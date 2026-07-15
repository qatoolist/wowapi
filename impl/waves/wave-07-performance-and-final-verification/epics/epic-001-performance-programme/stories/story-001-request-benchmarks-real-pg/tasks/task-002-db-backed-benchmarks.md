---
id: W07-E01-S001-T002
type: task
title: DB-backed benchmarks, all 6 profiles
status: complete
parent_story: W07-E01-S001
owner: W07-Phase-A-Execution.W07E01S001
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S001-T001
acceptance_criteria:
  - AC-W07-E01-S001-02
artifacts:
  - ART-W07-E01-S001-002
evidence:
  - EV-W07-E01-S001-002
---

# W07-E01-S001-T002 — DB-backed benchmarks, all 6 profiles

## Task Definition

### Task objective

Build DB-backed benchmarks for public/authenticated-read/authenticated-write/resource-authz/idempotent-write/async-enqueue profiles against real Postgres.

### Parent story

W07-E01-S001

### Owner

W07-Phase-A-Execution.W07E01S001

### Status

complete

### Dependencies

W07-E01-S001-T001 (benchmarks run on the reference environment T001 stands up).

### Detailed work

1. Build a benchmark for each of the 6 named workload profiles, against real Postgres, not fakes.
2. Record p50/p95/p99, allocations, SQL count, bytes, pool wait, tx duration, lock wait, and plan hash
   for each.
3. Confirm no RLS guard is weakened to achieve any result — the explicit directive prohibition.

### Expected files or components affected

New DB-backed benchmark files (exact location TBD).

### Expected output

A benchmark suite covering all 6 profiles against real Postgres.

### Required artifacts

ART-W07-E01-S001-002 (DB-backed benchmark suite).

### Required evidence

EV-W07-E01-S001-002 (benchmark run report, real Postgres, all 6 profiles).

### Related acceptance criteria

AC-W07-E01-S001-02.

### Completion criteria

All 6 profiles are benchmarked against real Postgres, no RLS guard weakened.

### Verification method

Direct execution of the benchmark suite against real Postgres.

### Risks

Medium — must not weaken RLS guards to win the benchmark, per PLAN T2's own explicit directive prohibition.

### Rollback or recovery considerations

If a benchmark result requires an RLS weakening to hit a target, treat this as a real finding about RLS performance, not license to weaken it — escalate rather than silently bypassing.

## Implementation Record

### What was actually implemented

Added six real-PostgreSQL request profiles through the production HTTP/auth/authz/tenant-transaction path with RLS enforced.

### Components and files changed

`perf/requestbench/requests_bench_test.go`, `perf/requestbench/requests_contract_test.go`

### Interfaces, configuration, schema, and security

Benchmark-only additive interfaces/configuration. No schema migration or production API changed; runtime RLS remains enforced.

### Tests, revision, and date

Focused contracts and real-PostgreSQL execution passed on the working tree based on `1626b11`; implemented 2026-07-13 through 2026-07-14. No commit or pull request was created.

### Relationship to the approved plan

Implemented as planned. Absolute SLOs remain outside scope pending DEC-Q9.

## Verification Record

### Actual result and pass/fail

PASS. See `EV-W07-E01-S001-002` and the story-level `verification.md` for the exact command, environment, result, and checksum-pinned output.

### Execution date and revision

2026-07-14; working tree based on entry SHA `1626b11`.

### Environment

Exact required local PostgreSQL environment for focused contracts; pinned Linux/amd64 Go 1.26.5 + PostgreSQL 16.9 containers for publication.

### Reviewer and findings

Independent review by `W05ReviewGateFinal` passed with zero open actionable issues.

### Retest and conclusion

Retested after fixes; task completion criterion satisfied.

## Deviations Record

*No deviations recorded yet.*

### Deviation ID

*Not applicable.*

### Approved plan

*Not applicable.*

### Actual implementation

*Not applicable.*

### Reason

*Not applicable.*

### Impact

*Not applicable.*

### Risks

*Not applicable.*

### Approval

*Not applicable.*

### Compensating controls

*Not applicable.*

### Follow-up work

*Not applicable.*
