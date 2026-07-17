package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"time"

	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/i18n"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/seeds"
	"github.com/qatoolist/wowapi/kernel/validation"
)

// Booted is the result of App.Boot: everything the process layer needs to
// start serving (or to migrate/seed). Modules have registered; the whole graph
// and registries are validated; nothing has started yet.
type Booted struct {
	Kernel     *kernel.Kernel
	Router     *httpx.Router
	Events     *outbox.HandlerRegistry // event subscriptions (drives the relay)
	Jobs       *jobs.Registry          // job kinds (drives the worker pools)
	OpenAPI    map[string][]byte
	Health     map[string]func(context.Context) error
	Migrations map[string]fs.FS
	Seeds      seeds.Bundle   // merged catalog seeds, ready for SeedSync
	Recurring  []RecurringJob // module-registered recurring jobs (run by the worker scheduler)
	// I18n is the merged message catalog (framework English + every module's
	// localized bundles). Pass it to httpx.Locale so responses negotiate and
	// localize (GAP-001). Never nil — at minimum it carries the framework's
	// English catalog.
	I18n *i18n.Catalog
}

// RecurringJob is a leader-safe per-tenant recurring job a module registered via
// module.Context.RecurringJob (roadmap E5/CA-5). StartWorker registers each on
// the scheduler; Run is invoked once per active tenant every Every, in that
// tenant's transaction.
type RecurringJob struct {
	Name  string
	Every time.Duration
	Run   func(ctx context.Context, db database.TenantDB) error
}

// BootOption tunes Boot. See SkipRLSEnforcementCheck.
type BootOption func(*bootOpts)

type bootOpts struct {
	skipRLSCheck bool
	i18nLayers   []i18n.Layer
}

// WithI18nLayers supplies product-configured i18n source layers (framework
// overrides, product/module catalog files, compiled Go bundles) to merge into
// the catalog after modules register and before it is frozen for serving. The
// generated api, worker, AND migrate binaries pass the SAME layers (resolved
// from the product's i18n config), so all three load one catalog through one
// lifecycle (B1 acceptance). Layers are applied in precedence order on top of
// the framework's embedded defaults; ownership violations fail boot like any
// other registration error. Omit it (zero-config) and boot ships the framework
// English catalog exactly as before.
func WithI18nLayers(layers ...i18n.Layer) BootOption {
	return func(o *bootOpts) { o.i18nLayers = append(o.i18nLayers, layers...) }
}

// SkipRLSEnforcementCheck disables the boot-time assertion that the runtime pool
// cannot bypass row-level security. Use it ONLY for a process that does not serve
// tenant traffic and runs as a privileged role by design — namely the migrate
// command, which boots the app merely to COLLECT module migration sets and
// connects with DDL (app_migrate/superuser) credentials. api/worker processes must
// NOT use it: their tenant-serving runtime pool must be a non-privileged app_rt
// role, and the default check keeps that safe-by-default (finding M3).
func SkipRLSEnforcementCheck() BootOption { return func(o *bootOpts) { o.skipRLSCheck = true } }

// Boot runs the module lifecycle up to (not including) Start: it registers every
// module against a capability-scoped context built from k — in dependency order
// so a module's ports are available to its dependents — then validates the whole
// graph and the shared registries (blueprint 06 §2). Boot fails, before anything
// serves, on: a module graph error (dup/unknown/cycle), a registration error, a
// route whose permission is not registered, a duplicate/invalid permission or
// resource type, or a seed ownership/parse error.
//
// namespaces is the loaded product config's module.* subtree; each module sees
// only its own slice via Context.Config().
func (a *App) Boot(ctx context.Context, k *kernel.Kernel, namespaces config.Namespaces, opts ...BootOption) (*Booted, error) {
	var bo bootOpts
	for _, o := range opts {
		o(&bo)
	}

	ordered, err := a.validateAndOrder()
	if err != nil {
		return nil, err
	}

	// Safe-by-default RLS enforcement (finding M3): fail boot if a pool that serves
	// data runs as a superuser / BYPASSRLS role, which would silently defeat FORCE
	// RLS. This covers BOTH the tenant-serving runtime pool AND the platform pool —
	// the platform pool does all cross-tenant kernel work (job runner, outbox relay,
	// webhook dispatch) over FORCE-RLS tables and relies on app_platform being a
	// non-privileged role served by permissive policies; a superuser platform DSN
	// would bypass those policies with no signal. Backstops the per-connection
	// (WithConnRLSGuard) and per-tx (WithRLSGuard) guards. Skipped only for
	// non-serving processes (migrate) that run privileged by design.
	if !bo.skipRLSCheck {
		if k.Pool != nil {
			if err := database.AssertRLSEnforced(ctx, k.Pool); err != nil {
				return nil, err
			}
		}
		if k.Platform != nil {
			if err := database.AssertRLSEnforced(ctx, k.Platform); err != nil {
				return nil, fmt.Errorf("platform pool: %w", err)
			}
		}
	}

	boot := newBootState()
	router := httpx.NewRouter()
	// FBL-08: boot-time request-contract enforcement is profile-flag gated
	// (compat-first). The mode must be set BEFORE modules register, since
	// Router.Handle applies the check at registration time; a rejected route
	// surfaces through router.Err() below with every other registration error.
	if k.Cfg.Security.EnforceRouteContracts {
		router.RequireRequestContracts()
	}
	val := validation.New()
	idgen := model.UUIDv7()
	events := outbox.NewHandlerRegistry()
	writer := outbox.NewWriter(idgen)
	jobReg := jobs.NewRegistry()

	var regErrs []error
	knownModules := make(map[string]struct{}, len(ordered))
	for _, m := range ordered {
		var view config.ModuleView
		if namespaces != nil {
			if v, ok := namespaces[m.Name()]; ok {
				view = v
			}
		}
		knownModules[m.Name()] = struct{}{}
		mc := newModuleContext(m.Name(), k.Log, view, moduleDeps{
			dependsOn: m.DependsOn(),
			router:    router, val: val, perms: k.Perms, rtypes: k.Resources,
			eval: k.Authz, tx: k.Tx, idgen: idgen,
			events: events, writer: writer, jobs: jobReg,
			rules: k.Rules, resolver: k.RulesResolver, wfReg: k.Workflows, wfRT: k.WorkflowRuntime,
			retClass: k.RetentionClasses,
			docClass: k.DocumentClasses, docHooks: k.DocumentHooks, docs: k.Documents,
			comments: k.Comments, attaches: k.Attachments,
			notifyReg: k.NotifyTemplates, notifySvc: k.Notify, webhooks: k.Webhooks,
			intReg: k.IntegrationProviders, intStore: k.Integrations,
			audit: k.Audit, sequence: k.Sequence, bulk: k.Bulk, artifacts: k.Artifacts,
			platformTx: k.PlatformTx, ruleStore: k.RuleStore, privCfg: k.Cfg.Privileged,
			boot: boot,
		})
		if err := m.Register(mc); err != nil {
			regErrs = append(regErrs, fmt.Errorf("module %q: Register: %w", m.Name(), err))
		}
	}

	// Compile and seal the extension model (F-10): ownership, duplicate, type,
	// and requirement violations collected through the appmodel compiler fail
	// boot here, and the sealed flag makes retained module contexts immutable —
	// runtime extensions go through the ownership-bound compiler, not bare maps.
	regErrs = append(regErrs, boot.portErrs...)
	if _, err := boot.compiler.Compile(); err != nil {
		regErrs = append(regErrs, fmt.Errorf("extension model: %w", err))
	}
	boot.sealed = true

	// Reject unknown module namespaces (AR-04 T1): a config `modules.<name>`
	// namespace with no corresponding registered module is otherwise retained as
	// opaque, unvalidated data and never rejected — silently masking a typo (e.g.
	// modules.polcy) or a stale namespace left behind by a removed module. Sort
	// the offending keys so the error is deterministic (ARCH-52).
	if len(namespaces) > 0 {
		var unknown []string
		for name := range namespaces {
			if _, ok := knownModules[name]; !ok {
				unknown = append(unknown, name)
			}
		}
		if len(unknown) > 0 {
			sort.Strings(unknown)
			regErrs = append(regErrs, fmt.Errorf("config: unknown module namespace(s) %v: no registered module matches", unknown))
		}
	}

	// Load and merge each module's seed bundle (strict, ownership-checked), and
	// register seed-declared permissions into the shared registry so the
	// evaluator recognizes them. Iterate module names in sorted order so the
	// merged bundle and any error messages are deterministic (ARCH-52).
	var bundle seeds.Bundle
	seedVersionSource := ""
	seedModules := make([]string, 0, len(boot.seeds))
	for name := range boot.seeds {
		seedModules = append(seedModules, name)
	}
	sort.Strings(seedModules)
	for _, name := range seedModules {
		fsys := boot.seeds[name]
		b, err := seeds.Load(fsys, name)
		if err != nil {
			regErrs = append(regErrs, err)
			continue
		}
		bundle.Permissions = append(bundle.Permissions, b.Permissions...)
		bundle.Roles = append(bundle.Roles, b.Roles...)
		bundle.ResourceTypes = append(bundle.ResourceTypes, b.ResourceTypes...)
		bundle.RelationshipTypes = append(bundle.RelationshipTypes, b.RelationshipTypes...)
		if b.Version != "" {
			if bundle.Version != "" && bundle.Version != b.Version {
				regErrs = append(regErrs, fmt.Errorf("module %q seed version %q conflicts with module %q seed version %q", name, b.Version, seedVersionSource, bundle.Version))
			} else {
				bundle.Version = b.Version
				seedVersionSource = name
			}
		}
		for _, p := range b.Permissions {
			perm := authz.Permission{Key: p.Key, Sensitive: p.Sensitive, GrantedVia: p.GrantedVia, StepUp: p.StepUp}
			// A richer step-up seed form (specific AMR subset and/or challenge
			// hint) becomes a StepUpPolicy on the registry entry — registry-
			// declared, in-memory only (not persisted; permissions.step_up stays
			// a plain bool). seeds.validate already rejected these fields set
			// without step_up: true, so p.StepUp is true here whenever either is
			// non-empty (B8).
			if len(p.StepUpAMR) > 0 || p.StepUpChallenge != "" {
				perm.StepUpPolicy = &authz.StepUpPolicy{RequiredAMR: p.StepUpAMR, Challenge: p.StepUpChallenge}
			}
			k.Perms.Register(perm)
		}
		for _, rt := range b.ResourceTypes {
			k.Resources.Register(name, resource.TypeSpec{Key: rt.Key, Description: rt.Description})
		}
	}

	// Whole-graph validation gates (all accumulated, boot fails with the full list).
	if err := k.Perms.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := k.Resources.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := router.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := events.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := jobReg.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := k.Rules.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := k.Workflows.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := k.DocumentClasses.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	// A module that registered a document class needs a document service to use
	// it; the service is nil when no object-storage adapter was wired. Fail boot
	// loudly rather than hand modules a nil Documents() at runtime.
	if len(k.DocumentClasses.Keys()) > 0 && k.Documents == nil {
		regErrs = append(regErrs, fmt.Errorf("document classes are registered (%v) but no storage adapter is wired: pass kernel.Deps.Storage", k.DocumentClasses.Keys()))
	}
	if err := k.NotifyTemplates.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	if err := k.IntegrationProviders.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	// Merge product-configured i18n source layers (framework overrides, product/
	// module catalog files, Go bundles) on top of the framework defaults and the
	// module bundles registered above, in precedence order. Ownership/duplicate
	// violations are recorded on the registry and surface via Err() below, so a
	// bad catalog fails boot with every other registration error.
	if len(bo.i18nLayers) > 0 {
		boot.i18n.ApplyLayers(bo.i18nLayers...)
	}
	// i18n bundle ownership (module-prefixed keys, no reserved kernel.* shadowing)
	// is boot-validated like every other registry.
	if err := boot.i18n.Err(); err != nil {
		regErrs = append(regErrs, err)
	}
	// Every route's permission must be a registered permission (deny-by-default
	// depends on the registry knowing it; an unknown permission is a boot bug).
	for _, p := range router.Permissions() {
		if !k.Perms.Has(p) {
			regErrs = append(regErrs, fmt.Errorf("route permission %q is not declared by any module seed or registration", p))
		}
	}

	if len(regErrs) > 0 {
		return nil, fmt.Errorf("app: boot validation failed: %w", errors.Join(regErrs...))
	}

	// Seal the catalog for request-time reads (Decision 3): every source and
	// module bundle has merged and validated, so no further writes are legitimate.
	// After Freeze, Catalog.Add is a no-op, so request-path Lookups never race a
	// write and a post-boot mutation cannot silently change served strings.
	boot.i18n.Freeze()

	// Seal every extension registry (closure review 2026-07-17, F-10): the
	// extension model is registration-at-boot only. Booted intentionally hands
	// out the live Router/Events/Jobs pointers for serving, and retained module
	// contexts still reference the shared kernel registries — from here on every
	// registration mutator on them panics, so neither path can add a route, job
	// kind, subscription, permission, resource type, rule point, workflow,
	// record class, document class/hook, template, or provider after the boot
	// gates above have validated the model. (RetentionClasses/DocumentHooks are
	// nil-guarded: unlike the others they are not required by the Err() gates.)
	router.Seal()
	events.Seal()
	jobReg.Seal()
	k.Perms.Seal()
	k.Resources.Seal()
	k.Rules.Seal()
	k.Workflows.Seal()
	k.DocumentClasses.Seal()
	k.NotifyTemplates.Seal()
	k.IntegrationProviders.Seal()
	if k.RetentionClasses != nil {
		k.RetentionClasses.Seal()
	}
	if k.DocumentHooks != nil {
		k.DocumentHooks.Seal()
	}

	// Booted carries SNAPSHOTS of the boot-time collectors, not the backing
	// structures (closure review 2026-07-17, F-10): the maps/slice handed out
	// here are fresh copies, so neither a retained module context nor a caller
	// mutating Booted's fields can alter the sealed extension state the runtime
	// (health handler, worker scheduler, migration runner) actually consumes —
	// and the reverse aliasing (Booted mutation racing a live reader) is
	// structurally impossible.
	openapi := make(map[string][]byte, len(boot.openapi))
	for k, v := range boot.openapi {
		openapi[k] = append([]byte(nil), v...)
	}
	health := make(map[string]func(context.Context) error, len(boot.health))
	for k, v := range boot.health {
		health[k] = v
	}
	migrationsFS := make(map[string]fs.FS, len(boot.migrations))
	for k, v := range boot.migrations {
		migrationsFS[k] = v
	}
	recurring := append([]RecurringJob(nil), boot.recurring...)

	return &Booted{
		Kernel:     k,
		Router:     router,
		Events:     events,
		Jobs:       jobReg,
		OpenAPI:    openapi,
		Health:     health,
		Migrations: migrationsFS,
		Seeds:      bundle,
		Recurring:  recurring,
		I18n:       boot.i18n.Catalog(),
	}, nil
}
