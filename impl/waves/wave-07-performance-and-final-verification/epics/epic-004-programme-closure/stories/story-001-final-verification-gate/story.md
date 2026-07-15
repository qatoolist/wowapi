---
id: W07-E04-S001
type: story
title: Final verification gate — programme-wide REVIEW §30 re-run, traceability completeness, disposition audit
status: planned
wave: W07
epic: W07-E04
owner: unassigned
reviewer: unassigned
priority: P0
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements: []
depends_on:
  - W07-E01
  - W07-E02
  - W07-E03
blocks:
  - W07-E04-S002
acceptance_criteria:
  - AC-W07-E04-S001-01
  - AC-W07-E04-S001-02
  - AC-W07-E04-S001-03
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W07-003
  - RISK-W07-E04-001
---

# W07-E04-S001 — Final verification gate — programme-wide REVIEW §30 re-run, traceability completeness, disposition audit

## Story ID

W07-E04-S001

## Title

Final verification gate — programme-wide REVIEW §30 re-run, traceability completeness, disposition
audit

## Objective

Re-run the REVIEW §30-style final approval gate across the whole 8-wave programme against current HEAD;
confirm the traceability matrix shows every `requirement-inventory.md` row with a disposition and no
silent drop; and audit that every item genuinely reached its recorded disposition, not merely claimed
to — with spot-check evidence, not blanket trust.

## Value to the framework

The source REVIEW document's own "Final approval gate (§30)" section models exactly the rigor this story
must apply: it did not merely assert a conclusion, it ran an independent completeness audit against
itself and recorded each finding's own adjudication — "ACCEPTED & CORRECTED," "REJECTED," "NOTED" —
rather than accepting audit output uncritically. This story applies that same discipline to the
programme's own final state: not "did every wave say it was done," but "is every wave's own claim of
`accepted` actually backed by genuine evidence, genuine independent review, and genuine acceptance-
criteria satisfaction." This is the single highest-leverage verification step in the entire programme,
since a false "everything is done" claim at this final gate would propagate into every downstream
consumer of this programme's own closure record, including the claim-upgrade decision package
(W07-E04-S002) this story directly feeds.

## Problem statement

The source REVIEW document's own §30 section states: "the production-readiness programme above is
APPROVED as the authoritative backlog. `HEAD` is **not** approved as production-ready — it is approved
as a *verified, honest foundation* with a corrected, sequenced, dependency-aware path." That statement
was made against `HEAD` as it existed on 2026-07-11, before any wave of this programme had executed. This
story's own purpose is to ask the same question again, against the programme's actual final state (all 8
waves executed) — not to assume the answer is unchanged, and not to merely restate the original verdict
without fresh verification.

## Source requirements

None directly — this story re-runs REVIEW's own §30 methodology against the programme's final state.

## Current-state assessment

At this story's own planning time, no wave has yet executed — this planning document itself is being
generated before any implementation work begins. This story's own current-state assessment is therefore
necessarily forward-looking: at the time this story is actually executed (after all 8 waves have reached
their own closure state), the "current state" to assess is each wave's own `closure-report.md`, each
epic's own `closure-report.md`, and each story's own `closure.md` — read directly, not summarized from
memory or trusted at face value.

## Desired state

A fresh, genuine re-run of the REVIEW §30-style gate exists, covering: (a) the capability-matrix-style
assessment REVIEW §H performed originally, re-assessed against the programme's actual final
implementation; (b) the mandatory-capability-readiness assessment REVIEW §I performed originally,
re-assessed the same way; (c) a traceability-completeness check confirming every `requirement-
inventory.md` row (§A plan findings, §B review findings/decisions, §C matrix verify-outcomes, §D
product-level items, §E session-delta facts) has a final disposition, with none silently dropped between
this planning generation and the programme's own actual execution; (d) a disposition audit,
spot-checking a meaningful sample of `accepted` claims across the programme's own stories against their
own actual evidence, not merely trusting each story's own self-reported `accepted` status.

## Scope

- Re-run the REVIEW §H-style capability-matrix assessment against the programme's actual final
  implementation.
- Re-run the REVIEW §I-style mandatory-capability-readiness assessment.
- Confirm traceability-matrix completeness across all of `requirement-inventory.md`'s own rows.
- Perform a disposition audit: spot-check a meaningful sample of `accepted` story claims against their
  own actual evidence.

## Out of scope

- **Fixing any gap this story's own re-run discovers** — recorded as an explicit open item for
  W07-E04-S002's own claim-upgrade decision package, per RISK-W07-003's own framing; not silently
  absorbed into this story's own bounded scope.
- **Making the production-readiness declaration itself** — that is explicitly not this story's own
  decision (nor W07-E04-S002's); the decision belongs to the human authority the claim-upgrade package
  is addressed to.

## Assumptions

- This story is executed only after every other epic in this wave (W07-E01, W07-E02, W07-E03) has
  reached its own closure state — its own re-run would be premature and incomplete otherwise.
- The exact sample size/selection methodology for the disposition audit's own spot-check is not
  specified by any source document — this story's own implementation determines a methodology
  proportionate to the programme's own total story count, favoring breadth across waves over depth within
  any single wave.

## Dependencies

Depends on W07-E01, W07-E02, W07-E03 (this wave's own other three epics) reaching their own closure
state first. Depends transitively on every prior wave (W00-W06) having reached its own closure state,
since the gate re-runs across the whole programme.

## Affected packages or components

None — this is a documentation/verification story with zero code change.

## Compatibility considerations

Not applicable.

## Security considerations

The disposition audit's own spot-check should weight toward P0/critical-priority stories (security,
tenant-isolation, release-gating work) when selecting its sample, since a false `accepted` claim on one
of those carries materially higher consequence than on a P2 documentation story.

## Performance considerations

Not applicable.

## Observability considerations

Not applicable.

## Migration considerations

Not applicable.

## Documentation requirements

This story's entire output is documentation: the fresh gate re-run's own report, the traceability-
completeness check's own output, and the disposition audit's own output.

## Acceptance criteria

- **AC-W07-E04-S001-01**: A fresh REVIEW §30-style gate re-run exists, covering both a §H-style capability-
  matrix reassessment and a §I-style mandatory-capability-readiness reassessment, against the
  programme's actual final implementation — not a restatement of REVIEW's own original 2026-07-11
  conclusions.
- **AC-W07-E04-S001-02**: The traceability-completeness check confirms every `requirement-inventory.md` row
  (§A-E) has a final disposition, with none silently dropped.
- **AC-W07-E04-S001-03**: The disposition audit spot-checks a meaningful, weighted-toward-P0/critical sample of
  `accepted` story claims across the programme against their own actual evidence, confirming each
  sampled claim is genuine, not merely self-reported.

## Required artifacts

- The fresh gate re-run report.
- The traceability-completeness check output.
- The disposition-audit report.
See `artifacts/index.md`.

## Required evidence

- The gate re-run's own evidence trail (what was actually re-checked, against what source).
- The traceability-completeness check's own row-by-row output.
- The disposition audit's own sampled-claim evidence trail.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, all three acceptance criteria numbered and measurable, dependency on all three
sibling epics recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming this story's own gate re-run is a genuine fresh
re-verification (per RISK-W07-E04-001's own framing), not a restatement of REVIEW's own original
conclusions with the dates changed.

## Risks

RISK-W07-003 (a genuine gap discovered in an earlier wave's own closure) and RISK-W07-E04-001 (the
re-run being performed as a restatement rather than genuine re-verification) — see epic-level `risks.md`
for full detail and mitigation/contingency.

## Residual-risk expectations

RISK-W07-E04-001's own residual risk is expected to be low once this story's own independent-review task
honors its explicit restatement-vs-re-verification check. RISK-W07-003 cannot be pre-resolved — its
outcome is a fact to be discovered by this story's own execution, carried forward honestly into
W07-E04-S002 if found.

## Plan

See `plan.md`.
