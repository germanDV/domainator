// Package certs holds the logic for the certs service.
package certs

import (
	"domainator/internal/httphelp"
	"domainator/internal/plans"
	"domainator/internal/tmpl"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// Controller is a controller that handles requests to the certs service.
type Controller struct {
	repo       Repo
	validator  *validator.Validate
	planGetter plans.Getter
	logger     *slog.Logger
}

// NewController returns a new certs controller.
func NewController(repo Repo, validate *validator.Validate, planGetter plans.Getter, logger *slog.Logger) *Controller {
	return &Controller{
		repo:       repo,
		validator:  validate,
		planGetter: planGetter,
		logger:     logger,
	}
}

// CertsNewForm renders the from to add a new Cert.
func (c *Controller) CertsNewForm(w http.ResponseWriter, r *http.Request) {
	templateData := tmpl.BaseData(r)
	tmpl.RenderPage(w, http.StatusOK, "certs_new.html.tmpl", &templateData)
}

// Save saves a new Cert, aka: a domain whose certificate will be checked.
func (c *Controller) Save(w http.ResponseWriter, r *http.Request) {
	var payload CreateCertReq
	httphelp.DecodeForm(r, &payload)

	ok := payload.Validate(c.validator)
	if !ok {
		templateData := tmpl.BaseData(r)
		templateData["Form"] = payload
		tmpl.RenderPage(w, http.StatusOK, "certs_new.html.tmpl", &templateData)
		return
	}

	userID := httphelp.GetUserIDFromCtx(w, r)
	count, err := c.repo.Count(r.Context(), userID)
	if err != nil {
		c.logger.Error(err.Error(), "handler", "Save", "trace", debug.Stack())
		httphelp.ServerError(w, err)
		return
	}

	planID := httphelp.GetPlanIDFromCtx(w, r)
	plan, err := c.planGetter(r.Context(), planID)
	if err != nil {
		c.logger.Error(err.Error(), "handler", "GetPlanIDFromCtx", "trace", debug.Stack())
		httphelp.ServerError(w, err)
		return
	}

	if count >= plan.CertsLimit {
		templateData := tmpl.BaseData(r)
		templateData["Form"] = payload
		templateData["Flash"] = fmt.Sprintf(
			`You have reached the limit for the %q plan. Please upgrade to track more domains.
			If this doesn't sound right, please log out and log back in.`,
			plan.Name,
		)
		tmpl.RenderPage(w, http.StatusOK, "certs_new.html.tmpl", &templateData)
		return
	}

	_, err = c.repo.Save(r.Context(), userID, &payload)
	if err != nil {
		c.logger.Error(err.Error(), "handler", "Save", "trace", debug.Stack())
		httphelp.ServerError(w, err)
		return
	}

	http.Redirect(w, r, "/certs", http.StatusSeeOther)
}

// GetSummary returns a summary of the Certs for the current user.
func (c *Controller) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID := httphelp.GetUserIDFromCtx(w, r)
	certs, err := c.repo.GetSummary(r.Context(), userID)
	if err != nil {
		c.logger.Error(err.Error(), "handler", "GetSummary", "trace", debug.Stack())
		httphelp.ServerError(w, err)
		return
	}

	templateData := tmpl.BaseData(r)
	templateData["Certs"] = certs
	tmpl.RenderPage(w, http.StatusOK, "certs.html.tmpl", &templateData)
}

// DeleteByID deletes a Cert.
func (c *Controller) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httphelp.ClientError(w, http.StatusBadRequest)
		return
	}

	userID := httphelp.GetUserIDFromCtx(w, r)
	err = c.repo.DeleteByID(r.Context(), id, userID)
	if err != nil {
		c.logger.Error(err.Error(), "handler", "DeleteByID", "trace", debug.Stack())
		httphelp.ServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
