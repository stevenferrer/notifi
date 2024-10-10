package callback

import "errors"

var (
	ErrCallbackExists    = errors.New("callback exists")
	ErrCallbackNotFound  = errors.New("callback not found")
	ErrCallbackURLNotSet = errors.New("callback url not set")
)
