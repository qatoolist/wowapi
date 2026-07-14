---
id: IMPL-W07-E02-S002
type: implementation-record
parent_story: W07-E02-S002
status: implemented
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Implementation record — W07-E02-S002

## What was actually implemented

- Converted every execution-time DB/S3/E2E prerequisite skip that is required by the authoritative
  gate into an actionable fail-closed branch while retaining documented local-only skips.
- Replaced the informational grep audit with a Go-AST manifest validator. The execution-time scan
  classified all 39 discovered sites, removed the probabilistic TOTP skip, and registered the
  remaining 38 with owners and rationales.
- Scoped the hosted race leg to DB/S3-backed integration packages and added a build-tagged seeded race
  whose wrapper succeeds only after observing `WARNING: DATA RACE`.
- Added native Go fuzz PR (10s/target) and scheduled (1m/target) jobs. The scheduled cache uses a
  per-run save key and stable restore prefix so generated `GOCACHE/fuzz` entries survive across runs.
- Added `fuzzproof`, which rejects seed-replay-only output, writes per-target logs and a JSON report,
  records positive elapsed executions, and inventories retained corpus files/mtime.

## Components changed

CI workflow; Make quality/test targets; DB/S3/E2E test prerequisites; skip policy tooling; opt-in race
fixture; fuzz proof runner; story artifacts/evidence.

## Files changed

Authoritative paths are registered in `artifacts/index.md`. Load-bearing implementation paths are
`.github/workflows/ci.yml`, `Makefile`, `miscellaneous/check_test_skips.sh`,
`miscellaneous/test-skip-manifest.json`, `miscellaneous/check_required_test_prerequisites.sh`,
`miscellaneous/check_test_skip_fixtures.sh`, `miscellaneous/check_race_detector.sh`,
`internal/tools/testskipmanifest/`, `internal/tools/fuzzproof/`, and
`internal/verificationfixtures/racefixture/`, plus the classified prerequisite test files.

## Interfaces introduced or changed

New repository commands: `make check-test-skips`, `make check-required-test-prerequisites`,
`make check-race-fixture`, `make test-race-integration`, `make test-fuzz-pr`, and
`make test-fuzz-scheduled`. The skip manifest JSON schema is version 1.

## Configuration changes

`ci.yml` now validates skip approvals, executes the seeded race fixture before its real integration
race suite, and runs separate short/long fuzz jobs with retained cache and uploaded proof artifacts.

## Schema or migration changes

None.

## Security changes

Required verification dependencies no longer silently erase security/integration coverage. Native
coverage-guided fuzzing now continuously explores untrusted filter, sort, and cursor parsers.

## Observability changes

Fuzz jobs publish JSON containing wall time, per-target positive fuzz elapsed time/executions, corpus
file count, and latest corpus mtime, plus raw native-fuzzer logs.

## Tests added or modified

AST validator unit tests and approved/unapproved fixtures; missing DB/S3 negative fixture; seeded race
fixture; fuzz-output/corpus unit tests; deterministic TOTP wrong-code test.

## Commits

Evidence is pinned to repository HEAD `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus the scoped
shared W07 working-tree provenance recorded in every evidence record. Per parent instruction, this
executor did not commit shared-worktree files.

## Pull requests

None created.

## Implementation dates

2026-07-14.

## Technical debt introduced

None.

## Known limitations

The scheduled fuzz corpus depends on GitHub Actions cache retention; proof reports are separately kept
as workflow artifacts for 14 days (PR) and 30 days (scheduled). First scheduled execution necessarily
starts without a prior scheduled cache, while every later run restores by stable prefix.

## Follow-up items

None within story scope.

## Relationship to the approved plan

Implemented the four planned mechanisms. The execution-time choices (per-change scoped integration
race, 10s PR fuzz, 1m scheduled fuzz, Actions-cache corpus retention) resolve the plan's explicit open
questions and are recorded in `deviations.md` as implementation-time decisions rather than hidden plan
rewrites.
