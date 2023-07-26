package main

import (
	"net/http"
	"strings"
	"testing"
)

func TestHomePage(t *testing.T) {
	app := newTestApp(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/")
	if code != http.StatusOK {
		t.Errorf("want status code %d; got %d", http.StatusOK, code)
	}

	doctype := "<!DOCTYPE html>"
	if !strings.Contains(body, doctype) {
		t.Errorf("want resp to contain %s", doctype)
	}

	title := "<title>Domainator</title>"
	if !strings.Contains(body, title) {
		t.Errorf("want resp to contain %s", title)
	}
}

func TestHealthcheck(t *testing.T) {
	app := newTestApp(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/healthcheck")
	if code != http.StatusOK {
		t.Errorf("want status code %d; got %d", http.StatusOK, code)
	}
	if body != "OK" {
		t.Errorf("want body %q; got OK", body)
	}
}
