package notif_test

import (
	"testing"

	"github.com/stevenferrer/notifi/callback"
	"github.com/stevenferrer/notifi/notif"
	"github.com/stevenferrer/notifi/token"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotifMsgIdemptKey(t *testing.T) {
	msg := notif.NotifMsg{
		NotifID:     notif.NewID(),
		DestTokenID: token.NewID(),
		CBType:      callback.CBType("INVOICE"),
		Payload: map[string]interface{}{
			"message": "hello",
		},
	}

	idempKey, err := msg.IdempKey()
	require.NoError(t, err)
	assert.NotEmpty(t, idempKey)

	// modifying retry count should not affect idemp key
	msg.RetryCount = 1

	gotIdempKey, err := msg.IdempKey()
	require.NoError(t, err)
	assert.Equal(t, idempKey, gotIdempKey)
}
