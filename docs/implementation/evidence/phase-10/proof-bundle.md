# Phase 10 — Proof Bundle

Scope (phase-plan row 10, blueprint 10 §2 E21): the installable `wowapi` CLI — scaffolding (init,
new-module), generation (gen crud), migration/seed/openapi/lint/deploy helpers — with embedded
templates and golden tests. No new DB tables. Date: 2026-07-04.

## 1. Decision evidence
D-0057 (Phase 10: CLI command surface + embedded-template scaffolding + generated-Go formatting + CLI
review fixes).

## 2. Discussion evidence
- **A single dispatcher, one file per command**: `cli.go` switches to a `runX(args, stdout, stderr) int`
  per command; each command lives in its own file. This kept the parallel build (lead + agent) conflict-
  free and makes each command independently testable with buffers.
- **Generated Go must be gofmt-clean**: `renderToFile` runs `go/format.Source` on every `.go` output —
  this both formats the result AND fails generation loudly if a template emitted invalid Go (a stronger
  check than the parse-only test the templates originally had).
- **Scaffold safety**: module/resource/field names are `identRE`-validated before any path is built, so
  a template can never be used for path traversal; `--force` gates every overwrite.
- **lint reuses the framework rules**: `wowapi lint boundaries` ports the import-layering + module-
  isolation law from `scripts/lint_boundaries.sh` (a pure `checkBoundaries` unit-tested without a live
  `go list`); the shell script stays the authoritative framework gate for vocabulary/Reveal/test-import.

## 3. Critique/review evidence
`review-findings.md`: 7 reproduced defects (3 med, 4 low) — unknown-field-type unbuildable Go (CLI-01),
silently-accepted null openapi fragment (CLI-02), missing boundary rule classes (CLI-06), plus exit-code
and error-propagation lows. All fixed with regression tests. Path-traversal safety, overwrite protection,
generated-migration RLS, and deploy secret handling verified solid.

## 4. Implementation evidence
Lead: `internal/cli/{migrate,openapi,seed,lint,deploy}_cmd.go` + `cmds_test.go`, dispatcher wiring in
`cli.go`, the `go/format.Source` improvement in `scaffold.go`, all review fixes. Agent: the scaffolding
commands `internal/cli/{init,new_module,gen}_cmd.go` + `scaffold.go` + `templates/` (init/module/crud) +
`scaffold_test.go`. One review agent.

## 5. Verification evidence
`command-log.md`: unit + golden tests for every command (parse-checking generated Go); an end-to-end
binary smoke test (`wowapi init → new-module → gen crud`) producing a gofmt-clean scaffolded tree; the
binary's `lint boundaries` agrees with `scripts/lint_boundaries.sh` on the framework repo. Host `make ci`
and `make ci-container` green.

## 6. Acceptance evidence
`acceptance-map.md`: all 14 Phase 10 exit criteria mapped (AC #6 module-in-an-hour via new-module + gen
crud; #23 CLI scaffolding/generation/migration/seed/openapi/lint; #28 config + deploy render). Carried
forward: gen sqlc/mocks, goreleaser binaries, and the shell script's framework-only lint checks.
Graphify `extract` blocked on LLM key (R11).
