---
id: PLAN-W06-E02-S003
type: plan
parent_story: W06-E02-S003
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W06-E02-S003

Per mandate §8.5. This story's plan is unusual in that it cannot fully specify T5's implementation
strategy until AR-03's own delivered shape (W05-E03) is known — this is stated explicitly rather than
invented. Confirmed facts, planned changes, and assumptions are distinguished explicitly below.

## Proposed architecture

Three independent compatibility-gate mechanisms, each consuming a specific upstream story's `accepted`
output: T3 consumes W06-E02-S001's (DX-06) merge-completeness closure to build a semantic diff over it;
T5 consumes AR-03's (W05-E03) typed-model shape (and, nominally, DX-03's design record, though DX-03
remains design-only) to build an event/schema compatibility check; T7 consumes W06-E01-S002's (DX-04)
golden-consumer fixture and reuses its upgrade-replay drill for a REL-03-specific invocation.

## Implementation strategy

**For T3** (once W06-E02-S001 is `accepted`):
1. Build a semantic-diff mechanism specifically for OpenAPI documents, classifying breaking changes per
   DX-06's own 3.1/2020-12 baseline (established by W06-E02-S001's T2).
2. Write a seeded breaking-OpenAPI fixture and confirm the gate fails it.

**For T5** (once both W06-E01-S001 and W05-E03 are `accepted`):
1. Determine, from AR-03's actual delivered shape, what compatibility-mode concept (e.g.
   `CompatibilityBackward`) exists to tie an event/schema compatibility check to.
2. Build the compatibility check against that concept.
3. Write a seeded breaking-event fixture and confirm the gate fails it when the compatibility mode is
   declared.

**For T7** (once W06-E01-S002 is `accepted`):
1. Reuse DX-04's own upgrade-replay drill (W06-E01-S002's T4) rather than building a second one.
2. Wire a REL-03-specific invocation of that drill, confirming golden-consumer contracts re-pass after
   an N-1-to-N upgrade.

## Expected package or module changes

T3: an extension within or alongside W06-E02-S001's own semantic-diff CI job. T5: a new package/CI job
(exact shape TBD pending AR-03's delivered form). T7: a CI-job wrapper invoking W06-E01-S002's existing
drill, not a new drill implementation.

## Expected file changes where determinable

Not determinable in full at this planning stage for T5, per mandate §18 — its exact file surface
depends on AR-03's own delivered shape, not yet known. T3 and T7's file changes are expected to be small
CI-configuration additions layered on their respective unblocking stories' own artifacts.

## Contracts and interfaces

T5's own contract (a `Compatibility` mode declaration) is not this story's to define — it is expected to
be AR-03's own delivered contract, consumed here, not designed here.

## Data structures

None new expected for T3/T7. T5's data structures (if any) depend on AR-03's own delivered shape.

## APIs

None affected directly — these are CI-time gates.

## Configuration changes

None anticipated beyond each gate's own CI-job configuration.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None.

## Error-handling strategy

Each gate must fail with a clear, specific error, consistent with the pattern established across this
wave's other compatibility gates.

## Security controls

None new beyond what each unblocking story already establishes.

## Observability changes

None beyond each gate's own clear-failure-reporting requirement.

## Testing strategy

- T3: seeded breaking-OpenAPI fixture, confirming the gate fails it.
- T5: seeded breaking-event fixture, confirming the gate fails it when the compatibility mode is
  declared.
- T7: golden-consumer N-1-to-N upgrade re-pass, reusing DX-04's own two-pass test structure.

## Regression strategy

Once each leg is implemented and wired into CI, it becomes an ongoing regression guard for its
respective compatibility class.

## Compatibility strategy

All three legs are themselves compatibility-enforcement mechanisms, each consuming an upstream story's
already-established compatibility baseline (DX-06's 3.1/2020-12 baseline for T3; AR-03's typed-model
shape for T5; DX-04's upgrade-replay drill for T7).

## Rollout strategy

Each leg lands independently, as soon as its own entry criterion is satisfied — no requirement that all
three land simultaneously.

## Rollback strategy

If a leg is implemented and later found to be premature (its unblocking story's `accepted` status is
later found to have been granted in error, or the unblocking story's own scope is later found
insufficient to support this leg), revert the leg's implementation and record the reversion as a
deviation, restating the correct entry criterion.

## Implementation sequence

No fixed sequence across T3/T5/T7 — each begins independently once its own entry criterion is
satisfied, in whatever order those criteria happen to be met.

## Task breakdown

- **W06-E02-S003-T001** — OpenAPI semantic diff (T3), entry-gated on W06-E02-S001.
- **W06-E02-S003-T002** — Event/schema compatibility (T5), entry-gated on W06-E01-S001 AND W05-E03.
- **W06-E02-S003-T003** — Generated-consumer upgrade check (T7), entry-gated on W06-E01-S002.
- **W06-E02-S003-T004** — Independent review (scoped to whichever legs actually complete within this story's
  execution window).

## Expected artifacts

The OpenAPI semantic-diff gate (T3); the event/schema compatibility-check mechanism (T5); the generated-
consumer upgrade check invocation (T7) — each produced only once its leg's entry criterion is satisfied.

## Expected evidence

Seeded breaking-OpenAPI-fixture test output (T3); seeded breaking-event-fixture test output (T5);
generated-consumer N-1-to-N upgrade re-pass output (T7).

## Unresolved questions

- T5's exact implementable scope pending AR-03's delivered shape — this is the story's own central
  unresolved question, explicitly not answerable until W05-E03 lands.
- Whether T3/T5/T7 land in the same PR/change as each other or independently as each unblocks — expected
  to be independently, per "Rollout strategy" above, but not yet confirmed.

## Approval conditions

This plan is approved for implementation on a per-leg basis: T3's implementation is approved once
W06-E02-S001 reaches `accepted`; T5's once both W06-E01-S001 and W05-E03 reach `accepted`; T7's once
W06-E01-S002 reaches `accepted`. The plan as a whole (this document) is approved for existence once the
owner and reviewer are assigned — it does not require all three entry criteria satisfied simultaneously
to be a valid plan.
