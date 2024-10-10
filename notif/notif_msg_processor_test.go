package notif_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stevenferrer/notifi/callback"
	"github.com/stevenferrer/notifi/notif"
	"github.com/stevenferrer/notifi/notifihttp"
	"github.com/stevenferrer/notifi/postgres"
	"github.com/stevenferrer/notifi/postgres/migration"
	"github.com/stevenferrer/notifi/postgres/txdb"
	"github.com/stevenferrer/notifi/token"
)

func TestNotifMsgProcessor(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	db := txdb.MustOpen()
	defer db.Close()

	migration.MustMigrate(db)

	tokenRepo := postgres.NewTokenRepository(db)
	tokenSvc := token.NewTokenService(tokenRepo)

	callbackRepo := postgres.NewCallbackRepository(db)
	notifRepo := postgres.NewNotifRepository(db)

	idempRepo := postgres.NewIdempRepository(db)

	requestSender := notifihttp.NewDefaultRequestSender()
	notifMsgProc := notif.NewNotifMessageProcessor(
		requestSender, callbackRepo, notifRepo,
		tokenRepo, idempRepo)

	ctx := context.TODO()

	// create a token
	tk, err := tokenSvc.CreateToken(ctx)
	require.NoError(t, err)

	baseURL := "https://example.org"

	t.Run("Failed send", func(t *testing.T) {
		cbURL := baseURL + "/failed"
		httpmock.RegisterResponder(
			http.MethodPost,
			cbURL,
			func(r *http.Request) (*http.Response, error) {
				idempKey := r.Header.Get("X-IDEMPOTENT-KEY")
				if idempKey == "" {
					return nil, errors.New("idemp key is empty")
				}
				cbToken := r.Header.Get("X-CALLBACK-TOKEN")
				if cbToken == "" {
					return nil, errors.New("callbak token empty")
				}

				return httpmock.NewJsonResponse(http.StatusBadRequest, nil)
			},
		)

		// create callback
		cb := callback.Callback{
			ID:      callback.NewID(),
			TokenID: tk.ID,
			CBType:  "INVOICE",
			URL:     cbURL,
		}
		err = callbackRepo.CreateCallback(ctx, cb)
		require.NoError(t, err)

		nf := notif.Notif{
			ID:          notif.NewID(),
			SrcTokenID:  tk.ID,
			DestTokenID: tk.ID,
			CBType:      cb.CBType,
			Status:      notif.StatusPending,
			Payload: map[string]interface{}{
				"message": "hello",
			},
		}
		err = notifRepo.CreateNotif(ctx, nf)
		require.NoError(t, err)

		notifMsg := notif.NotifMsg{
			NotifID:     nf.ID,
			DestTokenID: nf.DestTokenID,
			CBType:      nf.CBType,
			Payload:     nf.Payload,
		}
		buf := &bytes.Buffer{}
		err = json.NewEncoder(buf).Encode(notifMsg)

		err = notifMsgProc.Process(ctx, buf.Bytes())
		require.Error(t, err)

		// very notif status
		gotNf, err := notifRepo.GetNotif(ctx, nf.ID)
		require.NoError(t, err)

		assert.Equal(t, notif.StatusFailed, gotNf.Status)
	})

	t.Run("Complete send", func(t *testing.T) {
		cbURL := baseURL + "/complete"
		httpmock.RegisterResponder(
			http.MethodPost, cbURL,
			func(r *http.Request) (*http.Response, error) {
				idempKey := r.Header.Get("X-IDEMPOTENT-KEY")
				if idempKey == "" {
					return nil, errors.New("idemp key is empty")
				}
				cbToken := r.Header.Get("X-CALLBACK-TOKEN")
				if cbToken == "" {
					return nil, errors.New("callbak token empty")
				}

				return httpmock.NewJsonResponse(http.StatusOK, nil)
			},
		)

		// create callback
		cb := callback.Callback{
			ID:      callback.NewID(),
			TokenID: tk.ID,
			CBType:  "INVOICE2",
			URL:     cbURL,
		}
		err = callbackRepo.CreateCallback(ctx, cb)
		require.NoError(t, err)

		nf := notif.Notif{
			ID:          notif.NewID(),
			SrcTokenID:  tk.ID,
			DestTokenID: tk.ID,
			CBType:      cb.CBType,
			Status:      notif.StatusPending,
			Payload: map[string]interface{}{
				"message": "hello",
			},
		}
		err = notifRepo.CreateNotif(ctx, nf)
		require.NoError(t, err)

		notifMsg := notif.NotifMsg{
			NotifID:     nf.ID,
			DestTokenID: nf.DestTokenID,
			CBType:      nf.CBType,
			Payload:     nf.Payload,
		}
		buf := &bytes.Buffer{}
		err = json.NewEncoder(buf).Encode(notifMsg)

		err = notifMsgProc.Process(ctx, buf.Bytes())
		require.NoError(t, err)

		// very notif status
		gotNf, err := notifRepo.GetNotif(ctx, nf.ID)
		require.NoError(t, err)

		assert.Equal(t, notif.StatusComplete, gotNf.Status)
	})
}
