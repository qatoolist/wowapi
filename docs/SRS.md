# wowapi — Software Requirements Specification (SRS)

- **Product:** `wowapi` — a reusable, domain-agnostic enterprise backend framework ("platform kernel") in Go.
- **Module path:** `github.com/qatoolist/wowapi`
- **Status:** pre-1.0 (`v0.x`); public API additive-frozen at `v1.0.0`.
- **Document status:** Living. Synthesized from the original prompt/vision files (`Goal.md`, `Goal 1.1.md`, `Goal 1.2.md`, `Goal 2.md`), the authoritative design blueprint (`docs/blueprint/00–12`), the hardening tranche (`ROADMAP-wowapi.md`, `CHANGELOG.md`, `VERIFICATION-wowapi-hardening.md`), and cross-checked against the implemented code (50 commits; 251 Go files; 108 test files; 24 migrations; Go 1.26).
- **Companion:** [GOALS-TRACKER.md](GOALS-TRACKER.md) — what is done / deferred / pending, with the full backlog.

> **Provenance note.** The original vision (`Goal.md`) and blueprint §00–12 define the framework core. A later
> **compliance-hardening tranche** (`ROADMAP-wowapi.md` §14.x, verified in `VERIFICATION-wowapi-hardening.md`) added
> the evidence primitives — audit hash-chain, gap-free sequences, retention/DSR + legal hold, artifact pipeline,
> machine API keys, and step-up auth. Those requirements are tagged **[H]** below so traceability stays accurate.

---

## 1. Introduction

### 1.1 Purpose
This SRS captures the enduring vision and requirements for `wowapi`, independent of any single AI prompt or
conversation. It is the durable record of *what wowapi is, what it must do, and the constraints it operates under*,
so the project's intent survives beyond the working files that seeded it.

### 1.2 Scope
`wowapi` is a **modular-monolith backend framework** consumed as a **versioned third-party Go dependency**. It
provides the hard-to-get-right, cross-cutting spine of a multi-tenant enterprise SaaS — tenancy + isolation,
authorization, workflow/approvals, rules/config, audit, transactional messaging, jobs, notifications, documents,
observability — once, correctly, so that many products build on one kernel instead of forking per product.

It is **not** a product. The housing-society domain that motivated it is a *reference domain only*; no
society-specific concept may live in the kernel.

### 1.3 Definitions
| Term | Meaning |
|---|---|
| **Kernel** | The domain-blind framework core (`wowapi/kernel/*`). |
| **Module** | A product domain plugged in via the public SDK, living in a *consuming* repo. |
| **Actor** | An authenticated principal in an acting capacity — a user-in-capacity **or** a system/webhook principal. |
| **Tenant** | An isolated customer/org boundary; enforced by `tenant_id` + PostgreSQL RLS. |
| **RouteMeta** | Per-route metadata (permission, public, scope, idempotent, sensitive) that gates registration. |
| **Outbox** | Transactional-outbox table written in the same tx as the business change. |
| **RLS** | Row-Level Security (PostgreSQL), the primary tenant-isolation mechanism. |

### 1.4 References
- Vision: `Goal.md` (40 deliverable sections + patterns appendix).
- Distribution refinement: `Goal 1.1.md` → `docs/blueprint/11-framework-distribution-and-consumption.md`.
- Config/deployment refinement: `Goal 1.2.md` → `docs/blueprint/12-configuration-and-deployment.md`.
- Build plan: `Goal 2.md` (13 phases, Phase 0–12; 25-point Definition of Done).
- Authoritative design: `docs/blueprint/00-overview.md` … `12-configuration-and-deployment.md`.
- Hardening: `ROADMAP-wowapi.md`, `CHANGELOG.md`, `VERIFICATION-wowapi-hardening.md`.
- User docs: `README.md`, `docs/user-guide/*`. Quality culture: `docs/working/quality-gates.md`.

---

## 2. Overall Description

### 2.1 Product Perspective
`wowapi` is distributed as a Go module (`go get github.com/qatoolist/wowapi@vX.Y.Z`) plus an installable CLI
(`go install github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z`). A product application lives in **its own repository**,
imports the public packages, registers its domain modules through the module SDK, and ships one binary set
(`api`, `worker`, `migrate`) against **one PostgreSQL database**.

**Architecture style:** modular monolith with a domain-neutral kernel; hexagonal (ports & adapters) *at the edges
only* (DB, object storage, IdP, mail/SMS, payment, secrets); compile-time modules (no `.so` plugins); manual
dependency injection (no reflection container / service locator). Microservices, event-sourcing, CQRS-everywhere,
and a low-code runtime were explicitly rejected for v1, while preserving future extraction seams.

**Public surface (importable by consumers):** `kernel/*`, `module`, `app`, `adapters/*`, `testkit`, `migrations`,
`cmd/wowapi`. **Private:** `internal/*` (compiler-blocked for external imports) — no consumer-facing contract may
live under `internal/`.

### 2.2 Product Functions (summary)
Multi-tenancy + RLS · identity/authn integration · actor + acting-capacity model · RBAC/ReBAC/ABAC authorization +
policy · relationship graph · resource registry · workflow/approval engine · rule/config engine · audit (append-only,
tamper-evident **[H]**) · transactional outbox + events · background jobs + DLQ + leader-safe scheduler · notifications ·
webhooks (in/out) · integration providers · documents/files/comments/attachments · gap-free sequences **[H]** ·
retention/DSR/legal-hold **[H]** · artifacts **[H]** · machine API keys **[H]** · step-up auth **[H]** · config +
secrets (5 typed layers) · observability · module SDK · CLI/codegen · testkit.

### 2.3 User Classes
- **Framework maintainers** — build/evolve the kernel; own the boundary and versioning discipline.
- **Product/module developers (primary consumer)** — build a module in ~a day, CRUD in ~an hour, from their own repo.
- **Operators / small teams** — deploy boringly: one DB, a few processes, managed Postgres with PITR.
- **End-actor archetypes** (drive authz requirements): org admin, manager/approver, self-service user, time-boxed
  external auditor, vendor (assigned items only), system worker actor, webhook actor, multi-capacity user.

### 2.4 Constraints (see §5 for full list)
Go 1.26 · PostgreSQL 16 (RLS mandatory) · pgx/v5 + sqlc · goose migrations · River jobs · one-way import law
(kernel ← module ← product; adapters implement kernel ports) · pre-1.0 versioning discipline.

### 2.5 Assumptions & Dependencies (product-supplied, via adapters/hooks — not built into core)
External OIDC/JWT IdP (framework does authZ, consumes authN) · S3-compatible object storage (presigned up/download) ·
notification providers (email/SMS/WhatsApp/push) · integration/payment providers · malware-scan hook (e.g. ClamAV) ·
secret provider (env in dev, cloud manager in prod).

---

## 3. Functional Requirements

Each requirement has a stable ID (`FR-<area>-<n>`). **[H]** = hardening-tranche origin. Traceability to code and
tests is summarized in [GOALS-TRACKER.md](GOALS-TRACKER.md); package citations point at the implemented kernel.

### 3.1 Multi-Tenancy & Isolation (`kernel/database`)
- **FR-TEN-1** Shared-DB / shared-schema multitenancy keyed by `tenant_id` on every tenant-scoped table.
- **FR-TEN-2** PostgreSQL RLS `ENABLE` **and** `FORCE` on every tenant table; policy
  `tenant_id = current_setting('app.tenant_id')::uuid` for both `USING` and `WITH CHECK`.
- **FR-TEN-3** Every unit of work issues `SET LOCAL app.tenant_id` (and `app.actor_id`) as the first statement of
  the transaction. **No tenant in context ⇒ error, never a silent fallback** (fail-closed).
- **FR-TEN-4** Separate DB roles: `app_rt` (runtime), `app_migrate` (DDL), `app_platform` (break-glass). Runtime
  role has **no** UPDATE/DELETE on audit tables (DB-enforced immutability).
- **FR-TEN-5** `TenantDB` is the only door to tenant data; global tables (tenants, users, permissions) are
  kernel-service access only. Catalog-driven `AssertRLSIsolation` verifies isolation for every registered table.
- **FR-TEN-6** A `TxManager` escape hatch preserves schema-per-tenant / db-per-tenant as a future option without
  changing call sites.

### 3.2 Authorization (`kernel/authz`, `kernel/policy`, `kernel/relationship`)
- **FR-AUTHZ-1** **Deny-by-default.** A permission must exist in the registry or boot fails; default authenticator
  denies all until wired.
- **FR-AUTHZ-2** Layered evaluation: registry check → active assignments whose scope covers the target
  (`tenant ⊃ org-subtree ⊃ resource_type ⊃ resource`) within validity window → RBAC allow → ReBAC via declared
  `granted_via` relationship → ABAC policies (deny-overrides, priority) → record Decision.
- **FR-AUTHZ-3** `Evaluator.Filter` pushes record-level constraints **into SQL** (no load-then-filter).
- **FR-AUTHZ-4** Actors may be user-in-capacity **or** system/webhook principals; system actors are granted roles —
  there is **no bypass path**.
- **FR-AUTHZ-5** Special flows: scoped delegation; **break-glass** (≤60 min, reason required, event + notify);
  audited **impersonation** (may not approve or change security posture); emergency workflow override + ratification.
- **FR-AUTHZ-6** Denials of **sensitive** permissions are audited.
- **FR-AUTHZ-7 [H]** Machine principals: issuable, scoped, rotatable, revocable, expirable **API keys**
  (`kernel/apikey`, migration `00019`); only `sha256(secret)` stored; composite OIDC+API-key authenticator;
  audited issue/rotate/revoke; `wowapi apikey` CLI.
- **FR-AUTHZ-8 [H]** **Step-up / MFA**: `Permission.StepUp` + `Actor.AMR`; evaluation returns
  `Decision.StepUpRequired` when no strong factor present; HTTP gate replies `401` +
  `WWW-Authenticate: … step_up="mfa"`; `env.mfa` ABAC attribute. API-key actors carry no AMR ⇒ can never satisfy
  step-up (fail-secure by construction).

### 3.3 HTTP Layer (`kernel/httpx`)
- **FR-HTTP-1** **RouteMeta gate:** a route cannot register without a permission unless `Public:true`; setting both
  `Public` and `Permission` is a boot failure.
- **FR-HTTP-2** Fixed kernel middleware chain:
  `RealIP → RequestID → Recover → OTel → SecureHeaders → CORS → BodyLimit(1MB) → Timeout(30s) → AuthN →
  TenantResolve → CapacityResolve → RateLimit → AuthZ → handler`.
- **FR-HTTP-3** **Idempotency** for POSTs marked `RouteMeta.Idempotent`: key scoped `(tenant, actor, key)`; store
  request hash + response; same key+hash ⇒ replay; same key+different hash ⇒ 409; in-flight ⇒ `retry_later` (409);
  replay after expiry ⇒ 410.
- **FR-HTTP-4** **Rate limiting**: token buckets per `(tenant)` and `(tenant, actor)`; per-tenant quota from rule
  point `core.rate_limit.rpm`; `429` + `Retry-After`; in default chain (opt-out via `http.rate_limit`).
- **FR-HTTP-5** **Pagination**: cursor/keyset default (`?cursor=&limit=`, max 100); offset admin-only; filters/sorts
  strictly allowlisted (parameterized — injection impossible); `ETag`/`If-Match` (428/412) on versioned aggregates.
- **FR-HTTP-6** Responses: RFC 9457 problem details for errors; typed envelopes `APIResponse[T]`, `CursorPage[T]`,
  `OperationResponse` (202 + Location), `BulkResponse`, `UploadSessionResponse`.
- **FR-HTTP-7** Security posture middleware: secure headers, CORS allowlist, body-size limit, request timeout — all
  in-chain and unit-tested (`kernel/httpx/edge_test.go`).

### 3.4 Module SDK (`module`)
- **FR-MOD-1** A module implements `Name() / DependsOn() []string / Register(ctx Context) error`.
- **FR-MOD-2** `module.Context` is **capability-scoped** — registries + services only, never raw pools:
  `Routes, Permissions, Roles, ResourceTypes, RelationshipTypes, Rules, Workflows, Events, Jobs, DocumentClasses,
  NotificationTemplates, Hooks, Migrations(fs), Seeds(fs), Health, OpenAPI` + runtime deps
  `Tx, Authz, RulesResolver, WorkflowRuntime, Documents, Notify, Webhooks, Logger, Config, IDGen, Clock,
  Port/ProvidePort` + **[H]** `Audit, Sequence, Bulk, Artifacts, RetentionClasses`.
- **FR-MOD-3** A module may read **only** `modules.<name>.*` config via `ModuleView.Decode`; there is no API to read
  global or sibling config.
- **FR-MOD-4** Lifecycle: `Register → Validate → Migrate → SeedSync → Start → Stop`. **Validate** turns whole-graph
  problems into **boot failures**: duplicate permissions, routes without meta, unknown deps / dependency cycles,
  seed errors, unsatisfied ports, module-config errors.
- **FR-MOD-5** Modules communicate only via **declared ports** (`ProvidePort`/`Port`), never by importing each other.
- **FR-MOD-6** Composition is manual (`Kernel` + `App`), constructor injection, no container.

### 3.5 Transactional Outbox, Events, Jobs, Scheduler (`kernel/outbox`, `kernel/jobs`)
- **FR-JOB-1** The outbox row is written in the **same tx** as the business change (atomicity non-negotiable).
- **FR-JOB-2** Relay worker batch-claims with `FOR UPDATE SKIP LOCKED`, dispatches in-process, enqueues handler jobs.
- **FR-JOB-3** Jobs run on River (Postgres-backed) behind `jobs.Runner`; `Enqueue(ctx, db TenantDB, job)` shares the
  business tx.
- **FR-JOB-4** Retry with exponential backoff + jitter (default 5 attempts, 1s→5m); **DLQ**: discarded jobs land in
  `job_runs(status=dead)` + metric + admin requeue (`wowapi dlq`).
- **FR-JOB-5** Idempotent consumers via a `processed_events` inbox keyed `(handler, event_id)`.
- **FR-JOB-6** Worker executes: `SET LOCAL` tenant from payload → inbox dedupe → work in tx → graceful drain.
- **FR-JOB-7** No unbounded goroutines outside kernel async packages (lint-enforced).
- **FR-JOB-8** Per-aggregate **ordering**: relay dispatches in `occurred_at` order per resource; `pg_advisory_xact_lock`
  where handlers must serialize.
- **FR-JOB-9** **Leader-safe scheduler**: cron kinds (SLA sweep, retention sweep, webhook retry, digest) each claim
  with `FOR UPDATE SKIP LOCKED` and advance within the claim tx — exactly-once across replicas; each job is
  tenant-iterating and idempotent.

### 3.6 Rules & Workflow Engines (`kernel/rules`, `kernel/workflow`)
- **FR-RULE-1** Typed rule points + JSON-Schema-validated values; scope resolution `org → tenant → platform →
  code-default`; **historical `Resolve(at)`**; approval-gated activation; no-overlap exclusion constraint.
- **FR-RULE-2** Feature flags are `feature.*` rule points with rollout percentage.
- **FR-WF-1** Custom Postgres declarative workflow engine; versioned JSON definitions; step types
  `approval | task | auto | gateway | vote | terminal`; assignee resolution; SLA sweeper; delegation; emergency
  override + ratification; `Start` runs inside the caller's tenant tx.
- **FR-WF-2** Runtime gating is fail-closed on vote / `min_approvals` / self-approval; override requires an authz gate.

### 3.7 Data, Migrations, Seeds (`migrations`, `kernel/seeds`)
- **FR-DATA-1** goose migrations, per-module embedded FS; `app.RunMigrate` runs **kernel migrations first**, then
  product-module migrations topologically sorted by `DependsOn`; prefixed goose history rows coexist in one table.
- **FR-DATA-2** Breaking changes use expand-contract (N-1 pod compatible); the app refuses to serve on migration drift.
- **FR-DATA-3** Seeds are idempotent (`SeedSync`); seed errors fail boot validation.
- **FR-DATA-4** All migrations are reversible; a reversibility drill is part of the gate.

### 3.8 Compliance & Evidence Primitives **[H]**
- **FR-AUD-1** Append-only `audit_logs` (monthly partitions); Writer baked into `TenantDB` (same tx); runtime role
  lacks UPDATE/DELETE; every sensitive action + every sensitive-permission denial audited; impersonation/break-glass
  double-logged; rule-decision provenance (`VersionID`) recorded. Correlation `tx_id` (migration `00023`).
- **FR-AUD-2 [H]** **Tamper-evident hash chain** (migration `00018`): each row carries a per-tenant sequence +
  `row_hash = sha256(prev_hash ‖ row)`; `audit.Verify` recomputes the chain and detects any mutation/deletion; wired
  as the default audit sink. `wowapi audit verify` CLI (tamper detection verified end-to-end).
- **FR-SEQ-1 [H]** **Gap-free sequence allocator** (`kernel/sequence`, migration `00015`): per-tenant numbered series
  (receipts/vouchers/certificates); counter-row lock inside the business tx (rollback frees the number ⇒ gap-free);
  audited voids; proven under concurrency + rollback. Replaces the `MAX()+1` race.
- **FR-RET-1 [H]** **Retention / DSR / legal hold** (`kernel/retention`, migration `00020`): generalized legal hold
  over any entity (`Place/Release/IsHeld/List`); per-record-class retention/disposition registry
  (Dispose/Export/Erase callbacks) on the leader-safe scheduler; **DSR ledger** (`Open/Complete/Reject`) for
  export/erasure with statutory-override reason; exposed via `Context.RetentionClasses()`. Sweep re-checks holds
  in-tx (no hold-vs-sweep race).
- **FR-ART-1 [H]** **Artifact pipeline** (`kernel/artifact`, migration `00021`): dataset → immutable per-`(tenant,kind)`
  versioned artifact (sha256 content hash + structured sidecar + template-by-effective-date); `Verify` re-hashes;
  immutability grant-enforced. (Deviation: content stored in-row `bytea`, not object storage — decision D-0076.)
- **FR-BULK-1** Bulk import/export as chunked async jobs + `ProgressTracker`; `POST /…/bulk` ≤100 items sync, else
  202 → operation (`kernel/bulk`, migration `00016`).

### 3.9 Documents, Notifications, Webhooks, Integrations (`kernel/{document,attachment,comment,notify,webhook,integration}`)
- **FR-DOC-1** Document metadata + versions; storage adapter + fake; presigned upload→confirm→download; malware-scan
  hook; per-document grants; retention/redaction jobs (status lifecycle, not DELETE).
- **FR-NOT-1** `notify.Send(...)` writes `notifications` + per-channel `notification_deliveries` in the business tx;
  async send via jobs; channels in-app/email/SMS/WhatsApp/push; template resolution tenant → platform + locale
  fallback; `text/template` with allowlisted vars (unknown var ⇒ seed-validation fail); `importance=legal` writes
  audit + provider receipt; retries → DLQ. Unregistered channel **errors loudly** (never a silent no-op "sent").
- **FR-WH-1** Inbound webhooks: verify signature **before parse** → replay check (`external_event_id` unique + ±5 min
  window) → persist → 200 fast → async job.
- **FR-WH-2** Outbound webhooks: HMAC-SHA256 signed (`X-Signature/Timestamp/Event-Id`), 10s timeout, 5-attempt retry,
  **circuit breaker per endpoint** (open after 5 fails, half-open 5m), admin redeliver.
- **FR-INT-1** `integration.Provider` per kind with an anti-corruption layer (provider types die at the adapter);
  credentials only as secret references; per-tenant enablement.

### 3.10 Configuration & Secrets (`kernel/config`, `kernel/secrets`)
- **FR-CFG-1** **Five typed layers**: framework / product / module / deployment-env / tenant-runtime — kept separate,
  never mixed.
- **FR-CFG-2** Precedence: compiled defaults → `base.yaml` → `<env>.yaml` → env vars → secret refs → (flags
  local-only, refused in prod).
- **FR-CFG-3** **Fail-fast boot**: missing/unknown/out-of-range/unsafe-for-prod config aborts startup with the full
  error list.
- **FR-CFG-4** Config is **immutable after boot**; hot paths never re-read config; the only live tweak is log level
  via an audited admin endpoint; all other runtime/tenant change flows through the rule engine.
- **FR-CFG-5** **Secrets by reference**: `secretref://provider/path` resolved once at boot into `config.Secret` with
  **structural redaction** (`String/MarshalJSON/LogValuer` emit `[redacted:…]`; `Reveal()` is lint-restricted to
  adapters).
- **FR-CFG-6** Per-process narrowed views (api/worker/migrate) + `config_fingerprint` (SHA-256 over redacted config)
  with cross-process drift alerting.
- **FR-CFG-7** `environment` is fail-closed and non-downgradable (never from flags); core security guarantees have
  **no disabling config key**.

### 3.11 Observability (`kernel/observability`, `adapters/metrics`, `adapters/tracing`)
- **FR-OBS-1** slog JSON logs; every line carries `request_id, tenant_id, actor_id, capacity_id, trace_id, module` +
  a canonical per-request line.
- **FR-OBS-2** OTel traces span middleware → handler → service → SQL (pgx) → hooks → job (linked by event id);
  `telemetry.trace_sample_ratio` config key.
- **FR-OBS-3** Prometheus metrics: RED per route, pool stats, `outbox_pending` / `outbox_dispatch_lag_seconds`
  (alert > 60s), job depth/failed/dead, workflow SLA breaches, notify failure rate, webhook breaker state, authz
  denials, rate-limit drops.
- **FR-OBS-4** Health: `/healthz` (liveness), `/readyz` (DB ping, migrations current, registries validated, config
  valid + fingerprint).

### 3.12 CLI & Codegen (`cmd/wowapi`, `internal/cli`)
- **FR-CLI-1** Installable CLI with: `init`, `new-module`, `gen crud`, `migrate`, `seed validate`, `openapi merge`,
  `lint boundaries`, plus config tooling (`config init/validate/doctor/print --redacted/diff/schema`), `deploy render`,
  `apikey`, `audit verify`, `dlq`.
- **FR-CLI-2** Generated modules compile and pass the module contract; OpenAPI merge detects drift; boundary lint
  proves the import graph is acyclic and domain-clean.

---

## 4. External Interfaces

- **HTTP/REST** — RFC 9457 problem details; cursor pagination; typed envelopes; OpenAPI generated + drift-checked.
- **Database** — PostgreSQL 16 over pgx/v5; sqlc-generated queries; goose migrations; River job tables.
- **Object storage** — S3-compatible via `storage` adapter (presigned URLs); MinIO in dev.
- **IdP** — external OIDC/JWT (framework verifies tokens, owns authorization).
- **Notification/integration providers** — pluggable adapters; credentials as secret refs.
- **Secret provider** — env (dev) / cloud secret manager (prod) via `adapters/secrets`.
- **Telemetry** — OTLP traces; Prometheus scrape endpoint.

---

## 5. Non-Functional Requirements

### 5.1 Security (`docs/blueprint/10-delivery.md`, `09-patterns.md`)
Deny-by-default authz; fail-closed RLS (SET LOCAL + FORCE, defense-in-depth); route-metadata gate; append-only +
tamper-evident audit **[H]**; structural secret redaction (verified via `AssertNoSecretsInLogs`); core guarantees have
no disabling config key; OWASP API Top-10 mitigations. **Principle:** every security control has a *structural*
enforcement (type / middleware / DB) **and** a test.

### 5.2 Reliability
At-least-once delivery (transactional outbox + idempotent inbox); retry + backoff + jitter; DLQ; leader-safe
scheduler (exactly-once across replicas); graceful drain on shutdown; expand-contract migrations; crash-injection
test proves zero loss / zero duplicate. Targets: RTO 4h / RPO 15m.

### 5.3 Scalability & Performance
Stateless api/worker; pooled Postgres; async via outbox/jobs; per-aggregate ordering; keyset pagination everywhere;
caching added only at proven hot spots, tenant-keyed and event-invalidated (authz snapshot 30s, rules 60s +
invalidation). Goal: "10× tenants without redesign." Hot-path benchmark budgets gated (`bench-budgets.txt`).

### 5.4 Testability
Real-integration testing over mocks — Testcontainers Postgres, template-DB clone per test (< 60s/module); fakes only
at process/network boundaries (mail/SMS/push, scanner, IdP token minting, payment callbacks). Public `wowapi/testkit`
+ `RunModuleContract`; RLS / authz-matrix / audit assertions; fake clock + deterministic idgen via *production*
constructors. **CI gate runs entirely in containers** (`make ci` / `ci-container`) with `WOWAPI_REQUIRE_DB=1` so
DB-backed tests fail rather than silently skip.

### 5.5 Operability
`/healthz` + `/readyz`; single distroless image; compose (pg + minio + mailpit) for dev; prod = managed Postgres
(PITR) + 2× api + 1× worker; migrate as a pre-rollout job; backup/restore rehearsal runbook; monthly audit-partition
dump to an immutable bucket.

### 5.6 Maintainability & Governance
**Import law** (§5 below); module contract tests; ADR gate for any new kernel noun; small packages; composition over
inheritance; consumer-side interfaces; `wowapi lint boundaries` + `depguard` in CI; **quality gates** (`make
fmt-check/vet/lint-new/tidy-check/test`) wired through version-controlled git hooks + CI
(`docs/working/quality-gates.md`).

### 5.7 Domain-Neutrality (the defining NFR)
The kernel is **domain-blind**: "if a word in kernel code would mean something to a housing-society lawyer, it's a
bug." Enforced structurally by a lint **denylist** (building, wing, flat, member, committee, AGM, defaulter, parking,
visitor, conveyance, redevelopment, election, …) failing CI; by contract tests running the kernel against a *neutral*
fixture module; and physically, because product code lives in other repositories.

---

## 6. Constraints

- **C-1 Language:** Go 1.26 (blueprint floor Go ≥ 1.23; pin tightened in `go.mod`).
- **C-2 Database:** PostgreSQL 16; RLS **mandatory** for tenant isolation; pgx/v5 + sqlc; goose; River.
- **C-3 Import law (compiler + `wowapi lint boundaries`-enforced):**
  `L3 product modules → L2 wowapi/module → L1 wowapi/kernel/* ; L1 → L0 adapters (interfaces only)`.
  Kernel imports **no** module/app/adapters/testkit/examples/product code; `app` is the sole composition root;
  adapters implement kernel ports; product code imports only public packages (`internal/*` compiler-blocked);
  production code must not import `testkit`.
- **C-4 Distribution:** consumed as a versioned dependency; public surface is `kernel/module/app/adapters/testkit/
  migrations/cmd/wowapi`; no consumer contract under `internal/`.
- **C-5 Versioning:** pre-1.0 `v0.x` (surface may move); `v1.0.0` when `10-delivery.md` acceptance is green and the
  surface freezes to additive-only; `v2+` via `/v2` module path; deprecations survive ≥ 1 minor.
- **C-6 Architecture:** modular monolith; compile-time modules; hexagonal at edges only; manual DI; no reflection
  container / service locator; microservices/event-sourcing/CQRS-everywhere/low-code-runtime rejected for v1.

---

## 7. Acceptance Criteria (v1.0 gate — `docs/blueprint/10-delivery.md`, `Goal 2.md` Phase 12)
`make ci` passes entirely in containers; all blueprint acceptance criteria satisfied; no critical/high review
findings open; a **separate** product repo can `go get` wowapi, scaffold a module with the CLI, and have it compile +
pass the module contract without editing framework code; boundary lint proves no forbidden imports and no domain
leakage; api/worker/migrate E2E smoke green.

---

## 8. Traceability
Requirement → phase → code → test traceability, plus done/deferred/pending status, is maintained in
[GOALS-TRACKER.md](GOALS-TRACKER.md). Per-phase evidence bundles are archived in the `wowapi2` documentation
archive under `archive/evidence/phase-XX/` (see `docs/implementation/evidence/README.md` for the redirect map);
architectural decisions are recorded as `D-XXXX`; the hardening closure matrix is `VERIFICATION-wowapi-hardening.md`
§6 (archived to `wowapi2/archive/reviews/`; mirrored into the tracker). The retired prompt/hardening source files
named in this document's provenance notes live in the same archive — see GOALS-TRACKER §7.

> **Programme execution and verification (Waves 00–07, 2026-07-16).** The `impl/` directory contains the execution
> ledger for the Waves 00–07 programme (mandate, 8 waves, 75 stories, ~370 tasks, registers). An independent
> third-party audit (`impl/reports/implementation-autopsy-report-2026-07-16.md`, Fable 5) found **25 of 75 stories
> (33%) fully verified; the remaining 50 are incomplete, unreviewed, incorrectly implemented, or have contradictory
> status records.** The framework code that exists is largely good (lease/fencing, audit chain, release gating,
> online migration verified clean); the failure is one of **governance and truthfulness of completion claims**
> (statuses advanced without mandatory reviews, evidence records left unfilled or mis-pinned, a quality gate
> silently lowered). **Production-readiness claim rejected.** Remediation plan in `implementation-autopsy-report-2026-07-16.md`
> §13 (R-1 truth reconciliation is the gate for all else). Until W05–W07 genuinely close and the final gate runs,
> no production-readiness claim should be made. The SRS requirements themselves remain architecturally sound; the
> gap is execution and verification discipline, not the specification.
