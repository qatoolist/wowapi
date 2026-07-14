---
id: W05-E04-RISKS
type: epic-risks
epic: W05-E04
wave: W05
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W05-E04 — Risks

Epic-scoped elaboration of the wave-level risk register (`../../risks.md`). RISK-W05-005 originates
at wave scope and lands entirely within this epic's S002.

| Risk ID | Description | Likelihood | Impact | Severity | Affected items | Mitigation | Contingency | Owner | Status | Residual risk |
|---|---|---|---|---|---|---|---|---|---|---|
| RISK-W05-005 | SEC-04 T4 (per-tenant/global authorization epoch for cross-pod revocation) was PLAN's own "Highest-risk task" with "an open architecture decision (LISTEN/NOTIFY vs. epoch-row-poll)." D-06 resolves the architecture decision, but implementation risk remains in correctly wiring the epoch bump into every framework-side mutation path — MATRIX CS-17's own cross-CS note: "T2's epoch bumps must be added to the framework mutation paths that exist today (role/permission assignment writes in `kernel/authz`, seeds) and extended to SEC-01's grant table when it lands" | Medium | Medium — a missed epoch bump on one mutation path reintroduces the exact stale-cache-serves-revoked-access defect SEC-04 exists to close, narrowed to that one path | Medium | W05-E04-S002 | D-06 is referenced in `decisions/index.md`, closing the architecture-decision risk; S002's task breakdown enumerates every known framework-side mutation path explicitly (role/permission assignment writes in `kernel/authz`, seeds, and W03's grant-table writes, since W05 enters after W03-E01 acceptance) rather than treating epoch-bump wiring as a single undifferentiated task; S002 adds an independent-review task specifically scoped to confirming mutation-path completeness | If a mutation path is found missing an epoch bump after review, treat as a follow-up task recorded in `deviations.md`, not a silent gap | unassigned | open | Low once the enumerated mutation-path list is confirmed complete |

## Residual risk after mitigation

RISK-W05-005 is expected to reduce to low residual risk once the enumerated mutation-path list is
confirmed complete at implementation time and independently re-checked by S002's own review task.
