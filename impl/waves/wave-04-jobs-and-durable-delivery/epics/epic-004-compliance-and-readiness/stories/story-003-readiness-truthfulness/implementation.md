---
id: IMPL-W04-E04-S003
type: implementation-record
parent_story: W04-E04-S003
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W04-E04-S003

## What was actually implemented

- **T1 — migration-currency readiness check:** Added `app.MigrationCurrencyCheck` and
  `app.MigrationVersionDetail`, wired them through `app.ReadinessWithCatalogs`, and updated the
  generated `cmd/api/main.go.tmpl` to pass `migrations.Kernel()` / `migrations.SourceName`. Added
  migration `00042_goose_version_platform_select.sql` granting `app_platform` SELECT on
  `goose_version_wowapi` so the runtime readiness check can read the applied version.
- **T2 — seed/rule/model-hash readiness reporting:** Extended `app.ReadinessWithCatalogs` with detail
  providers for `migration_version`, `seed_catalog_hash`, `rule_hash`, and `model_hash`. Implemented
  `app.RuleHash` as a deterministic SHA-256 over the sorted rule-point registry. `model_hash` is a
  placeholder detail driven by `kernel.Kernel.ModelHash`; it is omitted until AR-01 lands (recorded
  as deviation DEV-W04-E04-S003-001).
- **T3 — `config doctor` product-root discovery:** Replaced CWD-relative `os.Stat` in
  `internal/cli/config_delegate.go` with `--project` / `go env GOMOD` discovery. The product checker
  now runs from the discovered root regardless of invocation directory, and stderr explicitly
  reports `product validation: engaged (<root>)` or `product validation: skipped (...)`.

DX-07 T4 (production-profile capacity/backpressure enforcement) was explicitly left out of scope,
per `story.md` and `plan.md`.

## Components changed

- `app/health.go` — readiness assembly, migration-currency check, hash detail providers.
- `kernel/kernel.go` — added `ModelHash` field (placeholder for AR-01).
- `internal/cli/templates/init/cmd_api_main.go.tmpl` — wires migration source into readiness.
- `internal/cli/config_delegate.go` — `go env GOMOD`/`--project` discovery and explicit reporting.
- `internal/cli/config_cmd.go` — added `--project` shared flag and usage docs.
- `migrations/00042_goose_version_platform_select.sql` — runtime SELECT grant for migration version.

## Files changed

- `app/health.go`
- `app/seed_readiness_test.go`
- `app/health_readiness_test.go` (new)
- `kernel/kernel.go`
- `internal/cli/templates/init/cmd_api_main.go.tmpl`
- `internal/cli/config_delegate.go`
- `internal/cli/config_cmd.go`
- `internal/cli/config_delegate_test.go` (new)
- `internal/cli/scaffold_test.go`
- `migrations/00042_goose_version_platform_select.sql` (new)
- `migrations/migrations_test.go`

## Interfaces introduced or changed

- `func app.ReadinessWithCatalogs(b *Booted, fingerprint config.Fingerprint, db database.DBTX, src fs.FS, source string, extra map[string]httpx.HealthCheck) *httpx.Health` — signature extended with migration source.
- `func app.MigrationCurrencyCheck(db database.DBTX, src fs.FS, source string) httpx.HealthCheck`
- `func app.MigrationVersionDetail(db database.DBTX, source string) httpx.DetailProvider`
- `func app.RuleHash(r *rules.Registry) string`
- `func app.MaxMigrationVersion(src fs.FS) (int64, error)`

## Configuration changes

None.

## Schema or migration changes

- `migrations/00042_goose_version_platform_select.sql` adds a single `GRANT SELECT ON goose_version_wowapi TO app_platform`.

## Security changes

- `app_platform` gains read access to `goose_version_wowapi` solely to support the migration-currency readiness check. No write access is granted.

## Observability changes

- `/readyz` payload now includes `details.migration_version`, `details.seed_catalog_hash`, and `details.rule_hash` when applicable.
- `wowapi config doctor` (and other delegated config subcommands) now prints explicit product-validation status to stderr.

## Tests added or modified

- `app/health_readiness_test.go`: migration-currency pass/fail integration tests and full readiness payload test.
- `internal/cli/config_delegate_test.go`: nested-subdirectory, outside-repo-with-`--project`, and skipped-product-validation unit tests.
- `app/seed_readiness_test.go`: updated `ReadinessWithCatalogs` call sites for new signature.
- `internal/cli/scaffold_test.go`: asserts generated api main contains `migration_currency`.
- `migrations/migrations_test.go`: updated expected file list.

## Implementation dates

2026-07-13

## Technical debt introduced

None.

## Known limitations

- `model_hash` is not reported until AR-01 (W05-E01-S003) populates `kernel.Kernel.ModelHash`.

## Follow-up items

- Once AR-01 lands, add an assertion for `model_hash` in `app/health_readiness_test.go` and resolve
  DEV-W04-E04-S003-001.

## Relationship to the approved plan

Matches `plan.md` with one deviation: T2's model-hash portion is partially deferred pending AR-01,
recorded in `deviations.md` as DEV-W04-E04-S003-001. No task attempted DX-07 T4.
