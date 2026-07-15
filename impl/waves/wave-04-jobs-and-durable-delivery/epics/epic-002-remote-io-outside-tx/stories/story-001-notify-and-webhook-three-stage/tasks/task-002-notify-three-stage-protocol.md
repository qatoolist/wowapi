---
id: W04-E02-S001-T002
type: task
title: Three-stage protocol for kernel/notify
status: done
parent_story: W04-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E02-S001-T001
acceptance_criteria:
  - AC-W04-E02-S001-02
artifacts:
  - ART-W04-E02-S001-002
  - ART-W04-E02-S001-004
evidence:
  - EV-W04-E02-S001-002
---

# W04-E02-S001-T002 — Three-stage protocol for kernel/notify

## Task Definition

### Task objective

Implement the three-stage claim-tx (assigns lease) → `sender.Send` outside any tx (delivery ID as
idempotency key) → finalize-tx (comparing lease token) protocol for `kernel/notify`, and delete or
update the self-documented "should move outside tx" comment at `notify/service.go:456-586`
(446-449) as part of the same change.

### Parent story

W04-E02-S001 — Notify and webhook three-stage remote-I/O protocol.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E02-S001-T001 (the lease-column migration must exist before the claim-tx stage can assign a
lease against it).

### Detailed work

1. Re-read `notify/service.go:446-586` at this task's actual start commit, confirming the current
   claim-and-send-inside-tx pattern and the self-documented comment (446-449) still exist.
2. Restructure the claim logic into a short claim-tx that assigns a lease via T001's migrated
   columns and W04-E01's shared primitive's claim API, then commits.
3. Move `sender.Send` entirely outside any open transaction, using the delivery ID as the effect
   stage's idempotency key.
4. Implement the finalize-tx stage, comparing the lease token against the current row's lease state
   before committing the outcome (fencing check against a reclaimed worker's stale finalize).
5. Delete or update the self-documented comment at `notify/service.go:446-449` to reflect the
   now-implemented protocol.
6. Write a test asserting no `sender.Send` call executes while a database transaction is open.
7. Add observability (log lines) for claim/effect/finalize stage transitions.
8. Document the three-stage protocol's stage boundaries for `kernel/notify`.

### Expected files or components affected

`notify/service.go` (claim/send/finalize logic around lines 446-586, including the comment at
446-449).

### Expected output

A three-stage protocol implementation for `kernel/notify` proving no `sender.Send` call executes
while a tx is open, with the self-documented comment resolved.

### Required artifacts

ART-W04-E02-S001-002 (notify three-stage protocol implementation), ART-W04-E02-S001-004
(documentation, shared with T003).

### Required evidence

EV-W04-E02-S001-002 (no-send-while-tx-open assertion test report).

### Related acceptance criteria

AC-W04-E02-S001-02.

### Completion criteria

The three-stage protocol is implemented; the no-send-while-tx-open assertion test passes; the
self-documented comment at `notify/service.go:446-449` is deleted or updated to reflect the
implemented protocol, confirmed by direct inspection.

### Verification method

Direct execution of the no-send-while-tx-open assertion test; direct inspection of
`notify/service.go:446-449` confirming the comment no longer describes an unresolved TODO.

### Risks

Deleting/updating the self-documented comment as part of this task (per PLAN DATA-03 T2's own risk
column) — if the comment is deleted without the protocol genuinely closing the gap it described,
this would silently regress from "documented known gap" to "undocumented known gap," which is
strictly worse. The independent-review task (T004) specifically re-checks this.

### Rollback or recovery considerations

Revert the three-stage restructuring if it destabilizes notify delivery throughput; the deleted
comment should be restorable from version control if a revert is needed before the gap is actually
closed.

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

*Not applicable — this task consumes T001's migration; it does not itself add a schema change.*

### Security changes

*Not applicable.*

### Observability changes

*Not yet implemented — stage-transition logging is planned, not yet added.*

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
| AC-W04-E02-S001-02 | Run no-send-while-tx-open assertion test; inspect comment at `notify/service.go:446-449` | Local dev or CI, Go toolchain | No `sender.Send` call while tx open; comment resolved, not a TODO | test report + code-inspection report | unassigned |

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
