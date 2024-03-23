package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/germandv/domainator/internal/cache"
)

const RequestsPerMin = 50

func rateLimiterBuilder(logger *slog.Logger, cacheClient cache.Client) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := getKey(r.RemoteAddr)

			current, err := cacheClient.Increment(key)
			if err != nil {
				logger.Error("rate limiter error incrementing counter", "key", key, "err", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if current > RequestsPerMin {
				logger.Info("too many requests", "ip", r.RemoteAddr, "count", current)
				http.Error(w, "Too many requests", http.StatusTooManyRequests)
				return
			} else if current == 1 {
				// First request, set expiration to 1 minute
				err = cacheClient.Expire(key, time.Minute)
				if err != nil {
					logger.Error("rate limiter error setting expiration", "key", key, "err", err)
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func getKey(ip string) string {
	return fmt.Sprintf("request_count_%s", ip)
}
