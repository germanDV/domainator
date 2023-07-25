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

	base := alice.New(app.authenticate)
	protected := base.Append(app.requireAuth)
	standard := alice.New(app.recoverPanic, app.logRequest, secureHeaders)

	mux.Handler(http.MethodGet, "/", base.ThenFunc(app.home))
	mux.Handler(http.MethodGet, "/pings", protected.ThenFunc(app.pings))
	mux.Handler(http.MethodGet, "/pings-new", protected.ThenFunc(app.pingsNewForm))
	mux.Handler(http.MethodPost, "/pings-new", protected.ThenFunc(app.pingsNew))
	mux.Handler(http.MethodGet, "/pings/:id", protected.ThenFunc(app.ping))
	mux.Handler(http.MethodDelete, "/pings/:id", protected.ThenFunc(app.pingDelete))
	mux.Handler(http.MethodGet, "/user/signup", base.ThenFunc(app.signupForm))
	mux.Handler(http.MethodPost, "/user/signup", base.ThenFunc(app.signup))
	mux.Handler(http.MethodGet, "/user/login", base.ThenFunc(app.loginForm))
	mux.Handler(http.MethodPost, "/user/login", base.ThenFunc(app.login))
	mux.Handler(http.MethodPost, "/user/logout", protected.ThenFunc(app.logout))
	mux.Handler(http.MethodGet, "/user/verify", base.ThenFunc(app.verifyForm))
	mux.Handler(http.MethodPost, "/user/verify", base.ThenFunc(app.verify))
	mux.Handler(http.MethodGet, "/settings", protected.ThenFunc(app.settings))
	mux.Handler(http.MethodPut, "/settings/email/:id", protected.ThenFunc(app.settingsUpsertEmail))
	mux.Handler(http.MethodPut, "/settings/toggle/:id", protected.ThenFunc(app.settingsToggle))

	return standard.Then(mux)
}
