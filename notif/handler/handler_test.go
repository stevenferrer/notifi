package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stevenferrer/notifi/callback"
	callbacksvc "github.com/stevenferrer/notifi/callback/service"
	"github.com/stevenferrer/notifi/notif"
	nfhandler "github.com/stevenferrer/notifi/notif/handler"
	"github.com/stevenferrer/notifi/notifihttp"
	"github.com/stevenferrer/notifi/postgres"
	"github.com/stevenferrer/notifi/postgres/migration"
	"github.com/stevenferrer/notifi/postgres/txdb"
	"github.com/stevenferrer/notifi/token"
	tkhandler "github.com/stevenferrer/notifi/token/handler"
)

func TestNotifHandler(t *testing.T) {
	db := txdb.MustOpen()
	defer db.Close()

	migration.MustMigrate(db)

	tokenRepo := postgres.NewTokenRepository(db)
	tokenSvc := token.NewTokenService(tokenRepo)
	tokenMw := tkhandler.NewTokenMw(tokenSvc)

	callbackRepo := postgres.NewCallbackRepository(db)
	requestSender := notifihttp.NewDefaultRequestSender()
	callbackSvc := callbacksvc.NewCallbackService(callbackRepo,
		tokenRepo, requestSender)

	notifRepo := postgres.NewNotifRepository(db)
	notifSvc := notif.NewNotifService(notifRepo, &notif.NopSender{})

	logger := zerolog.New(os.Stderr)
	notifHandler := nfhandler.NewNotifHandler(notifSvc, logger)
	authHandler := tokenMw(notifHandler)

	ctx := context.TODO()

	// create a token
	tk, err := tokenSvc.CreateToken(ctx)
	require.NoError(t, err)

	// create callback
	cb := callback.Callback{
		TokenID: tk.ID,
		CBType:  "INVOICE",
		URL:     "https://example.com",
	}
	_, err = callbackSvc.CreateCallback(ctx, cb)
	require.NoError(t, err)

	var notifID notif.ID
	t.Run("Create notification", func(t *testing.T) {
		var request = struct {
			DestTokenID token.ID               `json:"dest_token_id"`
			CBType      callback.CBType        `json:"callback_type"`
			Payload     map[string]interface{} `json:"payload"`
		}{
			DestTokenID: tk.ID,
			CBType:      cb.CBType,
			Payload: map[string]interface{}{
				"id":      "1234",
				"message": "Hello",
			},
		}

		buf := &bytes.Buffer{}
		err := json.NewEncoder(buf).Encode(request)
		require.NoError(t, err)

		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "/", buf)
		require.NoError(t, err)

		httpReq.Header.Add("X-API-KEY", string(tk.ID))

		rr := httptest.NewRecorder()
		authHandler.ServeHTTP(rr, httpReq)
		require.Equal(t, http.StatusOK, rr.Code)

		var response = struct {
			NotifID notif.ID `json:"notification_id"`
		}{}
		err = json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.NotifID)

		notifID = response.NotifID
	})

	t.Run("Get notification", func(t *testing.T) {
		httpReq, err := http.NewRequestWithContext(ctx,
			http.MethodGet, "/"+string(notifID), nil)
		require.NoError(t, err)

		httpReq.Header.Add("X-API-KEY", string(tk.ID))

		rr := httptest.NewRecorder()
		authHandler.ServeHTTP(rr, httpReq)

		var response = struct {
			ID      notif.ID               `json:"notification_id"`
			CBType  callback.CBType        `json:"callback_type"`
			Status  notif.Status           `json:"status"`
			Payload map[string]interface{} `json:"payload"`
		}{}
		err = json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.ID)
		assert.NotEmpty(t, response.CBType)
		assert.NotEmpty(t, response.Status)
		assert.NotNil(t, response.Payload)
	})
}
