---
id: W06-E02-S003-T003
type: task
title: Generated-consumer upgrade check (blocked on W06-E01-S002)
status: blocked
parent_story: W06-E02-S003
owner: W06E02Impl
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W06-E01-S002
acceptance_criteria:
  - AC-W06-E02-S003-03
artifacts:
  - ART-W06-E02-S003-003
evidence:
  - EV-W06-E02-S003-003
---

# W06-E02-S003-T003 — Generated-consumer upgrade check (blocked on W06-E01-S002)

## Task Definition

### Task objective

Implement the generated-consumer upgrade check, reusing DX-04's own drill. BLOCKED: cannot begin until W06-E01-S002 (DX-04) reaches accepted.

### Parent story

W06-E02-S003

### Owner

unassigned

### Status

todo

### Dependencies

W06-E01-S002 (DX-04) must be `accepted` — this task hard-depends on DX-04's golden-consumer fixture and its upgrade-replay drill existing first, per PLAN's own framing ('cannot exist before DX-04'). This task must not begin before that entry criterion is satisfied.

### Detailed work

1. Confirm W06-E01-S002 has reached `accepted`.
2. Reuse DX-04's own upgrade-replay drill (W06-E01-S002's T4) rather than building a second one.
3. Wire a REL-03-specific invocation of that drill, confirming golden-consumer contracts re-pass after
   an N-1-to-N upgrade.

### Expected files or components affected

A CI-job wrapper invoking W06-E01-S002's existing drill.

### Expected output

Confirmation that golden-consumer contracts re-pass after an N-1-to-N upgrade, via a REL-03-scoped invocation of DX-04's own drill.

### Required artifacts

ART-W06-E02-S003-003 (generated-consumer upgrade check invocation).

### Required evidence

EV-W06-E02-S003-003 (generated-consumer N-1-to-N upgrade re-pass output).

### Related acceptance criteria

AC-W06-E02-S003-03.

### Completion criteria

Contracts re-pass after the upgrade, via the reused drill.

### Verification method

Direct execution of the REL-03-scoped invocation, once unblocked.

### Risks

The primary risk is silently re-implementing a second upgrade-replay drill instead of reusing DX-04's — mitigated by this task's own explicit reuse requirement.

### Rollback or recovery considerations

If begun prematurely, or if a second drill is accidentally built instead of reusing DX-04's, halt and record a deviation.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not yet implemented.*

### Configuration changes

*Not yet implemented.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not yet implemented.*

### Observability changes

*Not yet implemented.*

### Tests added or modified

*Not yet implemented.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*None anticipated.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented. Once implementation occurs, record whether it matched `plan.md`; if not,
reference the corresponding entry in `deviations.md`.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E02-S003-03 | Run the REL-03-scoped invocation of DX-04's drill, once W06-E01-S002 is accepted | CI, real infrastructure | Contracts re-pass after the N-1-to-N upgrade | two-pass integration-test report | unassigned |

### Actual result

*Not yet executed.*

### Pass or fail

*Not yet executed.*

### Evidence identifier

*Not yet executed.*

### Execution date

*Not yet executed.*

### Commit or revision

*Not yet executed.*

### Environment

*Not yet executed.*

### Reviewer

*Not yet executed.*

### Findings

*Not yet executed.*

### Retest status

*Not yet executed.*

### Final conclusion

*Not yet executed.*

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
