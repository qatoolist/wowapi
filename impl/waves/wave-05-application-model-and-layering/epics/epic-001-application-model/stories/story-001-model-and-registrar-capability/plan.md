---
id: PLAN-W05-E01-S001
type: plan
parent_story: W05-E01-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W05-E01-S001

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below. D-02 and D-03 are treated as confirmed, ratified fact (from
`docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F); this plan does not invent
the typed-key's exact field-level design, which D-02 does not itself specify.

## Proposed architecture

A new `ApplicationModel` type and its `Compiler` accumulate module declarations through owner-bound
calls, validate, and seal into an immutable snapshot. A `Registrar` capability type — the security
boundary — is a single generic type (per D-02) minted only by the `Compiler` from a module's
`Manifest.ID`/`Module.Name()`, carrying an unexported seal method so module code cannot construct or
type-assert one for another owner. Per-subsystem capability confusion is prevented by typed keys
(`Key[T]`-shaped) bound to the `Registrar`, not by multiplying `Registrar` types, per D-02's explicit
resolution.

## Implementation strategy

1. Re-read `kernel/module`'s current registration surface at this story's actual start commit,
   confirming the current-state assessment (no lifecycle skeleton, no capability type) still holds.
2. Define the `ApplicationModel` type and its `collect → validate → seal → expose read-only
   snapshot` state machine.
3. Define the `Compiler` type: accumulates declarations via owner-bound calls; exposes `Compile()`
   which validates then seals.
4. Implement post-seal mutation behavior per D-03: error in production builds; panic only under an
   explicit dev/test build tag (Go build-tag-gated code path).
5. Define the `Registrar` capability type per D-02: one generic type, unexported seal method, minted
   only by the `Compiler` from `Manifest.ID`/`Module.Name()`.
6. Design the per-subsystem typed-key mechanism (`Key[T]`-shaped) that binds to the `Registrar`,
   preventing capability confusion across subsystems without multiplying `Registrar` types.
7. Write the state-machine transition unit tests (collecting → validating → sealed; post-seal
   mutation attempts).
8. Write the compile-fail fixture: a deliberate attempt, from simulated module code, to construct or
   type-assert a `Registrar` for another owner — must fail to compile.
9. Write the build-tag-scoped test confirming production builds error (not panic) and the explicit
   dev/test build tag panics (not silently errors) post-seal.
10. Document the lifecycle, the capability type, and the D-02/D-03 decisions this story enacts.

## Expected package or module changes

A new `ApplicationModel`/`Compiler` type and a new `Registrar` capability type (exact package
location TBD — expected within or adjacent to `kernel/module`, per "Unresolved questions" below). No
existing package is removed.

## Expected file changes where determinable

- A new file (or files) defining `ApplicationModel`, `Compiler`, and their lifecycle state machine
  (exact path TBD).
- A new file defining the `Registrar` capability type and its typed-key mechanism (exact path TBD).
- A new compile-fail fixture (a package intentionally excluded from the normal build, or a
  `// want`-style negative-compilation test harness, exact mechanism TBD at implementation time).
- New unit tests for the state-machine transitions and the D-03 build-tag-scoped behavior.

## Contracts and interfaces

`ApplicationModel` (immutable, read-only snapshot exposure); `Compiler` (accumulation + `Compile()`
entry point); `Registrar` (capability type, unexported seal method); per-subsystem typed `Key[T]`
(exact generic signature TBD, informed by D-02 but not dictated by it).

## Data structures

The `ApplicationModel`'s internal accumulated-declaration state (exact shape TBD, expected to be
extended by S002's per-registry ownership wrappers rather than fully specified here, since this
story defines the lifecycle skeleton, not every declaration class's own storage).

## APIs

None externally facing (no HTTP/gRPC surface) — this is internal Go type/API design consumed by
module registration code.

## Configuration changes

None.

## Persistence changes

None.

## Migration strategy

Not applicable — no schema or data migration.

## Concurrency implications

The sealed `ApplicationModel`'s read-only snapshot must be safe for concurrent access from multiple
goroutines once sealed (a requirement this story's own race tests do not yet prove — that is S003's
explicit scope, T10). This story's own scope is the lifecycle skeleton and capability type; it must
not introduce a data race in the collect/validate/seal transition itself, proven by this story's own
state-machine unit tests running under `-race`.

## Error-handling strategy

Post-seal mutation in production builds returns an explicit, typed error (not a generic error string)
identifying the attempted mutation as rejected-post-seal, per D-03. The dev/test build-tag panic path
must be clearly separated (Go build constraint) from the production error path, not a single code
path with a runtime environment check — this is the mechanism D-03 itself implies ("panic only under
an explicit dev/test build tag").

## Security controls

The `Registrar`'s unexported seal method and compiler-only minting are themselves the required
security control (PLAN T2's own risk note: "this is the actual security boundary"). This is not
optional hardening — it is the acceptance-criterion-defining property of this story.

## Observability changes

Boot-time log lines for the `collect → validate → seal` transitions (implementation-time addition,
not separately mandated by PLAN's own T1/T2 acceptance criteria — see `story.md` "Observability
considerations").

## Testing strategy

- State-machine transition unit tests: collecting → validating → sealed; a post-seal mutation
  attempt errors in production build configuration.
- Build-tag-scoped test: the explicit dev/test build tag panics post-seal; the default (production)
  build errors, never panics.
- Compile-fail fixture: simulated module code attempting to construct or type-assert a `Registrar`
  for another owner fails to compile — this fixture is itself the primary security proof for T2.

## Regression strategy

The compile-fail fixture, once established, is itself a permanent regression guard: any future
change that accidentally exposes a `Registrar`-construction path outside the compiler would need to
either fix the fixture (a reviewable, visible change) or the fixture would start compiling
unexpectedly, itself a signal.

## Compatibility strategy

Not applicable at this story's own scope — this story adds new types with no existing consumer; S004
is the compatibility story that wires the legacy path onto these types.

## Rollout strategy

Single story, landed as its own reviewable unit — no existing code consumes these types yet, so no
phased rollout is required at this story's own scope.

## Rollback strategy

Revert this story's types if the compile-fail fixture or build-tag-scoped test reveals the
capability-type design does not hold under the pattern S002's ownership wrappers need — escalate for
redesign rather than silently loosening the `Registrar`'s seal method's unexported-ness, which would
defeat the security property entirely.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1-10). Step 5 (the `Registrar` type) must be
designed before step 6 (the typed-key mechanism), since the typed key binds to the `Registrar`, not
the reverse.

## Task breakdown

- **W05-E01-S001-T001** — `ApplicationModel`/`Compiler` lifecycle skeleton and post-seal error/panic
  behavior per D-03 (steps 2-4, 7, 9 above).
- **W05-E01-S001-T002** — `Registrar` capability type and typed-key mechanism per D-02, plus the
  compile-fail fixture (steps 5-6, 8 above).
- **W05-E01-S001-T003** — Independent review (per mandate §14, scoped to this story, given T2's
  security-boundary status).

## Expected artifacts

The `ApplicationModel`/`Compiler` lifecycle skeleton (code); the `Registrar` capability type and
typed-key mechanism (code); lifecycle and capability-type documentation.

## Expected evidence

State-machine transition unit-test output; build-tag-scoped error/panic test output; compile-fail
fixture output.

## Unresolved questions

- Exact package location for the new `ApplicationModel`/`Compiler`/`Registrar` types (within
  `kernel/module` or a new adjacent package) — to be decided at implementation time.
- Exact typed-key (`Key[T]`) generic signature and how it binds to the `Registrar` — D-02 resolves
  "one `Registrar` type, typed keys" at the design-decision level; the concrete Go type signature is
  this story's own implementation-time design work.
- Exact compile-fail fixture mechanism (a permanently-excluded package, a `// want`-style negative
  compilation harness, or another approach) — to be chosen at implementation time.
- Whether the `ApplicationModel`'s internal accumulated-declaration storage is fully specified here
  or extended incrementally as S002's per-registry wrappers are added — leaning toward the latter
  (this story defines the lifecycle skeleton and capability type; S002 defines what gets stored per
  declaration class), but not yet confirmed.

## Approval conditions

This plan is approved for implementation once: (a) the exact package location and typed-key generic
signature are resolved (informed by, not contradicting, D-02's ratified resolution), and (b) the
owner and reviewer are assigned.
