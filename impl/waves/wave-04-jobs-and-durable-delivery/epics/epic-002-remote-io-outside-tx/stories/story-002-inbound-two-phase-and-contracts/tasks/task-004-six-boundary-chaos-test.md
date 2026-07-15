---
id: W04-E02-S002-T004
type: task
title: Named 6-boundary chaos test (notify and webhook)
status: done
parent_story: W04-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E02-S001-T002
  - W04-E02-S001-T003
  - W04-E02-S002-T001
  - W04-E01-S003
acceptance_criteria:
  - AC-W04-E02-S002-04
artifacts:
  - ART-W04-E02-S002-004
evidence:
  - EV-W04-E02-S002-004
  - EV-W04-E02-S002-005
---

# W04-E02-S002-T004 — Named 6-boundary chaos test (notify and webhook)

## Task Definition

### Task objective

Prove, with a named chaos test injecting failure at each of 6 specific boundaries — **before send,
during send, after success/before finalize, lease expiry, duplicate workers, provider timeout** —
applied to both `kernel/notify` and `kernel/webhook`, that zero duplicate external effects occur at
any of the 6 fault points. This task **reuses** the chaos harness built in **W04-E01-S003** (DATA-02
T7, `DATA-02/chaos/duplicate_worker_lease_expiry_test.go`'s underlying reusable harness) — it does
not design or build a new harness.

### Parent story

W04-E02-S002 — Inbound two-phase verification, adapter contracts, and the named 6-boundary chaos
test.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E02-S001-T002 (notify three-stage protocol), W04-E02-S001-T003 (webhook three-stage protocol),
W04-E02-S002-T001 (inbound two-phase verification) — T8's own source dependency column is "T2-T4."
**W04-E01-S003** (the shared chaos harness, built for DATA-02 T7 — "build as a reusable chaos
harness shared with DATA-03/DATA-04") — this task consumes that harness; it does not build its own.

### Detailed work

1. Confirm W04-E01-S003's chaos harness's finalized API/fixture shape (how a fault is injected at a
   named boundary, how "duplicate external effect" is observed/asserted) at this task's actual start
   commit. Do not redesign or reimplement any part of the harness itself.
2. Map each of the 6 named boundaries onto S001's three-stage protocol (for both notify and webhook)
   and this story's T001 (inbound two-phase verification, where applicable to the boundary):
   - **before send** — fault injected before the effect stage begins.
   - **during send** — fault injected mid-effect-stage (e.g. the remote call is interrupted).
   - **after success/before finalize** — the remote call succeeds but the finalize-tx has not yet
     committed.
   - **lease expiry** — the claim lease expires before finalize.
   - **duplicate workers** — two workers race to claim/finalize the same delivery.
   - **provider timeout** — the remote call times out without a definitive success/failure signal.
3. Implement the chaos-test suite for `kernel/notify`, exercising all 6 boundaries via the harness,
   asserting zero duplicate `sender.Send` effects at each.
4. Implement the same chaos-test suite for `kernel/webhook`, exercising all 6 boundaries,
   asserting zero duplicate DNS/secret-resolve/POST effects at each.
5. Confirm both suites produce a fail-first result against a version of the code lacking this
   epic's fixes (if feasible to construct such a baseline), or otherwise document why fail-first
   construction was not practical for this specific test.
6. Record the test output as evidence at `DATA-03/chaos/`.

### Expected files or components affected

New chaos-test files for notify and webhook (exact paths TBD, expected under `DATA-03/chaos/`);
no changes to W04-E01-S003's harness itself.

### Expected output

A passing 6-boundary chaos test for both notify and webhook, with zero duplicate external effects
observed at all 6 named boundaries, reusing W04-E01-S003's harness.

### Required artifacts

ART-W04-E02-S002-004 (6-boundary chaos-test suite for notify and webhook).

### Required evidence

EV-W04-E02-S002-004 (notify chaos-test report), EV-W04-E02-S002-005 (webhook chaos-test report).

### Related acceptance criteria

AC-W04-E02-S002-04.

### Completion criteria

Both notify's and webhook's chaos-test suites pass, asserting zero duplicate external effects at
each of the 6 named boundaries, using W04-E01-S003's harness without modification to the harness
itself.

### Verification method

Direct execution of both chaos-test suites; inspection confirming the harness used is
W04-E01-S003's own (import/dependency check), not a parallel reimplementation.

### Risks

"Most labor-intensive requirement in PF-DATA" (PLAN DATA-03 T8's own risk column) — this task's own
scope is deliberately narrowed to reusing an existing harness rather than building one, which is the
primary mitigation for that labor intensity; the remaining labor is mapping the 6 boundaries
correctly onto two distinct protocols (notify, webhook) and confirming correct behavior at each.

### Rollback or recovery considerations

Not applicable in the usual code-rollback sense — a failing chaos test blocks this story's
acceptance until the underlying protocol defect (in S001 or this story's T001) it reveals is fixed;
it does not itself get "rolled back."

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not applicable — this task consumes W04-E01-S003's harness API; it does not introduce a new one.*

### Configuration changes

*Not yet implemented.*

### Schema or migration changes

*Not applicable.*

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
| AC-W04-E02-S002-04 | Run 6-boundary chaos test for notify | Local dev or CI, PostgreSQL instance, W04-E01-S003 chaos harness | Zero duplicate `sender.Send` effects across all 6 boundaries | chaos-test report | unassigned |
| AC-W04-E02-S002-04 | Run 6-boundary chaos test for webhook | Local dev or CI, PostgreSQL instance, W04-E01-S003 chaos harness | Zero duplicate DNS/secret-resolve/POST effects across all 6 boundaries | chaos-test report | unassigned |

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
