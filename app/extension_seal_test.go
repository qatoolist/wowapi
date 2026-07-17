package app_test

import (
	"context"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/foundation/document"
	"github.com/qatoolist/wowapi/foundation/integration"
	"github.com/qatoolist/wowapi/foundation/notify"
	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/retention"
	"github.com/qatoolist/wowapi/kernel/rules"
	"github.com/qatoolist/wowapi/kernel/seeds"
	"github.com/qatoolist/wowapi/kernel/workflow"
	"github.com/qatoolist/wowapi/module"
	"github.com/qatoolist/wowapi/testkit"
)

// Closure-review regression (adversarial closure review 2026-07-17, F-10):
// after Boot returns, the extension model is SEALED for EVERY registry class —
// not only the collectors guarded by moduleContext.mustBeUnsealed. A retained
// module.Context hands out the live shared registries (routes, permissions,
// resources, events, jobs, rules, workflows, retention/document classes,
// hooks, templates, providers), and Booted intentionally exposes the live
// Router/Events/Jobs for serving; the runtime does LIVE lookups against Jobs
// (kernel/jobs/runner.go) and Events (relay dispatch), so an unsealed mutator
// would let post-boot code introduce routes, job kinds, or subscriptions that
// boot validation never saw. Every mutator must panic after boot.
func TestSealedExtensionModelRejectsEveryPostBootRegistration(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	var retained module.Context
	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(mc module.Context) error {
		retained = mc
		return nil
	}})
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	noopHandler := func(http.ResponseWriter, *http.Request) {}
	mutations := []struct {
		name string
		fn   func()
	}{
		// Registries reachable from a retained module context.
		{"Routes.Handle", func() {
			retained.Routes().Handle(http.MethodGet, "/late", httpx.RouteMeta{}, noopHandler)
		}},
		{"Permissions.Register", func() { retained.Permissions().Register(authz.Permission{}) }},
		{"Resources.Register", func() { retained.Resources().Register("widgets", resource.TypeSpec{}) }},
		{"Events.Subscribe", func() { retained.Events().Subscribe("late.event", "late", nil) }},
		{"Jobs.RegisterKind", func() { retained.Jobs().RegisterKind("late.kind", nil, jobs.RetryPolicy{}) }},
		{"Rules.Register", func() { retained.Rules().Register("widgets", rules.Point{}) }},
		{"Workflows.RegisterDefinition", func() { _ = retained.Workflows().RegisterDefinition(workflow.Definition{}) }},
		{"Workflows.RegisterAutoAction", func() { retained.Workflows().RegisterAutoAction("late", nil) }},
		{"Workflows.RegisterAssigneeResolver", func() { retained.Workflows().RegisterAssigneeResolver("late", nil) }},
		{"RetentionClasses.Register", func() { retained.RetentionClasses().Register(retention.RecordClass{}) }},
		{"DocumentClasses.Register", func() { retained.DocumentClasses().Register("widgets", document.Class{}) }},
		{"DocumentHooks.OnFileUpload", func() { retained.DocumentHooks().OnFileUpload(nil) }},
		{"DocumentHooks.OnDocumentAccess", func() { retained.DocumentHooks().OnDocumentAccess(nil) }},
		{"NotifyTemplates.Register", func() { retained.NotifyTemplates().Register("widgets", notify.TemplateSpec{}) }},
		{"IntegrationProviders.Register", func() { retained.IntegrationProviders().Register("widgets", integration.Provider(nil)) }},
		// The live collectors Booted itself exposes for serving.
		{"Booted.Router.Handle", func() {
			booted.Router.Handle(http.MethodGet, "/late2", httpx.RouteMeta{}, noopHandler)
		}},
		{"Booted.Events.Subscribe", func() { booted.Events.Subscribe("late.event2", "late2", nil) }},
		{"Booted.Jobs.RegisterKind", func() { booted.Jobs.RegisterKind("late.kind2", nil, jobs.RetryPolicy{}) }},
	}
	for _, m := range mutations {
		t.Run(m.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("%s after boot did not panic — the extension model is not sealed for this class", m.name)
				}
			}()
			m.fn()
		})
	}
}

// Second closure-audit regression (2026-07-17, F-10): recurring-job
// declarations are boot-validated — duplicate names silently share one
// scheduler row (one starves), a nonpositive interval hot-loops, and a nil
// callback panics only when first due. All must be collected boot errors.
func TestBootRejectsInvalidRecurringDeclarations(t *testing.T) {
	run := func(ctx context.Context, db database.TenantDB) error { return nil }
	for name, tc := range map[string]struct {
		reg  func(mc module.Context)
		want string
	}{
		"duplicate name": {func(mc module.Context) {
			mc.RecurringJob("sweep", time.Minute, run)
			mc.RecurringJob("sweep", time.Hour, run)
		}, "declared more than once"},
		"nonpositive interval": {func(mc module.Context) {
			mc.RecurringJob("sweep", 0, run)
		}, "nonpositive interval"},
		"nil callback": {func(mc module.Context) {
			mc.RecurringJob("sweep", time.Minute, nil)
		}, "nil callback"},
		"empty name": {func(mc module.Context) {
			mc.RecurringJob("", time.Minute, run)
		}, "non-empty name"},
	} {
		t.Run(name, func(t *testing.T) {
			err := bootModules(t, funcModule{name: "widgets", reg: func(mc module.Context) error {
				tc.reg(mc)
				return nil
			}})
			if err == nil {
				t.Fatalf("boot accepted a recurring declaration with %s", name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("boot error %v does not explain %q", err, tc.want)
			}
		})
	}
}

// Nil document hooks are boot errors, not deferred panics (F-10).
func TestBootRejectsNilDocumentHooks(t *testing.T) {
	err := bootModules(t, funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.DocumentHooks().OnFileUpload(nil)
		mc.DocumentHooks().OnDocumentAccess(nil)
		return nil
	}})
	if err == nil {
		t.Fatal("boot accepted nil document hooks (they panic on first invocation)")
	}
	if !strings.Contains(err.Error(), "nil hook") {
		t.Fatalf("boot error %v does not name the nil hook", err)
	}
}

// Second closure-audit regression (2026-07-17, F-10): Booted's exported
// collector fields are assignable, and the registry seal cannot prevent a
// caller REPLACING a field with a fresh unsealed registry or map. The
// framework's consumers must therefore read the boot-validated runtime view:
// replacing every replaceable field must leave RuntimeRouter/RuntimeEvents/
// RuntimeJobs/RuntimeMigrations and a Readiness handler built AFTER the
// replacement completely unaffected.
func TestBootedFieldReplacementCannotAlterRuntimeState(t *testing.T) {
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
		mc.Health("real", func(context.Context) error { return nil })
		mc.Migrations(fstest.MapFS{"0001_real.up.sql": &fstest.MapFile{Data: []byte("SELECT 1;")}})
		return nil
	}})
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	validatedRouter := booted.RuntimeRouter()
	validatedEvents := booted.RuntimeEvents()
	validatedJobs := booted.RuntimeJobs()

	// Wholesale field replacement — fresh, UNSEALED registries and maps.
	booted.Router = httpx.NewRouter()
	booted.Events = outbox.NewHandlerRegistry()
	booted.Jobs = jobs.NewRegistry()
	booted.Health = map[string]func(context.Context) error{
		"evil": func(context.Context) error {
			t.Error("replaced health map reached the live readiness handler")
			return nil
		},
	}
	booted.Migrations = map[string]fs.FS{
		"evil": fstest.MapFS{"0001_evil.up.sql": &fstest.MapFile{Data: []byte("DROP TABLE users;")}},
	}

	if booted.RuntimeRouter() != validatedRouter || booted.RuntimeRouter() == booted.Router {
		t.Fatal("RuntimeRouter follows the replaced Router field, not the boot-validated router")
	}
	if booted.RuntimeEvents() != validatedEvents || booted.RuntimeEvents() == booted.Events {
		t.Fatal("RuntimeEvents follows the replaced Events field")
	}
	if booted.RuntimeJobs() != validatedJobs || booted.RuntimeJobs() == booted.Jobs {
		t.Fatal("RuntimeJobs follows the replaced Jobs field")
	}
	migs := booted.RuntimeMigrations()
	if _, ok := migs["evil"]; ok {
		t.Fatal("RuntimeMigrations includes the replaced migration set")
	}
	if _, ok := migs["widgets"]; !ok {
		t.Fatalf("RuntimeMigrations lost the boot-validated module set: %v", migs)
	}

	// The readiness aggregator is built AFTER the replacement — a consumer of
	// the exported field would serve the injected check here.
	health := app.Readiness(booted, config.Fingerprint{}, nil)
	rec := httptest.NewRecorder()
	health.Readiness().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	body := rec.Body.String()
	if strings.Contains(body, "evil") {
		t.Fatalf("readiness built after field replacement serves the injected check: %s", body)
	}
	if !strings.Contains(body, "widgets.real") {
		t.Fatalf("readiness built after field replacement lost the boot-validated check: %s", body)
	}
}

// Third closure-audit regressions (2026-07-17, F-10): seed state and the i18n
// catalog are part of the boot-validated runtime view; migration content is
// MATERIALIZED at boot. Replacing the public mirrors, mutating nested seed
// slices, or mutating the module-owned migration filesystem after boot must
// not change what the runtime consumers (readiness, migrate, locale
// middleware) operate on.
func TestSeedsI18nAndMigrationContentAreBootCaptured(t *testing.T) {
	h := testkit.NewDB(t)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	migFS := fstest.MapFS{"0001_real.up.sql": &fstest.MapFile{Data: []byte("SELECT 1;")}}
	seedFS := fstest.MapFS{"catalog.yaml": &fstest.MapFile{Data: []byte(
		"permissions:\n  - key: widgets.thing.read\n    description: read things\nroles:\n  - key: widgets.reader\n    name: Reader\n    permissions: [widgets.thing.read]\n")}}
	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.Migrations(migFS)
		mc.Seeds(seedFS)
		return nil
	}})
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}
	frozenI18n := booted.RuntimeI18n()
	if frozenI18n == nil {
		t.Fatal("RuntimeI18n returned nil")
	}

	// (1) Replace the public mirrors wholesale.
	booted.Seeds = seeds.Bundle{Permissions: []seeds.PermissionSeed{{Key: "evil.thing.admin"}}}
	booted.I18n = nil

	rs := booted.RuntimeSeeds()
	if len(rs.Permissions) != 1 || rs.Permissions[0].Key != "widgets.thing.read" {
		t.Fatalf("RuntimeSeeds follows the replaced Seeds field: %+v", rs.Permissions)
	}
	if booted.RuntimeI18n() != frozenI18n {
		t.Fatal("RuntimeI18n follows the replaced I18n field")
	}

	// (2) Mutate nested slices on a RuntimeSeeds result AND on the replaced
	// public mirror; the validated catalog must be unaffected.
	rs.Roles[0].Permissions[0] = "evil.everything.admin"
	rs.Permissions[0].Key = "evil.thing.read"
	again := booted.RuntimeSeeds()
	if again.Permissions[0].Key != "widgets.thing.read" || again.Roles[0].Permissions[0] != "widgets.thing.read" {
		t.Fatalf("mutating a RuntimeSeeds result altered the validated catalog: %+v", again)
	}

	// (3) Mutate the module-owned migration filesystem: the boot-materialized
	// snapshot must keep serving the validated bytes.
	migFS["0001_real.up.sql"].Data = []byte("DROP TABLE users;")
	migFS["0002_evil.up.sql"] = &fstest.MapFile{Data: []byte("DROP TABLE tenants;")}
	snap := booted.RuntimeMigrations()["widgets"]
	if snap == nil {
		t.Fatal("RuntimeMigrations lost the module set")
	}
	data, err := fs.ReadFile(snap, "0001_real.up.sql")
	if err != nil {
		t.Fatalf("read materialized migration: %v", err)
	}
	if string(data) != "SELECT 1;" {
		t.Fatalf("materialized migration content changed after post-boot FS mutation: %q", data)
	}
	if _, err := fs.ReadFile(snap, "0002_evil.up.sql"); err == nil {
		t.Fatal("a file added to the module FS after boot appeared in the materialized snapshot")
	}
}

// Third closure-audit regressions (2026-07-17): declaration registries reject
// nil/empty declarations at boot instead of deferring the failure to first use.
func TestBootRejectsNilAndEmptyDeclarations(t *testing.T) {
	for name, tc := range map[string]struct {
		reg  func(mc module.Context)
		want string
	}{
		"nil migrations FS": {func(mc module.Context) { mc.Migrations(nil) }, "nil fs.FS"},
		"nil seeds FS":      {func(mc module.Context) { mc.Seeds(nil) }, "nil fs.FS"},
		"empty health name": {func(mc module.Context) {
			mc.Health("", func(context.Context) error { return nil })
		}, "non-empty check name"},
		"nil health check": {func(mc module.Context) { mc.Health("db", nil) }, "nil func"},
	} {
		t.Run(name, func(t *testing.T) {
			err := bootModules(t, funcModule{name: "widgets", reg: func(mc module.Context) error {
				tc.reg(mc)
				return nil
			}})
			if err == nil {
				t.Fatalf("boot accepted %s", name)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("boot error %v does not explain %q", err, tc.want)
			}
		})
	}
}
