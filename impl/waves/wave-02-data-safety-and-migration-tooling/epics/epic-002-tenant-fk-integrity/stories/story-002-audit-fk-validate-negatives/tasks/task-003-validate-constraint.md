---
id: W02-E02-S002-T003
type: task
title: VALIDATE CONSTRAINT each new composite FK (DATA-01 T5)
status: implemented
parent_story: W02-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W02-E02-S002-T002
acceptance_criteria:
  - AC-W02-E02-S002-03
artifacts:
  - ART-W02-E02-S002-003
evidence:
  - EV-W02-E02-S002-004
  - EV-W02-E02-S002-005
---

# W02-E02-S002-T003 — VALIDATE CONSTRAINT each new composite FK (DATA-01 T5)

## Task Definition

### Task objective

Run `VALIDATE CONSTRAINT` for each of the 8 new composite FKs without blocking concurrent DML,
scheduled per W02-E01-S002's backfill/validate-phase tooling, and produce a second zero-mismatch
confirmation as part of this step.

### Parent story

W02-E02-S002 — Cross-tenant mismatch audit, composite FK validation, and negative tests.

### Owner

unassigned

### Status

todo

> **Status note (2026-07-16):** marked implemented (schema deliverable applied and independently verified live 2026-07-16), but the completion criterion's named proof artifact (concurrent-writer-load test) was never built — evidence index row remains "not yet produced".

### Dependencies

**W02-E02-S002-T002** (PLAN's own Depends-on column for T5: "T4"). **Hard cross-wave gate, same
condition as T002: W02-E01-S001 and W02-E01-S002 must both have reached `accepted` before this task
begins.** This gate is restated here per this story's own design goal that a task-level reader cannot
miss it by reading only this file.

### Detailed work

1. **Checkpoint: confirm W02-E01-S001 and W02-E01-S002 have both reached `accepted`** (re-confirm;
   this task may start materially later than T002 given `VALIDATE CONSTRAINT`'s I/O-bound duration).
2. Schedule each of the 8 `VALIDATE CONSTRAINT` runs per W02-E01-S002's own backfill/validate-phase
   tooling, per PLAN's own T5 risk note: "I/O-bound — schedule per DATA-09's backfill/validate
   phases."
3. Run the load test under concurrent writer load confirming validation does not block concurrent
   DML.
4. Produce a second zero-mismatch confirmation as part of this step (distinct from T001's own
   pre-validation audit — this is the post-validation confirmation PLAN's own T5 acceptance criterion
   requires).
5. If `VALIDATE CONSTRAINT` itself surfaces a mismatch the T001 audit missed (e.g. a row written
   between the audit and the validate step), treat it with the same severity as a T001-detected
   mismatch — escalate per RISK-W02-002's path, do not silently retry.

### Expected files or components affected

The 8 `VALIDATE CONSTRAINT` statements (one per edge, following T002's migrations).

### Expected output

All 8 composite FKs validated without blocking concurrent DML; a second zero-mismatch confirmation
produced.

### Required artifacts

ART-W02-E02-S002-003 (the 8 `VALIDATE CONSTRAINT` migrations/statements).

### Required evidence

EV-W02-E02-S002-004 (concurrent-writer-load test report), EV-W02-E02-S002-005 (second zero-mismatch
confirmation).

### Related acceptance criteria

AC-W02-E02-S002-03.

### Completion criteria

All 8 composite FKs validated; the concurrent-writer-load test confirms no DML blocking; the second
zero-mismatch confirmation is produced and recorded; the W02-E01 gate confirmed honored.

### Verification method

Load test under concurrent writer load; direct execution of the second zero-mismatch confirmation,
output retained as evidence.

### Risks

Same RISK-W02-E02-002 gate risk as T002. Additionally: `VALIDATE CONSTRAINT` is I/O-bound and scans
every existing row on all 8 edges — a mismatch surfacing here (rather than in T001) is treated as a
finding requiring the same RISK-W02-002 escalation path, not a silently retried operation.

### Rollback or recovery considerations

If `VALIDATE CONSTRAINT` surfaces an unexpected mismatch or blocks concurrent DML despite the load
test passing pre-production, the composite FK can be dropped per-edge without affecting the other 7
edges while the issue is investigated.

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

*Not yet implemented — anticipated: 8 `VALIDATE CONSTRAINT` statements.*

### Security changes

*Not applicable.*

### Observability changes

*Not yet implemented — anticipated: progress/duration visibility during the I/O-bound validation
run.*

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
| AC-W02-E02-S002-03 | Run `VALIDATE CONSTRAINT` for each of the 8 composite FKs under concurrent writer load — only after confirming W02-E01-S001 and W02-E01-S002 are `accepted`, and after AC-W02-E02-S002-02 passes | Staging or prod-shaped environment with concurrent-writer-load test harness | Validation completes without blocking concurrent DML; second zero-mismatch confirmation produced | load-test report + second audit report | unassigned |

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
