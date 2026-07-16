---
id: W05-E01-S001
type: story
title: ApplicationModel lifecycle skeleton and Registrar capability type
status: planned
wave: W05
epic: W05-E01
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-01
depends_on: []
blocks:
  - W05-E01-S002
  - W05-E01-S003
  - W05-E01-S004
  - W05-E02-S001
  - W05-E03-S001
acceptance_criteria:
  - AC-W05-E01-S001-01
  - AC-W05-E01-S001-02
  - AC-W05-E01-S001-03
artifacts: []
evidence: []
decisions:
  - D-02
  - D-03
risks:
  - RISK-W05-E01-S001-001
---

# W05-E01-S001 — ApplicationModel lifecycle skeleton and Registrar capability type

## Story ID

W05-E01-S001

## Title

ApplicationModel lifecycle skeleton and Registrar capability type

## Objective

Define the `ApplicationModel` type and its `collect → validate → seal → expose read-only snapshot`
lifecycle skeleton, and the owner-bound `Registrar` capability type — minted only by the compiler
from `Manifest.ID`/`Module.Name()`, with an unexported seal method — so that module code can never
construct or type-assert a `Registrar` for another owner. This story enacts D-02 (one generic
owner-bound `Registrar` type, per-subsystem typed keys) and D-03 (post-seal mutation errors in
production, panics only under an explicit dev/test build tag).

## Value to the framework

Every later task in this epic (S002's per-registry ownership wrappers, S003's snapshot immutability
and post-seal rejection, S004's legacy adapter) and every later epic in this wave (E02's typed
ports, E03's authoritative manifest) is built directly on this story's two types. PLAN's own
PF-ARCH cross-cutting note states this explicitly: "AR-01 T1/T2 are the load-bearing prerequisite
for AR-02's `Registrar` reuse and AR-03's manifest-consumes-model dependency." Today, per PLAN's
own directive requirement, the framework has no immutable compiled model and no capability-typed
ownership boundary — registration APIs can accept an arbitrary owner string from module code, and
nothing structurally prevents one module from claiming another's declaration. This story replaces
that gap with the actual security boundary PLAN's own risk column names T2 as being.

## Problem statement

`requirement-inventory.md` row AR-01 states: "Ownership-bound `ApplicationModel` (T1–T11) | IMPL |
P1 | planned | W05-E01-S001..S004 | Core; D-02/D-03 resolve its open questions." PLAN's own AR-01
task table: T1 — "Define `ApplicationModel` type + `collect→validate→seal→expose` lifecycle
skeleton | Wave 0 exit | A `Compiler` accumulates declarations via owner-bound calls only;
`Compile()` validates then seals; post-seal, further calls error (panic only in an explicit dev/test
build tag) | Unit: state-machine transition tests | `AR-01/lifecycle_test_output.txt` | Medium —
load-bearing type every other AR-0x task depends on." T2 — "Owner-bound `Registrar` capability type
(unexported seal method) minted only by the compiler from `Manifest.ID`/`Module.Name()` | T1 |
Module code cannot construct/type-assert a `Registrar` for another owner | Compile-fail fixture
attempting to fabricate a `Registrar` | `AR-01/registrar_capability_test_output.txt` | High — this
is the actual security boundary." PLAN's own PF-ARCH cross-cutting note (5) frames the unresolved
design question T2 raises: "do all AR-01 per-subsystem registrars share one `Registrar` type
(capability-confusion risk) or does each get a distinct type (multiplies T2/T6 task count)?" — this
is precisely the question D-02 resolves. Note (6) frames the second unresolved question: "should
post-seal mutation panic in production builds, or only error? Recommend 'error, not panic' as the
default — wowsociety's harmless `s.rulesReg` retention would otherwise convert into a production
crash risk" — this is precisely the question D-03 resolves.

## Source requirements

AR-01 (T1, T2). D-02, D-03 (enacted; referenced from W00-E02-S003, not re-decided here).

## Current-state assessment

Per PLAN's own evidence, no `ApplicationModel` type, no `collect → validate → seal` lifecycle, and
no owner-bound `Registrar` capability type exist anywhere in wowapi today — module registration
proceeds through the current mutable `module.Context`, and registration APIs accept owner
identification (where checked at all) as an arbitrary string rather than a structurally-secured
capability. This is a confirmed absence, not a partial implementation. This story's own
re-confirmation step (per this programme's fail-first convention applied elsewhere, e.g.
W01-E01-S001, W02-E01-S001) is to re-read `kernel/module`'s current registration surface at this
story's actual start commit and confirm no lifecycle skeleton or capability type yet exists before
building one from zero.

## Desired state

A `Compiler` type accumulates module declarations via owner-bound calls only; `Compile()` validates
then seals the accumulated state into an immutable `ApplicationModel`; post-seal, further
registration calls return an error in production builds (per D-03), panicking only under an
explicit dev/test build tag. A `Registrar` capability type exists as a single generic type (per
D-02) with an unexported seal method, mintable only by the compiler from a module's
`Manifest.ID`/`Module.Name()`; module code has no path to construct or type-assert a `Registrar` for
an owner other than its own, proven by a compile-fail fixture that deliberately attempts to fabricate
one.

## Scope

- The `ApplicationModel` type definition and its `collect → validate → seal → expose read-only
  snapshot` lifecycle skeleton (state-machine transitions: collecting → validating → sealed).
- The `Compiler` type that accumulates declarations via owner-bound calls and exposes `Compile()`.
- Post-seal mutation error behavior (production) and panic behavior (explicit dev/test build tag
  only), per D-03.
- The `Registrar` capability type: a single generic type (per D-02) with an unexported seal method,
  minted only by the compiler from `Manifest.ID`/`Module.Name()`.
- Per-subsystem typed keys (`Key[T]`-shaped, per D-02's "typed keys rather than per-subsystem
  registrar types" resolution) as the mechanism preventing capability confusion across subsystems
  sharing the one `Registrar` type — the exact typed-key shape is this story's own design work,
  informed by D-02's resolution but not fully specified by it (D-02 resolves "one type vs. many
  types," not the typed-key's own field-level design).
- The compile-fail fixture proving module code cannot construct or type-assert a `Registrar` for
  another owner.

## Out of scope

- **Per-registry ownership wrappers for resource, rules, authz-permission, and the ~9+ remaining
  declaration classes** — W05-E01-S002's scope. This story produces the `Registrar` type those
  wrappers consume; it does not itself wrap any specific registry.
- **Snapshot immutability conversion, deterministic model hash, and race tests** — W05-E01-S003's
  scope, built on this story's sealed model.
- **The legacy compatibility adapter** — W05-E01-S004's scope.
- **AR-02's typed port-key API** — a separate epic (W05-E02) that reuses this story's `Registrar`
  type; not built here.

## Assumptions

- D-02's resolution ("one generic owner-bound `Registrar` capability type, with per-subsystem typed
  keys rather than per-subsystem registrar types... capability confusion is prevented by the key's
  phantom type + owner binding, not by multiplying registrar types") is taken as ratified fact from
  `docs/implementation/fable5-final-architecture-review-2026-07-11.md` §F item 3, referenced (not
  re-derived) via this story's `decisions/index.md`.
- D-03's resolution ("error in production builds; panic only under an explicit `dev`/test build
  tag") is taken as ratified fact from the same source, §F item 4, referenced via
  `decisions/index.md`.
- The exact shape of the per-subsystem typed key (`Key[T]`'s own field/phantom-type design) is not
  fully specified by D-02 itself — D-02 resolves the one-type-vs-many-types question, not the typed
  key's own concrete Go type signature. This story's plan records the exact typed-key design as an
  implementation-time decision informed by, but not dictated by, D-02.

## Dependencies

None within W05-E01 — this is the epic's first story and its own foundation. Depends on this wave's
entry gate (W03-E01 acceptance). Blocks W05-E01-S002, S003, S004 (all consume this story's
`Registrar` type and lifecycle skeleton); blocks W05-E02-S001 (AR-02 T1 reuses this story's
`Registrar` directly) and W05-E03-S001 (AR-03's manifest-derived-projection tooling depends on the
model this story defines) at wave scope.

## Affected packages or components

New: the `ApplicationModel` type and `Compiler` (expected location: `kernel/module` or an adjacent
new package — exact location TBD per `plan.md`'s "Unresolved questions"); the `Registrar` capability
type. No existing package is removed by this story; the legacy `module.Module`/`Context` surface
remains in place until S004's adapter and downstream module migration.

## Compatibility considerations

This story does not itself migrate any existing module — S004's legacy adapter is the compatibility
mechanism. This story's own compile-fail fixture is a new-code test, not a change to any existing
compiling code path, so it introduces no compatibility risk on its own.

## Security considerations

This is, per PLAN's own risk column for T2, "the actual security boundary." The `Registrar`
capability type's unexported seal method and compiler-only minting are the entire security property
this story establishes — a Go-level capability-security pattern (a value that cannot be constructed
or forged outside its issuing authority) rather than a runtime string comparison. The compile-fail
fixture (a fixture that deliberately attempts to construct/type-assert a `Registrar` for another
owner and is expected to fail to compile) is the primary proof mechanism, per T2's own "Tests"
column.

## Performance considerations

None material — this is boot-time/compile-time machinery, not a request-hot-path concern. AR-02
T3's own zero-hot-path-reflection requirement (a downstream, not this story's own, concern) confirms
the broader system's performance posture but is not this story's own scope.

## Observability considerations

The `Compiler`'s `collect → validate → seal` transitions should be observable at boot (log-level, at
minimum) for operator diagnosability — a reasonable implementation-time addition, not separately
mandated by PLAN's own AR-01 T1/T2 acceptance criteria beyond the state-machine transition tests
themselves.

## Migration considerations

None — this story adds new types; it does not migrate any existing schema, data, or module.

## Documentation requirements

Document the `ApplicationModel` lifecycle (collect/validate/seal/expose), the `Registrar` capability
type's minting and typed-key mechanism, and the D-02/D-03 decisions this story enacts, so a future
module author or reviewer understands why the API is shaped this way without re-reading the ADRs
directly.

## Acceptance criteria

- **AC-W05-E01-S001-01**: A `Compiler` accumulates declarations via owner-bound calls only;
  `Compile()` validates then seals; post-seal, further calls error in production builds — proven by
  unit state-machine transition tests (`AR-01/lifecycle_test_output.txt`).
- **AC-W05-E01-S001-02**: Post-seal mutation panics only under an explicit dev/test build tag, never
  in a production build — per D-03, proven by a build-tag-scoped test confirming the error/panic
  split.
- **AC-W05-E01-S001-03**: Module code cannot construct or type-assert a `Registrar` for another
  owner — proven by a compile-fail fixture attempting to fabricate a `Registrar`
  (`AR-01/registrar_capability_test_output.txt`), per D-02's single-type-with-typed-keys design.

## Required artifacts

- The `ApplicationModel` type and `Compiler` lifecycle skeleton (code).
- The `Registrar` capability type and its typed-key mechanism (code).
- Lifecycle and capability-type documentation.
See `artifacts/index.md`.

## Required evidence

- State-machine transition unit-test output (`AR-01/lifecycle_test_output.txt`).
- Compile-fail fixture output (`AR-01/registrar_capability_test_output.txt`).
- Build-tag-scoped error-vs-panic test output (D-03 proof).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, D-02/D-03 referenced (not
re-derived) via `decisions/index.md`, owner/reviewer assignment pending, the typed-key's exact
design recorded as an implementation-time decision rather than silently assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the compile-fail fixture genuinely fails to compile
(not merely asserted) and that D-02/D-03 are enacted as ratified, not reinterpreted.

## Risks

RISK-W05-E01-S001-001 (the `Registrar` capability type, once shipped, becomes the security boundary
every later ownership wrapper and every downstream module depends on — a flaw here propagates to
every consumer) — see epic-level `risks.md` for full detail and mitigation/contingency.

## Residual-risk expectations

Once the compile-fail fixture and build-tag-scoped D-03 test are executed as planned and confirmed
by independent review, residual risk is expected to be low — this is a foundational, well-bounded,
source-derived story whose two open design questions are already resolved by ratified ADRs (D-02,
D-03), removing the largest source of ambiguity a story like this would otherwise carry.

## Plan

See `plan.md`.

## Note (autopsy remediation R-1, 2026-07-16)

Status is unchanged — this story remains genuinely unexecuted as tracked (`planned`, all tasks
`todo`). However, the implementation-autopsy report
(`impl/reports/implementation-autopsy-report-2026-07-16.md`, §4 row W05-E01-S001, independent
verdict **contradictory**) found that related code has, in substance, already landed outside this
story's execution: a real `port.Key[T]` API and AR-01/AR-02 code (`kernel/appmodel`,
`kernel/port`, ~765 LOC, tested) exist but are built-but-not-wired (autopsy H-6), and the FBL-01
kernel re-home this programme's W05-E05 covers was executed on `main` while W05 tracking stayed
`planned` (autopsy H-7). See deviation **DEV-PROG-002** in
`impl/tracking/programme-deviations.md` for the full record. — autopsy remediation R-1,
2026-07-16.
