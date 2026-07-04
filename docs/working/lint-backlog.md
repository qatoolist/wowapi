# golangci-lint backlog

golangci-lint was only wired into the enforced gate on 2026-07-04. Before that a `.golangci.yml` existed
but nothing ran the *full* set in CI, so a backlog of pre-existing issues accumulated. The gate now blocks
on **changed code only** (`make lint-new`), so this backlog does not block work — but it is tracked here
and burned down incrementally. See [quality-gates.md](quality-gates.md).

## Snapshot (2026-07-04, after config fixes + auto-fixes)

`make lint` (full `golangci-lint run ./...`) reports **~160** issues:

| Linter | Count | Notes |
|---|---:|---|
| `errcheck` | 154 | Unchecked error returns in production code. The bulk of the backlog. Many are `defer x.Close()` / best-effort writes; each needs a real check or an explicit `_ =` with justification — never a blanket ignore. |
| `unused` | 3 | Dead code (`kernel/workflow/registry.go` `latestVersion`; two test-only symbols). Remove once confirmed no future use. |
| `unparam` | 2 | Params/results with a constant value (`runtime.go` `createTask` `def`; `testkit/db.go` `connectDB` `dbname`). |
| `unconvert` | 1 | Unnecessary conversion in `internal/tools/benchbudget`. |

Already resolved in the wiring pass: 31 `depguard` false positives (test files importing testkit — the
rule now excludes `*_test.go`), 10 `gofumpt` formatting, `misspell`, and most `staticcheck` quick-fixes.

## Burn-down plan

1. **errcheck** — tackle package-by-package. For each: check the error, or `_ =` with a one-line reason
   for genuinely best-effort calls (e.g. `defer rows.Close()`), or add a scoped `errcheck.exclude-functions`
   entry in `.golangci.yml` for a proven-safe stdlib call. Do NOT blanket-`//nolint`.
2. **unused / unparam / unconvert** — cheap; fix opportunistically when touching the file.
3. As categories reach zero, promote them from advisory to enforced on the *whole* tree (drop the
   `--new-from-merge-base` scoping for that linter, or fail `make lint`).

## Rule

Do not let the backlog grow: `make lint-new` (the gate) ensures new/changed code is fully clean, so this
number should only ever go **down**.
