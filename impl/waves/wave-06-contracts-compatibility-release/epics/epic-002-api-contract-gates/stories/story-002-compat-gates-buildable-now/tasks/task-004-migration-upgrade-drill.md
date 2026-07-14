---
id: W06-E02-S002-T004
type: task
title: Migration upgrade-from-oldest-supported drill
status: done
parent_story: W06-E02-S002
owner: W06E02Impl
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W06-E02-S002-04
artifacts:
  - ART-W06-E02-S002-004
evidence:
  - EV-W06-E02-S002-004
---

# W06-E02-S002-T004 — Migration upgrade-from-oldest-supported drill

## Task Definition

### Task objective

Extend TestIntegrationMigrationsReversible into an upgrade-from-oldest-supported-version drill.

### Parent story

W06-E02-S002

### Owner

W06E02Impl

### Status

done

### Dependencies

None.

### Detailed work

1. Extend TestIntegrationMigrationsReversible to seed at the oldest supported version.
2. Migrate forward to current.
3. Reverse on disposable data, confirming reversibility.

### Expected files or components affected

Extension to the existing TestIntegrationMigrationsReversible test file.

### Expected output

An extended reversibility test proving upgrade-from-oldest-supported works end to end.

### Required artifacts

ART-W06-E02-S002-004 (extended migration upgrade-drill test).

### Required evidence

EV-W06-E02-S002-004 (migration upgrade-drill test report).

### Related acceptance criteria

AC-W06-E02-S002-04.

### Completion criteria

Seed at oldest version, migrate forward, reverse on disposable data, all succeed.

### Verification method

Direct execution of the extended test against real Postgres.

### Risks

None beyond standard migration-tooling integration risk (this story reuses DATA-09's protocol where relevant, per W02's own tooling, rather than re-deriving migration mechanics).

### Rollback or recovery considerations

If reversibility fails for a specific migration, escalate to that migration's own owner rather than silently marking it irreversible without investigation.

## Implementation Record

Implemented and real-Postgres verified.

### What was actually implemented

The integration drill rolls to the v1.0.0 migration head, seeds disposable tenant data, upgrades to current, verifies preservation, reverses, and reconstructs current head.

### Components changed

Migration API and reversibility integration test.

### Files changed

`kernel/database/migrate.go`; `migrations/reversible_test.go`.

### Interfaces introduced or changed

Adds `database.MigrateTo` for controlled target-version migration.

### Configuration changes

Uses the supplied DB/S3-required integration environment.

### Schema or migration changes

No new migration; the test exercises existing migration history.

### Security changes

Disposable isolated test database only.

### Observability changes

None.

### Tests added or modified

Extended `TestIntegrationMigrationsReversible` with oldest-supported seed and preservation assertions.

### Commits

No commit; working tree based on `733ef3e`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

*None anticipated.*

### Known limitations

Requires real PostgreSQL.

### Follow-up items

None.

### Relationship to the approved plan

Matches the approved oldest-supported upgrade/reversibility plan.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E02-S002-04 | Run the extended reversibility test | CI, real Postgres | Seed-forward-reverse cycle succeeds | integration-test report | unassigned |

### Actual result

v1.0.0 seed, forward upgrade, data preservation, full reverse, and reconstruction all succeeded.

### Pass or fail

PASS.

### Evidence identifier

EV-W06-E02-S002-004.

### Execution date

2026-07-13.

### Commit or revision

Working tree based on `733ef3e`.

### Environment

Real PostgreSQL with supplied DATABASE_URL and both required-infrastructure flags.

### Reviewer

W06-E02-S002-Rerun reran the real PostgreSQL drill — PASS.

### Findings

No open functional finding.

### Retest status

Focused integration retest PASS.

### Final conclusion

Implemented, independently verified, and accepted.

## Deviations Record

*No deviations recorded yet.*

### Deviation ID

*Not applicable.*

### Approved plan

*Not applicable.*

### Actual implementation

*Not applicable.*

### Reason

*Not applicable.*

### Impact

*Not applicable.*

### Risks

*Not applicable.*

### Approval

*Not applicable.*

### Compensating controls

*Not applicable.*

### Follow-up work

*Not applicable.*
