---
id: W04-E01-S003-T002
type: task
title: Effect-ledger-vs-fencing test
status: done
parent_story: W04-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E01-S003-T001
acceptance_criteria:
  - AC-W04-E01-S003-02
artifacts:
  - ART-W04-E01-S003-002
  - ART-W04-E01-S003-005
evidence:
  - EV-W04-E01-S003-002
---

# W04-E01-S003-T002 — Effect-ledger-vs-fencing test

## Task Definition

### Task objective

Write a testable proof — not prose documentation — that fencing the `jobs_queue` row alone (S002's
own mechanism) does not undo an already-committed stale-worker domain transaction, and that the
effect ledger (not queue-row fencing) is what catches an idempotency-ignoring worker.

### Parent story

W04-E01-S003 — Worker idempotency contract and the shared duplicate-worker chaos harness.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E01-S003-T001 (the effect-ledger test needs the idempotency contract's key/lease-context
threading to construct a realistic worker scenario).

### Detailed work

1. Re-read `kernel/outbox/relay.go`'s existing inbox-dedup logic (per MATRIX CS-11's own evidence
   refinement: "DB effects are already exactly-once via the outbox inbox-dedup
   (`kernel/outbox/relay.go:191,205-219`)") at this task's actual start commit to confirm the current
   mechanism.
2. Construct a scenario where S002's queue-row fencing succeeds (the stale worker's finalize is
   correctly rejected) but the stale worker's domain transaction has already committed prior to
   fencing taking effect — this is the scenario T6's own acceptance criterion targets.
3. Write a test proving: in this scenario, the effect ledger — not the queue-row fencing — is what
   catches an idempotency-ignoring worker's duplicate effect (a worker that did not correctly
   declare/use one of the three duplicate-safety mechanisms).
4. Write a companion assertion (or a second test) proving: an idempotency-compliant worker's domain
   transaction, even if committed under a soon-to-be-fenced lease, is correctly recognized as
   already-applied by its own declared mechanism, not reprocessed.
5. Document the fencing/effect-ledger distinction (this task's share of ART-W04-E01-S003-005).

### Expected files or components affected

A new integration test exercising `kernel/jobs`'s fencing alongside `kernel/outbox/relay.go`'s (or
the relevant worker's own declared mechanism's) effect-ledger logic (exact file path TBD per
`plan.md`).

### Expected output

A passing test proving fencing alone does not undo a committed stale-worker domain transaction, and
that the effect ledger is the actual source of truth catching an idempotency-ignoring worker.

### Required artifacts

ART-W04-E01-S003-002 (the effect-ledger-vs-fencing test), ART-W04-E01-S003-005 (documentation,
shared with T001/T003).

### Required evidence

EV-W04-E01-S003-002 (effect-ledger-vs-fencing integration-test report).

### Related acceptance criteria

AC-W04-E01-S003-02.

### Completion criteria

The test proves, as a testable claim (not prose), that fencing the queue row alone does not undo an
already-committed stale-worker domain transaction, and that the effect ledger catches an
idempotency-ignoring worker.

### Verification method

Direct execution of the effect-ledger-vs-fencing test.

### Risks

Low, per PLAN DATA-02 T6's own risk column ("Low"). The primary risk is constructing a scenario that
does not actually isolate the fencing-vs-effect-ledger distinction (e.g. accidentally relying on
fencing to catch the duplicate, which would falsify the test's own premise).

### Rollback or recovery considerations

Not applicable — a test-only task with no production code change beyond the test fixture itself.

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
| AC-W04-E01-S003-02 | Run the effect-ledger-vs-fencing test | Local dev or CI, PostgreSQL instance | Fencing alone does not undo a committed stale-worker transaction; effect ledger catches the idempotency-ignoring worker | integration-test report | unassigned |

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
