# Changelog

All notable changes to wowapi are documented here. wowapi is a domain-neutral, reusable Go platform
kernel distributed as a third-party dependency (`go get github.com/qatoolist/wowapi`) with an installable
CLI (`go install github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z`).

The format is based on [Keep a Changelog](https://keepachangelog.com); this project is pre-1.0 (v0),
so the public surface (`kernel` / `module` / `app` / `adapters` / `testkit` / `migrations` + `cmd/wowapi`)
may still make breaking changes between minor versions.

## [Unreleased]

Hardening pass against ROADMAP-wowapi.md (see `docs/implementation/hardening-plan.md`) — phases H1, H2,
and selected H5/P1 items. All domain-neutral; each shipped behind the `make ci` + `make ci-container` gate.

### Hardening remediation — corrective actions CA-1…CA-15

Closure of the exit-gate gaps found by `VERIFICATION-wowapi-hardening.md` (status matrix in its §6):

- **Metrics actually emit** (CA-1): `kernel.Deps.Metrics`; RED middleware + `/metrics` in the generated
  api; scheduler-lag, webhook-breaker, rate-limit-drop, config-fingerprint gauges; reference Prometheus
  alerts/scrape in `deployments/reference/`.
- **Secure-by-default wiring** (CA-2): rate limiter in the default chain (opt-out `http.rate_limit`); real
  `telemetry.trace_sample_ratio` wired to the OTel adapter; composite API-key+OIDC authenticator; opt-in
  authz assignment cache; signed keyset cursors in `workflow.OpenTasksFor` and the generated CRUD list.
- **API keys** (CA-3): `apikey.Store.Rotate`, audited issue/rotate/revoke, and a `wowapi apikey` CLI.
- **Advisory-lock load envelope** (CA-4) documented with a repeatable test.
- **Module recurring jobs** (CA-5): `module.Context.RecurringJob`.
- **Audit** (CA-11): `module.Context` accessors for audit/sequence/bulk/artifact; `audit_logs.tx_id`
  (migration 00023); `wowapi audit verify` CLI.
- **Correctness/security nits**: idempotency replay-after-expiry now errors (410, CA-8); gate step-up
  test (CA-13); `integration.Config.Credential` is a redacted `config.Secret` (CA-14); an unregistered
  notification channel fails terminally instead of silently reporting `sent` (CA-15).
- Hosted CI workflow added (`.github/workflows/ci.yml`, CA-6). Remaining open/rescoped items (async trace
  propagation, read-replica routing, O2/O3/O5 finishers) are tracked in VERIFICATION §6.

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
