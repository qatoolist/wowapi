---
id: W02-E04-S001-T003
type: task
title: Reference-handler migration
status: done
parent_story: W02-E04-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W02-E04-S001-T001
  - W02-E04-S001-T002
acceptance_criteria:
  - AC-W02-E04-S001-03
artifacts:
  - ART-W02-E04-S001-003
evidence:
  - EV-W02-E04-S001-003
---

# W02-E04-S001-T003 — Reference-handler migration

## Task Definition

### Task objective

Migrate the framework's reference handler onto the new T1/T2 helper, so it no longer performs the
business-row write and mirror upsert as two independent statements, and confirm existing reference
tests continue to pass.

### Parent story

W02-E04-S001 — Typed aggregate write contract with mandatory mirror, audit, and outbox.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E04-S001-T001 (the helper must exist), W02-E04-S001-T002 (the actor-attribution fix must be in
place before the reference handler is migrated onto the completed contract).

### Detailed work

1. Confirm the exact identity (fully-qualified path) of "the reference handler" PLAN's DATA-06 T3
   refers to, at this task's actual start commit — resolving `plan.md`'s "Unresolved questions" item
   on this point; record the choice and rationale if more than one candidate exists.
2. Migrate the reference handler's business-row write and mirror upsert (currently two independent
   statements) onto a single call to the new T1/T2 helper.
3. Run the existing reference-handler test suite, confirming all tests pass with unchanged
   observable behavior.

### Expected files or components affected

The reference handler's source file (exact path TBD).

### Expected output

A reference handler that calls the new helper instead of performing two independent statements,
with its existing tests passing unmodified in behavior.

### Required artifacts

ART-W02-E04-S001-003 (migrated reference handler).

### Required evidence

EV-W02-E04-S001-003 (regression-test report).

### Related acceptance criteria

AC-W02-E04-S001-03.

### Completion criteria

The reference handler no longer manually performs two independent statements; all existing
reference tests pass, proven against a named commit SHA.

### Verification method

Direct execution of the existing reference-handler test suite before and after the migration,
confirming no behavior regression.

### Risks

PLAN T3's own named risk: "Fix the reference pattern before it's copied further" — this task exists
specifically because the reference handler is the pattern other module authors are expected to
copy; a failure to migrate it correctly perpetuates the exact defect class DATA-06 targets.

### Rollback or recovery considerations

Revert the migration if the reference-handler test suite regresses in a way that cannot be resolved
within this task's bounded scope; escalate rather than silently modifying the existing tests to
accommodate an unintended behavior change.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not applicable — this task consumes T001/T002's interface, it does not introduce a new one.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable — inherits T002's actor-attribution fix by consuming the helper; no additional
security change introduced by this task itself.*

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
| AC-W02-E04-S001-03 | Run the existing reference-handler test suite after migration | Local dev or CI | All existing tests pass; handler calls the new helper, not two independent statements | regression-test report | unassigned |

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
