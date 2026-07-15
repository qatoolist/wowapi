---
id: W02-E03-S001-T004
type: task
title: Scheduled GC sweep
status: todo
parent_story: W02-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W02-E03-S001-T002
  - W02-E03-S001-T003
acceptance_criteria:
  - AC-W02-E03-S001-04
artifacts:
  - ART-W02-E03-S001-004
  - ART-W02-E03-S001-005
evidence:
  - EV-W02-E03-S001-004
---

# W02-E03-S001-T004 — Scheduled GC sweep

## Task Definition

### Task objective

Implement a scheduled GC sweep that removes expired, unreferenced upload objects with metrics and
audit, conservative enough to never remove a referenced object.

### Parent story

W02-E03-S001 — Version-allocation races and upload-blob GC.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E03-S001-T002, W02-E03-S001-T003 (per PLAN DATA-05 T4's own Depends-on column: "T2, T3" — the
sweep needs T2's session records to exist and T3's confirmation path to correctly mark sessions
confirmed before it can safely distinguish a reclaimable object from a referenced one).

### Detailed work

1. Design the GC sweep's scheduling mechanism (cron-style scheduled job, periodic background worker,
   or manually-triggered operational command) — to be selected and documented at this task's
   implementation time (resolves `plan.md`'s "Unresolved questions" item on scheduling mechanism).
2. Design and select the sweep's grace window value — a conservative figure past a session's expiry
   before its object becomes eligible for removal, per PLAN T4's own risk note: "False-positive
   deletion is data loss — conservative grace window." Document the chosen value and rationale.
3. Implement the sweep: read session records for past-expiry, unconfirmed sessions within the grace
   window; remove the corresponding storage object; never act on a confirmed or still-pending
   session's object.
4. Emit metrics and an audit record for each sweep action (removal or explicit no-op), per PLAN T4's
   own acceptance framing.
5. Write the mixed-state test: a set of sessions in confirmed, expired-unconfirmed, and still-pending
   states in one test run, confirming the sweep removes only the expired-unconfirmed set's objects
   and never touches a confirmed or still-pending session's object.
6. Document the GC sweep's grace window, scheduling mechanism, and audit/metrics output.

### Expected files or components affected

A new GC sweep mechanism (scheduled job or command, exact location TBD); `kernel/document`'s session
and storage-object read paths (read-only from this task's perspective, no schema change beyond what
T002 already introduced).

### Expected output

A scheduled GC sweep proven, by the mixed-state test, to never remove a referenced object and to
remove every past-expiry unconfirmed session's object.

### Required artifacts

ART-W02-E03-S001-004 (the GC sweep mechanism), ART-W02-E03-S001-005 (documentation, shared with
T001/T002).

### Required evidence

EV-W02-E03-S001-004 (mixed confirmed/expired/pending GC sweep test output).

### Related acceptance criteria

AC-W02-E03-S001-04.

### Completion criteria

The mixed-state test passes: the sweep removes every past-expiry unconfirmed session's object and
never removes a confirmed or still-pending session's object.

### Verification method

Direct execution of the mixed-state test against a live PostgreSQL instance and object storage (or a
fake/mock storage backend), inspecting the sweep's actual removal decisions and its emitted
metrics/audit records.

### Risks

PLAN T4's own risk note states plainly: "False-positive deletion is data loss — conservative grace
window." This is the single highest-consequence failure mode in this story — an incorrect sweep could
destroy a referenced object with no recovery path. The mixed-state test and a deliberately
conservative grace window are the required controls, not optional hardening.

### Rollback or recovery considerations

Disable the GC sweep immediately (without reverting T001–T003's session-durability and confirmation
mechanisms) if any false-positive deletion is ever observed in any environment; escalate for redesign
of the sweep's eligibility logic before re-enabling, per this story's own `plan.md` "Rollback
strategy."

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

*Not yet implemented — expected: the sweep's grace window and scheduling interval, as constants or
configuration keys (see `plan.md` "Unresolved questions").*

### Schema or migration changes

*Not applicable — this task reads existing session records from T002's table; it introduces no new
schema of its own.*

### Security changes

*Not applicable.*

### Observability changes

*Not yet implemented — expected: metrics and an audit record for each sweep action.*

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
| AC-W02-E03-S001-04 | Run the mixed confirmed/expired/pending GC sweep test | Local dev or CI, PostgreSQL instance + object storage (or fake) | Sweep removes only past-expiry unconfirmed sessions' objects; never removes a referenced object | integration-test report | unassigned |

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
