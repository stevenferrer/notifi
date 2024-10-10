package notifihttp

import (
	"context"

	"github.com/stevenferrer/notifi/token"
)

type ctxKey int

const (
	tokenCtxKey ctxKey = iota
)

func TokenFromCtx(ctx context.Context) (*token.Token, bool) {
	tk, ok := ctx.Value(tokenCtxKey).(*token.Token)
	return tk, ok
}

func CtxWithToken(ctx context.Context, tk *token.Token) context.Context {
	return context.WithValue(ctx, tokenCtxKey, tk)
}
