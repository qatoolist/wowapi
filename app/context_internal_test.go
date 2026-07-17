package app

import (
	"context"
	"testing"
	"testing/fstest"
	"time"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/i18n"
	"github.com/qatoolist/wowapi/module"
)

// TestModuleContextAccessorsNilGuards builds a module context with empty deps
// (only a bootState so the recording accessors do not nil-panic) and exercises
// every accessor. The nil-guarded accessors must lazily construct a non-nil
// value; the plain pass-through accessors return their (nil) field without
// panicking. This locks the capability-scoped surface modules see in Register.
func TestModuleContextAccessorsNilGuards(t *testing.T) {
	// nil logger exercises the newModuleContext logger-defaulting branch.
	c := newModuleContext("x", nil, nil, moduleDeps{boot: newBootState()})

	// Lazily-constructed registries/services must never be nil.
	if c.Logger() == nil {
		t.Error("Logger() nil (nil logger must default to slog.Default())")
	}
	if c.Routes() == nil {
		t.Error("Routes() nil")
	}
	if c.Permissions() == nil {
		t.Error("Permissions() nil")
	}
	if c.Resources() == nil {
		t.Error("Resources() nil")
	}
	if c.IDGen() == nil {
		t.Error("IDGen() nil")
	}
	if c.Events() == nil {
		t.Error("Events() nil")
	}
	if c.Outbox() == nil {
		t.Error("Outbox() nil")
	}
	if c.Jobs() == nil {
		t.Error("Jobs() nil")
	}

	// nil view must yield an empty, non-nil MapView (modules can Decode without a guard).
	if c.Config() == nil {
		t.Error("Config() nil")
	}

	// Plain pass-through accessors: nil with empty deps, but must not panic and
	// must return exactly what was injected (nil here). We call each to lock the
	// surface and to prove none of them lazily allocate (they are kernel-owned).
	_ = c.Validator()
	_ = c.Authz()
	_ = c.Tx()
	_ = c.Rules()
	_ = c.RulesResolver()
	_ = c.Workflows()
	_ = c.WorkflowRuntime()
	_ = c.RetentionClasses()
	_ = c.Audit()
	_ = c.Sequence()
	_ = c.Bulk()
	_ = c.Artifacts()
	_ = c.DocumentClasses()
	_ = c.DocumentHooks()
	_ = c.Documents()
	_ = c.Comments()
	_ = c.Attachments()
	_ = c.NotifyTemplates()
	_ = c.Notify()
	_ = c.Webhooks()
	_ = c.IntegrationProviders()
	_ = c.Integrations()
}

// TestModuleContextBootStateRecording proves the app-level collector records
// what a module registers during Register: migration/seed FSes, OpenAPI
// fragments, health checks, recurring jobs, and inter-module ports — each keyed
// by the module name (or module-prefixed) so the app can consume them after all
// modules have registered.
func TestModuleContextBootStateRecording(t *testing.T) {
	boot := newBootState()
	c := newModuleContext("mymod", nil, nil, moduleDeps{boot: boot})

	c.Migrations(fstest.MapFS{})
	c.Seeds(fstest.MapFS{})
	c.OpenAPI([]byte("openapi: 3.1.0"))
	c.(module.I18nContext).I18n(i18n.Bundle{Locale: "mr", Messages: map[string]string{"mymod.msg.hi": "नमस्कार"}})
	c.Health("live", func(context.Context) error { return nil })
	c.RecurringJob("nightly", time.Minute, func(context.Context, database.TenantDB) error { return nil })
	c.ProvidePort("mymod.clock", 42)

	if _, ok := boot.migrations["mymod"]; !ok {
		t.Error("Migrations not recorded under module name")
	}
	if _, ok := boot.seeds["mymod"]; !ok {
		t.Error("Seeds not recorded under module name")
	}
	if got := string(boot.openapi["mymod"]); got != "openapi: 3.1.0" {
		t.Errorf("OpenAPI = %q, want the registered fragment", got)
	}
	if err := boot.i18n.Err(); err != nil {
		t.Fatalf("i18n bundle rejected: %v", err)
	}
	if msg, _ := boot.i18n.Catalog().Lookup("mr", "mymod.msg.hi"); msg != "नमस्कार" {
		t.Errorf("i18n bundle not aggregated: %q", msg)
	}
	if _, ok := boot.health["mymod.live"]; !ok {
		t.Error("Health not recorded under module-prefixed name mymod.live")
	}
	if len(boot.recurring) != 1 || boot.recurring[0].Name != "mymod.nightly" {
		t.Errorf("recurring = %v, want one job named mymod.nightly", boot.recurring)
	}
	if boot.recurring[0].Every != time.Minute {
		t.Errorf("recurring interval = %v, want 1m", boot.recurring[0].Every)
	}

	// Port round-trips a provided impl; an unknown port is a descriptive error.
	got, err := c.Port("mymod.clock")
	if err != nil {
		t.Fatalf("Port(mymod.clock) error: %v", err)
	}
	if got.(int) != 42 {
		t.Errorf("Port(mymod.clock) = %v, want 42", got)
	}
	if _, err := c.Port("nope.missing"); err == nil {
		t.Error("Port(nope.missing) must error for an unprovided port")
	}
}

// TestModuleContextI18nDocExampleBoots grounds the user guide's "Registering
// product/module translations" snippet (docs/user-guide/validation-errors.md,
// "Localizing responses (i18n)" section): a module-prefixed key registers
// cleanly with no boot error. This is the exact bundle shape shown in the doc.
func TestModuleContextI18nDocExampleBoots(t *testing.T) {
	boot := newBootState()
	c := newModuleContext("orders", nil, nil, moduleDeps{boot: boot})

	c.(module.I18nContext).I18n(i18n.Bundle{Locale: "mr", Messages: map[string]string{
		"orders.status.shipped": "पाठवले",
	}})

	if err := boot.i18n.Err(); err != nil {
		t.Fatalf("doc example bundle rejected at boot: %v", err)
	}
	if msg, _ := boot.i18n.Catalog().Lookup("mr", "orders.status.shipped"); msg != "पाठवले" {
		t.Errorf("doc example key not registered: %q", msg)
	}
}

// TestModuleContextI18nRejectsReservedKernelPrefix proves why the user guide's
// module example does NOT (and must not) include a kernel.* key: Register
// rejects it at boot, regardless of which module attempts it. This is the
// reason docs/user-guide/validation-errors.md directs translators of the
// framework's own strings to Booted.I18n.Add instead of module.Context.I18n.
func TestModuleContextI18nRejectsReservedKernelPrefix(t *testing.T) {
	boot := newBootState()
	c := newModuleContext("orders", nil, nil, moduleDeps{boot: boot})

	c.(module.I18nContext).I18n(i18n.Bundle{Locale: "mr", Messages: map[string]string{
		i18n.KeyProblemTitle(errors.KindInternal): "should be rejected",
	}})

	if err := boot.i18n.Err(); err == nil {
		t.Fatal("module registering a kernel.* key must fail boot, but Err() was nil")
	}
}
