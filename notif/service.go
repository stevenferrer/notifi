package notif

import (
	"context"

	"github.com/pkg/errors"
)

type Service interface {
	CreateNotif(context.Context, Notif) (ID, error)
	GetNotif(context.Context, ID) (*Notif, error)
	UpdateStatus(context.Context, ID, Status) error
	ResendNotif(context.Context, ID) error
}

type NotifService struct {
	notifRepo Repository
	sender    Sender
}

var _ Service = (*NotifService)(nil)

func NewNotifService(notifRepo Repository, sender Sender) *NotifService {
	return &NotifService{notifRepo: notifRepo, sender: sender}
}

func (ns *NotifService) CreateNotif(ctx context.Context, nf Notif) (ID, error) {
	// TODO: Validate that both src and dest token id exists??

	// Create notification record
	notifID := NewID()
	err := ns.notifRepo.CreateNotif(ctx, Notif{
		ID:          notifID,
		SrcTokenID:  nf.SrcTokenID,
		DestTokenID: nf.DestTokenID,
		CBType:      nf.CBType,
		Status:      StatusPending,
		Payload:     nf.Payload,
	})
	if err != nil {
		return NilID, errors.Wrap(err, "create notif")
	}

	//  Send the notification to queue
	err = ns.sender.Send(ctx, NotifMsg{
		NotifID:     notifID,
		DestTokenID: nf.DestTokenID,
		CBType:      nf.CBType,
		Payload:     nf.Payload,
	})
	if err != nil {
		return NilID, errors.Wrap(err, "send notif message to queue")
	}

	return notifID, nil
}

func (ns *NotifService) GetNotif(ctx context.Context, notifID ID) (*Notif, error) {
	nf, err := ns.notifRepo.GetNotif(ctx, notifID)
	if err != nil {
		if err == ErrNotifNotFound {
			return nil, err
		}

		return nil, errors.Wrap(err, "get notif")
	}

	return nf, nil
}

func (ns *NotifService) UpdateStatus(ctx context.Context, notifID ID, status Status) error {
	err := ns.notifRepo.UpdateStatus(ctx, notifID, status)
	if err != nil {
		if err == ErrNotifNotFound {
			return err
		}

		return errors.Wrap(err, "update status")
	}

	return nil
}

func (ns *NotifService) ResendNotif(ctx context.Context, notifID ID) error {
	nf, err := ns.GetNotif(ctx, notifID)
	if err != nil {
		return err
	}

	// check notif status before sending?
	if nf.Status != StatusComplete {
		return nil
	}

	//  Send the notification to queue
	err = ns.sender.Send(ctx, NotifMsg{
		NotifID:     notifID,
		DestTokenID: nf.DestTokenID,
		CBType:      nf.CBType,
		Payload:     nf.Payload,
	})
	if err != nil {
		return errors.Wrap(err, "resendsend notif message to queue")
	}

	return nil
}
