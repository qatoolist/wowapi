---
id: W07-E01-S001-T005
type: task
title: Publication against perf/reference-v1.json, DEC-Q9-conditional
status: complete
parent_story: W07-E01-S001
owner: W07-Phase-A-Execution.W07E01S001
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E01-S001-T001
  - W07-E01-S001-T002
  - W07-E01-S001-T003
  - W07-E01-S001-T004
acceptance_criteria:
  - AC-W07-E01-S001-05
artifacts:
  - ART-W07-E01-S001-005
evidence:
  - EV-W07-E01-S001-005
---

# W07-E01-S001-T005 — Publication against perf/reference-v1.json, DEC-Q9-conditional

## Task Definition

### Task objective

Publish results against perf/reference-v1.json as relative/container comparisons now, with absolute-SLO acceptance explicitly conditional on DEC-Q9.

### Parent story

W07-E01-S001

### Owner

W07-Phase-A-Execution.W07E01S001

### Status

complete

### Dependencies

W07-E01-S001-T001, W07-E01-S001-T002, W07-E01-S001-T003, W07-E01-S001-T004 (publication consumes all four prior tasks' output).

### Detailed work

1. Compile T002-T004's own benchmark and attribution results against T001's reference baseline.
2. Publish as a relative/container comparison report.
3. Write the report's own acceptance-criteria language to be explicitly conditional on DEC-Q9 for any
   absolute-latency claim — per this wave's own task-brief instruction, do not write an unconditional
   absolute-latency AC.

### Expected files or components affected

A published comparison report (exact format/location TBD).

### Expected output

A relative/container comparison report with explicit DEC-Q9-conditional absolute-SLO framing.

### Required artifacts

ART-W07-E01-S001-005 (published comparison report).

### Required evidence

EV-W07-E01-S001-005 (the report itself).

### Related acceptance criteria

AC-W07-E01-S001-05.

### Completion criteria

The report is published and its DEC-Q9 conditionality is explicit, not silently omitted.

### Verification method

Direct inspection of the published report's own language.

### Risks

Blocked until T001 exists — this is PERF-02's own actual closure gate, per PLAN T5's own risk note.

### Rollback or recovery considerations

If the report is found to have silently asserted an unconditional absolute-SLO claim, correct it immediately — this is exactly the overclaiming this story's whole framing exists to prevent.

## Implementation Record

### What was actually implemented

Published the pinned Linux/amd64 Go and PostgreSQL container result with relative ratios and explicit DEC-Q9-conditional absolute-SLO status.

### Components and files changed

`perf/results/request-reference-v1.json`, `perf/results/request-reference-v1.txt`

### Interfaces, configuration, schema, and security

Benchmark-only additive interfaces/configuration. No schema migration or production API changed; runtime RLS remains enforced.

### Tests, revision, and date

Focused contracts and real-PostgreSQL execution passed on the working tree based on `1626b11`; implemented 2026-07-13 through 2026-07-14. No commit or pull request was created.

### Relationship to the approved plan

Implemented as planned. Absolute SLOs remain outside scope pending DEC-Q9.

## Verification Record

### Actual result and pass/fail

PASS. See `EV-W07-E01-S001-005` and the story-level `verification.md` for the exact command, environment, result, and checksum-pinned output.

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
