package service

import (
	"bytes"
	"context"

	"github.com/pkg/errors"
	"github.com/stevenferrer/notifi/callback"
	"github.com/stevenferrer/notifi/notifihttp"
	"github.com/stevenferrer/notifi/token"
)

// CallbackService implements the callback service
type CallbackService struct {
	callbackRepo  callback.Repository
	tokenRepo     token.Repository
	requestSender notifihttp.RequestSender
}

var _ callback.Service = (*CallbackService)(nil)

func NewCallbackService(callbackRepo callback.Repository,
	tokenRepo token.Repository,
	requestSender notifihttp.RequestSender) *CallbackService {
	return &CallbackService{
		callbackRepo:  callbackRepo,
		tokenRepo:     tokenRepo,
		requestSender: requestSender,
	}
}

func (cbs *CallbackService) CreateCallback(ctx context.Context,
	cb callback.Callback) (callback.ID, error) {
	cbID := callback.NewID()
	err := cbs.callbackRepo.CreateCallback(ctx, callback.Callback{
		ID:      cbID,
		TokenID: cb.TokenID,
		CBType:  cb.CBType,
		URL:     cb.URL,
	})
	if err != nil {
		if err == callback.ErrCallbackExists {
			return callback.NilID, err
		}

		return callback.NilID, errors.Wrap(err, "create callback")
	}

	return cbID, nil
}

func (cbs *CallbackService) GetCallback(ctx context.Context,
	cbID callback.ID) (*callback.Callback, error) {
	cb, err := cbs.callbackRepo.GetCallback(ctx, cbID)
	if err != nil {
		if err == callback.ErrCallbackNotFound {
			return nil, err
		}

		return nil, errors.Wrap(err, "get callack")
	}

	return cb, nil
}

func (cbs *CallbackService) GetCbByTokenIDnCbType(ctx context.Context,
	tokenID token.ID, cbType callback.CBType) (*callback.Callback, error) {
	cb, err := cbs.callbackRepo.GetCbByTokenIDnCbType(ctx, tokenID, cbType)
	if err != nil {
		if err == callback.ErrCallbackNotFound {
			return nil, err
		}

		return nil, errors.Wrap(err, "get callack")
	}

	return cb, nil
}

func (cbs *CallbackService) TestCallback(ctx context.Context, cbID callback.ID) error {
	// TODO: Implement this
	cb, err := cbs.GetCallback(ctx, cbID)
	if err != nil {
		return errors.Wrap(err, "get callback")
	}

	tk, err := cbs.tokenRepo.GetToken(ctx, cb.TokenID)
	if err != nil {
		return errors.Wrap(err, "get token")
	}

	buf := &bytes.Buffer{}
	buf.WriteString(`{"message":"Hello"}`)

	headers := map[string]string{
		"X-IDEMPOTENT-KEY": "test-1234",
		"X-CALLBACK-TOKEN": string(tk.CBKey),
	}

	err = cbs.requestSender.SendRequest(ctx, cb.URL, buf, headers)
	if err != nil {
		return errors.Wrap(err, "send request")
	}

	return nil
}
