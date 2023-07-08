package main

import (
	"net/http"
	"time"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	app.render(w, http.StatusOK, "home.html.tmpl", &map[string]any{"Year": time.Now().Year()})
}
