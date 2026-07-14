---
id: W03-EPICS-INDEX
type: epics-index
wave: W03
status: planned
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03 — Epics index

| Epic | Title | Status | Stories | Objective |
|---|---|---|---|---|
| [W03-E01](epic-001-server-side-session-state/epic.md) | server-side-session-state | planned | 4 | Move tenant membership, break-glass, and impersonation state server-side, keyed on a new `identity_grant` table, replacing direct trust of client-presented JWT claims |
| [W03-E02](epic-002-outbound-security-governance/epic.md) | outbound-security-governance | planned | 1 | Govern the explicit outbound-security escape hatches (JWKS client injection, allowlisted egress hosts) with fingerprinting, boot-time reporting, change-audit, and a declared trusted-issuer config gate |
| [W03-E03](epic-003-webhook-authenticated-replay/epic.md) | webhook-authenticated-replay | planned | 1 | Bind webhook replay-window and dedup decisions exclusively to provider-authenticated data via a new `Verifier` interface |
| [W03-E04](epic-004-relationship-semantics/epic.md) | relationship-semantics | planned | 1 | Complete relationship-checker evaluation for party-subject edges and every schema-enumerated subject kind, with actor attribution shared from DATA-06 |
| [W03-E05](epic-005-workflow-privileged-completion/epic.md) | workflow-privileged-completion | planned | 1 | Close SEC-02's remaining P0 work: ratification implementation (or documented interim posture) and durable, grant-ID-attributed override audit |
