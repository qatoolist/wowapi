---
id: W02-E01-S003
type: story
title: Canary, switch, and contract-phase tooling with the full CI drill pipeline
status: accepted
wave: W02
epic: W02-E01
owner: unassigned
reviewer: unassigned
priority: critical
created_at: 2026-07-12
updated_at: 2026-07-12
source_requirements:
  - DATA-09
depends_on:
  - W02-E01-S002
blocks: []
acceptance_criteria:
  - AC-W02-E01-S003-01
  - AC-W02-E01-S003-02
  - AC-W02-E01-S003-03
  - AC-W02-E01-S003-04
artifacts: []
evidence: []
decisions: []
risks:
  - RISK-W02-003
---

# W02-E01-S003 — Canary, switch, and contract-phase tooling with the full CI drill pipeline

## Story ID

W02-E01-S003

## Title

Canary, switch, and contract-phase tooling with the full CI drill pipeline

## Objective

Build the final three phases of the online-migration protocol — canary/deploy-N tooling with soak
metrics (N alongside N-1), switch-phase tooling (observable compatibility flag, dual-schema-version
consumer support, proven application rollback after switch), and contract-phase tooling gated on an
evidenced no-N-1-remains precondition — and wire all six directive-named drills into a CI/scheduled
pipeline with a passing run artifact.

## Value to the framework

S001 and S002 give the protocol its manifest, budget, expand, backfill, and validation machinery;
this story delivers what PLAN calls "the core safety property this protocol exists to guarantee"
(T7's own risk column): the ability to roll an application back after a schema switch without a
destructive `Down`, and the gate preventing the single most dangerous step — contract — from running
before evidence proves no N-1 process remains. PLAN T8's own risk column: "Most safety-critical
piece — running contract too early is destructive and hard to detect pre-outage." The CI drill
pipeline (T9) converts these safety properties from one-time proofs into continuously-re-verified
guarantees, at the cost PLAN itself acknowledges: "Largest single infra investment in PF-DATA."

## Problem statement

PLAN DATA-09's task table gives four rows for this story's scope. T6: "Canary/deploy-N tooling: N
alongside N-1, soak metrics | T5 | **N-1 on expanded N schema + N code before/after backfill (both
explicitly required)** | This is the test | `DATA-09/canary-soak/` | No production telemetry
baseline exists — soak duration/thresholds are a genuine, currently unresolvable judgment gap |
Code for the harness; human decision on soak duration and go/no-go." T7: "Switch-phase tooling:
observable compatibility flag, dual-schema-version consumer support | T6 | **Application rollback
after switch (explicitly required)**, no destructive `Down` | This is the test |
`DATA-09/switch-rollback/` | The core safety property this protocol exists to guarantee | Code for
mechanics; **the decision to flip in production is human**." T8: "Contract-phase tooling: gated on
evidenced no-N-1-remains precondition | T7 | **Forward recovery from every failed phase +
delayed-contract-only-after-old-process-absence-proven (both required)** | This is the test |
`DATA-09/contract-gate/` | Most safety-critical piece — running contract too early is destructive
and hard to detect pre-outage | Code for the gate; human sign-off strongly advisable even with the
gate passing." T9: "Full CI drill pipeline covering all 6 directive-named drills | T1-T8 | All six
drills run in CI/scheduled pipeline | CI pipeline + passing run artifact |
`DATA-09/ci-drill-pipeline/` | Largest single infra investment in PF-DATA | Code; human decision on
which real migration is the first live exercise — DATA-01's composite-FK rollout is the natural
first candidate." None of this tooling exists today.

## Source requirements

DATA-09 (T6, T7, T8, T9).

## Current-state assessment

Per PLAN's evidence for DATA-09 as a whole (to be re-confirmed at this story's own execution
commit): no canary/soak, switch, or contract tooling and no drill pipeline exists anywhere in the
repository. The wowsociety-impact note in the same PLAN section confirms the gap extends to the
consuming product's deploy process: "`wowsociety/docs/DEPLOY.md:81-100` documents a single-shot
'migrate fully, then deploy everyone' model... no canary/soak, no N/N-1 dual-version window, no
interrupted-backfill resume, no telemetry-gated contract check" — meaning there is also no existing
deployment pattern this tooling can crib from within the two repositories; it is genuinely built
from zero. Two judgment gaps are confirmed by PLAN's own risk/classification columns and are
inherited by this story as explicit boundaries, not problems to silently solve: (1) soak
duration/thresholds have no production telemetry baseline to calibrate against ("a genuine,
currently unresolvable judgment gap" — RISK-W02-003); (2) the production flip decision (T7) and the
contract sign-off (T8) are human decisions the tooling supports but must not automate away.

## Desired state

Canary tooling can run N alongside N-1 and prove, via its named test, that N-1 code works against
the N-expanded schema both before and after backfill, with soak metrics collected against
configurable (not hardcoded-guessed) duration/threshold parameters. Switch tooling exposes an
observable compatibility flag, supports dual-schema-version consumers, and proves application
rollback after switch with no destructive `Down`. Contract tooling refuses to run unless the
no-N-1-remains precondition is evidenced, and forward recovery from every failed phase is proven.
All six directive-named drills run in a CI/scheduled pipeline producing a passing run artifact, with
the individual drill outputs aggregated into one consolidated evidence bundle.

## Scope

- Canary/deploy-N tooling: N alongside N-1, soak-metric collection, configurable soak
  duration/threshold parameters (PLAN T6).
- Switch-phase tooling: observable compatibility flag, dual-schema-version consumer support,
  application-rollback-after-switch mechanics, no destructive `Down` (PLAN T7).
- Contract-phase tooling: the evidenced no-N-1-remains precondition gate, forward recovery from
  every failed phase (PLAN T8).
- The full CI drill pipeline covering all six directive-named drills, with a passing run artifact
  (PLAN T9).
- An evidence-aggregation step consolidating T6/T7/T8's individual drill outputs and T9's pipeline
  run into one consolidated evidence bundle (see `tasks/index.md` grouping rationale).

## Out of scope

- **Calibrating soak duration/threshold numeric values** — PLAN T6's own risk column marks this "a
  genuine, currently unresolvable judgment gap" absent a production telemetry baseline. This story
  builds configurable parameters; the values are per-rollout human decisions (RISK-W02-003, accepted
  residual risk at wave level, not silently resolved here).
- **The production flip decision and contract sign-off** — PLAN T7/T8's own classification columns:
  "the decision to flip in production is human"; "human sign-off strongly advisable even with the
  gate passing." The tooling enforces preconditions and mechanics; it does not automate these
  decisions.
- **Choosing and executing the first real migration through the drills** — PLAN T9's own note names
  DATA-01's composite-FK rollout as "the natural first candidate"; that consumption happens in
  W02-E02, and the live-production scheduling decision is operational, outside this story.
- **Expand/backfill/validate tooling** — W02-E01-S002's scope (this story depends on it).
- **wowsociety's own deploy-process adoption of the N/N-1 window** — product-level process change,
  tracked per PLAN's wowsociety-impact note, not this framework story's scope.

## Assumptions

- The "6 directive-named drills" T9 covers are assumed to be the six explicitly-required named test
  scenarios PLAN's own T4–T8 rows bold: (1) interrupted/resumed backfill, (2) N-1 on expanded N
  schema before backfill, (3) N code before/after backfill (T6's second required leg), (4)
  application rollback after switch, (5) forward recovery from every failed phase, (6)
  delayed-contract-only-after-old-process-absence-proven. The directive document itself
  (`docs/implementation/architecture-directive-2026-07-11.md`) is the naming source; this mapping
  must be confirmed against the directive's own drill list at implementation time rather than
  treated as settled — recorded as an unresolved question in `plan.md`.
- The CI drill pipeline is assumed to extend the existing CI infrastructure
  (`.github/workflows/`-based) rather than requiring a new CI system — consistent with the wave's
  "Tooling dependencies: none new" statement, to be confirmed when T9's runtime cost (drills involve
  multi-version deploys against real PostgreSQL) is measured against CI runner constraints; a
  scheduled (nightly-style) pipeline rather than per-PR is the expected shape per PLAN's own "CI
  pipeline + passing run artifact" framing, exact trigger to be decided at implementation time.
- Dual-version testing (N alongside N-1) is assumed achievable in CI via container images or
  checked-out prior-version builds — the exact mechanism for materializing an "N-1 application
  version" in a test environment is an unresolved implementation question.

## Dependencies

Depends on W02-E01-S002 (PLAN T6's "Depends-on" column names T5; T9 depends on T1–T8, i.e. on both
S001 and S002 transitively through S002). Blocks nothing within this wave directly — W02-E02-S002's
gate is on S001+S002 acceptance per `impl/analysis/wave-allocation-detail.md`, not on this story —
but this epic's own closure (and therefore the full-protocol dependency W03-E01-S001 and
W04-E04-S001 consume) requires this story accepted.

## Affected packages or components

New: canary/soak tooling, switch-phase tooling, contract-phase gate (package locations TBD, expected
adjacent to S001/S002's protocol tooling); a new CI workflow (or extension of an existing one) for
the drill pipeline under `.github/workflows/`.

## Compatibility considerations

Dual-schema-version consumer support (T7) is itself a compatibility mechanism — the switch tooling
must let N-1 and N application versions coexist against the same database during the switch window.
The no-destructive-`Down` requirement means rollback is achieved by application-version rollback
against the still-compatible expanded schema, never by destructive schema reversal — this is the
protocol's compatibility contract, delivered (not merely considered) by this story.

## Security considerations

The contract-phase gate is safety-critical in the data-loss sense rather than the access-control
sense: PLAN T8's risk column — "running contract too early is destructive and hard to detect
pre-outage." The gate's evidenced-precondition check must fail closed: absent or ambiguous evidence
of N-1 absence blocks contract, it does not default-allow.

## Performance considerations

Soak metrics are this story's performance-observation surface — collected during canary against
configurable thresholds. The drill pipeline's own runtime cost (T9: "Largest single infra investment
in PF-DATA") is a CI-resource consideration: drill scheduling (per-PR vs. nightly) must balance
verification freshness against runner cost, decided at implementation time.

## Observability considerations

The switch-phase compatibility flag must be *observable* per PLAN T7's own wording — an operator
must be able to see which schema-version posture the application is running in. Soak metrics must be
collected and retained as part of the canary evidence. Drill pipeline runs must produce a durable
passing-run artifact (T9's own evidence requirement).

## Migration considerations

This story is itself migration-tooling (the final three phases). It executes no real production
migration; the drills run against fixture/test migrations in CI.

## Documentation requirements

Document: the canary tooling's configuration surface (soak duration/threshold parameters) with an
explicit note that value calibration is a per-rollout human judgment (RISK-W02-003); the switch
tooling's compatibility-flag semantics and the human flip decision boundary; the contract gate's
evidenced-precondition requirements and the human sign-off recommendation; the drill pipeline's
trigger, drill list, and artifact location.

## Acceptance criteria

- **AC-W02-E01-S003-01**: Canary/deploy-N tooling proves, via its named test, both explicitly-
  required legs: N-1 code runs correctly against the N-expanded schema, and N code runs correctly
  both before and after backfill. Soak duration/threshold parameters are configurable, not
  hardcoded.
- **AC-W02-E01-S003-02**: Switch-phase tooling proves, via its named test, application rollback
  after switch with no destructive `Down`; the compatibility flag is observable; dual-schema-version
  consumers are supported.
- **AC-W02-E01-S003-03**: Contract-phase tooling proves, via its named test, both explicitly-
  required properties: forward recovery from every failed phase, and contract gated on evidenced
  absence of N-1 processes (fail-closed on missing/ambiguous evidence).
- **AC-W02-E01-S003-04**: All six directive-named drills run in the CI/scheduled pipeline, producing
  a passing run artifact; the individual drill outputs are aggregated into one consolidated evidence
  bundle registered in `evidence/index.md`.

## Required artifacts

- Canary/deploy-N tooling (code).
- Switch-phase tooling (code).
- Contract-phase gate (code).
- The CI drill pipeline definition.
- The consolidated 6-drill evidence bundle (post-implementation).
- Documentation for all of the above.
See `artifacts/index.md`.

## Required evidence

- Canary named-test output (both required legs).
- Switch-rollback named-test output.
- Contract-gate named-test output (both required properties).
- CI drill pipeline passing-run artifact.
- The consolidated evidence bundle aggregating all of the above.
See `evidence/index.md`.

## Definition of ready

Confirmed against `governance/definition-of-ready.md` before this story moves to `ready`: `story.md`
and `plan.md` complete, acceptance criteria numbered and measurable, dependency on W02-E01-S002
recorded, the soak-threshold judgment gap recorded as an accepted boundary (RISK-W02-003) rather
than a problem this story claims to solve, owner/reviewer assignment pending.

## Definition of done

Confirmed against `governance/definition-of-done.md` before this story moves to `accepted`:
implementation matches `plan.md` or deviations are recorded in `deviations.md`; all four acceptance
criteria verified with evidence in `evidence/index.md`; `closure.md` completed; independent review
passed per mandate §14, specifically confirming the soak-threshold judgment gap is recorded as an
accepted residual risk, not silently resolved with an invented number (epic-level `acceptance.md`
AC-W02-E01-04's story-specific review focus).

## Risks

RISK-W02-003 (no production telemetry baseline for soak duration/thresholds — PLAN's own "currently
unresolvable judgment gap") — see wave-level and epic-level `risks.md` for full detail. Additionally,
T9's infra investment scale ("Largest single infra investment in PF-DATA") carries schedule risk: if
the drill pipeline's build cost threatens story boundedness (mandate §12), the contingency is to
split pipeline hardening into a follow-up task with the six drills' core wiring landed first, not to
silently narrow which drills run.

## Residual-risk expectations

RISK-W02-003 is expected to remain open and accepted at this story's closure — the tooling delivers
configurable parameters; calibration awaits a production telemetry baseline that does not exist
within this programme's current scope. This is recorded at wave, epic, and story level so no closure
reviewer can mistake it for an oversight. Beyond that, once the four named tests pass, residual risk
for this story's own mechanics is expected to be low; the human-decision boundaries (flip, contract
sign-off) are permanent process features, not residual risks.

## Plan

See `plan.md`.
