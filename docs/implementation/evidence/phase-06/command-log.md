# Phase 6 — Command Log

Exact commands with exit status; summarized output where long. Date: 2026-07-03.

| # | Command | Exit | Output summary |
|---|---|---|---|
| 1 | `make migrate` (00007 on fresh schema) | 0 | 7 migrations applied: events_outbox, processed_events, jobs_queue, job_runs + RLS/grants |
| 2 | `DATABASE_URL=… go test -run Integration ./kernel/outbox/` | 0 | atomic-with-business-tx (event iff commit), relay dispatch, inbox dedup (handler once), tenant isolation |
| 3 | `DATABASE_URL=… go test -run Integration ./kernel/jobs/` (agent) | 0 | atomic enqueue, worker success/tenant-aware, retry→DLQ, backoff reschedule, ReclaimStalled, tenant isolation |
| 4 | `go test ./kernel/jobs/` (units) | 0 | DefaultRetry, ExpJitterBackoff deterministic/capped, Registry dup-kind, Job.Kind |
| 5 | `DATABASE_URL=… go test -run TestIntegrationWorkerEndToEnd ./testkit/` | 0 | StartWorker: emitted event dispatched + enqueued job executed; graceful shutdown on ctx cancel |
| 6 | `unset DATABASE_URL; make ci` | 0 | vet, boundary lint, unit, race, build green |
| 7 | `make test-integration` (all packages) | 0 | outbox/jobs/authz/relationship/resource/testkit integration green |
| 8 | `docker compose run --rm tools ... go test -run Integration ./kernel/outbox/ ./kernel/jobs/` | 0 | outbox + jobs integration green inside the tools container (container-first) |
