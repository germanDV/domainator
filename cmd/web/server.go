package main

import (
	"domainator/internal/config"
	"domainator/internal/httphelp"
	"domainator/internal/tmpl"
	"encoding/json"
	"log/slog"
	"net/http"
	"runtime/debug"
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
	var revision string
	var dirty bool
	var lastCommit time.Time
	info, ok := debug.ReadBuildInfo()
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error reading build info"))
		return
	}

	for _, data := range info.Settings {
		switch data.Key {
		case "vcs.revision":
			revision = data.Value
		case "vcs.modified":
			dirty = true
		case "vcs.time":
			lastCommit, _ = time.Parse(time.RFC3339, data.Value)
		}
	}

	resp := map[string]any{
		"revision":   revision,
		"dirty":      dirty,
		"lastCommit": lastCommit,
		"go":         info.GoVersion,
	}

	w.WriteHeader(http.StatusOK)
	enc := json.NewEncoder(w)
	enc.Encode(resp)
}

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	httphelp.NotFound(w)
}
