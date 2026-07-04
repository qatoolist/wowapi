# wowapi — Skills & Knowledge Map

The concrete capabilities required to work on **this** codebase well. Each entry names the real packages,
files, and invariants involved and the traps observed in prior work. Not generic advice — grounded in the
kernel/module/app layering, the RLS tenancy model, and three review passes' findings.

> Companion docs: [best-practices.md](best-practices.md) · [working-persona.md](working-persona.md) ·
> [quality-gate-checklist.md](quality-gate-checklist.md) · [review-learning-register.md](review-learning-register.md) ·
> [internal-scripts-policy.md](internal-scripts-policy.md). Start at [README.md](README.md).

## 1. Architecture & layering

- **The layers and the import law.** `kernel/` (services), `module/` (the product-facing SDK contract),
  `app/` (composition root: `kernel.New` + `App.Boot`), `adapters/` (vendor bindings — prometheus, otel,
  secrets), `testkit/` (public test harness), `migrations/` (embedded goose SQL). Modules talk to the
  kernel only through `module.Context`; inter-module access is via declared ports. Enforced by
  `scripts/lint_boundaries.sh` (aka `make lint-boundaries` / `wowapi lint boundaries`) — run it after any
  new package or cross-package import.
- **Domain-neutrality.** The kernel carries no product/society concepts. Anything domain-specific belongs
  in a product repo (`wowapi init` scaffolds one). Blueprint: `docs/blueprint/`.
- **Composition root wiring.** `kernel/kernel.go` builds every service and the shared registries
  (`Perms`, `Resources`, `RetentionClasses`, `DocumentClasses`, …) that modules register into during
  `Boot`. **Trap:** a new kernel service is not usable until it is (a) a `Kernel` field, (b) exposed on
  `module.Context` (`app/context.go` + `module/module.go`), and (c) wired in `app/boot.go`. Missing any
  of these = "built-but-not-wired" (see learning register).

## 2. Data & tenancy (PostgreSQL + RLS)

- **RLS is the tenancy boundary.** Runtime connects AS the non-superuser `app_rt` login (never superuser
  + SET ROLE), tables use `FORCE ROW LEVEL SECURITY`, and `app_tenant_id()` is fail-closed. `app_platform`
  is the cross-tenant kernel role (relay, job runner, scheduler, sweeps); `app_migrate` owns DDL. Every
  tenant-scoped table needs: `ENABLE`+`FORCE` RLS, a `USING (tenant_id = app_tenant_id()) WITH CHECK (…)`
  policy, and the right grants (`GRANT … TO app_rt`).
- **Append-only via grants.** For audit/artifact tables, grant `app_rt` only `SELECT, INSERT` (no
  UPDATE/DELETE) — the runtime cannot rewrite history. Proven by tests that assert app_rt UPDATE/DELETE is
  denied (`kernel/audit`, `kernel/artifact`).
- **Cross-tenant reads/writes** need a permissive platform policy (`… TO app_platform USING (true)`) +
  the matching grant, mirroring `outbox_relay_all` (migration 00007). Examples: idempotency sweep (00012),
  DLQ delete grants (00013), api_keys verify (00019).
- **Transactions.** `database.TxManager`: `WithTenant` (RW, SET LOCAL role+tenant), `WithTenantRO` (READ
  ONLY — cannot INSERT; this is why the durable audit sink writes in its **own** tx), `Platform`
  (cross-tenant, no tenant bound). Do work in the caller's tenant tx so business writes + kernel writes
  (outbox event, audit row, sequence number) commit atomically.

## 3. Migrations

- **Every migration ships Up AND Down.** The reversibility drill (`TestIntegrationMigrationsReversible`
  via `database.MigrateReset`) runs forward→down→forward in `make ci-container`; a missing/incorrect Down
  fails CI (it already caught the 00010 `app_actor_id` bug). Down drops exactly what Up created — never
  cluster-scoped roles/extensions.
- **Register the file.** Add each `NNNNN_*.sql` to `expectedFiles` in `migrations/migrations_test.go`
  (numbers are contiguous). `miscellaneous/check_migrations.sh` verifies this.
- **Expand/contract for zero downtime** on live/journal tables: `docs/operations/migrations.md`.

## 4. Authentication & authorization

- **Deny-by-default evaluator** (`kernel/authz/evaluator.go`): RBAC scope-covering assignments → ReBAC
  `granted_via` relationships → ABAC deny-first policies. Extensions added and how they compose:
  **machine scope** (API keys: `Actor.Scopes`, allow if in scope, still subject to ABAC deny),
  **step-up** (`Permission.StepUp` + `Actor.AMR` → `Decision.StepUpRequired`), **caching**
  (`authz.CachingStore`, opt-in, explicit `Invalidate` to avoid stale-allow).
- **Runtime enforcement, not boot validation.** `httpx.SecureHandler`/`gateRoute` enforces `RouteMeta`
  per request (AuthN → bind tenant/actor → AuthZ). `DenyAllAuthenticator` is the fail-closed default.
  Machine auth = `kernel/apikey` (sha256 secret, cross-tenant verify as app_platform).
- **Denials are durably audited.** `kernel.durableAudit` (default sink) writes an `authz.denied`
  `audit_logs` row; sensitive-perm denials + explicit policy denies + break-glass/impersonation are always
  audited (`maybeAudit`).

## 5. HTTP, validation, errors, primitives

- `kernel/httpx`: router + `RouteMeta`, RFC 9457 problem details (`WriteError` maps error Kinds →
  status), the fixed middleware chain (`RequestID`→`Recover`→`Trace`→observability→`SecureHeaders`→`CORS`→
  `BodyLimit`→`Timeout`), idempotency, ETag, keyset pagination, filtering.
- **Error taxonomy** (`kernel/errors`): `KindValidation`/`Conflict`/`NotFound`/`Unauthenticated`/
  `Forbidden`/`RateLimited`/`Internal`/… → HTTP status. Use `kerr.E`/`kerr.Wrapf`; check with
  `kerr.KindOf`. Never leak internals on 500.
- **Keyset pagination** (`kernel/pagination` + `filtering.KeysetClause`): cursor encodes the LAST RETURNED
  row; carries an optional sort-spec signature so a changed sort fails loudly. **Trap:** off-by-one at page
  boundaries (shipped once) — always test the boundary.
- **Filter/sort DSL is injection-proof by construction**: client text only ever selects an allowlisted
  column/operator; values are always `$N` placeholders. Fuzzed (`make test-fuzz`).

## 6. Async & platform services

- **Transactional outbox** (`kernel/outbox`): event iff business commit; relay dispatches cross-tenant as
  app_platform with per-aggregate ordering (advisory lock) + event DLQ; inbox `processed_events` gives
  exactly-once handler effects. **Jobs** (`kernel/jobs`): Postgres runner, at-least-once (workers MUST be
  idempotent), retries/backoff/DLQ. **Scheduler** (`jobs.Scheduler` + `schedules` table): fixed-interval,
  leader-safe via atomic `FOR UPDATE SKIP LOCKED` + `next_run_at<=now` claim.
- **DLQ operability**: `jobs.{ListDead,ReplayDead,DiscardDead}`, `outbox.{…DeadEvents}`, `wowapi dlq` CLI.
- **Kernel maintenance tasks** are wired in `app/maintenance.go` (SLA sweep, idempotency sweep, retention
  disposition) as per-tenant scheduler fan-outs.

## 7. Compliance / evidence primitives (the "hand-rolled badly" set)

- `kernel/sequence` — gap-free per-tenant numbered series (receipts/vouchers), race-free (row lock), voids
  audited. Not a Postgres sequence (nextval doesn't roll back → gaps).
- `kernel/audit` — durable field-level audit + `Query` + `Redactor` + **hash-chaining** (`Verify` detects
  mutation/deletion, `Anchor` exports head).
- `kernel/retention` — generalized legal hold + DSR ledger + the disposition/DSR **engine**
  (`Registry`+`Engine`, product-supplied Dispose/Export/Erase callbacks — no dynamic-table SQL).
- `kernel/bulk` — chunked, resumable bulk ops with a partial-failure ledger.
- `kernel/artifact` — immutable versioned artifacts (hash + sidecar + template-by-effective-date);
  product supplies the rendered bytes (e.g. PDF/A) — no document-format lib in the kernel.

## 8. Config & deployment

- **Layered config** (`kernel/config`): defaults ← base.yaml ← env overlay ← env vars ← secretrefs;
  `config.Secret` is compiler-redacted (only `Reveal()`, lint-restricted); `Fingerprint` for drift. DSNs:
  `db.dsn` (app_rt), `db.migrate_dsn` (app_migrate), `db.platform_dsn` (app_platform — worker).
- **The CLI can't import product types** — `wowapi config validate|print|schema|doctor` delegate to the
  product-local `tools/configcheck` (scaffolded by `wowapi init`); framework-only fallback. `config diff`
  is framework-side.
- **Deployment**: `deployments/compose.yaml` (postgres/minio/mailpit/jaeger/tools), reference nginx +
  smoke, `docs/operations/deployment-checklist.md` (edge headers, drift alert, tracing, backup, migrations,
  rate limits).

## 9. Observability, performance, security testing

- `kernel/observability`: `Metrics` + `Tracer` ports, both NoOp-default and zero-cost when unwired;
  adapters in `adapters/{metrics/prometheus,tracing/otel}`. Tracing: `Trace` middleware, W3C traceparent
  `Inject`/`Extract`, Jaeger in compose (`:16686`).
- **Perf budgets** enforced in CI (`internal/tools/benchbudget` vs `bench-budgets.txt`) — hot paths stay
  reflection/lookup-free.
- **Security suite** (`make test-security`): RLS isolation, deny-by-default, privilege boundaries, secret
  redaction, unsafe-in-prod config rejection. Core guarantees have no disabling config key.

## 10. Testing, review & traceability discipline

- **TDD, real integration tests over mocks**; DB tests run against real Postgres and MUST run (not skip)
  under `WOWAPI_REQUIRE_DB=1` — the authoritative gate is `make ci-container`. `testkit` gives isolated
  per-test DBs + fixtures + `RunModuleContract`.
- **Traceability**: every deviation → `docs/implementation/decisions.md` (D-00NN) BEFORE code; every phase
  → an evidence bundle (`proof-bundle.md`, `review-findings.md`, `command-log.md`, `acceptance-map.md`)
  under `docs/implementation/evidence/`; `CHANGELOG.md` (Keep-a-Changelog, `[Unreleased]`).
- **Independent Review Gate** before any goal is complete (skill `independent-review-gate`; checklist
  [quality-gate-checklist.md](quality-gate-checklist.md)). Learnings feed
  [review-learning-register.md](review-learning-register.md).

## 11. Meta-skills (how to interpret and review)

- **Product-owner interpretation** — read the goal/roadmap literally; enumerate every sub-requirement;
  "partial/follow-up/deferred" ≠ done unless explicitly scoped out.
- **Development-architect review** — trace the full entry→effect path; a primitive without its adapter +
  wiring + infra is half-done; verify claims by running tests/artifacts/wiring, not by trusting docs.
- **Anti-hallucination** — do not invent packages/APIs/fields/config keys; grep/read first
  (`mcp__lumen__semantic_search` before Grep). Do not duplicate existing tests or implementations.
