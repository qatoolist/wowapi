---
id: IMPL-W02-E05-S001
type: implementation-record
parent_story: W02-E05-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W02-E05-S001

## What was actually implemented

FBL-02 production seed-sync path (MATRIX CS-21): idempotent, RLS-respecting, versioned-manifest-driven
sync with dry-run + audit, plus framework-level readiness wiring that reports the seed/catalog hash.

## Components changed

- `kernel/seeds` — added manifest `version`, canonical `Hash`, new `Apply` entrypoint (advisory-lock
  guarded, hash short-circuit, audit record), and `DryRun` change-plan output.
- `app` — added `ReadinessWithCatalogs` which auto-registers the `seed_catalogs` check and a dynamic
  `seed_catalog_hash` detail provider.
- `kernel/httpx` — added `DetailProvider` support so readiness can report dynamic details.
- `internal/cli` — `wowapi seed sync` now uses `seeds.Apply`, supports `--dry-run`.
- Generated templates — `cmd/migrate` calls `seeds.Apply`; `cmd/api` uses `app.ReadinessWithCatalogs`.
- Database — migration `00031_seed_sync_runs.sql` creates the append-only global audit/state table.

## Files changed

- `kernel/seeds/seeds.go` — `Bundle.Version`, `Hash`, version-conflict handling.
- `kernel/seeds/apply.go` — new `Apply`, `ApplyOptions`, `Report`, dry-run diff, audit insert.
- `kernel/seeds/apply_test.go`, `seeds_test.go` — new tests.
- `app/seed_health.go` — hash lookup, `latestSeedHash`.
- `app/health.go` — `ReadinessWithCatalogs`.
- `app/seed_readiness_test.go` — CS-21 fail-first/pass-after tests.
- `kernel/httpx/health.go` — `DetailProvider`, details in readiness payload.
- `internal/cli/seed_cmd.go` — `--dry-run`, `seeds.Apply`.
- `internal/cli/seed_cmd_db_test.go` — dry-run DB test.
- `internal/cli/scaffold_test.go`, `seed_lifecycle_drift_test.go`, `e2e_scaffold_harness_test.go` —
  updated generated-template assertions and unique E2E release version to avoid stale module cache.
- `internal/cli/templates/init/cmd_api_main.go.tmpl`, `cmd_migrate_main.go.tmpl` — use new APIs.
- `migrations/00031_seed_sync_runs.sql` — new audit/state table.
- `docs/user-guide/cli-reference.md`, `docs/user-guide/database-migrations.md`,
  `docs/operations/deployment-checklist.md` — updated.
- Removed `app/zz_cs21_before_probe_test.go` after capturing the before-state log.

## Interfaces introduced or changed

- `seeds.Apply(ctx, db, bundle, opts) (Report, error)` — new public entrypoint.
- `seeds.ApplyOptions`, `seeds.Report`, `seeds.ApplyCounts`, `seeds.ChangePlan` — new types.
- `seeds.Hash(Bundle) string` — new public function.
- `app.ReadinessWithCatalogs(b, fingerprint, db, extra)` — new public assembly.
- `httpx.Health.Detail(fn DetailProvider)` — new method.

## Configuration changes

None.

## Schema or migration changes

- `migrations/00031_seed_sync_runs.sql` adds global `seed_sync_runs` table with grants:
  `app_platform` SELECT+INSERT, `app_rt` none.

## Security changes

- Sync continues to run as `app_platform`; `Apply` validates the role is not `BYPASSRLS` in tests.
- `app_rt` cannot write `seed_sync_runs`.
- Advisory xact lock serializes concurrent `Apply` calls.

## Observability changes

- `/readyz` now includes `details.seed_catalog_hash` after a successful sync.
- `seed_sync_runs` audit rows carry `manifest_hash`, `version_label`, `actor`, `outcome`, `counts`,
  `error`, `created_at`.

## Tests added or modified

- `kernel/seeds/apply_test.go` — idempotency, dry-run no-writes, audit row, hash stability, version
  exclusion, RLS posture.
- `kernel/seeds/seeds_test.go` — version parsing/conflict.
- `app/seed_readiness_test.go` — empty-catalog named readiness failure, hash reporting after sync.
- `internal/cli/seed_cmd_db_test.go` — CLI dry-run DB test.
- Updated existing `internal/cli` scaffold/drift tests for `seeds.Apply` / `ReadinessWithCatalogs`.

## Commits

Working-tree changes on base commit `1626b113` (no git writes per session constraint).

## Pull requests

None in this session.

## Implementation dates

2026-07-13.

## Technical debt introduced

None anticipated.

## Known limitations

- Dry-run output is human-readable only; machine-readable format deferred until a consumer exists
  (per T001 decision Q8).
- Failed-run audit records are best-effort and only produced when `Apply` owns the transaction
  (i.e. a pool, not a caller-supplied tx).

## Follow-up items

- None within story scope. DX-07 migration-currency readiness and prod-profile capacity enforcement
  remain in W04-E04-S003.

## Relationship to the approved plan

Implementation matches the T001 design record (`artifacts/pre-implementation/design-decision-record.md`).
One plan revision was recorded in that document: execution order T003→T002→T005→T004, with audit folded
into `Apply` rather than a separate code path. No deviations beyond the documented design revision.
