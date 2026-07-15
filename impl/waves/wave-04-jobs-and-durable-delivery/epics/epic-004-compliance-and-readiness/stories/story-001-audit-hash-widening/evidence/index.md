---
id: W04-E04-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W04-E04-S001
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W04-E04-S001 — Evidence index

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W04-E04-S001-001 | tamper-test report (per-field results) | W04-E04-S001-T001 | AC-W04-E04-S001-01 | `go test ./kernel/audit/... -run TestIntegrationAuditChainDetectsPerFieldTampering -count=1 -v` | HEAD | PASS (17/17 fields break verification) | produced |
| EV-W04-E04-S001-002 | version-branch verification report (v1 historical + v2 new-row branches) | W04-E04-S001-T001 | AC-W04-E04-S001-02 | `go test ./kernel/audit/... -run 'TestIntegrationAuditHashVersionBranching|TestIntegrationAuditUnknownHashVersionFailsClosed' -count=1 -v` | HEAD | PASS (v1 + v2 verify; unknown version fails closed) | produced |
| EV-W04-E04-S001-003 | migration-classification report (manifest entry + lock-timeout compliance via W02-E01) | W04-E04-S001-T001 | AC-W04-E04-S001-03 | `go test ./migrations/... -run TestKernelMigrationsHaveManifests -count=1 -v` | HEAD | PASS (manifest valid, online, lock budget 2000 ms) | produced |
