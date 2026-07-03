# Phase 6 — Acceptance Map

Phase 6 exit criteria (Goal 2 Phase 6 + phase-plan row 6 + blueprint 07 §3/§7) → proof.

| # | Criterion | Proof |
|---|---|---|
| 1 | Transactional outbox (Writer) | `kernel/outbox/outbox.go`; `TestIntegrationOutboxAtomicWithBusinessTx` (event iff commit) |
| 2 | **Events commit atomically with business writes** | same test: rollback loses the event, commit persists it |
| 3 | Dispatcher + inbox (idempotent handlers) | `kernel/outbox/relay.go`; `TestIntegrationOutboxRelayDispatchAndInbox` (dispatch + dedup on redelivery) |
| 4 | Per-aggregate ordering (blueprint §7) | earliest-per-aggregate claim + advisory lock; `TestIntegrationOutboxPerAggregateOrderUnderRetry` |
| 5 | Event DLQ (poison ceiling) | `'dead'` status + max_attempts; `TestIntegrationOutboxDLQ` |
| 6 | Job runner (Postgres, D-0047) | `kernel/jobs/`; units + integration |
| 7 | Atomic enqueue in business tx | `TestIntegrationJobsEnqueueAtomic` (rollback → no job) |
| 8 | **Crash/retry tests pass** | `TestIntegrationJobsRetryToDLQ` (exact max_attempts → dead), `BackoffReschedules`, `ReclaimStalled`; outbox inbox dedup |
| 9 | **Jobs are tenant-aware** | `TestIntegrationJobsWorkerSucceeds`, `TestIntegrationJobsTenantIsolation` (worker sees only its tenant's rows) |
| 10 | Retry/backoff/DLQ correctness | DB-clock backoff (skew-immune), `job_runs` dead row, DLQ hook |
| 11 | Worker process + **graceful shutdown** | `app/worker.go` StartWorker (relay + runner, hard drain cap); `TestIntegrationWorkerEndToEnd` |
| 12 | Tenant isolation of events | `TestIntegrationOutboxTenantIsolation` (tenant B sees 0 of A's events via runtime RLS); relay reads cross-tenant only as app_platform |
| 13 | Bounded goroutines | runner pool is a semaphore = poolSize; `-race` clean |
| 14 | Migrations idempotent | `make migrate` ×2 (0 on rerun); migration 00007 markers/ordering |
| 15 | Container-first verification | host `make ci` + `make test-integration`; `docker compose run tools ... ./kernel/outbox ./kernel/jobs` green |
| 16 | Evidence bundle + review | this directory; review-findings.md (1 reproduced high ordering bug + 5 mediums fixed) |

Carried forward: optional job idempotency inbox (ARCH-59), claim-tx connection hold + durable
event-DLQ admin/requeue API (Phase 11 hardening). Module.Context Events/Jobs/Outbox accessors are
wired (D-0049); later-phase accessors (Rules/Workflows → 7, Documents → 8, Notify/Webhooks → 9)
arrive with their packages. Graphify `extract` blocked on LLM key (R11).
