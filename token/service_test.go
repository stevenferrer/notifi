package token_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stevenferrer/notifi/postgres"
	"github.com/stevenferrer/notifi/postgres/migration"
	"github.com/stevenferrer/notifi/postgres/txdb"
	"github.com/stevenferrer/notifi/token"
)

func TestTokenService(t *testing.T) {
	db := txdb.MustOpen()
	defer db.Close()

	migration.MustMigrate(db)

	tokenRepo := postgres.NewTokenRepository(db)
	tokenSvc := token.NewTokenService(tokenRepo)

	ctx := context.TODO()
	tk, err := tokenSvc.CreateToken(ctx)
	require.NoError(t, err)

	gotTk, err := tokenSvc.GetToken(ctx, tk.ID)
	require.NoError(t, err)

	assert.Equal(t, tk.ID, gotTk.ID)
	assert.Equal(t, tk.CBKey, gotTk.CBKey)

	// token not found
	_, err = tokenSvc.GetToken(ctx, token.NewID())
	assert.ErrorIs(t, err, token.ErrTokenNotFound)
}
