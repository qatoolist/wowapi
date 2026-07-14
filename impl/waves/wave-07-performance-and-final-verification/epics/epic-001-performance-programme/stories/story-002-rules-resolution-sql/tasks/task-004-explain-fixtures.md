---
id: W07-E01-S002-T004
type: task
title: EXPLAIN fixtures at representative cardinality
status: done
parent_story: W07-E01-S002
owner: W07-Scoping-Dispatch.W07E01S002
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S002-T003
acceptance_criteria:
  - AC-W07-E01-S002-04
artifacts:
  - ART-W07-E01-S002-004
evidence:
  - EV-W07-E01-S002-004
---

# W07-E01-S002-T004 — EXPLAIN fixtures at representative cardinality

## Task Definition

### Task objective

Produce EXPLAIN (ANALYZE, BUFFERS) fixtures at representative depth/history cardinality.

### Parent story

W07-E01-S002

### Owner

unassigned

### Status

todo

### Dependencies

W07-E01-S002-T003 (fixtures are produced against the confirmed/added indexes).

### Detailed work

1. Build a fixture harness with seeded org-ancestry test data.
2. Produce EXPLAIN fixtures for shallow and deep org ancestries.
3. Produce EXPLAIN fixtures for low and high historical-version counts.
4. Commit all 4 fixture combinations.

### Expected files or components affected

New EXPLAIN fixture files, one per cardinality combination.

### Expected output

Committed EXPLAIN fixtures for all 4 named cardinality combinations.

### Required artifacts

ART-W07-E01-S002-004 (EXPLAIN fixture files).

### Required evidence

EV-W07-E01-S002-004 (fixture inventory + output).

### Related acceptance criteria

AC-W07-E01-S002-04.

### Completion criteria

All 4 cardinality combinations have committed EXPLAIN fixtures.

### Verification method

Direct inspection of committed fixture files.

### Risks

Medium — needs seeded org-ancestry test data, per PLAN T3's own risk note.

### Rollback or recovery considerations

If seed data proves difficult to construct realistically, escalate rather than silently using an unrepresentative dataset.

## Implementation Record

### What was actually implemented

Added a real-PostgreSQL fixture harness that seeds a 20,000-row background, real organization chains,
and both current and superseded rule histories. It emits four
`EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)` fixtures, each containing the live historical resolution
plan and a current-predicate plan.

### Files changed

- `kernel/rules/resolver_perf_test.go`
- `perf/results/perf-03-explain-shallow-low.json`
- `perf/results/perf-03-explain-shallow-high.json`
- `perf/results/perf-03-explain-deep-low.json`
- `perf/results/perf-03-explain-deep-high.json`

### Cardinalities

Shallow/deep are ancestry depths 3/50. Low/high are 4/1000 versions per scope.

### Implementation dates

2026-07-14.

### Relationship to the approved plan

Matched T3 with all four required combinations.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Actual result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S002-04 | Generate, parse, and inspect committed fixtures | PostgreSQL 16.14 container | PASS — 4/4 fixtures generated and JSON-valid | EV-W07-E01-S002-004 | pending story independent review |

### Final conclusion

Passed on 2026-07-14; no fixture case skipped.

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
