---
id: W03-E02-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W03-E02-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W03-E02-S001 — Evidence index

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| [EV-W03-E02-S001-001](EV-W03-E02-S001-001-fingerprint-diff.md) | fingerprint-diff test report | W03-E02-S001-T001 | AC-W03-E02-S001-01 | `go test -v ./kernel/config/... -run 'TestSharedFingerprintChangesWithOutboundAllowlist\|TestSharedFingerprintChangesWithTrustedIssuers' -count=1` | 1626b11 (with working-tree changes) | pass | accepted |
| [EV-W03-E02-S001-002](EV-W03-E02-S001-002-egress-report.md) | report-output sample | W03-E02-S001-T002 | AC-W03-E02-S001-02 | `go run /tmp/egress_sample.go` (sample program using `Framework.EgressExceptions()`) | 1626b11 (with working-tree changes) | pass; no credentials in output | accepted |
| [EV-W03-E02-S001-003](EV-W03-E02-S001-003-allowlist-audit.md) | change-audit test report | W03-E02-S001-T003 | AC-W03-E02-S001-03 | `go test -v ./kernel/config/... -run 'TestRecordAllowlistChange' -count=1` | 1626b11 (with working-tree changes) | pass | accepted |
| [EV-W03-E02-S001-004](EV-W03-E02-S001-004-jwks-governance.md) | JWKS-governance negative-fixture test report | W03-E02-S001-T004 | AC-W03-E02-S001-04 | `go test -v ./kernel/auth/... -run 'TestNewJWKSKeySource_Prod' -count=1` | 1626b11 (with working-tree changes) | pass | accepted |
| [EV-W03-E02-S001-005](EV-W03-E02-S001-005-fitness-check.md) | fitness-check test report | W03-E02-S001-T005 | AC-W03-E02-S001-05 | `go test -v ./kernel/config/... -run 'TestFitnessCheck' -count=1` | 1626b11 (with working-tree changes) | pass | accepted |
| EV-W03-E02-S001-006 | review report | W03-E02-S001-T006 | AC-W03-E02-S001-01 through -05 | Independent review checklist per mandate §14 | TBD | TBD | not yet produced |

Evidence status vocabulary: `accepted` means the evidence item was produced,
verified, and supports its acceptance criterion.
