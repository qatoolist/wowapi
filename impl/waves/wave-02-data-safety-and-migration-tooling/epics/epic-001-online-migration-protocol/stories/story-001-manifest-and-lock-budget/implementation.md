---
id: IMPL-W02-E01-S001
type: implementation-record
parent_story: W02-E01-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W02-E01-S001

## What was actually implemented

- Inline `+wowapi:manifest` / `+wowapi:end` block parser and validator in
  `kernel/migration/manifest.go`.
- Manifest enforcement gate in `migrations/manifest_test.go` requiring every
  kernel migration ≥ 00031 to carry a valid manifest.
- Lock-timeout DDL executor in `kernel/migration/locktimeout.go` with a 2-second
  online budget, bounded retry ceiling (3 retries), and clean abort on SQLSTATE
  55P03.
- Manifest block added to the existing `migrations/00031_seed_sync_runs.sql`
  (W02-E05) so the new enforcement gate passes against the current working tree.

## Components changed

- New package `kernel/migration` (protocol tooling foundation).
- `migrations/manifest_test.go` (CI gate).
- `migrations/00031_seed_sync_runs.sql` (manifest block).
- `migrations/migrations_test.go` (expectedFiles updated for 00031).

## Files changed

- `kernel/migration/manifest.go`
- `kernel/migration/manifest_test.go`
- `kernel/migration/locktimeout.go`
- `kernel/migration/locktimeout_test.go`
- `migrations/manifest_test.go`
- `migrations/00031_seed_sync_runs.sql`
- `migrations/migrations_test.go`

## Interfaces introduced or changed

- `migration.Manifest` data contract.
- `migration.ParseManifest`, `Manifest.Validate`, `MigrationVersion`.
- `migration.ExecDDL` with budget/retry parameters.

## Configuration changes

None.

## Schema or migration changes

No application schema changes. Added manifest block to an existing migration.

## Security changes

Bounded retry ceiling enforced in `ExecDDL` to prevent deploy-time DoS from
unbounded lock-timeout retries.

## Observability changes

Lock-timeout abort/retry events are logged via `slog`.

## Tests added or modified

- `kernel/migration/manifest_test.go` — positive/negative fixture tests.
- `kernel/migration/locktimeout_test.go` — concurrent-lock abort/retry tests.
- `migrations/manifest_test.go` — kernel migration ledger enforcement.

## Commits

Working tree at base commit `1626b1132622aacc3e85475e4190e16a457ad1f6`.

## Pull requests

Not tracked in this session.

## Implementation dates

2026-07-13.

## Technical debt introduced

None.

## Known limitations

None.

## Follow-up items

None.

## Relationship to the approved plan

Matches `plan.md`. One deviation recorded: manifest block added to W02-E05's
`00031_seed_sync_runs.sql` to satisfy the ≥00031 enforcement boundary
(`deviations.md`).
