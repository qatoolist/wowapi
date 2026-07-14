---
id: W07-E01-S001-T003
type: task
title: Concurrency-matrix variants
status: complete
parent_story: W07-E01-S001
owner: W07-Phase-A-Execution.W07E01S001
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S001-T002
acceptance_criteria:
  - AC-W07-E01-S001-03
artifacts:
  - ART-W07-E01-S001-003
evidence:
  - EV-W07-E01-S001-003
---

# W07-E01-S001-T003 — Concurrency-matrix variants

## Task Definition

### Task objective

Build cold/warm cache × 1/10/100 concurrent-tenant variants, minimum 6 combinations per workload profile.

### Parent story

W07-E01-S001

### Owner

W07-Phase-A-Execution.W07E01S001

### Status

complete

### Dependencies

W07-E01-S001-T002 (the matrix operates on T002's own benchmark suite).

### Detailed work

1. Build the cold/warm cache × 1/10/100-concurrent-tenant variant matrix.
2. Ensure minimum 6 combinations per workload profile.
3. Build realistic seed data for the 100-tenant case.

### Expected files or components affected

A concurrency-matrix benchmark harness, extending T002's own benchmarks.

### Expected output

All 6 profiles benchmarked across the full concurrency matrix.

### Required artifacts

ART-W07-E01-S001-003 (concurrency-matrix harness).

### Required evidence

EV-W07-E01-S001-003 (concurrency-matrix run report).

### Related acceptance criteria

AC-W07-E01-S001-03.

### Completion criteria

Minimum 6 combinations per profile are benchmarked.

### Verification method

Direct execution of the concurrency-matrix harness.

### Risks

Medium — 100-tenant needs realistic seed data, per PLAN T3's own risk note.

### Rollback or recovery considerations

If realistic 100-tenant seed data proves difficult to construct, escalate rather than silently using an unrepresentative synthetic dataset without recording the limitation.

## Implementation Record

### What was actually implemented

Added the complete cold/warm by 1/10/100 tenant matrix over a deterministic 100-tenant dataset.

### Components and files changed

`perf/requestbench/requests_bench_test.go`, `perf/fixtures/request-workloads-v1.json`

### Interfaces, configuration, schema, and security

Benchmark-only additive interfaces/configuration. No schema migration or production API changed; runtime RLS remains enforced.

### Tests, revision, and date

Focused contracts and real-PostgreSQL execution passed on the working tree based on `1626b11`; implemented 2026-07-13 through 2026-07-14. No commit or pull request was created.

### Relationship to the approved plan

Implemented as planned. Absolute SLOs remain outside scope pending DEC-Q9.

## Verification Record

### Actual result and pass/fail

PASS. See `EV-W07-E01-S001-003` and the story-level `verification.md` for the exact command, environment, result, and checksum-pinned output.

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
