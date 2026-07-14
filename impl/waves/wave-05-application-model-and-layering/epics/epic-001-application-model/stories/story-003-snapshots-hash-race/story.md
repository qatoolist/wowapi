---
id: W05-E01-S003
type: story
title: Snapshot immutability, post-seal rejection, model hash, and race safety
status: planned
wave: W05
epic: W05-E01
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - AR-01
depends_on:
  - W05-E01-S002
blocks:
  - W05-E01-S004
acceptance_criteria:
  - AC-W05-E01-S003-01
  - AC-W05-E01-S003-02
  - AC-W05-E01-S003-03
artifacts: []
evidence: []
decisions: []
risks: []
---

# W05-E01-S003 — Snapshot immutability, post-seal rejection, model hash, and race safety

## Story ID

W05-E01-S003

## Title

Snapshot immutability, post-seal rejection, model hash, and race safety

## Objective

Convert every exported registry reader (across the registries S002 wrapped) to return
cloned/immutable data; reject Context retention after `Register()` returns with an explicit error
rather than a silent no-op or production panic; emit a deterministic model hash at startup/
readiness; and prove no runtime mutation of the sealed model under `go test -race`.

## Value to the framework

S001 and S002 establish the ownership-bound registration surface; this story closes the remaining
integrity properties that make the sealed `ApplicationModel` genuinely immutable and safe under
concurrency — without it, a caller could still mutate a registry's backing storage through a leaked
reference, or a retained registrar could silently no-op instead of erroring, undermining the
security and correctness properties the epic exists to establish.

## Problem statement

`requirement-inventory.md` row AR-01 groups this story's scope: "S003 snapshots-hash-race (T7, T9,
T10 + T8 post-seal rejection)." PLAN's own AR-01 task table: T7 — "Convert all snapshot-returning
reads to cloned/immutable data (`Specs()`, `Points()`, and equivalents on all registries) | T3-T6 |
No exported reader returns a backing map/slice | Unit: mutate returned value, assert registry
internal state unaffected | `AR-01/snapshot_immutability_test.go` | Low-medium." T8 — "Reject Context
retention after `Register()` returns | T1, T2 | A module retaining `ctx`/a registrar post-boot gets
an explicit error on mutation, never a silent no-op or a production panic | Adversarial: fixture
module retains registrar, calls it post-boot | `AR-01/post_seal_mutation_rejection_test.go` |
Medium — wowsociety's `policy` module already retains `mc.Rules()` today (harmlessly); this task has
a direct named consumer to validate against." T9 — "Deterministic model hash, emitted at
startup/readiness | T1-T8 | Two identical compiles → byte-identical hash; one changed declaration →
different hash | Unit: hash-determinism + hash-sensitivity tests |
`AR-01/model_hash_determinism_test.go` | Low — exclude non-deterministic inputs (map order,
timestamps)." T10 — "Race tests proving no runtime mutation of the sealed model | T1-T9 |
`go test -race` clean on concurrent legitimate reads; illegitimate write fails via T8, not a data
race | Race test | `AR-01/race_test_output.txt` | Low."

## Source requirements

AR-01 (T7, T8, T9, T10).

## Current-state assessment

Per this epic's own S002, the registries now have owner-bound registration, but their exported
readers have not yet been confirmed to return cloned/immutable data — this is a separate property
from ownership at registration time. No deterministic model hash and no race-test suite exist yet
for the sealed model, since the model itself does not yet fully exist prior to S001-S002. This
story's own re-confirmation step is to audit each wrapped registry's exported readers at this
story's actual start commit and confirm which, if any, still return a backing map/slice before
converting them.

## Desired state

No exported registry reader (`Specs()`, `Points()`, and equivalents on every registry S002 wrapped)
returns a backing map/slice — mutating a returned value does not affect registry internal state,
proven by `AR-01/snapshot_immutability_test.go`. A module retaining a registrar or the compiler's
`ctx` past `Register()` returning gets an explicit error on any post-boot mutation attempt, never a
silent no-op or a production panic — validated specifically against wowsociety's own
`internal/modules/policy/pack.go:334-338`'s retained `s.rulesReg` field, PLAN's own named real-world
consumer for this exact pattern. Two identical compiles of the `ApplicationModel` emit a
byte-identical hash; one changed declaration emits a different hash, excluding non-deterministic
inputs (map order, timestamps). `go test -race` is clean on concurrent legitimate reads of the
sealed model; an illegitimate write fails via T8's rejection mechanism, not as an unguarded data
race.

## Scope

- Conversion of every exported registry reader to cloned/immutable data across all registries S002
  wrapped (T7).
- Post-seal Context/registrar retention rejection — explicit error, never silent no-op or production
  panic (T8), validated against wowsociety's `s.rulesReg` retention pattern as a named real consumer.
- A deterministic model hash function, emitted at startup/readiness, excluding non-deterministic
  inputs (T9).
- Race tests proving no runtime mutation of the sealed model, with illegitimate writes failing via
  T8's rejection path, not as unguarded races (T10).

## Out of scope

- **The per-registry ownership wrappers themselves** — S002's scope, already built; this story
  converts their readers to immutable and validates their post-seal behavior.
- **The legacy compatibility adapter** — S004's scope, built after this story so it can wrap a
  model that is already immutable, deterministic, and race-safe.
- **Any change to the model hash's use in readiness reporting's HTTP/API surface** — this story
  produces the hash function itself; wiring it into a readiness endpoint's response shape, if not
  already covered by existing readiness infrastructure, is an implementation-time integration detail
  within this story's own T9 task, not a separate story.

## Assumptions

- T8's validation target — wowsociety's `internal/modules/policy/pack.go:334-338`'s retained
  `s.rulesReg` field — is taken as a confirmed real-world fixture from PLAN's own wowsociety-impact
  note for AR-01: "written once, never read again (dead code) — precisely the 'retained registrar'
  pattern T8 targets; must be dropped or replaced before wowsociety adopts the non-legacy v1
  registrar API... T8 must not reject" the legitimately-used `s.rulesStore`/`s.rulesResolver`
  distinction. This story's own adversarial test must distinguish the two, per PLAN's own explicit
  warning.
- T9's "exclude non-deterministic inputs (map order, timestamps)" is taken as a confirmed design
  constraint from PLAN's own risk column, not an invented detail this story adds.

## Dependencies

Depends on W05-E01-S002 (T7 depends on T3-T6's wrappers existing; T9-T10 depend on the fuller T1-T8/
T1-T9 surface). Blocks W05-E01-S004 (T11's legacy adapter wraps a model this story has made
immutable, deterministic, and race-safe).

## Affected packages or components

The registries wrapped in S002 (`kernel/resource`, `kernel/rules`, `kernel/authz`, and the ~9+
remaining declaration classes) — their exported reader methods (`Specs()`, `Points()`, equivalents);
the `ApplicationModel`/`Compiler` from S001 (model-hash function, post-seal rejection enforcement
point).

## Compatibility considerations

T8's post-seal rejection must not reject wowsociety's legitimately-used `s.rulesStore`/
`s.rulesResolver` (built over the registry, used live in request handlers) while correctly rejecting
the dead `s.rulesReg` retention — this distinction is explicitly named in PLAN's own risk note and
must be preserved by this story's adversarial test design, not silently collapsed into "reject
everything retained."

## Security considerations

T8's error-not-panic behavior (per D-03, referenced from S001) is itself a security-adjacent
availability control: a retained registrar that panicked in production would convert a benign
pattern (present in real code today) into a production crash. T7's immutability conversion prevents
a caller from mutating registry internal state through a leaked reference, a data-integrity concern
adjacent to but distinct from S002's ownership-at-registration concern.

## Performance considerations

T7's clone-on-read conversion has a performance cost proportional to registry size — acceptable
given these are boot-time/readiness-time reads, not request-hot-path reads (the framework's
hot-path-performance concerns are addressed separately, e.g. AR-02 T3's zero-reflection
requirement). This story's own plan should confirm this assumption holds for whichever registries
turn out to be read on a hot path, if any, at implementation time.

## Observability considerations

The deterministic model hash (T9) is itself an observability artifact — emitted at
startup/readiness so an operator or automated check can confirm two deployments compiled an
identical model, or detect an unexpected model change.

## Migration considerations

None — no schema or data migration.

## Documentation requirements

Document the immutability guarantee for exported registry readers, the post-seal rejection contract
(referencing D-03), and the model hash's determinism guarantees and readiness-reporting integration.

## Acceptance criteria

- **AC-W05-E01-S003-01**: No exported registry reader returns a backing map/slice — mutating a
  returned value does not affect registry internal state, proven by
  `AR-01/snapshot_immutability_test.go`.
- **AC-W05-E01-S003-02**: A module retaining a registrar or `ctx` past `Register()` returning gets
  an explicit error on post-boot mutation, never a silent no-op or a production panic — proven by
  `AR-01/post_seal_mutation_rejection_test.go`, validated specifically against wowsociety's
  `s.rulesReg` retention pattern without rejecting the legitimately-used `s.rulesStore`/
  `s.rulesResolver`.
- **AC-W05-E01-S003-03**: Two identical compiles emit a byte-identical model hash; one changed
  declaration emits a different hash — proven by `AR-01/model_hash_determinism_test.go`; `go test
  -race` is clean on concurrent legitimate reads, with illegitimate writes failing via T8's
  rejection mechanism, not as an unguarded data race — proven by `AR-01/race_test_output.txt`.

## Required artifacts

- Snapshot-immutability conversion (code, across all S002-wrapped registries).
- Post-seal Context/registrar retention rejection (code).
- The deterministic model-hash function (code).
- Race-test suite (code).
See `artifacts/index.md`.

## Required evidence

- `AR-01/snapshot_immutability_test.go` output.
- `AR-01/post_seal_mutation_rejection_test.go` output.
- `AR-01/model_hash_determinism_test.go` output.
- `AR-01/race_test_output.txt`.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on S002 recorded,
owner/reviewer assignment pending, T8's wowsociety-specific validation target recorded explicitly.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all three acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
per mandate §14 applied at epic-scope discretion given this story's Low-Medium risk profile relative
to S001/S002 — see epic `risks.md` for the epic's own risk-scaling rationale.

## Risks

No dedicated wave/epic-level risk entry — this story's task-level risks (PLAN's own Low-Medium
column values for T7/T9/T10, Medium for T8) are lower than S001/S002's High-risk items and are
tracked at this story's own task level, not escalated to epic/wave risk registers.

## Residual-risk expectations

Residual risk is expected to be low once T8's wowsociety-specific distinction (reject `s.rulesReg`,
do not reject `s.rulesStore`/`s.rulesResolver`) is proven by its own adversarial test, and once T10's
race test is confirmed clean.

## Plan

See `plan.md`.
