package token

import (
	"context"
)

type Repository interface {
	CreateToken(context.Context, Token) error
	GetToken(context.Context, ID) (*Token, error)
}
