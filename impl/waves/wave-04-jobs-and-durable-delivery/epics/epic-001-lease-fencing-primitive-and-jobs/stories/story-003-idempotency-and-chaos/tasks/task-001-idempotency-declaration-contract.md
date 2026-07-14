---
id: W04-E01-S003-T001
type: task
title: Idempotency-declaration contract
status: done
parent_story: W04-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E01-S002
acceptance_criteria:
  - AC-W04-E01-S003-01
artifacts:
  - ART-W04-E01-S003-001
  - ART-W04-E01-S003-005
evidence:
  - EV-W04-E01-S003-001
---

# W04-E01-S003-T001 — Idempotency-declaration contract

## Task Definition

### Task objective

Establish a stable job idempotency key and lease context passed to every worker, and require each
worker to declare exactly one duplicate-safety mechanism (inbox/effect ledger unique on
`(job_id, effect_name)`, domain CAS, or provider idempotency key) at registration time, rejecting
registration without a declaration.

### Parent story

W04-E01-S003 — Worker idempotency contract and the shared duplicate-worker chaos harness.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E01-S002 (the lease context threaded to workers is S002's fenced claim/finalize/reclaim chain's
own output).

### Detailed work

1. Re-read `kernel/jobs`'s worker-registration mechanism at this task's actual start commit to
   confirm no idempotency-declaration requirement currently exists.
2. Confirm whether PF-ARCH's typed operation model exists in the repository; if not, design the
   contract as a runtime registration-time check (resolves `plan.md`'s "Unresolved questions" item
   on enforcement mechanism).
3. Design and document the stable job idempotency key's derivation.
4. Implement the worker-invocation signature change threading the idempotency key and lease context
   to workers — record this explicitly as T5's confirmed-breaking change (RISK-W04-003), with the
   wowsociety coordination note recorded, not resolved.
5. Implement the registration-time contract: a worker must declare exactly one of the three allowed
   mechanisms; registration fails otherwise.
6. Write a duplicate-effect / registration-rejection test proving: a worker without a declaration
   cannot register; a worker with exactly one declared mechanism registers successfully.
7. Document the contract and the T5 coordination note (this task's share of ART-W04-E01-S003-005).

### Expected files or components affected

`kernel/jobs`'s worker-registration mechanism and worker-invocation signature (exact file paths TBD
per `plan.md`).

### Expected output

A registration-time contract enforcing exactly one declared duplicate-safety mechanism; the stable
idempotency key and lease context threaded to worker invocation.

### Required artifacts

ART-W04-E01-S003-001 (idempotency-declaration contract), ART-W04-E01-S003-005 (documentation,
shared with T002/T003).

### Required evidence

EV-W04-E01-S003-001 (duplicate-effect / registration-rejection test report).

### Related acceptance criteria

AC-W04-E01-S003-01.

### Completion criteria

A worker cannot register without declaring exactly one duplicate-safety mechanism; the idempotency
key and lease context are threaded to worker invocation — proven by a passing test.

### Verification method

Direct execution of the duplicate-effect / registration-rejection test.

### Risks

RISK-W04-003 (the worker-signature change is confirmed breaking, requiring wowsociety coordination)
— see epic-level `risks.md`. This is the task's primary risk, not incidental — the signature change
itself is unavoidable per T5's own acceptance criterion.

### Rollback or recovery considerations

If the registration contract is found to reject a legitimate worker's valid declaration due to a
contract-parsing defect, revert and fix the parsing logic — do not loosen the "exactly one declared
mechanism" requirement to work around a parsing bug.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not yet implemented — the worker-invocation signature change (T5, confirmed breaking) is expected
here once implemented.*

### Configuration changes

*Not yet implemented.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not yet implemented — the idempotency-declaration contract is itself the security control;
recorded here once implemented.*

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

*None anticipated beyond the already-tracked RISK-W04-003 coordination note.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E01-S003-01 | Run duplicate-effect / registration-rejection test | Local dev or CI, Go toolchain | Worker without declaration cannot register; worker with exactly one declared mechanism registers | test report | unassigned |

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
