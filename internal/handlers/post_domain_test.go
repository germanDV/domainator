package handlers

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/germandv/domainator/internal/certs"
	"github.com/germandv/domainator/internal/cntxt"
	"github.com/germandv/domainator/internal/db"
	"github.com/germandv/domainator/internal/tlser"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TODO: improve this mock and move to its own package (tlser_mock)
type MockTLSer struct{}

func (m MockTLSer) GetCertData(_ string) tlser.CertData {
	return tlser.CertData{
		Status: "OK",
		Expiry: time.Now().Add(24 * 30 * time.Hour),
		Issuer: "Test-Issuer",
	}
}

func TestRegisterDomain(t *testing.T) {
	db, _, err := NewTestDB()
	if err != nil {
		t.Fatal(err)
	}

	certsRepo := certs.NewRepo(db)
	certsService := certs.NewService(MockTLSer{}, certsRepo)

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
}

// TODO: make it a singleton to ensure it runs only once even for parallel tests
func NewTestDB() (*pgxpool.Pool, func(), error) {
	ctx := context.Background()
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16"),
		postgres.WithDatabase("domainator"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to strat posgres container: %w", err)
	}

	conn, err := pgContainer.ConnectionString(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get postgres connection string: %w", err)
	}

	fmt.Printf("Postgres Test Container ConnectionString: %s\n", conn)
	dbPool, err := db.InitWithConnStr(conn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to init postgres: %w", err)
	}

	err = dbPool.Ping(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	migrator, err := db.NewDbMigrator(conn, os.DirFS(filepath.Join("..", "..", "migrations")))
	if err != nil {
		return nil, nil, fmt.Errorf("failed create DB Migrator: %w", err)
	}
	err = migrator.Up(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("failed run migrations: %w", err)
	}

	terminate := func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			panic("failed to terminate container: " + err.Error())
		}
	}

	return dbPool, terminate, nil
}
