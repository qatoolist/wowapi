---
id: W01-E01-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W01-E01-S001
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E01-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
lifecycle subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are
created on first real content, not pre-populated empty. All entries produced 2026-07-13 (W01Lint).

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W01-E01-S001-001 | Updated `.golangci.yml` (zero-cost block) | configuration example | implementation | Enables sqlclosecheck, rowserrcheck, bodyclose, wastedassign, makezero, musttag, testifylint | FBL-05 | W01-E01-S001-T001 | `.golangci.yml` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S001-002 | `internal/cli/config_delegate.go` (noctx fix) | source-code change | implementation | `exec.Command` → `exec.CommandContext` | FBL-05 | W01-E01-S001-T002 | `internal/cli/config_delegate.go` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S001-003 | `internal/cli/lint_cmd.go` (noctx fix) | source-code change | implementation | `exec.Command` → `exec.CommandContext` | FBL-05 | W01-E01-S001-T002 | `internal/cli/lint_cmd.go` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S001-004 | `app/maintenance.go` (copyloopvar fix) | source-code change | implementation | Removes pre-1.22 loop-variable-capture idiom | FBL-05 | W01-E01-S001-T003 | `app/maintenance.go` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S001-005 | `kernel/config/config.go` (pool-lifetime keys) | source-code change / configuration schema | implementation | Adds `MaxConnLifetime`/`MaxConnIdleTime` config keys, validation, defaults | FBL-05 | W01-E01-S001-T004 | `kernel/config/config.go` | produced 2026-07-13 (working diff on HEAD 0a31186; conductor commits) |
| ART-W01-E01-S001-006 | Pool-lifetime config documentation update | documentation | post-implementation | Documents new config keys alongside existing `MaxConns` docs | FBL-05 | W01-E01-S001-T004 | `docs/user-guide/configuration.md` | produced 2026-07-13 |
