package main

import (
	"domainator/internal/httphelp"
	"domainator/internal/logger"
	"domainator/internal/tmpl"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

// buildServer builds the HTTP server, applying standard middleware and common/misc. routes.
// It returns the server and the router (to attach routes to).
func buildServer(addr string) (*http.Server, *httprouter.Router) {
	mux := httprouter.New()

	// Static files
	fs := http.FileServer(http.Dir("./ui/static/"))
	mux.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fs))

	// Misc. routes
	mux.Handler(http.MethodGet, "/", httphelp.Base.ThenFunc(homeHandler))
	mux.Handler(http.MethodGet, "/healthcheck", httphelp.Base.ThenFunc(healthcheckHandler))
	mux.NotFound = http.HandlerFunc(notFoundHandler)

	// Apply standard middleware
	handler := httphelp.Standard.Then(mux)

	srv := &http.Server{
		Addr:         addr,
		ErrorLog:     logger.Writer.ErrorLog,
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return srv, mux
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	templateData := tmpl.BaseData(r)
	tmpl.RenderPage(w, http.StatusOK, "home.html.tmpl", &templateData)
}

func healthcheckHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	httphelp.NotFound(w)
}
