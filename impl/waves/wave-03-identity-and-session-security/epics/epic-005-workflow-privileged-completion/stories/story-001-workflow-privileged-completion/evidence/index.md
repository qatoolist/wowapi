---
id: W03-E05-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W03-E05-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W03-E05-S001 — Evidence index

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W03-E05-S001-001 | ratification rejection-boundary test | W03-E05-S001-T001 | AC-W03-E05-S001-01 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/workflow/... -count=1 -run TestRatifyByDefinitionRejected -v` | 733ef3e930cbb3f89f5bbc53d8f562c60e426513 (working tree) | PASS | accepted |
| EV-W03-E05-S001-002 | audit-present/complete test | W03-E05-S001-T002 | AC-W03-E05-S001-02 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/workflow/... -count=1 -run TestOverrideAuditRowPresent -v` | 733ef3e930cbb3f89f5bbc53d8f562c60e426513 (working tree) | PASS | accepted |
| EV-W03-E05-S001-003 | fault-injection/rollback test | W03-E05-S001-T002 | AC-W03-E05-S001-02 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/workflow/... -count=1 -run TestOverrideAuditFailureRollsBack -v` | 733ef3e930cbb3f89f5bbc53d8f562c60e426513 (working tree) | PASS | accepted |
| EV-W03-E05-S001-004 | T1–T3 regression confirmation | W03-E05-S001-T003 | AC-W03-E05-S001-03 | `DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable WOWAPI_REQUIRE_DB=1 go test ./kernel/workflow/... -count=1 -run 'TestIntegrationOverrideAuthzGate|TestIntegrationOverrideFailsClosedWithoutPermission|TestIntegrationWorkflowOverride' -v` | 733ef3e930cbb3f89f5bbc53d8f562c60e426513 (working tree) | PASS | accepted |

AC-W03-E05-S001-03 is additionally covered by the full `./kernel/workflow/...` suite passing with
`-count=1`.
