---
id: W02-E02-S001-T003
type: task
title: CI gate wiring
status: todo
parent_story: W02-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W02-E02-S001-T002
acceptance_criteria:
  - AC-W02-E02-S001-03
artifacts:
  - ART-W02-E02-S001-003
  - ART-W02-E02-S001-004
  - ART-W02-E02-S001-005
evidence:
  - EV-W02-E02-S001-003
---

# W02-E02-S001-T003 — CI gate wiring

## Task Definition

### Task objective

Wire the T002 tenant-FK catalog scanner into a permanent CI gate, so that any future migration
adding a single-column (non-composite) tenant-table FK fails the build — per PLAN DATA-01 T6's own
acceptance criterion and its risk note: "Cheapest, most durable part — do first if sequencing
allows."

### Parent story

W02-E02-S001 — Parent tenant-scoped unique indexes, FK catalog scanner, and CI gate.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E02-S001-T002 (the scanner must exist before it can be wired into CI). PLAN's own Depends-on
column for T6 lists "T2."

### Detailed work

1. Extend the existing CI infrastructure to invoke the T002 scanner as a build step, consistent with
   how W02-E01-S001's manifest-schema validation is wired in.
2. Write a negative fixture migration: a migration adding a single-column, non-composite tenant-table
   FK.
3. Confirm the CI gate actually fails the build against the negative fixture migration, with an error
   message identifying the specific migration and FK.
4. Confirm the CI gate does not produce a false-positive rejection against the existing (post-T001)
   composite-compliant schema.
5. Document the CI gate's failure behavior and what a migration author should do instead.

### Expected files or components affected

CI configuration (e.g. `.github/workflows/`), extended to invoke the scanner. A new negative fixture
migration file.

### Expected output

A permanent CI gate that fails a build against a migration adding a non-composite tenant FK, proven
by the negative fixture migration's actual CI run.

### Required artifacts

ART-W02-E02-S001-003 (CI gate wiring), ART-W02-E02-S001-004 (negative fixture migration),
ART-W02-E02-S001-005 (documentation, shared with T002).

### Required evidence

EV-W02-E02-S001-003 (CI run output, negative fixture).

### Related acceptance criteria

AC-W02-E02-S001-03.

### Completion criteria

The negative fixture migration's CI run fails with an error message identifying the specific
non-composite FK; a legitimate migration continues to pass.

### Verification method

Direct execution of CI against both the negative fixture migration (expect fail) and the existing
composite-compliant schema (expect pass).

### Risks

A false-positive CI rejection of a legitimate migration would block unrelated work — this task's own
completion criteria explicitly requires confirming the gate does not misfire against the compliant
schema, not merely that it correctly fails the negative fixture.

### Rollback or recovery considerations

If the gate produces a false positive after landing, it may be reverted (disabled) — but any such
reversion must be recorded as a deviation per `story.md` "Rollback strategy," not silently applied.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

*Not yet implemented — expected: CI workflow configuration extended to invoke the scanner.*

### Schema or migration changes

*Not yet implemented — expected: one negative fixture migration (test-only, not applied to any real
environment).*

### Security changes

*Not applicable.*

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

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E02-S001-03 | Run CI against the negative fixture migration and against the compliant schema | CI | Negative fixture fails with a specific error; compliant schema passes | CI run output | unassigned |

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
