package handlers

import (
	"log/slog"
	"net/http"

	"github.com/germandv/domainator/internal/cache"
)

const RequestsPerMin int64 = 50

func CommonMdwBuilder(logger *slog.Logger, cacheClient cache.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		reqLogger := loggerBuilder(logger)
		rateLimiter := rateLimiterBuilder(logger, cacheClient, RequestsPerMin)
		recoverer := recoverBuilder(logger)

		return realIP(reqLogger(rateLimiter(recoverer(helmet(csp(contentType(next)))))))
	}
}
