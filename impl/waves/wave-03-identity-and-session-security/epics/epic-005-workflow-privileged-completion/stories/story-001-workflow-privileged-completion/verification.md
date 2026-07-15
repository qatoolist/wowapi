---
id: VER-W03-E05-S001
type: verification-record
parent_story: W03-E05-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W03-E05-S001

## Planned verification procedure

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E05-S001-01 | Run the rejection-boundary test | Local dev or CI, testkit DB | `ratify_by`-declaring definitions are rejected at validation time with a clear error; the reject choice is recorded | ratification test report | W03-E05-S001 implementer |
| AC-W03-E05-S001-02 | Run the audit-present/complete test; run the fault-injection test with an injected audit-write failure | Local dev or CI, testkit DB | Every override produces a complete audit row; an injected audit-write failure rolls back the override, leaving zero effect | audit test report + fault-injection test report | W03-E05-S001 implementer |
| AC-W03-E05-S001-03 | Re-run T1-T3's existing test coverage | Local dev or CI, testkit DB | `TestIntegrationOverrideAuthzGate`, `TestIntegrationOverrideFailsClosedWithoutPermission`, and `TestIntegrationWorkflowOverride` pass | regression confirmation report | W03-E05-S001 implementer |

## Post-execution record

### Actual result

All targeted tests passed against a real testkit database.

### Pass or fail

Pass.

### Evidence identifier

- EV-W03-E05-S001-001 — `TestRatifyByDefinitionRejected`
- EV-W03-E05-S001-002 — `TestOverrideAuditRowPresent`
- EV-W03-E05-S001-003 — `TestOverrideAuditFailureRollsBack`
- EV-W03-E05-S001-004 — T1–T3 regression confirmation (`TestIntegrationOverrideAuthzGate`,
  `TestIntegrationOverrideFailsClosedWithoutPermission`, `TestIntegrationWorkflowOverride`)

### Execution date

2026-07-13.

### Commit or revision

HEAD `733ef3e930cbb3f89f5bbc53d8f562c60e426513` with working-tree modifications for W03-E05-S001.

### Environment

Local dev; DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable;
WOWAPI_REQUIRE_DB=1.

### Reviewer

Independent review pending (W03-E05-S001-T003).

### Findings

None from verification execution.

### Retest status

Not required.

### Final conclusion

Acceptance criteria AC-01 and AC-02 are verified by passing tests. AC-03 regression guard is
verified by passing the existing Override/authz tests.
