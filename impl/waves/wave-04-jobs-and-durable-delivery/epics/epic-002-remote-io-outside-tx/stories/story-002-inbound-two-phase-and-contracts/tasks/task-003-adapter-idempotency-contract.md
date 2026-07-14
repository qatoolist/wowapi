---
id: W04-E02-S002-T003
type: task
title: Per-adapter idempotency-safety contract declaration
status: done
parent_story: W04-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E02-S001-T002
  - W04-E02-S001-T003
acceptance_criteria:
  - AC-W04-E02-S002-03
artifacts:
  - ART-W04-E02-S002-003
evidence:
  - EV-W04-E02-S002-003
---

# W04-E02-S002-T003 — Per-adapter idempotency-safety contract declaration

## Task Definition

### Task objective

Require every adapter (`Sender` and other high-impact-operation adapters) to declare its
duplicate-safety mechanism at registration time, and enforce at boot time that an adapter cannot be
registered for a non-idempotent high-impact operation without that declaration.

### Parent story

W04-E02-S002 — Inbound two-phase verification, adapter contracts, and the named 6-boundary chaos
test.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E02-S001-T002, W04-E02-S001-T003 (T6's own source dependency is "T2, T3" — the contract is
validated against the concrete `Sender`/webhook-delivery adapters those tasks produce in their
post-three-stage-protocol form).

### Detailed work

1. Inventory all existing `Sender` implementations (and any other adapter registered for a
   high-impact operation) across the framework, per PLAN DATA-03 T6's own risk note: "Inventory all
   existing `Sender` implementations first."
2. For each inventoried adapter, determine its current duplicate-safety posture: does it already use
   an inbox/effect ledger, a domain CAS, or a provider idempotency key, or does it have none.
3. Design the idempotency-safety-declaration contract/interface every adapter must satisfy at
   registration time, requiring it to declare which of the three duplicate-safety mechanisms it
   uses.
4. Implement the declaration mechanism and wire it into the existing adapter-registration flow for
   both `kernel/notify` and `kernel/webhook`.
5. Implement the boot-time enforcement check: reject registration of any adapter for a
   non-idempotent high-impact operation that has not made the declaration.
6. For each inventoried adapter found non-idempotent-and-undeclared (if any), record the finding —
   this task declares the contract and enforces it; it does not itself redesign an adapter's
   internal duplicate-safety logic (see `story.md` "Out of scope"). Record any such finding in
   `deviations.md` or as a follow-up item, not silently absorbed.
7. Write the boot-time fixture test: register a deliberately undeclared adapter for a
   non-idempotent high-impact operation, confirm the boot sequence rejects it with a clear,
   adapter-identifying error.
8. Write a positive fixture test confirming a correctly-declared adapter registers successfully.
9. Document the contract-declaration requirement and how to satisfy it.

### Expected files or components affected

`kernel/notify` and `kernel/webhook`'s existing `Sender`/adapter-registration mechanism; a new
idempotency-safety-declaration contract/interface (exact location TBD).

### Expected output

A boot-time-enforced idempotency-safety contract-declaration mechanism, proven by a boot-time
fixture test, with a complete inventory of all existing `Sender` implementations and their
declared (or newly-declared) duplicate-safety posture.

### Required artifacts

ART-W04-E02-S002-003 (per-adapter idempotency-safety contract declaration mechanism).

### Required evidence

EV-W04-E02-S002-003 (boot-time fixture test report plus `Sender` inventory report).

### Related acceptance criteria

AC-W04-E02-S002-03.

### Completion criteria

The boot-time fixture test rejects an undeclared adapter and accepts a correctly-declared one; the
`Sender` inventory is complete and every existing implementation is correctly declared before
enforcement is enabled.

### Verification method

Direct execution of the boot-time fixture test (positive and negative cases); inspection of the
inventory record for completeness against the actual set of registered adapters in the repository.

### Risks

"Inventory all existing `Sender` implementations first" (PLAN DATA-03 T6's own risk column) — an
incomplete inventory could mean enforcement goes live while a real, non-idempotent, currently-in-
production adapter is silently undeclared and passes only because it was never inventoried; this
task's step 1 is explicitly a hard prerequisite to step 5's enforcement, not a parallel-safe
activity.

### Rollback or recovery considerations

If enforcement (step 5) is found to reject a legitimate adapter due to an inventory gap, revert
enforcement (not the declaration mechanism itself) and complete the inventory before re-enabling.

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

*Not applicable.*

### Security changes

*Not yet implemented — the boot-time enforcement is itself the security control; recorded here once
implemented.*

### Observability changes

*Not yet implemented — boot-time rejection logging is planned, not yet added.*

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
| AC-W04-E02-S002-03 | Run boot-time fixture test (positive and negative); inspect `Sender` inventory completeness | Local dev or CI, Go toolchain | Undeclared adapter rejected with clear error; declared adapter registers; inventory complete | boot-time fixture test report + inventory report | unassigned |

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
