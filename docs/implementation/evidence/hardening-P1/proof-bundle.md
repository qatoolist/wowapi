# Hardening P1 — proof bundle

Self-contained P1 items from [../../hardening-plan.md](../../hardening-plan.md), each individually
QA-gated. (The large P1 items — R1 authz cache, O1 OTel — remain; this bundle accumulates the small,
self-contained ones.)

## S2 — Rate limiting (D-0064)

| Verdict | Fix |
|---|---|
| real (P1) — only middleware hooks, delegated to a proxy; no in-process limiter | `kernel/httpx.RateLimit` middleware + `TokenBucket` limiter (`NewTokenBucket`); 429 + `Retry-After` + RFC 7807 (`KindRateLimited`, already in the taxonomy); `KeyByIP` / `KeyByActor` key strategies; idle-bucket sweep bounds memory |

Enforceable in-process per the roadmap: per-principal (`KeyByActor`, after the authz gate),
per-permission (custom keyFn on expensive routes), per-IP (`KeyByIP`, edge). Opt-in — limits are
product-specific; wiring documented in `docs/operations/deployment-checklist.md` §5.

Tests (`kernel/httpx/ratelimit_test.go`): burst-then-limit, refill over time, independent keys, 429 body
+ Retry-After header, allow path, actor-key with IP fallback.

Gate: `make ci` + `make ci-container` green — 0 FAIL, 0 SKIP, 74 packages.

## R5 — Notification delivery receipts (D-0067)

| Verdict | Fix |
|---|---|
| partial (roadmap overstated "fire-and-forget") — delivery status WAS tracked in `notification_deliveries`, but there was no query API to read it per notification | `notify.Service.Deliveries(ctx, db, notificationID) []DeliveryReceipt` — per-channel status, attempts, provider message id (receipt), last error, timestamps; RLS-scoped to the caller's tenant |

Closes the concrete R5 gap ("delivery status queryable per notification; provider receipts stored").
**Per-user channel preferences** were added in the post-hardening review (D-0077):
`notification_channel_prefs` (migration 00022) + `notify.SetChannelPref`; `Send` skips a recipient's
opted-out channels.

Test (`kernel/notify/notify_test.go::TestIntegrationDeliveriesReceipts`): a 2-channel send yields 2
receipts (inapp + email) with the email destination + queued status. Gate green (76 packages).

## S3 — Step-up / MFA hooks (D-0073)

| Verdict | Fix |
|---|---|
| real (P1) — the token was the only factor; the authz layer could not demand elevated auth per permission | `authz.Permission.StepUp` marks a permission as requiring MFA; `authz.Actor.AMR` carries the surfaced auth-methods-references; `authz.Evaluate` turns an otherwise-allowed decision into a step-up challenge (`Decision.StepUpRequired`, reason `step_up_required`) when the AMR has no strong factor. `env.mfa` is surfaced as an ABAC attribute so policies can also condition on it. The httpx gate maps `StepUpRequired` to `401` + `WWW-Authenticate: … step_up="mfa"` (re-auth) rather than a flat 403. |

Deny-by-default preserved: step-up only *gates* an existing allow — it never grants, and a plain deny is
never masked as a step-up (tested). Strong factors = `mfa/otp/totp/hwk/sms/fpt/face`. MFA itself stays
the IdP's job; the framework gates on the surfaced `amr`.

Tests (`kernel/authz/step_up_test.go`): a granted actor without a strong factor gets a step-up challenge
and is admitted once AMR includes `mfa`; an ungranted actor on a StepUp perm gets a plain `default_deny`
(no step-up). All pre-existing authz + gate tests still green with the additions. Gate: 84 packages.

## R1 — Authz decision caching (D-0074)

| Verdict | Fix |
|---|---|
| real (P1) — every `Evaluate` hit the DB for the actor's assignments | `authz.CachingStore` — an **opt-in** `Store` decorator caching the hot read (`ActiveAssignments`) per `(tenant, actor)` for a short TTL (default 1s). Unwrapped, the evaluator behaves exactly as before (zero risk to existing deployments). `Invalidate(tenant, actor)` / `InvalidateTenant(tenant)` make a role revoke take effect immediately on the pod; the TTL is only the cross-pod bound. Other Store reads pass through. |

Correctness (the R1 requirement "no stale-allow after revocation"): proven by test — a revoke served
bounded-stale within the TTL, then **immediately denied after `Invalidate`** (no stale-allow), and a
cache hit serves 2 reads with 1 DB call. Read-replica routing (R1's second half) is a deployment seam:
point the Manager's `WithTenantRO` path at a replica pool — the evaluator already runs its reads in that
read-only transaction (documented on `CachingStore`).

Test (`kernel/authz/caching_internal_test.go`): 2 reads → 1 DB call; revoke bounded-stale within TTL;
`Invalidate` → immediate reload/deny; TTL expiry → reload. Gate: 84 packages.

## O1 — Distributed tracing seam (D-0075)

| Verdict | Fix |
|---|---|
| real (P1) — request-id propagation only; no tracing | `kernel/observability.Tracer`/`Span` port + `NoOpTracer` (a sibling of the `Metrics` port) + a `Trace` HTTP middleware that opens a server span per request (route/method/status/request-id attrs). Wired into the generated api chain with `NoOpTracer` — **zero-cost when disabled**. |

Consistent with the framework's port/adapter split (metrics ships the port in the kernel, prometheus in
`adapters/metrics/`): the kernel owns the tracing port; the OpenTelemetry binding is a thin adapter.
**The post-hardening review (D-0077, F3) delivered the end-to-end pieces:** the real
`adapters/tracing/otel` adapter (configurable ratio sampler, `NewOTLP` OTLP exporter), the `Tracer` port's
`Inject`/`Extract` for W3C-traceparent cross-process propagation (the HTTP middleware continues an inbound
trace), and the tracing infra — a Jaeger service in the compose stack + a deployment-checklist section.
Carrying the traceparent through outbox events / job payloads (the relay→worker leg) remains the last
wiring step.

Tests (`kernel/observability/tracing_test.go`): the middleware starts exactly one span per request, ends
it, and tags http.route/method/status; `NoOpTracer` returns ctx unchanged and swallows all calls. Gate:
84 packages; the generated api template still parses with the `Trace` line.
</content>
