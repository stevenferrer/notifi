package notif

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"

	"github.com/stevenferrer/notifi/callback"
	"github.com/stevenferrer/notifi/token"
)

type Sender interface {
	Send(context.Context, NotifMsg) error
}

type NotifMsg struct {
	NotifID     ID                     `json:"notif_id"`
	DestTokenID token.ID               `json:"dest_token_id"`
	CBType      callback.CBType        `json:"cb_type"`
	Payload     map[string]interface{} `json:"payload"`
	RetryCount  int                    `json:"retry_count" hash:"ignore"`
}

func (msg NotifMsg) IdempKey() (string, error) {
	hash, err := hashstructure.Hash(msg, hashstructure.FormatV2, nil)
	if err != nil {
		return "", errors.New("hash structure")
	}

	return strconv.FormatUint(hash, 16), nil
}

type NotifSender struct {
	ch *amqp.Channel
}

const (
	defaultExchange   = "notifs_exchange"
	defaultQueue      = "notifs_queue"
	defaultRoutingKey = "notifs_queue.default"
)

var _ Sender = (*NotifSender)(nil)

func NewNotifSender(ch *amqp.Channel) (*NotifSender, error) {
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
		return nil, errors.Wrap(err, "exchange declare")
	}

	return &NotifSender{ch: ch}, nil

}

func (sender *NotifSender) Send(ctx context.Context, nfMsg NotifMsg) error {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(nfMsg)
	if err != nil {
		return errors.Wrap(err, "json encode message")
	}

	err = sender.ch.Publish(
		defaultExchange,   // exchange
		defaultRoutingKey, // routing key
		false,             // mandatory
		false,             // immediate
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

type NopSender struct{}

var _ Sender = (*NopSender)(nil)

func (sender *NopSender) Send(ctx context.Context, nfMsg NotifMsg) error {
	return nil
}
