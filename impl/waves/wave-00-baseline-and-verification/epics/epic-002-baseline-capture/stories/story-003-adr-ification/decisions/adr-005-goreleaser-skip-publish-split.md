---
id: ADR-W00-E02-S003-005
type: decision
title: GoReleaser --skip=publish build-candidate + separate publish step
status: ratified
context: REL-01 GoReleaser split-mode — should the release pipeline use GoReleaser's built-in skip-publish mode, or a hand-rolled pipeline?
date: 2026-07-12
deciders:
  - Fable 5 (framework architecture lead role)
related_source_items:
  - D-05
  - W06-E03
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# ADR-W00-E02-S003-005 — GoReleaser --skip=publish build-candidate + separate publish step

**Formalization note:** This ADR formalizes a decision Fable 5 already made in
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F/§U; this ADR is the
programme's own durable record of it, not a new decision-making act.

## Decision ID

ADR-W00-E02-S003-005.

## Title

GoReleaser `--skip=publish` build-candidate + separate publish step.

## Status

ratified — the underlying decision was already made by Fable 5 in REVIEW §F row 6 (Q6); this ADR
file's own creation/registration is tracked separately by task W00-E02-S003-T002's own `status:
todo`→`done` lifecycle (see `../story.md` "Status discipline").

## Context

REVIEW §F question 6 asks: for REL-01's "release gated on exact published commit" requirement,
should the release pipeline build its release-candidate artifacts and then gate/publish them using
GoReleaser's own built-in split-mode support (`release --skip=publish` followed by a separate
`goreleaser publish` step), or should the team hand-roll a custom two-phase pipeline?

## Options considered

- **Hand-rolled two-phase release pipeline** (custom scripting to separate build-candidate from
  publish) — rejected. REVIEW §F row 6 states "no hand-rolled pipeline needed."
- **GoReleaser `release --skip=publish` for build-candidate + a separate `goreleaser publish`
  step** — chosen. See Decision below.

## Decision

**Use GoReleaser `release --skip=publish` for build-candidate + a separate `goreleaser publish`
step (not a hand-rolled pipeline) — verify against the pinned GoReleaser version at implementation
time (this is a caveat, not yet independently confirmed).** (REVIEW §F row 6, quoted verbatim,
combined with this task's own brief which restates the same caveat.)

### Safe default

No distinct safe-default stated as a separate fallback path — REVIEW §F row 6 states an
unconditional resolution ("resolved"). The closest analogue to a safe default is the source's own
explicitly acknowledged caveat (see Rationale/Consequences below): the decision's buildability is
conditioned on confirming this split-mode support actually exists as documented in the specific
GoReleaser version this programme pins, which had not yet been independently confirmed as of the
REVIEW document's writing.

## Rationale

REVIEW §F row 6, quoted verbatim: "Supported in current GoReleaser; no hand-rolled pipeline
needed." Using GoReleaser's own built-in mechanism avoids the maintenance burden and correctness
risk of a custom-built two-phase pipeline (a hand-rolled pipeline would have to independently
reimplement artifact-integrity guarantees GoReleaser already provides). The decision is qualified,
per this task's own brief: "verify against the pinned GoReleaser version at implementation time
(this is a caveat, not yet independently confirmed)" — i.e., REVIEW's claim that this is "supported
in current GoReleaser" was not independently re-verified against this specific repository's pinned
GoReleaser version as part of the REVIEW pass itself.

## Consequences

- REL-01 T6 (W06-E03, per `requirement-inventory.md`'s REL-01 row) implements the release pipeline
  using GoReleaser's `--skip=publish` + `publish` split, not a hand-rolled alternative.
- Before REL-01 T6's implementation is trusted as correct, the pinned GoReleaser version in this
  repository must be checked against its own documentation/changelog to confirm `--skip=publish`
  and a separate `publish` invocation behave as REVIEW assumes — this confirmation step is REL-01's
  own implementation-time responsibility, explicitly flagged here so it is not silently assumed to
  have already happened.
- `requirement-inventory.md`'s REL-01 row notes "~85% buildable now; final activation = DEC-Q10
  (admin)" — this ADR's decision governs the pipeline *design*, not the separate repo-admin
  activation blocker (`DEC-Q10`), which remains a distinct, tracked human decision.

## Related source items

D-05; downstream epic W06-E03 (REL-01) — unblocked by this ADR per
`../../../../dependencies.md` and `../story.md` "Dependencies."

## Date

2026-07-12.

## Deciders

Fable 5 (verify against pinned GoReleaser version at implementation).
