// Module registration context (D-0006; blueprint 06 §2).
//
// Phase 1 delivers the capability-scoped context handed to Module.Register:
// a logger pre-tagged with the module name and the module's own config
// namespace — and nothing else. Each later phase adds one accessor alongside
// the kernel capability it delivers (Routes in Phase 3, Tx in Phase 2, …).
package app

import (
	"log/slog"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/validation"
	"github.com/qatoolist/wowapi/module"
)

// moduleContext implements module.Context. It is unexported; callers receive
// the interface value, keeping the concrete type an implementation detail.
type moduleContext struct {
	name   string
	logger *slog.Logger
	view   config.ModuleView
	router *httpx.Router
	val    *validation.Validator
	perms  *authz.Registry
	rtypes *resource.Registry
	eval   authz.Evaluator
}

// moduleDeps bundles the shared registries/services the app injects into every
// module context, keeping the constructor signature stable as capabilities grow.
type moduleDeps struct {
	router *httpx.Router
	val    *validation.Validator
	perms  *authz.Registry
	rtypes *resource.Registry
	eval   authz.Evaluator
}

// newModuleContext returns the capability-scoped context handed to
// Module.Register (D-0006/D-0032; capabilities widen per phase). The logger is
// tagged once here so Logger() is allocation-free on every call.
func newModuleContext(name string, logger *slog.Logger, view config.ModuleView, deps moduleDeps) module.Context {
	return &moduleContext{
		name: name, logger: logger.With("module", name), view: view,
		router: deps.router, val: deps.val, perms: deps.perms, rtypes: deps.rtypes, eval: deps.eval,
	}
}

// Logger returns a logger pre-tagged with the module name so every log line
// emitted from within Module.Register carries the module identity.
func (c *moduleContext) Logger() *slog.Logger {
	return c.logger
}

// Config returns the module's isolated config namespace (modules.<name>.*).
// If no namespace was provided for this module, an empty MapView is returned
// so modules with no product configuration can still decode their defaults
// cleanly without a nil-check.
func (c *moduleContext) Config() config.ModuleView {
	if c.view == nil {
		return config.MapView{}
	}
	return c.view
}

// Routes returns the module's route registry. Registration enforces route
// metadata (permission or explicit public); errors surface at boot via
// Router.Err() (blueprint 05 §1).
func (c *moduleContext) Routes() *httpx.Router {
	if c.router == nil {
		c.router = httpx.NewRouter()
	}
	return c.router
}

// Validator returns the shared request validator for BindAndValidate.
func (c *moduleContext) Validator() *validation.Validator {
	return c.val
}

// Permissions returns the permission registry the module declares into.
func (c *moduleContext) Permissions() *authz.Registry {
	if c.perms == nil {
		c.perms = authz.NewRegistry()
	}
	return c.perms
}

// Resources returns the resource-type registry the module declares into.
func (c *moduleContext) Resources() *resource.Registry {
	if c.rtypes == nil {
		c.rtypes = resource.NewRegistry()
	}
	return c.rtypes
}

// Authz returns the authorization evaluator (nil until the app wires it).
func (c *moduleContext) Authz() authz.Evaluator {
	return c.eval
}
