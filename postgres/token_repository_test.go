package postgres_test

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

func TestTokenRepository(t *testing.T) {
	db := txdb.MustOpen()
	defer db.Close()

	migration.MustMigrate(db)

	tokenRepo := postgres.NewTokenRepository(db)
	ctx := context.TODO()

	tk := token.Token{
		ID:    token.NewID(),
		CBKey: token.NewCBKey(),
	}

	err := tokenRepo.CreateToken(ctx, tk)
	require.NoError(t, err)

	gotTk, err := tokenRepo.GetToken(ctx, tk.ID)
	require.NoError(t, err)

	assert.Equal(t, tk.ID, gotTk.ID)
	assert.Equal(t, tk.CBKey, gotTk.CBKey)
}
