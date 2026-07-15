---
id: CLOSURE-W04-E04-S003
type: closure-record
parent_story: W04-E04-S003
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Closure — W04-E04-S003

## Acceptance-criteria completion

- **AC-W04-E04-S003-01**: PASS. `/readyz` fails (503) when applied migration version lags expected.
  `TestIntegrationMigrationCurrencyCheckFailsWhenStale` boots against a rewound
  `goose_version_wowapi` and asserts 503.
- **AC-W04-E04-S003-02**: PASS with deviation. Readiness reports `migration_version`,
  `seed_catalog_hash`, and `rule_hash`. `model_hash` is implemented as a placeholder driven by
  `kernel.Kernel.ModelHash` but is omitted because AR-01 has not yet landed; recorded as
  DEV-W04-E04-S003-001.
- **AC-W04-E04-S003-03**: PASS. `config doctor` discovers the product root via `go env GOMOD` or
  `--project`, works from nested subdirectories and outside the repo, and explicitly reports whether
  product validation ran. Proven by `internal/cli/config_delegate_test.go`.

## Task completion

- W04-E04-S003-T001 — Migration-currency readiness check: COMPLETE.
- W04-E04-S003-T002 — Seed/rule/model-hash readiness reporting: COMPLETE (model-hash portion
  deferred pending AR-01, recorded in `deviations.md`).
- W04-E04-S003-T003 — `config doctor` product-root discovery fix: COMPLETE.
- W04-E04-S003-T004 — Independent review: PENDING.

## Artifact completeness

- Migration-currency readiness check: `app/health.go`, `internal/cli/templates/init/cmd_api_main.go.tmpl`.
- Seed/rule/model-hash readiness reporting: `app/health.go`.
- `config doctor` discovery fix: `internal/cli/config_delegate.go`, `internal/cli/config_cmd.go`.
- DX-07 T4 out-of-scope documentation: inline comments and this closure record.

## Evidence completeness

- Stale-migration 503 test: `go test ./app/... -run
  TestIntegrationMigrationCurrencyCheckFailsWhenStale -count=1 -v`.
- Full readiness payload test: `go test ./app/... -run
  TestIntegrationReadinessReportsSeedAndRuleHashes -count=1 -v`.
- Config-doctor discovery tests: `go test ./internal/cli/... -run
  'TestConfigDoctorDiscoversProductRoot|TestConfigDoctorReportsSkippedProductValidation' -count=1 -v`.

## Unresolved findings

None for T1-T3. T2's `model_hash` reporting is intentionally unresolved pending AR-01 and is
recorded as a deviation, not a silent gap.

## Accepted risks

- RISK-W04-004 (DX-07 T4's forward dependency on W05-E03-S002's waiver mechanism): remains open by
  design. This story correctly scoped T4 out and forward-referenced W05-E03-S002 / AR-04 T5.

## Deferred work

- DX-07 T4 (production-profile capacity/backpressure enforcement) is deferred to W05-E03-S002's
  AR-04 T5 waiver mechanism.
- PROD-03 (wowsociety's readiness backport) is a non-blocking coordination note, not implemented
  here.
- AR-01's `model_hash` integration is a forward dependency; the readiness detail provider will
  start reporting `model_hash` automatically once `kernel.Kernel.ModelHash` is populated.

## Reviewer conclusion

PENDING. The reviewer must confirm:
1. DX-07 T4 was correctly and explicitly scoped out — no task silently attempts capacity
   enforcement.
2. The forward reference to W05-E03-S002 / AR-04 T5 is present and not dropped.

## Acceptance authority

Data/reliability lead, per epic-level `acceptance.md`.

## Closure date

2026-07-13 (framework-side implementation and verification complete; pending independent review).

## Final status

`closed-pending-review` — T1-T3 implemented and evidenced; awaiting mandatory independent review per
mandate §14. DX-07 is NOT claimed complete; only T1-T3 are in scope.
