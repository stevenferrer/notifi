package service_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stevenferrer/notifi/callback"
	callbacksvc "github.com/stevenferrer/notifi/callback/service"
	"github.com/stevenferrer/notifi/notifihttp"
	"github.com/stevenferrer/notifi/postgres"
	"github.com/stevenferrer/notifi/postgres/migration"
	"github.com/stevenferrer/notifi/postgres/txdb"
	"github.com/stevenferrer/notifi/token"
)

func TestCallbackService(t *testing.T) {
	db := txdb.MustOpen()
	defer db.Close()

	migration.MustMigrate(db)

	tokenRepo := postgres.NewTokenRepository(db)
	tokenSvc := token.NewTokenService(tokenRepo)

	callbackRepo := postgres.NewCallbackRepository(db)
	requestSender := notifihttp.NewDefaultRequestSender()
	callbackSvc := callbacksvc.NewCallbackService(callbackRepo, tokenRepo, requestSender)

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
	cbID, err := callbackSvc.CreateCallback(ctx, cb)
	require.NoError(t, err)

	// create same callback twice should error
	_, err = callbackSvc.CreateCallback(ctx, cb)
	require.ErrorIs(t, err, callback.ErrCallbackExists)

	// get callback by id
	gotCb, err := callbackSvc.GetCallback(ctx, cbID)
	require.NoError(t, err)

	assert.Equal(t, cbID, gotCb.ID)
	assert.Equal(t, cb.TokenID, gotCb.TokenID)
	assert.Equal(t, cb.CBType, gotCb.CBType)
	assert.Equal(t, cb.URL, gotCb.URL)

	// get callback by token id and cb type
	gotCb, err = callbackSvc.GetCbByTokenIDnCbType(ctx, tk.ID, cb.CBType)
	require.NoError(t, err)

	assert.Equal(t, cbID, gotCb.ID)
	assert.Equal(t, cb.TokenID, gotCb.TokenID)
	assert.Equal(t, cb.CBType, gotCb.CBType)
	assert.Equal(t, cb.URL, gotCb.URL)

	// callback not found
	_, err = callbackSvc.GetCallback(ctx, callback.NewID())
	assert.ErrorIs(t, err, callback.ErrCallbackNotFound)

	_, err = callbackSvc.GetCbByTokenIDnCbType(ctx, token.NewID(), cb.CBType)
	assert.ErrorIs(t, err, callback.ErrCallbackNotFound)
}
