---
id: W07-E01-S003-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W07-E01-S003
status: verified
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W07-E01-S003 — Artifacts index

Per mandate §9.2, source and test artifacts stay in their production locations; measured outputs
are published under `perf/results/`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W07-E01-S003-001 | Bounded-batch SweepSLA code | source-code | implementation | Fixed 100-row guarded claims, batched materialization, safe job re-invocation | PERF-04 | W07-E01-S003-T001 | `kernel/workflow/sweeper.go`; `app/maintenance.go` | produced |
| ART-W07-E01-S003-002 | Set-based/batched operation conversions | source-code | implementation | Atomic set guard flips plus instance/definition loads by ID set | PERF-04 | W07-E01-S003-T002 | `kernel/workflow/sweeper.go`; `kernel/workflow/sweeper_perf_test.go` | produced |
| ART-W07-E01-S003-003 | `remind_after` partial index migration | schema migration | implementation | DATA-09-compliant concurrent partial index plus outbox lease fields | PERF-04 | W07-E01-S003-T003 | `migrations/00047_perf04_sweeper_outbox_leases.sql` | produced |
| ART-W07-E01-S003-004 | Batch-loaded RetryOutbound code | source-code | implementation | One `ANY(uuid[])` endpoint query per invocation | PERF-04 | W07-E01-S003-T004 | `foundation/webhook/service.go`; `foundation/webhook/retry_perf_test.go` | produced |
| ART-W07-E01-S003-005 | Leased-state-machine outbox rework | source-code | implementation | W04 lease primitive, committed claim, fenced finalize, ordering preserved | PERF-04 | W07-E01-S003-T005 | `kernel/outbox/relay.go`; `kernel/outbox/relay_lease_test.go` | produced |
| ART-W07-E01-S003-006 | Queue-lag/batch-duration metric emission | source-code | implementation | Bounded worker timing labels for sweeper, webhook retry, outbox relay | PERF-04 | W07-E01-S003-T006 | `kernel/workflow/runtime.go`; `foundation/webhook/service.go`; `kernel/outbox/relay.go`; `kernel/kernel.go`; `app/worker.go` | produced |
| ART-W07-E01-S003-007 | Bounded-batch benchmarks + budget entries | test suite + configuration | implementation | 10/1k/100k real-Postgres benchmark and same-change ceilings | PERF-04 | W07-E01-S003-T007 | `kernel/workflow/sweeper_perf_test.go`; `bench-budgets.txt`; `perf/results/perf-04-sweeper-after.txt` | produced |
| ART-W07-E01-S003-008 | Published before/after comparison | documentation + data | post-implementation | Same-host relative comparison against reference policy; absolute conditional | PERF-04 | W07-E01-S003-T008 | `perf/results/perf-04-comparison-v1.json`; `perf/results/perf-04-sweeper-before.txt`; `perf/results/perf-04-reminder-explain.txt` | produced |
