package lifecycle

// CurrentManifest is the hand-maintained descriptor graph for wowapi's ACTUAL
// wiring, read from kernel.New (kernel/kernel.go), app.Boot/moduleContext
// (app/context.go, app/boot.go) and module.Context (module/module.go) as of
// this package's introduction (backlog B9). It is not generated: whoever adds
// a field to Kernel/moduleContext/moduleDeps/module.Context should add or
// update the matching descriptor(s) here in the same change — the CI gate
// (`wowapi lint lifecycle`, and TestCurrentManifestLintsClean) exists so a
// forgotten or wrong update fails loudly rather than silently drifting.
//
// Naming convention: "kernel.<Field>" for a Kernel struct field built in
// kernel.New; "module.Context.<Method>" for an accessor module.Context
// exposes (all of which are APIRuntime: true, since module.Register runs in
// every product's api/worker boot path).
func CurrentManifest() Manifest {
	return Manifest{Descriptors: []ProviderDescriptor{
		// --- kernel.New: process-scoped infrastructure (kernel/kernel.go) ---

		// The runtime pool. RawPool: true and APIRuntime: false — kernel.Kernel.Pool
		// is never exposed through module.Context (module/module.go has no Pool()
		// accessor); only kernel.Tx (a TxManager wrapping it) reaches modules.
		{Provides: "kernel.Pool", Scope: ScopeProcess, RawPool: true},
		// The cross-tenant platform pool (Deps.Platform). Same raw-pool posture as
		// kernel.Pool: kernel.Kernel.Platform is nil for the migrate process and,
		// like Pool, is never handed to a module directly — only kernel.PlatformTx
		// (wrapped) reaches modules via Context.Privileged().
		{Provides: "kernel.Platform", Scope: ScopeProcess, RawPool: true},

		// database.TxManager built by the product main / testkit and injected via
		// Deps.Tx — the ONLY door to tenant data a module ever receives.
		{Provides: "kernel.Tx", Requires: []string{"kernel.Pool"}, Scope: ScopeProcess},
		// PlatformTx: a tenant-bindable TxManager over the app_platform pool
		// (WithRole app_platform + WithRLSGuard), built in kernel.New only when
		// Deps.Platform is non-nil. Backs kernel/privileged's scoped services.
		{Provides: "kernel.PlatformTx", Requires: []string{"kernel.Platform"}, Scope: ScopeProcess},

		{Provides: "kernel.Perms", Scope: ScopeProcess},
		{Provides: "kernel.Resources", Scope: ScopeProcess},
		{Provides: "kernel.Rules", Scope: ScopeProcess},
		{Provides: "kernel.RulesResolver", Requires: []string{"kernel.Rules"}, Scope: ScopeProcess},
		{Provides: "kernel.RuleStore", Requires: []string{"kernel.Rules"}, Scope: ScopeProcess},
		{Provides: "kernel.Workflows", Scope: ScopeProcess},
		{Provides: "kernel.WorkflowRuntime", Requires: []string{"kernel.Tx", "kernel.Workflows", "kernel.Authz"}, Scope: ScopeProcess},
		{Provides: "kernel.RetentionClasses", Scope: ScopeProcess},
		{Provides: "kernel.Retention", Requires: []string{"kernel.RetentionClasses"}, Scope: ScopeProcess},

		// Authz: the evaluator wraps a Store (optionally AuthzCache-wrapped) over
		// the shared Perms registry.
		{Provides: "kernel.Authz", Requires: []string{"kernel.Perms"}, Scope: ScopeProcess},
		{Provides: "kernel.AuthzCache", Scope: ScopeProcess},

		// Document / comment / attachment framework. kernel.Documents is nil unless
		// Deps.Storage is wired (an api-only process may run without object
		// storage) — modeled as a dependency on a storage adapter capability.
		{Provides: "kernel.DocumentClasses", Scope: ScopeProcess},
		{Provides: "kernel.DocumentHooks", Scope: ScopeProcess},
		{Provides: "kernel.storage.Adapter", Scope: ScopeProcess},
		{Provides: "kernel.Documents", Requires: []string{"kernel.DocumentClasses", "kernel.storage.Adapter", "kernel.Authz"}, Scope: ScopeProcess},
		{Provides: "kernel.Comments", Scope: ScopeProcess},
		{Provides: "kernel.Attachments", Scope: ScopeProcess},

		// Notify / webhook / integration framework.
		{Provides: "kernel.NotifyTemplates", Scope: ScopeProcess},
		{Provides: "kernel.Notify", Requires: []string{"kernel.NotifyTemplates"}, Scope: ScopeProcess},
		{Provides: "kernel.Webhooks", Scope: ScopeProcess},
		{Provides: "kernel.IntegrationProviders", Scope: ScopeProcess},
		{Provides: "kernel.Integrations", Requires: []string{"kernel.IntegrationProviders"}, Scope: ScopeProcess},

		{Provides: "kernel.Metrics", Scope: ScopeProcess},
		{Provides: "kernel.Tracer", Scope: ScopeProcess},

		// Evidence-layer services (roadmap CA-11).
		{Provides: "kernel.Audit", Scope: ScopeProcess},
		{Provides: "kernel.Sequence", Scope: ScopeProcess},
		{Provides: "kernel.Bulk", Scope: ScopeProcess},
		{Provides: "kernel.Artifacts", Scope: ScopeProcess},

		// --- module.Context accessors (module/module.go) ---
		//
		// All of these are process-scoped services HANDED to Module.Register,
		// which itself runs once at process boot (Phase 5/app.Boot) — the
		// descriptor's own Scope reflects the lifetime of the underlying VALUE
		// (process), not the lifetime of the Register call, matching how
		// app/context.go's moduleContext actually stores plain process-scoped
		// pointers/interfaces copied from moduleDeps (app/context.go
		// newModuleContext). APIRuntime: true on every one — module.Context IS the
		// api/worker runtime module surface.

		{Provides: "module.Context.Tx", Requires: []string{"kernel.Tx"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Authz", Requires: []string{"kernel.Authz"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Permissions", Requires: []string{"kernel.Perms"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Resources", Requires: []string{"kernel.Resources"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Rules", Requires: []string{"kernel.Rules"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.RulesResolver", Requires: []string{"kernel.RulesResolver"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Workflows", Requires: []string{"kernel.Workflows"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.WorkflowRuntime", Requires: []string{"kernel.WorkflowRuntime"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.RetentionClasses", Requires: []string{"kernel.RetentionClasses"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.DocumentClasses", Requires: []string{"kernel.DocumentClasses"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.DocumentHooks", Requires: []string{"kernel.DocumentHooks"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Documents", Requires: []string{"kernel.Documents"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Comments", Requires: []string{"kernel.Comments"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Attachments", Requires: []string{"kernel.Attachments"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.NotifyTemplates", Requires: []string{"kernel.NotifyTemplates"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Notify", Requires: []string{"kernel.Notify"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Webhooks", Requires: []string{"kernel.Webhooks"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.IntegrationProviders", Requires: []string{"kernel.IntegrationProviders"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Integrations", Requires: []string{"kernel.Integrations"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Audit", Requires: []string{"kernel.Audit"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Sequence", Requires: []string{"kernel.Sequence"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Bulk", Requires: []string{"kernel.Bulk"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Artifacts", Requires: []string{"kernel.Artifacts"}, Scope: ScopeProcess, APIRuntime: true},

		// Privileged() is built lazily from platformTx + ruleStore, scoped and
		// ownership-checked per module name (app/context.go moduleContext.priv);
		// the underlying platform manager is nil in a process wired without a
		// platform pool (migrate), but migrate registers no module that calls it.
		{Provides: "module.Context.Privileged", Requires: []string{"kernel.PlatformTx", "kernel.RuleStore"}, Scope: ScopeProcess, APIRuntime: true},

		// --- app.Boot-local, per-boot-run collectors (app/boot.go newBootState) ---
		// These are built once per Boot() call (effectively process-scoped: Boot
		// runs once at process startup) and handed to every moduleContext.

		{Provides: "app.boot.Router", Scope: ScopeProcess},
		{Provides: "app.boot.Events", Scope: ScopeProcess},
		{Provides: "app.boot.Jobs", Scope: ScopeProcess},
		{Provides: "module.Context.Routes", Requires: []string{"app.boot.Router"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Events", Requires: []string{"app.boot.Events"}, Scope: ScopeProcess, APIRuntime: true},
		{Provides: "module.Context.Jobs", Requires: []string{"app.boot.Jobs"}, Scope: ScopeProcess, APIRuntime: true},

		// --- tenant_tx-scoped: the per-callback handle a module actually reads/
		// writes through, obtained by CALLING kernel.Tx/PlatformTx.WithTenant(RO).
		// Never a field on Kernel/moduleContext (there is no such struct field —
		// database.TenantDB only exists inside the callback), so it depends on
		// the TxManager that produces it but nothing may depend ON it outside
		// that callback (checkTenantEscape enforces exactly this).
		{Provides: "database.TenantDB", Requires: []string{"kernel.Tx"}, Scope: ScopeTenantTx},
	}}
}
