---
id: W06-E01-S002
type: story
title: Golden consumer matrix — framework-repo-owned CLI/generator proof fixture
status: accepted
wave: W06
epic: W06-E01
owner: W06E01Impl
reviewer: W06-E01-S002-Verify
priority: P1
created_at: 2026-07-12
updated_at: 2026-07-14
source_requirements:
  - DX-04
depends_on:
  - W01-E04-S001
blocks:
  - W06-E02-S003
acceptance_criteria:
  - AC-W06-E01-S002-01
  - AC-W06-E01-S002-02
  - AC-W06-E01-S002-03
  - AC-W06-E01-S002-04
  - AC-W06-E01-S002-05
artifacts:
  - ART-W06-E01-S002-001
  - ART-W06-E01-S002-002
  - ART-W06-E01-S002-003
  - ART-W06-E01-S002-004
  - ART-W06-E01-S002-005
evidence:
  - EV-W06-E01-S002-001
  - EV-W06-E01-S002-002
  - EV-W06-E01-S002-003
  - EV-W06-E01-S002-004
  - EV-W06-E01-S002-005
  - EV-W06-E01-S002-006
  - EV-W06-E01-S002-007
  - EV-W06-E01-S002-008
  - EV-W06-E01-S002-009
  - EV-W06-E01-S002-010
  - EV-W06-E01-S002-011
  - EV-W06-E01-S002-012
  - EV-W06-E01-S002-013
  - EV-W06-E01-S002-014
decisions: []
risks:
  - RISK-W06-E01-001
---

# W06-E01-S002 — Golden consumer matrix — framework-repo-owned CLI/generator proof fixture

## Story ID

W06-E01-S002

## Title

Golden consumer matrix — framework-repo-owned CLI/generator proof fixture

## Objective

Build a framework-repo-owned golden-consumer fixture — installed via `go install`, not a repo-internal
import — that exercises resource, rule, workflow, event handler, recurring job, document flow,
notification, and webhook generation across at least two modules; boots API and worker processes
against real Postgres/MinIO/Mailpit/OTel; exercises authenticated CRUD, async delivery, restart/retry,
and RLS isolation; replays an upgrade-from-previous-version cycle; and is wired into CI as a required
gate.

## Value to the framework

PLAN's own DX-04 evidence explains precisely why this fixture must exist and why wowsociety cannot
substitute for it: "`internal/testmodules/requests/` is a single hand-authored in-process reference
module used for framework unit tests — not a CLI-scaffolded, installed-binary, two-module, upgrade-
tested consumer. Zero CI workflow references `wowapi init`/`gen crud`." Without this fixture, the
framework's own generator/CLI surface — the exact surface W01-E04-S001 already proved does not silently
produce dead-on-arrival output for a single resource — has never been exercised end-to-end across
multiple modules, multiple subsystems, real infrastructure, or an actual version upgrade. This story
converts "the generator produces text that individually inspects correctly" into "the generator
produces a multi-module product that boots, serves traffic, and survives an upgrade" — the only claim a
generator-driven framework should be allowed to make about its own consumer experience.

## Problem statement

`requirement-inventory.md` row DX-04 states: "Golden consumer + upgrade matrix | IMPL | P1 | planned |
W06-E01-S002 | Dep DX-01 T5." PLAN's own DX-04 evidence gives three concrete, independently-verified
reasons wowsociety cannot serve as this fixture: "(1) wowsociety consumes wowapi via a sibling-checkout
path `replace`, never an installed CLI binary — step 1 of the directive's own procedure ('Install the
built CLI') is never exercised. (2) `FRAMEWORK_VERSION` is a raw SHA with no 'previous supported
version' concept and `framework-verify` only diffs SHA equality — step 7 ('upgrade from previous
version, rerun contracts') has no mechanism. (3) wowsociety is a real product with domain-specific logic
(committee seats, OTP/TOTP, citation packs) — coupling wowapi's own release gate to wowsociety's roadmap
changes would violate the 'non-internal consumer fixture' intent." PLAN's own conclusion: "wowapi needs
its own separate, framework-repo-owned golden-consumer fixture, distinct from wowsociety."

## Source requirements

DX-04 (T1–T5).

## Current-state assessment

Per PLAN's own evidence (to be re-confirmed at this story's own execution commit): no framework-repo-
owned, `go install`-based, multi-module golden-consumer fixture exists today. `internal/testmodules/
requests/` is a single hand-authored in-process reference module used only for framework unit tests —
it is never installed as a binary, never boots against real infrastructure as part of a CI gate, and is
not upgrade-tested. Zero CI workflow references `wowapi init`/`gen crud` anywhere in the repository.

## Desired state

A CI-required fixture that: installs the CLI via `go install` (not a repo-internal import, proving the
released-binary experience, not merely the source-tree experience); generates a resource, rule,
workflow, event handler, recurring job, document flow, notification, and webhook across at least two
modules; boots the generated API and worker processes against real Postgres, MinIO, Mailpit, and OTel;
exercises authenticated CRUD, async delivery, restart/retry, and RLS isolation, all passing; replays an
upgrade-from-previous-version cycle (fixture generated at N-1, upgraded to N, contracts rerun and
passing); and runs as a required CI gate.

## Scope

- **T1** — Build the framework-repo-owned golden-consumer scaffold job, reusing DX-01 T5's isolated-
  temp-dir subprocess-scaffold harness (W01-E04-S001) as the shared primitive rather than building a
  second one; the fixture installs via `go install`.
- **T2** — Generate a resource, rule, workflow, event handler, recurring job, document flow,
  notification, and webhook across two modules, each subsystem exercised at least once.
- **T3** — Boot the API and worker processes against real Postgres/MinIO/Mailpit/OTel; exercise
  authenticated CRUD, async delivery, restart/retry, and RLS isolation.
- **T4** — Replay an upgrade-from-previous-version cycle: fixture at N-1 (per DX-05's already-ratified
  v1/N-1 policy), upgraded to N, contracts rerun and passing.
- **T5** — Wire the fixture into CI as a required gate, appearing in `ci/release-gates.yaml` at its
  Wave-4 boundary (REL-01 — this story only adds the gate reference; REL-01's own manifest mechanics are
  W06-E03-S001's scope).

## Out of scope

- **DX-01's own version-verification flags and fallback removal** — already built and accepted at
  W01-E04-S001; this story consumes that harness, it does not re-implement any part of it.
- **wowsociety as a golden-consumer substitute** — explicitly ruled out per PLAN's own three
  disqualifying reasons (see "Problem statement"); this fixture is framework-repo-owned, not
  wowsociety-based.
- **REL-01's own gate-manifest mechanics** (`ci/release-gates.yaml` schema, `required-gates.yml`,
  `build-candidate` split) — W06-E03-S001's scope; T5 here only adds this fixture's own entry to that
  manifest, it does not build the manifest mechanism.
- **REL-03 T7's generated-consumer upgrade check** — PLAN's own note: "Reuses DX-04's drill." REL-03 T7
  (W06-E02-S003) consumes this story's T4 upgrade-replay mechanism as a downstream dependency; it is not
  built here.

## Assumptions

- The "previous supported version" concept T4's upgrade replay requires is defined by DX-05's already-
  ratified v1/N-1 policy (W01-E04-S002, executed) — this story consumes that policy, it does not
  re-derive what "N-1" means.
- The exact two modules generated for T2's cross-module coverage (their names, domain content) are not
  specified by any source document — PLAN's own T2 acceptance criterion only requires "each subsystem
  exercised ≥1 time" across two modules, not a specific domain. This story's own plan records the exact
  module content as an implementation-time decision, consistent with mandate §18.
- T5's exact manifest-entry shape (what `ci/release-gates.yaml` requires of a Wave-4-boundary entry) is
  not yet defined until W06-E03-S001 builds the manifest schema — this story's own T5 task is recorded
  as depending on that schema existing, or, if sequencing requires T5 to land first, as adding a
  forward-compatible placeholder entry to be reconciled once the schema exists. This ambiguity is
  recorded as an unresolved question in `plan.md`, not silently resolved.

## Dependencies

Depends on W01-E04-S001 (the DX-01 T5 isolated-temp-dir scaffold harness, this story's own T1 shared
primitive) and, transitively, on this wave's own W05 entry gate. Within W06-E01, no dependency on
W06-E01-S001 (DX-03) — the two stories target disjoint surfaces. Blocks W06-E02-S003's T7 leg (REL-03's
generated-consumer upgrade check, "Hard-blocked on DX-04" per PLAN's own REL-03 T7 dependency row).

## Affected packages or components

The framework-owned fixture lives in `internal/cli/golden_consumer_test.go` and
`internal/cli/golden_consumer_infra_test.go`. `Makefile`, `.github/workflows/ci.yml`,
`.github/workflows/required-gates.yml`, and `ci/release-gates.yaml` wire it into both ordinary CI and
the exact-SHA required-gates runner.

## Compatibility considerations

None beyond what T4's upgrade replay itself tests — the fixture is additive CI infrastructure, it does
not change any existing public API or CLI flag surface.

## Security considerations

Not directly applicable — the fixture exercises existing authenticated-CRUD/RLS-isolation paths without
introducing new authorization logic. The fixture's own generated modules should not introduce a new
security-relevant surface beyond what the generator itself already produces (which is DX-01/DX-02's own
already-closed scope, not re-litigated here).

## Performance considerations

Not applicable at the CI-gate level — the fixture's own CI run time is a practical consideration
(installing a CLI, generating two modules, booting real infrastructure, running an upgrade replay is a
non-trivial CI job), but no performance SLO applies to this story's own acceptance bar.

## Observability considerations

The fixture boots against real OTel per T3's own scope — this is itself an observability-adjacent
exercise (proving the generated module's OTel wiring works), not a separate observability requirement
this story adds beyond what T3 already covers.

## Migration considerations

T4's upgrade-from-previous-version replay is itself a migration-adjacent exercise (the fixture's own
schema/state moving from N-1 to N), but this story does not introduce any new migration mechanism beyond
what the framework's existing migration tooling (and, where relevant, DATA-09's protocol from W02)
already provides — the fixture consumes existing migration tooling, it does not build new migration
tooling.

## Documentation requirements

Document how the golden-consumer fixture is structured, how to run it locally, and how it fits into the
CI gate — so a future contributor extending generator coverage knows where the fixture lives and how to
add a new subsystem to its exercised set.

## Acceptance criteria

- **AC-W06-E01-S002-01**: The golden-consumer fixture installs via `go install`, not a repo-internal import.
- **AC-W06-E01-S002-02**: A resource, rule, workflow, event handler, recurring job, document flow, notification,
  and webhook are each generated and exercised at least once, across at least two modules.
- **AC-W06-E01-S002-03**: The fixture boots API and worker processes against real Postgres, MinIO, Mailpit, and
  OTel; authenticated CRUD, async delivery, restart/retry, and RLS isolation all pass.
- **AC-W06-E01-S002-04**: An upgrade-from-previous-version replay (fixture generated at N-1, upgraded to N,
  contracts rerun) passes, as a genuine two-pass integration test.
- **AC-W06-E01-S002-05**: The fixture is wired into CI as a required gate, appearing in `ci/release-gates.yaml`
  at its Wave-4 boundary.

## Required artifacts

- The golden-consumer scaffold job (T1).
- The two-module, multi-subsystem generated fixture content (T2).
- The boot-and-exercise test harness (T3).
- The upgrade-replay test harness (T4).
- The CI gate wiring (T5).
See `artifacts/index.md`.

## Required evidence

- Fixture-installs-via-`go install` evidence (T1).
- Per-subsystem coverage evidence, one fixture per subsystem type (T2).
- Boot-and-exercise evidence against real infrastructure, covering all four named paths (T3).
- Two-pass upgrade-replay evidence (T4).
- CI-gate-wiring evidence (T5).
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, all five acceptance criteria numbered and measurable, dependency on
W01-E04-S001 recorded, owner/reviewer assignment pending, the two-module domain-content choice and T5's
manifest-schema-sequencing question recorded as unresolved questions rather than silently assumed.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all five acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the upgrade replay (AC-W06-E01-S002-04) is a genuine two-pass
integration test, not a single-pass assertion dressed up as a replay, and that the subsystem-coverage
claim (AC-W06-E01-S002-02) matches what was actually exercised.

## Risks

RISK-W06-E01-001 (DX-04 T2's broad subsystem-coverage requirement — "many kernel subsystems must be
generator-reachable," per PLAN's own risk note) — see epic-level `risks.md` for full detail and
mitigation/contingency.

## Residual-risk expectations

Once T2's minimal-exercise framing and T4's already-ratified-policy consumption are honored as planned,
residual risk is expected to be low — this is a well-bounded, source-derived fixture-building story with
a clear acceptance bar per PLAN's own T1–T5 acceptance-criteria columns.

## Plan

See `plan.md`.
