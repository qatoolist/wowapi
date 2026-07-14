---
id: W03-E02-S001-TASKS-INDEX
type: tasks-index
parent_story: W03-E02-S001
status: accepted
created_at: 2026-07-13
updated_at: 2026-07-13
---

# W03-E02-S001 — Tasks index

| Task | Title | Owner | Status | Dependencies | Output | Related AC | Implementation state | Verification state |
|---|---|---|---|---|---|---|---|---|
| [W03-E02-S001-T001](task-001-fingerprint-scope-confirmation.md) | Fingerprint-scope confirmation (SEC-06 T1) | unassigned | done | none | A fingerprint-diff regression test proving `SharedFingerprint()`'s scope covers the outbound allowlist | AC-W03-E02-S001-01 | done | pass |
| [W03-E02-S001-T002](task-002-boot-time-egress-report.md) | Boot-time egress-exception report (SEC-06 T2) | unassigned | done | none | A boot-time report enumerating every enabled egress exception, confirmed credential-free | AC-W03-E02-S001-02 | done | pass |
| [W03-E02-S001-T003](task-003-allowlist-change-audit.md) | Allowlist change-audit trail (SEC-06 T3) | unassigned | done | W03-E02-S001-T001 | An allowlist configuration change produces an audit-visible record, proven by a test | AC-W03-E02-S001-03 | done | pass |
| [W03-E02-S001-T004](task-004-jwks-client-governance-gate.md) | JWKS-client governance gate, D-07 enactment (SEC-06 T4) | unassigned | done | W03-E02-S001-T001, W03-E02-S001-T002, W03-E02-S001-T003 | A `prod`-profile boot with a custom JWKS client and no declared trusted-issuer allowlist fails readiness | AC-W03-E02-S001-04 | done | pass |
| [W03-E02-S001-T005](task-005-fitness-check.md) | No-tenant-controlled-allowlist fitness check (SEC-06 T5) | unassigned | done | W03-E02-S001-T001, W03-E02-S001-T002, W03-E02-S001-T003, W03-E02-S001-T004 | A static fitness check that fails if allowlist/JWKS-client construction reads request/tenant-scoped data | AC-W03-E02-S001-05 | done | pass |
| [W03-E02-S001-T006](task-006-independent-review.md) | Independent review | unassigned | todo | W03-E02-S001-T001, W03-E02-S001-T002, W03-E02-S001-T003, W03-E02-S001-T004, W03-E02-S001-T005 | A completed review report confirming the checklist above, recorded as evidence | AC-W03-E02-S001-01, AC-W03-E02-S001-02, AC-W03-E02-S001-03, AC-W03-E02-S001-04, AC-W03-E02-S001-05 | not started | not started |

## Grouping rationale

Per `plan.md`: tasks follow the PLAN task breakdown for this story. T006 is the
mandated independent-review gate.
