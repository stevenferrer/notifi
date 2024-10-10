package notif_test

import (
	"context"
	"testing"

	"github.com/stevenferrer/notifi/callback"
	callbacksvc "github.com/stevenferrer/notifi/callback/service"
	"github.com/stevenferrer/notifi/notif"
	"github.com/stevenferrer/notifi/notifihttp"
	"github.com/stevenferrer/notifi/postgres"
	"github.com/stevenferrer/notifi/postgres/migration"
	"github.com/stevenferrer/notifi/postgres/txdb"
	"github.com/stevenferrer/notifi/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifServie(t *testing.T) {
	db := txdb.MustOpen()
	defer db.Close()

	migration.MustMigrate(db)

	tokenRepo := postgres.NewTokenRepository(db)
	tokenSvc := token.NewTokenService(tokenRepo)

	callbackRepo := postgres.NewCallbackRepository(db)
	requestSender := notifihttp.NewDefaultRequestSender()
	callbackSvc := callbacksvc.NewCallbackService(callbackRepo,
		tokenRepo, requestSender)

	notifRepo := postgres.NewNotifRepository(db)
	notifSvc := notif.NewNotifService(notifRepo, &notif.NopSender{})

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

	// create notification
	nf := notif.Notif{
		SrcTokenID:  tk.ID,
		DestTokenID: tk.ID,
		CBType:      cb.CBType,
		Payload: map[string]interface{}{
			"id":      "1234",
			"message": "Hello",
		},
	}

	notifID, err := notifSvc.CreateNotif(ctx, nf)
	require.NoError(t, err)

	// get notification
	gotNf, err := notifSvc.GetNotif(ctx, notifID)
	require.NoError(t, err)

	assert.Equal(t, notifID, gotNf.ID)
	assert.Equal(t, nf.SrcTokenID, gotNf.SrcTokenID)
	assert.Equal(t, nf.DestTokenID, gotNf.DestTokenID)
	assert.Equal(t, nf.CBType, gotNf.CBType)
	assert.Equal(t, notif.StatusPending, gotNf.Status)
	assert.Equal(t, nf.Payload, gotNf.Payload)

	// get not found notif
	_, err = notifSvc.GetNotif(ctx, notif.NewID())
	require.ErrorIs(t, err, notif.ErrNotifNotFound)

	// update status
	err = notifSvc.UpdateStatus(ctx, notifID, notif.StatusComplete)
	require.NoError(t, err)

	gotNf, err = notifSvc.GetNotif(ctx, notifID)
	require.NoError(t, err)

	assert.Equal(t, notif.StatusComplete, gotNf.Status)

	// update not found notif
	err = notifSvc.UpdateStatus(ctx, notif.NewID(), notif.StatusComplete)
	require.ErrorIs(t, err, notif.ErrNotifNotFound)
}
