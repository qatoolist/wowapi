// Module registration context (D-0006/D-0040; blueprint 06 §2).
//
// The context is capability-scoped: modules receive registries and services,
// never raw pools or global config. Accessors grow per phase alongside the
// kernel capability each delivers. Phase 5 wires the full set the current
// kernel supports (routes, permissions, resource types, authz, tx, migrations,
// seeds, openapi, health, inter-module ports) and injects the shared registries
// modules register into during boot.
package app

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/validation"
	"github.com/qatoolist/wowapi/module"
)

// bootState is the app-level collector shared by every module context during
// boot: modules register migration/seed FSes, OpenAPI fragments, health checks,
// and inter-module ports into it, and the app consumes them after all modules
// have registered.
type bootState struct {
	migrations map[string]fs.FS
	seeds      map[string]fs.FS
	openapi    map[string][]byte
	health     map[string]func(context.Context) error
	ports      map[string]any
}

func newBootState() *bootState {
	return &bootState{
		migrations: map[string]fs.FS{},
		seeds:      map[string]fs.FS{},
		openapi:    map[string][]byte{},
		health:     map[string]func(context.Context) error{},
		ports:      map[string]any{},
	}
}

// moduleContext implements module.Context. Unexported; callers receive the
// interface value.
type moduleContext struct {
	name   string
	logger *slog.Logger
	view   config.ModuleView
	router *httpx.Router
	val    *validation.Validator
	perms  *authz.Registry
	rtypes *resource.Registry
	eval   authz.Evaluator
	tx     database.TxManager
	idgen  model.IDGen
	events *outbox.HandlerRegistry
	writer outbox.Writer
	jobs   *jobs.Registry
	boot   *bootState
}

// moduleDeps bundles the shared registries/services the app injects into every
// module context, keeping the constructor signature stable as capabilities grow.
type moduleDeps struct {
	router *httpx.Router
	val    *validation.Validator
	perms  *authz.Registry
	rtypes *resource.Registry
	eval   authz.Evaluator
	tx     database.TxManager
	idgen  model.IDGen
	events *outbox.HandlerRegistry
	writer outbox.Writer
	jobs   *jobs.Registry
	boot   *bootState
}

func newModuleContext(name string, logger *slog.Logger, view config.ModuleView, deps moduleDeps) module.Context {
	if logger == nil {
		logger = slog.Default()
	}
	return &moduleContext{
		name: name, logger: logger.With("module", name), view: view,
		router: deps.router, val: deps.val, perms: deps.perms, rtypes: deps.rtypes,
		eval: deps.eval, tx: deps.tx, idgen: deps.idgen,
		events: deps.events, writer: deps.writer, jobs: deps.jobs, boot: deps.boot,
	}
}

func (c *moduleContext) Logger() *slog.Logger { return c.logger }

func (c *moduleContext) Config() config.ModuleView {
	if c.view == nil {
		return config.MapView{}
	}
	return c.view
}

func (c *moduleContext) Routes() *httpx.Router {
	if c.router == nil {
		c.router = httpx.NewRouter()
	}
	return c.router
}

func (c *moduleContext) Validator() *validation.Validator { return c.val }

func (c *moduleContext) Permissions() *authz.Registry {
	if c.perms == nil {
		c.perms = authz.NewRegistry()
	}
	return c.perms
}

func (c *moduleContext) Resources() *resource.Registry {
	if c.rtypes == nil {
		c.rtypes = resource.NewRegistry()
	}
	return c.rtypes
}

func (c *moduleContext) Authz() authz.Evaluator { return c.eval }

func (c *moduleContext) Tx() database.TxManager { return c.tx }

func (c *moduleContext) IDGen() model.IDGen {
	if c.idgen == nil {
		c.idgen = model.UUIDv7()
	}
	return c.idgen
}

// Events returns the shared event-subscription registry.
func (c *moduleContext) Events() *outbox.HandlerRegistry {
	if c.events == nil {
		c.events = outbox.NewHandlerRegistry()
	}
	return c.events
}

// Outbox returns the event writer for emitting events in a business tx.
func (c *moduleContext) Outbox() outbox.Writer {
	if c.writer == nil {
		c.writer = outbox.NewWriter(c.IDGen())
	}
	return c.writer
}

// Jobs returns the shared job-kind registry.
func (c *moduleContext) Jobs() *jobs.Registry {
	if c.jobs == nil {
		c.jobs = jobs.NewRegistry()
	}
	return c.jobs
}

func (c *moduleContext) Migrations(fsys fs.FS) { c.boot.migrations[c.name] = fsys }
func (c *moduleContext) Seeds(fsys fs.FS)      { c.boot.seeds[c.name] = fsys }
func (c *moduleContext) OpenAPI(fragment []byte) {
	c.boot.openapi[c.name] = fragment
}

func (c *moduleContext) Health(name string, check func(context.Context) error) {
	c.boot.health[c.name+"."+name] = check
}

// ProvidePort registers an impl under a module-prefixed name so dependents can
// fetch it. A name must be prefixed with the providing module's name.
func (c *moduleContext) ProvidePort(name string, impl any) {
	c.boot.ports[name] = impl
}

// Port fetches a previously-provided port. Because Register runs in dependency
// order, a dependency's ports are available to its dependents; a missing port
// is an error the module surfaces (and Validate re-checks declared needs).
func (c *moduleContext) Port(name string) (any, error) {
	p, ok := c.boot.ports[name]
	if !ok {
		return nil, fmt.Errorf("module %q: port %q is not provided by any registered dependency", c.name, name)
	}
	return p, nil
}
