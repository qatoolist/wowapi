---
id: W03-E04-S001-T001
type: task
title: Checker.Has party-subject evaluation (DATA-07 T1)
status: todo
parent_story: W03-E04-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W03-E04-S001-01
artifacts:
  - ART-W03-E04-S001-001
evidence:
  - EV-W03-E04-S001-001
---

# W03-E04-S001-T001 — Checker.Has party-subject evaluation (DATA-07 T1)

## Task Definition

### Task objective

Resolve actor → active capacity → optional party through the post-SEC-01 authoritative principal
model, so `Checker.Has` can evaluate party-subject edges (today, per the code's own comment, "not
consulted yet").

### Parent story

W03-E04-S001 — Relationship semantics — party-subject evaluation, full subject-kind matrix, mutation
governance.

### Owner

unassigned

### Status

todo

### Dependencies

None within this task's own prerequisites, but **hard, blocking gate at story scope: W03-E01 must
have reached `accepted`** before this task begins — per PLAN's own DATA-07 T1 Depends-on column:
"Hard dependency on PF-SEC's SEC-01 — do not schedule before it lands." This gate is restated here
per this story's own design goal that a task-level reader cannot miss it by reading only this file.

### Detailed work

1. **Checkpoint: confirm W03-E01 has reached `accepted`.** Do not proceed past this step until
   confirmed.
2. Read `kernel/relationship/relationship.go:42-66` (`Checker.Has`) at this task's actual start
   commit, confirming its exact current `subject_kind='capacity'`-only filtering behavior.
3. Read W03-E01's actual, accepted principal-model implementation to confirm the exact "actor →
   active capacity → optional party" resolution path it exposes (or requires a small additive
   extension to expose — to be confirmed, not assumed).
4. Extend `Checker.Has` to resolve through that path and evaluate party-subject edges.
5. Write the seeded party-subject-edge test: seed a party-subject edge, resolve an actor carrying a
   party, assert the previously-false evaluation is now `true`.

### Expected files or components affected

`kernel/relationship/relationship.go:42-66`.

### Expected output

`Checker.Has` correctly evaluates party-subject edges via the post-SEC-01 principal model.

### Required artifacts

ART-W03-E04-S001-001 (`Checker.Has`'s extended party-subject evaluation logic).

### Required evidence

EV-W03-E04-S001-001 (party-subject-edge seeded test output).

### Related acceptance criteria

AC-W03-E04-S001-01.

### Completion criteria

The seeded test proves a party-subject edge, resolved through an actor carrying a party, evaluates
`true` where it was previously `false`.

### Verification method

Direct test execution against a testkit DB, logged output retained as evidence.

### Risks

If W03-E01's actual, accepted principal-model shape differs materially from what this task assumes
during premature implementation (i.e. if the W03-E01 gate is not honored), rework is required — see
RISK-W03-E04-002 (epic-level `risks.md`).

### Rollback or recovery considerations

This is an additive extension to `Checker.Has`'s evaluation logic — revertible independently of T002/
T003 if an issue is found, since it touches a distinct evaluation branch from T002's matrix work.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not applicable — `Checker.Has`'s external signature is not expected to change.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not yet implemented — anticipated: closes the party-subject-edge evaluation gap.*

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
| AC-W03-E04-S001-01 | Seed a party-subject edge; resolve an actor carrying a party through the post-SEC-01 principal model; run `Checker.Has` | Local dev or CI, testkit DB, W03-E01's principal model available and `accepted` | The previously-false evaluation is now correctly `true` | party-subject-edge test report | unassigned |

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
