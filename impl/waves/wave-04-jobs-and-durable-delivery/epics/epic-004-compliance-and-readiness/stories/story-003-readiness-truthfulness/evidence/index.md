---
id: W04-E04-S003-EVIDENCE-INDEX
type: evidence-index
parent_story: W04-E04-S003
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04-S003 — Evidence index

No evidence item for DX-07 T4 exists in this index — T4 is explicitly out of scope (see `story.md`).

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W04-E04-S003-001 | stale-migration integration-test report | W04-E04-S003-T001 | AC-W04-E04-S003-01 | `go test ./app/... -run TestIntegrationMigrationCurrencyCheckFailsWhenStale -count=1 -v` | HEAD | PASS (503 on stale DB) | produced |
| EV-W04-E04-S003-002 | full-readiness-payload integration-test report | W04-E04-S003-T002 | AC-W04-E04-S003-02 | `go test ./app/... -run TestIntegrationReadinessReportsSeedAndRuleHashes -count=1 -v` | HEAD | PASS (migration_version, seed_catalog_hash, rule_hash present; model_hash pending AR-01) | produced |
| EV-W04-E04-S003-003 | config-doctor discovery test report (nested-subdirectory + outside-repo --project) | W04-E04-S003-T003 | AC-W04-E04-S003-03 | `go test ./internal/cli/... -run 'TestConfigDoctorDiscoversProductRoot|TestConfigDoctorReportsSkippedProductValidation' -count=1 -v` | HEAD | PASS (discovery works and reports status) | produced |
