---
id: W02-E01-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W02-E01-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W02-E01-S002 — Evidence index

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W02-E01-S002-001 | compatibility-test report (old-reader-compatibility) | W02-E01-S002-T001 | AC-W02-E01-S002-01 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/migration/... -run 'TestExpandPhaseOldReaderCompatibility' -count=1 -v` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced |
| EV-W02-E01-S002-002 | integration-test report (named interrupted/resumed backfill test, `DATA-09/backfill-interrupt-resume/`) | W02-E01-S002-T002 | AC-W02-E01-S002-02 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/migration/... -run 'TestBackfillInterruptedAndResumed' -count=1 -v` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced |
| EV-W02-E01-S002-003 | artifact-schema test report | W02-E01-S002-T003 | AC-W02-E01-S002-03 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/migration/... -run 'TestValidationArtifactSchema' -count=1 -v` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced |

All evidence outputs are captured in `evidence/tests/`.
