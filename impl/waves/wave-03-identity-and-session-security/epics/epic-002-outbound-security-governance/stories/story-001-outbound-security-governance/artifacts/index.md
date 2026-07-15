---
id: W03-E02-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W03-E02-S001
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W03-E02-S001 — Artifacts index

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W03-E02-S001-001 | `SharedFingerprint()` scope confirmation/extension | source-code change + test | implementation | Fingerprint-diff regression test proving allowlist coverage | SEC-06 | W03-E02-S001-T001 | `kernel/config/shared.go`, `kernel/config/shared_test.go` | produced |
| ART-W03-E02-S001-002 | Boot-time egress-exception report | source-code change | implementation | Readiness/log report enumerating enabled egress exceptions | SEC-06 | W03-E02-S001-T002 | `kernel/config/egress.go`, `kernel/config/egress_report_test.go`, `kernel/kernel.go` | produced |
| ART-W03-E02-S001-003 | Allowlist change-audit trail | source-code change | implementation | Audit-visible record on allowlist configuration change | SEC-06 | W03-E02-S001-T003 | `kernel/config/egress.go`, `kernel/config/allowlist_audit_test.go`, `kernel/kernel.go` | produced |
| ART-W03-E02-S001-004 | JWKS trusted-issuer config-gate implementation | source-code change | implementation | D-07 enactment: declared, fingerprinted trusted-issuer config field; `prod`-profile fail-closed readiness gate | SEC-06, D-07 | W03-E02-S001-T004 | `kernel/config/security.go`, `kernel/auth/jwks.go`, `kernel/auth/jwks_governance_test.go`, `internal/cli/templates/init/cmd_api_main.go.tmpl` | produced |
| ART-W03-E02-S001-005 | No-tenant-controlled-allowlist fitness check | test / static-analysis rule | implementation | Static assertion allowlist/JWKS-client construction never reads request/tenant-scoped data | SEC-06 | W03-E02-S001-T005 | `kernel/config/egress_fitness_test.go` | produced |
