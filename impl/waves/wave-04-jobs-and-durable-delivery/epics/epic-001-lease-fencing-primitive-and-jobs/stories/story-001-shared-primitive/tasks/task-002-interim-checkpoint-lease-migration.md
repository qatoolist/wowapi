---
id: W04-E01-S001-T002
type: task
title: Interim-checkpoint-lease migration
status: done
parent_story: W04-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E01-S001-T001
acceptance_criteria:
  - AC-W04-E01-S001-03
artifacts:
  - ART-W04-E01-S001-002
  - ART-W04-E01-S001-003
evidence:
  - EV-W04-E01-S001-003
---

# W04-E01-S001-T002 — Interim-checkpoint-lease migration

## Task Definition

### Task objective

Migrate any migration-checkpoint state written under W02-E01-S002's interim checkpoint lease onto
the shared lease/fencing primitive's schema, remove the interim lease code path, and prove no
in-flight backfill checkpoint state is lost or duplicated across the cutover — the planned
supersession named by RISK-W04-001 and its mirror RISK-W02-001.

### Parent story

W04-E01-S001 — Shared lease/fencing primitive.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E01-S001-T001 (the migration re-expresses interim-lease state under the primitive's own schema,
which must exist first).

### Detailed work

1. Locate and read W02-E01-S002's interim checkpoint-lease implementation and any persisted
   checkpoint state it has written (production, staging, or test fixtures) at this task's actual
   start commit.
2. Design the migration mechanics: an explicit migration step (not a big-bang cutover), per
   RISK-W04-001's mitigation — read-then-translate-then-remove, or a dual-write transition window;
   choose and document the rationale (resolves `plan.md`'s "Unresolved questions" item on migration
   mechanics).
3. Implement the migration: read existing interim-lease checkpoint state and re-express it under the
   shared primitive's schema.
4. Remove the interim lease code path once migration is confirmed complete.
5. Write a test simulating an in-flight backfill checkpoint written under the interim lease format,
   execute the migration, and confirm the checkpoint state is correctly readable under the new
   primitive's schema with no loss or duplication.
6. Document the completed migration.

### Expected files or components affected

W02-E01-S002's interim checkpoint-lease implementation file(s) (modified to migrate state, then
interim-specific code removed; exact file path TBD, not yet confirmed by file/line pending this
task's own start-commit re-read).

### Expected output

A completed, tested migration of any interim-lease checkpoint state onto the shared primitive's
schema, with the interim lease code path removed.

### Required artifacts

ART-W04-E01-S001-002 (migration tooling), ART-W04-E01-S001-003 (documentation, shared with T001).

### Required evidence

EV-W04-E01-S001-003 (migration-test report).

### Related acceptance criteria

AC-W04-E01-S001-03.

### Completion criteria

Any existing interim-lease checkpoint state is correctly re-expressed under the shared primitive's
schema, the interim lease code path is removed, and a test proves no checkpoint state was lost or
duplicated across the cutover.

### Verification method

Direct execution of the migration test against simulated interim-lease checkpoint state; inspection
confirming the interim lease code path no longer exists post-migration.

### Risks

RISK-W04-001 (an incorrectly translated checkpoint could cause a backfill job to reprocess or skip
rows, undermining DATA-09's own "no reprocessing or skipping" acceptance bar) — see epic-level
`risks.md`. This is the migration's central correctness risk, not an incidental concern.

### Rollback or recovery considerations

If a live backfill is genuinely in flight at cutover time, pause it, complete the migration, and
resume — do not cut over underneath a running job. Record the pause/resume as a deviation if it
occurs, per RISK-W04-001's own contingency.

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

*Not yet implemented — this task migrates persisted checkpoint state written under W02-E01-S002's
interim lease; recorded here once executed.*

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

*None anticipated — this task closes the RISK-W04-001 window rather than introducing new debt.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E01-S001-03 | Run the interim-checkpoint-lease migration test | Local dev or CI, PostgreSQL instance with simulated interim-lease checkpoint state | Checkpoint state correctly re-expressed; no loss or duplication across the cutover | migration-test report | unassigned |

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
