package postgres

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"

	"github.com/stevenferrer/notifi/token"
)

type TokenRepository struct {
	db *sql.DB
}

var _ token.Repository = (*TokenRepository)(nil)

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

func (repo *TokenRepository) CreateToken(ctx context.Context, tk token.Token) error {
	stmnt := `insert into tokens (id, cb_key) values ($1, $2)`
	_, err := repo.db.ExecContext(ctx, stmnt, tk.ID, tk.CBKey)
	if err != nil {
		return errors.Wrap(err, "exec context")
	}

	return nil
}

func (repo *TokenRepository) GetToken(ctx context.Context, tokenID token.ID) (*token.Token, error) {
	stmnt := `select id, cb_key from tokens where id = $1`
	var tk token.Token
	err := repo.db.QueryRowContext(ctx, stmnt, tokenID).Scan(&tk.ID, &tk.CBKey)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, err
		}
		return nil, errors.Wrap(err, "query row context")
	}

	return &tk, nil
}
