---
id: W06-E01-S001
type: story
title: Module DSL design — state-of-the-art DSL design record (target, not implemented)
status: verified
wave: W06
epic: W06-E01
owner: W06E01Impl
reviewer: W06E04Impl (future-state labeling); W06-E01-E04-Execution.W06E01ReviewR (independent review)
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-13
source_requirements:
  - DX-03
depends_on: []
blocks:
  - W06-E02-S003
acceptance_criteria:
  - AC-W06-E01-S001-01
  - AC-W06-E01-S001-02
artifacts:
  - ART-W06-E01-S001-001
  - ART-W06-E01-S001-002
evidence:
  - EV-W06-E01-S001-001
  - EV-W06-E01-S001-002
  - EV-W06-E01-S001-003
decisions:
  - ADR-W06-E01-S001-001
risks: []
---

# W06-E01-S001 — Module DSL design — state-of-the-art DSL design record (target, not implemented)

## Story ID

W06-E01-S001

## Title

Module DSL design — state-of-the-art DSL design record (target, not implemented)

## Objective

Formalize the design of the state-of-the-art module DSL (`port`, `Manifest[T]`,
`Operation[Request,Response]`) into a design document and an ADR-style decision record, explicitly
labeled "target, not implemented" per AR-05's future-state-labeling discipline. **This story produces
no code.** It is a design-investigation story: its outputs are a design document and a decision record,
not an implementation.

## Value to the framework

PLAN's own DX-03 framing places this precisely: "Define the state-of-the-art module DSL (Wave 4, P1,
future design — not near-term implementation)." The value here is not a working DSL — that is
explicitly out of scope for this programme (see "Out of scope" below) — but a durable, reviewable record
of what the framework's own design direction is, so that (a) AR-05 T5's future-state-labeling lint has
something correctly labeled to point at instead of unlabeled aspirational prose, and (b) REL-03's T5
leg (event/schema compatibility, MATRIX CS-15's own framing: "Blocked on DX-03/AR-03 — the concept
doesn't exist in current source") has a concrete design target to reference when it eventually
unblocks, even though this story does not itself unblock T5 by implementing anything.

## Problem statement

PLAN's own DX-03 evidence states plainly: "confirmed no `port`/`Manifest[T]`/`Operation[Request,
Response]` DSL exists anywhere in wowapi today — the directive's 'proposed API, not current source'
framing is accurate. Current DSL-adjacent surface (`module.Context`, string/any-keyed registries, the
closed authz verb set) is exactly what AR-01/AR-02/DX-02 already target from the correctness-fix side;
DX-03 formalizes the type-system-level version once those land." PLAN's own task table for DX-03 is
minimal and explicit about scope: "DX-03-T0. Formalize the design into an ADR under `decisions.md`,
explicitly labeled 'target, not implemented' per AR-05 | Wave 1 (AR-01 ApplicationModel, AR-02 typed
ports) complete | Design-only" followed by "DX-03-T1..Tn. Implementation | Wave 1-3 exit gates; DX-02
P1-T4 reuses this compiler | Deferred — out of near-term scope per §12 Wave 4." There is, today, no
design document and no decision record for this DSL anywhere in the repository — only the directive's
own prose description of what such a DSL might look like.

## Source requirements

DX-03 (T0 only — T1..Tn are explicitly out of scope, see below).

## Current-state assessment

Per PLAN's own evidence (to be re-confirmed at this story's own execution commit, per this programme's
fail-first re-confirmation convention applied elsewhere): no `port` type, no `Manifest[T]` generic, no
`Operation[Request,Response]` generic, and no compiler or code-generation mechanism consuming any such
type exists anywhere in `wowapi` today. The framework's actual current DSL-adjacent surface is
`module.Context` (the mutable, ~39-method interface MATRIX CS-02 describes), string/any-keyed
registries, and the closed authz verb set — all of which AR-01/AR-02/DX-02 target from the correctness-
fix side, not the type-system-redesign side. No design document or ADR-style decision record for the
future DSL exists in `docs/` or `impl/` today.

## Desired state

A design document exists describing the module DSL's proposed shape (`port`, `Manifest[T]`,
`Operation[Request,Response]`) at a level of detail sufficient for a future implementer to begin work
without re-deriving the design from the directive's own prose, together with an ADR-style decision
record capturing the key design trade-offs considered and the rationale for the chosen shape. Both
documents are explicitly and visibly labeled "target, not implemented," consistent with AR-05 T5's own
lint requirement (a future-state design block that is not labeled fails AR-05 T5's lint) — this story's
own output must itself pass that lint once it exists, not merely describe the labeling convention in
the abstract.

## Scope

- Drafting the module-DSL design's options and trade-offs (storage/typing approach for `port`, the
  generic shape of `Manifest[T]` and `Operation[Request,Response]`, how the design relates to AR-01's
  `Registrar` capability and AR-02's typed `port.Key[T]` mechanism it is expected to build on).
- Formalizing the chosen design into an ADR-style decision document, explicitly labeled "target, not
  implemented" per AR-05's labeling convention (PLAN DX-03-T0's own exact instruction: "Formalize the
  design into an ADR under `decisions.md`").
- Confirming the design document and decision record are correctly labeled such that AR-05 T5's lint
  (once it exists, per W06-E04-S002) would not flag them as unlabeled future-state prose.

## Out of scope

- **DX-03-T1..Tn (implementation)** — the DSL's actual compiler, code generation, or runtime type-
  system change. PLAN's own table marks this "Deferred — out of near-term scope per §12 Wave 4." This
  story produces zero implementation code.
- **DX-02 P1-T4's reuse of this compiler** — PLAN's own dependency note ("DX-02 P1-T4 reuses this
  compiler") describes a future consumer of DX-03's eventual implementation, not this story's own scope;
  DX-02's P1/Wave-4 tasks are themselves out of scope for the whole programme as currently planned
  (DX-02's Wave-0 slice was already executed at W01; its P1 remainder is not scheduled in any wave
  W00–W07 covers).
- **Actually implementing AR-05 T5's future-state-labeling lint** — that is W06-E04-S002's own scope;
  this story only needs its own output to be correctly labeled such that the lint (once built) would
  pass against it, it does not build the lint itself.

## Assumptions

- The exact internal shape of `port`, `Manifest[T]`, and `Operation[Request,Response]` — field lists,
  generic constraints, exact method sets — is not specified by any source document beyond the directive's
  own "proposed API, not current source" prose framing referenced in PLAN's DX-03 evidence. This story's
  own design work is the first place these types would be given a concrete shape; per mandate §18, this
  plan does not invent that shape in advance of the story's own design task actually being executed —
  the design document's content is this story's output, not a pre-determined fact this planning
  document already knows.
- The ADR-style decision record's exact format is expected to follow the same shape as the D-01..D-09
  decision records already produced at W00 (`impl/waves/wave-00-baseline-and-verification/epics/
  epic-002-baseline-capture/stories/story-003-adr-ification/decisions/`), though this story's own
  decision record is a story-produced design artifact, not a consumed programme-level D-0N decision —
  it is not added to the D-01..D-09 register, and no `decisions/` subdirectory is pre-created under this
  story for it (see "Required artifacts" below for where it is expected to live).

## Dependencies

Depends on W05's AR-01 (W05-E01) and AR-02 (W05-E02) reaching `accepted`, per PLAN DX-03-T0's own
dependency row ("Wave 1 (AR-01 ApplicationModel, AR-02 typed ports) complete"). No dependency within
W06-E01 — this story is the epic's first story alphabetically but has no code-level dependency on
W06-E01-S002 (DX-04), and the two may proceed in either order. Blocks W06-E02-S003 (REL-03b's T5 leg,
per MATRIX CS-15's "Blocked on DX-03/AR-03" framing) — though W06-E02-S003's T5 leg also depends on
W05-E03 (AR-03 remainder), so this story alone does not fully unblock it.

## Affected packages or components

None — this is a documentation/design-record story. No Go package is created or modified.

## Compatibility considerations

Not applicable — no code is produced, so no compatibility surface is affected.

## Security considerations

Not applicable — no code is produced.

## Performance considerations

Not applicable — no code is produced.

## Observability considerations

Not applicable — no code is produced.

## Migration considerations

Not applicable — no code, schema, or data change is produced.

## Documentation requirements

This story's entire output is documentation: a design document and an ADR-style decision record, both
explicitly labeled "target, not implemented." No other documentation-file update is required by this
story's own scope.

## Acceptance criteria

- **AC-W06-E01-S001-01**: A module-DSL design document exists, describing `port`, `Manifest[T]`, and
  `Operation[Request,Response]` at a level of detail sufficient for a future implementer to begin work
  without re-deriving the design from directive prose alone; the document's design-trade-off discussion
  references how the design relates to AR-01's `Registrar` capability and AR-02's typed `port.Key[T]`
  mechanism.
- **AC-W06-E01-S001-02**: An ADR-style decision record exists, formalizing the design document's chosen shape,
  explicitly and visibly labeled "target, not implemented" per AR-05's labeling convention; no
  implementation code, compiler, or runtime type-system change accompanies either document.

## Required artifacts

- The module-DSL design document (exact location TBD at implementation time — expected under
  `docs/blueprint/` or a new `docs/design/` location, consistent with where the framework's other
  design-facing documentation lives; this story's own plan does not pre-select the path per mandate
  §18).
- The ADR-style decision record (exact location TBD — this story's own output, not added to the
  W00-E02-S003 D-01..D-09 register since it is not a consumed programme-level architecture decision but
  a story-produced design artifact).
See `artifacts/index.md`.

## Required evidence

- None beyond the design document and decision record themselves — this is a design-investigation
  story with no test surface. The evidence that AC-W06-E01-S001-02's labeling requirement is met is the
  decision record's own visible label, inspectable directly, not a separate test-execution artifact.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, both acceptance criteria numbered and measurable, dependency on W05's AR-01/
AR-02 acceptance recorded, owner/reviewer assignment pending, the exact design-document/decision-record
storage location recorded as an unresolved question rather than silently assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation (i.e., the design document and decision record) matches `plan.md` or deviations are
recorded in `deviations.md`; both acceptance criteria verified with evidence in `evidence/index.md`;
`closure.md` completed. Given this story's P1 (not P0/critical) priority and its design-only nature, no
independent-review task is added per mandate §14's own scoping to critical stories — see
`tasks/index.md` "Grouping rationale" for why.

## Risks

None recorded at this story's own scope beyond the general risk (recorded at epic scope, not
story-specific) that a design document produced without implementation experience may need revision
once DX-03's eventual implementation (out of this programme's scope) actually begins — this is an
accepted, inherent property of design-before-implementation work, not a risk this story's own planning
can mitigate further.

## Residual-risk expectations

Once both acceptance criteria are met, no residual risk is expected — this is a bounded documentation
story with a clear, source-derived scope (PLAN DX-03-T0's own single-line task description) and no
implementation surface to leave incomplete.

## Plan

See `plan.md`.
