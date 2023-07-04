package main

import (
	"domainator/internal/services"
	"net/http"
	"time"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "home.html", &map[string]any{"Year": time.Now().Year()})
}

func (app *application) pings(w http.ResponseWriter, r *http.Request) {
	dummyUserID := "00000000-0000-0000-0000-000000000000"

	pings, err := app.pingSvc.GetSummary(dummyUserID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	templateData := map[string]any{
		"Year":  time.Now().Year(),
		"Pings": pings,
	}

	app.render(w, http.StatusOK, "pings.html", &templateData)
}

func (app *application) pingsNewForm(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "pings_new.html", &map[string]any{"Year": time.Now().Year()})
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
		app.render(w, http.StatusOK, "pings_new.html", &templateData)
		return
	}

	// TODO: save payload to database
	http.Redirect(w, r, "/pings", http.StatusSeeOther)
}
