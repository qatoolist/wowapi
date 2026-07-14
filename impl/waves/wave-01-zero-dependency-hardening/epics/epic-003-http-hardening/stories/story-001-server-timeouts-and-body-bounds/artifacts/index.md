---
id: W01-E03-S001-ARTIFACTS-INDEX
type: artifact-index
parent_story: W01-E03-S001
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E03-S001 — Artifacts index

Per mandate §9.2. Artifacts are the working-tree diffs at SHA 0a31186cada5c275a588c74081cf977adf346e61 (conductor owns the wave
commit); each is identified by file path rather than a copied bundle, per the wave's
evidence-in-repo convention.

| Artifact ID | Title | Type | Lifecycle stage | Location | Source requirement | Producing task | Status |
|---|---|---|---|---|---|---|---|
| ART-W01-E03-S001-001 | Scaffold-template timeout wiring | source-code (template) | implementation | `internal/cli/templates/init/cmd_api_main.go.tmpl` (http.Server literal) + `configs_base.yaml.tmpl` (http keys) | FBL-09 | T001 | produced |
| ART-W01-E03-S001-002 | HTTP config schema addition | schema | implementation | `kernel/config/config.go` (HTTP struct, Defaults, Validate prod block) | FBL-09 | T001/T002 | produced |
| ART-W01-E03-S001-003 | CSRF defensive-bound | source-code | implementation | `kernel/httpx/csrf.go` (CSRFPolicy.MaxFormBytes + MaxBytesReader wrap) | FBL-09 | T003 | produced |
