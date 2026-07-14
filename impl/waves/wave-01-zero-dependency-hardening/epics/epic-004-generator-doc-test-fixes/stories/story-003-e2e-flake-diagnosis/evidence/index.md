---
id: W01-E04-S003-EVIDENCE-INDEX
type: evidence-index
parent_story: W01-E04-S003
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E04-S003 — Evidence index

Per mandate §10. All evidence produced 2026-07-13 at commit
`0a31186cada5c275a588c74081cf977adf346e61`, at path `evidence/premier/T-TEST-01/`.

| Evidence ID | Type | Task | Acceptance criteria proven | Result | Status |
|---|---|---|---|---|---|
| EV-W01-E04-S003-001 | Test execution logs (repeated `-count`+parallel, race, stress runs) — record: `premier/T-TEST-01/reproduction-runs.md`, logs: `premier/T-TEST-01/logs/` (16 files) | W01-E04-S003-T001 | AC-W01-E04-S003-01 | Documented non-reproduction: 29/29 clean executions at pinned SHA; 16 contaminated failures (run-01..04) preserved with `failed` status, superseded by run-05..08 (`retested`) | final |
| EV-W01-E04-S003-002 | Diagnosis/decision note — `premier/T-TEST-01/diagnosis-note.md` | W01-E04-S003-T001 | AC-W01-E04-S003-01 | DB-wiring determined (own wiring via raw `DATABASE_URL`; does NOT use `testkit.NewDB`); decision for T002: monitoring-only, no code fix | final |
| EV-W01-E04-S003-003 | Diagnosis-note update (monitoring-only branch) — `diagnosis-note.md` §5-§6 + task-002 Implementation Record | W01-E04-S003-T002 | AC-W01-E04-S003-02 | Monitoring-only outcome implemented, traceable to T001 findings (task-002 branch 3) | final |

Per mandate §10: "Failed evidence must be preserved and marked appropriately... Do not delete earlier
failed verification merely because a later run passes." Any run in the T001 reproduction sequence that
fails is preserved in full, not discarded even if a subsequent run of the same protocol passes.

Addendum (conductor, 2026-07-13): internal/e2e/e2e_test.go modified post-evidence (--local-framework flag added to integrate DX-01 fail-closed init); fresh runs TestE2EScaffoldedRepoBuild PASS (10.5s, 13.3s) + reviewer independent re-run PASS (11.4s). Diagnosis conclusions unaffected (change is harness wiring, not timing/DB behavior).
