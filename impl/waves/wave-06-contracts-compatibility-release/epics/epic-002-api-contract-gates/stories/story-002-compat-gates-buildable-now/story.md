---
id: W06-E02-S002
type: story
title: Compatibility gates buildable now — REL-03a (Go API diff, compile matrix, config compat, migration drill, arch smoke, SBOM verify)
status: accepted
wave: W06
epic: W06-E02
owner: W06E02Impl
reviewer: W06-E02-S002-Rerun
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-14
source_requirements:
  - REL-03
depends_on: []
blocks: []
acceptance_criteria:
  - AC-W06-E02-S002-01
  - AC-W06-E02-S002-02
  - AC-W06-E02-S002-03
  - AC-W06-E02-S002-04
  - AC-W06-E02-S002-05
  - AC-W06-E02-S002-06
artifacts:
  - ART-W06-E02-S002-001
  - ART-W06-E02-S002-002
  - ART-W06-E02-S002-003
  - ART-W06-E02-S002-004
  - ART-W06-E02-S002-005
  - ART-W06-E02-S002-006
evidence:
  - EV-W06-E02-S002-001
  - EV-W06-E02-S002-002
  - EV-W06-E02-S002-003
  - EV-W06-E02-S002-004
  - EV-W06-E02-S002-005
  - EV-W06-E02-S002-006
  - EV-W06-E02-S002-007
decisions: []
risks: []
---

# W06-E02-S002 — Compatibility gates buildable now — REL-03a

## Story ID

W06-E02-S002

## Title

Compatibility gates buildable now — REL-03a (Go API diff, compile matrix, config compat, migration
drill, arch smoke, SBOM verify)

## Objective

Build the six REL-03 compatibility-gate tasks that carry no unresolved upstream architectural
dependency and are buildable today: Go public API diff (T1), module compile matrix (T2), config-schema
compatibility (T4), migration upgrade-from-oldest-supported drill (T6), container architecture smoke
(T8), and SBOM/provenance/signature verification folded in from REL-01 T8/T9 (T9).

## Value to the framework

PLAN's own text makes the case for splitting REL-03 explicitly: "Recommend splitting into REL-03a
(buildable now...) and REL-03b (hard-blocked on Wave 1/4 architecture work...) — do not schedule as one
monolithic P1 item, or 5 of 9 sub-tasks silently block the other 4." This story is REL-03a: six of
REL-03's nine tasks that have no dependency on DX-03, DX-04, or DX-06 and can be built, tested, and
evidenced against today's codebase, without waiting for this wave's own E01/E02-S001 work to land. Its
value is that a genuinely blocked minority of REL-03's scope (three tasks, W06-E02-S003) does not hold
hostage the majority that is ready today — a direct application of this programme's own doability-over-
theoretical-completeness planning principle.

## Problem statement

`requirement-inventory.md` row REL-03 states: "Compatibility gates (split a/b) | QG | P1 | planned |
W06-E02-S002..S003 | a=T1,T2,T4,T6,T8,T9 now; b=T3(DX-06),T5(AR-03/DX-03),T7(DX-04)." MATRIX CS-15
confirms the exact same split: "REL-03 split into REL-03a/REL-03b **per the plan's own recommendation**
(`premier-framework-implementation-plan.md:694`): **REL-03a (buildable now) = T1, T2, T4, T6, T8, T9**
(Go API diff via `golang.org/x/exp/apidiff`/`gorelease`, compile matrix, config compat, migration drill,
arch smoke, SBOM-verify fold-in)." Today, none of these six mechanisms exist: PLAN's own REL-03
evidence confirms zero `apidiff`/`gorelease` hits anywhere in Makefile/CI/docs, and no compile matrix,
config-compat gate, migration upgrade-drill extension, architecture-smoke-on-candidate-image job, or
SBOM/provenance fold-in exists today beyond what REL-01 T8/T9 (this wave's W06-E03-S001) itself builds.

## Source requirements

REL-03 (T1, T2, T4, T6, T8, T9 — the REL-03a buildable-now subset).

## Current-state assessment

Per PLAN's own evidence (to be re-confirmed at this story's own execution commit): zero
`apidiff`/`gorelease` hits anywhere in Makefile/CI/docs (T1). No documented module compile matrix exists
(T2). `kernel/config/schema.go` is the existing source of truth for config schema, but no compatibility
gate against it exists (T4). `TestIntegrationMigrationsReversible` exists but has not been extended into
an upgrade-from-oldest-supported drill (T6). No container-architecture-smoke job runs against the
*candidate* image specifically (T8; REL-01 T6/T7's build-candidate split, W06-E03-S001, must exist
first for this to run against the right artifact). No SBOM/provenance/signature verification exists yet
independent of REL-01 T8/T9's own build (T9 folds directly into that work per PLAN's own framing: "not
separate work, just REL-03's naming of a property REL-01 already builds").

## Desired state

A Go public API diff tool (`apidiff`/`gorelease`) is wired as a CI job, correctly classifying
added/removed/changed exported symbols per DX-05's already-ratified v1/N-1 policy. A module compile
matrix runs across supported Go/dependency versions with explicit, not silent, exclusions. A config
schema compatibility gate fails a field removal/type change and passes additive optional fields,
against `kernel/config/schema.go` as source of truth. A migration upgrade-from-oldest-supported-version
drill, extending `TestIntegrationMigrationsReversible`, seeds at the oldest supported version, migrates
forward, and reverses on disposable data. A container architecture smoke test runs against every
published architecture, specifically against the *candidate* image produced by REL-01's build-candidate
stage, not an already-published one. SBOM/provenance/signature verification for REL-03's own naming
purposes is satisfied by REL-01 T8/T9's existing golden-failure tests, shared as evidence, not
re-implemented.

## Scope

- **T1** — Go public API diff tool (`apidiff`/`gorelease`) wired as a CI job; correctly classifies
  added/removed/changed exported symbols against DX-05's v1/N-1 policy; a seeded breaking-API fixture
  fails the gate.
- **T2** — Module compile matrix across supported Go/dependency versions; excluded versions are
  explicit, not silently ignored.
- **T4** — Config schema compatibility gate against `kernel/config/schema.go`; field removal/type
  change fails; additive optional fields pass; a generated fixture migration test.
- **T6** — Migration upgrade-from-oldest-supported drill, extending
  `TestIntegrationMigrationsReversible`; seed at oldest version, migrate forward, reverse on disposable
  data.
- **T8** — Container architecture smoke on every published architecture, run against the *candidate*
  image in the `build-candidate` stage (not an already-published image).
- **T9** — SBOM/provenance/signature verification, folded in from REL-01 T8/T9 as a shared-evidence
  naming exercise, not separate implementation.

## Out of scope

- **T3, T5, T7 (REL-03b)** — the three still-blocked legs; W06-E02-S003's own scope, not duplicated
  here.
- **REL-01's own T6/T7 build-candidate split mechanics** — W06-E03-S001's scope; T8 here depends on that
  split existing to have a candidate image to smoke-test against, but does not itself build the split.
- **DX-05's own v1/N-1 policy design** — already ratified at W01; T1/T6 consume it, they do not
  re-derive it.

## Assumptions

- T8's dependency on REL-01 T6/T7 (the `build-candidate` split, W06-E03-S001) means this task cannot be
  fully proven end-to-end until that story's own artifact-producing mechanism exists — this story's
  plan records T8 as sequenced after or in parallel with W06-E03-S001 rather than strictly before it,
  since both stories may proceed concurrently once each satisfies its own upstream dependencies; the
  exact coordination point (does T8 wait for W06-E03-S001's `accepted` status, or can it be developed
  against a stub candidate image and reconciled later) is recorded as an unresolved question, not
  silently assumed.
- T9's exact evidence-sharing mechanism with REL-01 T8/T9 (a shared evidence path, `REL-01/verify_
  release/`, per MATRIX CS-15's own citation) is confirmed from source — PLAN REL-03 T9's row states
  explicitly: "Same acceptance as REL-01 T8/T9 | Same golden-failure tests | `REL-01/verify_release/`
  (shared)." This story's T9 task does not duplicate that evidence path, it references it.

## Dependencies

None within W06-E02 for T1, T2, T4, T6 (these five have no dependency on this epic's S001 or S003).
T8 depends cross-story on W06-E03-S001 (REL-01 T6/T7's build-candidate split, needed to have a candidate
image to smoke-test). T9 shares evidence with, but does not depend on completing before, W06-E03-S001's
own REL-01 T8/T9 work — see "Assumptions" above.

## Affected packages or components

New CI jobs for T1 (API diff), T2 (compile matrix), T4 (config compat), T8 (arch smoke); an extension
to the existing `TestIntegrationMigrationsReversible` test for T6; T9 references REL-01's own artifacts,
touching no new package.

## Compatibility considerations

T1, T4, and T6 are themselves compatibility-enforcement mechanisms — they exist specifically to prevent
an unintentional breaking change from shipping undetected. This story does not itself introduce a
compatibility-breaking change to any existing surface.

## Security considerations

T9's SBOM/provenance/signature verification is itself a supply-chain security control, shared with
REL-01. No new security surface is introduced beyond what REL-01 T8/T9 already covers.

## Performance considerations

Not applicable — these are CI-time gates, not runtime request paths. T2's compile matrix and T8's
architecture smoke may have meaningful CI wall-clock cost (multiple Go versions, multiple
architectures, arm64 via QEMU per PLAN's own risk note: "arm64 via QEMU is slow, consider native
runners") — a practical CI-budget consideration, not a runtime performance concern.

## Observability considerations

Each gate should report clearly which specific check failed (which exported symbol changed, which Go
version excluded and why, which config field is incompatible, which migration step failed the drill,
which architecture failed the smoke test) so a developer can diagnose a CI failure without re-deriving
the gate's own internal logic.

## Migration considerations

T6 is itself a migration-testing mechanism; it does not perform any migration of its own beyond the
disposable-data drill it runs as a test.

## Documentation requirements

Document each of the six gates' purpose, invocation, and failure-diagnosis guidance, so a developer
encountering a REL-03a gate failure in CI has a clear path to understanding and fixing it.

## Acceptance criteria

- **AC-W06-E02-S002-01**: The Go public API diff tool correctly classifies added/removed/changed exported
  symbols; a seeded breaking-API fixture fails the gate.
- **AC-W06-E02-S002-02**: The module compile matrix runs across supported Go/dependency versions; excluded
  versions are explicit in CI configuration, not silently ignored.
- **AC-W06-E02-S002-03**: A seeded breaking-config fixture (field removal or type change) fails the config
  compatibility gate; additive optional fields pass.
- **AC-W06-E02-S002-04**: The migration upgrade-from-oldest-supported drill seeds at the oldest supported
  version, migrates forward, and reverses on disposable data, extending
  `TestIntegrationMigrationsReversible`.
- **AC-W06-E02-S002-05**: Each published container architecture boots and passes minimal smoke before
  `publish`, run against the candidate image (not an already-published one).
- **AC-W06-E02-S002-06**: SBOM/provenance/signature verification for REL-03's own naming purpose is satisfied
  by REL-01 T8/T9's shared evidence, correctly cross-referenced, not re-implemented.

## Required artifacts

- The Go API diff CI job (T1).
- The module compile matrix CI configuration (T2).
- The config schema compatibility gate (T4).
- The extended migration upgrade-drill test (T6).
- The container architecture smoke CI job (T8).
- The REL-03 T9 cross-reference to REL-01's shared SBOM/provenance evidence (T9).
See `artifacts/index.md`.

## Required evidence

- Seeded breaking-API-fixture test output (T1).
- Compile-matrix CI run output, with explicit exclusions (T2).
- Seeded breaking-config-fixture test output (T4).
- Migration upgrade-drill test output (T6).
- Architecture-smoke test output per published architecture (T8).
- The shared REL-01 T8/T9 evidence reference (T9).
See `evidence/index.md`.

## Definition of ready

Confirmed before implementation: `story.md` and `plan.md` were complete, all six acceptance criteria
were numbered and measurable, the T8 cross-story dependency was recorded, and owner/reviewer were assigned.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all six acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming T9's shared-evidence cross-reference to REL-01 T8/T9 is
accurate and not a duplicated re-implementation claim.

## Risks

None recorded at this story's own scope beyond the general CI-wall-clock-cost consideration noted under
"Performance considerations" — this is a well-bounded, source-derived closure story with a clear MATRIX
CS-15/PLAN REL-03 acceptance bar for each of its six tasks.

## Residual-risk expectations

T8's candidate dependency is satisfied and all six gates are verified; residual story risk is low.

## Plan

See `plan.md`.
