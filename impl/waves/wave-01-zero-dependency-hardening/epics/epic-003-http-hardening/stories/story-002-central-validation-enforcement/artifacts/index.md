---
id: W01-E03-S002-ARTIFACTS-INDEX
type: artifact-index
parent_story: W01-E03-S002
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E03-S002 — Artifacts index

Per mandate §9.2. Artifacts are the working-tree diffs at SHA 0a31186cada5c275a588c74081cf977adf346e61 (conductor owns the wave
commit); each is identified by file path.

| Artifact ID | Title | Type | Lifecycle stage | Location | Source requirement | Producing task | Status |
|---|---|---|---|---|---|---|---|
| ART-W01-E03-S002-001 | RouteMeta.Request contract seam + boot check + waiver | source-code | implementation | `kernel/httpx/router.go` (RouteMeta.Request/NoRequestBody, RequireRequestContracts, checkRequestContract) | FBL-08 | T001 | produced |
| ART-W01-E03-S002-002 | ValidatedHandler adaptor | source-code | implementation | `kernel/httpx/decode.go` (ValidatedHandler[T]) | FBL-08 | T002 | produced |
| ART-W01-E03-S002-003 | Enforcement profile flag + boot wiring | source-code / schema | implementation | `kernel/config/security.go` (EnforceRouteContracts) + `app/boot.go` (RequireRequestContracts wiring) | FBL-08 | T001 | produced |
| ART-W01-E03-S002-004 | Migrated crud template | source-code (template) | implementation | `internal/cli/templates/crud/resource.go.tmpl` | FBL-08 | T003 | produced |
