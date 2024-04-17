package handlers

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/germandv/domainator/internal/cntxt"
)

func TestGetAccess_Redirect(t *testing.T) {
	handler := GetAccess()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/login", nil)

	userID := "018ec52b-dd69-7df4-b8e7-edcdc9a3a891"
	handler.ServeHTTP(w, cntxt.SetUserID(r, userID))

	if w.Code != 303 {
		t.Errorf("Should redirect logged in user, got status code %d", w.Code)
	}
}

func TestGetAccess(t *testing.T) {
	handler := GetAccess()
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/login", nil)
	handler.ServeHTTP(w, r)

	if w.Code != 200 {
		t.Errorf("Want status code %d, got %d", 200, w.Code)
	}

	if !strings.Contains(w.Body.String(), "<title>Domainator | Login</title>") {
		t.Errorf("Login page does not have the expected title")
	}
}
