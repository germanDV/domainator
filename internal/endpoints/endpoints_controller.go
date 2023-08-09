// Package endpoints holds the logic for the endpoints service.
package endpoints

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

// Controller is a controller that handles requests to the endpoints service.
type Controller struct {
	repo      Repo
	validator *validator.Validate
	plansRepo plans.Repo
}

// NewController returns a new endpoints controller.
func NewController(repo Repo, validate *validator.Validate, plansRepo plans.Repo) *Controller {
	return &Controller{
		repo:      repo,
		validator: validate,
		plansRepo: plansRepo,
	}
}

// GetSummary returns a summary of the endpoints and healthchecks for the current user.
func (c *Controller) GetSummary(w http.ResponseWriter, r *http.Request) {
	userID := httphelp.GetUserIDFromCtx(w, r)
	endpoints, err := c.repo.GetSummary(r.Context(), userID)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	templateData := tmpl.BaseData(r)
	templateData["Endpoints"] = endpoints
	tmpl.RenderPage(w, http.StatusOK, "endpoints.html.tmpl", &templateData)
}

// EndpointNewForm renders the page for creating a new Endpoint.
func (c *Controller) EndpointNewForm(w http.ResponseWriter, r *http.Request) {
	templateData := tmpl.BaseData(r)
	tmpl.RenderPage(w, http.StatusOK, "endpoints_new.html.tmpl", &templateData)
}

// CreateEndpoint creates a new Endpoint and saves it to the database.
func (c *Controller) CreateEndpoint(w http.ResponseWriter, r *http.Request) {
	var payload CreateEndpointReq
	httphelp.DecodeForm(r, &payload)

	ok := payload.Validate(c.validator)
	if !ok {
		templateData := tmpl.BaseData(r)
		templateData["Form"] = payload
		tmpl.RenderPage(w, http.StatusOK, "endpoints_new.html.tmpl", &templateData)
		return
	}

	userID := httphelp.GetUserIDFromCtx(w, r)
	count, err := c.repo.Count(r.Context(), userID)
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
		templateData["Flash"] = fmt.Sprintf(
			`You have reached the limit for the %q plan. Please upgrade to add more endpoints.
			If this doesn't sound right, please log out and log back in.`,
			plan.Name,
		)
		tmpl.RenderPage(w, http.StatusOK, "endpoints_new.html.tmpl", &templateData)
		return
	}

	_, err = c.repo.Save(r.Context(), userID, &payload)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	http.Redirect(w, r, "/endpoints", http.StatusSeeOther)
}

// GetEndpoint returns the details and Healthchecks of an Endpoint.
func (c *Controller) GetEndpoint(w http.ResponseWriter, r *http.Request) {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		httphelp.ClientError(w, http.StatusBadRequest)
		return
	}

	userID := httphelp.GetUserIDFromCtx(w, r)
	endpoint, err := c.repo.GetByID(r.Context(), id, userID)
	if err != nil {
		if err == validation.ErrNotFound {
			httphelp.NotFound(w)
		} else {
			httphelp.ServerError(w, err)
		}
		return
	}

	Healthchecks, err := c.repo.GetHealthcheckByID(r.Context(), id, userID)
	if err != nil {
		httphelp.ServerError(w, err)
		return
	}

	templateData := tmpl.BaseData(r)
	templateData["Endpoint"] = endpoint
	templateData["Healthchecks"] = Healthchecks
	tmpl.RenderPage(w, http.StatusOK, "endpoint.html.tmpl", &templateData)
}

// DeleteByID deletes an Endpoint.
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
