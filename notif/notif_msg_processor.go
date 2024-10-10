package notif

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"go.uber.org/multierr"

	"github.com/stevenferrer/notifi/callback"
	"github.com/stevenferrer/notifi/idemp"
	"github.com/stevenferrer/notifi/notifihttp"
	"github.com/stevenferrer/notifi/token"
)

type NotifMsgProcessor struct {
	requestSender notifihttp.RequestSender
	callbackRepo  callback.Repository
	notifRepo     Repository
	tokenRepo     token.Repository
	idempRepo     idemp.Repository
}

func NewNotifMessageProcessor(
	requestSender notifihttp.RequestSender,
	callbackRepo callback.Repository,
	notifRepo Repository,
	tokenRepo token.Repository,
	idemprepo idemp.Repository,
) *NotifMsgProcessor {
	return &NotifMsgProcessor{
		requestSender: requestSender,
		callbackRepo:  callbackRepo,
		notifRepo:     notifRepo,
		tokenRepo:     tokenRepo,
		idempRepo:     idemprepo,
	}
}

func (nmp *NotifMsgProcessor) Process(ctx context.Context, msgBody []byte) error {
	var notifMsg NotifMsg
	err := json.NewDecoder(bytes.NewBuffer(msgBody)).Decode(&notifMsg)
	if err != nil {
		return errors.Wrap(err, "json decode message")
	}

	// Compute and save idempkey
	idempKey, err := notifMsg.IdempKey()
	if err != nil {
		return errors.Wrap(err, "notif msg idemp key")
	}

	// save idempkey
	err = nmp.idempRepo.SaveKey(ctx, idempKey)
	if err != nil {
		return errors.Wrap(err, "save idemp key")
	}

	// 1. Retrieve URL from database
	cb, err := nmp.callbackRepo.GetCbByTokenIDnCbType(
		ctx, notifMsg.DestTokenID, notifMsg.CBType)
	if err != nil {
		return errors.Wrap(err, "get cb by token id and cb type")
	}

	tk, err := nmp.tokenRepo.GetToken(ctx, cb.TokenID)
	if err != nil {
		return errors.Wrap(err, "get token")
	}

	buf := &bytes.Buffer{}
	err = json.NewEncoder(buf).Encode(notifMsg.Payload)
	if err != nil {
		// Update notif status to failed and return an error
		err2 := nmp.notifRepo.UpdateStatus(ctx, notifMsg.NotifID, StatusFailed)
		if err != nil {
			err = multierr.Append(err, err2)
			return errors.Wrap(err, "update notif status")
		}

		return errors.Wrap(err, "json encode notif payload")
	}

	headers := map[string]string{
		"X-IDEMPOTENT-KEY": idempKey,
		"X-CALLBACK-TOKEN": string(tk.CBKey),
	}

	// 2. Send the notification request
	err = nmp.requestSender.SendRequest(ctx, cb.URL, buf, headers)
	if err != nil {
		// Update notif status to failed and return an error
		err2 := nmp.notifRepo.UpdateStatus(ctx, notifMsg.NotifID, StatusFailed)
		if err != nil {
			err = multierr.Append(err, err2)
			return errors.Wrap(err, "update notif status")
		}

		return errors.Wrap(err, "send notif request")
	}

	// Yay, no error! Update notif status to complete
	err = nmp.notifRepo.UpdateStatus(ctx, notifMsg.NotifID, StatusComplete)
	if err != nil {
		return errors.Wrap(err, "update notif status")
	}

	return nil
}
