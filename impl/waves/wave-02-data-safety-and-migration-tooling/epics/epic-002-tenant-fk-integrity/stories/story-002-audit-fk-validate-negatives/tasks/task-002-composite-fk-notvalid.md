---
id: W02-E02-S002-T002
type: task
title: Composite FK NOT VALID add, all 8 edges (DATA-01 T4)
status: todo
parent_story: W02-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W02-E02-S002-T001
acceptance_criteria:
  - AC-W02-E02-S002-02
artifacts:
  - ART-W02-E02-S002-002
evidence:
  - EV-W02-E02-S002-003
---

# W02-E02-S002-T002 — Composite FK NOT VALID add, all 8 edges (DATA-01 T4)

## Task Definition

### Task objective

Add composite FK `NOT VALID` for all 8 tenant-scoped edges, run per-table as separate statements,
each add staying under the DATA-09 2-second lock-timeout budget.

### Parent story

W02-E02-S002 — Cross-tenant mismatch audit, composite FK validation, and negative tests.

### Owner

unassigned

### Status

todo

### Dependencies

**W02-E02-S002-T001** (PLAN's own Depends-on column for T4: "T1, T3" — T001 here is this story's T3
equivalent; T1's parent-index prerequisite is W02-E02-S001's own T001, already a story-level
dependency). **Hard cross-wave gate, not merely a sequencing preference: W02-E01-S001 and
W02-E01-S002 must both have reached `accepted` before this task begins**, per PLAN's own PF-DATA
cross-cutting note (6): "sequence DATA-09 T1-T5 ahead of DATA-01 T4/T5... in the real release plan."
This gate is restated here per this story's own design goal that a task-level reader cannot miss it
by reading only this file.

### Detailed work

1. **Checkpoint: confirm W02-E01-S001 and W02-E01-S002 have both reached `accepted`.** Do not
   proceed past this step until confirmed.
2. Confirm T001's mismatch audit resolved to zero-mismatch (or a resolved remediation decision).
3. For each of the 8 edges, run `ALTER TABLE ... ADD CONSTRAINT ... FOREIGN KEY (tenant_id, parent_id)
   REFERENCES parent(tenant_id, id) NOT VALID` as its own statement (not batched with any other
   edge's statement), per PLAN's own T4 risk note: "Run per-table as separate statements."
4. Measure each statement's lock duration and confirm it stays under the DATA-09 2-second
   lock-timeout budget provided by W02-E01-S001.
5. Write the migration lock-duration test.

### Expected files or components affected

8 new migration files (or migration statement groups), one per edge, located under the existing
migration directory structure.

### Expected output

All 8 edges carry a composite FK added `NOT VALID`, each add measured and confirmed under the
DATA-09 2-second lock-timeout budget.

### Required artifacts

ART-W02-E02-S002-002 (the 8 composite-FK `NOT VALID` migrations).

### Required evidence

EV-W02-E02-S002-003 (migration lock-duration test report).

### Related acceptance criteria

AC-W02-E02-S002-02.

### Completion criteria

All 8 `NOT VALID` adds applied; each statement's lock duration confirmed under the DATA-09 budget by
the lock-duration test; the W02-E01 gate confirmed honored (task started only after both W02-E01-S001
and W02-E01-S002 reached `accepted`).

### Verification method

Direct migration execution (per-table), measured lock-duration test, logged output retained as
evidence.

### Risks

If the W02-E01 gate is not honored (this task starts before W02-E01-S001/S002 reach `accepted`), the
framework's highest-severity confirmed data-integrity fix ships without the safety tooling DATA-09
exists to provide — see RISK-W02-E02-002 in `../../risks.md`.

### Rollback or recovery considerations

A `NOT VALID` add is metadata-only (no existing-row scan); it can be dropped per-edge without
affecting the other 7 edges if an unexpected issue is found before `VALIDATE CONSTRAINT` runs.

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

*Not yet implemented — anticipated: 8 composite FK `NOT VALID` adds.*

### Security changes

*Not applicable — schema-integrity change, not an access-control change.*

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
| AC-W02-E02-S002-02 | Run the composite FK `NOT VALID` add per-table, measure lock duration against the DATA-09 budget — only after confirming W02-E01-S001 and W02-E01-S002 are `accepted` | CI or staging environment with DATA-09 lock-timeout tooling available | Each per-table `NOT VALID` add stays under the 2-second lock-timeout budget | migration lock-duration report | unassigned |

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
