---
id: W07-E04-S002
type: story
title: Closure and claim decision — programme closure report + production-readiness claim-upgrade decision package
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
  - W07-E04-S001
blocks: []
acceptance_criteria:
  - AC-W07-E04-S002-01
  - AC-W07-E04-S002-02
  - AC-W07-E04-S002-03
artifacts: []
evidence: []
decisions: []
risks: []
---

# W07-E04-S002 — Closure and claim decision — programme closure report + production-readiness claim-upgrade decision package

## Story ID

W07-E04-S002

## Title

Closure and claim decision — programme closure report + production-readiness claim-upgrade decision
package

## Objective

Produce the programme's own closure report, consolidating all 8 waves' own closure states, and a
separate, explicit production-readiness claim-upgrade decision package addressed to the human authority.
**This story does not itself declare the framework production-ready** — that determination is
explicitly reserved for the human authority the decision package is addressed to, mirroring the same
human-gated framing this programme already applies to W06-E03-S002's own DEC-Q10 activation.

## Value to the framework

`impl/index.md`'s own "Programme acceptance" section states the exact separation this story exists to
honor: "production-readiness claim upgrade is a separate, explicit decision." Without this story, the
programme's own closure could be misread as an implicit production-readiness claim simply because every
wave reached `accepted` — exactly the failure mode mandate §7 warns against at the task/story level
("A story must not be accepted solely because all tasks are marked complete"), applied here at the
whole-programme level. This story's own value is making that separation explicit and unmissable: the
programme's own technical closure (every wave done, every finding addressed or honestly deferred) is one
document; the decision about what that closure means for a production-readiness claim is a second,
separate document addressed to a human, not self-issued by this programme's own execution.

## Problem statement

The source REVIEW document's own §30 gate modeled this exact distinction already, at the pre-programme
stage: "the production-readiness programme above is APPROVED as the authoritative backlog. `HEAD` is
**not** approved as production-ready." This story re-applies that same distinction at the
post-programme stage: the programme's own closure report can honestly state "every wave's own exit
criteria were satisfied" (or honestly state which were not, and why) without that statement itself
constituting a production-readiness claim. `impl/index.md`'s own programme-acceptance language is the
authoritative source for why these must be two separate documents, not one document with two sections
that risk being read as equivalent in weight.

## Source requirements

None directly — this story consumes W07-E04-S001's own gate re-run output and every prior wave's own
`closure-report.md`.

## Current-state assessment

At this story's own planning time, no wave has yet executed. This story's own current-state assessment
is necessarily forward-looking: at the time this story is actually executed, the "current state" is
W07-E04-S001's own completed gate re-run, plus all 8 waves' own `closure-report.md` files, read directly.

## Desired state

A programme closure report exists, consolidating: every wave's own exit-criteria satisfaction status;
every epic's own `closure-report.md` status; the traceability-completeness and disposition-audit
findings from W07-E04-S001; any deviation record across the entire programme; any accepted residual
risk across the entire programme (e.g. DEC-Q9's own continued-open status if unresolved, DEC-Q10's own
activation status). A separate production-readiness claim-upgrade decision package exists, addressed to
the human authority, presenting: the programme's own closure state as *input* to a decision, not as the
decision itself; every open item the closure report surfaces (unresolved DEC-Qs, any gap
W07-E04-S001's own gate re-run found, any deferred work); and an explicit statement that the decision to
upgrade any production-readiness claim rests with the human authority, not with this programme's own
execution.

## Scope

- Compile the programme closure report from all 8 waves' own `closure-report.md` files and
  W07-E04-S001's own gate re-run output.
- Compile the production-readiness claim-upgrade decision package as a separate document, presenting the
  closure state as decision input, explicitly not as a self-issued declaration.
- Ensure every open item (unresolved decisions, gaps, deferred work) is carried into the decision
  package, not silently dropped from the closure report to the decision package.

## Out of scope

- **Making the production-readiness declaration itself** — reserved for the human authority.
- **Resolving any open item the closure report surfaces** (e.g. DEC-Q9, DEC-Q10, any gap
  W07-E04-S001 found) — these are carried forward as explicit open items in the decision package, not
  resolved by this story.

## Assumptions

- The human authority the claim-upgrade decision package is addressed to is not named by any source
  document available to this planning generation — this story's own implementation records the package
  as addressed to "the framework's designated production-readiness decision authority" generically,
  leaving the specific named recipient as an implementation-time detail (consistent with mandate §18's
  own instruction not to invent a specific fact the source does not give).
- The exact format of both the closure report and the decision package (a single consolidated document
  vs. two genuinely separate files) is not specified by any source document beyond `impl/index.md`'s own
  "production-readiness claim upgrade is a separate, explicit decision" language — this story's own
  implementation treats "separate" as requiring two genuinely distinct documents, not two sections of one
  document, per the strongest reading of that language.

## Dependencies

Depends on W07-E04-S001 (the final verification gate) reaching `accepted` first — this story's own
closure report and decision package both consume S001's own gate re-run output directly.

## Affected packages or components

None — this is a documentation story with zero code change.

## Compatibility considerations

Not applicable.

## Security considerations

Not applicable directly, though the decision package's own content (if it surfaces an unresolved
security-relevant open item, e.g. an unwaived finding from W07-E02-S001's own SEC-05 assessment) must
carry that forward honestly, not soften or omit it to present a cleaner closure narrative.

## Performance considerations

Not applicable.

## Observability considerations

Not applicable.

## Migration considerations

Not applicable.

## Documentation requirements

This story's entire output is documentation: the programme closure report and the production-readiness
claim-upgrade decision package.

## Acceptance criteria

- **AC-W07-E04-S002-01**: A programme closure report exists, consolidating all 8 waves' own exit-criteria
  status, epic closure statuses, W07-E04-S001's own gate re-run findings, and every deviation/accepted-
  risk record across the programme.
- **AC-W07-E04-S002-02**: A separate production-readiness claim-upgrade decision package exists, addressed to
  the human authority, presenting the closure state as decision input — with an explicit statement that
  this programme's own execution does not itself declare the framework production-ready.
- **AC-W07-E04-S002-03**: Every open item the closure report surfaces (unresolved DEC-Qs, any gap
  W07-E04-S001 found, any deferred work) is carried into the decision package, none silently dropped
  between the two documents.

## Required artifacts

- The programme closure report.
- The production-readiness claim-upgrade decision package.
See `artifacts/index.md`.

## Required evidence

- The closure report's own completeness confirmation (every wave/epic accounted for).
- The decision package's own open-item-carry-forward confirmation (nothing dropped between the two
  documents).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, all three acceptance criteria numbered and measurable, dependency on
W07-E04-S001 recorded, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the decision package genuinely does not contain a
self-issued production-readiness declaration anywhere in its own text, and that every open item from the
closure report genuinely appears in the decision package.

## Risks

None recorded at this story's own scope beyond the general risk (mitigated by this story's own explicit
two-document structure and its own AC-03) that a closure report's own positive framing could
inadvertently soften or omit an open item when compiled into the decision package — this is exactly why
AC-03 exists as its own, separately-verified acceptance criterion.

## Residual-risk expectations

Once all three acceptance criteria are verified, residual risk is expected to be low.

## Plan

See `plan.md`.
