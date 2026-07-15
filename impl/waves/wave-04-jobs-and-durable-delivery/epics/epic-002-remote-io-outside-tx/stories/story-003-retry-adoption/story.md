---
id: W04-E02-S003
type: story
title: Adopt cenkalti/backoff/v5 for duplicated retry logic
status: accepted
wave: W04
epic: W04-E02
owner: W04-Rerun
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - FBL-04
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W04-E02-S003-01
  - AC-W04-E02-S003-02
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W04-E02-S003-001
---

# W04-E02-S003 — Adopt cenkalti/backoff/v5 for duplicated retry logic

## Story ID

W04-E02-S003

## Title

Adopt cenkalti/backoff/v5 for duplicated retry logic

## Objective

Replace the framework's two duplicated hand-rolled retry implementations with
`cenkalti/backoff/v5`, proving retry-schedule parity with the prior hand-rolled schedules and
correct behavior under fault injection.

## Value to the framework

REVIEW §K's reuse-opportunity register identifies this precisely: "Retry/backoff (hand-rolled ×2) |
custom, duplicated | Replace → `cenkalti/backoff/v5` (already in module graph, unused) |
Duplication + a mature lib already transitively present. FBL-04." Removing duplicated retry logic
in favor of a single, well-tested, already-present library reduces the framework's own maintenance
surface and behavioral-inconsistency risk between the two current implementations, without
introducing a new external dependency (the library is already transitively present in the module
graph).

## Problem statement

REVIEW §O's detailed task register states, verbatim: "FBL-04 (P1): Replace duplicated hand-rolled
retry with `cenkalti/backoff/v5`. Tests: retry-schedule parity + fault injection." REVIEW §L's
approved dependency register confirms: "New approvals for reuse work: `cenkalti/backoff/v5` (MIT,
already transitive)." `requirement-inventory.md` row FBL-04: "Adopt cenkalti/backoff for duplicated
retry | IMPL | P1 | planned | W04-E02-S003 | Approved dep; parity + fault-injection tests." This is
a small, well-bounded item: two duplicated hand-rolled retry implementations exist somewhere in the
framework's remote-I/O paths, and this story replaces both with the one approved library.

## Source requirements

FBL-04.

## Current-state assessment

Per REVIEW §K's own framing, two hand-rolled retry implementations exist today, duplicated rather
than shared, and `cenkalti/backoff/v5` is "already in module graph, unused" — meaning the library
is already a transitive dependency of the module but is not yet imported or used directly anywhere
in the framework's own code. This story's own re-confirmation step is to identify both hand-rolled
retry implementations' exact locations at this story's actual start commit (not yet pinpointed by
this plan, since the source text names the duplication pattern but not the exact file/line
locations) before replacing either.

## Desired state

Both hand-rolled retry implementations are replaced with `cenkalti/backoff/v5`, configured to match
(or intentionally and documented-ly improve upon) each prior implementation's own retry schedule. A
retry-schedule-parity test proves the new library's behavior matches or improves on each prior
schedule. A fault-injection test proves correct retry/backoff behavior under induced remote-call
failure.

## Scope

- Identifying both existing hand-rolled retry implementations' exact locations in the codebase.
- Adding `cenkalti/backoff/v5` as a direct dependency (already transitively present; this story adds
  the direct `go.mod` requirement and the import).
- Replacing both hand-rolled implementations with `cenkalti/backoff/v5`, configured for
  retry-schedule parity with (or a documented, deliberate improvement over) each prior schedule.
- Retry-schedule-parity and fault-injection tests, per REVIEW §O's own required test coverage.

## Out of scope

- Any redesign of the call sites that use the retry logic beyond swapping the retry mechanism itself
  — this story does not restructure `kernel/notify`/`kernel/webhook`'s transaction boundaries (that
  is W04-E02-S001/S002's scope) or any other caller's control flow beyond the retry call itself.
- Coordinating the exact retry mechanism used inside W04-E02-S001's three-stage protocol's effect
  stage, if that stage's own retry behavior is one of the two hand-rolled implementations being
  replaced — if so, this story's implementer and S001's implementer must coordinate to avoid
  building two incompatible retry mechanisms for the same call site; this coordination note is
  carried in `plan.md`, not silently resolved by assumption here.

## Assumptions

- The exact two hand-rolled retry implementations are not pinpointed by file/line in the source
  text available to this plan — REVIEW §K names the duplication pattern ("Retry/backoff (hand-rolled
  ×2)") without citing exact locations. This story's own first implementation step is to locate
  both, recorded as an implementation-time discovery, not invented here.
- `cenkalti/backoff/v5`'s API (constructor, retry-policy configuration, context-aware retry) is
  assumed sufficient to express both prior hand-rolled schedules' behavior — this is a reasonable
  assumption given the library's maturity and REVIEW's own approval of it for this exact purpose,
  but is confirmed only once both replacements are actually implemented and parity-tested.

## Dependencies

None within this epic — FBL-04 has no task dependency on S001 or S002 per `wave-allocation-detail.md`'s
own grouping ("S003 retry-adoption"), which places it in this epic by shared-package proximity, not
by a task dependency. This story may proceed in parallel with S001/S002, subject to the coordination
note above if a shared call site is involved.

## Affected packages or components

The two locations hosting the current hand-rolled retry implementations (exact packages TBD at
implementation time — likely candidates include `kernel/notify` and/or `kernel/webhook`'s own
remote-I/O retry paths, given REVIEW's framing of FBL-04 alongside DATA-03's remote-I/O restructuring,
but this is not confirmed by the source and must be established by the implementation-time location
step, not assumed here).

## Compatibility considerations

Replacing a hand-rolled retry schedule with `cenkalti/backoff/v5`'s configured equivalent could
subtly change retry timing/behavior even when "parity" is the goal — the retry-schedule-parity test
is the explicit control for this risk, per REVIEW §O's own required test.

## Security considerations

None beyond general retry-logic correctness — an incorrectly configured backoff (e.g. too-aggressive
immediate retry) could itself become a denial-of-service-adjacent concern against a remote provider,
which the parity/fault-injection tests are expected to catch.

## Performance considerations

`cenkalti/backoff/v5`'s own overhead is expected to be negligible relative to the remote calls it
wraps; no separate performance investigation is warranted beyond the parity test confirming
comparable retry timing.

## Observability considerations

Retry/backoff events should remain observable (logged) after the swap, consistent with any existing
logging on the current hand-rolled implementations — this story should not silently regress
observability while changing the retry mechanism.

## Migration considerations

No schema or data migration. This is a pure code-level dependency swap.

## Documentation requirements

Document that both retry call sites now use `cenkalti/backoff/v5`, with their configured retry
schedules, so a future maintainer does not reintroduce a third hand-rolled implementation.

## Acceptance criteria

- **AC-W04-E02-S003-01**: Both of the framework's hand-rolled retry implementations are replaced
  with `cenkalti/backoff/v5`, with no hand-rolled retry logic remaining at either original call
  site — proven by a retry-schedule-parity test confirming the new library's configured behavior
  matches or improves on each prior implementation's own retry schedule.
- **AC-W04-E02-S003-02**: A fault-injection test proves correct retry/backoff behavior under
  induced remote-call failure for both replaced call sites (correct number of attempts, correct
  backoff timing, correct terminal behavior on exhausted retries).

## Required artifacts

- The `cenkalti/backoff/v5` integration at both replaced call sites.
- Retry-schedule-parity and fault-injection test suites.
See `artifacts/index.md`.

## Required evidence

- Retry-schedule-parity test output.
- Fault-injection test output.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, no blocking dependency recorded
(none exists), owner/reviewer assignment pending, the exact locations of both hand-rolled
implementations recorded as an implementation-time discovery step rather than silently assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; both acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; a lightweight review
(per `tasks/index.md`'s own documented rationale, appropriate to this story's P1/low-risk,
well-bounded profile) confirms both replacements are genuine (no hand-rolled logic silently left in
place alongside the new library) and both tests are meaningful, not merely present.

## Risks

RISK-W04-E02-S003-001 (a parity test could pass while subtly mis-configuring the new library's
backoff parameters relative to the original schedule's intent, if the original schedule's exact
intent was itself under-documented) — see "Risks" in `plan.md` for the mitigating discovery step.

## Residual-risk expectations

This is a small, well-bounded, P1 item — per this story's own prompt framing, "do not artificially
inflate scope." Residual risk is expected to be low once both hand-rolled locations are correctly
identified and the parity/fault-injection tests pass.

## Plan

See `plan.md`.
