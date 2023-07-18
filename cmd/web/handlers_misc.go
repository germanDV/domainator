package main

import (
	"net/http"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	templateData := initialTmplData(r)
	app.render(w, http.StatusOK, "home.html.tmpl", &templateData)
}
