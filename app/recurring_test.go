package app_test

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/app"
	"github.com/qatoolist/wowapi/v2/kernel"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/module"
	"github.com/qatoolist/wowapi/v2/testkit"
)

type recurringModule struct{}

func (recurringModule) Name() string        { return "widgets" }
func (recurringModule) DependsOn() []string { return nil }
func (recurringModule) Register(mc module.Context) error {
	mc.RecurringJob("nightly", time.Hour, func(context.Context, database.TenantDB) error { return nil })
	return nil
}

type evidenceModule struct{ auditNil, seqNil, bulkNil, artNil bool }

func (evidenceModule) Name() string        { return "evidence" }
func (evidenceModule) DependsOn() []string { return nil }
func (m *evidenceModule) Register(mc module.Context) error {
	m.auditNil = mc.Audit() == nil
	m.seqNil = mc.Sequence() == nil
	m.bulkNil = mc.Bulk() == nil
	m.artNil = mc.Artifacts() == nil
	return nil
}

// TestIntegrationModuleEvidenceAccessors is the CA-11 regression: the audit,
// sequence, bulk, and artifact evidence-layer services are wired and reachable
// from module.Context (previously built-but-not-wired).
func TestIntegrationModuleEvidenceAccessors(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{Pool: h.Runtime, Tx: h.TxM})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	a := app.New()
	m := &evidenceModule{}
	a.Register(m)
	if _, err := a.Boot(context.Background(), k, nil); err != nil {
		t.Fatalf("boot: %v", err)
	}
	if m.auditNil || m.seqNil || m.bulkNil || m.artNil {
		t.Fatalf("evidence accessors must be wired: audit=%v sequence=%v bulk=%v artifacts=%v",
			!m.auditNil, !m.seqNil, !m.bulkNil, !m.artNil)
	}
}

// TestIntegrationModuleRecurringJobCollected is the E5/CA-5 regression: a module
// registers a recurring job via module.Context, and it surfaces on Booted
// (module-prefixed) for the worker scheduler to run per active tenant.
func TestIntegrationModuleRecurringJobCollected(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{Pool: h.Runtime, Tx: h.TxM})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	a := app.New()
	a.Register(recurringModule{})
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("boot: %v", err)
	}
	if len(app.CapturedRecurring(booted)) != 1 {
		t.Fatalf("expected 1 recurring job on Booted, got %d", len(app.CapturedRecurring(booted)))
	}
	rj := app.CapturedRecurring(booted)[0]
	if rj.Name != "widgets.nightly" {
		t.Errorf("recurring job name = %q, want widgets.nightly (module-prefixed)", rj.Name)
	}
	if rj.Every != time.Hour {
		t.Errorf("interval = %v, want 1h", rj.Every)
	}
	if rj.Run == nil {
		t.Error("recurring job Run must be set")
	}
}
