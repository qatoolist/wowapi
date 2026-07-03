# Phase 10 — Review Findings

One focused code-review agent audited the `internal/cli` package on 2026-07-04 (proportionate for a
single-package, no-DB tooling phase). It reproduced 7 defects — 3 medium, 4 low — all fixed with
regression tests. Path-traversal safety, overwrite protection, generated-migration RLS, openapi
determinism, and deploy secret-ref handling were verified solid.

| ID | Sev | Finding (reproduced) | Resolution | Status |
|---|---|---|---|---|
| CLI-01 | med | `gen crud` with an unknown `--fields` type (e.g. `price:decimal`) passed the raw string through as the Go type → syntactically valid but undefined → `go build` fails downstream | `mapFieldType` returns `ok bool`; an unknown type errors (exit 1) before any file is written; `TestGenCRUDRejectsUnknownFieldType` | **fixed** |
| CLI-02 | med | `openapi merge` silently accepted a `null` (or non-object) fragment — it contributed nothing but exited 0 | `mergeFragment` rejects any fragment whose first token is not `{`; `TestOpenAPIMergeRejectsNullFragment` | **fixed** |
| CLI-06 | med | `wowapi lint boundaries` `checkBoundaries` was missing rule classes vs `scripts/lint_boundaries.sh` (adapters/cmd/internal/cli/internal/tools layering + the hard "no production package imports testkit" rule) | added the missing framework layer rules + the hard testkit rule (testkit removed from the per-layer lists to avoid double-reporting); still OK on the framework repo, matching the shell gate | **fixed** |
| CLI-03 | low | missing required flag exited 1 in `migrate create` / `seed validate` (should be 2, matching init/new-module/gen) | split empty-vs-invalid: missing `--name`/`--module` → exit 2; malformed → exit 1; tests updated + `TestMigrateCreateMissingNameIsUsageError` | **fixed** |
| CLI-04 | low | `listImports` discarded `go list` stderr, so a compile error / import cycle gave `exit status 1` with no detail | captures `cmd.Stderr` and appends it to the wrapped error | **fixed** |
| CLI-05 | low | stdout write errors ignored in `openapi merge` + `deploy render` → exit 0 on a broken pipe | both propagate the write error (exit 1) | **fixed** |
| CLI-07 | low | `gen crud` derived the package name via `filepath.Base(--module)` without validating it — a non-identifier segment produced a confusing "not valid Go" error | validates the derived package name against `identRE` with a clear message | **fixed** |

Reviewer-verified solid (positive): `--name`/`--resource`/`--fields`-name are `identRE`-validated BEFORE
any path is built (no `..`/`/` traversal); `renderToFile` refuses to clobber without `--force`;
generated `.go` is gofmt-clean (`go/format.Source`, which also catches template bugs at generation
time); generated migrations carry ENABLE+FORCE RLS + tenant policy + app_rt grants (no DELETE);
`nextMigrationNumber` handles empty dir/gaps/non-migration files + 5-digit padding; `openapi merge`
output is deterministic (sorted keys) and faithfully round-trips fragment JSON with duplicate
detection; `deploy render` emits secrets as `${VAR}` refs (no inlining, no template injection);
dispatcher reaches all commands with consistent exit codes.

Residual / carried forward (honest): `checkBoundaries` covers the import-layering + module-isolation
law; the shell `scripts/lint_boundaries.sh` remains the authoritative FRAMEWORK gate for its
vocabulary-denylist, `Secret.Reveal()`, and test-import checks (not replicated in the CLI, which
targets product repos where those framework-specific rules do not apply). `gen sqlc`/`gen mocks` and
goreleaser release binaries remain incremental follow-ups (the CLI ships the scaffolding + crud
generator). Graphify `extract` blocked on LLM key (R11).
