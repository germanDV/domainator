package httphelp

import (
	"bytes"
	"domainator/internal/logger"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

var (
	infoLogs = io.Discard
	errLogs  = new(bytes.Buffer)
)

func TestMain(m *testing.M) {
	setup()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func setup() {
	logger.Init(infoLogs, errLogs)
}

func TestSecureHeaders(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	secureHeaders(next).ServeHTTP(rr, r)
	rs := rr.Result()

	want := "default-src 'self'; style-src 'self' 'unsafe-inline'; font-src 'self'"
	got := rs.Header.Get("Content-Security-Policy")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = "origin-when-cross-origin"
	got = rs.Header.Get("Referrer-Policy")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = "nosniff"
	got = rs.Header.Get("X-Content-Type-Options")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = "deny"
	got = rs.Header.Get("X-Frame-Options")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	want = "0"
	got = rs.Header.Get("X-XSS-Protection")
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	bytes.TrimSpace(body)
	want = "OK"
	got = string(body)
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestRecoverPanic(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	panicMsg := "oops"
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(panicMsg)
	})

	recoverPanic(next).ServeHTTP(rr, r)
	rs := rr.Result()

	wantCode := http.StatusInternalServerError
	if rs.StatusCode != wantCode {
		t.Errorf("want %d, got %d", wantCode, rs.StatusCode)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}

	bytes.TrimSpace(body)
	wantText := http.StatusText(wantCode) + "\n"
	if string(body) != wantText {
		t.Errorf("want %q, got %q", wantText, string(body))
	}

	if !strings.Contains(errLogs.String(), panicMsg) {
		t.Errorf("want error log to contain %q", panicMsg)
	}
}
