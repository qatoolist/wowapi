---
id: W04-E01-S003-T003
type: task
title: Chaos harness and named chaos test
status: done
parent_story: W04-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E01-S003-T001
acceptance_criteria:
  - AC-W04-E01-S003-03
artifacts:
  - ART-W04-E01-S003-003
  - ART-W04-E01-S003-005
evidence:
  - EV-W04-E01-S003-003
---

# W04-E01-S003-T003 — Chaos harness and named chaos test

## Task Definition

### Task objective

Build the named chaos test `DATA-02/chaos/duplicate_worker_lease_expiry_test.go`, exercising all
three named boundaries (domain, external, finalize) exactly as PLAN DATA-02 T7 states — "pause
worker A after claim, expire, reclaim via B, B completes, resume A and attempt finalize at every
domain/external/finalize boundary — exactly one logical effect recorded, A's writes rejected" — and
build it as a **reusable chaos harness explicitly shared with W04-E02 (DATA-03) and W04-E03
(DATA-04)**, not reimplemented by either.

### Parent story

W04-E01-S003 — Worker idempotency contract and the shared duplicate-worker chaos harness.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E01-S003-T001 (the chaos test invokes workers using the idempotency key/lease-context shape
T001 establishes).

### Detailed work

1. Design the chaos harness's pause-after-claim mechanism (test hook, synchronization primitive, or
   controlled delay) as reusable, parameterizable test infrastructure — not inline, single-test-only
   logic (resolves `plan.md`'s "Unresolved questions" item on the pause mechanism).
2. Implement the harness: pause worker A after claim; expire A's lease; reclaim via worker B; B
   completes; resume A; A attempts to finalize.
3. Implement the named test file `DATA-02/chaos/duplicate_worker_lease_expiry_test.go`
   (or its repository-actual equivalent path, documented if it diverges — see `plan.md`), exercising
   the harness at all three named boundaries: domain (A's domain transaction attempt after resume),
   external (A's external-effect attempt after resume, if applicable to jobs specifically), and
   finalize (A's finalize attempt, consuming S002's fenced finalize path).
4. Confirm the test proves: exactly one logical effect is recorded (B's, not A's duplicate); A's
   writes are rejected at every one of the three boundaries, not a subset.
5. Structure and document the harness's public surface so W04-E02's 6-boundary chaos test and
   W04-E03's chaos test can parameterize it for their own effect types (notify/webhook,
   bulk-operation effects) without reimplementing the pause/expire/reclaim/resume mechanics.
6. Document the harness's reuse contract (this task's share of ART-W04-E01-S003-005).

### Expected files or components affected

A new chaos-harness package and the named test file (exact locations TBD per `plan.md`, expected to
structurally correspond to the source's own `DATA-02/chaos/` path notation).

### Expected output

A passing named chaos test proving exactly one logical effect recorded and worker A's writes
rejected at all three named boundaries; a documented, reusable harness consumable by W04-E02 and
W04-E03.

### Required artifacts

ART-W04-E01-S003-003 (chaos test + harness), ART-W04-E01-S003-005 (documentation, shared with
T001/T002).

### Required evidence

EV-W04-E01-S003-003 (named chaos-test report, all three boundaries).

### Related acceptance criteria

AC-W04-E01-S003-03.

### Completion criteria

The named chaos test exists at the required path, passes, exercises all three named boundaries, and
the harness is structured and documented for W04-E02/W04-E03's own reuse without reimplementation.

### Verification method

Direct execution of the named chaos test; inspection confirming the harness's public surface is
documented and parameterizable, not hardcoded to jobs-only effect types.

### Risks

"Must exercise all 3 named boundaries" per PLAN DATA-02 T7's own risk note — a chaos test that
exercises only a subset (e.g. finalize only, skipping domain/external) would fail to prove the
epic's own closure contract, even if it passes. The harness's reusability is a second, explicit
risk — a harness too tightly coupled to jobs-specific effect types would force W04-E02/W04-E03 to
reimplement rather than reuse, defeating this task's second stated purpose.

### Rollback or recovery considerations

If the harness is found too tightly coupled to jobs-specific effect types once W04-E02 or W04-E03
attempts to consume it, treat this as a harness-extension task (recorded in `deviations.md` at
whichever epic discovers the gap, with a cross-reference back to this task) — not a silent
reimplementation that defeats the shared-harness intent.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not yet implemented — the harness's public reuse-facing surface is expected here once
implemented.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not yet implemented — the chaos test itself requires observability hooks at the domain and
external boundaries in addition to S002's existing finalize-boundary observability.*

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
| AC-W04-E01-S003-03 | Run the named chaos test `DATA-02/chaos/duplicate_worker_lease_expiry_test.go` | Local dev or CI, PostgreSQL instance, multi-goroutine/multi-process test harness | Exactly one logical effect recorded; A's writes rejected at all 3 named boundaries | chaos-test report | unassigned |

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
