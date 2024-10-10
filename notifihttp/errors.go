package notifihttp

import (
	"net/http"
	"strings"
)

// Error is an API error
type Error struct {
	Err     error  `json:"-"`
	Status  int    `json:"-"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Err.Error()
}

func NewBadRequestError(err error) *Error {
	status := http.StatusBadRequest
	return &Error{
		Err:     err,
		Status:  status,
		Message: statusText(status),
	}
}

func NewNotFoundError(err error) *Error {
	status := http.StatusNotFound
	return &Error{
		Err:     err,
		Status:  status,
		Message: statusText(status),
	}
}

func NewInternalServerError(err error) *Error {
	status := http.StatusInternalServerError
	return &Error{
		Err:     err,
		Status:  status,
		Message: statusText(status),
	}
}

func NewForbiddenError(err error) *Error {
	status := http.StatusForbidden
	return &Error{
		Err:     err,
		Status:  status,
		Message: statusText(status),
	}
}

func NewHTTPError(status int, err error) *Error {
	return &Error{
		Err:     err,
		Status:  status,
		Message: statusText(status),
	}
}

func statusText(status int) string {
	return strings.ToLower(http.StatusText(status))
}
