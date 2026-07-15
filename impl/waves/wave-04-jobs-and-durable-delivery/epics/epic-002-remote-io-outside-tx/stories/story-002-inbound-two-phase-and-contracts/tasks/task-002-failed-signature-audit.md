---
id: W04-E02-S002-T002
type: task
title: Failed-signature audit path
status: done
parent_story: W04-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E02-S002-T001
acceptance_criteria:
  - AC-W04-E02-S002-02
artifacts:
  - ART-W04-E02-S002-002
evidence:
  - EV-W04-E02-S002-002
---

# W04-E02-S002-T002 — Failed-signature audit path

## Task Definition

### Task objective

Implement a body-free audit row write, in its own short transaction, for every failed inbound
signature verification, so that no raw request body is ever persisted on a failed verification.

### Parent story

W04-E02-S002 — Inbound two-phase verification, adapter contracts, and the named 6-boundary chaos
test.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E02-S002-T001 (the failed-signature audit fires from the two-phase verification's failure path;
T5's own source dependency is "T4").

### Detailed work

1. Design the failed-signature-audit row schema: signature, timestamp, endpoint reference, and
   failure reason — explicitly excluding the raw request body.
2. Implement the audit-row write, triggered on every failed verification outcome from T001's
   two-phase protocol, in its own short transaction separate from the verification transactions.
3. Write the empty-body-field test: deliberately trigger a failed verification, inspect the
   resulting audit row, and confirm the body field is empty (or the column/field does not exist at
   all, if the schema design excludes it structurally rather than leaving it nullable).
4. Document the failed-signature audit's body-free guarantee.

### Expected files or components affected

A new failed-signature-audit table/write path (exact location TBD); `HandleInbound`'s failure-path
wiring.

### Expected output

A body-free audit row written on every failed verification, in its own short transaction, proven by
the empty-body-field test.

### Required artifacts

ART-W04-E02-S002-002 (failed-signature audit path).

### Required evidence

EV-W04-E02-S002-002 (empty-body-field test report).

### Related acceptance criteria

AC-W04-E02-S002-02.

### Completion criteria

The empty-body-field test passes for every failed-verification case exercised; the audit row is
confirmed, by direct inspection, to be written in its own transaction, separate from the
verification transactions.

### Verification method

Direct execution of the empty-body-field test; inspection of the audit-write transaction boundary.

### Risks

Low (per PLAN DATA-03 T5's own risk column: "Low") — the primary risk is an implementation error
that accidentally captures the raw body despite the schema/design intent; the empty-body-field test
is the direct mitigation.

### Rollback or recovery considerations

Revert the audit-write path if it introduces a performance regression on the failure path or a
schema issue; because the audit row is additive (a new table/write), a revert should not affect
existing verification behavior.

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

*Not yet implemented — the failed-signature-audit table is planned, not yet created.*

### Security changes

*Not yet implemented — the body-free guarantee is itself a security/compliance control; recorded
here once implemented.*

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
| AC-W04-E02-S002-02 | Run empty-body-field test on a deliberately failed verification | Local dev or CI, Go toolchain | Audit row written in own short tx; body field empty | test report | unassigned |

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
