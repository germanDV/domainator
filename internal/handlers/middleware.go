package handlers

import (
	"log/slog"
	"net/http"

	"github.com/germandv/domainator/internal/cache"
)

func CommonBuilder(logger *slog.Logger, cacheClient cache.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		reqLogger := loggerBuilder(logger)
		rateLimiter := rateLimiterBuilder(logger, cacheClient)
		recoverer := recoverBuilder(logger)

		return realIP(reqLogger(rateLimiter(recoverer(helmet(csp(contentType(next)))))))
	}
}
