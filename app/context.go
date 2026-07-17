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
	"reflect"
	"strings"
	"time"

	"github.com/qatoolist/wowapi/foundation/artifact"
	"github.com/qatoolist/wowapi/foundation/attachment"
	"github.com/qatoolist/wowapi/foundation/bulk"
	"github.com/qatoolist/wowapi/foundation/comment"
	"github.com/qatoolist/wowapi/foundation/document"
	"github.com/qatoolist/wowapi/foundation/integration"
	"github.com/qatoolist/wowapi/foundation/notify"
	"github.com/qatoolist/wowapi/foundation/webhook"
	"github.com/qatoolist/wowapi/kernel/appmodel"
	kaudit "github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/i18n"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/privileged"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/retention"
	"github.com/qatoolist/wowapi/kernel/rules"
	"github.com/qatoolist/wowapi/kernel/sequence"
	"github.com/qatoolist/wowapi/kernel/validation"
	"github.com/qatoolist/wowapi/kernel/workflow"
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
	recurring  []RecurringJob
	// compiler is the ownership-bound extension compiler (kernel/appmodel):
	// every runtime ProvidePort routes through it so owner-prefix, duplicate,
	// nil/type, and post-seal violations are enforced at boot (adversarial
	// review 2026-07-17, F-10 — the raw map alone accepted all of them).
	compiler *appmodel.Compiler
	// portErrs accumulates extension-contract violations; Boot fails on any.
	portErrs []error
	// sealed flips after Boot compiles the model: a retained module context
	// must not mutate extensions post-boot.
	sealed bool
	// i18n aggregates the framework's English catalog plus every module's
	// localized bundles (GAP-001). Shared across module contexts; ownership
	// (module-prefixed keys) is enforced per Register and surfaced at boot.
	i18n *i18n.Registry
}

func newBootState() *bootState {
	return &bootState{
		compiler:   appmodel.NewCompiler(),
		migrations: map[string]fs.FS{},
		seeds:      map[string]fs.FS{},
		openapi:    map[string][]byte{},
		health:     map[string]func(context.Context) error{},
		ports:      map[string]any{},
		i18n:       i18n.NewRegistry(),
	}
}

// moduleContext implements module.Context. Unexported; callers receive the
// interface value.
type moduleContext struct {
	registrar appmodel.Registrar[any]
	depSet    map[string]struct{}
	name      string
	logger    *slog.Logger
	view      config.ModuleView
	router    *httpx.Router
	val       *validation.Validator
	perms     *authz.Registry
	rtypes    *resource.Registry
	eval      authz.Evaluator
	tx        database.TxManager
	idgen     model.IDGen
	events    *outbox.HandlerRegistry
	writer    outbox.Writer
	jobs      *jobs.Registry
	rules     *rules.Registry
	resolver  *rules.Resolver
	wfReg     *workflow.Registry
	wfRT      *workflow.Runtime
	retClass  *retention.Registry
	docClass  *document.Registry
	docHooks  *document.Hooks
	docs      *document.Service
	comments  *comment.Service
	attaches  *attachment.Service
	notifyReg *notify.Registry
	notifySvc *notify.Service
	webhooks  *webhook.Service
	intReg    *integration.Registry
	intStore  *integration.Store
	audit     *kaudit.Writer
	sequence  *sequence.Allocator
	bulk      *bulk.Service
	artifacts *artifact.Pipeline
	boot      *bootState

	// privileged deps (GAP-006): the tenant-bindable app_platform manager and rule
	// store backing the scoped privileged services. priv is the per-module Services
	// value, built lazily on first Privileged() call from these shared deps.
	platformTx database.TxManager
	ruleStore  *rules.Store
	priv       *privileged.Services
	// privCfg is this module's boot-validated allow-list widening (backlog B10,
	// config.Framework.Privileged), or the zero value when the product declares
	// none for this module — in which case Privileged() behaves EXACTLY as
	// before (prefix-ownership only).
	privCfg config.PrivilegedGrant
}

// moduleDeps bundles the shared registries/services the app injects into every
// module context, keeping the constructor signature stable as capabilities grow.
type moduleDeps struct {
	dependsOn []string
	router    *httpx.Router
	val       *validation.Validator
	perms     *authz.Registry
	rtypes    *resource.Registry
	eval      authz.Evaluator
	tx        database.TxManager
	idgen     model.IDGen
	events    *outbox.HandlerRegistry
	writer    outbox.Writer
	jobs      *jobs.Registry
	rules     *rules.Registry
	resolver  *rules.Resolver
	wfReg     *workflow.Registry
	wfRT      *workflow.Runtime
	retClass  *retention.Registry
	docClass  *document.Registry
	docHooks  *document.Hooks
	docs      *document.Service
	comments  *comment.Service
	attaches  *attachment.Service
	notifyReg *notify.Registry
	notifySvc *notify.Service
	webhooks  *webhook.Service
	intReg    *integration.Registry
	intStore  *integration.Store
	audit     *kaudit.Writer
	sequence  *sequence.Allocator
	bulk      *bulk.Service
	artifacts *artifact.Pipeline
	boot      *bootState

	platformTx database.TxManager
	ruleStore  *rules.Store
	// privCfg carries each module's boot-validated allow-list widening
	// (backlog B10, config.Framework.Privileged) keyed by module name; a
	// module absent from the map gets the zero value (prefix-ownership only,
	// unchanged from before this config section existed).
	privCfg config.Privileged
}

func newModuleContext(name string, logger *slog.Logger, view config.ModuleView, deps moduleDeps) module.Context {
	if logger == nil {
		logger = slog.Default()
	}
	if deps.boot == nil {
		// Direct constructions (unit tests, tools) get a self-contained boot
		// state; App.Boot always supplies the shared one.
		deps.boot = newBootState()
	}
	depSet := make(map[string]struct{}, len(deps.dependsOn))
	for _, d := range deps.dependsOn {
		depSet[d] = struct{}{}
	}
	return &moduleContext{
		registrar: deps.boot.compiler.GetRegistrar(name), depSet: depSet,
		name: name, logger: logger.With("module", name), view: view,
		router: deps.router, val: deps.val, perms: deps.perms, rtypes: deps.rtypes,
		eval: deps.eval, tx: deps.tx, idgen: deps.idgen,
		events: deps.events, writer: deps.writer, jobs: deps.jobs,
		rules: deps.rules, resolver: deps.resolver, wfReg: deps.wfReg, wfRT: deps.wfRT,
		retClass: deps.retClass,
		docClass: deps.docClass, docHooks: deps.docHooks, docs: deps.docs,
		comments: deps.comments, attaches: deps.attaches,
		notifyReg: deps.notifyReg, notifySvc: deps.notifySvc, webhooks: deps.webhooks,
		intReg: deps.intReg, intStore: deps.intStore,
		audit: deps.audit, sequence: deps.sequence, bulk: deps.bulk, artifacts: deps.artifacts,
		platformTx: deps.platformTx, ruleStore: deps.ruleStore,
		privCfg: deps.privCfg[name],
		boot:    deps.boot,
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

// RecurringJob collects a leader-safe per-tenant recurring job; the worker's
// scheduler runs it (roadmap E5/CA-5). The name is module-prefixed to avoid
// collisions with kernel maintenance tasks and other modules. Declarations are
// boot-validated (second closure audit 2026-07-17, F-10): an empty name, a
// nonpositive interval, a nil callback, or a duplicate full name is a
// collected boot error — a duplicate would silently share one scheduler row
// (one declaration advances the schedule while the other starves), and a nil
// callback would panic only when first due.
func (c *moduleContext) RecurringJob(name string, every time.Duration, fn func(ctx context.Context, db database.TenantDB) error) {
	c.mustBeUnsealed("RecurringJob")
	full := c.name + "." + name
	switch {
	case name == "":
		c.boot.portErrs = append(c.boot.portErrs,
			fmt.Errorf("module %q: RecurringJob requires a non-empty name", c.name))
		return
	case every <= 0:
		c.boot.portErrs = append(c.boot.portErrs,
			fmt.Errorf("module %q: recurring job %q has nonpositive interval %v", c.name, full, every))
		return
	case fn == nil:
		c.boot.portErrs = append(c.boot.portErrs,
			fmt.Errorf("module %q: recurring job %q has a nil callback (would panic when due)", c.name, full))
		return
	}
	for _, existing := range c.boot.recurring {
		if existing.Name == full {
			c.boot.portErrs = append(c.boot.portErrs,
				fmt.Errorf("module %q: recurring job %q declared more than once (duplicates share one scheduler row and starve each other)", c.name, full))
			return
		}
	}
	c.boot.recurring = append(c.boot.recurring, RecurringJob{
		Name:  full,
		Every: every,
		Run:   fn,
	})
}

// Rules returns the rule-point registry; RulesResolver the resolver.
func (c *moduleContext) Rules() *rules.Registry         { return c.rules }
func (c *moduleContext) RulesResolver() *rules.Resolver { return c.resolver }

// Workflows returns the workflow registry; WorkflowRuntime the runtime.
func (c *moduleContext) Workflows() *workflow.Registry      { return c.wfReg }
func (c *moduleContext) WorkflowRuntime() *workflow.Runtime { return c.wfRT }

func (c *moduleContext) RetentionClasses() *retention.Registry { return c.retClass }

// Evidence-layer services (roadmap CA-11).
func (c *moduleContext) Audit() *kaudit.Writer         { return c.audit }
func (c *moduleContext) Sequence() *sequence.Allocator { return c.sequence }
func (c *moduleContext) Bulk() *bulk.Service           { return c.bulk }
func (c *moduleContext) Artifacts() *artifact.Pipeline { return c.artifacts }

// Privileged returns the module's scoped privileged-service surface (GAP-006),
// built once and bound to THIS module's name so ownership is enforced against
// the calling module. Prefix-ownership only by default: the module may manage
// relationship types and rule keys prefixed "<name>.". A product config can
// widen this per module (backlog B10) via config.Framework.Privileged, an
// explicit allow-list of concrete relationship types / rule keys — boot-
// validated to reject wildcards/globs/empty entries (config.Privileged.Validate,
// fail closed). A module the product declares no allow-list for keeps EXACTLY
// today's prefix-only behavior; one module's allow-list never widens another's
// (privCfg is looked up per module name in newModuleContext). In a process
// wired without a platform pool (the migrate process, which registers no
// modules that perform privileged writes) the underlying platform manager is
// nil, so a Grant/Revoke/ActivateTenant call would fail at invocation — but
// such a process never reaches those calls.
func (c *moduleContext) Privileged() *privileged.Services {
	if c.priv == nil {
		c.priv = privileged.New(c.name, c.platformTx, c.ruleStore, c.Audit(), c.IDGen(), privileged.Config{
			AllowRelTypes: c.privCfg.AllowRelTypes,
			AllowRuleKeys: c.privCfg.AllowRuleKeys,
		})
	}
	return c.priv
}

// DocumentClasses/DocumentHooks are the shared document registration pointers;
// Documents/Comments/Attachments are the runtime services.
func (c *moduleContext) DocumentClasses() *document.Registry { return c.docClass }
func (c *moduleContext) DocumentHooks() *document.Hooks      { return c.docHooks }
func (c *moduleContext) Documents() *document.Service        { return c.docs }
func (c *moduleContext) Comments() *comment.Service          { return c.comments }
func (c *moduleContext) Attachments() *attachment.Service    { return c.attaches }

// NotifyTemplates/Notify/Webhooks/IntegrationProviders/Integrations expose the
// Phase 9 notification, webhook, and integration framework.
func (c *moduleContext) NotifyTemplates() *notify.Registry           { return c.notifyReg }
func (c *moduleContext) Notify() *notify.Service                     { return c.notifySvc }
func (c *moduleContext) Webhooks() *webhook.Service                  { return c.webhooks }
func (c *moduleContext) IntegrationProviders() *integration.Registry { return c.intReg }
func (c *moduleContext) Integrations() *integration.Store            { return c.intStore }

// mustBeUnsealed guards every boot-time collector: a retained module context
// must not mutate extension state after Boot has compiled and sealed the model
// (F-10). Health is the sharpest case — its map is read by the live health
// handler, so a post-boot write is also a concurrent map hazard.
func (c *moduleContext) mustBeUnsealed(what string) {
	if c.boot.sealed {
		panic(fmt.Sprintf("module %q: %s after boot: the extension model is sealed", c.name, what))
	}
}

// rejectDuplicate accumulates a boot error when a module registers the same
// collector twice — the second call previously overwrote the first silently
// (closure review 2026-07-17, F-10).
func (c *moduleContext) rejectDuplicate(what string, exists bool) bool {
	if exists {
		c.boot.portErrs = append(c.boot.portErrs,
			fmt.Errorf("module %q: duplicate %s registration (would silently overwrite the first)", c.name, what))
	}
	return exists
}

func (c *moduleContext) Migrations(fsys fs.FS) {
	c.mustBeUnsealed("Migrations")
	if fsys == nil {
		c.boot.portErrs = append(c.boot.portErrs,
			fmt.Errorf("module %q: Migrations registered a nil fs.FS", c.name))
		return
	}
	if _, dup := c.boot.migrations[c.name]; c.rejectDuplicate("Migrations", dup) {
		return
	}
	c.boot.migrations[c.name] = fsys
}

func (c *moduleContext) Seeds(fsys fs.FS) {
	c.mustBeUnsealed("Seeds")
	if fsys == nil {
		c.boot.portErrs = append(c.boot.portErrs,
			fmt.Errorf("module %q: Seeds registered a nil fs.FS", c.name))
		return
	}
	if _, dup := c.boot.seeds[c.name]; c.rejectDuplicate("Seeds", dup) {
		return
	}
	c.boot.seeds[c.name] = fsys
}

func (c *moduleContext) OpenAPI(fragment []byte) {
	c.mustBeUnsealed("OpenAPI")
	if _, dup := c.boot.openapi[c.name]; c.rejectDuplicate("OpenAPI", dup) {
		return
	}
	c.boot.openapi[c.name] = fragment
}

// I18n registers a module's localized message bundle under this module's name.
// Ownership (keys prefixed "<name>.") is enforced by the shared registry;
// violations accumulate and fail boot via the registry's Err().
func (c *moduleContext) I18n(bundle i18n.Bundle) {
	c.mustBeUnsealed("I18n")
	c.boot.i18n.Register(c.name, bundle)
}

func (c *moduleContext) Health(name string, check func(context.Context) error) {
	c.mustBeUnsealed("Health")
	if name == "" {
		c.boot.portErrs = append(c.boot.portErrs,
			fmt.Errorf("module %q: Health requires a non-empty check name", c.name))
		return
	}
	if check == nil {
		c.boot.portErrs = append(c.boot.portErrs,
			fmt.Errorf("module %q: health check %q has a nil func (would panic when probed)", c.name, name))
		return
	}
	key := c.name + "." + name
	if _, dup := c.boot.health[key]; c.rejectDuplicate("Health("+name+")", dup) {
		return
	}
	c.boot.health[key] = check
}

// ProvidePort registers an impl under a module-prefixed name so dependents can
// fetch it. A name must be prefixed with the providing module's name; ownership,
// duplicates, nil implementations, and post-boot mutation are enforced through
// the appmodel compiler and fail Boot (F-10) — never silently overwritten.
func (c *moduleContext) ProvidePort(name string, impl any) {
	if c.boot.sealed {
		panic(fmt.Sprintf("module %q: ProvidePort(%q) after boot: the extension model is sealed", c.name, name))
	}
	if !strings.HasPrefix(name, c.name+".") {
		c.boot.portErrs = append(c.boot.portErrs,
			fmt.Errorf("module %q: ProvidePort(%q): a port must be prefixed with its providing module's name", c.name, name))
		return
	}
	if impl == nil {
		c.boot.portErrs = append(c.boot.portErrs,
			fmt.Errorf("module %q: ProvidePort(%q): nil implementation", c.name, name))
		return
	}
	// A typed nil ((*T)(nil), nil map/func/chan/slice) passes impl == nil but
	// panics at first use — reject it at boot like an untyped nil (closure
	// review 2026-07-17, F-10).
	if v := reflect.ValueOf(impl); (v.Kind() == reflect.Ptr || v.Kind() == reflect.Map ||
		v.Kind() == reflect.Slice || v.Kind() == reflect.Func || v.Kind() == reflect.Chan ||
		v.Kind() == reflect.Interface) && v.IsNil() {
		c.boot.portErrs = append(c.boot.portErrs,
			fmt.Errorf("module %q: ProvidePort(%q): typed-nil implementation (%T)", c.name, name, impl))
		return
	}
	t := reflect.TypeOf(impl)
	if err := c.registrar.DefinePort(name, t); err != nil {
		c.boot.portErrs = append(c.boot.portErrs, fmt.Errorf("module %q: ProvidePort(%q): %w", c.name, name, err))
		return
	}
	if err := c.registrar.ProvidePort(name, impl, t); err != nil {
		c.boot.portErrs = append(c.boot.portErrs, fmt.Errorf("module %q: ProvidePort(%q): %w", c.name, name, err))
		return
	}
	c.boot.ports[name] = impl
}

// Port fetches a previously-provided port. Because Register runs in dependency
// order, a dependency's ports are available to its dependents. The provider
// (the port name's module prefix) must be this module or one of its DECLARED
// dependencies (F-10) — resolution is not a global grab-bag — and each resolve
// is recorded as a requirement the compiled model re-validates at boot.
func (c *moduleContext) Port(name string) (any, error) {
	if c.boot.sealed {
		return nil, fmt.Errorf("module %q: Port(%q) after boot: the extension model is sealed", c.name, name)
	}
	// Every resolution failure is BOTH returned to the module AND accumulated
	// into boot validation (closure review 2026-07-17, F-10): a module that
	// ignores the error and returns nil from Register must still fail boot —
	// "unsatisfied dependency fails boot" is a boot contract, not a courtesy
	// return value.
	fail := func(err error) (any, error) {
		c.boot.portErrs = append(c.boot.portErrs, err)
		return nil, err
	}
	provider, _, ok := strings.Cut(name, ".")
	if !ok {
		return fail(fmt.Errorf("module %q: port %q is not module-prefixed", c.name, name))
	}
	if provider != c.name {
		if _, declared := c.depSet[provider]; !declared {
			return fail(fmt.Errorf("module %q: port %q belongs to module %q, which is not a declared dependency", c.name, name, provider))
		}
	}
	p, ok := c.boot.ports[name]
	if !ok {
		return fail(fmt.Errorf("module %q: port %q is not provided by any registered dependency", c.name, name))
	}
	if err := c.registrar.RequirePort(name, reflect.TypeOf(p)); err != nil {
		return fail(fmt.Errorf("module %q: Port(%q): %w", c.name, name, err))
	}
	return p, nil
}
