package callback

import (
	"github.com/stevenferrer/notifi/token"
)

// ID is the callback ID
type ID string

const (
	NilID = ID("")
)

// CBType is the callback type
type CBType string

// Callback is a callback subsription
type Callback struct {
	// ID is the callback ID
	ID ID
	// TokenID is the reference to token
	TokenID token.ID
	// CBType is the callback type
	CBType CBType
	// URL is the url to send callback
	URL string
}
