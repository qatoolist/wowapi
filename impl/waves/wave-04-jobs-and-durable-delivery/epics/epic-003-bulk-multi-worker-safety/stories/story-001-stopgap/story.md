---
id: W04-E03-S001
type: story
title: Bulk multi-worker stopgap — correct false safety claim, enforce single-processor
status: accepted
wave: W04
epic: W04-E03
owner: W04BulkSafety
reviewer: W04BulkSafety
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DATA-04
depends_on: []
blocks:
  - W04-E03-S002
acceptance_criteria:
  - AC-W04-E03-S001-01
  - AC-W04-E03-S001-02
artifacts: []
evidence: []
decisions: []
risks: []
---

# W04-E03-S001 — Bulk multi-worker stopgap — correct false safety claim, enforce single-processor

## Story ID

W04-E03-S001

## Title

Bulk multi-worker stopgap — correct false safety claim, enforce single-processor

## Objective

Correct the false "safe across replicas" claim in migration `00016`'s header comment, and enforce
single-processor execution via an advisory lock or CAS at the `Service` API boundary, so that a
second concurrent processor attempting to process the same `bulkID` is mechanically rejected rather
than silently racing against the first.

## Value to the framework

This story closes the false-documentation P0 sub-issue independently and fast, before the full
leased-claim rewrite (`W04-E03-S002`) lands. Per the source's own framing, DATA-04 is "P1; P0
before advertising multi-worker" — the immediate risk this story addresses is not that multi-worker
processing is unsafe (it was never actually claimed by the code itself: `Service.next`'s own doc
comment already concedes "no lock — single processor per operation"), but that the migration's own
header comment tells operators and future maintainers something false about the system's safety
properties. A false safety claim in a migration file is itself a hazard — an operator reading that
comment could reasonably conclude it is safe to run multiple bulk processors today, and be wrong.
This story ships independently and fast — closing that false-documentation P0 sub-issue before the
full rewrite — per the source's own task-row framing: "Ships independently and fast — closes the
false-documentation P0 sub-issue before the full rewrite."

## Problem statement

The source's own evidence, reproduced verbatim from the epic's parent finding: "migration `00016`'s
header claims 'safe across replicas' via `FOR UPDATE SKIP LOCKED`; `Service.next`
(`kernel/bulk/bulk.go:123-144`) actually does a plain unlocked `SELECT ... LIMIT 1`, with the
function's own doc comment conceding 'no lock — single processor per operation.'" This is a direct
contradiction within the codebase itself: the migration comment and the function's own doc comment
disagree about what the code actually does, and the migration comment is the one that is false. PLAN
DATA-04 T1's own row: "Immediate stopgap: correct the false migration comment; enforce
single-processor via advisory lock or CAS at the `Service` API boundary | — | False 'replica-safe'
claim removed; a second concurrent processor is rejected, not silently racing | Concurrency test: 2
processors on the same `bulkID` | `DATA-04/stopgap/` | Ships independently and fast — closes the
false-documentation P0 sub-issue before the full rewrite."

## Source requirements

DATA-04 (T1).

## Current-state assessment

Per the source's own evidence (to be re-confirmed at this story's own execution commit): migration
`00016`'s header comment claims cross-replica safety via `FOR UPDATE SKIP LOCKED`; `Service.next`
(`kernel/bulk/bulk.go:123-144`) does not use `FOR UPDATE SKIP LOCKED` at all — it issues a plain
unlocked `SELECT ... LIMIT 1`. `Service.next`'s own doc comment already states "no lock — single
processor per operation," which is the honest description of the code's actual current behavior.
There is today no mechanical enforcement preventing a second processor from calling `Service.next`
concurrently against the same `bulkID` — the "single processor per operation" property is
documentation-only, not enforced. This story's own re-confirmation step is to read migration
`00016`'s header and `kernel/bulk/bulk.go`'s `Service.next` function (lines 123-144 per the source's
citation) at this story's actual start commit and confirm both the false claim and the absence of
enforcement still hold before making the fix.

## Desired state

Migration `00016`'s header comment no longer claims cross-replica safety it does not provide; it
either states the code's actual current property ("single processor per operation, mechanically
enforced") or is corrected in place once the stopgap lands. A second processor attempting to call
`Service.next` (or any equivalent claim path) against a `bulkID` already being processed is rejected
by an advisory lock or a CAS check at the `Service` API boundary — not silently allowed to race
against the first processor's unlocked `SELECT`.

## Scope

- Correcting migration `00016`'s header comment to remove the false "safe across replicas" claim.
- Implementing single-processor enforcement at the `Service` API boundary via either a PostgreSQL
  advisory lock keyed on `bulkID`, or a CAS (compare-and-swap) check against a processing-owner
  column — the exact mechanism choice is an implementation-time decision, per `plan.md`'s
  "Unresolved questions."
- A concurrency test proving a second processor attempting to claim the same `bulkID` while the
  first is active is rejected.

## Out of scope

- **The full leased-claim rewrite** — lease columns via the shared primitive, the atomic `SKIP
  LOCKED` claim SQL, item idempotency/finalize fencing, retry policy, cancellation, and
  pause/resume/cancel lifecycle controls — all `W04-E03-S002`'s scope (DATA-04 T2–T6). This story
  is explicitly the fast, interim fix, not the final architecture.
- **Any dependency on `W04-E01`'s shared lease/fencing primitive** — this story is designed and
  scoped to require nothing from `W04-E01`, per the source's own dependency column ("—") and per
  `wave-allocation-detail.md`'s framing ("S001 T1 stopgap (can start at wave entry)").
- **Multi-worker throughput or batch-size scaling** — this story enforces exclusivity (single
  processor), it does not add any multi-worker capability; that arrives only once `W04-E03-S002`'s
  leased-claim mechanism lands.

## Assumptions

- The exact mechanism (advisory lock vs. CAS) is not specified by the source beyond "advisory lock
  or CAS at the `Service` API boundary" — this story's plan records the exact choice as an
  implementation-time decision, per mandate §18, rather than inventing a specific mechanism here
  without evidence the source prescribes one over the other.
- This story's fix is confirmed, not assumed, to be superseded by `W04-E03-S002`'s T2 lease-column
  mechanism once that lands — see this epic's `risks.md` (RISK-W04-E03-001) for the sequencing risk
  this creates and its required mitigation (an explicit supersession step recorded in S002's
  `plan.md`).
- Migration `00016` is assumed, pending this story's own re-confirmation step, to be an existing,
  already-applied migration in the repository — this story corrects its header comment in place (or
  via a follow-up migration/comment update, depending on the repository's own convention for
  amending historical migration comments, to be determined at implementation time) rather than
  reverting or replacing the migration itself.

## Dependencies

None within `W04-E03` or upstream — this story has no dependency on `W04-E01`'s shared primitive
and may start at wave entry. Depends on W00's exit gate at wave scope (this wave's own entry
ordering). Blocks `W04-E03-S002` only in the interim sense described in this epic's
`dependencies.md`: S002's T2 depends on `W04-E01-S001` primarily, with this story's stopgap serving
as the bridge mechanism until T2 lands, not as a hard completion gate on S002 starting its own
design work.

## Affected packages or components

`kernel/bulk` (specifically `kernel/bulk/bulk.go`'s `Service.next` and the `Service` API boundary
generally); the migration `00016` header comment (exact file path to be confirmed at this story's
own start-commit re-confirmation step, expected under the repository's migrations directory).

## Compatibility considerations

The single-processor enforcement mechanism (advisory lock or CAS) must not change `Service.next`'s
existing single-processor-caller behavior — a caller already respecting the "single processor per
operation" contract should observe no behavioral change; only a second, concurrent caller against
the same `bulkID` is newly rejected where it was previously silently allowed to race.

## Security considerations

None beyond the correctness property itself — rejecting a second concurrent processor is a
data-integrity control (preventing a race that could corrupt or duplicate bulk-item processing), not
a security-boundary control in the authn/authz sense.

## Performance considerations

An advisory lock or CAS check at the `Service` API boundary adds a small, bounded per-call overhead
to `Service.next` (or its equivalent claim path); this is expected to be negligible relative to the
cost of processing a bulk item and is not separately budgeted by the source beyond the correctness
requirement itself.

## Observability considerations

A rejected second-processor attempt should be observable (logged, at minimum, with the `bulkID` and
the rejecting mechanism) so an operator can distinguish "a second processor was correctly rejected"
from "the first processor silently failed to advance" — a reasonable implementation-time addition
given the stopgap's own purpose, though not separately mandated by the source beyond the rejection
behavior itself.

## Migration considerations

This story's own migration-adjacent change is limited to correcting migration `00016`'s header
comment; it does not add, alter, or drop any database column or table (that arrives only with
`W04-E03-S002`'s T2 lease-column work). If the advisory-lock/CAS mechanism requires a new column
(for the CAS path) or none at all (for the advisory-lock path), that determination is made at
implementation time per `plan.md`'s "Unresolved questions" — an advisory lock requires no schema
change; a CAS check would require a processing-owner column, which, if chosen, is itself a small,
additive migration scoped to this story.

## Documentation requirements

Document the corrected migration `00016` header comment's actual claim (single-processor enforced
via [mechanism], not cross-replica safety); document the chosen enforcement mechanism (advisory
lock or CAS) and its rejection behavior, so a future reader of `kernel/bulk` understands the current,
honest safety property without needing to re-derive it from the code.

## Acceptance criteria

- **AC-W04-E03-S001-01**: Migration `00016`'s header comment no longer claims "safe across
  replicas" via `FOR UPDATE SKIP LOCKED`; the comment is corrected to state the code's actual
  single-processor-enforced property.
- **AC-W04-E03-S001-02**: A second processor attempting to process the same `bulkID` while a first
  processor is active is rejected (not silently allowed to race), proven by a concurrency test with
  2 processors on the same `bulkID`.

## Required artifacts

- The corrected migration `00016` header comment.
- The advisory-lock or CAS single-processor enforcement mechanism (code).
- Documentation of the corrected claim and the enforcement mechanism.
See `artifacts/index.md`.

## Required evidence

- 2-processor concurrency test output (`DATA-04/stopgap/`), proving the second processor is
  rejected.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependencies (none) recorded,
owner/reviewer assignment pending, the mechanism choice (advisory lock vs. CAS) explicitly recorded
as an implementation-time decision rather than silently assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; both acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; the corrected
migration comment and the rejection mechanism are confirmed by review to genuinely match the source's
acceptance bar ("False 'replica-safe' claim removed; a second concurrent processor is rejected, not
silently racing"), not merely partially addressed.

## Risks

None epic-specific beyond RISK-W04-E03-001 (this epic's `risks.md`), which concerns the sequencing
between this story's stopgap and `W04-E03-S002`'s T2 lease-column mechanism — this story's own scope
is the first half of that sequencing risk's mitigation (a clean, clearly-scoped stopgap that S002 can
explicitly supersede, not a mechanism that lingers ambiguously alongside S002's later fix).

## Residual-risk expectations

Low, once the concurrency test (AC-W04-E03-S001-02) passes and the migration comment correction
(AC-W04-E03-S001-01) is confirmed — this is a small, well-bounded, fast-track fix with a clear,
source-derived acceptance bar and no dependency on unfinished upstream work.

## Plan

See `plan.md`.
