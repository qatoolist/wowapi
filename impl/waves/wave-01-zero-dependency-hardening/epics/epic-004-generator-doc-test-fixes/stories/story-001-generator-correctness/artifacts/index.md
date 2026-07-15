---
id: W01-E04-S001-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W01-E04-S001
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E04-S001 — Artifacts index

Per mandate §9.2. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2": lifecycle
subdirectories (`pre-implementation/`, `implementation/`, `post-implementation/`) are created on first
real content, not pre-populated empty. All entries below are produced.

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Path | Status |
|---|---|---|---|---|---|---|---|---|
| ART-W01-E04-S001-001 | Updated `internal/cli/init_cmd.go` + new `internal/cli/init_version.go` (version-resolution flags) | source-code change | implementation | Adds `--framework-version`/`--local-framework` flags with pre-write verification, shape-classified stamped-version default (tagged release / verified pseudo-version / fail-closed `+dirty` / VCS-derived devel), deletes the `v0.0.0` fallback; go.mod template gains a conditional dev-mode `replace` block; tests in `internal/cli/init_version_test.go` | DX-01 | W01-E04-S001-T001 | `internal/cli/init_cmd.go`, `internal/cli/init_version.go`, `internal/cli/templates/init/go.mod.tmpl`, `internal/cli/init_version_test.go`, `internal/cli/scaffold_test.go` (test-harness accommodation) | produced (uncommitted working-tree change at 05dce5c8; conductor commits) |
| ART-W01-E04-S001-002 | Isolated-temp-dir E2E scaffold harness | test infrastructure | implementation | Reusable generate→build→boot→smoke pipeline (`scaffoldPipeline` + `buildWowapiCLI` released/source builds + `buildFrameworkProxy` hermetic file:// module proxy); shared primitive reused by T004 (via the same underlying scaffold layer) and callable by future DX-04 (W06, out of scope) | DX-01 | W01-E04-S001-T002 | `internal/cli/e2e_scaffold_harness_test.go` | produced (new uncommitted file at 05dce5c8; conductor commits) |
| ART-W01-E04-S001-003 | Updated `internal/cli/templates/crud/resource.go.tmpl` (verb fix) | source-code change (generator template) | implementation | Emits `{{.PermPrefix}}.deactivate` instead of `{{.PermPrefix}}.delete` for the generated DELETE route (line 54; one-token change) | DX-02 | W01-E04-S001-T003 | `internal/cli/templates/crud/resource.go.tmpl` | produced (uncommitted working-tree change at 05dce5c8; conductor commits) |
| ART-W01-E04-S001-004 | Updated `internal/cli/scaffold_test.go` (`TestGenCRUDPermissionKeys` fix) | source-code change (test) | implementation | Corrects the test-locking assertion from `"widgets.widget.delete"` to `"widgets.widget.deactivate"` (one line; test at lines 985-1001 after sibling-story drift, formerly 937-953) | DX-02 | W01-E04-S001-T003 | `internal/cli/scaffold_test.go` | produced (uncommitted working-tree change at 05dce5c8; conductor commits) |
| ART-W01-E04-S001-005 | Generator-output-boots CI test | test infrastructure | implementation | Fail-first CI test proving `gen crud` output boots without closed-verb-set rejection; reuses the existing `buildRenderedProduct` scaffold primitive (`internal/cli/scaffold_test.go:568`) rather than reimplementing it; CI-wired via `make test-unit` (`go test ./...`, no `-short`) | DX-02 | W01-E04-S001-T004 | `internal/cli/gen_crud_boots_test.go` | produced (new uncommitted file at 05dce5c8; conductor commits) |
| ART-W01-E04-S001-006 | Updated `internal/cli/templates/init/configs_base.yaml.tmpl` (i18n block commented) + `TestInitScaffoldConfigValidates` | source-code change (scaffold template) + test infrastructure | implementation | Scope addition (DEV-W01-E04-S001-03): pristine scaffold configs must pass the framework-only `config validate` fallback; the active product-owned `i18n:` block becomes a commented example per the file's own convention | DX-02 (generator correctness charter) | W01-E04-S001-T005 | `internal/cli/templates/init/configs_base.yaml.tmpl`, `internal/cli/gen_crud_boots_test.go` | produced (uncommitted at 05dce5c8; conductor commits) |

## Notes

- ART-W01-E04-S001-002 (the harness) is the story's single most important artifact for downstream reuse
  — its "Description" field explicitly records the shared-primitive relationship so a future reader of
  this index (including a future DX-04 story's own planning) does not need to rediscover that fact from
  `story.md`/`plan.md` prose.
- ART-W01-E04-S001-002 (the full T5 generate→build→boot→smoke harness incl. the released-CLI vs
  source-built-CLI distinction) remains T002's deliverable and is NOT claimed by T004's work. T004
  deliberately reused the already-existing `buildRenderedProduct` scaffold primitive
  (`internal/cli/scaffold_test.go:568`: init → replace-to-local-checkout → tidy) as its scaffold layer,
  honoring the story's "implemented once and reused, not reimplemented" instruction; T002 is expected to
  build on the same primitive.
- No pre-implementation artifact is expected for this story (no baseline capture beyond the current-state
  citations already recorded in `story.md` "Current-state assessment," which are themselves confirmed by
  direct source-code inspection at story-authoring time, not by a separate artifact).
- No post-implementation artifact (release notes, runbook, upgrade guidance) is expected — this story's
  scope is CLI/generator-internal correctness, not a user-facing release event in its own right; any
  release-notes entry for DX-01/DX-02 landing is a concern of whatever wave/story governs the next actual
  release cut, not this story.
