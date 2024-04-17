package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContentTypeStatic(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/static/test.jpg", nil)
	contentType(handler).ServeHTTP(w, r)
	if w.Header().Get("Content-Type") != "" {
		t.Errorf("Content-Type header should be empty for static files")
	}
}

func TestContentTypeHealthcheck(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/healthcheck", nil)

	contentType(handler).ServeHTTP(w, r)

	want := "application/json; charset=utf-8"
	got := w.Header().Get("Content-Type")

	if got != want {
		t.Errorf("Want Content-Type %q, got %q", want, got)
	}
}

func TestContentTypeDefault(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	contentType(handler).ServeHTTP(w, r)

	want := "text/html; charset=utf-8"
	got := w.Header().Get("Content-Type")

	if want != got {
		t.Errorf("Want Content-Type %q, got %q", want, got)
	}
}
