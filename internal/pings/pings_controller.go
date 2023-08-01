// Package pings holds the logic for the pings service
package pings

import (
	"domainator/internal/httphelp"
	"domainator/internal/plans"
	"domainator/internal/tmpl"
	"domainator/internal/validation"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

// Controller is a controller that handles requests to the pings service
type Controller struct {
	repo      Repo
	validator *validator.Validate
	plansRepo plans.Repo
}

// NewController returns a new pings controller
func NewController(repo Repo, validate *validator.Validate, plansRepo plans.Repo) *Controller {
	return &Controller{
		repo:      repo,
		validator: validate,
		plansRepo: plansRepo,
	}
}

// GetSummary returns a summary of the pings for the current user
func (c *Controller) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID := httphelp.GetUserIDFromCtx(w, r)
	pings, err := c.repo.GetSummary(r.Context(), userID)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	templateData := tmpl.BaseData(r)
	templateData["Pings"] = pings
	tmpl.RenderPage(w, http.StatusOK, "pings.html.tmpl", &templateData)
}

// PingsNewForm renders the page for creating a new ping.
func (c *Controller) PingsNewForm(w http.ResponseWriter, r *http.Request) {
	templateData := tmpl.BaseData(r)
	tmpl.RenderPage(w, http.StatusOK, "pings_new.html.tmpl", &templateData)
}

// CreatePing creates a new ping and saves it to the database.
func (c *Controller) CreatePing(w http.ResponseWriter, r *http.Request) {
	var payload CreatePingReq
	httphelp.DecodeForm(r, &payload)

	ok := payload.Validate(c.validator)
	if !ok {
		templateData := tmpl.BaseData(r)
		templateData["Form"] = payload
		tmpl.RenderPage(w, http.StatusOK, "pings_new.html.tmpl", &templateData)
		return
	}

	userID := httphelp.GetUserIDFromCtx(w, r)
	count, err := c.repo.CountSettings(r.Context(), userID)
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

	if count >= plan.DomainsLimit {
		templateData := tmpl.BaseData(r)
		templateData["Form"] = payload
		templateData["Flash"] = fmt.Sprintf("You have reached the limit for the %q plan. Please upgrade to create more pings.", plan.Name)
		tmpl.RenderPage(w, http.StatusOK, "pings_new.html.tmpl", &templateData)
		return
	}

	c.repo.SaveSettings(r.Context(), userID, &payload)
	http.Redirect(w, r, "/pings", http.StatusSeeOther)
}

// GetPing returns the details for a ping.
func (c *Controller) GetPing(w http.ResponseWriter, r *http.Request) {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httphelp.ClientError(w, http.StatusBadRequest)
		return
	}

	userID := httphelp.GetUserIDFromCtx(w, r)
	settings, err := c.repo.GetSettingsByID(r.Context(), id, userID)
	if err != nil {
		if err == validation.ErrNotFound {
			httphelp.NotFound(w)
		} else {
			httphelp.ServerError(w, err)
		}
		return
	}

	pingChecks, err := c.repo.GetByID(r.Context(), id, userID)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	templateData := tmpl.BaseData(r)
	templateData["Settings"] = settings
	templateData["Checks"] = pingChecks
	tmpl.RenderPage(w, http.StatusOK, "ping.html.tmpl", &templateData)
}

func (c *Controller) DeleteByID(w http.ResponseWriter, r *http.Request) {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httphelp.ClientError(w, http.StatusBadRequest)
		return
	}

	userID := httphelp.GetUserIDFromCtx(w, r)
	c.repo.DeleteSettingsByID(r.Context(), id, userID)
	w.WriteHeader(http.StatusOK)
}
