---
id: W06-E01
type: epic
title: Consumer and DSL
status: planned
wave: W06
owner: unassigned
reviewer: unassigned
priority: high
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - DX-03
  - DX-04
depends_on: []
stories:
  - W06-E01-S001
  - W06-E01-S002
decisions: []
risks:
  - RISK-W06-E01-001
---

# W06-E01 — Consumer and DSL

## Epic objective

Produce a design-only record of the state-of-the-art module DSL (DX-03) — explicitly labeled "target,
not implemented," with no code — and build the framework-repo-owned golden-consumer fixture (DX-04)
that proves `wowapi init`/`gen crud` and the wider generator surface actually work end-to-end, installed
as a real binary, exercised across two modules and multiple subsystems, booted against real
infrastructure, and replayed across an upgrade — the CI-authoritative fixture PLAN's own DX-04 evidence
says wowsociety cannot substitute for.

## Problem being solved

`requirement-inventory.md` row DX-03 states: "Module DSL design | ARCH/FUT | P1 | deferred |
W06-E01-S001 | Design-investigation story only (Wave-4-class per plan)." PLAN's own DX-03 section
confirms no `port`/`Manifest[T]`/`Operation[Request,Response]` DSL exists anywhere in wowapi today — "the
directive's 'proposed API, not current source' framing is accurate." Row DX-04 states: "Golden consumer +
upgrade matrix | IMPL | P1 | planned | W06-E01-S002 | Dep DX-01 T5." PLAN's own DX-04 evidence is blunt
about why wowsociety cannot serve as this fixture: "wowsociety consumes wowapi via a sibling-checkout
path `replace`, never an installed CLI binary — step 1 of the directive's own procedure ('Install the
built CLI') is never exercised... `FRAMEWORK_VERSION` is a raw SHA with no 'previous supported version'
concept... wowsociety is a real product with domain-specific logic... coupling wowapi's own release gate
to wowsociety's roadmap changes would violate the 'non-internal consumer fixture' intent." The gap this
epic closes, for DX-04, is structural: without a framework-repo-owned, installed-binary, two-module,
upgrade-tested consumer, there is no CI-authoritative proof the generator/CLI surface actually produces
working software, only `internal/testmodules/requests/`'s single hand-authored in-process reference
module used for framework unit tests. For DX-03, the gap is that the framework's own stated design
direction (a typed-port, manifest-driven module DSL) exists only as directive prose, with no design
document or decision record inside the programme itself.

## Scope

- DX-03-T0: formalize the module-DSL design into a design doc and an ADR-style decision record,
  explicitly labeled "target, not implemented" per AR-05's labeling discipline (S001).
- DX-04 T1: build the framework-repo-owned golden-consumer scaffold job, reusing DX-01 T5's isolated-
  temp-dir subprocess-scaffold harness as its shared primitive (S002).
- DX-04 T2: generate a resource, rule, workflow, event handler, recurring job, document flow,
  notification, and webhook across two modules, exercising each named subsystem at least once (S002).
- DX-04 T3: boot the API and worker processes against real Postgres/MinIO/Mailpit/OTel; exercise
  authenticated CRUD, async delivery, restart/retry, and RLS isolation (S002).
- DX-04 T4: replay an upgrade-from-previous-version cycle — fixture at N-1, upgraded to N, contracts
  rerun and pass (S002).
- DX-04 T5: wire the fixture into CI as a required gate (S002).

## Out of scope

- **DX-03-T1..Tn (module-DSL implementation)** — explicitly deferred per PLAN's own framing ("Deferred
  — out of near-term scope per §12 Wave 4"); this epic produces the design record only, no compiler, no
  runtime type system change.
- **wowsociety as a golden-consumer substitute** — PLAN's own three disqualifying reasons (sibling-
  checkout path-replace, no N/N-1 concept, real product-specific domain logic) rule this out; this
  epic's S002 builds a wowapi-repo-owned fixture instead, not a wowsociety-based one.
- **DX-05's v1/N-1 compatibility-class policy design** — already decided at W01 (DX-05 T1/T2 executed);
  this epic's DX-04 T4 consumes that already-ratified policy to define "previous supported version," it
  does not re-derive it.

## Source requirements

DX-03, DX-04. No MATRIX CS-ID directly owns either as a dedicated closure spec at this epic's grain —
DX-03 is discussed only within PLAN §5.4's own prose; DX-04 likewise. Neither has a D-0N architecture-
decision dependency recorded in `requirement-inventory.md` §B or REVIEW §F/§U.

## Architectural context

DX-03 and DX-04 are grouped into one epic because both concern the framework's *consumer-facing*
generator/DSL surface — one is the future design of that surface (DX-03), the other is proof the
current surface actually works for a real consumer (DX-04) — rather than because they share
implementation dependencies; DX-03 produces no code and DX-04 depends on infrastructure DX-03 does not
touch. `impl/analysis/wave-allocation-detail.md`'s own W06-E01 grouping states this exactly: "S001
module-dsl-design (DX-03 — DESIGN INVESTIGATION story: outputs design doc + decision, no code); S002
golden-consumer-matrix (DX-04; dep W01-E04-S001 harness)." This grouping is fixed by the canonical
allocation and is not to be regrouped.

## Included stories

- **W06-E01-S001 — module-dsl-design** (PLAN DX-03-T0): a design-investigation story producing a design
  doc and an ADR-style decision record, explicitly labeled "target, not implemented" — no code.
- **W06-E01-S002 — golden-consumer-matrix** (PLAN DX-04 T1–T5): the framework-repo-owned golden-consumer
  fixture, built on W01-E04-S001's DX-01 T5 harness, exercising the named subsystem set across two
  modules, booted against real infrastructure, replayed across an upgrade, wired into CI.

## Dependencies

No dependency on any other W06 epic for this epic's own entry — this epic's two stories may proceed in
either order or in parallel relative to each other (DX-03's design work and DX-04's fixture-building
target disjoint surfaces). This epic depends on W05's exit gate (AR-01/AR-02 for DX-03-T0's own
dependency: "Wave 1 (AR-01 ApplicationModel, AR-02 typed ports) complete") and cross-wave on
W01-E04-S001 (the DX-01 T5 harness DX-04 T1 reuses). Downstream: W06-E02-S003's T5 leg depends on this
epic's S001 (DX-03 design); W06-E02-S003's T7 leg depends on this epic's S002 (DX-04).

## Risks

RISK-W06-E01-001 (DX-04's broad subsystem-coverage surface — "many kernel subsystems must be
generator-reachable," per PLAN's own risk note on T2) originates at this epic's S002. See `risks.md`
for the epic-scoped elaboration.

## Required decisions

None. Neither DX-03 nor DX-04 carries a D-0N architecture-decision dependency in the source (confirmed
— no D-0N row in `requirement-inventory.md` §B or REVIEW §F/§U targets either). This epic's stories
accordingly carry no `decisions/` directory (S001's own ADR-style output is a story-produced artifact,
not a consumed programme-level D-0N decision — see S001's `story.md` for why it is not pre-written
here).

## Epic acceptance criteria

- **AC-W06-E01-01**: A module-DSL design doc and an ADR-style decision record exist, explicitly labeled
  "target, not implemented"; no DX-03 implementation code is produced by this epic.
- **AC-W06-E01-02**: The golden-consumer fixture installs via `go install`, not a repo-internal import;
  exercises resource, rule, workflow, event handler, recurring job, document flow, notification, and
  webhook generation across at least two modules; boots against real Postgres/MinIO/Mailpit/OTel with
  authenticated CRUD, async delivery, restart/retry, and RLS isolation all passing.
- **AC-W06-E01-03**: The fixture replays an upgrade-from-previous-version cycle (fixture at N-1, upgraded
  to N, contracts rerun and pass) and is wired into CI as a required gate at its Wave-4 boundary
  (REL-01).
- **AC-W06-E01-04**: Both stories have passed independent review per mandate §14 (S002 only — S001 is a
  design-investigation story without an independent-review task, per this epic's own tasks-index
  rationale).

## Closure conditions

Both stories reach `accepted` (each satisfying its own `closure.md`); AC-W06-E01-01 through
AC-W06-E01-04 above are all satisfied; `closure-report.md` for this epic is completed with reviewer
conclusion and acceptance date; DX-03's design record is confirmed to introduce no implementation code,
and DX-04's upgrade-replay evidence is confirmed to be a genuine two-pass integration test, not a
single-pass assertion dressed up as a replay.
