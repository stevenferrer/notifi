package postgres

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"

	"github.com/stevenferrer/notifi/callback"
	"github.com/stevenferrer/notifi/token"
)

type CallbackRepository struct{ db *sql.DB }

var _ callback.Repository = (*CallbackRepository)(nil)

func NewCallbackRepository(db *sql.DB) *CallbackRepository {
	return &CallbackRepository{db: db}
}

func (repo *CallbackRepository) CreateCallback(ctx context.Context,
	cb callback.Callback) error {
	// check of callback type already exists for token (user)
	exists, err := repo.checkCbExists(ctx, cb.TokenID, cb.CBType)
	if err != nil {
		return errors.Wrap(err, "check callback exists")
	}

	// callback exists
	if exists {
		return callback.ErrCallbackExists
	}

	stmnt := `insert into callbacks (id, token_id, cb_type, cb_url)
		values ($1, $2, $3, $4)`
	_, err = repo.db.ExecContext(ctx, stmnt, cb.ID, cb.TokenID, cb.CBType, cb.URL)
	if err != nil {
		return errors.Wrap(err, "exec context")
	}

	return nil
}

func (repo *CallbackRepository) GetCallback(ctx context.Context, cbID callback.ID) (*callback.Callback, error) {
	stmnt := `select id, token_id, cb_type, cb_url from callbacks where id=$1`
	var cb callback.Callback
	err := repo.db.QueryRowContext(ctx, stmnt, cbID).
		Scan(&cb.ID, &cb.TokenID, &cb.CBType, &cb.URL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, callback.ErrCallbackNotFound
		}
		return nil, errors.Wrap(err, "query row context")
	}

	return &cb, nil
}

func (repo *CallbackRepository) GetCbByTokenIDnCbType(ctx context.Context,
	tokenID token.ID, cbType callback.CBType) (*callback.Callback, error) {
	stmnt := `select id, token_id, cb_type, cb_url from callbacks 
		where token_id=$1 and cb_type=$2`
	var cb callback.Callback
	err := repo.db.QueryRowContext(ctx, stmnt, tokenID, cbType).
		Scan(&cb.ID, &cb.TokenID, &cb.CBType, &cb.URL)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, callback.ErrCallbackNotFound
		}
		return nil, errors.Wrap(err, "query row context")
	}

	return &cb, nil
}

func (repo *CallbackRepository) checkCbExists(ctx context.Context,
	tokenID token.ID, cbType callback.CBType) (bool, error) {
	stmnt := `select exists(select 1 from callbacks 
		where token_id=$1 and cb_type=$2)`
	var exists bool
	err := repo.db.QueryRowContext(ctx, stmnt,
		tokenID, cbType).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, errors.Wrap(err, "query row context")
	}

	return exists, nil
}
