---
id: PLAN-W07-E02-S002
type: plan
parent_story: W07-E02-S002
status: approved
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Plan — W07-E02-S002

Per mandate §8.5. T8's own implementation strategy is written to explicitly close both REL-04 T8 and
PERF-06 T3/T4's identical scope in one place, per CONFLICT-02's resolution — not duplicated under either
finding's own name. Confirmed facts, planned changes, and assumptions are distinguished explicitly
below.

## Proposed architecture

Four independent CI-truthfulness mechanisms: a fail-not-skip conversion for the authoritative E2E job's
own prerequisite checks; a machine-checked skip manifest extending the existing `check_test_skips.sh`;
a race-test CI job (per-PR or scheduled) over DB/S3-backed packages; and real coverage-guided fuzzing
wired into both PR (short, time-bounded) and scheduled (longer, corpus-retained) CI runs.

## Implementation strategy

1. Classify each of the 22 inventoried skip sites as legitimately optional or masking required coverage
   (T5); convert masking-required-coverage cases from skip to fail.
2. Build the machine-checked skip manifest, extending `check_test_skips.sh` (T6); write a fixture adding
   an unguarded `t.Skip()` and confirm it fails CI.
3. Wire `go test -race` over DB/S3-backed packages in CI, deciding per-PR vs. scheduled based on CI-time
   budget (T7); write a seeded data-race fixture.
4. Wire real `-fuzz=<Name> -fuzztime=Ns` execution into PR CI (short duration) and a separate scheduled
   job (longer duration, corpus retained) (T8); write a fuzz-duration/corpus-mtime test confirming
   non-zero time beyond seed replay; confirm this closes both REL-04 T8 and PERF-06 T3/T4 with no
   duplicate implementation.

## Expected package or module changes

CI workflow configuration changes for T5 (E2E job), T6 (skip manifest, likely alongside the existing
`check_test_skips.sh`), T7 (race-test job), T8 (PR + scheduled fuzz jobs).

## Expected file changes where determinable

- The authoritative E2E job's own workflow configuration (T5).
- `check_test_skips.sh` (extended) plus a new skip-manifest file (T6).
- New CI workflow configuration for the race-test job (T7).
- New CI workflow configuration for PR and scheduled fuzz jobs (T8); `make test-fuzz`'s own wiring
  confirmed/fixed.

## Contracts and interfaces

The skip-manifest's own file format (owner/rationale per approved skip entry) is the primary new
contract (T6).

## Data structures

The skip-manifest entry structure itself.

## APIs

None affected.

## Configuration changes

CI-only configuration changes; no application configuration.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

T7's own race-test job is itself the concurrency-testing surface this task wires into CI; no new
framework concurrency primitive is introduced.

## Error-handling strategy

T5's own fail-not-skip conversion must produce a clear, diagnosable failure message (which prerequisite
was unmet), not merely a generic non-zero exit.

## Security controls

T8's own real fuzzing is itself a security-adjacent defect-discovery mechanism; see `story.md` "Security
considerations."

## Observability changes

T8's own fuzz artifacts (proving non-zero fuzzing time) are the primary new observability output.

## Testing strategy

- T5: kill a required E2E dependency, confirm the job now fails (not "0 tests ran, green").
- T6: fixture adding an unguarded `t.Skip()`, confirm CI failure; confirm an approved skip with
  rationale passes.
- T7: seeded data-race fixture, confirm `-race` catches it in CI.
- T8: fuzz-duration/corpus-mtime test confirming non-zero fuzzing time beyond seed replay.

## Regression strategy

T6's own skip manifest becomes the ongoing regression guard against a future unapproved `t.Skip()`. T8's
own real-fuzz wiring becomes the ongoing coverage-guided defect-discovery mechanism.

## Compatibility strategy

T5's own fail-not-skip conversion is an intentional behavior change over the prior false-positive-
prone state, per `story.md` "Compatibility considerations."

## Rollout strategy

T5, T6, T7, T8 may proceed independently, each landing once its own implementation is complete; T8's own
scheduled-job component may land slightly after its PR-job component if CI-budget tuning requires
iteration.

## Rollback strategy

If T7 or T8 produce excessive CI-time cost once wired in, move the more expensive component (race tests,
long-fuzz runs) to a scheduled-only cadence rather than every PR — this is an anticipated, not an
exceptional, tuning step per both tasks' own PLAN risk notes.

## Implementation sequence

T5, T6, T7, T8 may proceed in any order, in parallel — no cross-task dependency exists among them.

## Task breakdown

- **W07-E02-S002-T001** — Fail-not-skip E2E prerequisites (T5).
- **W07-E02-S002-T002** — Machine-checked skip manifest (T6).
- **W07-E02-S002-T003** — Race tests over integration-relevant packages (T7).
- **W07-E02-S002-T004** — Real time-bounded coverage-guided fuzzing, owning PERF-06 T3/T4 (T8).
- **W07-E02-S002-T005** — Independent review.

## Expected artifacts

The fail-not-skip E2E job configuration + skip-site classification record; the machine-checked skip
manifest; the race-test CI job configuration; the real-fuzz PR and scheduled CI job configuration.

## Expected evidence

Kill-a-required-dependency test output; unguarded-`t.Skip()` fixture fail-test output; seeded data-race
fixture test output; fuzz-duration/corpus-mtime test output.

## Unresolved questions

- The exact classification of each of the 22 inventoried skip sites — this task's own central work, not
  knowable in advance.
- T7's per-PR vs. scheduled-only decision.
- T8's exact fuzz-time-bound duration for PR vs. scheduled runs.
- T8's corpus-retention mechanism (artifact vs. commit).

## Approval conditions

This plan is approved for implementation once the owner and reviewer are assigned.
