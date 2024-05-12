package handlers

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/germandv/domainator/internal/certs"
	"github.com/germandv/domainator/internal/cntxt"
	"github.com/germandv/domainator/internal/common"
	"github.com/germandv/domainator/internal/tlsermock"
	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

func TestMain(m *testing.M) {
	db = common.TestDB.GetPool()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestRegisterDomain(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	certsRepo := certs.NewRepo(db)
	certsService := certs.NewService(tlsermock.New(), certsRepo, 2)

	t.Run("register_new_domain", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("domain", "example.com")
		body := strings.NewReader(formData.Encode())

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/domain", body)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r = cntxt.SetUserID(r, "018ec52b-dd69-7df4-b8e7-edcdc9a3a891")

		handler := RegisterDomain(logger, certsService)
		handler.ServeHTTP(w, r)

		if w.Code != 200 {
			t.Errorf("Expected status code 200, got %d", w.Code)
		}

		resp := w.Body.String()
		if !strings.Contains(resp, `<th scope="row" class="w-250">example.com</th>`) {
			t.Errorf("Domain col not found in response: %s", resp)
		}
		if !strings.Contains(resp, `<td class="w-250">Test-Issuer</td>`) {
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

		handler := RegisterDomain(logger, certsService)
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

		handler := RegisterDomain(logger, certsService)
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

		handler := RegisterDomain(logger, certsService)
		handler.ServeHTTP(w, r)

		if w.Code != 400 {
			t.Errorf("Expected status code 400, got %d", w.Code)
		}
	})

	t.Run("exceed_domain_limit", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("domain", "third-domain.com")
		body := strings.NewReader(formData.Encode())

		w := httptest.NewRecorder()
		r := httptest.NewRequest("POST", "/domain", body)
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r = cntxt.SetUserID(r, "018ec52b-dd69-7df4-b8e7-edcdc9a3a891")

		handler := RegisterDomain(logger, certsService)
		handler.ServeHTTP(w, r)

		if w.Code != 400 {
			t.Errorf("Expected status code 400, got %d", w.Code)
		}

		resp := w.Body.String()
		if !strings.Contains(resp, "cannot have more than 2 certs") {
			t.Errorf("Expected error message 'cannot have more than 2 certs', got %s", resp)
		}
	})
}

func TestDeleteDomain(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	certsRepo := certs.NewRepo(db)
	certsService := certs.NewService(tlsermock.New(), certsRepo, 2)

	// Register a domain.
	formData := url.Values{}
	formData.Set("domain", "foo.bar")
	body := strings.NewReader(formData.Encode())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/domain", body)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r = cntxt.SetUserID(r, "018ec52b-dd69-7df4-b8e7-edcdc9a3a444")
	handler := RegisterDomain(logger, certsService)
	handler.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Expected 200 when registering, got %d", w.Code)
	}

	// Fetch domains.
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/", nil)
	r = cntxt.SetUserID(r, "018ec52b-dd69-7df4-b8e7-edcdc9a3a444")
	handler = GetDashboard(certsService)
	handler.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Expected 200 when fetching dashboard page, got %d", w.Code)
	}
	resp := w.Body.String()
	if !strings.Contains(resp, `<th scope="row" class="w-250">foo.bar</th>`) {
		t.Errorf("Domain not found in dashboard page: %s", resp)
	}

	// Find cert.
	cert, err := getCert(certsService, "foo.bar", "018ec52b-dd69-7df4-b8e7-edcdc9a3a444")
	if err != nil {
		t.Error(err)
	}

	// Delete domain.
	w = httptest.NewRecorder()
	r = httptest.NewRequest("DELETE", fmt.Sprintf("/domain/%s", cert.ID.String()), nil)
	r.SetPathValue("id", cert.ID.String())
	r = cntxt.SetUserID(r, "018ec52b-dd69-7df4-b8e7-edcdc9a3a444")
	handler = DeleteDomain(logger, certsService)
	handler.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Expected 200 when deleting domain, got %d", w.Code)
	}

	// Fetch domains.
	w = httptest.NewRecorder()
	r = httptest.NewRequest("GET", "/", nil)
	r = cntxt.SetUserID(r, "018ec52b-dd69-7df4-b8e7-edcdc9a3a444")
	handler = GetDashboard(certsService)
	handler.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Expected 200 when fetching dashboard page, got %d", w.Code)
	}
	resp = w.Body.String()
	if strings.Contains(resp, `<th scope="row" class="w-250">foo.bar</th>`) {
		t.Errorf("Dashboard should not longer contain deleted domain: %s", resp)
	}
}

func TestUpdateDomain(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	certsRepo := certs.NewRepo(db)
	certsService := certs.NewService(tlsermock.New(), certsRepo, 2)

	// Register a domain.
	formData := url.Values{}
	formData.Set("domain", "foobar.io")
	body := strings.NewReader(formData.Encode())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/domain", body)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r = cntxt.SetUserID(r, "018ec52b-dd69-7df4-b8e7-edcdc9a3a077")
	handler := RegisterDomain(logger, certsService)
	handler.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Expected 200 when registering, got %d", w.Code)
	}

	// Find created cert.
	certBefore, err := getCert(certsService, "foobar.io", "018ec52b-dd69-7df4-b8e7-edcdc9a3a077")
	if err != nil {
		t.Error(err)
	}

	// Update domain.
	w = httptest.NewRecorder()
	r = httptest.NewRequest("PUT", fmt.Sprintf("/domain/%s", certBefore.ID.String()), nil)
	r.SetPathValue("id", certBefore.ID.String())
	r = cntxt.SetUserID(r, "018ec52b-dd69-7df4-b8e7-edcdc9a3a077")
	handler = UpdateDomain(logger, certsService)
	handler.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Errorf("Expected 200 when updating domain, got %d", w.Code)
		t.Error(w.Body.String())
	}

	// Fetch updated cert.
	certAfter, err := getCert(certsService, "foobar.io", "018ec52b-dd69-7df4-b8e7-edcdc9a3a077")
	if err != nil {
		t.Error(err)
	}
	if certAfter.ID != certBefore.ID {
		t.Errorf("Domain ID should not have changed")
	}
	if certAfter.UpdatedAt.Before(certBefore.UpdatedAt) || certAfter.UpdatedAt.Equal(certBefore.UpdatedAt) {
		t.Errorf("Domain should have been updated. Before: %s, After: %s", certBefore.UpdatedAt, certAfter.UpdatedAt)
	}
}

func getCert(svc certs.Service, domain string, userID string) (*certs.Cert, error) {
	id, err := common.ParseID(userID)
	if err != nil {
		return nil, err
	}

	certs, err := svc.GetAll(context.Background(), certs.GetAllReq{UserID: id})
	if err != nil {
		return nil, err
	}

	for _, cert := range certs {
		if cert.Domain.String() == domain {
			return &cert, nil
		}
	}

	return nil, fmt.Errorf("could not find cert for domain %s", domain)
}
