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
	"github.com/stevenferrer/notifi/callback"
	cbhandler "github.com/stevenferrer/notifi/callback/handler"
	callbacksvc "github.com/stevenferrer/notifi/callback/service"
	"github.com/stevenferrer/notifi/notifihttp"
	"github.com/stevenferrer/notifi/postgres"
	"github.com/stevenferrer/notifi/postgres/migration"
	"github.com/stevenferrer/notifi/postgres/txdb"
	"github.com/stevenferrer/notifi/token"
	tkhandler "github.com/stevenferrer/notifi/token/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallbackHandler(t *testing.T) {
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

	logger := zerolog.New(os.Stderr)
	cbHandler := cbhandler.NewCallbackHandler(callbackSvc, logger)
	authHandler := tokenMw(cbHandler)

	ctx := context.TODO()

	// create a token
	tk, err := tokenSvc.CreateToken(ctx)
	require.NoError(t, err)

	var cbID callback.ID
	t.Run("Create callback", func(t *testing.T) {
		var request = struct {
			CBType string `json:"callback_type"`
			URL    string `json:"url"`
		}{
			CBType: "INVOICE",
			URL:    "https://example.com",
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
			CBID string `json:"callback_id"`
		}{}
		err = json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.CBID)

		cbID = callback.ID(response.CBID)
	})

	t.Run("Get callback", func(t *testing.T) {
		httpReq, err := http.NewRequestWithContext(ctx,
			http.MethodGet, "/"+string(cbID), nil)
		require.NoError(t, err)

		httpReq.Header.Add("X-API-KEY", string(tk.ID))

		rr := httptest.NewRecorder()
		authHandler.ServeHTTP(rr, httpReq)

		var response = struct {
			CallbackID   string `json:"callback_id"`
			CallbackType string `json:"callback_type"`
			URL          string `json:"url"`
		}{}
		err = json.NewDecoder(rr.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.CallbackID)
		assert.NotEmpty(t, response.CallbackType)
		assert.NotEmpty(t, response.URL)
	})
}
