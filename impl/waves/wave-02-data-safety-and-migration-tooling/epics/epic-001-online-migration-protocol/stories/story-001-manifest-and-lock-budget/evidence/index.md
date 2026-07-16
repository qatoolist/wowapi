---
id: W02-E01-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W02-E01-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W02-E01-S001 — Evidence index

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W02-E01-S001-001 | schema-validation report (positive/negative fixture pair + kernel ledger enforcement) | W02-E01-S001-T001 | AC-W02-E01-S001-01 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/migration/... ./migrations/... -run 'TestParseManifest\|TestValidate\|TestMigrationVersion\|TestKernelMigrationsHaveManifests' -count=1 -v` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced |
| EV-W02-E01-S001-002 | review report | W02-E01-S001-T001 | AC-W02-E01-S001-02 | Independent review completed via W02Proto.ManifestSchemaReview (peer reviewer). Schema locked in `artifacts/manifest-schema-design.md`. | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced |
| EV-W02-E01-S001-003 | integration-test report (concurrently-locked-table lock-timeout abort/retry) | W02-E01-S001-T002 | AC-W02-E01-S001-03 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/migration/... -run 'TestExecDDL' -count=1 -v` | 1626b1132622aacc3e85475e4190e16a457ad1f6 | pass | produced |
| EV-W02-E01-S001-004 | review report (independent review, task-003, re-verifying AC-01/AC-02/AC-03 against real code and a fresh test re-run; supplements the unverifiable EV-002 prose claim as the operative artifact-backed evidence for AC-02) | W02-E01-S001-T003 | AC-W02-E01-S001-01, AC-W02-E01-S001-02, AC-W02-E01-S001-03 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable go test ./kernel/migration/... -run 'TestExecDDLLockTimeoutAbortAndRetry\|TestManifest' -v -count=1` | HEAD 43b6e12 + remediation working tree 2026-07-16 | pass (AC-01, AC-03 confirmed; AC-02 evidence gap noted, see task-003 Findings) | produced |

All evidence outputs are captured in `evidence/tests/`. Reviewer for EV-004: Independent review
agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy remediation R-3).
Environment: macOS (darwin/arm64), go1.26.5. Superseded evidence: EV-004 does not delete or
overwrite EV-002; it supplements it as the newly-produced, artifact-backed record for AC-02.
