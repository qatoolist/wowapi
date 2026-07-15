---
id: W04-E04-S002-T004
type: task
title: Explicit partial/not-applicable per-class DSR status
status: done
parent_story: W04-E04-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E04-S002-T002
  - W04-E04-S002-T003
acceptance_criteria:
  - AC-W04-E04-S002-04
artifacts:
  - ART-W04-E04-S002-005
  - ART-W04-E04-S002-006
evidence:
  - EV-W04-E04-S002-005
---

# W04-E04-S002-T004 — Explicit partial/not-applicable per-class DSR status

## Task Definition

### Task objective

Ensure the DSR result set explicitly lists every registered record class with a status (exported,
erased, not-applicable, or partial), coordinated with T002's DSR export artifact manifest shape, so
that a record class without an export/erase callback is never silently omitted from the result.

### Parent story

W04-E04-S002 — External anchoring, DSR export artifact, central legal-hold, and explicit per-class
status.

### Owner

unassigned

### Status

done

### Dependencies

W04-E04-S002-T002 (this task's status reporting is coordinated with T002's manifest shape, per PLAN
DATA-08 W6-T5's own dependency column: "W6-T3, W6-T4"), W04-E04-S002-T003 (status reporting must
reflect the enumerated `RecordClass` set T003 establishes).

### Detailed work

1. Confirm T002's manifest shape (per-class result schema) and T003's `RecordClass` enumeration
   record are both available and stable enough to build against.
2. Design the explicit-status vocabulary (exported, erased, not-applicable, partial) and integrate it
   into the DSR result set's own schema.
3. Implement the logic ensuring every registered record class — including those without an export/
   erase callback — appears in the result set with an explicit status, never a silent omission.
4. Write the explicit-status test: confirm the result set lists every registered class with a status
   for a representative mix of classes (some with callbacks, some without).
5. Document the status vocabulary and how it integrates with T002's manifest.

### Expected files or components affected

The DSR result-set construction logic (exact location TBD, expected within or near `kernel/
retention`'s orchestration code); a new test file for the explicit-status test.

### Expected output

A DSR result set that explicitly lists every registered class with a status; a passing explicit-
status test; documentation of the status vocabulary.

### Required artifacts

ART-W04-E04-S002-005 (explicit per-class DSR status reporting mechanism), ART-W04-E04-S002-006
(documentation, shared with T001/T002/T003).

### Required evidence

EV-W04-E04-S002-005 (explicit-status test report).

### Related acceptance criteria

AC-W04-E04-S002-04 (status half, jointly with T003's enumeration half).

### Completion criteria

The DSR result set explicitly lists every registered record class with a status; the explicit-status
test confirms no class is silently omitted, including classes without an export/erase callback.

### Verification method

Direct execution of the explicit-status test.

### Risks

Low-medium — the primary risk is the status schema being incompatible with T002's manifest shape if
the two are not genuinely coordinated, per PLAN W6-T5's own dependency framing.

### Rollback or recovery considerations

Additive result-set schema change — revertible without a data-migration concern, since the underlying
`Dispose`/`Erase` callback behavior is unchanged by this task; only the result set's own reporting
completeness changes.

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
| AC-W04-E04-S002-04 (status half) | Run explicit-status test | Local dev or CI | Every registered class appears in the result set with a status, none omitted | explicit-status test report | unassigned |

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
