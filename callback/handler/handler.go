package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/unrolled/render"

	"github.com/stevenferrer/notifi/callback"
	"github.com/stevenferrer/notifi/notifihttp"
)

type callbackHandler struct {
	callbackSvc callback.Service
	mux         *chi.Mux
	logger      zerolog.Logger
	render      *render.Render
}

func NewCallbackHandler(callbackSvc callback.Service, logger zerolog.Logger) http.Handler {
	cbh := &callbackHandler{
		callbackSvc: callbackSvc,
		logger:      logger,
		mux:         chi.NewMux(),
		render:      render.New(),
	}

	addRoute := func(method, pattern string, h notifihttp.Handler) {
		cbh.mux.Method(method, pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := h.ServeHTTP(w, r)
			if err != nil {
				cbh.handleError(w, r, err)
			}
		}))
	}

	addRoute(http.MethodPost, "/", createCallback(cbh))
	addRoute(http.MethodGet, "/{callback_id}", getCallback(cbh))
	addRoute(http.MethodPost, "/{callback_id}/test", testCallback(cbh))

	return cbh
}

func (cbh *callbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	cbh.mux.ServeHTTP(w, r)
}

func (cbh *callbackHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	cbh.logger.Error().Err(err).Msg("callback handler error")

	apiErr := notifihttp.NewInternalServerError(err)
	if e, ok := err.(*notifihttp.Error); ok {
		apiErr = e
	}

	w.WriteHeader(apiErr.Status)
	err = json.NewEncoder(w).Encode(apiErr)
	if err != nil {
		cbh.logger.Error().Err(err).Msg("json encode")
	}
}

type createCbRequest struct {
	CallbackType callback.CBType `json:"callback_type"`
	URL          string          `json:"url"`
}

type createCbResponse struct {
	CallbackID callback.ID `json:"callback_id"`
}

func createCallback(cbh *callbackHandler) notifihttp.Handler {
	return notifihttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		token, ok := notifihttp.TokenFromCtx(r.Context())
		if !ok {
			return errors.New("token not injected into context")
		}

		var request createCbRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			return notifihttp.NewBadRequestError(err)
		}

		cbID, err := cbh.callbackSvc.CreateCallback(r.Context(), callback.Callback{
			ID:      callback.NewID(),
			TokenID: token.ID,
			CBType:  request.CallbackType,
			URL:     request.URL,
		})
		if err != nil {
			if err == callback.ErrCallbackExists {
				return notifihttp.NewBadRequestError(err)
			}

			return err
		}

		return cbh.render.JSON(w, http.StatusOK, createCbResponse{
			CallbackID: cbID,
		})
	})
}

type getCallbackResponse struct {
	ID     callback.ID     `json:"callback_id"`
	CBType callback.CBType `json:"callback_type"`
	URL    string          `json:"url"`
}

func getCallback(cbh *callbackHandler) notifihttp.Handler {
	return notifihttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		cbID := callback.ID(chi.URLParam(r, "callback_id"))
		cb, err := cbh.callbackSvc.GetCallback(r.Context(), cbID)
		if err != nil {
			if err == callback.ErrCallbackNotFound {
				return notifihttp.NewNotFoundError(err)
			}

			return errors.Wrap(err, "get callback")
		}

		response := getCallbackResponse{
			ID:     cb.ID,
			CBType: cb.CBType,
			URL:    cb.URL,
		}

		return cbh.render.JSON(w, http.StatusOK, response)
	})
}

func testCallback(cbh *callbackHandler) notifihttp.Handler {
	return notifihttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		cbID := callback.ID(chi.URLParam(r, "callback_id"))
		cb, err := cbh.callbackSvc.GetCallback(r.Context(), cbID)
		if err != nil {
			if err == callback.ErrCallbackNotFound {
				return notifihttp.NewNotFoundError(err)
			}

			return errors.Wrap(err, "get callback")
		}

		err = cbh.callbackSvc.TestCallback(r.Context(), cb.ID)
		if err != nil {
			cbh.logger.Error().Err(err).Msg("callback test fail")
			return cbh.render.JSON(w, http.StatusOK, map[string]interface{}{
				"message": "Callback test failed :(",
			})
		}

		return cbh.render.JSON(w, http.StatusOK, map[string]interface{}{
			"message": "Callback test success :)",
		})
	})
}
