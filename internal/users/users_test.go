package users

import (
	"context"
	"domainator/internal/config"
	"domainator/internal/db"
	"domainator/internal/notifier"
	"domainator/internal/plans"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	dbpool     *pgxpool.Pool
	controller *Controller
)

func TestMain(t *testing.M) {
	setup()
	code := t.Run()
	teardown()
	os.Exit(code)
}

func setup() {
	config.LoadEnv()
	dbpool = db.MustInit(config.GetString("DSN"))
	repo := NewPostgresRepo(dbpool)

	controller = &Controller{
		repo:       repo,
		validator:  validator.New(),
		mailer:     notifier.NewMailer(),
		planGetter: mockPlanGetter,
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
	}

	u, err := newUser("tester_one@test.io", "tester1password")
	if err != nil {
		panic(err)
	}
	repo.Save(context.Background(), u)
}

func teardown() {
	dbpool.Close()
}

func mockPlanGetter(_ context.Context, _ int) (*plans.Plan, error) {
	return &plans.Plan{
		ID:           2,
		Name:         "Pro",
		Price:        500,
		DomainsLimit: 100,
		CertsLimit:   100,
	}, nil
}

func TestLogin(t *testing.T) {
	t.Run("good_credentials", func(t *testing.T) {
		w := httptest.NewRecorder()

		r, err := http.NewRequest(http.MethodPost, "/user/login", nil)
		if err != nil {
			t.Error(err)
		}
		form := url.Values{}
		form.Set("email", "tester_one@test.io")
		form.Set("password", "tester1password")
		r.PostForm = form
		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		controller.Login(w, r)
		resp := w.Result()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("want status code %d, got %d", http.StatusOK, resp.StatusCode)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		wantMsg := "activate your account in order to continue"
		if !strings.Contains(string(body), wantMsg) {
			t.Errorf("body does not include %q, got: %s", wantMsg, string(body))
		}
	})
}
