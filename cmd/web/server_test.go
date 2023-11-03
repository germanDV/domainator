package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthcheckHandler(t *testing.T) {
	rr := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	healthcheckHandler(rr, r)
	rs := rr.Result()

	want := http.StatusOK
	if rs.StatusCode != want {
		t.Errorf("want %d, got %d", want, rs.StatusCode)
	}

	defer rs.Body.Close()
	_, err = io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNotFoundHandler(t *testing.T) {
	rr := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	notFoundHandler(rr, r)
	rs := rr.Result()

	want := http.StatusNotFound
	if rs.StatusCode != want {
		t.Errorf("want %d, got %d", want, rs.StatusCode)
	}
}

func TestHomeHandler(t *testing.T) {
	rr := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	homeHandler(rr, r)
	rs := rr.Result()

	want := http.StatusOK
	if rs.StatusCode != want {
		t.Errorf("want %d, got %d", want, rs.StatusCode)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	doctype := "<!DOCTYPE html>"
	if !strings.Contains(string(body), doctype) {
		t.Errorf("want resp to contain %s", doctype)
	}

	title := "<title>Domainator</title>"
	if !strings.Contains(string(body), title) {
		t.Errorf("want resp to contain %s", title)
	}
}
