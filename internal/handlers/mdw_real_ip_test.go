package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRealIP_TrueClientIP(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set(trueClientIP, "1.2.3.4")
	r.Header.Set(xRealIP, "1.2.3.5")
	r.Header.Set(xForwardedFor, "1.2.3.6")

	realIP(handler).ServeHTTP(w, r)

	if r.RemoteAddr != "1.2.3.4" {
		t.Errorf("Want RemoteAddr %s, got %s", "1.2.3.4", r.RemoteAddr)
	}
}

func TestRealIP_XRealIP(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set(xRealIP, "1.2.3.5")
	r.Header.Set(xForwardedFor, "1.2.3.6")

	realIP(handler).ServeHTTP(w, r)

	if r.RemoteAddr != "1.2.3.5" {
		t.Errorf("Want RemoteAddr %s, got %s", "1.2.3.5", r.RemoteAddr)
	}
}

func TestRealIP_XForwardedFor(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set(xForwardedFor, "1.2.3.6")

	realIP(handler).ServeHTTP(w, r)

	if r.RemoteAddr != "1.2.3.6" {
		t.Errorf("Want RemoteAddr %s, got %s", "1.2.3.6", r.RemoteAddr)
	}
}

func TestRealIP_XForwardedForMultiple(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set(xForwardedFor, "1.2.3.6, 1.2.3.7, 1.2.3.8")

	realIP(handler).ServeHTTP(w, r)

	if r.RemoteAddr != "1.2.3.6" {
		t.Errorf("Want RemoteAddr %s, got %s", "1.2.3.6", r.RemoteAddr)
	}
}

func TestRealIP_NoHeaders(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	before := r.RemoteAddr
	realIP(handler).ServeHTTP(w, r)

	if r.RemoteAddr != before {
		t.Errorf("Want middleware not to modify RemoteAddr, before: %s, after %s", before, r.RemoteAddr)
	}
}
