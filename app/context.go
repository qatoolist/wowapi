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
	"time"

	"github.com/qatoolist/wowapi/kernel/artifact"
	"github.com/qatoolist/wowapi/kernel/attachment"
	kaudit "github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/bulk"
	"github.com/qatoolist/wowapi/kernel/comment"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/document"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/i18n"
	"github.com/qatoolist/wowapi/kernel/integration"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/notify"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/privileged"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/retention"
	"github.com/qatoolist/wowapi/kernel/rules"
	"github.com/qatoolist/wowapi/kernel/sequence"
	"github.com/qatoolist/wowapi/kernel/validation"
	"github.com/qatoolist/wowapi/kernel/webhook"
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
	// i18n aggregates the framework's English catalog plus every module's
	// localized bundles (GAP-001). Shared across module contexts; ownership
	// (module-prefixed keys) is enforced per Register and surfaced at boot.
	i18n *i18n.Registry
}

func newBootState() *bootState {
	return &bootState{
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
}

// moduleDeps bundles the shared registries/services the app injects into every
// module context, keeping the constructor signature stable as capabilities grow.
type moduleDeps struct {
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
}

func newModuleContext(name string, logger *slog.Logger, view config.ModuleView, deps moduleDeps) module.Context {
	if logger == nil {
		logger = slog.Default()
	}
	return &moduleContext{
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
		boot: deps.boot,
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
// collisions with kernel maintenance tasks and other modules.
func (c *moduleContext) RecurringJob(name string, every time.Duration, fn func(ctx context.Context, db database.TenantDB) error) {
	c.boot.recurring = append(c.boot.recurring, RecurringJob{
		Name:  c.name + "." + name,
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
// relationship types and rule keys prefixed "<name>."; a product that must widen
// this can build its own privileged.Services with a Config allow-list from its
// own wiring. In a process wired without a platform pool (the migrate process,
// which registers no modules that perform privileged writes) the underlying
// platform manager is nil, so a Grant/Revoke/ActivateTenant call would fail at
// invocation — but such a process never reaches those calls.
func (c *moduleContext) Privileged() *privileged.Services {
	if c.priv == nil {
		c.priv = privileged.New(c.name, c.platformTx, c.ruleStore, c.Audit(), c.IDGen(), privileged.Config{})
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

func (c *moduleContext) Migrations(fsys fs.FS) { c.boot.migrations[c.name] = fsys }
func (c *moduleContext) Seeds(fsys fs.FS)      { c.boot.seeds[c.name] = fsys }
func (c *moduleContext) OpenAPI(fragment []byte) {
	c.boot.openapi[c.name] = fragment
}

// I18n registers a module's localized message bundle under this module's name.
// Ownership (keys prefixed "<name>.") is enforced by the shared registry;
// violations accumulate and fail boot via the registry's Err().
func (c *moduleContext) I18n(bundle i18n.Bundle) { c.boot.i18n.Register(c.name, bundle) }

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
