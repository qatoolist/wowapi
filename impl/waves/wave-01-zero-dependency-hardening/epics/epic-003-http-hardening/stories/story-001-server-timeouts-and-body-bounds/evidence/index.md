---
id: W01-E03-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W01-E03-S001
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E03-S001 — Evidence index

Per mandate §10. All records pinned to commit SHA 0a31186cada5c275a588c74081cf977adf346e61 (working tree; conductor owns the wave commit).

| Evidence ID | Type | Acceptance criteria proven | Producing task | Status | File |
|---|---|---|---|---|---|
| EV-W01-E03-S001-001 | unit-test report — template-render assertion (fail-first pair) | AC-W01-E03-S001-02 | T001 | produced (failed half preserved, resolved) | tests/ev-001-template-render-fail-first.md |
| EV-W01-E03-S001-002 | unit-test report — config defaults assertion | AC-W01-E03-S001-01 | T001 | produced | tests/ev-002-config-defaults.md |
| EV-W01-E03-S001-003 | unit-test report — prod-profile zero-timeout rejection (fail-first pair) | AC-W01-E03-S001-03 | T002 | produced (failed half preserved, resolved) | tests/ev-003-prod-zero-rejection-fail-first.md |
| EV-W01-E03-S001-004 | static-analysis report — scoped gosec run | AC-W01-E03-S001-04 | T003 | produced (see honesty note re rule id) | static-analysis/ev-004-gosec-scoped.md |
| EV-W01-E03-S001-005 | unit-test report — CSRF oversized-form-body rejection (fail-first pair) | AC-W01-E03-S001-04 | T003 | produced (failed half preserved, resolved) | tests/ev-005-csrf-oversized-body-fail-first.md |

Race-detector run: `go test -race -count=1 ./kernel/httpx/ ./kernel/config/ ./app/ ./internal/cli/` — all ok (2026-07-13, same SHA).
