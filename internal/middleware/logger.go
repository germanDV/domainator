package middleware

import (
	"log/slog"
	"net/http"
	"os"
)

func Logger(next http.Handler) http.Handler {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info(
			"Serving Request",
			"method", r.Method,
			"path", r.URL.Path,
			"query", r.URL.RawQuery,
			"ip", r.RemoteAddr,
		)
		next.ServeHTTP(w, r)
	})
}
