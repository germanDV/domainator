package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/germandv/domainator/internal/cntxt"
)

func TestGetLanding_Authenticated(t *testing.T) {
	handler := GetLanding()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)

	userID := "018ec52b-dd69-7df4-b8e7-edcdc9a3a891"
	handler.ServeHTTP(w, cntxt.SetUserID(r, userID))

	if w.Code != 200 {
		t.Errorf("Want status code %d, got %d", 200, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Go To My Dashboard") {
		t.Errorf("Landing page does not have `Go To My Dashboard` link")
	}
}

func TestGetLanding_Unauthenticated(t *testing.T) {
	handler := GetLanding()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	handler.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("Want status code %d, got %d", 200, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Continue With GitHub") {
		t.Errorf("Landing page does not have `Continue With GitHub` link")
	}
}
