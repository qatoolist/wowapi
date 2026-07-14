---
id: IMPL-W02-E01-S002
type: implementation-record
parent_story: W02-E01-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W02-E01-S002

## What was actually implemented

- Expand-phase DDL helpers in `kernel/migration/expand.go` (nullable/default-safe
  column, NOT VALID check constraint, CREATE INDEX CONCURRENTLY, compatibility
  view).
- Resumable, tenant-scoped, keyset-paginated backfill harness in
  `kernel/migration/backfill.go` with an interim checkpoint-lease primitive
  (`migration.backfill_checkpoint`).
- Validation-phase tooling in `kernel/migration/validate.go` (VALIDATE CONSTRAINT
  wrapper, reconciliation query, JSON artifact schema).

## Components changed

- `kernel/migration/expand.go`
- `kernel/migration/backfill.go`
- `kernel/migration/validate.go`

## Files changed

- `kernel/migration/expand.go`
- `kernel/migration/expand_test.go`
- `kernel/migration/backfill.go`
- `kernel/migration/backfill_test.go`
- `kernel/migration/validate.go`
- `kernel/migration/validate_test.go`

## Interfaces introduced or changed

- `ExpandPhase` helper methods.
- `BackfillConfig`, `Backfill.Run`, `ProcessBatch`.
- `ValidationReport`, `ValidationCheck`, `Reconcile`, `ValidateConstraint`.

## Configuration changes

Backfill batch size, rate limit, and runtime window are configurable via
`BackfillConfig`.

## Persistence changes

New `migration.backfill_checkpoint` table created lazily by
`EnsureCheckpointTable`.

## Migration strategy

No application data migration performed; this story adds tooling.

## Concurrency implications

The backfill harness commits the checkpoint in the same transaction as the
batch, so an interruption leaves a resumable position. The interrupted/resumed
named test proves no reprocessing or skipping.

## Error-handling strategy

Batch errors rollback the transaction; `ErrBackfillStopped` is the only path
that commits and halts cleanly.

## Security changes

None beyond S001's lock-timeout wrapper.

## Observability changes

Checkpoint state is queryable in `migration.backfill_checkpoint`.

## Tests added or modified

- `TestExpandPhaseOldReaderCompatibility`
- `TestBackfillInterruptedAndResumed`
- `TestBackfillTenantScoped`
- `TestValidationArtifactSchema`

## Commits

Working tree at base commit `1626b1132622aacc3e85475e4190e16a457ad1f6`.

## Pull requests

Not tracked in this session.

## Implementation dates

2026-07-13.

## Technical debt introduced

The interim checkpoint-lease is a deliberate, scope-bounded substitute for
DATA-02 T1's full shared lease/fencing primitive. It provides only checkpoint
token + resumability; it does not provide job-claim fencing or heartbeats.
Forward reference to W04-E01-S001 recorded in `deviations.md` and `closure.md`.

## Known limitations

The interim checkpoint-lease will be replaced by W04-E01-S001.

## Follow-up items

- W04-E01-S001: migrate `migration.backfill_checkpoint` onto the shared
  lease/fencing primitive.

## Relationship to the approved plan

Matches `plan.md`. The interim-lease scope boundary is explicitly documented in
source comments and in this record.
