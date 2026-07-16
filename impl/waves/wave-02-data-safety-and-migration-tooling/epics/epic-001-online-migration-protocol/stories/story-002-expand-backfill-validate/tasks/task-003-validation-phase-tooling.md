---
id: W02-E01-S002-T003
type: task
title: Validation-phase tooling
status: done
parent_story: W02-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W02-E01-S002-T002
acceptance_criteria:
  - AC-W02-E01-S002-03
artifacts:
  - ART-W02-E01-S002-003
  - ART-W02-E01-S002-004
evidence:
  - EV-W02-E01-S002-003
---

# W02-E01-S002-T003 — Validation-phase tooling

## Task Definition

### Task objective

Implement validation-phase tooling — `VALIDATE CONSTRAINT` orchestration plus reconciliation
queries with artifact capture — such that zero-mismatch reports are machine-checked artifacts, not
prose, per PLAN DATA-09 T5's acceptance criterion.

### Parent story

W02-E01-S002 — Expand-phase tooling, resumable backfill harness, and validation-phase tooling.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E01-S002-T002 (PLAN T5's own "Depends-on" column names T4 — validation runs after backfill).

### Detailed work

1. Define the validation-report artifact schema: per-constraint/per-query mismatch counts,
   pass/fail, timestamps, and whatever identity fields make the report machine-checkable rather
   than free-form prose.
2. Implement `VALIDATE CONSTRAINT` orchestration: run validation for each `NOT VALID` constraint
   the expand phase added, within the lock-budget discipline S001 established.
3. Implement reconciliation-query execution: run each migration's manifest-declared validation
   query (the manifest field S001's schema defines) and capture results into the artifact.
4. Write the artifact-schema test: confirm a validation report conforms to its schema and correctly
   reports zero mismatches on clean data; where feasible, also confirm a seeded mismatch is
   correctly surfaced rather than silently passed (fail-loud behavior per `plan.md`'s error-handling
   strategy).
5. Document the artifact schema and note PLAN T5's own human-in-the-loop boundary: "human review of
   the report before canary" — the tooling produces the machine-checked artifact; the go-to-canary
   decision remains a human step, consumed by W02-E01-S003.

### Expected files or components affected

New validation-phase tooling package (exact location TBD per `plan.md`).

### Expected output

Validation-phase tooling producing machine-checked, artifact-schema-conformant zero-mismatch
reports.

### Required artifacts

ART-W02-E01-S002-003 (validation-phase tooling + artifact schema), ART-W02-E01-S002-004
(documentation, shared with T001/T002).

### Required evidence

EV-W02-E01-S002-003 (artifact-schema test output).

### Related acceptance criteria

AC-W02-E01-S002-03.

### Completion criteria

The artifact-schema test passes: reports conform to the schema, zero mismatches on clean data are
reported as such, and (where implemented) a seeded mismatch is surfaced — evidenced against a named
commit SHA.

### Verification method

Direct execution of the artifact-schema test, logged output retained as evidence per
`evidence/index.md`.

### Risks

PLAN T5's own classification: "Code for the harness; human review of the report before canary" —
the residual risk is process-level (a human skipping the report review before canary), not
tooling-level; the tooling's own risk is low once the artifact-schema test passes.

### Rollback or recovery considerations

Reverting the tooling is a plain code revert with no data impact; a validation run itself is
read-only apart from `VALIDATE CONSTRAINT`'s metadata effect, which is not destructive.

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

*Not applicable — validation tooling reads and validates; it does not alter application schema.*

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
| AC-W02-E01-S002-03 | Artifact-schema test against the validation-phase report | Local dev or CI, PostgreSQL | Report conforms to schema; zero mismatches correctly reported; seeded mismatch surfaced (where implemented) | artifact-schema test report | unassigned |

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
