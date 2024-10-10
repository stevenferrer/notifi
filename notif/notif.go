package notif

import (
	"time"

	"github.com/stevenferrer/notifi/callback"
	"github.com/stevenferrer/notifi/token"
)

// ID is the notification ID
type ID string

const (
	NilID ID = ""
)

type Notif struct {
	// ID is the reference ID
	ID ID
	//SrcTokenID is the token id of notif sender
	SrcTokenID token.ID
	// DestTokenID is the token id of notif receiver
	DestTokenID token.ID
	// CBType is the callback type
	CBType callback.CBType
	// Status is the notification status (e.g. complete, failed, pending, etc.)
	Status Status
	// Payload is the notification payload in JSON format
	Payload map[string]interface{}
	// CreatedAt is the created timestamp
	CreatedAt *time.Time

	// TODO: We could probably attach the send events here? or use a separate table?
}
