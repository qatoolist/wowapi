package app_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

	// Simulate a stale database relative to the clean embedded head (version 1)
	// by marking that head row unapplied, leaving goose's version-0 marker as the
	// highest applied version. Assert the fixture is genuinely behind before
	// exercising readiness so a future baseline squash cannot silently turn this
	// regression into a current/current comparison.
	expected, err := app.MaxMigrationVersion(migrations.Kernel())
	if err != nil {
		t.Fatalf("read embedded head: %v", err)
	}
	tag, err := h.Admin.Exec(context.Background(),
		"UPDATE goose_version_wowapi SET is_applied = false WHERE version_id = $1 AND is_applied", expected)
	if err != nil {
		t.Fatalf("rewind migration version: %v", err)
	}
	if tag.RowsAffected() != 1 {
		t.Fatalf("rewind affected %d applied rows, want exactly 1", tag.RowsAffected())
	}
	var applied int64
	if err := h.Admin.QueryRow(context.Background(), `SELECT version_id
		FROM goose_version_wowapi WHERE is_applied ORDER BY version_id DESC LIMIT 1`).Scan(&applied); err != nil {
		t.Fatalf("read rewound migration version: %v", err)
	}
	if applied >= expected {
		t.Fatalf("stale fixture is not stale: applied=%d expected=%d", applied, expected)
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
	check, ok := body.Checks["migration_currency"]
	if !ok {
		t.Fatalf("missing migration_currency check; checks=%v", body.Checks)
	}
	want := "applied version 0 lags expected 1"
	if !strings.Contains(check, want) {
		t.Fatalf("migration_currency = %q, want error containing %q", check, want)
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
