package main

import "net/http"

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	app.infoLog.Println("User logging out")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
