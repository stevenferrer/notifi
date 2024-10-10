package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stevenferrer/notifi/postgres"
	"github.com/stevenferrer/notifi/postgres/migration"
	"github.com/stevenferrer/notifi/postgres/txdb"
	"github.com/stevenferrer/notifi/token"
	tokenhandler "github.com/stevenferrer/notifi/token/handler"
)

func TestTokenHandler(t *testing.T) {
	db := txdb.MustOpen()
	defer db.Close()

	migration.MustMigrate(db)

	tokenRepo := postgres.NewTokenRepository(db)
	tokenSvc := token.NewTokenService(tokenRepo)

	ctx := context.TODO()
	logger := zerolog.New(os.Stderr)
	tokenHandler := tokenhandler.NewTokenHandler(tokenSvc, logger)
	authHandler := tokenhandler.NewTokenMw(tokenSvc)(tokenHandler)

	var apiKey string
	t.Run("Create token", func(t *testing.T) {
		// create token
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "/", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		tokenHandler.ServeHTTP(rr, httpReq)
		require.Equal(t, http.StatusOK, rr.Code)

		var resp = struct {
			APIKey string `json:"api_key"`
			CBKey  string `json:"cb_key"`
		}{}
		err = json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.APIKey)
		assert.NotEmpty(t, resp.CBKey)

		apiKey = resp.APIKey
	})

	// get token
	t.Run("Get token", func(t *testing.T) {
		httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "/me", nil)
		require.NoError(t, err)

		httpReq.Header.Add("X-API-KEY", apiKey)

		rr := httptest.NewRecorder()
		authHandler.ServeHTTP(rr, httpReq)

		var resp = struct {
			APIKey string `json:"api_key"`
			CBKey  string `json:"cb_key"`
		}{}
		err = json.NewDecoder(rr.Body).Decode(&resp)
		require.NoError(t, err)

		assert.NotEmpty(t, resp.APIKey)
		assert.NotEmpty(t, resp.CBKey)
	})
}
