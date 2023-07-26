package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecureHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	secureHeaders(next).ServeHTTP(rr, r)
	rs := rr.Result()

	want := "default-src 'self'; style-src 'self' 'unsafe-inline'; font-src 'self'"
	got := rs.Header.Get("Content-Security-Policy")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = "origin-when-cross-origin"
	got = rs.Header.Get("Referrer-Policy")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = "nosniff"
	got = rs.Header.Get("X-Content-Type-Options")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = "deny"
	got = rs.Header.Get("X-Frame-Options")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = "0"
	got = rs.Header.Get("X-XSS-Protection")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	bytes.TrimSpace(body)
	want = "OK"
	got = string(body)
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
