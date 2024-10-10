package idemp

import (
	"context"
)

type Repository interface {
	SaveKey(ctx context.Context, idempKey string) error
}
