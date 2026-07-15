---
id: W07-E02-S002-T003
type: task
title: Race tests over integration-relevant packages
status: done
parent_story: W07-E02-S002
owner: W07-E02-S002 executor
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W07-E02-S002-03
artifacts:
  - ART-W07-E02-S002-003
evidence:
  - EV-W07-E02-S002-003
---

# W07-E02-S002-T003 — Race tests over integration-relevant packages

## Task Definition

### Task objective

Run go test -race over DB/S3-backed packages in CI.

### Parent story

W07-E02-S002

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Wire go test -race over DB/S3-backed packages in CI.
2. Decide per-PR vs. scheduled-only, based on CI-time budget.
3. Write a seeded data-race fixture and confirm -race catches it.

### Expected files or components affected

New CI workflow configuration for the race-test job.

### Expected output

go test -race runs over DB/S3-backed packages, catching a seeded data race.

### Required artifacts

ART-W07-E02-S002-003 (race-test CI job configuration).

### Required evidence

EV-W07-E02-S002-003 (seeded data-race fixture test output).

### Related acceptance criteria

AC-W07-E02-S002-03.

### Completion criteria

The seeded data-race fixture is caught by -race in CI.

### Verification method

Direct execution of the race-test job against the seeded fixture.

### Risks

Medium — may need a separate scheduled job, not every PR, per PLAN T7's own risk note.

### Rollback or recovery considerations

If per-PR race testing proves too CI-time-expensive, move to a scheduled-only cadence rather than silently disabling the check.

## Implementation Record

### What was actually implemented

The existing per-change container race leg now runs an explicit integration target over S3, E2E,
tenant-FK, database, migration, outbox, and testkit packages with both requirement flags. Before the
real suite, it runs a build-tagged intentional race and refuses to proceed unless the Go detector emits
`DATA RACE`.

### Components and files changed

`.github/workflows/ci.yml`, `Makefile`, `miscellaneous/check_race_detector.sh`, and
`internal/verificationfixtures/racefixture/`.

### Interfaces and configuration

New commands: `make check-race-fixture` and `make test-race-integration`. CI starts PostgreSQL, MinIO,
and Mailpit, then passes DB/S3 requirement flags into the toolbox.

### Tests

Seeded negative fixture plus uncached real `-race -count=1` execution against the seven packages.

### Revision, date, debt, and plan relationship

Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus scoped shared-worktree provenance; implemented
2026-07-14; no debt. The planned per-PR versus scheduled choice resolved to per-change scoped CI.

## Verification Record

| Acceptance criterion | Actual result | Result | Evidence | Reviewer |
|---|---|---|---|---|
| AC-W07-E02-S002-03 | Intentional race detected; all seven real DB/S3 integration packages passed under `-race`. | PASS | EV-W07-E02-S002-003 | W05ReviewGateFinal: PASS |

Environment: Darwin arm64, Go 1.26.5, real PostgreSQL/MinIO, both requirement flags. Retest passed on
2026-07-14. Final task conclusion: verified and artifact/evidence registered.

## Deviations Record

No plan divergence; the per-change scoped choice resolves a planned open question.
