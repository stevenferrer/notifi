package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stevenferrer/notifi/callback"
	"github.com/stevenferrer/notifi/notif"
	"github.com/stevenferrer/notifi/postgres"
	"github.com/stevenferrer/notifi/postgres/migration"
	"github.com/stevenferrer/notifi/postgres/txdb"
	"github.com/stevenferrer/notifi/token"
)

func TestIdempRepository(t *testing.T) {
	db := txdb.MustOpen()
	defer db.Close()

	migration.MustMigrate(db)

	idempRepo := postgres.NewIdempRepository(db)

	msg := notif.NotifMsg{
		NotifID:     notif.NewID(),
		DestTokenID: token.NewID(),
		CBType:      callback.CBType("INVOICE"),
		Payload: map[string]interface{}{
			"message": "Hello",
		},
	}

	idemKey, err := msg.IdempKey()
	require.NoError(t, err)

	ctx := context.TODO()
	err = idempRepo.SaveKey(ctx, idemKey)
	require.NoError(t, err)

	// saving twice should not error
	err = idempRepo.SaveKey(ctx, idemKey)
	require.NoError(t, err)
}
