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
	"log/slog"
	"regexp"

	"github.com/qatoolist/wowapi/kernel/config"
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

	// Later phases add, alongside the capability they deliver:
	//   Routes() httpx.Router                  (Phase 3)
	//   Tx() database.TxManager                (Phase 2)
	//   Authz() authz.Evaluator                (Phase 4)
	//   Migrations(fs fs.FS) / Seeds(fs fs.FS) (Phase 5)
	//   Events() outbox.HandlerRegistry / Jobs() jobs.Registry (Phase 6)
	//   Rules() rules.Registry / Workflows() workflow.Registry (Phase 7)
	//   Documents() document.Service           (Phase 8)
	//   Notify() notify.Sender / Webhooks() webhook.Service    (Phase 9)
}

// nameRE constrains module names; ValidName is used by app.Validate.
var nameRE = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

// ValidName reports whether s is a legal module name.
func ValidName(s string) bool { return nameRE.MatchString(s) }
