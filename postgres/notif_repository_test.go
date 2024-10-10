package postgres_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stevenferrer/notifi/callback"
	"github.com/stevenferrer/notifi/notif"
	"github.com/stevenferrer/notifi/postgres"
	"github.com/stevenferrer/notifi/postgres/migration"
	"github.com/stevenferrer/notifi/postgres/txdb"
	"github.com/stevenferrer/notifi/token"
)

func TestNotifRepository(t *testing.T) {
	db := txdb.MustOpen()
	defer db.Close()

	migration.MustMigrate(db)

	tokenRepo := postgres.NewTokenRepository(db)
	callbackRepo := postgres.NewCallbackRepository(db)
	notifRepo := postgres.NewNotifRepository(db)

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

	// create notif
	nf := notif.Notif{
		ID:          notif.NewID(),
		SrcTokenID:  tk.ID,
		DestTokenID: tk.ID,
		CBType:      cb.CBType,
		Status:      notif.StatusPending,
		Payload: map[string]interface{}{
			"id":      "1234",
			"message": "Hello",
		},
	}
	err = notifRepo.CreateNotif(ctx, nf)
	require.NoError(t, err)

	// get notification
	gotNf, err := notifRepo.GetNotif(ctx, nf.ID)
	require.NoError(t, err)

	assert.Equal(t, nf.ID, gotNf.ID)
	assert.Equal(t, nf.SrcTokenID, gotNf.SrcTokenID)
	assert.Equal(t, nf.DestTokenID, gotNf.DestTokenID)
	assert.Equal(t, nf.CBType, gotNf.CBType)
	assert.Equal(t, nf.Status, gotNf.Status)
	assert.Equal(t, nf.Payload, gotNf.Payload)

	// update status to failed
	err = notifRepo.UpdateStatus(ctx, nf.ID, notif.StatusFailed)
	require.NoError(t, err)

	// verify status has been updated to failed
	gotNf, err = notifRepo.GetNotif(ctx, nf.ID)
	require.NoError(t, err)

	assert.Equal(t, notif.StatusFailed, gotNf.Status)
}
