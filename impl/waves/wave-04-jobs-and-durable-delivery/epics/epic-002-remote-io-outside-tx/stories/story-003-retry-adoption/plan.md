---
id: PLAN-W04-E02-S003
type: plan
parent_story: W04-E02-S003
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Plan — W04-E02-S003

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below. This plan does not invent the exact file/line locations of the two hand-rolled
retry implementations, since the source text (REVIEW §K/§O) names the duplication pattern without
citing exact locations — locating them is this story's own first implementation step.

## Proposed architecture

A single shared retry mechanism (`cenkalti/backoff/v5`) replacing two independently-maintained
hand-rolled retry loops. No new abstraction layer is proposed beyond what the library itself
provides — this is a direct library-adoption swap, not a new internal retry-abstraction design,
consistent with the story's own small, well-bounded scope.

## Implementation strategy

1. Locate both hand-rolled retry implementations in the codebase (this story's own confirmed-open
   discovery step — see "Unresolved questions").
2. For each, document its current retry schedule (attempt count, backoff timing/growth, jitter if
   any, terminal behavior) as the parity baseline.
3. Add `cenkalti/backoff/v5` as a direct `go.mod` dependency (already transitively present).
4. Replace the first hand-rolled implementation with a `cenkalti/backoff/v5`-configured equivalent,
   matching (or documented-ly improving) its retry schedule.
5. Replace the second hand-rolled implementation the same way.
6. Write the retry-schedule-parity test, comparing each new configuration's observed behavior
   against its documented baseline from step 2.
7. Write the fault-injection test, inducing remote-call failure at each replaced call site and
   confirming correct attempt count, backoff timing, and terminal behavior on exhausted retries.
8. Document both call sites' new retry configuration.
9. Coordinate with W04-E02-S001's implementer if either hand-rolled implementation is found to be
   inside the `kernel/notify`/`kernel/webhook` effect stage S001 is simultaneously restructuring, to
   avoid two incompatible retry mechanisms landing on the same call site.

## Expected package or module changes

The two packages hosting the current hand-rolled retry implementations (exact packages TBD per step
1); `go.mod`/`go.sum` gaining `cenkalti/backoff/v5` as a direct dependency.

## Expected file changes where determinable

Not yet determinable — the exact files are this story's own first-step discovery. Once located, this
plan's implementer should update this section (not silently) before implementation proceeds past
step 1.

## Contracts and interfaces

None new — this story consumes `cenkalti/backoff/v5`'s existing public API; it does not define a new
internal interface.

## Data structures

None new.

## APIs

No caller-facing API change expected — the retry mechanism is an internal implementation detail of
whatever function currently performs the hand-rolled retry; to be confirmed once both locations are
identified.

## Configuration changes

`cenkalti/backoff/v5`'s configuration (max attempts, initial/max interval, multiplier, jitter) is
set per call site to match each prior schedule's parity baseline — whether these become
runtime-configurable values or fixed constants is an implementation-time decision, to be recorded
here once made.

## Persistence changes

None.

## Migration strategy

Not applicable — no schema or data migration.

## Concurrency implications

None beyond what each call site's own concurrency model already required — the retry mechanism swap
does not change either call site's concurrency semantics, only its retry-loop implementation.

## Error-handling strategy

`cenkalti/backoff/v5`'s own error-handling/retry-decision API (permanent vs. retryable errors)
replaces each hand-rolled implementation's equivalent logic — this distinction must be preserved
correctly at both call sites (a permanent error must not be retried; a retryable error must be,
within the configured bound).

## Security controls

None new beyond correct retry-bound configuration (an unbounded or too-aggressive retry
configuration would itself be a denial-of-service-adjacent concern against a remote provider) —
this mirrors, at a smaller scope, the bounded-retry security rationale used elsewhere in this
programme (e.g. W02-E01-S001-T002).

## Observability changes

Retry/backoff events remain logged at both call sites, preserving whatever observability the prior
hand-rolled implementations already provided, adapted to `cenkalti/backoff/v5`'s own retry-notify
hooks if the library exposes them.

## Testing strategy

- Retry-schedule-parity test: for each replaced call site, confirm the new library's configured
  behavior matches or documented-ly improves on the original hand-rolled schedule's baseline
  (step 2's documented behavior).
- Fault-injection test: for each replaced call site, induce remote-call failure and confirm correct
  attempt count, backoff timing, and terminal (give-up) behavior on exhausted retries.

## Regression strategy

The parity and fault-injection tests, once landed, become permanent regression guards against a
future change silently altering either call site's retry behavior.

## Compatibility strategy

No caller-facing API change expected (see "APIs" above); if either hand-rolled implementation is
found embedded in a call site with external callers depending on its exact timing behavior
(unlikely, but not yet confirmed), record that as a compatibility consideration once discovered, not
silently assumed away here.

## Rollout strategy

Single story, landed as its own reviewable unit. Both replacements may land as one commit/PR or two
separable ones — to be determined once both locations and their coordination needs (with S001, if
applicable) are known.

## Rollback strategy

Revert either replacement independently if it destabilizes the affected call site's retry
behavior; because both are library-configuration swaps rather than structural rewrites, reverting
either is expected to be low-risk and mechanically simple.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–9). Step 1 (locating both
implementations) is a hard prerequisite to every subsequent step.

## Task breakdown

- **W04-E02-S003-T001** — Locate and replace both hand-rolled retry implementations with
  `cenkalti/backoff/v5` (steps 1–5, 8–9 above).
- **W04-E02-S003-T002** — Retry-schedule-parity and fault-injection tests (steps 6–7 above).
- **W04-E02-S003-T003** — Lightweight review (see `tasks/index.md` "Grouping rationale" for why a
  lighter approach than mandate §14's full independent-review pattern is judged appropriate here,
  and what it still requires).

## Expected artifacts

The `cenkalti/backoff/v5` integration at both replaced call sites; retry-schedule-parity and
fault-injection test suites; documentation of both call sites' new retry configuration.

## Expected evidence

Retry-schedule-parity test output; fault-injection test output.

## Unresolved questions

- The exact file/line locations of both hand-rolled retry implementations — not pinpointed by the
  source text available to this plan; this story's own first implementation step (T001) resolves
  this.
- Whether either hand-rolled implementation is embedded inside a call site W04-E02-S001 is
  simultaneously restructuring (requiring coordination) — to be determined once both locations are
  found.
- Whether `cenkalti/backoff/v5`'s retry-bound/backoff parameters become runtime-configurable or
  fixed constants at each call site.

## Approval conditions

This plan is approved for implementation once: (a) both hand-rolled retry implementations' exact
locations are confirmed (T001's own first step), and (b) the owner and reviewer are assigned.
