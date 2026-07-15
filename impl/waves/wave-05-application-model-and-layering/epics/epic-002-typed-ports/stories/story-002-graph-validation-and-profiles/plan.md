---
id: PLAN-W05-E02-S002
type: plan
parent_story: W05-E02-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W05-E02-S002

Per mandate §8.5.

## Proposed architecture

A type-erased provider graph built over S001's `port.Key[T]` API, with type-erasure occurring only
at compile/boot time (never at `Resolve` call time). Boot-time validation walks the graph checking
for the five named failure classes, reusing `kernel/lifecycle`'s existing scope-rank ordering logic
rather than reimplementing it. A projection compiler derives API/worker/migrate profile subsets from
the one graph.

## Implementation strategy

1. Design the provider graph's internal representation such that `Resolve` dispatch does not require
   `reflect.*` at call time (e.g. via generated/cached type-safe accessors set up once at boot).
2. Implement the graph.
3. Write `AR-02/hotpath_no_reflection_bench.txt`'s producing benchmark, and a static lint check
   specifically scanning for `reflect.*` calls on the `Resolve` code path.
4. Study `kernel/lifecycle`'s existing scope-rank ordering logic to identify the reusable component.
5. Implement boot-time graph validation for the five failure classes (duplicate providers, missing
   requirements, undeclared edges, cycles, invalid scope/lifetime edges), reusing the identified
   `kernel/lifecycle` logic.
6. Write `AR-02/boot_graph_validation_test.go`: one adversarial fixture per failure class, asserting
   error messages name both owners.
7. Implement the three-profile projection compiler.
8. Write `AR-02/three_profile_projection_test.go`: build all three profiles from one fixture, assert
   capability subsets.
9. Document all three components.

## Expected package or module changes

A new provider-graph package; `kernel/lifecycle` (read/reused, not yet removed).

## Expected file changes where determinable

New provider-graph implementation files; new validation files; new projection-compiler files; new
benchmark, lint, and test files as named above.

## Contracts and interfaces

The provider graph's own internal type (exact shape TBD); the projection compiler's output type
(profile-scoped capability subset, exact shape TBD).

## Data structures

The compiled provider graph's internal representation (exact shape TBD, designed for zero-reflection
dispatch).

## APIs

None externally facing.

## Configuration changes

None anticipated.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

The compiled graph, once built at boot, must be safe for concurrent `Resolve` calls from multiple
goroutines (consistent with the framework's general boot-then-serve concurrency model) — this
story's own benchmark and tests should confirm this holds under the zero-reflection design.

## Error-handling strategy

T4's validation errors must name both owners involved in a failure (e.g. both the duplicate
provider's registering modules, or both ends of an undeclared edge) — a specific, diagnosability-
driven error-message requirement from PLAN's own acceptance criterion.

## Security controls

None new beyond this epic's existing capability-security posture (S001).

## Observability changes

None beyond the named-both-owners error-message requirement above.

## Testing strategy

- `AR-02/hotpath_no_reflection_bench.txt`: benchmark proving zero-reflection dispatch.
- Static lint check for `reflect.*` calls on the `Resolve` path.
- `AR-02/boot_graph_validation_test.go`: one adversarial fixture per failure class.
- `AR-02/three_profile_projection_test.go`: all three profiles built from one fixture, capability
  subsets asserted.

## Regression strategy

The lint check is itself a permanent regression guard against reintroduced hot-path reflection; the
adversarial validation suite guards against a reintroduced unvalidated failure class.

## Compatibility strategy

T4's reuse of `kernel/lifecycle`'s scope-rank ordering avoids introducing a second, divergent
ordering scheme.

## Rollout strategy

Single story, landed as its own reviewable unit, sequenced after S001.

## Rollback strategy

Revert the provider-graph implementation if the benchmark/lint reveals hot-path reflection that
cannot be readily redesigned away — escalate for a redesign of the dispatch mechanism.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-9).

## Task breakdown

- **W05-E02-S002-T001** — Type-erased provider graph with zero-hot-path-reflection proof (T3; steps
  1-3 above).
- **W05-E02-S002-T002** — Boot-time graph validation, reusing `kernel/lifecycle`'s scope-rank
  ordering (T4; steps 4-6 above).
- **W05-E02-S002-T003** — Three-profile projection compiler (T5; steps 7-8 above).

No independent-review task is added for this story — PLAN's own risk column values (Medium for all
three tasks) are lower than S001's High-risk T2, and this story's own zero-reflection and
validation-completeness properties are proven by dedicated benchmark/lint/adversarial-suite
mechanisms rather than resting on code-review confidence alone.

## Expected artifacts

The provider graph (code); boot-time validation (code); the three-profile projection compiler
(code).

## Expected evidence

The three named test/benchmark outputs.

## Unresolved questions

- Exact mechanism for achieving zero-reflection dispatch (generated code, cached type-safe closures,
  or another approach) — to be decided at implementation time.
- Exact reusable component within `kernel/lifecycle`'s scope-rank ordering logic — to be identified
  by this story's own step-4 study, not assumed in advance.

## Approval conditions

This plan is approved for implementation once: (a) the zero-reflection dispatch mechanism is chosen,
and (b) the owner and reviewer are assigned.
