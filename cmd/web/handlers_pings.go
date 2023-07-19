package main

import (
	"domainator/internal/services"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

func (app *application) pings(w http.ResponseWriter, r *http.Request) {
	userID := app.GetUserIDFromCtx(w, r)
	if userID == uuid.Nil {
		app.clientError(w, http.StatusUnauthorized)
		return
	}

	pings, err := app.pingSvc.GetSummary(r.Context(), userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData := initialTmplData(r)
	templateData["Pings"] = pings
	app.render(w, http.StatusOK, "pings.html.tmpl", &templateData)
}

// TODO: only allow the user to see their own pings
func (app *application) ping(w http.ResponseWriter, r *http.Request) {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	settings, err := app.pingSvc.GetSettingsByID(r.Context(), id)
	if err != nil {
		if err == services.ErrNotFound {
			app.notFound(w)
		} else {
			app.serverError(w, err)
		}
		return
	}

	pingChecks, err := app.pingSvc.GetChecksByID(r.Context(), id)
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData := initialTmplData(r)
	templateData["Settings"] = settings
	templateData["Checks"] = pingChecks
	app.render(w, http.StatusOK, "ping.html.tmpl", &templateData)
}

func (app *application) pingsNewForm(w http.ResponseWriter, r *http.Request) {
	templateData := initialTmplData(r)
	app.render(w, http.StatusOK, "pings_new.html.tmpl", &templateData)
}

func (app *application) pingsNew(w http.ResponseWriter, r *http.Request) {
	var payload services.PingCreate
	app.decodeForm(r, &payload)

	ok := app.pingSvc.Validate(&payload)
	if !ok {
		templateData := initialTmplData(r)
		templateData["Form"] = payload
		app.render(w, http.StatusOK, "pings_new.html.tmpl", &templateData)
		return
	}

	userID := app.GetUserIDFromCtx(w, r)
	if userID == uuid.Nil {
		app.serverError(w, errors.New("Missing user ID in context"))
		return
	}

	app.pingSvc.SaveSettings(r.Context(), userID, &payload)
	http.Redirect(w, r, "/pings", http.StatusSeeOther)
}

// TODO: only allow the user to delete their own pings
func (app *application) pingDelete(w http.ResponseWriter, r *http.Request) {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	app.pingSvc.DeleteSettingsByID(r.Context(), id)
	w.WriteHeader(http.StatusOK)
}
