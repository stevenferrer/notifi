package handler

import (
	"net/http"

	"github.com/stevenferrer/notifi/notifihttp"
	"github.com/stevenferrer/notifi/token"
)

func NewTokenMw(tokenSvc token.Service) notifihttp.StdMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow only token path
			if r.URL.Path == "/token" ||
				r.URL.Path == "/test" {
				next.ServeHTTP(w, r)
				return
			}

			apiKey := r.Header.Get("X-API-KEY")
			if apiKey == "" {
				status := http.StatusUnauthorized
				http.Error(w, http.StatusText(status), status)
				return
			}

			tk, err := tokenSvc.GetToken(r.Context(), token.ID(apiKey))
			if err != nil {
				status := http.StatusInternalServerError
				if err == token.ErrTokenNotFound {
					status = http.StatusUnauthorized
				}

				http.Error(w, http.StatusText(status), status)
				return
			}

			next.ServeHTTP(w, r.WithContext(notifihttp.CtxWithToken(r.Context(), tk)))
		})
	}
}
