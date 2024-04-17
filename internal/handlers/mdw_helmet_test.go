package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHelmet(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	helmet(handler).ServeHTTP(w, r)

	want := "DENY"
	got := w.Header().Get("X-Frame-Options")
	if got != want {
		t.Errorf("Want X-Frame-Options %q, got %q", want, got)
	}

	want = "nosniff"
	got = w.Header().Get("X-Content-Type-Options")
	if got != want {
		t.Errorf("Want X-Content-Type-Options %q, got %q", want, got)
	}

	want = "max-age=31536000"
	got = w.Header().Get("Strict-Transport-Security")
	if got != want {
		t.Errorf("Want Strict-Transport-Security %q, got %q", want, got)
	}
}
