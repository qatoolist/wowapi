package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/rules"
	"github.com/qatoolist/wowapi/kernel/seeds"
	"github.com/qatoolist/wowapi/migrations"
	"github.com/qatoolist/wowapi/module"
	"github.com/qatoolist/wowapi/testkit"
)

func TestIntegrationMigrationCurrencyCheckPassesWhenCurrent(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	a := app.New()
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	cfg := config.Defaults()
	fp, _ := config.FingerprintOf(cfg)

	health := app.ReadinessWithCatalogs(booted, fp, h.Platform, migrations.Kernel(), migrations.SourceName, nil)
	rec := httptest.NewRecorder()
	health.Readiness()(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	var body struct {
		Status  string            `json:"status"`
		Checks  map[string]string `json:"checks"`
		Details map[string]any    `json:"details"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusOK {
		t.Fatalf("readiness = %d, want 200; checks=%v", rec.Code, body.Checks)
	}
	if _, ok := body.Checks["migration_currency"]; !ok {
		t.Fatalf("missing migration_currency check; checks=%v", body.Checks)
	}
	if body.Details["migration_version"] == nil {
		t.Fatalf("missing migration_version detail; details=%v", body.Details)
	}
}

func TestIntegrationMigrationCurrencyCheckFailsWhenStale(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	a := app.New()
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	cfg := config.Defaults()
	fp, _ := config.FingerprintOf(cfg)

	// Simulate a stale-migrated database by rewinding the goose version table.
	if _, err := h.Admin.Exec(context.Background(),
		"UPDATE goose_version_wowapi SET version_id = 1, is_applied = true"); err != nil {
		t.Fatalf("rewind migration version: %v", err)
	}

	health := app.ReadinessWithCatalogs(booted, fp, h.Platform, migrations.Kernel(), migrations.SourceName, nil)
	rec := httptest.NewRecorder()
	health.Readiness()(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	var body struct {
		Status  string            `json:"status"`
		Checks  map[string]string `json:"checks"`
		Details map[string]any    `json:"details"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("readiness = %d, want 503 for stale DB; checks=%v", rec.Code, body.Checks)
	}
	if body.Status != "not_ready" {
		t.Fatalf("status = %q, want not_ready", body.Status)
	}
	if _, ok := body.Checks["migration_currency"]; !ok {
		t.Fatalf("missing migration_currency check; checks=%v", body.Checks)
	}
}

func TestIntegrationReadinessReportsSeedAndRuleHashes(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	m := funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.Seeds(fstest.MapFS{"seed.yaml": &fstest.MapFile{Data: []byte(`
permissions:
  - key: widgets.widget.create
    description: create a widget
`)}})
		mc.Rules().Register("widgets", rules.Point{
			Key:              "widgets.feature.flag",
			Module:           "widgets",
			ValueSchema:      []byte(`{"type":"boolean"}`),
			Default:          []byte(`false`),
			AllowedScopes:    []rules.ScopeKind{rules.ScopePlatform},
			RequiresApproval: false,
			Description:      "feature flag",
		})
		return nil
	}}

	a := app.New()
	a.Register(m)
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	// Seed-sync the declared catalogs so seed_catalogs check passes and the hash
	// is recorded.
	if _, err := seeds.Apply(context.Background(), h.Platform, booted.RuntimeSeeds(), seeds.ApplyOptions{Actor: "test"}); err != nil {
		t.Fatalf("seed sync: %v", err)
	}

	cfg := config.Defaults()
	fp, _ := config.FingerprintOf(cfg)

	health := app.ReadinessWithCatalogs(booted, fp, h.Platform, migrations.Kernel(), migrations.SourceName, nil)
	rec := httptest.NewRecorder()
	health.Readiness()(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))

	var body struct {
		Status  string            `json:"status"`
		Checks  map[string]string `json:"checks"`
		Details map[string]any    `json:"details"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if rec.Code != http.StatusOK {
		t.Fatalf("readiness = %d, want 200; checks=%v details=%v", rec.Code, body.Checks, body.Details)
	}
	if body.Details["seed_catalog_hash"] == nil {
		t.Fatalf("missing seed_catalog_hash detail; details=%v", body.Details)
	}
	if body.Details["rule_hash"] == nil {
		t.Fatalf("missing rule_hash detail; details=%v", body.Details)
	}
}
