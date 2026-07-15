---
id: W03-E01-S002-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W03-E01-S002
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W03-E01-S002 — Artifacts index

Per mandate §9.2.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W03-E01-S002-001 | Capacity-selection enforcement logic | source-code change | implementation | Server-side-validated explicit capacity choice, rejecting no-choice and unentitled-assertion cases | SEC-01 | W03-E01-S002-T001 | `kernel/auth/auth.go` | produced |
| ART-W03-E01-S002-002 | Privileged-session resolver implementation | source-code change | implementation | Grant-table lookup by opaque grant ID; six-condition rejection matrix | SEC-01 | W03-E01-S002-T002 | `kernel/auth/auth.go`; `adapters/auth/pgprincipal/pgprincipal.go` | produced |
| ART-W03-E01-S002-003 | Capacity-selection and resolver documentation | documentation | post-implementation | Documents the capacity-selection mechanism and resolver rejection matrix | SEC-01 | W03-E01-S002-T001, T002 | `kernel/auth/auth.go` doc comments; this story's implementation record | produced |
