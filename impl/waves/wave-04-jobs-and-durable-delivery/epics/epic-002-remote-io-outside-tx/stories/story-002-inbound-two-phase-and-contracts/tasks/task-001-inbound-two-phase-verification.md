---
id: W04-E02-S002-T001
type: task
title: Inbound two-phase verification for HandleInbound
status: done
parent_story: W04-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E02-S001-T001
acceptance_criteria:
  - AC-W04-E02-S002-01
artifacts:
  - ART-W04-E02-S002-001
evidence:
  - EV-W04-E02-S002-001
---

# W04-E02-S002-T001 — Inbound two-phase verification for HandleInbound

## Task Definition

### Task objective

Restructure `kernel/webhook.HandleInbound` into a two-phase protocol — short read-tx endpoint
snapshot → verification outside any tx → short write-tx re-check of version/status with
discard+retry on mismatch — so that a secret rotation or deactivation occurring between snapshot and
verification cannot cause an inbound signature to be accepted under a stale policy.

### Parent story

W04-E02-S002 — Inbound two-phase verification, adapter contracts, and the named 6-boundary chaos
test.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E02-S001-T001 (shares the claim-row/lease-column infrastructure where relevant; T4's own source
dependency is "T1").

### Detailed work

1. Re-read `HandleInbound`'s current transaction structure at this task's actual start commit,
   confirming it currently holds a single enclosing transaction across snapshot-read and
   verification.
2. Design the snapshot data shape (secret, version, status, and any other field whose mid-window
   change would invalidate a verification).
3. Implement the read-tx snapshot stage, closing the transaction immediately after the snapshot is
   captured.
4. Implement the out-of-tx verification stage, using the snapshot's secret to verify the inbound
   signature with no open transaction.
5. Implement the write-tx re-check stage: re-read the endpoint's current version/status, compare
   against the snapshot; on match, commit the verification outcome; on mismatch, discard the result
   and retry from step 3, bounded by an explicit, documented retry ceiling.
6. Enumerate and update in-repo callers of `HandleInbound` for the breaking signature change to its
   transaction-ownership contract; record the change explicitly as a compatibility consideration.
7. Write the rotation-during-verification test: deliberately rotate or deactivate the endpoint's
   secret in the snapshot-to-verification window, confirming discard+retry occurs and no
   accept-under-stale-policy is possible.
8. Add observability (log lines) for discard+retry events.
9. Document the two-phase protocol's stage boundaries and its retry-ceiling behavior.

### Expected files or components affected

`webhook/service.go` or its `HandleInbound`-owning file; any in-repo callers of `HandleInbound`
requiring a signature-change update.

### Expected output

A two-phase inbound-verification protocol proven immune to a rotation-during-verification race by a
dedicated test, with the breaking-change note explicitly recorded.

### Required artifacts

ART-W04-E02-S002-001 (inbound two-phase verification implementation).

### Required evidence

EV-W04-E02-S002-001 (rotation-during-verification integration-test report).

### Related acceptance criteria

AC-W04-E02-S002-01.

### Completion criteria

The rotation-during-verification test passes; the retry ceiling is confirmed bounded, not
unbounded, by direct inspection; the breaking signature change is documented as a compatibility
consideration with all in-repo callers updated.

### Verification method

Direct execution of the rotation-during-verification test against a live PostgreSQL instance;
inspection of `HandleInbound`'s new signature and its callers.

### Risks

Breaking signature change to `HandleInbound`'s transaction-ownership contract (per PLAN DATA-03 T4's
own risk column) — must be recorded as an explicit compatibility consideration, not silently
absorbed; bound retry attempts — an unbounded discard+retry loop would itself become a
denial-of-service-adjacent defect, mirroring W02-E01-S001-T002's own bounded-retry security
rationale.

### Rollback or recovery considerations

Revert the two-phase protocol if it produces false-positive discard+retry loops under legitimate,
non-rotation load; because the signature change is breaking, a revert must also revert any
in-repo caller updates made in step 6.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not yet implemented — `HandleInbound`'s breaking signature change is planned, not yet made.*

### Configuration changes

*Not yet implemented.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not yet implemented — the two-phase protocol is itself the security fix; recorded here once
implemented.*

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
| AC-W04-E02-S002-01 | Run rotation-during-verification test | Local dev or CI, PostgreSQL instance | Discard+retry on mismatch, no accept-under-stale-policy, bounded retries | integration-test report | unassigned |

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
