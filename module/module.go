// Package module is wowapi's public module SDK: the contract a product
// module implements and the capability-scoped Context it registers against.
//
// Modules live in consuming product repositories (see
// docs/blueprint/06-module-sdk.md and 11-framework-distribution-and-
// consumption.md); the framework repo keeps only private neutral fixtures
// under internal/testmodules.
//
// Phase 0 ships the minimal contract (D-0006): Context grows one accessor at
// a time alongside the kernel capability each phase delivers (Routes in
// Phase 3, Authz in Phase 4, Migrations/Seeds in Phase 5, …). Interface
// widening is an accepted breaking change while wowapi is v0.
package module

import (
	"context"
	"io/fs"
	"log/slog"
	"regexp"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/validation"
)

// Module is implemented by every product module (and by the framework's
// private test fixtures).
type Module interface {
	// Name is the unique module identifier: lowercase, [a-z][a-z0-9_]*,
	// e.g. "requests". It prefixes permissions, resource types, events,
	// rule points, and migration history entries.
	Name() string

	// DependsOn lists module names this module requires. The app topo-sorts
	// registration by this graph; unknown names and cycles fail boot.
	DependsOn() []string

	// Register wires the module into the framework: routes, permissions,
	// seeds, migrations, jobs, event handlers, … via ctx. Register must only
	// wire — no I/O, no business logic.
	Register(ctx Context) error
}

// Context is the capability-scoped registration surface handed to
// Module.Register. Modules receive registries and services — never raw
// pools, never global config.
type Context interface {
	// Logger returns a module-scoped structured logger (pre-tagged with the
	// module name).
	Logger() *slog.Logger

	// Config returns the module's namespaced configuration view
	// (modules.<name>.* only — see docs/blueprint/12 §2).
	Config() config.ModuleView

	// Routes returns the module's route registry. Every route must declare
	// metadata (a permission or explicit public opt-out); registration errors
	// surface at boot (Phase 3; blueprint 05 §1).
	Routes() *httpx.Router

	// Validator returns the shared request validator used by
	// httpx.BindAndValidate (Phase 3).
	Validator() *validation.Validator

	// Permissions returns the authorization permission registry the module
	// declares its permissions into; an unregistered permission can never be
	// authorized (deny-by-default, boot-validated — Phase 4, blueprint 01 §3).
	Permissions() *authz.Registry

	// Resources returns the resource-type registry the module declares its
	// resource types into (Phase 4).
	Resources() *resource.Registry

	// Authz returns the authorization evaluator for record-level checks and
	// list filtering (Phase 4).
	Authz() authz.Evaluator

	// Tx returns the tenant transaction manager — the only door to the
	// database for module work (Phase 2/5).
	Tx() database.TxManager

	// IDGen returns the id generator (UUIDv7); Clock returns the wall clock —
	// both injectable so tests run deterministic sequences (Phase 5).
	IDGen() model.IDGen

	// Migrations registers the module's goose migrations. fsys must be ROOTED
	// at the .sql files (use fs.Sub if they live in a subdirectory), matching
	// goose's convention. Seeds registers its embedded YAML catalog bundle;
	// OpenAPI registers its spec fragment. Applied/synced by the app at boot
	// (Phase 5, blueprint 06 §2).
	Migrations(fsys fs.FS)
	Seeds(fsys fs.FS)
	OpenAPI(fragment []byte)

	// Health registers a named readiness check (Phase 5).
	Health(name string, check func(context.Context) error)

	// ProvidePort declares an implementation another module may consume;
	// Port fetches a declared port (both checked at boot — an unsatisfied
	// Port dependency fails Validate). Inter-module access is via ports only,
	// never another module's internals (Phase 5, blueprint 06 §2).
	ProvidePort(name string, impl any)
	Port(name string) (any, error)

	// Later phases add, alongside the capability they deliver:
	//   Events() outbox.HandlerRegistry / Jobs() jobs.Registry (Phase 6)
	//   Rules() rules.Registry / Workflows() workflow.Registry (Phase 7)
	//   Documents() document.Service           (Phase 8)
	//   Notify() notify.Sender / Webhooks() webhook.Service    (Phase 9)
}

// nameRE constrains module names; ValidName is used by app.Validate.
var nameRE = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// ValidName reports whether s is a legal module name.
func ValidName(s string) bool { return nameRE.MatchString(s) }
