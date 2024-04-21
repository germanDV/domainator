package handlers

import (
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/germandv/domainator/internal/certs"
	"github.com/germandv/domainator/internal/cntxt"
	"github.com/germandv/domainator/internal/common"
	"github.com/germandv/domainator/internal/tlser_mock"
)

func TestRegisterDomain(t *testing.T) {
	db := common.TestDB.GetPool()
	certsRepo := certs.NewRepo(db)
	certsService := certs.NewService(tlser_mock.New(), certsRepo)

	t.Run("register_new_domain", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("domain", "example.com")
		body := strings.NewReader(formData.Encode())

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/domain", body)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r = cntxt.SetUserID(r, "018ec52b-dd69-7df4-b8e7-edcdc9a3a891")

		handler := RegisterDomain(certsService)
		handler.ServeHTTP(w, r)

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		resp := w.Body.String()
		if !strings.Contains(resp, `<th scope="row">example.com</th>`) {
			t.Errorf("Domain col not found in response: %s", resp)
		}
		if !strings.Contains(resp, `<td>Test-Issuer</td>`) {
			t.Errorf("Issuer col not found in response: %s", resp)
		}
		if !strings.Contains(resp, `<td><span class="chip">Expires in 29 days</span></td>`) {
			t.Errorf("Expiration col not included in response: %s", resp)
		}
	})

	t.Run("register_duplicate_domain", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("domain", "example.com")
		body := strings.NewReader(formData.Encode())

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/domain", body)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r = cntxt.SetUserID(r, "018ec52b-dd69-7df4-b8e7-edcdc9a3a891")

		handler := RegisterDomain(certsService)
		handler.ServeHTTP(w, r)

		if w.Code != 400 {
			t.Errorf("Expected status code 400, got %d", w.Code)
		}
	})

	t.Run("register_expired_domain", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("domain", "expired.com")
		body := strings.NewReader(formData.Encode())

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/domain", body)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r = cntxt.SetUserID(r, "018ec52b-dd69-7df4-b8e7-edcdc9a3a891")

		handler := RegisterDomain(certsService)
		handler.ServeHTTP(w, r)

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		resp := w.Body.String()
		if !strings.Contains(resp, `<td><span class="chip error-text">Expired</span></td>`) {
			t.Errorf("Expiration col not included in response: %s", resp)
		}
	})

	t.Run("register_bad_domain", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("domain", "notconnect.com")
		body := strings.NewReader(formData.Encode())

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/domain", body)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r = cntxt.SetUserID(r, "018ec52b-dd69-7df4-b8e7-edcdc9a3a891")

		handler := RegisterDomain(certsService)
		handler.ServeHTTP(w, r)

		if w.Code != 400 {
			t.Errorf("Expected status code 400, got %d", w.Code)
		}
	})
}
