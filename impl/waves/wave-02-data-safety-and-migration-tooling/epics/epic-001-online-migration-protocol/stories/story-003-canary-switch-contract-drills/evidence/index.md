---
id: W02-E01-S003-EVIDENCE-INDEX
type: evidence-index
parent_story: W02-E01-S003
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W02-E01-S003 — Evidence index

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W02-E01-S003-001 | integration-test report (canary named test — N-1/N both legs + partial fleet rollout) | W02-E01-S003-T001 | AC-W02-E01-S003-01 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/migration/... -run 'TestCanaryNAndNMinusOne\|TestPartialFleetRollout' -count=1 -v` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced |
| EV-W02-E01-S003-002 | integration-test report (switch-rollback named test) | W02-E01-S003-T002 | AC-W02-E01-S003-02 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/migration/... -run 'TestSwitchRollbackAfterSwitch' -count=1 -v` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced |
| EV-W02-E01-S003-003 | integration-test report (contract-gate named test, both required properties) | W02-E01-S003-T003 | AC-W02-E01-S003-03 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/migration/... -run 'TestContractGateAndForwardRecovery' -count=1 -v` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced |
| EV-W02-E01-S003-004 | CI pipeline run artifact (all six drills) | W02-E01-S003-T004 | AC-W02-E01-S003-04 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/migration/... -count=1` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced |
| EV-W02-E01-S003-005 | consolidated evidence bundle (aggregates EV-001 through EV-004) | W02-E01-S003-T005 | AC-W02-E01-S003-04 | Not applicable (aggregation artifact) | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced |

The six directive-named drills covered are:
1. N-1 code on expanded N schema — `TestCanaryNAndNMinusOne` (N-1 leg)
2. N code before/after backfill — `TestCanaryNAndNMinusOne` (N legs)
3. Interrupted/resumed backfill — `TestBackfillInterruptedAndResumed`
4. Partial fleet rollout — `TestPartialFleetRollout`
5. Application rollback after switch — `TestSwitchRollbackAfterSwitch`
6. Forward recovery from every failed phase + delayed contract gate — `TestContractGateAndForwardRecovery`

All evidence outputs are captured in `evidence/tests/` and `evidence/pipeline/`.
The CI drill pipeline definition is `.github/workflows/migration-drills.yml`.
