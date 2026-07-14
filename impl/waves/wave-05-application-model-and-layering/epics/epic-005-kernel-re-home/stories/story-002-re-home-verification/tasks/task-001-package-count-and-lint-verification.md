---
id: W05-E05-S002-T001
type: task
title: Kernel package-count and lint verification
status: todo
parent_story: W05-E05-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E05-S002-01
artifacts:
  - ART-W05-E05-S002-001
evidence:
  - EV-W05-E05-S002-001
---

# W05-E05-S002-T001 — Kernel package-count and lint verification

## Task Definition

### Task objective

Run and record `go list ./kernel/... | wc -l` against S001's post-move state, confirming it is at or
below S001-T004's own target-list count, and confirm depguard and boundaries lint are both green.

### Parent story

W05-E05-S002 — Kernel package-count and wowsociety identity-suite verification.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T002); depends on W05-E05-S001 at story scope.

### Detailed work

1. Confirm W05-E05-S001 has landed.
2. Run `go list ./kernel/... | wc -l` against the current state.
3. Compare against S001-T004's own final target-list count.
4. Run the full depguard and boundaries-lint suites, confirming green.
5. Record the results in the verification-results document.

### Expected files or components affected

None (verification-only; produces a documentation artifact).

### Expected output

A recorded package count at or below target, with both lints confirmed green.

### Required artifacts

ART-W05-E05-S002-001 (shared with T002).

### Required evidence

EV-W05-E05-S002-001.

### Related acceptance criteria

AC-W05-E05-S002-01.

### Completion criteria

The count is confirmed at or below target; both lints are confirmed green.

### Verification method

Direct execution of `go list ./kernel/... | wc -l` and both lint suites.

### Risks

Low — a mechanical verification step, assuming S001 has landed correctly.

### Rollback or recovery considerations

If the count exceeds target or either lint is red, record as a finding requiring S001's own
follow-up — do not silently adjust the target count to make this check pass.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not applicable.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not applicable — this task runs existing checks.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*Not applicable.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W05-E05-S002-01 | Run `go list ./kernel/... \| wc -l` and both lint suites | Local dev or CI, Go toolchain (lint) | Count at or below target; both lints green | count + lint report | unassigned |

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
