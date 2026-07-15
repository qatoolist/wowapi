---
id: VER-W07-E01-S003
type: verification-record
parent_story: W07-E01-S003
status: verified
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Verification record — W07-E01-S003

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S003-01 | Run traced cardinality tests and three-tier real-DB benchmark | PostgreSQL 16.9 | Fixed 100-row claim and bounded allocations at 10/1k/100k | EV-W07-E01-S003-001 | W05ReviewGateFinal |
| AC-W07-E01-S003-02 | Run traced reinvocation and concurrent sweeper tests | PostgreSQL 16.9 | Fixed guard/load counts and no double reminder | EV-W07-E01-S003-002 | W05ReviewGateFinal |
| AC-W07-E01-S003-03 | Run real `EXPLAIN` without planner override | PostgreSQL 16.9 | `Index Scan using wft_remind_after` | EV-W07-E01-S003-003 | W05ReviewGateFinal |
| AC-W07-E01-S003-04 | Trace 10 deliveries across 3 endpoints | PostgreSQL 16.9 | One `ANY(uuid[])` endpoint query | EV-W07-E01-S003-004 | W05ReviewGateFinal |
| AC-W07-E01-S003-05 | Run commit-boundary, lease/race, inherited chaos and ordering suites | PostgreSQL 16.9 + race detector | No outer claim transaction; stale worker fenced; order/idempotency pass | EV-W07-E01-S003-005 | W05ReviewGateFinal |
| AC-W07-E01-S003-06 | Run recording metric sinks and production-wiring tests | PostgreSQL 16.9 | Lag gauge and duration histogram with bounded labels | EV-W07-E01-S003-006 | W05ReviewGateFinal |
| AC-W07-E01-S003-07 | Run three-tier benchmark→budget gate and inspect publication | Same-host real Postgres; reference policy | Same-change budgets and truthful relative publication; absolute conditional on DEC-Q9 | EV-W07-E01-S003-007, -008 | W05ReviewGateFinal |

## Post-execution record

Focused execution and independent review are complete.

### Actual result

All focused implementation packages, inherited chaos, race stress, migration contracts, benchmark
budget gate, and publication checks passed as recorded in EV-W07-E01-S003-001 through -008.

### Pass or fail

PASS; absolute SLO assessment remains intentionally conditional on DEC-Q9.

### Evidence identifier

EV-W07-E01-S003-001 through EV-W07-E01-S003-008.

### Execution date

2026-07-13.

### Commit or revision

Working tree based on entry SHA `733ef3e`.

### Environment

darwin/arm64, Apple M3 Max, Go 1.26.5; real PostgreSQL 16.9 Docker service; race detector where cited.

### Reviewer

W07-Scoping-Dispatch.W07E01S003ReviewR — PASS, no open issues.

### Findings

Executor found and fixed one test-instrumentation timing issue: the reclaiming worker now receives a
five-second lease while the stale worker retains the 75ms expiry trigger; ten race runs pass.

### Retest status

Focused packages, ten race repetitions, and the benchmark budget gate passed.

### Final conclusion

All seven ACs are accepted with observed evidence and a clean independent gate.
