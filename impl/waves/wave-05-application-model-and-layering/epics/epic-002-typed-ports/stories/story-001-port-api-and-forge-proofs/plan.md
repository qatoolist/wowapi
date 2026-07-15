---
id: PLAN-W05-E02-S001
type: plan
parent_story: W05-E02-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W05-E02-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below.

## Proposed architecture

A `port` package defines `Key[T]` (a phantom-typed key identifying a provided value's type) and four
generic free functions (`Define`, `Provide`, `Require`, `Resolve`), each taking a `Registrar` (reused
from W05-E01-S001, not a new type) as a first-class argument — since Go lacks type-parameterized
methods, the API is necessarily function-based rather than method-based on a generic receiver.

## Implementation strategy

1. Re-read the current ad hoc wiring mechanism at this story's start commit to confirm no
   `port.Key[T]` API exists.
2. Design and implement `port.Key[T]`.
3. Implement `Define`, `Provide`, `Require`, `Resolve` as generic free functions bound to a
   `Registrar`.
4. Extend the internal compiler factory (from W05-E01-S001) to mint registrars for AR-02's own
   port-key registration flow, reusing the same minting mechanism, not duplicating it.
5. Write the happy-path round-trip unit test (`AR-02/port_api_unit_test.go`).
6. Write the adversarial compile-fail fixture (`AR-02/registrar_forge_compile_fail_fixture/`):
   simulated module code attempting to manufacture a `Registrar` from a bare string, specifically
   probing the shared-`Registrar`-type scenario where a capability minted for a resource-registration
   purpose (AR-01) might be misusable for a port-registration purpose (AR-02), or vice versa.
7. Document the API and the safety proof.

## Expected package or module changes

A new `port` package (exact location TBD).

## Expected file changes where determinable

- New `port.Key[T]` type definition file.
- New `Define`/`Provide`/`Require`/`Resolve` function file(s).
- New unit test and compile-fail fixture files as named above.

## Contracts and interfaces

`port.Key[T]` (phantom-typed); `Define[T](r Registrar, key Key[T], ...)`, `Provide[T](...)`,
`Require[T](...)`, `Resolve[T](...)` (exact signatures TBD at implementation time, per PLAN's own
note that Go's lack of type-parameterized methods forces a first-class-argument API).

## Data structures

None new beyond `Key[T]`'s own internal representation (a phantom type, likely a small struct
carrying an identifier and a zero-sized type parameter).

## APIs

Internal Go generic API only; no external HTTP/gRPC surface.

## Configuration changes

None.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None material at this story's own scope — registration-time API, not a runtime concurrency concern
(S002's own T3/T4 address the compiled graph's runtime behavior).

## Error-handling strategy

The compile-fail fixture's expected outcome is a compile error, not a runtime error — this is a
Go-level capability-security proof, following the same pattern as W05-E01-S001-T002's own
compile-fail fixture.

## Security controls

T2's registrar-forge safety proof is the required security control this story delivers.

## Observability changes

None material.

## Testing strategy

- `AR-02/port_api_unit_test.go`: happy-path define/provide/resolve round-trip.
- `AR-02/registrar_forge_compile_fail_fixture/`: adversarial fixture proving module code cannot
  manufacture a `Registrar`, specifically probing the shared-`Registrar`-type capability-confusion
  scenario.

## Regression strategy

The compile-fail fixture is a permanent regression guard, matching the pattern established by
W05-E01-S001-T002.

## Compatibility strategy

Not applicable — no existing consumer.

## Rollout strategy

Single story, landed as its own reviewable unit.

## Rollback strategy

Revert if the compile-fail fixture reveals a capability-confusion bypass — escalate for redesign of
the port-key minting mechanism rather than silently narrowing the fixture's own adversarial scope.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-7).

## Task breakdown

- **W05-E02-S001-T001** — `port.Key[T]` and the four generic free functions (steps 2-3, 5 above).
- **W05-E02-S001-T002** — Internal compiler factory extension and the registrar-forge compile-fail
  fixture (steps 4, 6 above).
- **W05-E02-S001-T003** — Independent review (per mandate §14, scoped to this story, given T2's
  High-risk capability-confusion proof).

## Expected artifacts

`port.Key[T]` and the four generic functions (code); the compiler factory extension (code); port-API
documentation.

## Expected evidence

`AR-02/port_api_unit_test.go` output; `AR-02/registrar_forge_compile_fail_fixture/` output.

## Unresolved questions

- Exact function signatures for `Define`/`Provide`/`Require`/`Resolve` — PLAN's own risk note flags
  "ergonomics review needed," implying the exact API shape is not fully specified by the source and
  is this story's own design work.
- Exact `port` package location.

## Approval conditions

This plan is approved for implementation once: (a) the API's ergonomics are reviewed (per T1's own
risk note), and (b) the owner and reviewer are assigned.
