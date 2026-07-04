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
**Per-user channel preferences** (opt-out per channel) remains a follow-up — it needs a preferences
table + a send-path check, out of scope for this small increment.

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
</content>
