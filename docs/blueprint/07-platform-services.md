# 07 — Security, Performance, Concurrency, Documents, Notifications, Webhooks, Events/Jobs, API, Observability

## 1. Security as a framework primitive

Middleware chain (fixed order, kernel-owned):
`RealIP → RequestID → Recover → OTel → SecureHeaders → CORS → BodyLimit(1MB default) → Timeout(30s)
→ AuthN(JWT/OIDC) → TenantResolve → CapacityResolve → RateLimit → AuthZ(RouteMeta) → handler`.

| Control | Mechanism (kernel) |
|---|---|
| AuthN | OIDC discovery + JWKS cache; `aud`/`iss`/`exp` strict; revocation hook (`TokenRevoked(jti) bool` port — Redis/DB set later); MFA = IdP concern, `amr` claim surfaced for policies (`env.mfa == true` conditions). |
| Tenant guardrail | tenant-scoped queries impossible without `SET LOCAL` (TenantDB is the only door) + RLS FORCED as DB backstop = defense in depth. |
| AuthZ guardrail | route registration refuses missing `RouteMeta`; `Public: true` is a searchable explicit marker reviewed in PRs. |
| Record-level | `Evaluator.Filter` pushes constraints into SQL; single-record checks in services. |
| Rate limiting | token bucket per (tenant) and (tenant,actor); in-memory per pod v1 (limits are guardrails, not billing), Redis adapter later; 429 + Retry-After. |
| Input | strict JSON decode (unknown fields rejected), body limits, param validation, allowlisted filters/sorts (injection impossible), sqlc = parameterized always. |
| Headers/CORS | `nosniff`, `frame-ancestors 'none'`, HSTS; CORS allowlist per env config. CSRF: token APIs (Bearer) exempt; any future cookie session must enable the kernel's SameSite+token middleware. |
| Uploads | presigned PUT only (no proxying bytes through API), MIME sniff + extension check on confirm, size from rule point `core.upload.max_size_mb`, checksum verify, `OnFileUpload` scan hook (ClamAV adapter; `scan_status` gate before download), download always via short-lived presigned GET + audit row. |
| Webhooks | signature verify before parse (per-provider verifier), replay window + `external_event_id` unique, secrets via secret provider refs. |
| Secrets | `secrets.Provider` port (env for dev, cloud secret manager in prod); config structs hold refs; `String()` redaction on secret types; no secrets in logs test. |
| Encryption | TLS everywhere (LB-terminated + enforced), Postgres/S3 encryption at rest (managed), column-level `pgcrypto` only for designated PII fields (party contacts) if compliance demands. |
| Masking | slog `ReplaceAttr` redactor for known keys (email, phone, token) + `logsafe.String` wrapper types. |
| Audit | sensitive actions auto-audited via TenantDB (writers baked into tx); denials of `sensitive` permissions always audited; impersonation & break-glass double-logged (see 01). |
| OWASP API top-10 | BOLA → record-level filters; broken auth → OIDC + revocation; unrestricted resource consumption → body/rate/timeout/page-size caps; BFLA → RouteMeta + deny-default; SSRF → outbound webhook URL validation (deny private ranges); misconfig → secure defaults + config validation at boot. |

## 2. Performance guardrails

Budgets (CI benchmark suite fails on 2× regression): authenticated simple GET p95 < 50ms / < 30
allocs in framework code; authz evaluate (cached assignments) < 1ms; middleware chain overhead < 1ms.

- **Optimize early (it's structural):** connection pooling (pgx pool sized `min(4×cores, db_max/replicas)`), per-query timeouts, keyset pagination everywhere lists can grow, allowlisted filters (prevents accidental table scans), COPY for batch, prepared statements (pgx auto-cache), JSON encoding via stdlib now with an `encoding` seam (goccy/sonic swap-in later if profiles demand).
- **Do NOT optimize early:** caching domain reads (introduce per proven hot spot, tenant-keyed, event-invalidated), read replicas, materialized views (add per painful report), sharding.
- **Hot-path rules:** no reflection, no per-request registry lookups (resolve at boot), pre-compiled route table, assignment snapshot cache (30s) for authz, rules cache (60s + event invalidation). Async everything non-critical: notifications, webhooks, projections ride outbox/jobs — request tx stays minimal.
- **N+1 defense:** repos expose batch getters (`GetByIDs`), list queries join what they need; review checklist item; `slowquery` log (pgx tracer, >100ms warn) catches escapes.
- **pprof** on internal admin port always; continuous profiling optional later. `make bench` runs the budget suite.

## 3. Concurrency & async primitives (`kernel/jobs`, `kernel/outbox`)

```go
type Job interface { Kind() string }                       // payload struct implements
type Worker[T Job] interface { Work(ctx context.Context, j T) error }
type Registry interface { RegisterKind(kind string, w AnyWorker, rp RetryPolicy) }
type Runner interface {                                    // wraps River
    Enqueue(ctx context.Context, db database.TenantDB, j Job, opts ...Opt) error // SAME TX as business write
    EnqueueGlobal(ctx context.Context, j Job, opts ...Opt) error
    Schedule(kind string, cron string)                     // periodic kinds
}
type RetryPolicy struct { MaxAttempts int; Backoff BackoffPolicy }   // default: 5 attempts, exp+jitter (1s→5m)
type BackoffPolicy func(attempt int) time.Duration
// DeadLetter: exhausted jobs land in River's discarded state; kernel mirrors to job_runs(status=dead)
// + metric + admin API for inspect/requeue.
type ProgressTracker interface { Set(ctx, done, total int64, note string) error } // bulk ops → job_runs.progress
```

Fixed worker pools per queue (default 10; per-kind override), bounded by config — **no unbounded
goroutines anywhere** (lint: `go` keyword allowed only in kernel jobs/outbox/httpx packages).
Backpressure = queue depth; producers never block (DB insert), consumers pace themselves.
Every worker: `SET LOCAL` tenant from payload → inbox-dedupe (for event handlers) → work in tx →
ctx cancellation respected → graceful drain on shutdown (in-flight finish, deadline, then release).

**Parallelize:** outbox relay (batch claim `FOR UPDATE SKIP LOCKED`), notification fan-out (per
delivery), webhook deliveries (per endpoint), bulk import (chunked jobs + ProgressTracker), report
generation, SLA sweeps (per tenant). **Do NOT parallelize:** inside a business transaction (splitting
one aggregate write across goroutines destroys atomicity), request-path fan-out to own DB (pool
exhaustion for microseconds), event handlers for the same aggregate (serialize via per-aggregate
advisory lock `pg_advisory_xact_lock(hash(resource_id))` when ordering matters). `singleflight` only
for cache fill of expensive shared reads (JWKS, rule bundles).

## 4. Document / file / comment / attachment framework — key flows

Upload: `POST /documents` (metadata) → `POST /documents/{id}/versions:initiate` → `UploadSessionResponse`
(presigned PUT, tenant-prefixed key) → client uploads → `:confirm` → kernel verifies size/checksum/MIME
sniff, runs `OnFileUpload` hooks (scan enqueued; `scan_status=pending` blocks download for
`sensitivity>=confidential`), writes immutable `document_versions` row + event.
Download: authz (document class policy + grants + `granted_via` relationships) → `OnDocumentAccess`
hooks (watermark slot) → 302 to presigned GET (60s) → audit row `document.downloaded`.
Retention: `retention_until` from document-class rule point; nightly sweep voids expired versions
(storage delete + tombstone row; legal-hold flag blocks). Deletion = voiding; hard erasure = explicit
redaction job (GDPR path) that also redacts comments.
Comments/attachments: plain services against `ResourceRef` (see DDL); comment edit keeps history in
audit detail; void ≠ delete.

## 5. Notification framework

`notify.Send(ctx, db, Message{TemplateKey, RecipientParty, Vars, Importance, ResourceRef})` writes
`notifications` + one `notification_deliveries` per resolved channel **in the business tx**; actual
sending is async (jobs). Channel resolution: party contact points ∩ per-party preferences
(rule point `notify.preferences` default) ∩ template availability; template lookup tenant→platform,
locale fallback chain (`hi-IN → hi → en`). Templates are Go `text/template` with an allowlisted
variable set per template key (unknown vars fail seed validation, not send time). Providers behind
`ChannelSender` ports (smtp, sms, whatsapp, push adapters); retries per RetryPolicy; exhausted →
`dead` + metric; `importance=legal` deliveries additionally write an audit row with provider receipt.
In-app channel = rows queried by `/notifications` API + unread counts. Synchronous send exists only
for security-critical flows (break-glass alert) via direct adapter call with 2s timeout + async fallback.

## 6. Webhook & integration framework

**Inbound:** `HandleWebhook(provider)` route factory (Public + `webhook:{provider}` actor):
verify signature (provider verifier from registry) → replay check (`external_event_id` unique +
timestamp window ±5m) → persist `webhook_events` → ack 200 fast → process async via job (handler
registered by module). Failure to verify → 401 + audit + metric (no body logging).
**Outbound:** modules/tenants register endpoints + subscribed event types; outbox dispatcher fans
matching events to delivery jobs: HMAC-SHA256 signature header (`X-Signature`, `X-Timestamp`,
`X-Event-Id`), 10s timeout, retry policy (5 attempts exp backoff), circuit breaker per endpoint
(open after 5 consecutive failures, half-open probe 5m) — breaker state in memory + endpoint
`status=degraded` persisted; dead → admin requeue API.
**Integration providers:** `integration.Provider` adapter interface per kind; anti-corruption layer
rule: provider payloads are translated to kernel/module types at the adapter — provider types never
cross into services. Credentials only as secret refs; per-tenant enablement rows; health checks per
provider surfaced in readiness detail (non-fatal).

## 7. Event conventions

- Naming `module.resource.verb_past` (`core.rule.version_activated`, `requests.request.approved`).
- Envelope = outbox row: `id (uuidv7)`, `type`, `schema_version`, `tenant_id`, `resource`, `actor`,
  `occurred_at`, `payload`. Payload structs live in module `domain/events.go`; additive evolution
  only within a schema_version; breaking → bump version, handlers declare versions they accept.
- Consumers are idempotent by construction (`processed_events` inbox keyed handler+event_id).
- Ordering: guaranteed only per aggregate (relay dispatches in `occurred_at` order per resource);
  handlers must not assume global order.

## 8. REST API conventions

- Base: `/v1` (URI versioning; additive within v1, breaking → v2 side-by-side).
- Tenancy: `/v1/t/{tenantSlug}/…` for tenant resources; `X-Tenant` header honored for machine
  clients; platform admin under `/v1/platform/…` (separate authz).
- Resources plural kebab; verbs only as `:action` suffix for non-CRUD transitions
  (`POST /requests/{id}:approve`, `:initiate`, `:confirm`).
- Pagination: cursor default (`?cursor=&limit=`, max 100); offset only on admin endpoints.
  Filtering `?filter[status]=active&filter[created_at][gte]=…` (allowlist); sort `?sort=-created_at`.
  Search `?q=` where a module implements it.
- Concurrency: `ETag: "v<version>"` on GET; `If-Match` required on PUT/PATCH/DELETE of versioned
  aggregates (else 428); mismatch → 412.
- Idempotency: `Idempotency-Key` header honored on all POSTs with `RouteMeta.Idempotent`.
- Bulk: `POST /…/bulk` ≤100 items sync (`BulkResponse`) else 202 → operation. Long-running:
  202 + `Location: /v1/operations/{id}` (backed by job_runs).
- Kernel endpoint groups (society module adds `/society/*` later, kernel unchanged):
  `tenants, organizations, users, parties, capacities, resources, relationships, roles, permissions,
  assignments, policies, rules (+versions:activate), workflows (definitions, instances, tasks:decide),
  documents (+versions, grants), comments, attachments, notifications, audit-logs (read-only),
  integrations, webhooks (+deliveries:redeliver), jobs/operations (admin), healthz/readyz (public)`.
- OpenAPI: module fragments merged by `wowapi openapi merge --check`; CI diff-checks spec vs routes.

## 9. Observability & operations

- **Logs:** slog JSON; every line carries `request_id, tenant_id, actor_id, capacity_id, trace_id,
  module`; canonical log line per request (method, route, status, dur, bytes).
- **Traces:** OTel spans: middleware → handler → service ops → SQL (pgx tracer) → hooks → job spans
  linked to originating request via event id.
- **Metrics (Prometheus):** RED per route; pool stats; `outbox_pending`, `outbox_dispatch_lag_seconds`
  (alert >60s), job queue depth/failed/dead, workflow open tasks + SLA breaches, notification
  delivery failure rate, webhook breaker state, authz denials, rate-limit drops.
- **Health:** `/healthz` liveness (process); `/readyz` readiness (DB ping, migrations current,
  registries validated, config valid — includes the redacted `config_fingerprint`); module checks
  via `ctx.Health(...)`; workers expose readiness on admin port. Shared-section fingerprint drift
  between api and worker raises an alert ([12](12-configuration-and-deployment.md) §7).
- **Ops (small-team pragmatic):** single Dockerfile (distroless, multi-stage); docker-compose for
  dev (pg + minio + mailpit); production = one managed Postgres (PITR backups, tested restore
  runbook) + 2× api + 1× worker containers on a managed runtime (Cloud Run / ECS / Fly / k8s if
  already owned); CI = lint (golangci-lint + boundaries) → unit → integration (testcontainers) →
  race → bench-budget → build/push; migrations as release step (`cmd/migrate`) before rollout;
  expand-contract keeps them backward-compatible with N-1 pods. Configuration and deployment
  layering (typed layered config, secret references, per-process views, compose/k8s rendering,
  CI config gates) is specified in [12-configuration-and-deployment.md](12-configuration-and-deployment.md).
  Audit export: monthly partition dump to object storage (immutable bucket).
  DR: restore-from-backup rehearsal + `RTO 4h / RPO 15m` targets documented.
