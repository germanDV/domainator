package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/germandv/domainator/internal/cachemock"
)

func TestRateLimiter(t *testing.T) {
	t.Parallel()

	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	cacheClient := cachemock.New()
	limiter := rateLimiterBuilder(logger, cacheClient, 5)

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/", nil)

	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		limiter(handler).ServeHTTP(w, r)
		if w.Code != http.StatusOK {
			t.Errorf("Want status code %d, got %d", http.StatusOK, w.Code)
		}
	}

	w := httptest.NewRecorder()
	limiter(handler).ServeHTTP(w, r)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Want status code %d, got %d", http.StatusTooManyRequests, w.Code)
	}
}
