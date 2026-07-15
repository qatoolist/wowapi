---
id: W03-E01-S003-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W03-E01-S003
status: accepted
derived: false
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W03-E01-S003 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories are created on first real content, not pre-populated empty. All entries
below are `not yet produced`.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W03-E01-S003-001 | Assurance-freshness binding/enforcement code change | source-code change | implementation | Binds `auth_time`/`acr`/`amr`; enforces step-up freshness | SEC-01 | W03-E01-S003-T001 | `kernel/auth/auth.go`, `kernel/authz/evaluator.go`, `kernel/authz/registry.go`, `testkit/auth.go` | produced |
| ART-W03-E01-S003-002 | Credential-scheme distinction implementation | source-code change | implementation | Explicit user/API-key/webhook/internal distinction at the permission-check layer | SEC-01 | W03-E01-S003-T002 | `kernel/authz/authz.go`, `kernel/authz/evaluator.go`, `kernel/auth/auth.go`, `kernel/apikey/apikey.go` | produced |
