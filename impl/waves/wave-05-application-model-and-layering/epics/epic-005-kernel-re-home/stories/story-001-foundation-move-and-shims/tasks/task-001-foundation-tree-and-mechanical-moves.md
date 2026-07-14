---
id: W05-E05-S001-T001
type: task
title: Foundation tree creation and 8 mechanical package moves
status: todo
parent_story: W05-E05-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E05-S001-01
artifacts:
  - ART-W05-E05-S001-001
evidence:
  - EV-W05-E05-S001-001
---

# W05-E05-S001-T001 — Foundation tree creation and 8 mechanical package moves

## Task Definition

### Task objective

Create the `foundation/` tree, and `git mv` the 8 zero-consumer-outside-wowapi packages
(`webhook, notify, document, artifact, attachment, comment, bulk, integration`) from `kernel/` to
`foundation/`, updating import paths repo-wide, with a passing full build as proof.

### Parent story

W05-E05-S001 — Foundation tree, package moves, and mfa forwarding shim.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T002 — disjoint package set); depends on W05-E01 and
W05-E02 at story scope.

### Detailed work

1. Re-confirm `go list ./kernel/...` and the zero-external-consumer status of the 8 packages (beyond
   `mfa`) at this task's actual start commit.
2. Create the `foundation/` top-level tree.
3. `git mv` each of the 8 packages, preserving history.
4. Update every import path repo-wide referencing the old `kernel/<pkg>` location.
5. Run a full repository build, confirming success.
6. Document the move.

### Expected files or components affected

`foundation/webhook`, `foundation/notify`, `foundation/document`, `foundation/artifact`,
`foundation/attachment`, `foundation/comment`, `foundation/bulk`, `foundation/integration`; every
file repo-wide importing any of these 8.

### Expected output

The 8 packages moved to `foundation/`, all import paths updated, full build passing.

### Required artifacts

ART-W05-E05-S001-001.

### Required evidence

EV-W05-E05-S001-001.

### Related acceptance criteria

AC-W05-E05-S001-01.

### Completion criteria

The full build succeeds post-move; `git log` confirms history preservation for the moved files.

### Verification method

Direct execution of the full repository build; `git log --follow` on a sample of moved files to
confirm history preservation.

### Risks

Low — MATRIX CS-01's own framing: "mechanical, 8 of 9 are zero-consumer outside wowapi,"
"behaviour-preserving moves."

### Rollback or recovery considerations

If the build fails post-move, fix the specific broken import path(s) before proceeding — a
mechanical move should not require any logic change.

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
| AC-W05-E05-S001-01 | Run a full repository build; inspect git history | Local dev or CI, Go toolchain | Build succeeds; history preserved for moved files | build-output report | unassigned |

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
