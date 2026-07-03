# Phase 6 — Review Findings

One reliability+security critique agent reviewed the outbox/relay/jobs/worker slice (2026-07-03)
with live probes. It verified atomicity and tenant isolation are production-grade (SEC-35 positive)
and reproduced four dispatch/retry/shutdown correctness gaps against the blueprint's own §3/§7.

| ID | Sev | Finding | Resolution | Status |
|---|---|---|---|---|
| ARCH-53 | high | per-aggregate ordering NOT guaranteed — the §3 advisory lock was absent; a transient handler failure reordered events, and concurrent relays reorder — reproduced | claim picks only the earliest undispatched event per (tenant,resource); tx-scoped `pg_advisory_xact_lock` per aggregate serializes concurrent relays; `TestIntegrationOutboxPerAggregateOrderUnderRetry` (D-0050) | **fixed** |
| ARCH-54 | med | no poison ceiling / event DLQ — failed events retried forever | `events_outbox` gains max_attempts + `'dead'` status; poison events dead-letter; `TestIntegrationOutboxDLQ` | **fixed** |
| ARCH-55 | med | RequeueFailed cooldown keyed on occurred_at (write time) → no cooldown — reproduced | added `failed_at`; cooldown keyed on it | **fixed** |
| ARCH-56 | med | drainTimeout overloaded as per-job timeout AND outcome-write ctx (jobs cut off early; recordFailure on an expired ctx) — reproduced | separate per-job `jobTimeout` (default 2m); outcomes written with a fresh short-lived ctx | **fixed** |
| ARCH-57 | med | graceful-shutdown drain was advisory — a ctx-ignoring worker hangs shutdown forever | `StartWorker` races drain against a HARD `ShutdownDrain` cap and returns | **fixed** |
| ARCH-58 | med | reclaim visibility-timeout race → a live over-running job reclaimed and run concurrently | `stalledTimeout` floored above `jobTimeout + drainTimeout + 1m` in NewRunner so a live job is never reclaimable | **fixed** |
| ARCH-59 | med | jobs at-least-once with NO framework dedup (unlike events) — under-documented | documented loudly on `jobs.Worker`: workers must be idempotent; external side effects need their own idempotency key. Optional job inbox is a future enhancement | **documented** |
| info | claim tx holds two pooled conns across the batch; created_by hardcoded zero; onDead at-most-once; no replay for late-added handlers | accepted/noted — claim-then-close and a durable audit DLQ hook are Phase 11 hardening; created_by is a jsonb-actor design artifact | **accepted (noted)** |

Reviewer-verified solid (positive): outbox atomicity (event iff business commit); app_rt cannot
read cross-tenant outbox (relay reads as app_platform via the role-scoped policy); no tenant mix-up
in dispatch (per-event WithTenantID); an event is never marked dispatched without a durable handler
effect (crash before the mark → re-dispatch → inbox dedups); job atomic enqueue; exact max_attempts
(no off-by-one); backoff on the DB clock (skew-immune); bounded worker pool (no unbounded
goroutines); `-race` clean.

Residual risk (honest):
- Per-aggregate ordering is now enforced for the shipped single-relay path AND concurrent relays
  (advisory lock). Aggregate-less events (no resource) remain unordered by design.
- Jobs stay at-least-once; the framework provides no job inbox — workers with external effects must
  self-dedup (documented). An optional job idempotency key is a candidate enhancement.
- Claim-tx connection hold and a durable event-DLQ admin/requeue API are Phase 11 items.
