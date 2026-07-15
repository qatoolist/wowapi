---
id: W07-E01-S003-T008
type: task
title: Publication against perf/reference-v1.json
status: complete
parent_story: W07-E01-S003
owner: W07-Scoping-Dispatch.W07E01S003
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S003-T006
  - W07-E01-S003-T007
acceptance_criteria:
  - AC-W07-E01-S003-07
artifacts:
  - ART-W07-E01-S003-008
evidence:
  - EV-W07-E01-S003-008
---

# W07-E01-S003-T008 — Publication against perf/reference-v1.json

## Task Definition

### Task objective

Publish before/after evidence against perf/reference-v1.json.

### Parent story

W07-E01-S003

### Owner

W07-Scoping-Dispatch.W07E01S003

### Status

complete

### Dependencies

W07-E01-S003-T006, W07-E01-S003-T007 (publication consumes prior evidence); cross-story on W07-E01-S001's T001.

### Detailed work

1. Confirm W07-E01-S001's perf/reference-v1.json exists.
2. Compile before/after evidence from T001-T007.
3. Publish the comparison.

### Expected files or components affected

A published comparison report.

### Expected output

Before/after evidence published against perf/reference-v1.json.

### Required artifacts

ART-W07-E01-S003-008 (published comparison).

### Required evidence

EV-W07-E01-S003-008 (the report itself).

### Related acceptance criteria

AC-W07-E01-S003-07.

### Completion criteria

The comparison is published and references perf/reference-v1.json.

### Verification method

Direct inspection of the published report.

### Risks

Blocked on the reference environment (W07-E01-S001), per PLAN T8's own dependency.

### Rollback or recovery considerations

If W07-E01-S001's environment is not yet available, wait rather than publishing against an ad hoc substitute baseline.

## Implementation Record

### What was actually implemented

Published raw same-host pre/post outputs and a machine-readable median comparison. The report points
to `perf/reference-v1.json`, identifies its provisional-advisory DEC-Q9 policy, records an environment
mismatch, and marks all absolute SLO results not assessed.

### Components changed

Performance results and story evidence.

### Files changed

`perf/results/perf-04-sweeper-before.txt`, `perf/results/perf-04-sweeper-after.txt`,
`perf/results/perf-04-comparison-v1.json`, `perf/results/perf-04-reminder-explain.txt`.

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

Local database DSN is redacted from published outputs.

### Observability changes

None.

### Tests added or modified

Artifact checksums and evidence records pin each publication.

### Commits

Working tree based on entry SHA `733ef3e`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

Absolute reference-environment SLOs await DEC-Q9; this publication is same-host relative only.

### Follow-up items

Re-run using the accepted reference environment after DEC-Q9.

### Relationship to the approved plan

Matches plan T8 and does not overstate conditional evidence.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S003-07 | Inspect raw outputs, medians, hashes, reference pointer and qualification fields | Same-host before/after + accepted reference policy | Truthful relative deltas; absolute result conditional | publication report | W07-Scoping-Dispatch.W07E01S003ReviewR |

### Actual result

PASS: median runtime improved 34.93%/86.75%/95.45%, bytes improved 44.44%/89.10%/99.89%,
and allocations improved 38.84%/94.57%/99.94% at 10/1k/100k. Absolute result is explicitly
`not-assessed-pending-DEC-Q9`.

### Pass or fail

PASS for truthful relative publication.

### Evidence identifier

EV-W07-E01-S003-008.

### Execution date

2026-07-13.

### Commit or revision

Working tree based on entry SHA `733ef3e`.

### Environment

Same-host darwin/arm64 Apple M3 Max, Go 1.26.5, PostgreSQL 16.9.

### Reviewer

W07-Scoping-Dispatch.W07E01S003ReviewR.

### Findings

No absolute SLO conclusion is permitted while DEC-Q9 remains open.

### Retest status

Checksums and source reference verified.

### Final conclusion

Publication is complete and truthful; absolute reference acceptance remains conditional.
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
