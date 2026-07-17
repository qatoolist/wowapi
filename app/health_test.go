package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qatoolist/wowapi/v2/app"
	"github.com/qatoolist/wowapi/v2/kernel"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/httpx"
	"github.com/qatoolist/wowapi/v2/module"
	"github.com/qatoolist/wowapi/v2/testkit"
)

func TestReadinessWiresModuleAndFrameworkChecks(t *testing.T) {
	// Readiness consumes the boot-validated runtime view, so the Booted must
	// come from a real Boot (a hand-constructed value fails loudly by design —
	// third closure audit 2026-07-17).
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.Health("live", func(context.Context) error { return nil })
		return nil
	}})
	b, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	fp, _ := config.FingerprintOf(config.Defaults())
	hh := app.Readiness(b, fp, map[string]httpx.HealthCheck{
		"db":         func(context.Context) error { return nil },
		"migrations": func(context.Context) error { return errors.New("00012 pending") },
	})

	rec := httptest.NewRecorder()
	hh.Readiness()(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
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
	if body.Checks["module.widgets.live"] != "ok" {
		t.Fatalf("module check not wired: %+v", body.Checks)
	}
	if _, ok := body.Checks["db"]; !ok {
		t.Fatalf("framework db check not wired: %+v", body.Checks)
	}
}
