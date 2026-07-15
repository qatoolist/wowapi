---
id: PLAN-W06-E01-S001
type: plan
parent_story: W06-E01-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W06-E01-S001

Per mandate §8.5. This is a design-investigation story: its "implementation strategy" is a design and
documentation process, not a code-change process. Confirmed facts, planned changes, and assumptions are
distinguished explicitly below.

## Proposed architecture

Not applicable in the code-architecture sense — this story produces no code. The "architecture" this
story addresses is documentation architecture: where the design document and decision record live, and
how they relate to AR-01/AR-02's already-`accepted` typed-registration mechanisms they build on.

## Implementation strategy

1. Re-read the directive's own module-DSL prose (the source of the "proposed API, not current source"
   framing PLAN's DX-03 evidence cites) and AR-01/AR-02's `accepted` implementation (W05-E01, W05-E02)
   to ground the design in what the framework's typed-registration mechanisms actually provide today.
2. Draft the module-DSL design's options and trade-offs: the shape of `port` (a typed capability
   descriptor, building on AR-02's `port.Key[T]`), `Manifest[T]` (a typed per-module declaration
   surface, building on AR-01's `Registrar`), and `Operation[Request,Response]` (a typed request/
   response contract for a generated or hand-written operation).
3. Select a design and document the rationale, including how it relates to AR-01/AR-02's existing
   mechanisms (this design is explicitly framed by PLAN as "formalizes the type-system-level version"
   of what AR-01/AR-02/DX-02 already target from the correctness-fix side — it is evolutionary, not a
   replacement).
4. Formalize the selected design into an ADR-style decision record, explicitly labeled "target, not
   implemented" per AR-05's labeling convention.
5. Confirm both documents are correctly labeled such that AR-05 T5's future-state-labeling lint (once
   it exists, W06-E04-S002) would not flag them.

## Expected package or module changes

None — no Go package is created or modified by this story.

## Expected file changes where determinable

- A new design document (exact path TBD — see `story.md` "Required artifacts").
- A new ADR-style decision record (exact path TBD — see `story.md` "Required artifacts").

## Contracts and interfaces

The design document's own content *describes* proposed contracts and interfaces (`port`, `Manifest[T]`,
`Operation[Request,Response]`) — but describing a future contract is not the same as defining one in
code; this story's own artifact is prose/design content, not a Go type declaration.

## Data structures

Not applicable — no code is produced.

## APIs

Not applicable — no code is produced.

## Configuration changes

Not applicable.

## Persistence changes

Not applicable.

## Migration strategy

Not applicable.

## Concurrency implications

Not applicable.

## Error-handling strategy

Not applicable — no code is produced.

## Security controls

Not applicable — no code is produced.

## Observability changes

Not applicable.

## Testing strategy

Not applicable in the code-test sense. The only "test" this story's output must pass is inspection: the
decision record must visibly carry the "target, not implemented" label, and its content must be
sufficiently detailed that a future implementer would not need to re-derive the design from directive
prose alone — both are confirmed by direct reading at verification time, not by an automated test.

## Regression strategy

Not applicable — no code is produced, so no regression surface exists.

## Compatibility strategy

Not applicable.

## Rollout strategy

Single story, landed as its own reviewable documentation change — no phased rollout.

## Rollback strategy

If the design document or decision record is found to be materially wrong after being produced (e.g. a
later implementer discovers the design does not actually build cleanly on AR-01/AR-02's real shape),
revise the documents directly — a documentation artifact does not require the same rollback discipline
as a code change, since it has no runtime behavior to roll back.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–5).

## Task breakdown

- **W06-E01-S001-T001** — Draft the module-DSL design options and trade-offs (steps 1–3 above).
- **W06-E01-S001-T002** — Formalize the design into a labeled, ADR-style decision record and confirm labeling
  correctness (steps 4–5 above).

## Expected artifacts

The module-DSL design document; the ADR-style decision record, explicitly labeled "target, not
implemented."

## Expected evidence

None beyond the two documents themselves (see `story.md` "Required evidence").

## Unresolved questions

- Exact storage location for the design document and decision record (a new `docs/design/` location, or
  under existing `docs/blueprint/`) — to be decided at implementation time.
- The exact internal shape of `port`, `Manifest[T]`, and `Operation[Request,Response]` (field lists,
  generic constraints, method sets) — this is the story's own design-task output, not a fact this
  planning document can state in advance.

## Approval conditions

This plan is approved for implementation once: (a) W05's AR-01/AR-02 have reached `accepted` (this
story's upstream dependency), and (b) the owner and reviewer are assigned.
