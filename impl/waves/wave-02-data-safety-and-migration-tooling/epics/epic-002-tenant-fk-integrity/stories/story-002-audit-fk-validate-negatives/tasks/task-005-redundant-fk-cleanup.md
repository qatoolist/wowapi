---
id: W02-E02-S002-T005
type: task
title: Optional redundant single-column FK cleanup (DATA-01 T8)
status: done
parent_story: W02-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W02-E02-S002-T003
  - W02-E02-S002-T004
acceptance_criteria:
  - AC-W02-E02-S002-05
artifacts:
  - ART-W02-E02-S002-005
evidence:
  - EV-W02-E02-S002-007
---

# W02-E02-S002-T005 — Optional redundant single-column FK cleanup (DATA-01 T8)

## Task Definition

### Task objective

Remove the redundant single-column FKs, only after all consumers and rollback paths have been
verified not to depend on the old FK's cascade behavior. **Optional — per PLAN's own T8 framing,
non-completion of this task does not block this story's or the epic's P0 closure.**

### Parent story

W02-E02-S002 — Cross-tenant mismatch audit, composite FK validation, and negative tests.

### Owner

unassigned

### Status

todo

### Dependencies

**W02-E02-S002-T003, W02-E02-S002-T004** (PLAN's own Depends-on column for T8: "T5, T7" — T003/T004
here are this story's T5/T7 equivalents).

### Detailed work

1. Grep-sweep the codebase and all migrations for any reference to the old single-column FK
   constraint name, on any of the 8 edges.
2. Run the full regression suite to confirm no code path relies on the old FK's cascade behavior.
3. If the sweep and regression both come back clean, author and run the FK-removal migration for the
   8 redundant single-column FKs.
4. If either the sweep or the regression surfaces a dependency on the old FK, do not proceed with
   removal — record the finding and treat this task as intentionally not completed (an optional,
   deferred item), not a failed acceptance criterion.

### Expected files or components affected

If pursued: 8 further migration files removing the old single-column FK definitions.

### Expected output

Either: the 8 redundant single-column FKs removed, with a recorded consumer/rollback verification
showing no dependency existed; or: an explicit, recorded decision to defer this task, per its own
optional status.

### Required artifacts

ART-W02-E02-S002-005 (the FK-removal migrations and consumer/rollback verification record, if
pursued).

### Required evidence

EV-W02-E02-S002-007 (regression + grep sweep output).

### Related acceptance criteria

AC-W02-E02-S002-05.

### Completion criteria

Either the removal is completed with a clean regression + grep sweep recorded, or the task is
explicitly recorded as deferred per its own optional, non-blocking status — not left in an ambiguous
"todo forever" state without a recorded disposition.

### Verification method

Grep sweep across the codebase and migrations; full regression suite execution; output retained as
evidence regardless of which branch (proceed or defer) is taken.

### Risks

Low — PLAN's own risk framing for T8 does not assign it an elevated risk; the primary risk is
proceeding with removal without a genuinely thorough consumer/rollback-path check, which the grep +
regression combination is designed to catch.

### Rollback or recovery considerations

If pursued and later found to have broken an undiscovered consumer, the removed single-column FK
would need to be re-created and re-validated — a more expensive rollback than T002/T003's own
per-edge drop, which is exactly why this task requires the sweep to be clean before proceeding and
remains optional.

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

*Not applicable.*

### Schema or migration changes

*Not yet implemented — anticipated, if pursued: 8 FK-removal migrations.*

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

*None anticipated. If this task is deferred, the redundant single-column FKs remaining in place is
recorded as an intentional, non-blocking deferral, not technical debt in the negative sense.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented. If deferred, recorded here as a follow-up item for a future story or wave.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E02-S002-05 (optional) | If pursued: grep sweep for references to the old single-column FK name, plus a full regression run, before the FK-removal migration | CI environment | No code relies on the old FK name for cascade behavior; full regression passes | regression report + grep sweep output | unassigned |

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
