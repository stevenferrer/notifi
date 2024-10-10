package notif

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/streadway/amqp"
)

type NotifWorker struct {
	ch           *amqp.Channel
	doneChan     chan bool
	msgProcessor *NotifMsgProcessor
	logger       zerolog.Logger
}

func NewNotifWorker(
	ch *amqp.Channel,
	msgProcessor *NotifMsgProcessor,
	logger zerolog.Logger,
) (*NotifWorker, error) {
	// declare exchange
	err := ch.ExchangeDeclare(
		defaultExchange,
		"topic",
		true,  // durable
		false, // auto-delete
		false, //internal
		false, // no-wait,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "declare exchange")
	}

	// declare default queue
	_, err = ch.QueueDeclare(
		defaultQueue, // name,
		true,         // durable
		false,        // delete when unused
		false,        // exclusive,
		false,        // no-wait
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "declare queue")
	}

	// bind channel to default queue
	err = ch.QueueBind(
		defaultQueue,
		defaultRoutingKey,
		defaultExchange,
		false,
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "queue bind")
	}

	return &NotifWorker{
		ch:           ch,
		msgProcessor: msgProcessor,
		doneChan:     make(chan bool),
		logger:       logger,
	}, nil
}

func (worker *NotifWorker) Start(ctx context.Context) error {
	msgs, err := worker.ch.Consume(
		defaultQueue, // queue
		"",           // consumer
		false,        // auto-ack
		false,        // exclusive
		false,        // no local
		false,        // no wait
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "consume messages")
	}

	// TODO: add semaphore

	var wg sync.WaitGroup
	for {
		select {
		case msg := <-msgs:
			wg.Add(1)
			go func() {
				defer wg.Done()
				// Process new messages
				err = worker.msgProcessor.Process(ctx, msg.Body)
				if err != nil {
					worker.logger.Error().Err(err).Msg("process message")
					// Retry
					err = worker.retryMsg(ctx, msg)
					if err != nil {
						worker.logger.Error().Err(err).Msg("retrying message")
					}
				}

				// Send ack whether success or not
				err = msg.Ack(false)
				if err != nil {
					worker.logger.Error().Err(err).Msg("ack")
				}
			}()

		case <-worker.doneChan:
			close(worker.doneChan)
			// Wait for goroutines to finish
			wg.Wait()
			return nil
		}
	}
}

func (worker *NotifWorker) Stop() {
	worker.doneChan <- true
}

func (worker *NotifWorker) retryMsg(ctx context.Context, msg amqp.Delivery) error {
	var notifMsg NotifMsg
	err := json.NewDecoder(bytes.NewBuffer(msg.Body)).Decode(&notifMsg)
	if err != nil {
		return errors.Wrap(err, "json decode message")
	}

	// compute delay based on retry count
	delay := secondsToDelay(notifMsg.RetryCount)
	// format retry queue name
	routingKey := fmt.Sprintf("%s.%d", defaultQueue, delay)

	retryQueueName, err := worker.createRetryQueue(delay)
	if err != nil {
		return errors.Wrap(err, "create retry queue")
	}

	// Bind retyr queue to default exchanges
	err = worker.ch.QueueBind(
		retryQueueName,
		routingKey,
		defaultExchange,
		false,
		nil,
	)
	if err != nil {
		return errors.Wrap(err, "bind retry queue to exchange")
	}

	// Increase retry count
	notifMsg.RetryCount += 1

	buf := &bytes.Buffer{}
	err = json.NewEncoder(buf).Encode(notifMsg)
	if err != nil {
		return errors.Wrap(err, "json encode message")
	}

	err = worker.ch.Publish(
		defaultExchange, // exchange
		routingKey,      // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         buf.Bytes(),
		},
	)
	if err != nil {
		return errors.Wrap(err, "publish message")
	}

	return nil
}

func (worker *NotifWorker) createRetryQueue(delay int) (string, error) {
	queueName := fmt.Sprintf("%s.retry.%d", defaultQueue, delay)

	// declare retry queue
	q, err := worker.ch.QueueDeclare(
		queueName, // name,
		true,      // durable
		false,     // delete when unused
		false,     // exclusive,
		false,     // no-wait
		amqp.Table{
			"x-dead-letter-exchange":    defaultExchange,
			"x-dead-letter-routing-key": "notifs_queue.default",
			"x-message-ttl":             delay * 1000,
			"x-expires":                 delay * 1000 * 2,
		},
	)
	if err != nil {
		return "", errors.Wrap(err, "declare retry queue")
	}

	return q.Name, nil
}

func secondsToDelay(count int) int {
	base := count + 1
	return base * base
}
