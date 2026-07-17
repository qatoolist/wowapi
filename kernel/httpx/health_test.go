package httpx_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qatoolist/wowapi/kernel/httpx"
)

func TestLivenessAlwaysOK(t *testing.T) {
	h := httpx.NewHealth("fp123")
	// Even with a failing readiness check, liveness stays 200.
	h.Register("db", func(context.Context) error { return errors.New("down") })
	rec := httptest.NewRecorder()
	h.Liveness()(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("liveness = %d, want 200", rec.Code)
	}
}

func TestReadinessAllPass(t *testing.T) {
	h := httpx.NewHealth("fp-abc").
		Register("db", func(context.Context) error { return nil }).
		Register("migrations", func(context.Context) error { return nil })
	rec := httptest.NewRecorder()
	h.Readiness()(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("readiness = %d, want 200", rec.Code)
	}
	var body struct {
		Status            string            `json:"status"`
		ConfigFingerprint string            `json:"config_fingerprint"`
		Checks            map[string]string `json:"checks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Status != "ready" || body.ConfigFingerprint != "fp-abc" {
		t.Fatalf("bad body: %+v", body)
	}
	if body.Checks["db"] != "ok" || body.Checks["migrations"] != "ok" {
		t.Fatalf("checks: %+v", body.Checks)
	}
}

func TestReadinessFailsWhenACheckFails(t *testing.T) {
	h := httpx.NewHealth("fp").
		Register("db", func(context.Context) error { return nil }).
		Register("migrations", func(context.Context) error { return errors.New("pending migration 00012") })
	rec := httptest.NewRecorder()
	h.Readiness()(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readiness with a failing check = %d, want 503", rec.Code)
	}
	var body struct {
		Status string            `json:"status"`
		Checks map[string]string `json:"checks"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Status != "not_ready" {
		t.Fatalf("status = %q, want not_ready", body.Status)
	}
	if body.Checks["migrations"] == "ok" {
		t.Fatal("failing migration check should not report ok")
	}
}
