package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/germandv/domainator/internal/cache_mock"
)

type MockDbPinger struct{}

func (m MockDbPinger) Ping(_ context.Context) error {
	return nil
}

func TestGetHealthcheck(t *testing.T) {
	handler := GetHealthcheck(cache_mock.New(), MockDbPinger{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/healthcheck", nil)
	handler.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("Want status code %d, got %d", 200, w.Code)
	}

	resp := map[string]any{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("Error unmarshalling response: %s", err)
	}

	goVersion := "go1.22.3"
	if resp["go"] != goVersion {
		t.Errorf("Want go version %s, got %s", goVersion, resp["go"])
	}
}

func TestGetHealthcheck_Deep(t *testing.T) {
	handler := GetHealthcheck(cache_mock.New(), MockDbPinger{})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/healthcheck?deep=true", nil)
	handler.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("Want status code %d, got %d", 200, w.Code)
	}

	resp := map[string]any{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		t.Errorf("Error unmarshalling response: %s", err)
	}

	if resp["redis"] != "up" {
		t.Errorf("Want redis to be up, got %s", resp["redis"])
	}

	if resp["postgres"] != "up" {
		t.Errorf("Want postgres to be up, got %s", resp["postgres"])
	}
}
