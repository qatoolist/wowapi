# Decision Log

Format per entry: context → options → decision → tradeoffs → affected files/tests.
Blueprint deviations MUST land here before the code that implements them.

## D-0001 — Preflight: `kernel/secrets` added to the package map
- **Context:** 12-configuration-and-deployment referenced `kernel/secrets` types but the 04 package
  map never defined the package (Goal 2 preflight item).
- **Options:** (a) fold secrets into `kernel/config`; (b) separate `kernel/secrets` base package.
- **Decision:** (b). `kernel/secrets` = `Provider` port + `Ref` parsing, stdlib-only, graph base;
  `kernel/config` imports it. Adapters implement providers; `app` resolves refs at boot.
- **Tradeoffs:** one more public package; in exchange adapters don't import `kernel/config` and the
  graph stays layered.
- **Affected:** docs/blueprint/04 §2 (new row), 11 §2 kernel list, kernel/secrets (Phase 1 code).

## D-0002 — Preflight: config type naming standardized
- **Context:** blueprint mixed `config.Config`, `config.MustLoad()`, and `config.Framework`.
- **Decision:** framework-owned struct is **`config.Framework`** (in `wowapi/kernel/config`);
  the product-owned type is **`Config`** in the product's **`internal/appcfg`** package
  (scaffolded by `wowapi init`), embedding `config.Framework`, loaded via `appcfg.Load/MustLoad`.
  `kernel.Kernel.Cfg` is `config.Framework`.
- **Affected:** docs/blueprint/06 §3, 11 §3, 12 §2.

## D-0003 — Preflight: CLI config tooling never imports product packages
- **Context:** installed CLI is prebuilt; it cannot link product config types, but
  `config validate/schema/...` must operate on the *product's* composed config.
- **Options:** (a) CLI parses YAML against a generated JSON schema only; (b) generated
  product-local checker binary the CLI shells out to; (c) plugin loading (rejected: runtime magic).
- **Decision:** (b) with (a) as its transport: `wowapi init` scaffolds `tools/configcheck/main.go`
  (imports `internal/appcfg` + `wowapi/kernel/config`; emits schema/validation/redacted-effective
  JSON on stdout); the CLI runs `go run ./tools/configcheck` in the product repo and formats the
  result. Framework-repo fallback: `config.Framework` alone.
- **Tradeoffs:** requires Go toolchain for config commands in product repos (already required);
  in exchange, full typed validation with zero import-direction violations.
- **Affected:** docs/blueprint/12 §8.

## D-0004 — Preflight: CLI command listings use one command per line
- **Context:** 11 §5 used `wowapi config init | validate | …` which reads as a shell pipe.
- **Decision:** every doc lists each command on its own line. **Affected:** docs/blueprint/11 §5.

## D-0005 — Preflight: acyclicity re-verified with `kernel/secrets`
- **Decision:** graph remains acyclic: `kernel/secrets` (stdlib only) ← `kernel/config` ←
  other `kernel/*` (receive sub-structs by value; no config imports needed) ← `module` ← `app`;
  `adapters` → `kernel/*` only. Encoded in `scripts/lint_boundaries.sh` from Phase 0.
- **Amendment (Phase 1, ARCH-13):** "receive sub-structs by value" still requires importing
  `kernel/config` for the *types* (e.g. `kernel/logging` imports `config.Log`/`config.Fingerprint`).
  That is a types-only kernel→kernel edge, cycle-free and consistent with 04 §2; what stays
  forbidden is other packages *loading* config or reading stores at runtime.

## D-0006 — Phase 0: walking-skeleton scope for `module.Context`
- **Context:** the full Context interface (06 §2) references many kernel packages that don't exist
  yet; stubbing them all would create broad partial implementations (banned by preflight rule 3).
- **Decision:** Phase 0 ships `module.Module` exactly as specified plus a **minimal** `Context`
  (Logger, Config→`config.ModuleView`) with the blueprint-documented growth path; each later phase
  adds its own accessor alongside the capability it delivers. Interface widening pre-v0.1.0 is an
  accepted breaking change (semver v0 rules).
- **Affected:** module/module.go; noted in evidence/phase-00/proof-bundle.md.

## D-0007 — Phase 0: Go toolchain version
- **Context:** blueprint says Go ≥ 1.23; local toolchain is 1.26.4.
- **Decision:** `go.mod` declares `go 1.26` (repo floor). CI pins the same; revisit at v1.
- **Affected:** go.mod.

## D-0008 — Phase 0: `wowapi version` implementation
- **Decision:** CLI version from `runtime/debug.ReadBuildInfo` (main module version when installed
  via `go install …@vX.Y.Z`; `(devel)` in-repo), with `-ldflags -X` override hook for goreleaser.
  Dependency-mismatch warning parses the nearest `go.mod` for the wowapi requirement.
- **Affected:** cmd/wowapi, internal/buildinfo.

## D-0009 — Phase 0: vocabulary denylist pragmatics in boundary lint
- **Context:** blueprint 00 §5 lists denylist words including over-generic ones (building, wing,
  flat, member) that would false-positive constantly in code ("building the request", struct
  members).
- **Decision:** the grep-based Phase 0 lint enforces the unambiguous terms (society, housing,
  chairman, treasurer, defaulter, conveyance, redevelopment, agm, maintenance_bill); generic terms
  are covered by code review until the Phase 5 AST-based lint can check identifiers only.
- **Affected:** scripts/lint_boundaries.sh; revisit at Phase 5.

## D-0035 — Phase 4: migration numbering maps blueprint 002–004 to on-disk 00004–00006
- **Context:** blueprint 03 §5 numbers the identity/resource/authz migrations 002/003/004, but
  on-disk 00003 was taken by idempotency (D-0031). goose numbers are per-source and only need to be
  monotonic.
- **Decision:** 00004_org_party_capacity.sql (blueprint 002: organizations, parties, persons,
  legal_entities, party_contacts, acting_capacities), 00005_resource_relationship.sql (003:
  resource_types, resources, relationship_types, relationships), 00006_authz.sql (004: permissions,
  roles, role_permissions, actor_assignments, policies, policy_conditions). All tenant-scoped tables
  get ENABLE+FORCE RLS + app_rt grants; global registries (resource_types, relationship_types,
  permissions) get app_platform grants (kernel-service access, per SEC-13/D-0026).
- **Affected:** migrations/00004–00006; docs/blueprint/03 §5 note.

## D-0036 — Phase 4: authz evaluator is deny-by-default with a Store port; registry validated at boot
- **Context:** 01 §3 specifies the layered Evaluate algorithm (RBAC → ReBAC → ABAC, deny-first) and
  a permission registry where an unknown permission is a boot error, not a runtime 403.
- **Decision:** `kernel/authz` defines Actor/Target/Decision/Evaluator + a `Store` port (loads
  active assignments, role permissions, relationship grants, policies for an actor/target) so the
  evaluator is pure and unit-testable with a fake store; the pg-backed store lands beside it. The
  permission registry is a validated set built from module route permissions + seeded permissions;
  Evaluate on an unregistered permission returns an error (surfaced at boot when routes register,
  not per request). Filter returns a structured `ListFilter` (org/resource id constraints) the
  store translates to SQL — never load-then-filter.
- **Affected:** kernel/authz, kernel/policy, kernel/relationship, kernel/resource.

## D-0037 — Phase 4: OIDC verifier with an injectable JWKS source + local test issuer
- **Context:** 01 §3 / auth middleware needs an OIDC token verifier, but tests must mint tokens the
  verifier accepts without an external IdP.
- **Decision:** `kernel/auth` verifies JWTs against a `KeySource` port (JWKS by key id); production
  wires a caching JWKS-over-HTTPS adapter, tests wire a local RSA signer (`testkit.IssueToken`).
  The verifier maps validated claims → `authz.Actor` (user id, tenant, capacity) after resolving
  the user's active capacity in the tenant. Break-glass/impersonation carry explicit ctx markers
  and are audited.
- **Affected:** kernel/auth, testkit/auth.go, adapters/oidc (JWKS adapter, later).

## D-0038 — Phase 4 review: closed verb set extended with `ingest` and `activate` (ARCH-41)
- **Context:** the 01 §3 closed action verb set is `create|read|list|update|deactivate|restore|approve|
  reject|assign|export|admin`, but the blueprint's own matrix uses `payments.callback.ingest`
  (webhook ingest) and break-glass needs an `activate` verb.
- **Decision:** extend the closed set with `ingest` (inbound webhook/event ingestion) and `activate`
  (break-glass / feature activation). The set stays closed and small; both have concrete blueprint
  usages. 01 §3's list is updated to match so code and blueprint agree.
- **Affected:** kernel/authz/registry.go, docs/blueprint/01 §3.

## D-0039 — Phase 4 review: evaluator runs in the caller's tenant tx (ARCH-36); caching + list-ReBAC deferred
- **Context:** the pg Store/Checker each opened their own `WithTenantRO` tx per method, so one
  Evaluate spanned ~5 separate transactions — a different MVCC snapshot from the request's business
  tx (a just-written resources mirror row would be invisible), N round-trips, and second-connection
  deadlock risk on the hot path.
- **Decision:** the Store/Checker/Evaluator methods take the caller's `database.TenantDB` and run
  their reads on it — one snapshot, one connection, consistent with the request's writes. The pure
  evaluator is unchanged; only the seam moves. Per-request memoization and the 30s assignment
  snapshot cache (01 §3) are deferred to Phase 5/6 with the live wiring (a TODO on the evaluator);
  ReBAC list visibility (`ListFilter` from relationship-derived resource ids) needs a
  `Store.RelationshipResourceIDs` seam and is completed in Phase 5 when list endpoints ship
  (ARCH-37) — until then `Filter` covers RBAC scopes only, documented in code.
- **Affected:** kernel/authz (store.go, evaluator.go, store_pg.go), kernel/relationship,
  kernel/resource/registrar; phase-plan rows 4/5.

## D-0040 — Phase 5: Context accessor scope (which of 06 §2 ships now)
- **Context:** 06 §2's full Context references kernel packages that arrive in later phases (rules→7,
  workflow→7, outbox/jobs→6, document→8, notify/webhook→9). D-0006 grows Context per phase.
- **Decision:** Phase 5 ships the accessors whose kernel capabilities exist:
  Routes/Permissions/Roles/ResourceTypes/RelationshipTypes, Migrations/Seeds/OpenAPI, Tx/Authz/
  Logger/Config/IDGen/Clock/Health, and Port/ProvidePort (inter-module ports checked at boot). The
  later-phase accessors (Rules/Workflows/Events/Jobs/Documents/Notify/Webhooks) are added with their
  packages. Interface widening pre-v0.1.0 is an accepted breaking change (D-0006).
- **Affected:** module/module.go, app/context.go.

## D-0041 — Phase 5: Kernel + App composition root; boot wires the evaluator and gates on registries
- **Context:** Phase 4 left the evaluator, permission registry, and PrincipalStore dangling (ARCH-39,
  ARCH-44). Phase 5 is where the app boot builds them.
- **Decision:** `kernel.Kernel` (New(ctx, cfg, deps) → owns pool, Tx, Authz evaluator, Log, Health,
  Audit sink) and `app.App` (Register/Validate/StartAPI/StartWorker/Shutdown). Boot order per 06 §2
  lifecycle: construct kernel → per-module Register (collect into registries) → Validate (whole-graph:
  dup permissions, routes without meta, unknown deps/cycles, unsatisfied ports, module-config decode,
  seed-schema, **permission registry Err()**) → SeedSync (idempotent catalog upsert) → Start. The
  evaluator is built from the composed permission registry + PgStore + policy engine + relationship
  checker + audit sink and injected into every module.Context.Authz(). Boot aborts on any Validate
  error — the permission registry gate is now enforced (closes the Phase 4 deferral).
- **Affected:** kernel/ (new package), app/app.go + run.go + context.go.

## D-0042 — Phase 5: seed loader is declarative YAML → idempotent catalog upsert
- **Context:** modules ship `seeds/*.yaml` declaring permissions, roles (+role_permissions),
  resource_types, relationship_types; SeedSync upserts them idempotently (never touches tenant data).
- **Decision:** `kernel/seeds` parses a typed seed bundle (strict YAML, unknown keys fail) and
  SeedSync upserts into the global catalogs as app_platform (the catalogs are app_platform-writable,
  per SEC-13/D-0026). Seed permission/role keys feed the boot permission registry. Idempotent:
  ON CONFLICT DO UPDATE; running twice is a no-op diff. Contract-tested (run twice).
- **Affected:** kernel/seeds, migrations grants (already app_platform), testkit contract suite.

## D-0043 — Phase 5: scratch-consumer test builds a real external module in a tmpdir
- **Context:** the headline exit criterion — an external product repo can import wowapi, define a
  module, and pass the contract suite without framework edits.
- **Decision:** a `test-consumer` flow (host+container) scaffolds a tiny product module in
  t.TempDir(), `go mod init` + `go mod edit -replace github.com/qatoolist/wowapi => <repo>`, writes
  a module using only public packages, and runs `testkit.RunModuleContract`. Proves the public API
  surface is sufficient and import-direction-clean from outside the repo.
- **Affected:** testkit/contract.go, a consumer test under testkit or internal, Makefile test-consumer.

## D-0044 — Phase 5 review: seed ownership covers role grants + granted_via; grants reconciled
- **Context:** the seed prefix-ownership check validated declared keys but NOT the role grant-list or
  `granted_via` — so a module could grant itself a foreign permission (SEC-32, reproduced) or wire
  its permission to another module's relationship (SEC-34). Sync was also insert-only, so removed
  grants never pruned (ARCH-47).
- **Decision:** `seeds.validate` prefix-checks every `RoleSeed.Permissions` entry and `GrantedVia`,
  and requires `granted_via` to name a relationship type the same bundle declares. `Sync`
  reconciles each role's grants (deletes grants not in the seed) so a demoted role sheds
  privileges across redeploys. Regression tests in seeds_test.go.
- **Affected:** kernel/seeds/seeds.go.

## D-0045 — Phase 5 review: seeds run as app_platform; hybrid-table RLS uses a forgiving tenant fn
- **Context:** the contract ran `seeds.Sync` as superuser, never testing the SEC-13 grant boundary
  (SEC-33). Running as app_platform hit the roles/policies RLS `WITH CHECK`, which calls the strict
  `app_tenant_id()` (raises when unset) — a platform connection has no tenant, so NULL-template
  writes aborted.
- **Decision:** add `app_tenant_id_or_null()` (missing_ok → NULL) and use it ONLY in the
  roles/policies policies (`tenant_id IS NULL OR tenant_id = app_tenant_id_or_null()`), so a
  platform/catalog connection can read/write NULL-tenant templates while a tenant connection still
  sees only its rows + templates. Pure tenant tables keep the strict raising `app_tenant_id()`
  (loud fail-closed + AssertRLSIsolation unchanged). testkit provisions an `app_platform` login +
  Platform pool; the contract syncs seeds under it (SEC-33) and asserts effect-idempotency via a
  catalog checksum (ARCH-49). app_rt is still SELECT-only on roles/policies, so this does not widen
  it.
- **Affected:** migrations/00001, 00006; testkit/db.go (Platform pool), testkit/contract.go.

## D-0046 — Phase 5 review: contract RLS check is diff-based, not name-prefix (ARCH-48)
- **Context:** the RLS assertion matched tables by `<module>_` prefix — evadable by naming — and a
  module with zero conforming tables passed silently.
- **Decision:** the contract snapshots public tables before/after the module migrate, and asserts
  ENABLE+FORCE RLS on every table the migration actually created (excluding goose bookkeeping);
  a module that ran migrations but produced no RLS-forced table fails.
- **Affected:** testkit/contract.go.

## D-0047 — Phase 6: Postgres-backed job runner behind the interfaces, not River
- **Context:** Goal 2 says "River OR the selected Postgres-backed job runner behind framework
  interfaces". River is a large dependency with its own migration set and API surface; the module
  portability contract only depends on `jobs.Registry`/`Runner`/`Worker`.
- **Decision:** implement a focused Postgres job queue (`kernel/jobs`) behind those interfaces:
  a `jobs_queue` table, `FOR UPDATE SKIP LOCKED` claim, bounded fixed worker pool per queue,
  exponential backoff + jitter retry, DLQ (status=discarded mirrored to `job_runs`). Interfaces
  match the blueprint so a future River swap is internal. Keeps the dependency surface small and
  the retry/DLQ semantics ours to test precisely.
- **Affected:** kernel/jobs, migration 00007.

## D-0048 — Phase 6: outbox relay reads cross-tenant as app_platform; dispatches per-tenant
- **Context:** `events_outbox` is tenant-scoped (RLS) so modules write/read only their tenant's
  events in the business tx. The relay must dispatch ALL tenants' pending events.
- **Decision:** a role-scoped RLS policy grants `app_platform` (the relay/kernel role) SELECT+UPDATE
  across all outbox rows; the relay claims a batch with `FOR UPDATE SKIP LOCKED` as app_platform,
  then for each event RE-ENTERS a tenant transaction bound to the event's tenant_id (SET LOCAL) to
  run handlers under normal tenant RLS + the inbox dedup. Ordering is per-aggregate
  (`occurred_at` per resource). This keeps app_rt strictly tenant-isolated while giving the kernel
  relay the cross-tenant read it needs — mirrors the app_platform posture from Phase 5.
- **Affected:** migration 00007 (events_outbox policies), kernel/outbox relay.

## D-0049 — Phase 6: TenantDB.Outbox()/Events() + module.Context Events()/Jobs()
- **Context:** 05 §2 TenantDB carries `Outbox()`; 06 §2 Context carries `Events()`/`Jobs()`.
- **Decision:** `database.TenantDB` grows `Outbox() outbox.Writer` (same-tx event write); the
  per-tx writer is attached by the TxManager. module.Context grows `Events() outbox.HandlerRegistry`
  (Subscribe) and `Jobs() jobs.Registry` (RegisterKind). The worker process (`app.RunWorker`) starts
  the relay + job pools and drains gracefully on shutdown.
- **Affected:** kernel/database (TenantDB), kernel/outbox, kernel/jobs, module/module.go,
  app/context.go, app worker start.

## D-0050 — Phase 6 review: per-aggregate ordering enforced; event DLQ; job timeout/drain separation
- **Context:** the review reproduced that per-aggregate ordering was NOT actually held (the
  blueprint's advisory lock was absent; a transient handler failure reordered events, ARCH-53),
  failed events retried forever with an ineffective cooldown (ARCH-54/55), and the job runner
  conflated the shutdown drain with the per-job timeout (ARCH-56/57).
- **Decision:**
  - Relay: the claim only picks the earliest still-undispatched event per (tenant, resource) — a
    later event never overtakes an earlier pending/failed one — plus a tx-scoped
    `pg_advisory_xact_lock` per aggregate so concurrent relays serialize. Per-aggregate ordering is
    now real (regression test under retry).
  - Event DLQ: `events_outbox` gains `failed_at`, `max_attempts`, `last_error` and a `'dead'` status;
    a poison event dead-letters after max_attempts; `RequeueFailed` keys its cooldown on `failed_at`.
  - Jobs: a per-job `jobTimeout` (default 2m) separate from the shutdown `drainTimeout`; outcomes are
    written with a fresh short-lived context; `stalledTimeout` is floored above jobTimeout+drain so a
    live job can't be reclaimed and run concurrently (ARCH-58); `StartWorker` enforces a HARD drain
    cap so a ctx-ignoring worker can't hang shutdown (ARCH-57).
  - Semantics documented: jobs are at-least-once with NO framework dedup (workers with external side
    effects must carry their own idempotency key); event handlers get exactly-once DB effect via the
    inbox (ARCH-59).
- **Affected:** migrations/00007, kernel/outbox/relay.go, kernel/jobs/{runner,jobs}.go, app/worker.go.

## D-0051 — Phase 7: migrations 00008 (rules) + 00009 (workflow); custom Postgres engines
- **Context:** blueprint 02 §1.1 recommends a small custom Postgres-backed workflow engine over
  Temporal/Camunda (approval/state-machine shaped, tenant-editable, shares the business tx/RLS/audit/
  outbox). Rules likewise are a Postgres-backed versioned config engine.
- **Decision:** `kernel/rules` (rule-point registry + version storage + resolution) and
  `kernel/workflow` (definition model + runtime + SLA sweeper) as custom engines. Migration 00008
  = rule_definitions (global) + rule_versions (tenant+platform hybrid, temporal, exclusion
  constraint one-active-per-scope); 00009 = workflow_definitions (global+tenant) + workflow_instances
  + workflow_tasks + workflow_task_assignees (tenant-scoped RLS). Both engines share the tenant tx
  and emit outbox events + audit in the same transaction as state changes.
- **Affected:** kernel/rules, kernel/workflow, migrations/00008–00009.

## D-0052 — Phase 7: rule resolution is org-ancestry → tenant → platform → code default, historical by `at`
- **Decision:** `rules.Resolver.Resolve(key, tenant, org?, at)` picks the first active version
  (effective_from <= at < effective_to) walking org ancestry upward, then tenant, then platform,
  then the code-registered default; the value is JSON-Schema validated (defense in depth) and
  returned with provenance. Versions are immutable (never mutated, only superseded), so any
  historical `at` resolves deterministically. Approval-gated points require an `active` version to
  have passed approval; a draft/pending version never resolves. Resolution runs on the caller's
  TenantDB (one snapshot).
- **Affected:** kernel/rules resolver + tests.

## D-0053 — Phase 7: workflow step-type set is closed; definitions validated at boot
- **Decision:** closed step types (approval|task|auto|gateway|vote|terminal); assignee kinds
  (actor|role-at-scope|relationship|resource_owner|resolver). Definitions are validated at
  registration (graph connectivity, no orphan steps, terminals reachable, unknown auto-actions
  fail boot). Instances pin their definition version (immutable per version). Every transition
  re-checks the actor (assignee + `workflow.task.decide`), mutates with optimistic locking, and
  writes audit + outbox in the same tenant tx. testkit `WorkflowSim` drives definitions over a real
  test DB.
- **Affected:** kernel/workflow, testkit/workflowsim.

## D-0060 — Review-findings pass: runtime authz gate, deploy/config-scaffold fixes, CI DB gate
- **Context:** an external review reproduced six findings against the Goal-2 framework; five were real
  (one a false-premise-free but expected deferral). Fixed each with existing conventions + regression tests.
- **Decisions:**
  - **Runtime authz enforcement (High).** The RouteMeta permission gate was boot-validated but NEVER
    enforced per request — a deployed API served every route unauthenticated/unauthorized. Added
    `httpx.SecureHandler`/`gateRoute`: for each non-Public route, AuthN (via a pluggable `Authenticator`
    port — the product supplies OIDC/tenant strategy) → bind tenant+actor → AuthZ(permission) at tenant
    scope → serve; deny-by-default. The generated api wires it with `DenyAllAuthenticator` (fail-closed:
    business routes 401 until a real Authenticator is set). Fine-grained resource checks stay per-handler.
  - **Workflow pagination off-by-one (Medium).** `OpenTasksFor` encoded the cursor from the dropped
    lookahead row, skipping one task per page boundary; now encodes the last RETURNED item. Regression
    test proven by revert (skips 1 → paged 4/5).
  - **deploy render (High).** Defaulted `--env production` (invalid; valid is `prod`) and rendered
    `${WOWAPI_DB_DSN}` (config.DB.DSN is a Secret needing `secretref://`). Now defaults `prod`, validates
    `--env` via `config.Env.Valid()`, and renders `secretref://env/WOWAPI_DB_DSN` (+ MIGRATE_DSN).
  - **Product config scaffolding (Medium).** `wowapi init` now scaffolds `internal/appcfg` (product
    Config embedding config.Framework + Modules namespaces, D-0002) and `tools/configcheck` (D-0003); the
    generated api/worker load via `appcfg.Load` and pass `cfg.Modules` to `Boot` (was `nil`).
  - **CI DB-skip hygiene (Medium).** DB-backed tests SKIP without a DSN, so host `make ci` could be
    green-but-hollow. Added `testkit.RequireDB()` (WOWAPI_REQUIRE_DB=1) → FAIL not skip; `make ci-container`
    and `make test-integration` set it, so the authoritative gate cannot silently skip DB/E2E proofs.
  - **Deferrals (Lower) — no change.** Workflow vote/min_approvals>1/self_approval are fail-closed
    (D-0054), audit_logs is the logging sink, gen-crud emits honest TODO handlers — all already
    accurately documented as deferrals; verified no doc overclaims them complete.
- **Affected:** kernel/httpx/{authz_gate,router}.go, kernel/workflow/runtime.go, internal/cli/{deploy_cmd,
  init_cmd}.go + templates, testkit/db.go + consumer_test, internal/e2e, internal/testmodules/requests,
  Makefile; evidence/phase-12 acceptance-map (#18 now runtime-enforced).

## D-0059 — Phase 12: `wowapi init` produces a framework-wired product repo; E2E acceptance
- **Context:** Phase 12 (capstone) must prove a blank repo builds a WORKING API binary (AC #19) and runs
  kernel + module migrations from cmd/migrate (AC #22). The Phase-10 init mains were framework-import-free
  stubs — a gap.
- **Decisions:**
  - **The scaffolded mains wire the framework.** `wowapi init` now renders real `cmd/api|worker|migrate`
    mains: config load → pool (runtime AS app_rt + RLS guard; worker also a platform pool) → `kernel.New`
    → `app.New().Register(wire.Modules()...).Boot` → serve the router behind the observability middleware
    chain + `/healthz`//`/readyz`, graceful shutdown; worker runs `app.StartWorker`; migrate runs
    `migrations.Kernel()` then each module's migrations. Modules are registered via a generated
    `internal/wire/modules.go` (manual list — auto-append is a documented follow-up).
  - **Config scaffold uses secret references.** `configs/local.yaml` renders `secretref://env/DATABASE_URL`
    (raw/empty DSN strings fail `Secret.UnmarshalText` by design) — the secret-ref-only guarantee shows up
    in the scaffold itself.
  - **E2E test = acceptance through the real CLI.** `internal/e2e` runs `wowapi init`, replaces wowapi with
    the local tree, `go build`s the repo, and (with a DB) runs the migrate binary + curls the api binary's
    `/healthz` — following the consumer test's offline-skip discipline.
  - **Release notes + full acceptance sweep.** `CHANGELOG.md` (v0.1.0); the 28-criterion acceptance map.
- **Affected:** internal/cli/templates/init/* (cmd mains + internal/wire + config), internal/cli/init_cmd.go,
  internal/e2e/e2e_test.go, CHANGELOG.md; evidence/phase-12/. **Goal 2 complete (Phases 0–12).**

## D-0058 — Phase 11: observability + performance budgets + security suite + config drift
- **Context:** Phase 11 hardens the framework (blueprint 07 §1–2/§9; AC #17/#18/#26/#27) — observability
  wiring, perf budgets, a security gate, and cross-process config drift. Additive; no new domain tables.
- **Decisions:**
  - **Observability = ports + adapters:** `kernel/observability` defines a small `Metrics` port
    (ObserveRequest/IncCounter/SetGauge) + a NoOp default + RED and AccessLog middleware; the Prometheus
    client lives ONLY in `adapters/metrics/prometheus` (with a `/metrics` handler). The RED middleware
    labels by the matched route PATTERN (bounded cardinality). Full OTel span export is a product adapter.
  - **Health:** `kernel/httpx/health.go` — liveness runs NO checks (a failing dep must not trip a
    liveness probe); readiness runs checks → 200/503 and reports the redacted config fingerprint.
    `app.Readiness` assembles module `ctx.Health` + framework checks (DB ping / migrations-current,
    supplied by the composition root) + fingerprint.
  - **Performance budgets (#17):** 24 hot-path benchmarks + a pure-Go `internal/tools/benchbudget` gate
    reading piped `go test -bench` output against `bench-budgets.txt`, wired into `make ci`. Config field
    reads at 0.3 ns/op, 0 allocs prove the hot path is reflection/lookup-free.
  - **Security suite (#18/#26):** a curated `make test-security` gate over the existing RLS/authz/
    privilege/secret tests + new per-knob unsafe-config matrix + a structural-secret-redaction gap test.
    Audit found the core guarantees (deny-by-default, secret-ref-only, structural redaction, RLS,
    unsafe-config-fails-startup) have no disabling config key.
  - **Config drift (#27):** `kernel/config/shared.go` — `SharedFingerprint` covers env/schema/DB
    (excludes process-specific HTTP/Log); `CheckSharedDrift(expected)` fails a mis-deployed process.
- **Affected:** kernel/observability, adapters/metrics/prometheus, kernel/httpx/health.go,
  kernel/config/shared.go, app/health.go, internal/tools/benchbudget, bench-budgets.txt, Makefile
  (bench/bench-budget/test-security + bench-budget in ci), benchmarks + security tests; evidence/phase-11/.

## D-0057 — Phase 10: installable `wowapi` CLI (scaffolding, codegen, tooling) + review fixes
- **Context:** Phase 10 delivers the CLI command surface (blueprint 10 §2 E21): init, new-module,
  gen crud, migrate create, seed validate, openapi merge, lint boundaries, deploy render — plus the
  existing version/config. No new DB tables.
- **Decisions:**
  - **Dispatcher = one file per command:** `internal/cli/cli.go` switches to a `runX(args, stdout,
    stderr) int`; each command is its own file, buffer-testable. Enabled a conflict-free parallel build
    (lead: transform commands; agent: scaffolding).
  - **Generated Go is gofmt-clean:** `renderToFile` runs `go/format.Source` on `.go` output — formats
    AND fails generation loudly on an invalid-Go template (stronger than a parse-only check).
  - **Scaffold path safety:** module/resource/field names are `identRE`-validated before any path is
    built (no traversal); `--force` gates every overwrite.
  - **lint reuses the framework law:** `wowapi lint boundaries` ports the import-layering + module-
    isolation rules from `scripts/lint_boundaries.sh` as a pure, unit-tested `checkBoundaries`; the
    shell script remains the authoritative framework gate for vocabulary/Reveal/test-import checks.
  - **Review fixes (D-0057):** unknown `gen crud` field type rejected instead of emitting unbuildable Go
    (CLI-01); `openapi merge` rejects non-object fragments (CLI-02); `checkBoundaries` gained the missing
    adapters/cmd/internal-cli/internal-tools layer rules + hard testkit rule (CLI-06); usage-error exit
    codes normalized to 2 (CLI-03); `go list` stderr surfaced (CLI-04); stdout write errors propagated
    (CLI-05); derived package name validated (CLI-07).
- **Affected:** internal/cli/ (all command files, scaffold.go, templates/, tests), cmd/wowapi;
  evidence/phase-10/.

## D-0056 — Phase 9: notify / webhook / integration framework + review fixes
- **Context:** Phase 9 delivers the notification, webhook, and integration subsystems (migration 00011,
  blueprint 07 §5/§6). Two parallel review agents reproduced 13 defects (evidence/phase-09/review-findings.md).
- **Decisions:**
  - **Config tables are app_platform-written (SEC-13):** notification_templates, integration_providers,
    and webhook_endpoints are behavior-changing config (which channels/endpoints fire, which credentials
    sign) — app_rt SELECT-only. notifications is module-written in a business tx; notification_deliveries
    and webhook_events are append-only to app_rt with status advanced by the app_platform sender/relay.
  - **Notifications:** template registry (module-declared, allowlisted vars, `text/template` — but
    `html/template` for the email channel to auto-escape, SEC-51); `Send` writes the notification + one
    delivery per resolved channel in the caller's tenant tx and dry-run-renders each body so a missing
    var fails synchronously (ARCH-77); `SendPending` (app_platform) claims + delivers with a
    `next_attempt_at` backoff and a maxAttempts dead-letter (ARCH-75).
  - **Webhooks:** inbound `HandleInbound` verifies the provider signature (constant-time HMAC), enforces
    replay via a synthesized-or-provided dedup id over a PARTIAL unique index (SEC-49) and a ±5m window;
    a signature-failure audit row carries a NULL dedup id so it cannot block a real event (SEC-50);
    outbound signing covers `timestamp + "." + body` (SEC-52). `RetryOutbound` (app_platform) is the
    worker that actually drives outbound backoff/DLQ — DispatchOutbound alone gave one attempt (ARCH-70).
    A per-endpoint circuit breaker opens after N failures, half-opens after a cooldown, and clears the
    persisted `degraded` status on recovery (ARCH-72).
  - **Integrations:** a provider-adapter registry (anti-corruption boundary) + a store that resolves
    per-tenant/platform config and a credential from a secret REFERENCE (plaintext rejected); `Upsert`
    uses `RETURNING id` so the conflict path returns the real row id (ARCH-71); `HealthChecks` probes
    configured providers for readiness.
  - **Hybrid RLS backstop (SEC-53):** a RESTRICTIVE policy on the platform+tenant hybrid tables forbids
    a tenant-bound session from writing a NULL-tenant (platform) row.
  - **events_outbox INSERT for app_platform:** granted in 00011 so tenant-bound workers (inbound
    handlers, the delivery sender) can emit events; the relay's WITH CHECK admits it, the outbox Writer
    stamps the tenant.
- **Affected:** kernel/notify, kernel/webhook, kernel/integration,
  migrations/00011_notify_webhook_integration.sql, kernel/kernel.go, module/module.go,
  app/{context,boot}.go; evidence/phase-09/.

## D-0055 — Phase 8: document/file framework (storage port, append-only versions, grant RLS) + review fixes
- **Context:** Phase 8 delivers documents/versions/grants/comments/attachments (migration 00010,
  blueprint 07 §4). Two parallel review agents reproduced 13 defects (evidence/phase-08/review-findings.md).
- **Decisions:**
  - **Object storage is a port (`kernel/storage.Adapter`):** PresignPut/Get + Stat + Peek + Delete;
    blob bytes never transit the API process (client ↔ store via presigned URLs). A memory adapter
    backs tests + local dev; an S3/minio adapter implements the same five methods.
  - **Append-only versions + privilege split:** `document_versions` is INSERT-only to app_rt;
    scan-status settlement and retention voiding run as app_platform (tenant-bound via a PlatformTxM),
    so a module can neither rewrite an immutable file pointer nor clear an infected scan flag.
  - **Download authorization is deny-first + owner + capacity-grant:** an explicit deny policy from
    the authz evaluator is authoritative; otherwise the document owner, an authz role/policy allow, or
    a valid (windowed) capacity grant permits. Two kernel-owned permissions (`kernel.document.read`,
    `kernel.document.update`) are registered at boot.
  - **Grant writes are RLS-ownership-enforced (SEC-41/42):** a new `app_actor_id()` SQL function +
    a RESTRICTIVE policy pin every `document_access_grants` INSERT/UPDATE to a document the acting
    actor owns — a module cannot self-grant or redirect a grant even via raw SQL. Chosen over an
    app_platform-only grant path to keep grant creation composable in the module's business tx.
  - **Governance columns are app_platform-only (SEC-44):** app_rt gets column-level UPDATE on
    documents (title/sensitivity/version/updated_*) but NOT status/legal_hold/retention_until — a
    module cannot clear a legal hold or void a document to dodge retention.
  - **Download is a pure read (ARCH-65):** it emits NO outbox event (that INSERT broke read-only-tx
    callers); durable download audit is deferred to the audit_logs writer.
  - **Retention sweep ordering (SEC-48):** rows are tombstoned inside the tx; blobs are deleted only
    AFTER commit — a failure orphans a blob (safe) rather than leaving an active row over a deleted blob.
  - **Random storage keys (ARCH-66):** the upload key uses a UUID suffix, not the version number, so
    concurrent InitiateUpload calls never clobber each other's blob.
  - **Comment/attachment author guards (SEC-45/46):** Go-level author/creator checks (fail-closed on
    no actor) for edit/void/detach — the realistic user-vs-user protection; a trusted in-process
    module issuing raw SQL can still touch its own tenant's rows (accepted; DB-level protection is
    reserved for the cross-authorization/legal controls).
- **Affected:** kernel/storage, kernel/document, kernel/comment, kernel/attachment,
  migrations/00010_documents.sql, kernel/kernel.go, module/module.go, app/{context,boot}.go,
  testkit/db.go; evidence/phase-08/.

## D-0054 — Phase 7 review: temporal resolution, write-time schema, draft/activate split, workflow fail-closed
- **Context:** two parallel review agents (security + architecture) reproduced eight gaps in the
  rules + workflow slice (see evidence/phase-07/review-findings.md).
- **Decisions:**
  - **Historical resolution includes superseded (ARCH-60):** the resolver reads
    `status IN ('active','superseded')` within the temporal `effective_from/to` window, not
    `status='active'` — a value active in the past then superseded must still resolve for an `at`
    inside its old window rather than falling through to the code default.
  - **Write-time schema validation (SEC-40):** `Propose` validates the value against the point's
    `value_schema` (focused `type`+`enum` validator, `kernel/rules/schema.go`) before INSERT —
    defense in depth over read-path Decode. Full JSON Schema deferred.
  - **Draft/activate privilege split (SEC-13):** `Propose` inserts a DRAFT on app_rt (never
    resolves); `Activate` supersedes+activates on app_platform via a role-scoped
    `rule_versions_platform_all` policy. Activation changes runtime behavior, so it stays off the
    module role. `created_by` is the proposing actor from `ActorIDFrom(ctx)` (ARCH-62).
  - **Workflow fail-closed on unenforced gating (SEC-36/37/38):** the runtime does not yet tally
    votes, enforce `min_approvals > 1`, or exclude self-approval, so the definition validator
    REJECTS such definitions at boot rather than accepting and mis-enforcing them. `Policy.SelfApproval`
    is `*bool` to distinguish unset from explicit false. Approval steps must define both
    `on_approve.next` and `on_reject.next` (ARCH-64). Per R7, fail-closed is the acceptable posture
    for an unshipped control.
  - **Override authz gate (SEC-39):** `Runtime.Override(ctx, actor, id, to, reason)` evaluates
    `workflow.instance.override` on the instance resource before forcing a step; deny → `KindForbidden`.
  - **Test-suite fix:** `TestVerify_TamperedSignature` flipped the trailing base64url char of the
    JWT signature, which can carry only discarded padding bits → non-deterministic (passed on host,
    failed in-container). Now flips the first char (always 6 significant bits); 200× stable.
- **Affected:** kernel/rules/{resolver,store,schema}.go, kernel/workflow/{definition,runtime}.go,
  kernel/auth/auth_test.go; evidence/phase-07/.

## D-0010 — Phase 0→1: `environment` is fail-closed in deployed processes (SEC-1)
- **Context:** security review: `Defaults()` sets `environment=local`; a prod deploy that forgets
  to set it would silently validate under local (lenient) rules.
- **Decision:** the Phase 1 loader errors when `environment` is absent from every layer; the
  compiled `local` default serves only `Defaults()` in tests/local tooling. Blueprint 12 §4
  updated; Phase 1 exit criteria include a test for this.
- **Affected:** docs/blueprint/12 §4; kernel/config loader (Phase 1).

## D-0011 — Phase 1: first third-party dependency, `gopkg.in/yaml.v3`
- **Context:** the layered loader must parse `configs/*.yaml`; blueprint 12 §2 already assumes YAML
  (product `Modules map[string]yaml.Node` example). Repo had zero deps.
- **Options:** (a) hand-rolled YAML subset (rejected: config parsing is exactly where correctness
  bugs hide); (b) `gopkg.in/yaml.v3` (stable, no transitive deps); (c) JSON-only config (rejected:
  blueprint mandates YAML overlays).
- **Decision:** (b). The "kernel/config imports only stdlib + kernel/secrets" rule in 12 §2 governs
  the *internal package graph* (acyclicity), not third-party libs; yaml.v3 keeps the graph acyclic.
- **Affected:** go.mod, kernel/config loader.

## D-0012 — Phase 1: binder scope — `conf`/`default`/`required` tags + `Validate()` hook
- **Context:** blueprint 12 §2 shows a full tag DSL (`conf`, `default`, `validate:"min=…,max=…"`,
  `unsafe`, `redact`, `doc`); Phase 0 shipped hand-written `Framework.Validate()` with accumulated
  errors; risk R5 warns against a reflection-heavy config system.
- **Decision:** ONE audited binder implementing: `conf` key mapping (embedded structs flatten),
  `default:"…"` tags, `required:"true"`, strict unknown-key rejection, scalar conversion
  (string/bool/ints/floats/duration/Env/Secret/slices), `unsafe:"true"` prod refusal (stage warns),
  and `doc` tags (feed `config schema`). Range/cross-field/enum checks stay in code via a
  `Validate() error` hook (already accumulates all errors) — no min/max tag mini-language.
  A drift-guard test asserts tag defaults reproduce `Defaults()`.
- **Tradeoffs:** two places express constraints (tags for shape, code for ranges); in exchange the
  binder stays small enough to audit and R5 stays contained.
- **Affected:** kernel/config (bind/load/schema), config_test.go.

## D-0013 — Phase 1: env secret provider lives at `adapters/secrets/envprovider`
- **Context:** D-0001 put the `Provider` port in `kernel/secrets` with implementations in adapters;
  blueprint 04 §1 lists `adapters/secrets/`.
- **Decision:** first provider is `adapters/secrets/envprovider` (`secretref://env/<VAR>` →
  process environment), with an injectable lookup func for tests. Cloud providers follow the same
  layout later (`adapters/secrets/<name>provider`).
- **Affected:** adapters/secrets/envprovider, app boot wiring, CLI config commands.

## D-0014 — Phase 1: loader API is `Load[T]` (blueprint signature) + `LoadDetailed[T]`
- **Context:** blueprint 12 §2 fixes `Load[T any](opts Options) (T, Fingerprint, error)`, but
  `config doctor` needs per-key provenance and stage-unsafe warnings need a channel out.
- **Decision:** keep the blueprint signature as the primary API; add
  `LoadDetailed[T any](opts Options) (Loaded[T], error)` where `Loaded` carries Config,
  Fingerprint, Provenance (key → layer) and Warnings. `Load` delegates to `LoadDetailed`.
  Fingerprint = SHA-256 of the canonical *redacted* effective config JSON (structural `Secret`
  redaction makes this safe by construction).
- **Affected:** kernel/config/load.go, internal/cli (validate/print/doctor), app views.

## D-0015 — Phase 1: `unsafe` knob mechanism ships now; first framework knob later
- **Context:** 12 §4 requires a per-knob prod-refusal matrix, but every listed dev convenience
  (fake token issuer, SQL echo, public pprof, permissive CORS) belongs to a later-phase component;
  adding a dead config field now would be a partial implementation (banned by preflight rule 3).
- **Decision:** the binder's `unsafe:"true"` handling (prod=error, stage=warning) is implemented
  and matrix-tested in Phase 1 against test-local structs (the binder is generic, so the tests are
  real end-to-end loader tests); `AllowFlags`-style CLI flags refused in prod is the one live
  production rule now. Each later phase adds its real knobs with `unsafe:"true"` + a matrix entry.
- **Affected:** kernel/config loader + tests; later phases' config sections.

## D-0016 — Phase 1 review: `config.Options` final shape (supersedes blueprint 12 §2 sketch)
- **Context:** review finding ARCH-12 — the implemented Options diverged from the blueprint sketch
  (`AllowFlags bool` dropped; `Environ []string` and `Flags map[string]string` added).
- **Decision:** keep the implemented shape. `Flags` presence + the prod refusal rule subsumes
  `AllowFlags` (an empty map IS "flags not allowed"); `Environ` makes the env layer hermetic in
  tests instead of mutating the process environment. Blueprint 12 §2 updated to match.
- **Affected:** kernel/config/load.go, docs/blueprint/12 §2.

## D-0017 — Phase 1 review: the environment gate is not overridable downward (SEC-5)
- **Context:** security review reproduced two downgrades: an env var could flip a committed
  `environment: prod` to `local` (disabling every prod check), and a flag setting `environment`
  escaped the flags-refused-in-prod guard by lowering the value the guard reads.
- **Decision:** trust rules in the loader: (1) `environment` may never come from the flag layer;
  (2) an env var may *supply* `environment` only when no config file sets it — any mismatch with a
  file value is an error, not an override; (3) prod checks and the flag guard key off the
  file-layer value when present. The blueprint §1 table's "env vars set `environment`" reading is
  narrowed accordingly (12 §4 updated).
- **Tradeoffs:** a platform can no longer "promote" an image whose files say `dev` by env var —
  intentional; environment changes ship as config changes.
- **Affected:** kernel/config/load.go; tests TestLoadEnvironmentNotDowngradableByEnvVar,
  TestLoadEnvironmentNeverFromFlags, TestLoadFlagDowngradeStillRefusedInProd; docs/blueprint/12 §4.

## D-0018 — Phase 1 review: module namespaces are file-layer only (for now) (ARCH-8)
- **Context:** env-var/flag values reach the tree as strings; a module's strict typed Decode would
  fail with a confusing per-module JSON error at boot (`"4"` into an int field).
- **Decision:** the loader rejects `modules.*` keys sourced from the env-var or flag layers with a
  clear error at load time. Lifted when module config decoding learns scalar string coercion
  (revisit at Phase 5 with the module SDK).
- **Affected:** kernel/config/bind.go (namespaces case); TestLoadModuleNamespaceViaEnvVarRejected;
  docs/blueprint/12 §3.

## D-0019 — Phase 1 review: unsafe knobs are judged on final bound values (SEC-3/SEC-4)
- **Context:** security review reproduced two fail-open holes: an unsafe knob whose unsafe value
  is its compiled default was never checked (check lived on the "value present in tree" path), and
  unsafe tags on struct/Secret/slice/pointer fields were silently unenforced.
- **Decision:** enforcement moved to a post-bind pass over the fully bound struct: any
  `unsafe:"true"` field with a non-zero final value refuses prod / warns stage, regardless of
  which layer (or default tag) produced the value and regardless of field kind.
- **Affected:** kernel/config/bind.go (enforceUnsafe), load.go; tests
  TestLoadUnsafeDefaultRefusedInProd, TestLoadUnsafeStructKnobRefusedInProd.

## D-0020 — Phase 2: `kernel/model` ships complete now
- **Context:** phase-plan row 2 doesn't name kernel/model, but TenantDB helpers key on
  `model.TenantScoped`, testkit fixtures return typed handles, and migrations follow its column
  conventions — building database/testkit against ad-hoc types would create the partial
  implementations preflight rule 3 bans.
- **Decision:** implement 04 §3 verbatim in Phase 2: BaseFields/TenantScoped/Auditable/CreatedOnly/
  Versioned/Temporal/Statused + Ref value objects + `IDGen` port with a UUIDv7 default.
  Deps: google/uuid (v7 support), shopspring/decimal (Money).
- **Affected:** kernel/model; go.mod.

## D-0021 — Phase 2: DB DSNs validated at process-view narrowing, not by a required tag
- **Context:** blueprint 12 §2 sketches `DSN Secret validate:"required"`, but a tag-required DSN
  would make every Framework load (CLI schema/validate in the framework repo, config-only tests,
  Defaults()) fail without a database — and §7 says each process receives only what it needs.
- **Decision:** `config.DB` fields are optional at load; `app.NewAPIConfig`/`NewWorkerConfig`
  error when the runtime DSN is unset, `app.NewMigrateConfig` errors when the migrate DSN is
  unset. Raw (non-secretref) DSN strings remain structurally impossible (Secret.UnmarshalText).
- **Affected:** kernel/config/config.go (DB section), app/views.go, tests.

## D-0022 — Phase 2: integration tests use env-DSN + template-database clones, not testcontainers
- **Context:** test-strategy sketched testcontainers; the compose stack already provides Postgres
  both on the host (localhost:5432) and inside the tools container (DATABASE_URL), and
  testcontainers-go would be the largest dependency in the tree by far.
- **Decision:** testkit connects via `WOWAPI_TEST_DSN` (fallback `DATABASE_URL`); tests skip with
  a clear message when neither is set. Speed: kernel migrations run once per process into a
  template database; each test gets `CREATE DATABASE … TEMPLATE …` + drop on cleanup.
  Testcontainers can be layered later without API changes. test-strategy.md updated.
- **Affected:** testkit/db.go, Makefile test-integration, docs/implementation/test-strategy.md.

## D-0023 — Phase 2: runtime RLS identity is a non-superuser login (revised after SEC-11/SEC-12)
- **Context:** RLS must be enforced against a role that is non-owner, non-superuser, and lacks
  BYPASSRLS. The original decision (superuser admin login + `SET ROLE app_rt`) was reproduced by
  the Phase 2 security review to be escapable: a module running arbitrary SQL as designed can
  `RESET ROLE` back to the superuser login mid-transaction and read every tenant (SEC-11), and a
  pool wired against an over-privileged DSN silently disables RLS with no signal (SEC-12).
- **Decision:** deployed processes MUST authenticate as a **non-superuser login mapped to app_rt**;
  `SET ROLE` from a superuser is no longer an accepted production posture. Defense in depth, all
  shipped:
  1. `database.WithConnRLSGuard()` refuses, at connect, any pool whose effective role is superuser
     or BYPASSRLS (fail-closed pool construction).
  2. `database.Manager` `WithRole` re-asserts `SET LOCAL ROLE` per tenant tx (survives pool-state
     leaks across checkouts), and `WithRLSGuard` re-checks enforcement per tenant tx.
  3. `app_rt`/`app_platform` stay NOLOGIN in the committed migration — no password ships. The
     testkit grants `app_rt` a local-only LOGIN out-of-band (never committed) and connects as it,
     modelling production exactly; the SEC-11 escalation test passes only because the login is a
     genuine non-superuser.
- **Tradeoffs:** product deployment docs must state the non-superuser-login requirement plainly
  (Phase 10/12); `WithSetRole` is retained only as a session baseline for tooling, not a security
  boundary.
- **Affected:** migrations/00001_bootstrap.sql, kernel/database (pool guards, per-tx role),
  testkit/db.go, docs/blueprint/12 (deployment note, Phase 10).

## D-0026 — Phase 2 review: global identity tables granted to app_platform, not app_rt (SEC-13)
- **Context:** global tables carry no RLS (03 §1); granting them to `app_rt` let any module read or
  tamper with the whole cross-tenant membership graph via ordinary tenant-tx SQL.
- **Decision:** 00002 grants SELECT/INSERT/UPDATE on tenants/users/user_tenant_access to
  `app_platform` only. Kernel identity services run platform transactions under that role via a
  dedicated pool; that pool is wired when the first such service lands (Phase 4). In Phase 2 the
  runtime `app_rt` simply cannot touch the global spine — correct for now.
- **Affected:** migrations/00002_core_identity.sql; kernel/database.Manager.Platform (pool wiring
  deferred to Phase 4, tracked in phase-plan row 4).

## D-0027 — Phase 2 review: per-source migration history tables (ARCH-16)
- **Context:** goose derives a version from the leading filename digits and tracks one history
  table; kernel `00001..` and a module's `0001..` would collide, making the documented
  multi-source model impossible.
- **Decision:** `database.Migrate(ctx, pool, src, source)` uses a per-source history table
  (`goose_version_<source>`); the kernel source is `migrations.SourceName` ("wowapi"), each module
  supplies its own. Independently-numbered sources coexist. `Migrate` returns `MigrateResult{Version,
  Applied}` so idempotency (`Applied==0` on rerun) is assertable.
- **Affected:** kernel/database/migrate.go, migrations/migrations.go, internal/tools/migrate,
  testkit; docs/blueprint/03 §5 wording.

## D-0028 — Phase 2 review: ExpectOneRow distinguishes 0-row conflict from >1-row bug (ARCH-20)
- **Decision:** 0 rows → `ErrVersionConflict` (409/412); >1 row → a distinct internal error (500),
  never masked as a conflict — a too-broad WHERE on a versioned aggregate is a bug, not contention.
- **Affected:** kernel/database/errors.go.

## D-0029 — Phase 2 review: `config.Pool` sub-struct absorbs shared pool knobs (ARCH-17)
- **Decision:** pool knobs live in `config.Pool`, embedded in `config.DB` and in the app views'
  `RuntimeDB`/`MigrateDB`; new pool fields propagate to every narrowed view without editing the
  narrowing code, closing the silent-drop drift.
- **Affected:** kernel/config/config.go, app/views.go.

## D-0030 — Phase 2 review: actor binding stays optional until the actor model exists (ARCH-19)
- **Context:** 05 §2 says `WithTenant` binds `app.tenant_id` AND `app.actor_id` "error if absent".
  The Phase 2 TxManager hard-fails on missing tenant but binds actor only when present. There is no
  actor model, no audit triggers, and no `created_by` defaults reading `app.actor_id` until Phase 4.
- **Decision:** keep actor binding optional for Phase 2 (tenant remains fail-closed). When Phase 4
  introduces the actor/audit machinery that actually consumes `app.actor_id`, `WithTenant` (RW)
  will require it (fail-closed at the door), while `WithTenantRO` read paths stay actor-optional.
  Recorded now so the deviation from 05 §2 is explicit, not silent.
- **Affected:** kernel/database/txmanager.go; revisit at phase-plan row 4.

## D-0031 — Phase 3: idempotency_keys migration (00003) ships now, out of 03 §5 order
- **Context:** phase-plan row 3 requires tested idempotency helpers; 05 §2's `IdemStore` needs the
  `idempotency_keys` table, which blueprint 03 §5 lists in migration 009 (a Phase 6 batch).
- **Decision:** pull the single `idempotency_keys` table forward into kernel migration
  `00003_idempotency.sql` (tenant-scoped, ENABLE+FORCE RLS, granted to app_rt) so the Phase 3
  idempotency store is real and integration-tested against RLS now. The remaining migration-009
  tables (outbox, processed_events, job_runs, audit_logs) still land in Phase 6. Migration numbers
  are per-source and monotonic, so pulling one table forward is safe.
- **Affected:** migrations/00003_idempotency.sql; kernel/database/idempotency.go (IdemStore + pg
  impl); kernel/httpx/idempotency.go (WithIdempotency); docs/blueprint/03 §5 note.

## D-0032 — Phase 3: module.Context gains Routes() and Validator()
- **Context:** D-0006 grows Context per phase; Phase 3 delivers httpx + validation, so modules can
  now register routes and validate input.
- **Decision:** add `Routes() *httpx.Router` and `Validator() *validation.Validator` to
  module.Context (and the app-side moduleContext). Route registration errors surface at boot via
  Router.Err(). Tx()/Authz()/etc. still arrive in their phases.
- **Affected:** module/module.go, app/context.go.

## D-0033 — Phase 3 review: the database layer may emit taxonomy Kinds (ARCH-30)
- **Context:** D-0024 kept `kernel/database` on exported sentinels mapped upstream. `IdemStore`
  naturally produces conflict / retry_later / in-flight outcomes that ARE taxonomy Kinds
  (KindConflict, KindIdempotencyInFlight); returning sentinels and re-mapping them in httpx would
  duplicate the taxonomy.
- **Decision:** `kernel/database` MAY import `kernel/errors` and return `*errors.Error` for
  outcomes that map cleanly to a Kind (idempotency, and version-conflict helpers may migrate to
  this too). The graph stays acyclic — `kernel/errors` imports only stdlib. Encoded a `depguard`
  rule in `.golangci.yml` (kernel must not import module/app/adapters/testkit) so the import law is
  machine-checked, not just documented.
- **Affected:** kernel/database/idempotency.go, .golangci.yml.

## D-0034 — Phase 3 review: idempotency review-finding resolutions
- **SEC-16/ARCH-27 (critical, reproduced):** the claim raced (SELECT-FOR-UPDATE cannot lock a
  non-existent row, so concurrent first-uses both went Fresh and the unconditional upsert clobbered
  a completed response). Rewritten to atomic `INSERT … ON CONFLICT DO NOTHING RETURNING` — only a
  real insert is Fresh; otherwise `SELECT … FOR UPDATE` and branch (completed→replay, hash
  mismatch→conflict, expired→re-claim, else in-flight). Concurrency regression test
  (`TestIntegrationIdempotencyConcurrent`, 8 goroutines, exactly-once, passes ×5 under `-race`).
- **SEC-18 (medium, reproduced):** `Recover` appended a problem body to already-written responses
  and swallowed `http.ErrAbortHandler`. Now tracks whether bytes were written (skips the problem
  body if so) and re-panics on ErrAbortHandler.
- **ARCH-32/SEC-23:** `WithIdempotency` now stores only 2xx responses; non-2xx claims are discarded
  (stay retryable) via the new `IdemStore.Discard`.
- **SEC-19:** `RequestHash` now includes the URL query string.
- **ARCH-29:** `DecodeJSON` rejects a literal `null` body like an empty one.
- **ARCH-31/SEC-22:** added `filtering.KeysetClause` (blueprint 05 §2, previously missing) with
  cursor-key allowlisting + `Sort.Terms()` accessors; columns come only from the sort allowlist,
  cursor supplies only bound values.
- **ARCH-34:** `RequireIfMatch` rejects `*` (optimistic concurrency requires a concrete version).
- **Accepted/deferred:** ARCH-28/SEC-21 (Router.Err() enforced at boot) → Phase 5 app wiring;
  ARCH-35 (ScopeExtractor `any` → authz.Target) → Phase 4; SEC-20 (duplicate JSON keys / no
  Content-Type check) → defense-in-depth noted, strict decode + domain validation suffice.
- **Affected:** kernel/database/idempotency.go, kernel/httpx/{idempotency,middleware,decode,etag}.go,
  kernel/filtering/{sort,keyset}.go; evidence/phase-03/review-findings.md.

## D-0024 — Phase 2: TenantDB grows per-phase accessors; sentinel errors until kernel/errors
- **Context:** 05 §2's TenantDB carries Outbox()/Audit()/Resources(), owned by Phases 4/6; the
  error taxonomy arrives in Phase 3.
- **Decision:** Phase 2 TenantDB = DBTX only (D-0006 growth pattern; accessors land with their
  capabilities). Version-conflict/no-tenant failures are exported sentinel errors in
  kernel/database now and get mapped into the Phase 3 taxonomy when it exists.
- **Affected:** kernel/database; revisit notes in phase-plan rows 3/4/6.

## D-0025 — Phase 2: only kernel migrations 000–001 ship; RLS proven on probe tables
- **Context:** tenants/users/user_tenant_access (001) are GLOBAL tables — RLS-bearing kernel
  tables start at migration 002+ (later phases). Phase 2 must still prove the RLS mechanics.
- **Decision:** ship 000 (extensions, roles, `app_tenant_id()`), 001 (tenants/users/access) per
  phase plan; `testkit.AssertRLSIsolation` + integration tests create standard-convention probe
  tables (tenant_id + ENABLE/FORCE + policy) to prove SET LOCAL binding, isolation, WITH CHECK,
  and no-tenant-context failure. Each later migration adding tenant tables reuses the same
  assertion catalog-driven.
- **Affected:** migrations/, testkit/asserts.go, kernel/database integration tests.
