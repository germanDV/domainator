package main

import "net/http"

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	app.logit.Info("User logging out")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
