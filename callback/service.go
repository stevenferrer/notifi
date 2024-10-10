package callback

import (
	"context"

	"github.com/stevenferrer/notifi/token"
)

// Service is the callback service
type Service interface {
	CreateCallback(context.Context, Callback) (ID, error)
	GetCallback(context.Context, ID) (*Callback, error)
	GetCbByTokenIDnCbType(context.Context, token.ID, CBType) (*Callback, error)
	TestCallback(context.Context, ID) error
	// UpdateCallback - update callback definition
}
