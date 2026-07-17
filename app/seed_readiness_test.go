package app_test

// FBL-02 / MATRIX CS-21 post-fix readiness assertions: the framework-level
// seed_catalogs check fails loudly when the database catalogs are empty but
// modules declare seeds, and the readiness payload reports the seed/catalog
// hash once seed-sync has run.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/qatoolist/wowapi/v2/app"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/seeds"
	"github.com/qatoolist/wowapi/v2/module"
	"github.com/qatoolist/wowapi/v2/testkit"
)

func TestIntegrationReadinessEmptyCatalogsFailsNamed(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	m := funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.Seeds(fstest.MapFS{"seed.yaml": &fstest.MapFile{Data: []byte(`
permissions:
  - key: widgets.widget.create
    description: create a widget
`)}})
		return nil
	}}

	a := app.New()
	a.Register(m)
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	cfg := config.Defaults()
	cfg.Environment = config.EnvProd
	fp, _ := config.FingerprintOf(cfg)

	health := app.ReadinessWithCatalogs(booted, fp, h.Platform, nil, "", nil)
	rec := httptest.NewRecorder()
	health.Readiness()(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	var body struct {
		Status  string            `json:"status"`
		Checks  map[string]string `json:"checks"`
		Details map[string]any    `json:"details"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readiness = %d, want 503", rec.Code)
	}
	if body.Status != "not_ready" {
		t.Fatalf("status = %q, want not_ready", body.Status)
	}
	chk, ok := body.Checks["seed_catalogs"]
	if !ok {
		t.Fatalf("seed_catalogs check missing: %v", body.Checks)
	}
	if !strings.Contains(chk, "seed sync") {
		t.Fatalf("seed_catalogs error must name the fix: %q", chk)
	}
	if len(body.Details) != 0 {
		t.Fatalf("expected no details when unseeded, got %v", body.Details)
	}
}

func TestIntegrationReadinessAfterSyncReportsHash(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	seedYAML := `
version: v1
permissions:
  - key: widgets.widget.create
    description: create a widget
`
	m := funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.Seeds(fstest.MapFS{"seed.yaml": &fstest.MapFile{Data: []byte(seedYAML)}})
		return nil
	}}

	a := app.New()
	a.Register(m)
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	report, err := seeds.Apply(context.Background(), h.Platform, booted.RuntimeSeeds(), seeds.ApplyOptions{Actor: "test"})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}

	cfg := config.Defaults()
	cfg.Environment = config.EnvProd
	fp, _ := config.FingerprintOf(cfg)

	health := app.ReadinessWithCatalogs(booted, fp, h.Platform, nil, "", nil)
	rec := httptest.NewRecorder()
	health.Readiness()(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	var body struct {
		Status  string            `json:"status"`
		Checks  map[string]string `json:"checks"`
		Details map[string]any    `json:"details"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusOK {
		t.Fatalf("readiness = %d, want 200", rec.Code)
	}
	if body.Status != "ready" {
		t.Fatalf("status = %q, want ready", body.Status)
	}
	if body.Checks["seed_catalogs"] != "ok" {
		t.Fatalf("seed_catalogs check = %q, want ok", body.Checks["seed_catalogs"])
	}
	if body.Details == nil || body.Details["seed_catalog_hash"] != report.Hash {
		t.Fatalf("details.seed_catalog_hash = %v, want %q", body.Details, report.Hash)
	}
}
