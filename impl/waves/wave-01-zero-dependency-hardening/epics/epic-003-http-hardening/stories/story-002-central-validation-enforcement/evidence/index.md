---
id: W01-E03-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W01-E03-S002
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E03-S002 — Evidence index

Per mandate §10. All records pinned to commit SHA 0a31186cada5c275a588c74081cf977adf346e61 (working tree; conductor owns the wave commit).

| Evidence ID | Type | Acceptance criteria proven | Producing task | Status | File |
|---|---|---|---|---|---|
| EV-W01-E03-S002-001 | unit-test report — boot-rejection fail-first pair (3-stage: pre-fix boots / stub red / post-fix green) | AC-01, AC-02 | T001 | produced (failed stage preserved, resolved) | tests/ev-001-boot-rejection-fail-first.md |
| EV-W01-E03-S002-002 | unit-test report — adversarial invalid-DTO 400 with field errors | AC-03 | T002 | produced | tests/ev-002-adversarial-400-field-errors.md |
| EV-W01-E03-S002-003 | unit-test report — waiver-exemption boot success (+ contradiction guard) | AC-04 | T001 | produced | tests/ev-003-waiver-exemption.md |
| EV-W01-E03-S002-004 | unit-test report — crud template migration (generator output) | AC-03 (transitive) | T003 | produced | tests/ev-004-crud-template-migration.md |

Race-detector run: `go test -race -count=1 ./kernel/httpx/ ./kernel/config/ ./app/ ./internal/cli/` — all ok (2026-07-13, same SHA).
Config-flag compat proof: `TestEnforceRouteContractsDefaultsOff` (kernel/config) — flag defaults false everywhere, validates when enabled in every env (embedded in story-001's ev-003 pass log, same command).
