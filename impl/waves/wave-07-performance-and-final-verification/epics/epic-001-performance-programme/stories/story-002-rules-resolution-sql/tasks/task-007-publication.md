---
id: W07-E01-S002-T007
type: task
title: Publication against perf/reference-v1.json
status: done
parent_story: W07-E01-S002
owner: W07-Scoping-Dispatch.W07E01S002
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S002-T004
  - W07-E01-S002-T005
  - W07-E01-S002-T006
acceptance_criteria:
  - AC-W07-E01-S002-06
artifacts:
  - ART-W07-E01-S002-006
evidence:
  - EV-W07-E01-S002-007
---

# W07-E01-S002-T007 — Publication against perf/reference-v1.json

## Task Definition

### Task objective

Publish before/after evidence against perf/reference-v1.json.

### Parent story

W07-E01-S002

### Owner

unassigned

### Status

todo

### Dependencies

W07-E01-S002-T004, W07-E01-S002-T005, W07-E01-S002-T006 (publication consumes prior tasks' evidence); cross-story on W07-E01-S001's T001 (perf/reference-v1.json must exist).

### Detailed work

1. Confirm W07-E01-S001's perf/reference-v1.json exists.
2. Compile before/after evidence from T002-T006.
3. Publish the comparison.

### Expected files or components affected

A published comparison report.

### Expected output

Before/after evidence published against perf/reference-v1.json.

### Required artifacts

ART-W07-E01-S002-006 (published comparison).

### Required evidence

EV-W07-E01-S002-007 (the report itself).

### Related acceptance criteria

AC-W07-E01-S002-06.

### Completion criteria

The comparison is published and references perf/reference-v1.json.

### Verification method

Direct inspection of the published report.

### Risks

Blocked on the reference environment (W07-E01-S001), per PLAN T6's own dependency.

### Rollback or recovery considerations

If W07-E01-S001's environment is not yet available, wait rather than publishing against an ad hoc substitute baseline.

## Implementation Record

### What was actually implemented

Published `perf/results/perf-03-comparison.json` against the accepted
`perf/reference-v1.json` SHA-256. The report contains real same-container legacy/set-based SQL counts,
all four EXPLAIN observations, the reference's zero statement-count-increase ceiling, and an explicit
claim boundary.

### Measured result

- Legacy total SQL statements at depths 3/10/50: 11/18/58.
- Set-based total SQL statements at depths 3/10/50: 8/8/8.
- Relative statement reductions: 27.27%/55.56%/86.21%.
- No absolute latency, throughput, or SLO claim is made.

### Files changed

- `perf/results/perf-03-comparison.json`

### Implementation dates

2026-07-14.

### Known limitations

DEC-Q9 remains open. Local Darwin/arm64 container timings are observations, not accepted reference-runner
measurements and not absolute SLO evidence.

### Relationship to the approved plan

Matched T6 and the wave's DEC-Q9 conditionality.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Actual result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S002-06 | JSON parse, reference hash, relative-count, and DEC-Q9 contract assertions | Repository + generated data | PASS — report is valid and honestly conditional | EV-W07-E01-S002-007 | pending story independent review |

### Final conclusion

Passed on 2026-07-14. Publication references `perf/reference-v1.json` by path and SHA-256.

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
