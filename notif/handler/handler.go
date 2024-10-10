package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/unrolled/render"

	"github.com/stevenferrer/notifi/callback"
	"github.com/stevenferrer/notifi/notif"
	"github.com/stevenferrer/notifi/notifihttp"
	"github.com/stevenferrer/notifi/token"
)

type notifHandler struct {
	notifSvc notif.Service
	mux      *chi.Mux
	render   *render.Render
	logger   zerolog.Logger
}

func NewNotifHandler(notifSvc notif.Service, logger zerolog.Logger) http.Handler {
	nth := &notifHandler{
		notifSvc: notifSvc,
		mux:      chi.NewMux(),
		render:   render.New(),
		logger:   logger,
	}

	// helper for adding http route
	addRoute := func(method, pattern string, h notifihttp.Handler) {
		nth.mux.Method(method, pattern, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := h.ServeHTTP(w, r)
			if err != nil {
				nth.handleError(w, r, err)
			}
		}))
	}

	addRoute(http.MethodPost, "/", createNotif(nth))
	// addRoute(http.MethodGet, "/", getNotifs(nth))
	addRoute(http.MethodGet, "/{notif_id}", getNotif(nth))
	addRoute(http.MethodPost, "/{notif_id}/resend", resendNotif(nth))

	return nth
}

func (nth *notifHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	nth.mux.ServeHTTP(w, r)
}

func (nth *notifHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	nth.logger.Error().Err(err).Msg("notification handler error")

	apiErr := notifihttp.NewInternalServerError(err)
	if e, ok := err.(*notifihttp.Error); ok {
		apiErr = e
	}

	w.WriteHeader(apiErr.Status)
	err = json.NewEncoder(w).Encode(apiErr)
	if err != nil {
		nth.logger.Error().Err(err).Msg("json encode")
	}
}

type createNotifRequest struct {
	DestTokenID token.ID               `json:"dest_token_id"`
	CBType      callback.CBType        `json:"callback_type"`
	Payload     map[string]interface{} `json:"payload"`
}

type createNotifResponse struct {
	NotifID notif.ID `json:"notification_id"`
}

func createNotif(nth *notifHandler) notifihttp.Handler {
	return notifihttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		token, ok := notifihttp.TokenFromCtx(r.Context())
		if !ok {
			return errors.New("token not injected into context")
		}

		var request createNotifRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			return notifihttp.NewBadRequestError(err)
		}

		notifID, err := nth.notifSvc.CreateNotif(r.Context(), notif.Notif{
			SrcTokenID:  token.ID,
			DestTokenID: request.DestTokenID,
			CBType:      request.CBType,
			Payload:     request.Payload,
		})
		if err != nil {
			return errors.Wrap(err, "create notif")
		}

		return nth.render.JSON(w, http.StatusOK, createNotifResponse{
			NotifID: notifID,
		})
	})
}

type getNotifResponse struct {
	ID      notif.ID               `json:"notification_id"`
	CBType  callback.CBType        `json:"callback_type"`
	Status  notif.Status           `json:"status"`
	Payload map[string]interface{} `json:"payload"`
}

func getNotif(nth *notifHandler) notifihttp.Handler {
	return notifihttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		notifID := notif.ID(chi.URLParam(r, "notif_id"))
		nf, err := nth.notifSvc.GetNotif(r.Context(), notifID)
		if err != nil {
			if err == notif.ErrNotifNotFound {
				return notifihttp.NewNotFoundError(err)
			}

			return errors.Wrap(err, "get notif")
		}

		response := getNotifResponse{
			ID:      nf.ID,
			CBType:  nf.CBType,
			Status:  nf.Status,
			Payload: nf.Payload,
		}

		return nth.render.JSON(w, http.StatusOK, response)
	})
}

func resendNotif(nth *notifHandler) notifihttp.Handler {
	return notifihttp.HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		notifID := notif.ID(chi.URLParam(r, "notif_id"))
		err := nth.notifSvc.ResendNotif(r.Context(), notifID)
		if err != nil {
			return err
		}

		return nth.render.JSON(w, http.StatusOK, map[string]string{
			"message": "resend ok",
		})
	})
}
