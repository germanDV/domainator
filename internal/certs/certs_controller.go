// Package certs holds the logic for the certs service.
package certs

import (
	"domainator/internal/httphelp"
	"domainator/internal/plans"
	"domainator/internal/tmpl"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// Controller is a controller that handles requests to the certs service
type Controller struct {
	repo      Repo
	validator *validator.Validate
	plansRepo plans.Repo
}

// NewController returns a new certs controller
func NewController(repo Repo, validate *validator.Validate, plansRepo plans.Repo) *Controller {
	return &Controller{
		repo:      repo,
		validator: validate,
		plansRepo: plansRepo,
	}
}

// CertsNewForm renders the from to add a new domain.
func (c *Controller) CertsNewForm(w http.ResponseWriter, r *http.Request) {
	templateData := tmpl.BaseData(r)
	tmpl.RenderPage(w, http.StatusOK, "certs_new.html.tmpl", &templateData)
}

// SaveDomain saves a new domain for which certificates will be checked.
func (c *Controller) SaveDomain(w http.ResponseWriter, r *http.Request) {
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
	count, err := c.repo.CountDomains(r.Context(), userID)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	planID := httphelp.GetPlanIDFromCtx(w, r)
	plan, err := c.plansRepo.GetByID(r.Context(), planID)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	if count >= plan.CertsLimit {
		templateData := tmpl.BaseData(r)
		templateData["Form"] = payload
		templateData["Flash"] = fmt.Sprintf("You have reached the limit for the %q plan. Please upgrade to track more domains.", plan.Name)
		tmpl.RenderPage(w, http.StatusOK, "certs_new.html.tmpl", &templateData)
		return
	}

	_, err = c.repo.Save(r.Context(), userID, &payload)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	http.Redirect(w, r, "/certs", http.StatusSeeOther)
}

// GetSummary returns a summary of the certificates for the current user.
func (c *Controller) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID := httphelp.GetUserIDFromCtx(w, r)
	certs, err := c.repo.GetSummary(r.Context(), userID)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	templateData := tmpl.BaseData(r)
	templateData["Certs"] = certs
	tmpl.RenderPage(w, http.StatusOK, "certs.html.tmpl", &templateData)
}

// DeleteByID deletes a domain whose certificate no longer needs to be checked.
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
		httphelp.ServerError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
