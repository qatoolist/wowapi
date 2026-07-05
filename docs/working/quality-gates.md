# Quality Gates — format, vet, lint, tidy, test

The formalized pre-commit / pre-push / CI culture for this repo. It follows standard Go practice and is
wired through the `Makefile` + version-controlled git hooks (`.githooks/`) + the CI workflow, so the same
checks run on every machine and in CI.

## One-time setup

```bash
make setup          # installs golangci-lint, installs the git hooks, downloads modules
# or just the hooks:
make hooks          # git config core.hooksPath .githooks
```

`golangci-lint` **v2** bundles the formatters (**gofumpt** + **goimports**), so there are no separate
formatter binaries to install.

## The targets

| Target | What it does | Where it runs |
|---|---|---|
| `make fmt` | Apply **gofumpt + goimports** to the tree | manual |
| `make fmt-check` | Fail if any file needs formatting | pre-commit, CI |
| `make vet` | `go vet ./...` | pre-push, CI |
| `make lint` | **Full** golangci-lint across the tree — now **clean** (backlog B-1 closed 2026-07-05, D-0087); still `manual` until golangci-lint is pinned and it is promoted to CI | manual |
| `make lint-new` | golangci-lint on **changed code only** (`--new-from-merge-base`) — **the enforced gate** | pre-commit, pre-push, CI |
| `make tidy` / `tidy-check` | `go mod tidy` / fail if `go.mod`/`go.sum` drift | pre-push, CI |
| `make check` | `fmt-check vet lint-new tidy-check test-unit` — fast local pre-flight | manual |
| `make coverage` | Statement coverage against the **real DB** (`WOWAPI_REQUIRE_DB=1`), over the measured package set | manual |
| `make coverage-check` | Fail if total coverage drops below `COVERAGE_FLOOR` (**90%**) | CI coverage job |
| `make ci` / `ci-container` | Authoritative correctness gate (vet, boundaries, unit, race, bench, build); `ci-container` runs it with a real DB (`WOWAPI_REQUIRE_DB=1`) | CI gate job |

## Coverage floor

`make coverage-check` enforces a **90%** statement-coverage floor (`COVERAGE_FLOOR`, override to raise it).
Coverage is measured **with a real Postgres** — much of the kernel (RLS, outbox, jobs, audit) only executes
against the DB, so a no-DB run understates it. The measured set **excludes** packages that cannot be
meaningfully unit-tested: `cmd/wowapi` (process main), `internal/tools/migrate` (tool main),
`internal/testmodules` (test fixture), and `module` (interface-only) — see `COVER_EXCLUDE` in the `Makefile`.
The CI `coverage` job runs `make coverage-check` inside the toolbox container so DB-backed tests contribute.

## When each check runs

- **On commit** (`.githooks/pre-commit`, fast): `fmt-check` + `lint-new` on the staged Go changes.
- **On push** (`.githooks/pre-push`): `go vet` + `lint-new` + `go test ./...` + `go.mod` tidy check.
- **In CI** (`.github/workflows/ci.yml`):
  - *unit job* (no DB): `fmt-check` + `vet` + `lint-new` + `tidy-check` + boundaries + unit tests + build.
  - *gate job* (authoritative): `make ci-container` (full suite against a real DB) + fuzz seeds.
  - *coverage job*: `make coverage-check` inside the toolbox (real DB) — enforces the 90% floor.

Bypass a hook only in a genuine emergency: `git commit --no-verify` / `git push --no-verify`.

## Why `lint-new` and not full `lint` as the gate

golangci-lint was not previously wired into `make ci`, so a backlog of pre-existing issues accumulated
(see [lint-backlog.md](lint-backlog.md)). Blocking on the *full* set would either force a large risky
cleanup or invite blanket `//nolint`. Instead we gate on **new/changed code**
(`--new-from-merge-base=origin/main`): every new line is fully linted immediately, while the backlog is
burned down incrementally and honestly. Run `make lint` any time to see the full picture.

## Adding/adjusting linters

The linter set and the import-law `depguard` rules live in `.golangci.yml`. Formatting is `gofumpt` +
`goimports` (v2 `formatters`). Keep the enforced set green on new code; widen it as the backlog shrinks.
