---
id: PLAN-W06-E04-S002
type: plan
parent_story: W06-E04-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W06-E04-S002

Per mandate §8.5. T4's own implementation strategy cannot be fully specified until AR-03's delivered
shape (W05-E03) is known — this is stated explicitly rather than invented. Confirmed facts, planned
changes, and assumptions are distinguished explicitly below.

## Proposed architecture

Two independent mechanisms: T4, a reference-doc-generation pipeline consuming AR-03's own authoritative
model export (once it exists) and producing generated reference tables, verified against the model
export via an integration golden-diff test. T5, a lint scanning `docs/blueprint/` for normative-sounding
prose blocks lacking a "target, not implemented" (or equivalent) label.

## Implementation strategy

**For T4** (once W05-E03 is `accepted`):
1. Confirm AR-03's own delivered model-export format (W05-E03's own output).
2. Build a reference-doc-generation pipeline consuming that model export.
3. Write an integration golden-diff test proving the generated reference tables byte-match the model
   export.

**For T5** (independent of T4):
1. Review `docs/blueprint/` for the current set of future-state design blocks and how they are (or are
   not) labeled today.
2. Build a lint detecting a normative-sounding future-state block lacking the "target, not implemented"
   (or equivalent) label.
3. Write fixture tests: an unlabeled block fails; a correctly-labeled block passes.

## Expected package or module changes

T4: a new reference-doc-generation pipeline (exact location TBD, dependent on AR-03's own delivered
package structure). T5: a new lint tool (exact location TBD).

## Expected file changes where determinable

Not fully determinable for T4 at this planning stage, per mandate §18 — its exact file surface depends
on AR-03's own delivered shape. T5's file changes are expected to be a new lint tool plus its fixture
tests, independent of AR-03.

## Contracts and interfaces

T4's own contract (the model-export format it consumes) is AR-03's own delivered contract, consumed
here, not designed here.

## Data structures

T4: none new beyond what AR-03's model export already defines. T5: none new.

## APIs

None affected.

## Configuration changes

None.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None.

## Error-handling strategy

T5's lint must report clearly which specific unlabeled block triggered the failure and its exact
location.

## Security controls

None new.

## Observability changes

None beyond T5's own clear-failure-reporting requirement.

## Testing strategy

- T4: integration golden-diff test proving generated reference tables byte-match the model export.
- T5: fixture tests — unlabeled block fails, correctly-labeled block passes.

## Regression strategy

T4, once wired in, becomes the ongoing regression guard against generated-doc drift from the model
export. T5 becomes the ongoing regression guard against a future unlabeled future-state block.

## Compatibility strategy

Not applicable.

## Rollout strategy

T5 may land independently of T4, as soon as its own implementation is complete; T4 lands once W05-E03
is `accepted`.

## Rollback strategy

If T5's lint produces false positives against a legitimate non-future-state block, revise the lint's
detection logic; do not silently narrow its scope without recording why.

## Implementation sequence

T5 may proceed immediately. T4 begins only once W05-E03 reaches `accepted`.

## Task breakdown

- **W06-E04-S002-T001** — Generated reference docs byte-matching AR-03's model export (T4), blocked on W05-E03.
- **W06-E04-S002-T002** — Future-state-labeling lint (T5), independent of T4.
- **W06-E04-S002-T003** — Independent review (scoped to whichever of T4/T5 actually completed within this
  story's execution window).

## Expected artifacts

The reference-doc-generation pipeline (T4, once unblocked); the future-state-labeling lint (T5).

## Expected evidence

Integration golden-diff test output (T4); lint fixture test output (T5).

## Unresolved questions

- T4's exact pipeline mechanism (a `go generate` directive, a standalone tool, a Makefile target) — not
  specified by any source document.
- Whether T5's lint scope should extend beyond `docs/blueprint/` (e.g. to this programme's own `impl/`
  tree, which itself produces future-state-labeled content such as W06-E01-S001's DX-03 design record)
  — PLAN's own acceptance criterion scopes T5 to `docs/blueprint/` specifically; extending it further is
  an implementation-time scoping decision, not assumed here.

## Approval conditions

This plan is approved for implementation on a per-task basis: T5's implementation is approved once the
owner and reviewer are assigned; T4's implementation is approved once W05-E03 additionally reaches
`accepted`.
