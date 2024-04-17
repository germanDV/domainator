package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCsp(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	csp(handler).ServeHTTP(w, r)

	want := "default-src 'self'; script-src 'self'; style-src 'self' 'sha256-pgn1TCGZX6O77zDvy0oTODMOxemn0oj0LeCnQTRj7Kg='; frame-ancestors 'none'; form-action 'self'"
	got := w.Header().Get("Content-Security-Policy")

	if got != want {
		t.Errorf("Want Content-Security-Policy %q, got %q", want, got)
	}
}
