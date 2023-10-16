package main

import (
	"domainator/internal/config"
	"domainator/internal/httphelp"
	"domainator/internal/tmpl"
	"log/slog"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

// buildServer builds the HTTP server, applying standard middleware and common/misc. routes.
// It returns the server and the router (to attach routes to).
func buildServer(addr string, logger *slog.Logger) (*http.Server, *httprouter.Router) {
	mux := httprouter.New()

	// Static files
	fs := http.FileServer(http.Dir("./ui/static/"))
	mux.Handler(http.MethodGet, "/static/*filepath", http.StripPrefix("/static", fs))

	// Misc. routes
	mux.Handler(http.MethodGet, "/", httphelp.Base.ThenFunc(homeHandler))
	mux.Handler(http.MethodGet, "/healthcheck", httphelp.Base.ThenFunc(healthcheckHandler))
	mux.NotFound = http.HandlerFunc(notFoundHandler)

	// Apply standard middleware
	handler := httphelp.Standard(logger).Then(mux)

	srv := &http.Server{
		Addr:         addr,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return srv, mux
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	expiryThreshold := config.GetDuration("CERT_EXPIRY_THRESHOLD")
	templateData := tmpl.BaseData(r)
	templateData["HealthcheckInterval"] = 15
	templateData["CertcheckInterval"] = 15
	templateData["ExpirationThreshold"] = expiryThreshold.Hours() / 24
	tmpl.RenderPage(w, http.StatusOK, "home.html.tmpl", &templateData)
}

func healthcheckHandler(w http.ResponseWriter, _ *http.Request) {
	// TODO: add git revision
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	httphelp.NotFound(w)
}
