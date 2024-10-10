package notifihttp

import "net/http"

// Handler is the same as http.Handler except ServeHTTP may return an error
type Handler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request) error
}

// HandlerFunc is a convenience type like http.HandlerFunc
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

// ServeHTTP implements the Handler interface
func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return h(w, r)
}

type StdMiddleware func(http.Handler) http.Handler
