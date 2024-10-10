package callback

import (
	"context"

	"github.com/stevenferrer/notifi/token"
)

// Repository is the callback repository
type Repository interface {
	CreateCallback(context.Context, Callback) error
	GetCallback(context.Context, ID) (*Callback, error)
	GetCbByTokenIDnCbType(context.Context, token.ID, CBType) (*Callback, error)
}
