package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/httpx"
)

func TestReadinessWiresModuleAndFrameworkChecks(t *testing.T) {
	b := &app.Booted{
		Health: map[string]func(context.Context) error{
			"widgets": func(context.Context) error { return nil },
		},
	}
	fp, _ := config.FingerprintOf(config.Defaults())
	h := app.Readiness(b, fp, map[string]httpx.HealthCheck{
		"db":         func(context.Context) error { return nil },
		"migrations": func(context.Context) error { return errors.New("00012 pending") },
	})

	rec := httptest.NewRecorder()
	h.Readiness()(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	// The failing migrations check must make the whole endpoint 503.
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readiness = %d, want 503", rec.Code)
	}
	var body struct {
		ConfigFingerprint string            `json:"config_fingerprint"`
		Checks            map[string]string `json:"checks"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body.ConfigFingerprint != fp.String() {
		t.Fatalf("fingerprint = %q, want %q", body.ConfigFingerprint, fp.String())
	}
	if body.Checks["module.widgets"] != "ok" {
		t.Fatalf("module check not wired: %+v", body.Checks)
	}
	if _, ok := body.Checks["db"]; !ok {
		t.Fatalf("framework db check not wired: %+v", body.Checks)
	}
}
