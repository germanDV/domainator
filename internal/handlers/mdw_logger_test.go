package handlers

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type logWriter struct {
	logs []string
}

func newLogWriter() *logWriter {
	return &logWriter{logs: make([]string, 0)}
}

func (lw *logWriter) Write(data []byte) (int, error) {
	lw.logs = append(lw.logs, string(data))
	return len(data), nil
}

func TestLoggerBuilder(t *testing.T) {
	lw := newLogWriter()
	logger := slog.New(slog.NewTextHandler(lw, nil))
	mdw := loggerBuilder(logger)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/foo?bar=baz", nil)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	mdw(handler).ServeHTTP(w, r)

	if len(lw.logs) < 1 {
		t.Fatal("Should have at least one log entry")
	}

	want := `level=INFO msg="Serving Request" method=GET path=/foo query="bar=baz"`
	if !strings.Contains(lw.logs[0], want) {
		t.Errorf("Log entry %q should include %q", lw.logs[0], want)
	}
}
