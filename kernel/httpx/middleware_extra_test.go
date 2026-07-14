package httpx_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/httpx"
)

// TestRecoverLogsPanicWithLogger covers the logger != nil branch of Recover: a
// panic is logged (to the provided logger) and still yields a clean 500 whose
// body never carries the panic value.
func TestRecoverLogsPanicWithLogger(t *testing.T) {
	logged := false
	logger := slog.New(&countingHandler{onHandle: func() { logged = true }})

	h := httpx.Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("secret panic detail") }),
		httpx.RequestID(),
		httpx.Recover(logger),
	)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	if !logged {
		t.Fatal("Recover with a logger must emit a log line for the panic")
	}
	if body := rec.Body.String(); body == "" || strings.Contains(body, "secret") {
		t.Errorf("panic detail must not reach the wire: %q", body)
	}
}

func TestRecoverSanitizesUntrustedLogFields(t *testing.T) {
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))
	h := httpx.Chain(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }),
		httpx.RequestID(),
		httpx.Recover(logger),
	)
	req := httptest.NewRequest(http.MethodGet, "/safe", nil)
	req.URL.Path = "/safe\nlevel=ERROR forged=true"
	req.Header.Set("X-Request-Id", "id\r\nforged=true")
	h.ServeHTTP(httptest.NewRecorder(), req)

	got := logs.String()
	trimmed := strings.TrimSuffix(got, "\n")
	if strings.Contains(trimmed, "\n") || strings.Contains(trimmed, "\r") {
		t.Fatalf("untrusted request metadata forged an additional log line: %q", got)
	}
	for _, escaped := range []string{`\\n`, `\\r`} {
		if !strings.Contains(got, escaped) {
			t.Errorf("sanitized log missing %q: %q", escaped, got)
		}
	}
}

// countingHandler is a minimal slog.Handler that flags when a record is handled.
type countingHandler struct{ onHandle func() }

func (h *countingHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *countingHandler) Handle(context.Context, slog.Record) error {
	h.onHandle()
	return nil
}
func (h *countingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *countingHandler) WithGroup(string) slog.Handler      { return h }
