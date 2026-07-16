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
| EV-W04-E04-S003-001 | stale-migration integration-test report | W04-E04-S003-T001 | AC-W04-E04-S003-01 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable go test ./app/... -run TestIntegrationMigrationCurrencyCheckFailsWhenStale -count=1 -v` | HEAD 43b6e12 + remediation working tree 2026-07-16 | PASS (503 on stale DB) | retested (previous record cited "HEAD," a moving target, not a pinned SHA, per evidence-policy.md's revision-pinning rule — superseded by this pinned re-run) |
| EV-W04-E04-S003-002 | full-readiness-payload integration-test report | W04-E04-S003-T002 | AC-W04-E04-S003-02 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable go test ./app/... -run TestIntegrationReadinessReportsSeedAndRuleHashes -count=1 -v` | HEAD 43b6e12 + remediation working tree 2026-07-16 | PASS (migration_version, seed_catalog_hash, rule_hash present; model_hash pending AR-01 per `deviations.md` DEV-W04-E04-S003-001) | retested (supersedes the prior "HEAD"-cited record) |
| EV-W04-E04-S003-003 | config-doctor discovery test report (nested-subdirectory + outside-repo --project) | W04-E04-S003-T003 | AC-W04-E04-S003-03 | `go test ./internal/cli/... -run 'TestConfigDoctorDiscoversProductRootFromNestedSubdir|TestConfigDoctorDiscoversProductRootFromOutsideRepo|TestConfigDoctorReportsSkippedProductValidation' -count=1 -v` | HEAD 43b6e12 + remediation working tree 2026-07-16 | PASS — all 3 subtests (discovery works from a nested subdir and from outside the repo via `--project`; product-validation-ran status is reported explicitly in both success and skipped cases) | retested (supersedes the prior "HEAD"-cited record; also corrects the execution command, which previously named a non-matching pattern `TestConfigDoctorDiscoversProductRoot` — the actual test names are the two `...FromNestedSubdir`/`...FromOutsideRepo` funcs) |

Reviewer: Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor
(autopsy remediation R-3). Date: 2026-07-16. Full analysis in
`tasks/task-004-independent-review.md`'s Verification Record.
