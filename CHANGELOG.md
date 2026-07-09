# Changelog

All notable changes to wowapi are documented here. wowapi is a domain-neutral, reusable Go platform
kernel distributed as a third-party dependency (`go get github.com/qatoolist/wowapi`) with an installable
CLI (`go install github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z`).

The format is based on [Keep a Changelog](https://keepachangelog.com) and this project follows
[Semantic Versioning](https://semver.org). As of **v1.0.0** the public surface
(`kernel` / `module` / `app` / `adapters` / `testkit` / `migrations` + `cmd/wowapi`) is stable: breaking
changes to it require a new major version.

## [Unreleased]

### Added
- **Rule registry definitions lifecycle + expanded schema validation (`kernel/rules.SyncDefinitions`)**
  — closes the gap that forced `wowsociety` to hand-mirror its Go rule-point declarations into a SQL
  migration and run a drift-guard test against it (GAP-007). `rules.SyncDefinitions(ctx, db, registry)`
  upserts every registered `Point` into `rule_definitions` idempotently (schema, default value, allowed
  scopes, approval requirement, description all converge on re-sync; no duplicate rows), satisfying the
  `rule_versions.rule_key` foreign key without any product SQL. It is the rule-registry analogue of
  `kernel/seeds.Sync` and runs at the same lifecycle point: the generated `cmd/migrate` calls it
  immediately after `seeds.Sync`, on the same `app_platform`-privileged connection. There is no
  standalone `wowapi rules sync` framework CLI subcommand — unlike seed catalogs (declarative YAML the
  framework can load off disk), rule points exist only as Go declarations inside a booted product
  process, so a product with a custom migrate main calls `SyncDefinitions` itself the same way. Also
  expands `kernel/rules`' focused schema validator (`kernel/rules/schema.go`) with the common JSON
  Schema bounds keywords: `minimum`, `maximum`, `exclusiveMinimum`, `exclusiveMaximum`, `minLength`,
  `maxLength`, `pattern`, `minItems`, `maxItems`, and a shallow `required`-property-presence check for
  object values — so `Store.Propose` now rejects an out-of-bounds numeric/string/array value at write
  time instead of a product re-implementing the same bounds check (as `wowsociety`'s `rulepoints.go`
  `checkValue` did). Full recursive JSON Schema (nested `properties` sub-schemas, `additionalProperties`,
  `items` sub-schemas) stays explicitly out of scope — the validator's doc comment narrows the contract
  rather than silently under-supporting it. All prior validator behavior (top-level `type` + `enum`)
  is preserved unchanged. Documented in the user guide (Database & Migrations → Rule definitions).
- **Scoped privileged framework services (`kernel/privileged`, `module.Context.Privileged()`)** —
  a sanctioned, audited surface for the valid tenant-scoped operations that require *platform*
  privilege at the database, so a product module no longer has to write its own `SECURITY DEFINER`
  SQL bridge (SEC-24 / SEC-13). `Privileged().Relationships()` grants/revokes ReBAC relationship
  edges (`Grant`/`Revoke`); `Privileged().Rules()` activates tenant-scope rule versions
  (`ActivateTenant`, with an optional atomic product `Gate`). Each service is bound to the calling
  module and enforces, in Go, every check the product bridges performed: tenant binding (fail-closed
  when unbound), relationship-type / rule-key **ownership** (module-name prefix or a declared
  allow-list), subject/object **resource existence** in the bound tenant, **scope restriction**
  (tenant-scope only, of the caller's tenant), soft-revoke that preserves history, double-revoke /
  double-activate conflict detection, and an **audit** row on the kernel hash chain. Operations run
  in a tenant-bound `app_platform` transaction, so the existing `relationships` / `rule_versions`
  RLS and the one-active-per-instant `EXCLUDE` constraint still enforce isolation and arbitrate
  races — no module ever sees a platform pool or raw SQL door. Migration `00030` grants
  `app_platform` the minimum it needs to run the services as a role (SELECT on `acting_capacities`;
  SELECT/INSERT on `audit_logs`; INSERT/UPDATE on `audit_chain`) while leaving the `relationships`
  and `rule_versions` table protections untouched. Documented in the user guide
  (Modules → Privileged services).
- **Locale negotiation + API i18n (`kernel/i18n`, `httpx.Locale`, `module.Context.I18n`)** —
  cross-cutting localization for synchronous API responses so every product built on the framework
  handles `Accept-Language` and localized errors consistently instead of re-implementing risky
  translation plumbing (GAP-001). `kernel/i18n` adds a `Catalog` (in-process `(locale, key)` map with
  deterministic fallback to a default locale, and a final fallback to the key so a missing translation
  never breaks a response), a `Registry` that ships the framework's own **English** catalog (problem
  titles + validation-tag messages, under the reserved `kernel.` namespace) and accepts module bundles
  under their `<module>.` prefix, and an RFC 9110 §12.5.4 `Negotiate` (q-values, primary-subtag match).
  `httpx.Locale(catalog)` middleware negotiates the request locale, binds it to the request context,
  and sets `Content-Language`; `httpx.WriteError` localizes problem `title` and `kernel/validation`'s
  `StructCtx` localizes field messages — **machine `code`s and field paths stay byte-stable across
  locales**, and internal logs stay technical English. `module.Context.I18n(bundle)` registers a
  module's localized bundle; `app.Boot` merges all bundles with the framework catalog and exposes it as
  `Booted.I18n` to pass to `httpx.Locale`. **Zero-config behavior is unchanged**: with no catalog wired,
  responses stay English-only, byte-for-byte as before. `testkit` adds `AssertNegotiatedLocale`,
  `NewLocaleRequest`, and `AssertLocalizedProblem`. English is the default locale and ultimate fallback;
  product translations stay in product-owned bundles (the framework ships only its own English strings).
  Documented in the user guide (Validation & Error Handling → Localizing responses).
- **`adapters/storage/s3`** — production S3/MinIO object-storage adapter implementing the
  `kernel/storage.Adapter` port (presigned PUT/GET with TTL clamping, checksum-verified `Stat`,
  ranged `Peek`, idempotent `Delete`, `KindNotFound` mapping, path-style addressing, fail-closed
  bucket validation with optional local/dev auto-create). Integration-tested against real MinIO
  (gated on `WOWAPI_REQUIRE_S3`, mirroring the DB gate) plus a Memory↔S3 contract test and a full
  document upload/confirm e2e. Adds `github.com/minio/minio-go/v7` as a direct dependency. Wiring
  is documented in the user guide (Build & Deploy → Object storage); framework config/scaffold
  wiring is a tracked follow-up.
- **Step-up/MFA seedability + JWT `amr` propagation**: `seeds.PermissionSeed.StepUp` (`step_up` YAML key,
  strict-decoded) lets a module declare a step-up-gated permission in its seed catalog instead of
  registering it directly against `authz.Registry`; `app.Boot` propagates it into `authz.Permission.StepUp`
  and `seeds.Sync` persists it to a new `permissions.step_up` column (migration `00029`, idempotent —
  re-syncing after the seed changes updates the existing row). `auth.Claims` gains the standard `amr`
  claim (RFC 8176) and `Verifier.Actor` copies it onto `authz.Actor.AMR`, so a product's JWT authenticator
  gets step-up enforcement without reparsing the bearer token to recover `amr` itself. Adds
  `testkit.WithAMR(...string)` alongside the existing token options. Closes the two framework gaps
  `wowsociety`'s identity module previously worked around (direct `StepUp` registration + a manual
  catalog migration + a JWT-reparsing authenticator wrapper).

### Fixed
- `wowapi init` next-steps hint no longer implies a bare `make migrate-up` works — it needs `APP_ENV` + the DB
  DSNs + a running Postgres (fail-closed). The hint now points to the generated README's "Getting started".

## [1.0.0] — 2026-07-06 — first stable release

wowapi is now **API-stable and production-hardened**. Beyond the 0.1.0 framework build, this release adds
exhaustive multi-tenant isolation hardening, safe-by-default RLS enforcement, generated-scaffold correctness
fixes, a full enterprise supply-chain release pipeline, and a clean, enforced lint gate.

### Security & tenant-isolation hardening

- **Safe-by-default RLS enforcement**: `app.Boot` fails closed if the runtime **or** platform pool's effective
  role can bypass row-level security (superuser / `BYPASSRLS`), so a misconfigured DSN cannot ship an
  RLS-inert deployment. Backstops the per-connection (`WithConnRLSGuard`) and per-tx (`WithRLSGuard`) guards;
  connect-time guards are wired on every serving pool (api/worker runtime + platform, and the `dlq` CLI).
- **Tenant-isolation footgun hardening**: rate-limit keys are tenant-prefixed (no cross-tenant bucket
  collapse); webhook `DispatchOutbound` binds the tenant from the event and fails closed on a mismatch;
  leader-safe per-tenant `notify`/`webhook` pollers; `jobs_queue` / `job_runs` gain RLS + FORCE with a strict
  `app_tenant_id()` `WITH CHECK`; and a table-driven RLS-isolation census guards every tenant table.
- Error strings stored in `last_error` / `job_runs` are truncated on a **UTF-8 rune boundary** (never invalid
  UTF-8 that a Postgres `text` column would reject).

### Generated product scaffold — correctness

- `wowapi init <name>` takes a positional name that creates a new `./<name>/` directory and scaffolds the
  product inside it (name defaults from it); flags may appear before or after it. The flag-only form
  (`wowapi init --module … --dir …`, scaffold into `--dir`) still works.
- `wowapi config validate --env <e>` now honours `--env` (previously validated whatever `APP_ENV` pointed at)
  and fails closed when the composed environment doesn't match.
- `cmd/migrate` fails closed on a bad config and has real `up` / `down` — `down` is a guarded full reset,
  **refused outside local/dev** (production schema change is forward-only, expand-contract).
- `config diff` is delegated to the product checker with the secret provider wired; `deploy render` emits all
  three DSN references (runtime + **platform** + migrate) so the rendered manifest actually boots; the api
  middleware sets security/CORS headers even on rate-limited (429) responses; and the scaffold README's
  getting-started works out of the box.

### Delivery & supply chain

- **Tag-driven release**: GoReleaser builds cross-platform CLI archives + checksums + SBOMs, **cosign**
  keyless-signs them, and attaches **SLSA** build provenance; a multi-arch distroless GHCR image ships with
  provenance + SBOM attestations. Hosted CI adds actionlint, CodeQL, Scorecard, govulncheck, and secret scanning.
- **Lint gate closed and enforced**: the pre-existing golangci-lint backlog is fully burned down
  (`make lint` = 0); golangci-lint and actionlint are pinned; the **full-tree** `make lint` is the enforced CI gate.
- **Reference-stack header smoke** in CI: a scaffolded product runs behind the reference nginx over TLS and its
  security-header posture is smoke-tested through the proxy.
- Test coverage is enforced at a **90 % floor** against the real database.

### Earlier hardening pass (ROADMAP-wowapi.md — H1/H2 + selected H5/P1)

All domain-neutral; each shipped behind the `make ci` + `make ci-container` gate.

### Hardening remediation — corrective actions CA-1…CA-15

Closure of the exit-gate gaps found by `VERIFICATION-wowapi-hardening.md` (status matrix in its §6):

- **Metrics actually emit** (CA-1): `kernel.Deps.Metrics`; RED middleware + `/metrics` in the generated
  api; scheduler-lag, webhook-breaker, rate-limit-drop, config-fingerprint, and **DLQ-depth**
  (`dlq_depth{queue}`, leader-safe) gauges; reference Prometheus alerts/scrape in `deployments/reference/`.
- **Secure-by-default wiring** (CA-2): rate limiter in the default chain (opt-out `http.rate_limit`); real
  `telemetry.trace_sample_ratio` wired to the OTel adapter; composite API-key + OIDC/JWT authenticator —
  the OIDC user leg (`kernel/auth`: JWKS RS256/ES256 verifier + `Authenticator`, DB principal via
  `adapters/auth/pgprincipal`) activates when `auth.oidc` is configured in the product, else the composite
  falls through to deny-by-default (`DenyAllAuthenticator`); opt-in authz assignment cache with
  `InvalidateAll` wired into `seeds.Sync` so a spine change is reflected immediately (D-0079); signed keyset
  cursors in `workflow.OpenTasksFor` and the generated CRUD list.
- **Runtime/platform privilege separation** (CF-1, second review): the generated api/worker **fail closed**
  when `db.platform_dsn` is unset instead of reusing the runtime DSN + `SET ROLE app_platform` (which would
  require the cluster-global `app_rt → app_platform` membership); the product-dev box uses a dedicated
  `app_platform` login, and `authz.TestIntegrationRuntimeRoleNotMemberOfPlatform` guards against poisoning
  (D-0078).
- **Perf budgets extended** (B-2): benchmarks + budgets for audit Record/chain, `sequence.Allocate`, the HTTP
  token bucket, `authz.CachingStore`, and the edge middleware chain; `make bench-budget` enforces 30 benches.
- **API keys** (CA-3): `apikey.Store.Rotate`, audited issue/rotate/revoke, and a `wowapi apikey` CLI.
- **Advisory-lock load envelope** (CA-4) documented with a repeatable test.
- **Module recurring jobs** (CA-5): `module.Context.RecurringJob`.
- **Audit** (CA-11): `module.Context` accessors for audit/sequence/bulk/artifact; `audit_logs.tx_id`
  (migration 00023); `wowapi audit verify` CLI.
- **Correctness/security nits**: idempotency replay-after-expiry now errors (410, CA-8); gate step-up
  test (CA-13); `integration.Config.Credential` is a redacted `config.Secret` (CA-14); an unregistered
  notification channel fails terminally instead of silently reporting `sent` (CA-15).
- Hosted CI workflow added (`.github/workflows/ci.yml`, CA-6). A second independent review then closed the
  remaining items — async trace propagation now covers jobs + notify (CA-9), the audit anchor-export ships
  (CA-11), and the reversibility/PITR/object-storage restore drills are scripted (CA-12); read-replica
  routing remains a recorded deployment-concern rescope. Full per-item status lives in
  [docs/GOALS-TRACKER.md](docs/GOALS-TRACKER.md) §4.

### Added
- `kernel/httpx` edge middleware — `SecureHeaders`, `CORS`, `BodyLimit`, `Timeout` — completing the
  blueprint's fixed chain. `BodyLimit`/`Timeout` now enforce the previously dead `http.max_body_bytes`
  / `http.request_timeout` config. New `http.cors_allowed_origins` config (deny-by-default allowlist).
  The generated api wires the full chain.
- Keyset cursors can carry a sort-spec signature (`pagination.EncodeCursorWithSig`, `Cursor.Sig`,
  `filtering.Sort.Signature`, `filtering.NextCursor`); `KeysetClause` now rejects a cursor minted under
  a different sort **order/direction** with a validation error instead of silently returning wrong pages.
- `database.IdemStore.SweepExpired` — cross-tenant purge of expired idempotency keys (migration 00012).
- Native fuzz targets for the filter DSL parser and cursor decoder (`make test-fuzz`).
- Reference reverse-proxy deployment (`deployments/reference/`) and an operations deployment checklist
  (security headers, TLS, config-drift alerting convention).
- Dead-letter-queue operability: `wowapi dlq <jobs|events> <list|inspect|replay|discard>` and the
  kernel admin functions behind it (`jobs.{ListDead,ReplayDead,DiscardDead}`,
  `outbox.{ListDeadEvents,ReplayDeadEvent,DiscardDeadEvent}`). Migration 00013 grants app_platform
  DELETE on the queue tables.
- Notification delivery receipts: `notify.Service.Deliveries(notificationID)` returns per-channel
  delivery status + provider message ids (RLS-scoped), making delivery queryable per notification.
- Distributed-tracing seam: `kernel/observability.Tracer`/`Span` port + `NoOpTracer` + a `Trace`
  HTTP middleware (server span per request), wired into the generated api with the NoOp tracer
  (zero-cost when disabled). The OpenTelemetry SDK binding is a thin adapter (kernel stays otel-free).
- Authz decision caching: `authz.CachingStore` (opt-in `Store` decorator) caches the hot
  `ActiveAssignments` read per (tenant, actor) with a short TTL + explicit `Invalidate` so a role
  revoke applies immediately (no stale-allow). Unwrapped, behavior is unchanged.
- Step-up / MFA hooks: `authz.Permission.StepUp` + `authz.Actor.AMR`; `Evaluate` challenges an
  otherwise-allowed decision (`Decision.StepUpRequired`) when no strong auth factor is present, and
  the httpx gate returns 401 + `WWW-Authenticate: … step_up="mfa"`. `env.mfa` added as an ABAC attribute.
- Snapshot/artifact pipeline (`kernel/artifact`, migration 00021): immutable, per-(tenant,kind)
  versioned artifacts with sha256 content hash, structured sidecar, and template-by-effective-date
  resolution; `Verify` re-hashes to detect tampering. The product supplies the rendered bytes
  (e.g. PDF/A) — no document-format library in the kernel.
- Data lifecycle (`kernel/retention`, migration 00020): generalized legal hold over any entity
  (`Place`/`Release`/`IsHeld`/`List`, not just documents) + a DSR ledger (`Open`/`Complete`/`Reject`)
  for export/erasure requests with a statutory-override reason.
- Machine authentication (`kernel/apikey`, migration 00019): issuable, scoped, rotatable, revocable,
  expirable API keys / service principals (only sha256(secret) stored). `apikey.Authenticator`
  satisfies the httpx gate port; a verified key becomes an `ActorSystem` whose scopes authorize it
  via a new machine fast-path in `authz.Evaluate` (a scope acts like an RBAC grant, still subject to
  ABAC deny). `authz.Actor` gains a `Scopes` field.
- Audit tamper-evidence (hash-chaining, migration 00018): each audit row carries a per-tenant seq +
  `row_hash = sha256(prev_hash ‖ row)`; `audit.Verify` recomputes the chain and detects any mutation
  (hash mismatch) or deletion (seq gap); `audit.Anchor` exports the head for external notarization.
- Durable field-level audit trail (`kernel/audit`, migration 00017): append-only `audit_logs` with a
  `Record`/`Query` API capturing entity/field/before/after/actor/request-id + a per-record redaction
  hook. Append-only is grant-enforced (app_rt has no UPDATE/DELETE). Basis for S6 hash-chaining.
- Bulk-operation framework (`kernel/bulk`, migration 00016): chunked processing of large item sets
  with progress reporting, a partial-failure ledger, and resumability — each item isolated in its own
  transaction (success commits atomically with the done mark; a failure rolls back but is ledgered).
- Gap-free per-tenant sequence allocator (`kernel/sequence`, migration 00015): transactional
  statutory numbered series (receipts/vouchers/certificates) with audited voids — gap-free (a
  rolled-back tx frees the number) and race-free (concurrent allocations serialize), replacing
  hand-rolled `MAX()+1`.
- Recurring scheduler (`jobs.Scheduler` + `schedules` table, migration 00014): fixed-interval kernel
  maintenance tasks, leader-safe across worker replicas via an atomic per-row claim. Wires the workflow
  SLA sweep (per-tenant) and the idempotency-key expiry sweep so they now actually run on a schedule
  without N replicas double-firing.
- In-process rate limiting: `kernel/httpx.RateLimit` middleware + `TokenBucket` limiter
  (`NewTokenBucket`), `KeyByIP` / `KeyByActor` strategies, 429 + `Retry-After` + RFC 7807. Opt-in.
- Migration reversibility: `database.MigrateReset` (goose down-to-0) and a CI forward→down→forward drill
  (`TestIntegrationMigrationsReversible`). Operations docs for zero-downtime expand/contract migrations
  and a backup/restore runbook + `scripts/backup_restore_drill.sh`.

- Data-lifecycle disposition/DSR engine (`kernel/retention.Registry`/`Engine`): per-record-class
  Dispose/Export/Erase callbacks, a scheduled per-tenant disposition sweep, and DSR export/erasure
  fulfilment. Exposed via `module.Context.RetentionClasses()`.
- OpenTelemetry tracing adapter (`adapters/tracing/otel`) with a configurable ratio sampler and an
  OTLP exporter; the `Tracer` port gains `Inject`/`Extract` (W3C traceparent) and the HTTP middleware
  continues an inbound trace. Compose stack gains a Jaeger backend (OTLP + UI).
- Per-user notification channel preferences (`notify.SetChannelPref`, migration 00022).
- `wowapi config` delegates to the product `tools/configcheck`; new `config diff` subcommand.
- `db.platform_dsn` config so the worker can use a dedicated app_platform login.

### Fixed
- Authorization denials are now written to the durable `audit_logs` (not only WARN-logged):
  the default sink writes an `authz.denied` row in its own tenant tx.
- Retention sweep legal-hold race: a hold applied concurrently with `SweepRetention` could be voided.
  The candidate scan now locks rows `FOR UPDATE` and the void re-asserts `legal_hold = false`.
- Migration 00010 down: it created `app_actor_id()` but never dropped it, so a rollback + re-apply
  failed ("function already exists"). Caught by the new reversibility drill.

## [0.1.0] — 2026-07-04 — initial framework release

The first complete framework build (Goal 2, Phases 0–12). Everything below is domain-neutral: the
kernel contains no product/society-specific concepts.

### Data & tenancy
- PostgreSQL with **row-level security** as the tenancy boundary: the runtime connects AS a non-superuser
  `app_rt` login (never superuser + SET ROLE), RLS FORCE + fail-closed `app_tenant_id()`. A separate
  `app_platform` role owns cross-tenant/catalog work.
- `kernel/database`: pgx pool, `TxManager`/`TenantDB` (SET LOCAL role + tenant), per-source goose
  migrations, RLS guard (rejects superuser/BYPASSRLS).
- `kernel/model`, `kernel/config` (typed layered config: defaults ← base ← env overlay ← env vars ←
  secret references; fingerprint; per-process narrowed views; shared-section **drift detection**).

### HTTP, errors, primitives
- `kernel/httpx`: router with route metadata (permission required unless `Public`), RFC 9457 problem
  details, middleware chain (request-id, recover, metrics, access log), idempotency, ETag, keyset
  pagination, filtering, liveness `/healthz` + readiness `/readyz` (reports the config fingerprint).
- `kernel/errors` taxonomy; `kernel/validation`; `kernel/pagination`; `kernel/filtering`.

### Identity & authorization
- `kernel/auth` (OIDC verification); `kernel/authz` **deny-by-default** evaluator (RBAC scope-covering
  assignments → ReBAC `granted_via` relationships → ABAC deny-first policies); `kernel/policy`;
  `kernel/relationship`; `kernel/resource`. The authz spine is SELECT-only to the module role — a module
  cannot self-grant.

### Module SDK & composition
- `module` (the public contract) + `app` (the composition root): `kernel.New` + `App.Boot` register
  modules in dependency order against a capability-scoped `Context`, gate the whole graph, and load
  ownership-checked seed bundles. Inter-module access is via declared ports only.
- `testkit`: `RunModuleContract`, one-line fixtures, RLS assertions, Admin/Runtime/Platform pools —
  importable and passing from an external product repo.

### Async & platform services
- `kernel/outbox`: transactional outbox (event iff business commit), relay (cross-tenant read as
  app_platform, per-tenant dispatch, per-aggregate ordering via advisory lock, event DLQ), inbox
  (exactly-once effect). `kernel/jobs`: Postgres job runner (retries/backoff/DLQ, at-least-once).
- `kernel/rules`: versioned, temporally-resolved rule engine (org-ancestry → tenant → platform → default;
  draft→activate approval gating; write-time schema validation).
- `kernel/workflow`: closed step-type set, boot-validated definitions, optimistic-locked runtime with
  same-tx outbox, authz-gated override, SLA sweeper.
- `kernel/document` + `kernel/storage`: object-storage port (presigned upload/download), append-only
  versioned file pointers, deny-first authorized downloads, malware-scan gate, retention sweep + legal
  hold. `kernel/comment`, `kernel/attachment`.
- `kernel/notify`: templates (html-escaped email), transactional send + async delivery with backoff/DLQ.
  `kernel/webhook`: inbound verify/replay + outbound HMAC delivery with retry + per-endpoint circuit
  breaker. `kernel/integration`: provider registry with secret-reference credentials + health checks.

### Observability, performance, security
- `kernel/observability`: metrics port (RED per route, counters, gauges) + NoOp; access-log middleware;
  `adapters/metrics/prometheus` (`/metrics`).
- Performance budgets enforced in CI (`internal/tools/benchbudget`); hot paths are reflection/lookup-free.
- Curated security suite (`make test-security`): RLS isolation, deny-by-default authz, privilege
  boundaries, secret redaction (logs/CLI/dumps), per-knob unsafe-config prod rejection. Core guarantees
  have no disabling config key.

### Tooling — the `wowapi` CLI
- `init` (scaffold a framework-wired product repo: api/worker/migrate binaries), `new-module`,
  `gen crud` (gofmt-clean generated Go), `migrate create`, `seed validate`, `openapi merge`,
  `lint boundaries`, `deploy render`, `config init|validate|doctor|print|schema`, `version`
  (go.mod mismatch warning).

### Delivery
- Container-first CI (`make ci` / `make ci-container`): vet + boundary lint + unit + race + perf budgets
  + build. Single Dockerfile; docker-compose dev stack (Postgres + MinIO + Mailpit + tools runner).
- Full evidence trail under `docs/implementation/` (decision log D-0001…D-0058, per-phase evidence
  bundles, the 28-criterion acceptance map).
