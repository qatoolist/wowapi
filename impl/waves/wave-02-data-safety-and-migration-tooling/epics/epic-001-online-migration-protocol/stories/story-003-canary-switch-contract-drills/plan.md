---
id: PLAN-W02-E01-S003
type: plan
parent_story: W02-E01-S003
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W02-E01-S003

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below. This plan deliberately does not invent soak-duration/threshold values (PLAN T6's
"currently unresolvable judgment gap," RISK-W02-003), does not automate the human flip/sign-off
decisions (PLAN T7/T8's own classification columns), and records the drill-list mapping as an
assumption to confirm against the directive document, not settled fact.

## Proposed architecture

Three phase-tools layered on S001/S002's protocol foundation, plus a CI pipeline: (1) canary tooling
that orchestrates an N-1-alongside-N deployment against an expanded schema and collects soak
metrics; (2) switch tooling exposing an observable compatibility flag and supporting
dual-schema-version consumers, with rollback achieved by application-version rollback (never
destructive schema `Down`); (3) a contract gate that verifies an evidenced no-N-1-remains
precondition before allowing the contract phase, failing closed; (4) a scheduled CI workflow running
all six directive-named drills against fixture migrations and producing a durable passing-run
artifact.

## Implementation strategy

1. Confirm the six-drill mapping against
   `docs/implementation/architecture-directive-2026-07-11.md`'s own drill naming (resolving the
   assumption in `story.md`).
2. Decide the mechanism for materializing an N-1 application version in a test environment
   (container image of the prior release vs. checked-out prior build) — this is a prerequisite for
   both canary and switch testing.
3. Implement canary/deploy-N tooling: run N-1 against the N-expanded schema, run N before and after
   backfill, collect soak metrics against configurable duration/threshold parameters.
4. Write the canary named test covering both explicitly-required legs (PLAN T6's bolded acceptance
   criterion).
5. Implement switch-phase tooling: observable compatibility flag, dual-schema-version consumer
   support, application-rollback mechanics, no destructive `Down` in any generated/managed
   migration path.
6. Write the switch-rollback named test (PLAN T7's bolded "Application rollback after switch"
   requirement).
7. Implement the contract-phase gate: define what constitutes admissible evidence of N-1 absence
   (e.g. version-tagged connection/process registry, deploy-system attestation — exact evidence
   source is an unresolved question below), and gate the contract step on it, failing closed.
8. Write the contract-gate named test covering both explicitly-required properties: forward
   recovery from every failed phase, and delayed-contract-only-after-old-process-absence-proven
   (PLAN T8's bolded requirements).
9. Build the CI drill pipeline: a scheduled workflow running all six drills against fixture
   migrations, producing a durable passing-run artifact.
10. Aggregate the drill outputs: consolidate T6/T7/T8's individual named-test outputs and T9's
    pipeline run artifact into one consolidated evidence bundle registered in `evidence/index.md`.
11. Document everything, including the human-decision boundaries and the soak-calibration gap.

## Expected package or module changes

New packages for canary, switch, and contract tooling (locations TBD, adjacent to S001/S002's
protocol tooling); a new or extended workflow under `.github/workflows/` for the drill pipeline.

## Expected file changes where determinable

Not yet determinable by file/line beyond the `.github/workflows/` addition — this is new tooling
whose package structure depends on S001/S002's own implementation-time layout decisions.

## Contracts and interfaces

The compatibility flag's read interface (how an operator/process observes the current
schema-version posture); the contract gate's evidence-input contract (what artifact proves N-1
absence). Both are new, additive interfaces — no existing public contract changes.

## Data structures

Soak-metric records (metric name, value, threshold, timestamp, canary run ID); the compatibility-
flag state; the contract gate's evidence record. Exact shapes TBD at implementation time.

## APIs

None affected at the application-runtime level. The tooling itself gains command/entry-point
surfaces (exact CLI/API shape TBD).

## Configuration changes

Soak duration/threshold parameters (configurable per PLAN T6's judgment-gap boundary); drill
pipeline schedule/trigger configuration. Exact keys TBD.

## Persistence changes

Possibly a table for compatibility-flag state and/or canary run records, if file/CI-artifact-based
storage proves insufficient — to be determined at implementation time.

## Migration strategy

Not applicable in the consuming sense — this story builds the final phases of the migration
protocol; it performs no real production migration. Drills run against fixture migrations.

## Concurrency implications

The canary window is inherently a concurrent-versions scenario — N-1 and N processes operating
against the same database simultaneously; the named canary test exercises exactly this. The contract
gate must be safe against a race between its evidence check and a late-starting N-1 process — the
fail-closed posture plus the "delayed-contract-only-after-old-process-absence-proven" drill are the
controls.

## Error-handling strategy

Every phase must support forward recovery from failure (PLAN T8's bolded requirement applies to
"every failed phase," not only contract): a failed canary, switch, or contract attempt leaves the
system in a recoverable, well-defined state with a documented forward path. The contract gate fails
closed on missing or ambiguous evidence.

## Security controls

The contract gate's fail-closed evidence check (see `story.md` "Security considerations") — absent
or ambiguous evidence of N-1 absence blocks contract, never default-allows.

## Observability changes

The observable compatibility flag (T7's own requirement); soak-metric collection; the drill
pipeline's durable passing-run artifact.

## Testing strategy

- Canary named test: both required legs (N-1 on expanded N schema; N before/after backfill) —
  `DATA-09/canary-soak/`.
- Switch named test: application rollback after switch, no destructive `Down` —
  `DATA-09/switch-rollback/`.
- Contract named test: forward recovery from every failed phase + contract-only-after-absence-proven
  — `DATA-09/contract-gate/`.
- The CI drill pipeline itself: all six drills, passing run artifact — `DATA-09/ci-drill-pipeline/`.
- Per PLAN's own "Tests" columns, each named test *is* the acceptance criterion ("This is the
  test") — no substitute or weaker proxy is acceptable.

## Regression strategy

The scheduled drill pipeline is itself the regression mechanism: the six safety properties are
re-proven on every scheduled run, not only once at this story's closure. A drill regression fails
the pipeline loudly.

## Compatibility strategy

Dual-schema-version consumer support and the no-destructive-`Down` rule are this story's delivered
compatibility mechanics (see `story.md` "Compatibility considerations").

## Rollout strategy

Single story, landed as its own reviewable unit; the drill pipeline activates on its schedule once
merged. No production migration is executed.

## Rollback strategy

Application-version rollback against the still-compatible expanded schema is the rollback mechanism
this story *builds* (T7). For the story's own code: each phase tool and the pipeline can be reverted
independently; a drill-pipeline failure blocks nothing at runtime (it is a verification surface, not
a runtime dependency).

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–11). Steps 1–2 (drill-list confirmation,
N-1 materialization mechanism) are prerequisites for everything else; steps 3–8 follow the T6→T7→T8
phase order (each phase's tooling feeds the next's test scenario); steps 9–10 close with the
pipeline and aggregation.

## Task breakdown

- **W02-E01-S003-T001** — Canary/deploy-N tooling and its named test (steps 2–4).
- **W02-E01-S003-T002** — Switch-phase tooling and its named test (steps 5–6).
- **W02-E01-S003-T003** — Contract-phase gate and its named test (steps 7–8).
- **W02-E01-S003-T004** — CI drill pipeline (steps 1, 9).
- **W02-E01-S003-T005** — Evidence aggregation: the consolidated 6-drill evidence bundle (step 10).
- **W02-E01-S003-T006** — Independent review (per mandate §14, scoped to this story, with specific
  attention to the soak-threshold gap being honestly recorded).

## Expected artifacts

Canary, switch, and contract tooling; the CI drill pipeline definition; the consolidated 6-drill
evidence bundle; documentation including the human-decision boundaries and the soak-calibration gap.

## Expected evidence

The three named-test outputs (canary both legs; switch rollback; contract gate both properties); the
pipeline passing-run artifact; the consolidated evidence bundle.

## Unresolved questions

- The exact six-drill list per the directive document's own naming — assumed to be the six bolded
  requirements across PLAN T4–T8, to be confirmed against
  `docs/implementation/architecture-directive-2026-07-11.md` at implementation time (step 1).
- The mechanism for materializing an N-1 application version in CI (prior-release container image
  vs. checked-out prior build).
- The admissible evidence source for the contract gate's no-N-1-remains precondition
  (version-tagged process registry, deploy-system attestation, or another mechanism).
- The drill pipeline's trigger (nightly schedule vs. per-PR vs. both with different drill subsets),
  balancing verification freshness against runner cost.
- Whether compatibility-flag/canary-run state is database-persisted or CI-artifact-based.

## Approval conditions

This plan is approved for implementation once: (a) W02-E01-S002 has landed (the drills exercise its
backfill/validation tooling), (b) the six-drill mapping is confirmed against the directive document,
and (c) the owner and reviewer are assigned.
