# Phase 10 — Acceptance Map

Phase 10 exit criteria (Goal 2 Phase 10 + phase-plan row 10 + blueprint 10 §2 E21; AC #6/#23/#28) → proof.

| # | Criterion | Proof |
|---|---|---|
| 1 | `wowapi version` (go.mod dependency mismatch warning) | `internal/cli/cli.go` runVersion; existing `cli_test.go` |
| 2 | `wowapi config init/validate/doctor/print/schema` | `internal/cli/config_cmd.go` (Phase 1); config_cmd_test.go |
| 3 | **`wowapi migrate create`** — next-numbered goose migration | `internal/cli/migrate_cmd.go`; `TestMigrateCreateNextNumber`, `TestMigrateCreateEmptyDirStartsAtOne`, `TestMigrateCreateRejectsBadName` |
| 4 | **`wowapi seed validate`** — strict seed-bundle validation (CI gate) | `internal/cli/seed_cmd.go` (uses kernel/seeds.Load); `TestSeedValidateOK`, `TestSeedValidateForeignKeyFails`, `TestSeedValidateRequiresModule` |
| 5 | **`wowapi openapi merge`** — merge fragments, fail on collision | `internal/cli/openapi_cmd.go`; `TestOpenAPIMerge`, `TestOpenAPIMergeDuplicatePathFails` |
| 6 | **`wowapi lint boundaries`** — module isolation + framework layering | `internal/cli/lint_cmd.go` (pure `checkBoundaries` + `go list`); `TestCheckBoundaries*`; binary agrees with `scripts/lint_boundaries.sh` on the framework repo |
| 7 | **`wowapi deploy render`** — compose/env deployment manifest (secrets as refs) | `internal/cli/deploy_cmd.go`; `TestDeployRenderCompose`, `TestDeployRenderEnvAndBadFormat` |
| 8 | **`wowapi init`** — scaffold a compiling product repo (AC #23/#28) | `internal/cli/` scaffold command + embedded templates; golden tests (parse-check generated Go) |
| 9 | **`wowapi new-module`** — scaffold a module package (AC #6) | scaffold command; golden test (parseable module.go, correct Name()) |
| 10 | **`wowapi gen crud`** — CRUD scaffolding for a resource (AC #6) | gen command; golden test (parseable resource.go with permission keys + RLS migration) |
| 11 | Honest help / planned-command roadmap | `wowapi help` lists implemented + remaining commands |
| 12 | Golden / unit tests for every command | `internal/cli/*_test.go` |
| 13 | Container-first verification | host `make ci`; `make ci-container` |
| 14 | Evidence bundle + review | this directory; review-findings.md |

Carried forward: `gen sqlc`/`gen mocks` and goreleaser release binaries (E21 mentions them; the CLI ships
the scaffolding + generator framing, richer generators are incremental); `wowapi config diff` if not
already present. Graphify `extract` blocked on LLM key (R11).
