package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stevenferrer/notifi/callback"
	"github.com/stevenferrer/notifi/postgres"
	"github.com/stevenferrer/notifi/postgres/migration"
	"github.com/stevenferrer/notifi/postgres/txdb"
	"github.com/stevenferrer/notifi/token"
)

func TestCallbackRepository(t *testing.T) {
	db := txdb.MustOpen()
	defer db.Close()

	migration.MustMigrate(db)

	tokenRepo := postgres.NewTokenRepository(db)
	callbackRepo := postgres.NewCallbackRepository(db)

	ctx := context.TODO()

	// create token
	tk := token.Token{
		ID:    token.NewID(),
		CBKey: token.NewCBKey(),
	}
	err := tokenRepo.CreateToken(ctx, tk)
	require.NoError(t, err)

	// create callback
	cb := callback.Callback{
		ID:      callback.NewID(),
		TokenID: tk.ID,
		CBType:  "INVOICE",
		URL:     "https://example.com",
	}
	err = callbackRepo.CreateCallback(ctx, cb)
	require.NoError(t, err)

	// creating same callback twice should give an error
	err = callbackRepo.CreateCallback(ctx, cb)
	require.ErrorIs(t, err, callback.ErrCallbackExists)

	// get callback
	gotCb, err := callbackRepo.GetCallback(ctx, cb.ID)
	require.NoError(t, err)

	assert.Equal(t, cb.ID, gotCb.ID)
	assert.Equal(t, cb.TokenID, gotCb.TokenID)
	assert.Equal(t, cb.CBType, gotCb.CBType)
	assert.Equal(t, cb.URL, gotCb.URL)

	// get callback by token id and cb type
	gotCb, err = callbackRepo.GetCbByTokenIDnCbType(ctx, tk.ID, cb.CBType)
	require.NoError(t, err)

	assert.Equal(t, cb.ID, gotCb.ID)
	assert.Equal(t, cb.TokenID, gotCb.TokenID)
	assert.Equal(t, cb.CBType, gotCb.CBType)
	assert.Equal(t, cb.URL, gotCb.URL)

	// callback not exist
	_, err = callbackRepo.GetCallback(ctx, callback.NewID())
	assert.ErrorIs(t, err, callback.ErrCallbackNotFound)

	_, err = callbackRepo.GetCbByTokenIDnCbType(ctx, token.NewID(), cb.CBType)
	require.ErrorIs(t, err, callback.ErrCallbackNotFound)
}
