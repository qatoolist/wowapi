---
id: W04-E02-S001-T003
type: task
title: Three-stage protocol for kernel/webhook.deliverToEndpoint
status: done
parent_story: W04-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E02-S001-T001
acceptance_criteria:
  - AC-W04-E02-S001-03
artifacts:
  - ART-W04-E02-S001-003
  - ART-W04-E02-S001-004
evidence:
  - EV-W04-E02-S001-003
---

# W04-E02-S001-T003 — Three-stage protocol for kernel/webhook.deliverToEndpoint

## Task Definition

### Task objective

Implement the same three-stage claim-tx → effect-outside-tx → finalize-tx protocol for
`kernel/webhook.deliverToEndpoint`, moving the current-row-state check into the claim stage so the
effect stage (DNS resolution, secret resolution, POST) requires no mid-flight database read.

### Parent story

W04-E02-S001 — Notify and webhook three-stage remote-I/O protocol.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E02-S001-T001 (the lease-column migration must exist before the claim-tx stage can assign a
lease against it). Parallel-safe with T002 — disjoint code surface (`kernel/webhook` vs
`kernel/notify`); both may proceed independently once T001 lands.

### Detailed work

1. Re-read `webhook/service.go`'s `deliverToEndpoint` function and its delivery loop / secret
   resolution at this task's actual start commit, confirming both currently run inside
   `plat.WithTenant(...)`.
2. Restructure the claim logic into a short claim-tx that assigns a lease via T001's migrated
   columns and W04-E01's shared primitive's claim API, then commits — and relocate the current-row-
   state check (confirming the endpoint/delivery row's state before proceeding) into this claim
   stage.
3. Move DNS resolution, secret resolution, and the POST entirely outside any open transaction, using
   the delivery ID as the effect stage's idempotency key.
4. Implement the finalize-tx stage, comparing the lease token against the current row's lease state
   before committing the outcome.
5. Write a test asserting no DNS resolution, secret resolution, or POST call executes while a
   database transaction is open, and confirming no mid-flight DB read occurs during the effect
   stage.
6. Add observability (log lines) for claim/effect/finalize stage transitions.
7. Document the three-stage protocol's stage boundaries for `kernel/webhook`.

### Expected files or components affected

`webhook/service.go` (`deliverToEndpoint` and its delivery loop / secret resolution).

### Expected output

A three-stage protocol implementation for `kernel/webhook.deliverToEndpoint` proving no
DNS/secret-resolve/POST call executes while a tx is open, and that the current-row-state check
occurs only in the claim stage.

### Required artifacts

ART-W04-E02-S001-003 (webhook three-stage protocol implementation), ART-W04-E02-S001-004
(documentation, shared with T002).

### Required evidence

EV-W04-E02-S001-003 (no-network-call-while-tx-open assertion test report).

### Related acceptance criteria

AC-W04-E02-S001-03.

### Completion criteria

The three-stage protocol is implemented; the no-network-call-while-tx-open assertion test passes;
the current-row-state check is confirmed, by direct inspection, to occur only in the claim stage,
with no mid-flight DB read during the effect stage.

### Verification method

Direct execution of the no-network-call-while-tx-open assertion test; direct inspection of the
effect-stage code confirming no DB read occurs mid-flight.

### Risks

The current-row-state check must move into the claim stage so `Execute` (the effect stage) needs no
mid-flight DB reads (per PLAN DATA-03 T3's own risk column) — an incorrectly-placed check (left in
the effect stage) would silently reintroduce a DB dependency inside what is meant to be a
tx-free effect stage, defeating this task's own purpose. The independent-review task (T004)
specifically re-checks this placement.

### Rollback or recovery considerations

Revert the three-stage restructuring if it destabilizes webhook delivery throughput or breaks
endpoint-state consistency; because the current-row-state check moves to claim stage, a revert must
restore the original inline check location, not merely delete the new claim-stage check.

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
| AC-W04-E02-S001-03 | Run no-network-call-while-tx-open assertion test; inspect effect-stage code for mid-flight DB reads | Local dev or CI, Go toolchain | No DNS/secret-resolve/POST call while tx open; current-row-state check confirmed in claim stage only | test report + code-inspection report | unassigned |

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
