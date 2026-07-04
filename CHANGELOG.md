# Changelog

All notable changes to wowapi are documented here. wowapi is a domain-neutral, reusable Go platform
kernel distributed as a third-party dependency (`go get github.com/qatoolist/wowapi`) with an installable
CLI (`go install github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z`).

The format is based on [Keep a Changelog](https://keepachangelog.com); this project is pre-1.0 (v0),
so the public surface (`kernel` / `module` / `app` / `adapters` / `testkit` / `migrations` + `cmd/wowapi`)
may still make breaking changes between minor versions.

## [Unreleased]

Hardening pass against ROADMAP-wowapi.md (see `docs/implementation/hardening-plan.md`). Phase H1:

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

### Fixed
- Retention sweep legal-hold race: a hold applied concurrently with `SweepRetention` could be voided.
  The candidate scan now locks rows `FOR UPDATE` and the void re-asserts `legal_hold = false`.

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
