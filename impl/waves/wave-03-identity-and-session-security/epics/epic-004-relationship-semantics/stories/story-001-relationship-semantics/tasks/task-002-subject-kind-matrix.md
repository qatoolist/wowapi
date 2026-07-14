---
id: W03-E04-S001-T002
type: task
title: Checker.Has full subject-kind matrix (DATA-07 T2)
status: todo
parent_story: W03-E04-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W03-E04-S001-T001
acceptance_criteria:
  - AC-W03-E04-S001-02
artifacts:
  - ART-W03-E04-S001-002
evidence:
  - EV-W03-E04-S001-002
---

# W03-E04-S001-T002 — Checker.Has full subject-kind matrix (DATA-07 T2)

## Task Definition

### Task objective

Extend `Checker.Has` to cover every schema-enumerated `subject_kind`, with an unsupported or
unenumerated kind failing closed.

### Parent story

W03-E04-S001 — Relationship semantics — party-subject evaluation, full subject-kind matrix, mutation
governance.

### Owner

unassigned

### Status

todo

### Dependencies

W03-E04-S001-T001 — PLAN's own Depends-on column for T2: "T1."

### Detailed work

1. Enumerate every schema-defined `subject_kind` value at this task's actual start commit.
2. **Confirm which enumerated kinds are live requirements versus dead schema surface first** — per
   T2's own risk note, verbatim: "Confirm which enumerated kinds are live requirements vs. dead
   schema surface first." Do not write speculative evaluation branches for schema surface with no
   live requirement without first confirming that classification.
3. Add an explicit evaluation branch in `Checker.Has` for each live-requirement `subject_kind`.
4. Add an explicit fail-closed default branch for any unenumerated or unsupported kind — returning a
   distinguishable "denied, unsupported kind" result, not a generic or ambiguous error.
5. Write the subject-kind matrix test covering every enumerated kind, plus a fail-closed test case for
   a deliberately unenumerated kind.

### Expected files or components affected

`kernel/relationship/relationship.go:42-66`.

### Expected output

Every schema-enumerated, live-requirement `subject_kind` has an explicit evaluation branch; an
unsupported/unenumerated kind fails closed.

### Required artifacts

ART-W03-E04-S001-002 (the full subject-kind evaluation matrix with fail-closed handling).

### Required evidence

EV-W03-E04-S001-002 (subject-kind matrix test output, including the fail-closed case).

### Related acceptance criteria

AC-W03-E04-S001-02.

### Completion criteria

The matrix test confirms every live-requirement kind has a correct evaluation branch; the fail-closed
test proves an unenumerated kind is denied, not silently `true` or silently ignored.

### Verification method

Direct test execution against a testkit DB, logged output retained as evidence.

### Risks

If the dead-schema-surface-versus-live-requirement classification (step 2) is skipped or done
carelessly, this task risks either under-covering a genuinely live kind or over-building evaluation
logic for schema surface with no real caller — both are explicitly flagged by PLAN's own risk note.

### Rollback or recovery considerations

Additive evaluation-branch extensions, plus the fail-closed default — independently revertible per
branch if an issue is found with a specific kind's evaluation logic.

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

*Not applicable.*

### Security changes

*Not yet implemented — anticipated: the fail-closed default for unsupported kinds is itself a
security control.*

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
| AC-W03-E04-S001-02 | Run the subject-kind matrix test across every schema-enumerated `subject_kind`, including a deliberately unenumerated kind | Local dev or CI, testkit DB | Every enumerated kind has a correct evaluation branch; the unenumerated kind fails closed | subject-kind matrix test report | unassigned |

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
