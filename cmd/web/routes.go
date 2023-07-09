package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := httprouter.New()

	fs := http.FileServer(http.Dir("./ui/static/"))
	mux.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fs))

	mux.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	mux.HandlerFunc(http.MethodGet, "/", app.home)

	mux.HandlerFunc(http.MethodGet, "/pings", app.pings)
	mux.HandlerFunc(http.MethodGet, "/pings-new", app.pingsNewForm)
	mux.HandlerFunc(http.MethodPost, "/pings-new", app.pingsNew)
	mux.HandlerFunc(http.MethodGet, "/pings/:id", app.ping)
	mux.HandlerFunc(http.MethodPost, "/pings-delete/:id", app.pingDelete)

	mux.HandlerFunc(http.MethodPost, "/user/logout", app.logout)

	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	return standard.Then(mux)
}
