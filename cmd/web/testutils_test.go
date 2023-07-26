package main

import (
	"domainator/internal/logger"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/form/v4"
	"github.com/go-playground/validator/v10"
)

type testServer struct {
	*httptest.Server
}

func newTestApp(t *testing.T) *application {
	return newTestAppWithLogger(t, io.Discard, io.Discard)
}

func newTestAppWithLogger(t *testing.T, infoOut, errOut io.Writer) *application {
	validate := validator.New()
	logit := logger.New(infoOut, errOut)

	templateCache, err := newTemplateCache()
	if err != nil {
		logit.Fatal(err)
	}

	fragmentCache, err := newFragmentsCache()
	if err != nil {
		logit.Fatal(err)
	}

	return &application{
		logit:         logit,
		templateCache: templateCache,
		fragmentCache: fragmentCache,
		formDecoder:   form.NewDecoder(),
		validate:      validate,
		// pingSvc:       pinger,
		// userSvc:       userSvc,
		// inspector:     inspector.New(db, pinger, logit),
		// mailer:        notifier.NewMailer(),
	}
}

func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)
	return &testServer{ts}
}

func (ts *testServer) get(t *testing.T, path string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + path)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	return rs.StatusCode, rs.Header, string(body)
}
