# EV-W01-E04-S002-002 — command log: blueprint-11 CLI example verification

Executed 2026-07-13T07:25Z against a CLI binary built at HEAD
`0a31186cada5c275a588c74081cf977adf346e61` (`go build -o /tmp/w01docs-wowapi ./cmd/wowapi`,
go1.26.5 darwin/arm64). Fail-first discipline: every stale documented form was executed and its
failure captured BEFORE the doc correction was finalized; every corrected form was then executed
in a freshly scaffolded product repo (`wowapi init --module example.com/acme-ops` in a temp dir).

## Stale forms as previously documented (expected: fail or misbehave)

| Command | Exit | Symptom |
|---|---|---|
| `wowapi init --module example.com/acme-ops --wowapi-version v1.0.0` | 2 | flag provided but not defined |
| `wowapi new-module requests` | 2 | `--name is required` |
| `wowapi gen` | 2 | usage (no bare-gen; only `crud` subcommand exists) |
| `wowapi gen crud --module requests --resource request` | **0 (misleading)** | wrote `./requests/request.go` + migration at cwd — wrong location; `--module` is a directory |
| `wowapi migrate create --module requests --name create_requests` | 2 | flag provided but not defined (`--dir`/`--name` only) |
| `wowapi seed validate` | 2 | `--module is required` |
| `wowapi config init` | 2 | unknown subcommand (validate/print/schema/doctor/diff/capacity) |
| `wowapi openapi merge --check` | 2 | flag provided but not defined |

## Corrected forms as now documented (expected: parse + run)

| Command | Exit |
|---|---|
| `wowapi init --module example.com/acme-ops` | 0 |
| `wowapi new-module --name requests` | 0 |
| `wowapi gen crud --module internal/modules/requests --resource request` | 0 |
| `wowapi migrate create --dir internal/modules/requests/migrations --name create_requests` | 0 |
| `wowapi seed validate --module requests --dir internal/modules/requests/seeds` | 0 |
| `wowapi openapi merge --dir internal/modules/requests` | 0 |
| `wowapi version` | 0 |
| `wowapi config schema` | 0 |
| `wowapi deploy render --env prod` | 0 |
| `wowapi lint boundaries` | parse OK; run fails in scaffolded repo — scaffolded go.mod invalid pseudo-version (known DX-01/SF-7 defect, not a doc-example defect) |
| `wowapi config validate --env prod` / `doctor` / `print --redacted` / `diff --from dev --to prod` | parse OK (all flags exist); runs fail in scaffolded repo on (a) the same DX-01 go.mod defect via product-checker delegation and (b) scaffolded configs carrying `i18n.*` keys unknown to the framework schema — code-level generator defects routed to W01-E04-S001's owner (deviations.md DEV-03), not doc-example defects |

## Scaffold observation backing decision-table row 19

`wowapi init` at HEAD emits `configs/{base,local}.yaml` only (verified by listing the scaffolded
`configs/` dir) — blueprint step-6 claim of `{base,local,dev,stage,prod}.yaml` corrected.

## HEAD-clean retest (contamination check)

The first run's binary was built from the shared working tree, which carries sibling workers'
in-flight W01 edits (flagged by the S001 owner). Retested 2026-07-13T07:35Z with a binary built
from a pristine `git archive HEAD` extraction — HEAD had by then advanced to
`05dce5c8a548f7dce3222637ab2c82024236a2a0` (impl/-only delta vs `0a31186`, no CLI/docs change;
see the carry-forward note in `../index.md`): all corrected forms and all stale-form failures
reproduce identically; `wowapi init` still emits only `configs/{base,local}.yaml`; the
scaffolded-config `i18n.default_locale`/`i18n.go_bundles`/`i18n.locales_dir` unknown-key
validation failure reproduces with the clean binary — confirmed a genuine committed-tree defect,
not working-tree contamination. Doc decisions unaffected.
