package token

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
)

type Service interface {
	CreateToken(context.Context) (*Token, error)
	GetToken(context.Context, ID) (*Token, error)
}

type TokenService struct{ repo Repository }

var _ Service = (*TokenService)(nil)

func NewTokenService(repo Repository) *TokenService {
	return &TokenService{repo: repo}
}

func (tks *TokenService) CreateToken(ctx context.Context) (*Token, error) {
	tk := Token{
		ID:    NewID(),
		CBKey: NewCBKey(),
	}

	err := tks.repo.CreateToken(ctx, tk)
	if err != nil {
		return nil, errors.Wrap(err, "create token")
	}

	return &tk, nil
}

func (tks *TokenService) GetToken(ctx context.Context, tkID ID) (*Token, error) {
	tk, err := tks.repo.GetToken(ctx, tkID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrTokenNotFound
		}

		return nil, errors.Wrap(err, "get token")
	}

	return tk, nil
}
