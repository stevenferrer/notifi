package notif

import "context"

type Repository interface {
	CreateNotif(context.Context, Notif) error
	GetNotif(context.Context, ID) (*Notif, error)
	UpdateStatus(context.Context, ID, Status) error
}
