---
id: W05-E02-S003-T001
type: task
title: Lifecycle manifest retirement
status: todo
parent_story: W05-E02-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E02-S003-01
artifacts:
  - ART-W05-E02-S003-001
evidence:
  - EV-W05-E02-S003-001
---

# W05-E02-S003-T001 — Lifecycle manifest retirement

## Task Definition

### Task objective

Retire the hand-maintained `kernel/lifecycle` manifest (`lifecycle.go`/`manifest.go`) in favor of
generation from S002's provider graph, preserving the existing 5 lint-failure classes as
data-driven checks.

### Parent story

W05-E02-S003 — Lifecycle manifest retirement and legacy port adapter.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T002); depends on W05-E02-S002 at story scope.

### Detailed work

1. Determine whether `lifecycle.go`/`manifest.go` can be deleted outright or must be generated.
2. Implement the chosen approach.
3. Write the regression test proving the existing 5 lint-failure classes still pass, now
   data-driven.

### Expected files or components affected

`kernel/lifecycle/lifecycle.go`, `kernel/lifecycle/manifest.go`.

### Expected output

`lifecycle.go`/`manifest.go` deleted or generated; existing lint classes pass.

### Required artifacts

ART-W05-E02-S003-001.

### Required evidence

EV-W05-E02-S003-001.

### Related acceptance criteria

AC-W05-E02-S003-01.

### Completion criteria

The regression test confirms all 5 existing lint-failure classes still pass.

### Verification method

Direct execution of the regression test.

### Risks

Low-medium, per PLAN T6's own risk column.

### Rollback or recovery considerations

If the regression test fails for any lint class, fix before proceeding — do not ship a manifest
retirement that silently drops lint coverage.

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

*Not applicable.*

### Observability changes

*Not applicable.*

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

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W05-E02-S003-01 | Run the lifecycle-lint regression test | Local dev or CI, Go toolchain | Existing 5 lint-failure classes pass, now data-driven | regression-test report | unassigned |

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
