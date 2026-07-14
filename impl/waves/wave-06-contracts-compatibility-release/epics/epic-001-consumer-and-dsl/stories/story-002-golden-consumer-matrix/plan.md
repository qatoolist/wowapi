---
id: PLAN-W06-E01-S002
type: plan
parent_story: W06-E01-S002
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W06-E01-S002

Per mandate §8.5. Confirmed facts, planned changes, and implementation assumptions are distinguished
explicitly below.

## Proposed architecture

A new, framework-repo-owned golden-consumer fixture, built as a CI job that: (1) installs the CLI via
`go install`; (2) invokes the CLI to scaffold two modules and generate the full named-subsystem set
within them, reusing DX-01 T5's isolated-temp-dir subprocess-scaffold harness (W01-E04-S001) rather than
building a second harness; (3) boots the generated product against real infrastructure; (4) replays an
upgrade cycle; (5) is registered as a required CI gate. No change to any existing kernel or CLI package
is required — the fixture is purely a consumer of the existing generator/CLI surface.

## Implementation strategy

1. Confirm W01-E04-S001's isolated-temp-dir harness is `accepted` and reusable as-is (no modification
   expected, per that story's own forward-reference note).
2. Build the golden-consumer scaffold job (T1): install the CLI via `go install`, invoke `wowapi init`
   for two modules.
3. Generate the named subsystem set (resource, rule, workflow, event handler, recurring job, document
   flow, notification, webhook) across the two modules (T2), each subsystem exercised at least once.
4. Boot the generated API and worker processes against real Postgres/MinIO/Mailpit/OTel (T3); exercise
   authenticated CRUD, async delivery, restart/retry, and RLS isolation.
5. Implement the upgrade-from-previous-version replay (T4): generate the fixture at N-1 (per DX-05's
   ratified v1/N-1 policy), upgrade to N, rerun contracts.
6. Wire the fixture into CI as a required gate (T5), coordinating with W06-E03-S001's manifest schema
   once it exists (or landing a forward-compatible placeholder entry if sequencing requires T5 first —
   see "Unresolved questions").

## Expected package or module changes

A new golden-consumer fixture package/directory (exact location TBD); CI workflow file changes; no
change to any existing kernel or CLI package.

## Expected file changes where determinable

- A new golden-consumer fixture directory (exact path TBD).
- New CI workflow configuration wiring the fixture in as a required gate.
- No changes to `internal/cli/`, `kernel/`, or any existing generator template.

## Contracts and interfaces

None new — the fixture consumes existing CLI/generator contracts, it does not define new ones.

## Data structures

None new at the framework level. The fixture's own generated modules will have their own data
structures (per whatever the generator produces for a resource/rule/workflow/etc.), but these are
generator output, not new framework data structures this story hand-designs.

## APIs

None affected at the framework level.

## Configuration changes

None anticipated beyond the fixture's own CI job configuration (environment variables, infrastructure
connection strings for Postgres/MinIO/Mailpit/OTel).

## Persistence changes

None at the framework level — the fixture's own generated modules will have their own migrations (per
whatever the generator produces), scoped to the fixture's own throwaway database, not a framework
schema change.

## Migration strategy

T4's upgrade replay is itself the migration strategy under test — this story does not introduce new
framework migration tooling.

## Concurrency implications

T3's restart/retry exercise implicitly tests some concurrency-adjacent behavior (a worker restarting
mid-job), but this story does not introduce new concurrency primitives — it exercises existing ones.

## Error-handling strategy

The fixture's own CI job must fail clearly and diagnosably if any of T1–T4's steps fail, so a developer
debugging a CI-gate failure can identify which stage (install, generate, boot, upgrade-replay) broke.

## Security controls

None new — the fixture exercises existing RLS-isolation and authenticated-CRUD paths without
introducing new authorization logic.

## Observability changes

None beyond T3's own OTel-boot exercise.

## Testing strategy

- T1: fixture-installs-via-`go install` evidence.
- T2: one fixture per subsystem type, confirming generation succeeds and the subsystem is exercised.
- T3: integration test against real infrastructure, covering authenticated CRUD, async delivery,
  restart/retry, and RLS isolation, all passing.
- T4: two-pass integration test — generate at N-1, upgrade to N, rerun contracts, confirm pass.
- T5: CI config test confirming the manifest entry exists at the correct Wave-4 boundary.

## Regression strategy

Once wired into CI as a required gate (T5), this fixture becomes the ongoing regression guard against
any future generator/CLI change that silently breaks multi-module, multi-subsystem generation or the
upgrade path.

## Compatibility strategy

Not directly applicable — the fixture is additive CI infrastructure. T4's upgrade replay is itself a
compatibility-proving mechanism, consuming DX-05's already-ratified v1/N-1 policy rather than defining a
new one.

## Rollout strategy

Single story, landed as its own reviewable unit. T5's CI-gate wiring is the final rollout step, made
required only once T1–T4 are demonstrated working.

## Rollback strategy

If the fixture proves flaky or too CI-time-expensive once wired in as a required gate, it can be
temporarily demoted to advisory (non-blocking) while the flake is diagnosed — this is a standard CI-gate
rollback pattern, not specific to this story.

## Implementation sequence

As listed under "Implementation strategy" above (steps 1–6), matching T1→T2→T3→T4→T5's own natural
dependency order (each step's fixture content depends on the prior step's scaffold/generation having
succeeded).

## Task breakdown

- **W06-E01-S002-T001** — Golden-consumer scaffold job (T1).
- **W06-E01-S002-T002** — Cross-module, cross-subsystem generation (T2).
- **W06-E01-S002-T003** — Boot-and-exercise against real infrastructure (T3).
- **W06-E01-S002-T004** — Upgrade-from-previous-version replay (T4).
- **W06-E01-S002-T005** — CI gate wiring (T5).
- **W06-E01-S002-T006** — Independent review.

## Expected artifacts

The golden-consumer scaffold job; the two-module generated fixture content; the boot-and-exercise
harness; the upgrade-replay harness; the CI gate wiring.

## Expected evidence

Fixture-install evidence; per-subsystem coverage evidence; boot-and-exercise evidence; two-pass
upgrade-replay evidence; CI-gate-wiring evidence.

## Unresolved questions

- Exact domain content of the two generated modules (names, fields) — not specified by any source
  document; to be chosen at implementation time.
- Whether T5's CI-gate entry lands before or after W06-E03-S001's manifest schema exists, and if before,
  what the forward-compatible placeholder entry's exact shape is — to be resolved at implementation time
  by coordinating with W06-E03-S001's own sequencing.
- Exact fixture directory location.

## Approval conditions

This plan is approved for implementation once: (a) W01-E04-S001 has reached `accepted` (the harness this
story's T001 reuses), and (b) the owner and reviewer are assigned.
