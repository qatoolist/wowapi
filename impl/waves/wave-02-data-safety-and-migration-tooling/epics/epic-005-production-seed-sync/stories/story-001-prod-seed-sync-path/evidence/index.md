---
id: W02-E05-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W02-E05-S001
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-16
---

# W02-E05-S001 — Evidence index

All evidence was produced against base commit `1626b113` with the FBL-02 implementation changes in
the working tree (no git writes performed per session constraint). PostgreSQL test instance:
`postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W02-E05-S001-001 | design-decision record | W02-E05-S001-T001 | AC-W02-E05-S001-01 | N/A (document record) | 1626b113 + WIP | Decision record complete, predates implementation | produced |
| EV-W02-E05-S001-002 | integration-test report | W02-E05-S001-T002 | AC-W02-E05-S001-02 | `go test ./kernel/seeds/... -run 'TestApplyIdempotentNoop\|TestApplyRLSPostureRespectsPlatformRole\|TestApplyHashStableAcrossOrdering\|TestApplyHashExcludesVersionLabel\|TestLoadParsesVersion\|TestLoadRejectsConflictingVersion'` | 1626b113 + WIP | Idempotency proven (noop second run, xmin unchanged); RLS posture (app_platform, no BYPASSRLS); version conflict rejected | produced |
| EV-W02-E05-S001-003 | integration-test report | W02-E05-S001-T003 | AC-W02-E05-S001-03 | `go test ./kernel/seeds/... -run 'TestApplyDryRunNoWrites\|TestApplyRecordsAuditRow' && go test ./internal/cli/... -run 'TestSeedSyncDBDryRun'` | 1626b113 + WIP | Dry-run writes nothing and emits plan; audit row recorded with hash/actor/counts | produced |
| EV-W02-E05-S001-004 | integration-test report (fail-first pair) | W02-E05-S001-T004 | AC-W02-E05-S001-05 | Before probe: `app/zz_cs21_before_probe_test.go` (removed after capture); after: `go test ./app/... -run 'TestIntegrationReadinessEmptyCatalogsFailsNamed\|TestIntegrationReadinessAfterSyncReportsHash'` | 1626b113 + WIP | Before: 200/ready with no seed_catalogs check; after: 503/not_ready named failure, then 200/ready | produced |
| EV-W02-E05-S001-005 | integration-test report | W02-E05-S001-T005 | AC-W02-E05-S001-05 | `go test ./app/... -run 'TestIntegrationReadinessAfterSyncReportsHash'` | 1626b113 + WIP | Readiness payload includes `details.seed_catalog_hash` matching the applied manifest hash | produced |
| EV-W02-E05-S001-006 | review report | W02-E05-S001-T006 | AC-W02-E05-S001-01..-05 | Independent review checklist | 1626b113 + WIP | Review completed, no open issues (see `006-independent-review/review-report.md`) | **superseded** by EV-W02-E05-S001-007 on 2026-07-16 (autopsy finding M-1: EV-006 lacks every mandatory evidence-policy field — evidence ID, execution command, commit SHA, branch/tag, environment, tool versions, date/time, reviewer, file/URI — and per `impl/governance/evidence-policy.md` "must not be cited as proof of an acceptance criterion." EV-006 is preserved unmodified, not deleted, per the failed-evidence preservation rule.) |
| EV-W02-E05-S001-007 | review report (independent review, field-complete remediation of EV-006; supersedes it) | W02-E05-S001-T006 | AC-W02-E05-S001-01, AC-W02-E05-S001-02, AC-W02-E05-S001-03, AC-W02-E05-S001-04, AC-W02-E05-S001-05, AC-W02-E05-S001-06 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable go test ./kernel/seeds/... -v -count=1 ; go test ./app/... -run 'TestIntegrationReadinessEmptyCatalogsFailsNamed\|TestIntegrationReadinessAfterSyncReportsHash' -v -count=1` | HEAD 43b6e12 + remediation working tree 2026-07-16 | pass (14/14 kernel/seeds tests + 2/2 readiness tests) | produced, see `007-independent-review-remediation/review-report.md` |

## Evidence locations

- `001-design-decision/` — points to `../artifacts/pre-implementation/design-decision-record.md`.
- `002-idempotency-rls/` — `apply-tests.log`.
- `003-dry-run-audit/` — `apply-tests.log` and `cli-tests.log`.
- `004-failfirst-readiness/` — `before-probe.log` and `after-readiness.log`.
- `005-hash-reporting/` — `readiness-hash.log`.
- `006-independent-review/` — `review-report.md`.

## Notable outcomes

- The pre-fix fail-first probe was captured by temporarily fixing and running
  `app/zz_cs21_before_probe_test.go`, then removing it. The post-fix behavior is asserted permanently
  by `app/seed_readiness_test.go`.
- No core dumps or crash logs were found for the previous W02Seed attempt; see `006-independent-review/review-report.md`.
