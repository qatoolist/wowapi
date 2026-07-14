---
id: W07-E04
type: epic
title: Programme closure
status: planned
wave: W07
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements: []
depends_on:
  - W07-E01
  - W07-E02
  - W07-E03
stories:
  - W07-E04-S001
  - W07-E04-S002
decisions: []
risks:
  - RISK-W07-003
---

# W07-E04 — Programme closure

## Epic objective

Re-run the REVIEW §30-style final approval gate across the whole 8-wave programme against current HEAD,
confirm the requirement-traceability matrix shows every `requirement-inventory.md` row with a
disposition and no silent drop, and audit that every item genuinely reached its recorded disposition —
not merely claimed to; then produce the programme's own closure report and a separate, explicit
production-readiness claim-upgrade decision package for the human authority, since this wave — and this
epic specifically — does not itself declare the framework production-ready.

## Problem being solved

`impl/index.md`'s own "Programme acceptance" section states the exact bar this epic exists to satisfy:
"All waves closed per their `closure-report.md`; requirement-traceability matrix shows every `planned`
item accepted/deferred-with-approval; the REVIEW §30-style final gate re-run passes; no unexplained
deviation; production-readiness claim upgrade is a separate, explicit decision." The source REVIEW
document's own "Final approval gate (§30)" section gives this epic its exact re-run shape: "Fable 5
verdict: the three review commits are APPROVED TO REMAIN... and the production-readiness programme above
is APPROVED as the authoritative backlog. `HEAD` is **not** approved as production-ready — it is
approved as a *verified, honest foundation* with a corrected, sequenced, dependency-aware path." This
epic's own job is to re-run that same class of judgment against the programme's actual final state (all
8 waves executed), not merely restate REVIEW's own 2026-07-11 conclusions as if they still describe
current HEAD without re-verification.

## Scope

- Re-run the REVIEW §30-style final approval gate across all 8 waves' own closure states.
- Check the requirement-traceability matrix for completeness: every `requirement-inventory.md` row has a
  disposition, none silently dropped.
- Audit that every item genuinely reached its recorded disposition (not merely claimed to) —
  spot-checking, not merely trusting, closure claims across the programme.
- Produce the programme closure report.
- Produce a separate, explicit production-readiness claim-upgrade decision package for the human
  authority.

## Out of scope

- **Declaring the framework production-ready** — explicitly not this epic's own decision to make; the
  claim-upgrade decision package is a recommendation/decision-input document for a human authority, not
  a self-issued declaration.
- **Fixing any gap the final gate re-run discovers** — per RISK-W07-003's own framing (wave-scoped), a
  genuine gap found this late in the programme is recorded as an explicit open item for the human
  authority's own decision, not silently absorbed into this epic's own already-defined scope by reopening
  an earlier wave's own closed story.

## Source requirements

None directly (this epic re-runs REVIEW's own §30 gate against the programme's actual final state,
across all prior findings/waves).

## Architectural context

This epic is the programme's own terminal closure mechanism — `impl/analysis/wave-allocation-detail.md`'s
own W07-E04 grouping states this exactly: "S001 final-verification-gate (re-run REVIEW §30-style gate
across the programme; traceability-matrix completeness; disposition audit); S002 closure-and-claim-
decision (programme closure report; production-readiness claim upgrade decision package for the human
authority)." The two-story split mirrors the two distinct outputs a closure process needs: a verification
step (S001, "is everything actually as claimed") and a decision-packaging step (S002, "given that
verification, what should a human decide") — these are kept separate because conflating them would risk
S002's own decision-packaging work quietly substituting for S001's own independent verification, exactly
the failure mode mandate §7 warns against ("A story must not be accepted solely because all tasks are
marked complete").

## Included stories

- **W07-E04-S001 — final-verification-gate**: re-run REVIEW §30's gate across the programme;
  traceability-matrix completeness check; disposition audit.
- **W07-E04-S002 — closure-and-claim-decision**: the programme closure report; the production-readiness
  claim-upgrade decision package for the human authority.

## Dependencies

Depends on W07-E01, W07-E02, W07-E03 (this wave's own other three epics) — the final gate's own re-run
scope cannot be meaningful until every other epic in this wave has reached its own closure state. Depends
transitively on every prior wave (W00-W06), since the gate re-runs across the whole programme, not just
this wave.

## Risks

RISK-W07-003 (the final gate discovers an unresolved gap in an earlier wave's own closure that this
epic cannot itself fix without reopening that wave's own scope) originates at wave scope and lands
entirely within this epic's S001/S002. See `risks.md` for the epic-scoped elaboration.

## Required decisions

None new in the D-0N sense. This epic's own S002 produces a decision *package* for a human authority —
it does not itself make a new architecture decision.

## Epic acceptance criteria

- **AC-W07-E04-01**: The REVIEW §30-style final approval gate has been re-run against current HEAD
  (across all 8 waves), producing a genuine, freshly-derived verdict, not a restatement of REVIEW's own
  original 2026-07-11 conclusions.
- **AC-W07-E04-02**: The traceability matrix shows every `requirement-inventory.md` row with a
  disposition; no row is silently dropped.
- **AC-W07-E04-03**: The disposition audit confirms every item genuinely reached its recorded
  disposition, with spot-check evidence, not blanket trust.
- **AC-W07-E04-04**: The programme closure report is complete, covering all 8 waves' own closure states.
- **AC-W07-E04-05**: A separate, explicit production-readiness claim-upgrade decision package exists for
  the human authority — this epic does not itself declare the framework production-ready.

## Closure conditions

Both stories reach `accepted`; AC-W07-E04-01 through AC-W07-E04-05 above are all satisfied;
`closure-report.md` for this epic is completed with reviewer conclusion and acceptance date; any gap the
final gate re-run discovers is recorded honestly in the claim-upgrade decision package, not silently
absorbed or dropped.
