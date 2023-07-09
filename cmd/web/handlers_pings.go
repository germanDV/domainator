package main

import (
	"domainator/internal/services"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/julienschmidt/httprouter"
)

func (app *application) pings(w http.ResponseWriter, r *http.Request) {
	dummyUserID := uuid.New()

	pings, err := app.pingSvc.GetSummary(r.Context(), dummyUserID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData := map[string]any{
		"Year":  time.Now().Year(),
		"Pings": pings,
	}

	app.render(w, http.StatusOK, "pings.html.tmpl", &templateData)
}

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

	templateData := map[string]any{
		"Year":     time.Now().Year(),
		"Settings": settings,
		"Checks":   pingChecks,
	}

	app.render(w, http.StatusOK, "ping.html.tmpl", &templateData)
}

func (app *application) pingsNewForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "pings_new.html.tmpl", &map[string]any{"Year": time.Now().Year()})
}

func (app *application) pingsNew(w http.ResponseWriter, r *http.Request) {
	var payload services.PingCreate
	app.decodeForm(r, &payload)

	ok := app.pingSvc.Validate(&payload)
	if !ok {
		templateData := map[string]any{
			"Year": time.Now().Year(),
			"Form": payload,
		}
		app.render(w, http.StatusOK, "pings_new.html.tmpl", &templateData)
		return
	}

	app.pingSvc.SaveSettings(r.Context(), &payload)
	http.Redirect(w, r, "/pings", http.StatusSeeOther)
}

func (app *application) pingDelete(w http.ResponseWriter, r *http.Request) {
	idStr := httprouter.ParamsFromContext(r.Context()).ByName("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	app.pingSvc.DeleteSettingsByID(r.Context(), id)
	http.Redirect(w, r, "/pings", http.StatusSeeOther)
}
