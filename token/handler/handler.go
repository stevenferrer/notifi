package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/unrolled/render"

	"github.com/stevenferrer/notifi/notifihttp"
	"github.com/stevenferrer/notifi/token"
)

type tokenHandler struct {
	tks    token.Service
	mux    *chi.Mux
	logger zerolog.Logger
	render *render.Render
}

func NewTokenHandler(tks token.Service, logger zerolog.Logger) http.Handler {
	tkh := &tokenHandler{
		tks:    tks,
		mux:    chi.NewMux(),
		logger: logger,
		render: render.New(),
	}

	// helper for adding http route
	addRoute := func(method, pattern string, h notifihttp.Handler) {
		tkh.mux.Method(method, pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := h.ServeHTTP(w, r)
			if err != nil {
				tkh.handleError(w, r, err)
			}
		}))
	}

	addRoute(http.MethodPost, "/", createToken(tkh))
	addRoute(http.MethodGet, "/me", getToken(tkh))

	return tkh
}

func (tkh *tokenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tkh.mux.ServeHTTP(w, r)
}

func (tkh *tokenHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	tkh.logger.Error().Err(err).Msg("token handler error")

	apiErr := notifihttp.NewInternalServerError(err)
	if e, ok := err.(*notifihttp.Error); ok {
		apiErr = e
	}

	w.WriteHeader(apiErr.Status)
	err = json.NewEncoder(w).Encode(apiErr)
	if err != nil {
		tkh.logger.Error().Err(err).Msg("json encode")
	}
}

func createToken(tkh *tokenHandler) notifihttp.Handler {
	return notifihttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		tk, err := tkh.tks.CreateToken(r.Context())
		if err != nil {
			return err
		}

		return tkh.render.JSON(w, http.StatusOK, tk)
	})
}

func getToken(tkh *tokenHandler) notifihttp.Handler {
	return notifihttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		tk, ok := notifihttp.TokenFromCtx(r.Context())
		if !ok {
			return errors.New("token is not injected into context")
		}

		return tkh.render.JSON(w, http.StatusOK, tk)
	})
}
