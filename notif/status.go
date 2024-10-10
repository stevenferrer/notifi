package notif

// Status is the notification status
type Status string

const (
	StatusPending  Status = "PENDING"
	StatusComplete Status = "COMPLETE"
	StatusFailed   Status = "FAILED"
)
