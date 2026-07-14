---
id: ART-W01-E04-S002-002
type: artifact
title: DX-05 T3 — blueprint-11 CLI example reconciliation decision table
parent_story: W01-E04-S002
producing_task: W01-E04-S002-T002
source_requirement: DX-05 (T3)
status: produced
created_at: 2026-07-13
---

# DX-05 T3 — blueprint-11 CLI example decision table

Reconciles every CLI example in `docs/blueprint/11-framework-distribution-and-consumption.md`
against `internal/cli/cli.go` and its command files at revision
`0a31186cada5c275a588c74081cf977adf346e61` — re-verified unchanged at the advanced HEAD
`05dce5c8` (impl/-only delta; see `../evidence/index.md`).
Decision vocabulary per `../tasks/task-002-dx05-residual-reconciliation.md`: **implement** = example
intent valid, doc corrected to match the real CLI; **delete** = example describes surface that does
not exist and is not landing in this wave; **keep** = already matches reality (no edit).

Every corrected form was executed against a CLI binary built at this HEAD (see
`../evidence/ev-002-command-log.md`); every stale form was executed and confirmed to fail (or, for
`gen crud --module requests`, to succeed misleadingly by writing into the wrong directory).

| # | Blueprint location | Example (as previously documented) | Reality at HEAD (`internal/cli/`) | Decision | Corrected form |
|---|---|---|---|---|---|
| 1 | §3 step 1, §5 block | `go install .../cmd/wowapi@vX.Y.Z` | Not a wowapi command; matches exact-pin policy (DX-05 T1) | keep | — |
| 2 | §3 step 2, §5 block | `wowapi init --module example.com/acme-ops --wowapi-version vX.Y.Z` | `init_cmd.go:50-57`: flags are `--dir/--module/--name/--force`; no `--wowapi-version` (exit 2) | implement | `wowapi init --module example.com/acme-ops` (version-pin flag is DX-01 scope, W01-E04-S001 plan — not landed in that story's current W01 slice; see deviations) |
| 3 | §3 step 3, §5 block | `wowapi new-module requests` | `new_module_cmd.go:35,42-45`: `--name` is required; bare positional exits 2 | implement | `wowapi new-module --name requests` |
| 4 | §5 block | `wowapi gen crud --module requests --resource request` | `gen_cmd.go:19,66`: `--module` is a module *directory*; passing `requests` writes `./requests/` at cwd — wrong location, silently misleading | implement | `wowapi gen crud --module internal/modules/requests --resource request` |
| 5 | §5 block | `wowapi gen` (`# sqlc + mocks + mappers (idempotent)`) | `gen_cmd.go:46-58`: bare `gen` exits 2; only subcommand is `crud`; no sqlc/mocks/mappers generation exists anywhere in `internal/cli` | delete | removed from block |
| 6 | §5 block | `wowapi migrate create --module requests --name create_requests` | `migrate_cmd.go:52-55`: flags are `--dir` (default `migrations`) and `--name`; no `--module` (exit 2) | implement | `wowapi migrate create --dir internal/modules/requests/migrations --name create_requests` |
| 7 | §5 block | `wowapi seed validate` (bare) | `seed_cmd.go:106-109`: `--module` is required (exit 2); `--dir` default `seeds` wrong for the module layout the same doc scaffolds | implement | `wowapi seed validate --module requests --dir internal/modules/requests/seeds` |
| 8 | §5 block | `wowapi openapi merge` | `openapi_cmd.go:42-56`: valid (flags `--dir/--title/--version/--out`) | keep | — |
| 9 | §5 block | `wowapi lint boundaries` | `lint_cmd.go:93-96`: valid (`--pkgs` optional, default `./...`) | keep | — |
| 10 | §5 block | `wowapi version` | `cli.go:67-90`: valid; prints CLI version + go.mod comparison, matching the §5 "Version alignment" prose | keep | — |
| 11 | §5 block | `wowapi config init` | `config_cmd.go:140-166`: subcommands are `validate/print/schema/doctor/diff/capacity`; no `init` (exit 2). Config scaffolding is done by `wowapi init` | delete | removed from block |
| 12 | §5 block | `wowapi config validate --env prod` | `config_cmd.go:64-67`: `--env` exists | keep | — |
| 13 | §5 block | `wowapi config doctor` | dispatched, valid | keep | — |
| 14 | §5 block | `wowapi config print --redacted` | `config_cmd.go:215`: `--redacted` exists (and is required) | keep | — |
| 15 | §5 block | `wowapi config diff --from dev --to prod` | `config_delegate.go:52-63`: `--from/--to` exist and are required | keep | — |
| 16 | §5 block | `wowapi config schema` | dispatched, valid (exit 0 with no files on disk) | keep | — |
| 17 | §5 block | `wowapi deploy render --env prod` | `deploy_cmd.go:48-54`: `--format/--name/--image/--env/--out`; valid | keep | — |
| 18 | §5 CI-usage bullet | `wowapi openapi merge --check` | no `--check` flag exists (exit 2); merge already fails loudly on duplicate paths/schemas, which is the CI value | implement | `wowapi openapi merge` |
| 19 | §3 step 6 (prose claim) | "`wowapi init` seeds `configs/{base,local,dev,stage,prod}.yaml`" | scaffold at HEAD emits only `configs/{base,local}.yaml` (verified by running `init`) | implement | claim corrected to `{base,local}.yaml` + "add further overlays per environment" |
| 20 | §5 fallback bullet | `go run .../cmd/wowapi@vX.Y.Z <cmd>` | generic Go invocation, valid | keep | — |

## Observations recorded, deliberately NOT edited (out of this task's scope)

- `internal/cli/cli.go:101` usage text lists `config validate|print|schema|doctor` but omits the
  real `diff`/`capacity` subcommands, and blueprint-11 omits the real `i18n`, `dlq`, `apikey`,
  `audit`, `seed sync`, `lint lifecycle` commands entirely. T3's contract is examples→reality
  (implement-or-delete per existing example), not doc completeness; adding new examples or editing
  Go source is out of scope. Recorded here so it is not silently dropped.
- The scaffolded product's `configs/{base,local}.yaml` fail `wowapi config validate` at HEAD with
  `i18n.default_locale: unknown key` etc., and the scaffolded `go.mod` carries an invalid `+dirty`
  pseudo-version (SF-7/DX-01). Both are code-level generator defects owned by W01-E04-S001/DX-01
  scope, routed to that story's owner via IRC (see `../deviations.md` DEV-03).
