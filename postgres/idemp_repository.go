package postgres

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"
	"github.com/stevenferrer/notifi/idemp"
)

type IdempRepository struct {
	db *sql.DB
}

var _ idemp.Repository = (*IdempRepository)(nil)

func NewIdempRepository(db *sql.DB) *IdempRepository {
	return &IdempRepository{db: db}
}

func (repo *IdempRepository) SaveKey(ctx context.Context, idempKey string) error {
	exists, err := repo.checkKeyExists(ctx, idempKey)
	if err != nil {
		return errors.Wrap(err, "check key exists")
	}

	if exists {
		return nil
	}

	stmnt := `insert into idemp_keys (key) values ($1)`
	_, err = repo.db.ExecContext(ctx, stmnt, idempKey)
	if err != nil {
		return errors.Wrap(err, "exec context")
	}

	return nil
}

func (repo *IdempRepository) checkKeyExists(ctx context.Context,
	idemPkey string) (bool, error) {
	stmnt := `select exists(select 1 from idemp_keys where key=$1)`
	var exists bool
	err := repo.db.QueryRowContext(ctx, stmnt, idemPkey).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, errors.Wrap(err, "query row context")
	}

	return exists, nil
}
