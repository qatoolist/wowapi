package app_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/foundation/document"
	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/module"
	"github.com/qatoolist/wowapi/testkit"
)

// widgetsSeed is a valid, ownership-clean seed catalog for a "widgets" module.
const widgetsSeed = `
permissions:
  - key: widgets.widget.read
    description: read a widget
resource_types:
  - key: widgets.widget
    description: a widget aggregate
`

func discardKernel(t *testing.T, h *testkit.DBHandle) *kernel.Kernel {
	t.Helper()
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{
		Pool: h.Runtime, Platform: h.Platform, Tx: h.TxM,
	})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	return k
}

// funcModule is a minimal module whose registration logic is supplied inline.
type funcModule struct {
	name string
	deps []string
	reg  func(mc module.Context) error
}

func (m funcModule) Name() string        { return m.name }
func (m funcModule) DependsOn() []string { return m.deps }
func (m funcModule) Register(mc module.Context) error {
	if m.reg == nil {
		return nil
	}
	return m.reg(mc)
}

// TestBootHappyPathWiresEverything boots a module that registers a seed
// (permissions + resource types), a migration FS, an OpenAPI fragment, a health
// check, an inter-module port, and a route whose permission the seed declares.
// Boot must succeed and surface every artifact on Booted, and the seed-declared
// permission must be registered in the shared authz registry.
func TestBootHappyPathWiresEverything(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	m := funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.Seeds(fstest.MapFS{"seed.yaml": &fstest.MapFile{Data: []byte(widgetsSeed)}})
		mc.Migrations(fstest.MapFS{"0001_init.sql": &fstest.MapFile{Data: []byte("SELECT 1;")}})
		mc.OpenAPI([]byte("openapi: 3.1.0"))
		mc.Health("ready", func(context.Context) error { return nil })
		mc.ProvidePort("widgets.clock", 7)
		mc.Routes().Handle(http.MethodGet, "/widgets",
			httpx.RouteMeta{Permission: "widgets.widget.read"},
			func(http.ResponseWriter, *http.Request) {})
		return nil
	}}

	a := app.New()
	a.Register(m)
	booted, err := a.Boot(context.Background(), k, nil)
	if err != nil {
		t.Fatalf("Boot: %v", err)
	}

	if !k.Perms.Has("widgets.widget.read") {
		t.Error("seed-declared permission not registered into the shared authz registry")
	}
	if _, ok := booted.RuntimeMigrations()["widgets"]; !ok {
		t.Error("Booted.Migrations missing widgets FS")
	}
	if got := string(booted.RuntimeOpenAPI()["widgets"]); got != "openapi: 3.1.0" {
		t.Errorf("Booted.OpenAPI[widgets] = %q", got)
	}
	if _, ok := app.CapturedHealth(booted)["widgets.ready"]; !ok {
		t.Error("Booted.Health missing widgets.ready")
	}
	if len(booted.RuntimeSeeds().Permissions) != 1 || booted.RuntimeSeeds().Permissions[0].Key != "widgets.widget.read" {
		t.Errorf("merged RuntimeSeeds().Permissions = %+v", booted.RuntimeSeeds().Permissions)
	}
	if len(booted.RuntimeSeeds().ResourceTypes) != 1 {
		t.Errorf("merged RuntimeSeeds().ResourceTypes = %+v", booted.RuntimeSeeds().ResourceTypes)
	}
	if len(booted.RuntimeRouter().Routes()) != 1 {
		t.Errorf("router has %d routes, want 1", len(booted.RuntimeRouter().Routes()))
	}
}

// stepUpSeed declares one plain and one step-up-gated permission.
const stepUpSeed = `
permissions:
  - key: widgets.widget.read
    description: read a widget
  - key: widgets.widget.approve
    description: approve a widget
    step_up: true
`

// TestBootPropagatesStepUp proves a seed-declared step_up: true permission
// reaches the shared authz registry as Permission.StepUp — the seed→boot→
// registry plumbing GAP-004 connects. A permission that omits step_up must
// register as StepUp: false.
func TestBootPropagatesStepUp(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	m := funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.Seeds(fstest.MapFS{"seed.yaml": &fstest.MapFile{Data: []byte(stepUpSeed)}})
		return nil
	}}

	a := app.New()
	a.Register(m)
	if _, err := a.Boot(context.Background(), k, nil); err != nil {
		t.Fatalf("Boot: %v", err)
	}

	readPerm, ok := k.Perms.Get("widgets.widget.read")
	if !ok {
		t.Fatal("widgets.widget.read not registered")
	}
	if readPerm.StepUp {
		t.Error("widgets.widget.read should not require step-up (seed omits step_up)")
	}

	approvePerm, ok := k.Perms.Get("widgets.widget.approve")
	if !ok {
		t.Fatal("widgets.widget.approve not registered")
	}
	if !approvePerm.StepUp {
		t.Error("widgets.widget.approve should require step-up (seed sets step_up: true)")
	}
	if approvePerm.StepUpPolicy != nil {
		t.Errorf("plain step_up: true should NOT populate StepUpPolicy: %+v", approvePerm.StepUpPolicy)
	}
}

// stepUpAMRSeed declares a permission requiring a SPECIFIC AMR (hwk), not the
// deployment default set, via the richer seed form (B8).
const stepUpAMRSeed = `
permissions:
  - key: widgets.widget.export
    description: export a widget
    step_up: true
    step_up_amr: [hwk]
    step_up_challenge: hwk
`

// TestBootPropagatesStepUpPolicy proves the richer step_up_amr/step_up_challenge
// seed form reaches the shared authz registry as Permission.StepUpPolicy — the
// same seed→boot→registry plumbing as the plain bool, extended (B8).
func TestBootPropagatesStepUpPolicy(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	m := funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.Seeds(fstest.MapFS{"seed.yaml": &fstest.MapFile{Data: []byte(stepUpAMRSeed)}})
		return nil
	}}

	a := app.New()
	a.Register(m)
	if _, err := a.Boot(context.Background(), k, nil); err != nil {
		t.Fatalf("Boot: %v", err)
	}

	exportPerm, ok := k.Perms.Get("widgets.widget.export")
	if !ok {
		t.Fatal("widgets.widget.export not registered")
	}
	if !exportPerm.StepUp {
		t.Error("widgets.widget.export should carry StepUp: true (seed sets step_up: true)")
	}
	if exportPerm.StepUpPolicy == nil {
		t.Fatal("widgets.widget.export should carry a StepUpPolicy (seed sets step_up_amr)")
	}
	if len(exportPerm.StepUpPolicy.RequiredAMR) != 1 || exportPerm.StepUpPolicy.RequiredAMR[0] != "hwk" {
		t.Errorf("StepUpPolicy.RequiredAMR = %v, want [hwk]", exportPerm.StepUpPolicy.RequiredAMR)
	}
	if exportPerm.StepUpPolicy.Challenge != "hwk" {
		t.Errorf("StepUpPolicy.Challenge = %q, want %q", exportPerm.StepUpPolicy.Challenge, "hwk")
	}
	// Boot populates the in-memory registry only; DB persistence of the plain
	// step_up bool (via seeds.Sync) is proven separately by
	// kernel/seeds.TestSyncPersistsStepUp and the httpx step-up e2e test —
	// the cheapest-correct path keeps StepUpPolicy registry-declared, never
	// DB-persisted (see authz.Permission.StepUpPolicy doc comment).
}

// TestBootConfigNamespaceIsolation confirms Boot threads each module's own
// config namespace into its context and nothing else (blueprint 06 §2). Both
// namespaces belong to registered modules (AR-04 T1 now rejects a namespace
// with no matching module), so "other" is a second real module rather than a
// phantom namespace.
func TestBootConfigNamespaceIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	var seen map[string]any
	m := funcModule{name: "widgets", reg: func(mc module.Context) error {
		return mc.Config().Decode(&seen)
	}}
	other := funcModule{name: "other", reg: func(mc module.Context) error {
		var v map[string]any
		return mc.Config().Decode(&v)
	}}
	a := app.New()
	a.Register(m)
	a.Register(other)
	ns := config.Namespaces{
		"widgets": config.MapView{"size": "large"},
		"other":   config.MapView{"secret": "hidden"},
	}
	if _, err := a.Boot(context.Background(), k, ns); err != nil {
		t.Fatalf("Boot: %v", err)
	}
	if seen["size"] != "large" {
		t.Errorf("module must see its own key: %v", seen)
	}
	if _, leaked := seen["secret"]; leaked {
		t.Errorf("module leaked another namespace's key: %v", seen)
	}
}

// TestBootFailsOnUnknownConfigNamespace proves Boot rejects a config
// `modules.<name>` namespace that has no corresponding registered module
// (AR-04 T1) — e.g. a typo like modules.polcy instead of modules.policy, or a
// leftover namespace for a module that was removed. Before this check such a
// namespace was silently retained as opaque, unvalidated data.
func TestBootFailsOnUnknownConfigNamespace(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	m := funcModule{name: "widgets", reg: func(mc module.Context) error {
		var seen map[string]any
		return mc.Config().Decode(&seen)
	}}
	a := app.New()
	a.Register(m)
	ns := config.Namespaces{
		"widgets": config.MapView{"size": "large"},
		"polcy":   config.MapView{"enabled": true}, // typo: no "polcy" module registered
	}
	_, err := a.Boot(context.Background(), k, ns)
	if err == nil || !strings.Contains(err.Error(), "polcy") ||
		!strings.Contains(err.Error(), "unknown module namespace") {
		t.Fatalf("Boot error = %v, want unknown-module-namespace failure naming %q", err, "polcy")
	}
}

// TestBootFailsOnGraphError proves Boot short-circuits on a module-graph error
// (here an unknown dependency) before any registration happens.
func TestBootFailsOnGraphError(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	a := app.New()
	a.Register(funcModule{name: "widgets", deps: []string{"ghost"}})
	_, err := a.Boot(context.Background(), k, nil)
	if err == nil || !strings.Contains(err.Error(), `unknown module "ghost"`) {
		t.Fatalf("Boot error = %v, want unknown-module graph error", err)
	}
}

// TestBootFailsOnRegisterError proves a module's Register error aborts boot with
// the offending module named.
func TestBootFailsOnRegisterError(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	boom := errors.New("register exploded")
	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(module.Context) error { return boom }})
	_, err := a.Boot(context.Background(), k, nil)
	if err == nil || !strings.Contains(err.Error(), "register exploded") ||
		!strings.Contains(err.Error(), `module "widgets": Register`) {
		t.Fatalf("Boot error = %v, want the module Register failure", err)
	}
}

// TestBootFailsOnUnknownRoutePermission proves deny-by-default: a route whose
// permission is declared by no seed/registration fails boot (an unknown
// permission could otherwise silently deny at runtime).
func TestBootFailsOnUnknownRoutePermission(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.Routes().Handle(http.MethodGet, "/w",
			httpx.RouteMeta{Permission: "widgets.widget.read"}, // never declared
			func(http.ResponseWriter, *http.Request) {})
		return nil
	}})
	_, err := a.Boot(context.Background(), k, nil)
	if err == nil || !strings.Contains(err.Error(), "widgets.widget.read") ||
		!strings.Contains(err.Error(), "not declared") {
		t.Fatalf("Boot error = %v, want unknown-route-permission failure", err)
	}
}

// TestBootFailsOnDocumentClassWithoutStorage proves the built-but-not-wired
// guard: a module registering a document class needs a storage adapter; without
// one Boot fails loudly rather than handing modules a nil Documents() service.
func TestBootFailsOnDocumentClassWithoutStorage(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h) // Deps has no Storage -> k.Documents == nil

	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.DocumentClasses().Register("widgets", document.Class{Key: "widgets.doc", Module: "widgets"})
		return nil
	}})
	_, err := a.Boot(context.Background(), k, nil)
	if err == nil || !strings.Contains(err.Error(), "no storage adapter is wired") {
		t.Fatalf("Boot error = %v, want document-class-without-storage failure", err)
	}
}

// TestBootFailsOnSeedOwnershipViolation proves seeds are ownership-checked: a
// module may only seed keys prefixed with its own name; a foreign key fails the
// seed load and thus boot.
func TestBootFailsOnSeedOwnershipViolation(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	const foreign = "permissions:\n  - key: other.widget.read\n    description: not mine\n"
	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.Seeds(fstest.MapFS{"seed.yaml": &fstest.MapFile{Data: []byte(foreign)}})
		return nil
	}})
	_, err := a.Boot(context.Background(), k, nil)
	if err == nil || !strings.Contains(err.Error(), "may only seed keys prefixed") {
		t.Fatalf("Boot error = %v, want seed ownership violation", err)
	}
}

// TestBootAccumulatesRegistryErrors proves Boot gathers ALL shared-registry
// validation failures (permissions, resource types, routes, event subscriptions,
// job kinds) into one error rather than failing on the first — so a module
// author sees the full list at once.
func TestBootAccumulatesRegistryErrors(t *testing.T) {
	h := testkit.NewDB(t)
	k := discardKernel(t, h)

	a := app.New()
	a.Register(funcModule{name: "widgets", reg: func(mc module.Context) error {
		mc.Permissions().Register(authz.Permission{Key: "not a valid key"})                                      // perms.Err
		mc.Resources().Register("widgets", resource.TypeSpec{Key: "not_valid_key"})                              // resources.Err
		mc.Routes().Handle(http.MethodGet, "/x", httpx.RouteMeta{}, func(http.ResponseWriter, *http.Request) {}) // router.Err (neither Permission nor Public)
		mc.Events().Subscribe("", "", nil)                                                                       // events.Err
		mc.Jobs().RegisterKind("", nil, jobs.Idempotency{}, jobs.DefaultRetry())                                 // jobs.Err
		return nil
	}})
	_, err := a.Boot(context.Background(), k, nil)
	if err == nil {
		t.Fatal("Boot must fail when shared registries have validation errors")
	}
	msg := err.Error()
	for _, want := range []string{
		"permission", // invalid permission key
		"resource",   // invalid resource type key
		"route",      // route metadata failure
		"subscription",
		"job",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("accumulated boot error missing %q signal:\n%s", want, msg)
		}
	}
}
