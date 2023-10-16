// Package events holds the logic for the events service.
package events

import (
	"domainator/internal/httphelp"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-playground/validator/v10"
)

// Controller is a controller that handles requests to the events service
type Controller struct {
	repo      Repo
	validator *validator.Validate
	logger    *slog.Logger
}

// NewController returns a new events controller.
func NewController(repo Repo, validate *validator.Validate, logger *slog.Logger) *Controller {
	return &Controller{
		repo:      repo,
		validator: validate,
		logger:    logger,
	}
}

// Save stores an event in the database.
func (c *Controller) Save(w http.ResponseWriter, r *http.Request) {
	var payload CreateEventReq
	httphelp.DecodeForm(r, &payload)

	ok := payload.Validate(c.validator)
	if !ok {
		httphelp.ClientError(w, http.StatusBadRequest)
		return
	}

	userID := httphelp.GetUserIDFromCtx(w, r)
	_, err := c.repo.Save(r.Context(), userID, &payload)
	if err != nil {
		c.logger.Error(err.Error(), "handler", "Save", "trace", debug.Stack())
		httphelp.ServerError(w, err)
		return
	}

	w.Write([]byte(`<div class="center">Thanks for letting us know!</div>`))
}
