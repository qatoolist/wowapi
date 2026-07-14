---
id: IMPL-W07-E01-S003
type: implementation-record
parent_story: W07-E01-S003
status: implemented
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Implementation record — W07-E01-S003

This record aggregates the verified implementation reality across T001-T008.

## What was actually implemented

- Fixed-size, atomic 100-row SLA reminder/escalation claims with batched instance/definition loads.
- Safe scheduler reinvocation, no-double-remind concurrency, and bounded real-Postgres benchmarks.
- DATA-09 online partial reminder index and W04-compatible outbox lease columns.
- One-query RetryOutbound endpoint loading.
- Leased outbox claim/commit/tenant-handler/fenced-finalize state machine directly using `lease.Lease`.
- Queue-lag gauges and batch-duration histograms with bounded labels.
- Raw same-host before/after output, same-change budgets, and a DEC-Q9-qualified comparison.

## Components changed

Workflow sweeper/runtime, webhook retry, outbox relay, observability metrics port, kernel/worker wiring,
migration inventory, focused tests, benchmark budgets, and performance publications.

## Files changed

Production: `kernel/workflow/{runtime,sweeper}.go`, `foundation/webhook/service.go`,
`kernel/outbox/relay.go`, `kernel/observability/metrics.go`, `kernel/kernel.go`,
`app/{maintenance,worker}.go`.

Schema/tests/results: `migrations/00047_perf04_sweeper_outbox_leases.sql`,
`migrations/migrations_test.go`, `testkit/db.go`, `kernel/workflow/sweeper_perf_test.go`,
`foundation/webhook/retry_perf_test.go`, `kernel/outbox/relay_lease_test.go`,
`bench-budgets.txt`, and `perf/results/perf-04-*`.

## Interfaces introduced or changed

`observability.Metrics` adds histogram observation; workflow and outbox add metrics options;
outbox also adds a lease-TTL option for deterministic testing. No domain/public API changed.

## Configuration changes

Three `BenchmarkSweepSLABatch` ceilings were added to `bench-budgets.txt`.

## Schema or migration changes

Online migration 00047 creates `wft_remind_after`, adds outbox lease token/generation/expiry fields,
and creates the pending-claim index; the down section removes all additions.

## Security changes

Tenant RLS paths remain intact. Outbox claim/finalization uses app-platform access, while every
handler runs after claim commit inside its own tenant-bound transaction. Metrics expose no
tenant/resource/endpoint identifiers.

## Observability changes

Adds `worker_queue_lag_seconds` and `worker_batch_duration_seconds` for exactly `workflow_sla`,
`webhook_retry`, and `outbox_relay`.

## Tests added or modified

Real-Postgres query-count/cardinality/reinvocation/concurrency/EXPLAIN tests; RetryOutbound query
tracing; outbox commit-boundary/lease-expiry fencing/race tests; metrics tests; three-tier benchmark
and fail-closed budget gate; migration manifest update.

## Commits

Working tree based on entry SHA `733ef3e`.

## Pull requests

None.

## Implementation dates

2026-07-13 through 2026-07-14.

## Technical debt introduced

None.

## Known limitations

External outbox side effects retain the accepted at-least-once W04 semantics. Absolute performance
SLOs remain conditional on DEC-Q9; published measurements are same-host relative evidence only.

## Follow-up items

Re-run absolute measurements in the final accepted DEC-Q9 reference environment.

## Relationship to the approved plan

T1-T8 were implemented as planned. No story deviation was required.
