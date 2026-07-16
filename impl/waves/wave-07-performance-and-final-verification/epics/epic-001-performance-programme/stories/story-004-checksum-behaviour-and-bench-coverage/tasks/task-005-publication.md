---
id: W07-E01-S004-T005
type: task
title: Publication against perf/reference-v1.json
status: done
parent_story: W07-E01-S004
owner: W07-Scoping-Dispatch.W07E01S004
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W07-E01-S004-T001
  - W07-E01-S004-T002
  - W07-E01-S004-T003
  - W07-E01-S004-T004
acceptance_criteria:
  - AC-W07-E01-S004-05
artifacts:
  - ART-W07-E01-S004-005
evidence:
  - EV-W07-E01-S004-005
---

# W07-E01-S004-T005 — Publication against perf/reference-v1.json

## Task Definition

### Task objective

Publish before/after evidence against perf/reference-v1.json.

### Parent story

W07-E01-S004

### Owner

W07-Scoping-Dispatch.W07E01S004

### Status

complete

### Dependencies

W07-E01-S004-T001 through W07-E01-S004-T004 (publication consumes prior evidence); cross-story on W07-E01-S001's T001.

### Detailed work

1. Confirm W07-E01-S001's perf/reference-v1.json exists (only needed for the quantified-latency
   claim; T001's own "no body download" proof stands independently).
2. Compile before/after evidence.
3. Publish the comparison, noting which claims are independently proven now vs. which depend on the
   reference environment.

### Expected files or components affected

A published comparison report.

### Expected output

Before/after evidence published, with independent-vs-reference-env claims distinguished.

### Required artifacts

ART-W07-E01-S004-005 (published comparison).

### Required evidence

EV-W07-E01-S004-005 (the report itself).

### Related acceptance criteria

AC-W07-E01-S004-05.

### Completion criteria

The comparison is published; AC-01's behavioral proof is noted as standing independently.

### Verification method

Direct inspection of the published report.

### Risks

Partially blocked — the 'no body download' behavioral proof (T1) is independently testable now; only the quantified latency claim needs the reference environment, per PLAN T5's own risk note.

### Rollback or recovery considerations

If the reference environment is not yet available, publish the independently-provable claims now and note the quantified-latency claim as pending, rather than delaying the entire publication.

## Implementation Record

Published `perf/results/perf-05-comparison-v1.json` and
`perf/results/perf-05-checksum-inventory-v1.json`. The comparison records real
local measurements and explicitly classifies them as not like-for-like with the
accepted Linux/amd64 reference because the seven CS-16 benchmarks are new and
have no before observations. It makes no absolute SLO claim and keeps DEC-Q9
open. Implemented 2026-07-14, working tree based on `733ef3e`; no PR, debt, or
plan deviation.
## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S004-05 | inspect comparison against `perf/reference-v1.json` | local measured run plus accepted reference | truthful publication; DEC-Q9 conditional | data report | independent story reviewer |

**PASS**, 2026-07-14, working tree based on `733ef3e`.
EV-W07-E01-S004-005 publishes measured values, comparability limitations, and
the required reference-environment follow-up. Independent review: correct,
confidence 1, no findings.
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
