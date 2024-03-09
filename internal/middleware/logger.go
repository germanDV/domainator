package middleware

import (
	"log/slog"
	"net/http"
)

func loggerBuilder(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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
}
