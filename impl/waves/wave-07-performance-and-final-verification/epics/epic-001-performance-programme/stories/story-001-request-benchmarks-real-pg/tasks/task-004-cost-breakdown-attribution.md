---
id: W07-E01-S001-T004
type: task
title: Cost-breakdown attribution
status: complete
parent_story: W07-E01-S001
owner: W07-Phase-A-Execution.W07E01S001
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S001-T002
acceptance_criteria:
  - AC-W07-E01-S001-04
artifacts:
  - ART-W07-E01-S001-004
evidence:
  - EV-W07-E01-S001-004
---

# W07-E01-S001-T004 — Cost-breakdown attribution

## Task Definition

### Task objective

Attribute cost by pool wait / tx setup / authz query / handler query / serialization / middleware, separately, no single aggregate number.

### Parent story

W07-E01-S001

### Owner

W07-Phase-A-Execution.W07E01S001

### Status

complete

### Dependencies

W07-E01-S001-T002 (attribution instruments T002's own benchmarks).

### Detailed work

1. Build cost-breakdown instrumentation, reusing existing tracing/span infrastructure where
   practical.
2. Confirm cost is attributed separately by each of the 6 named components, not reported as a single
   aggregate.

### Expected files or components affected

New cost-breakdown instrumentation (exact location TBD).

### Expected output

Per-component cost attribution for the benchmark suite.

### Required artifacts

ART-W07-E01-S001-004 (cost-breakdown instrumentation).

### Required evidence

EV-W07-E01-S001-004 (attribution report).

### Related acceptance criteria

AC-W07-E01-S001-04.

### Completion criteria

Cost is separately attributed per component.

### Verification method

Direct inspection of the attribution output.

### Risks

Medium — no existing scaffolding for this breakdown today, per PLAN T4's own risk note.

### Rollback or recovery considerations

If span-based attribution proves insufficiently granular, fall back to EXPLAIN-correlated instrumentation and record the choice.

## Implementation Record

### What was actually implemented

Added query-span SQL counting, pool/transaction/lock timing, six representative plan hashes, and six non-overlapping cost components.

### Components and files changed

`perf/requestbench/requests_bench_test.go`, `perf/results/request-reference-v1.json`

### Interfaces, configuration, schema, and security

Benchmark-only additive interfaces/configuration. No schema migration or production API changed; runtime RLS remains enforced.

### Tests, revision, and date

Focused contracts and real-PostgreSQL execution passed on the working tree based on `1626b11`; implemented 2026-07-13 through 2026-07-14. No commit or pull request was created.

### Relationship to the approved plan

Implemented as planned. Absolute SLOs remain outside scope pending DEC-Q9.

## Verification Record

### Actual result and pass/fail

PASS. See `EV-W07-E01-S001-004` and the story-level `verification.md` for the exact command, environment, result, and checksum-pinned output.

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
