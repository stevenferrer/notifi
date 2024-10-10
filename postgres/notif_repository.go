package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/stevenferrer/notifi/notif"
)

type NotifRepository struct{ db *sql.DB }

var _ notif.Repository = (*NotifRepository)(nil)

func NewNotifRepository(db *sql.DB) *NotifRepository {
	return &NotifRepository{db: db}
}

func (repo *NotifRepository) CreateNotif(ctx context.Context, nf notif.Notif) error {
	payload, err := json.Marshal(nf.Payload)
	if err != nil {
		return errors.Wrap(err, "marshal payload")
	}

	stmnt := `insert into notifications (id, src_token_id, 
			dest_token_id, cb_type, status, payload)
		values ($1, $2, $3, $4, $5, $6)`
	_, err = repo.db.ExecContext(ctx, stmnt, nf.ID, nf.SrcTokenID,
		nf.DestTokenID, nf.CBType, nf.Status, payload)
	if err != nil {
		return errors.Wrap(err, "exec context")
	}

	return nil
}

func (repo *NotifRepository) GetNotif(ctx context.Context, notifID notif.ID) (*notif.Notif, error) {
	stmnt := `select id, src_token_id, dest_token_id, cb_type, 
		status, payload from notifications where id=$1`
	var (
		nf      notif.Notif
		payload []byte
	)
	err := repo.db.QueryRowContext(ctx, stmnt, notifID).
		Scan(&nf.ID, &nf.SrcTokenID, &nf.DestTokenID,
			&nf.CBType, &nf.Status, &payload)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, notif.ErrNotifNotFound
		}

		return nil, errors.Wrap(err, "query row context")
	}

	err = json.Unmarshal(payload, &nf.Payload)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal payload")
	}

	return &nf, nil
}

func (repo *NotifRepository) UpdateStatus(ctx context.Context,
	notifID notif.ID, status notif.Status) error {
	nf, err := repo.GetNotif(ctx, notifID)
	if err != nil {
		if err == notif.ErrNotifNotFound {
			return err
		}

		return errors.Wrap(err, "get notif")
	}

	// No need to update
	if nf.Status == status {
		return nil
	}

	stmnt := `update notifications set status=$1, 
		updated_at=NOW() where id=$2`
	_, err = repo.db.ExecContext(ctx, stmnt, status, notifID)
	if err != nil {
		return errors.Wrap(err, "exec context")
	}

	return nil
}
