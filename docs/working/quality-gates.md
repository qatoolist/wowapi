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
| `make lint` | **Full** golangci-lint across the whole tree — **enforced** (pinned `GOLANGCI_VERSION`; backlog B-1 closed, D-0087/D-0089) | CI, manual |
| `make lint-new` | golangci-lint on **changed code only** (`--new-from-merge-base`) — fast local pre-check (CI enforces the full `make lint` above) | pre-commit, pre-push |
| `make tidy` / `tidy-check` | `go mod tidy` / fail if `go.mod`/`go.sum` drift | pre-push, CI |
| `make check` | `fmt-check vet lint-new tidy-check test-unit` — fast local pre-flight | manual |
| `make coverage` | Statement coverage against the **real DB** (`WOWAPI_REQUIRE_DB=1`), over the measured package set | manual |
| `make coverage-check` | Fail if total coverage drops below `COVERAGE_FLOOR` (**84.0%, interim** — target 90%) | CI coverage job |
| `make ci` / `ci-container` | Authoritative correctness gate (vet, boundaries, unit, race, bench, build); `ci-container` runs it with a real DB (`WOWAPI_REQUIRE_DB=1`) | CI gate job |

## Coverage floor

`make coverage-check` enforces `COVERAGE_FLOOR` (`Makefile`), currently **84.0% — an interim floor**, not the
approved target. The approved requirement is a **90%** floor; `COVERAGE_FLOOR` was lowered to 84.0 alongside a
genuine coverage regression (92.3% → 84.5%) with no decision record at the time, and DEC-PROG-001
(`impl/tracking/programme-decisions.md`) now documents that as an acknowledged interim floor with a
ratchet-upward plan back to 90%. **DEC-PROG-001 is `proposed` and awaits human ratification** — a drop below
84.0 remains release-blocking in the meantime. Coverage is measured **with a real Postgres** — much of the
kernel (RLS, outbox, jobs, audit) only executes against the DB, so a no-DB run understates it. The measured set
**excludes** packages that cannot be meaningfully unit-tested: `cmd/wowapi` (process main),
`internal/tools/migrate` (tool main), `internal/testmodules` (test fixture), and `module` (interface-only) —
see `COVER_EXCLUDE` in the `Makefile`. The CI `coverage` job runs `make coverage-check` inside the toolbox
container so DB-backed tests contribute.

## When each check runs

- **On commit** (`.githooks/pre-commit`, fast): a `gitleaks` secret scan of the staged diff (all file types,
  not just code — mirrors CI's blocking gitleaks job) + `fmt-check` + `lint-new` on the staged Go changes.
- **On push** (`.githooks/pre-push`): a `gitleaks` secret scan of the outgoing commit range (same detect
  semantics as CI) + `go vet` + `lint-new` + `go test ./...` + `go.mod` tidy check. `make secret-scan` gives a
  full-range CI-parity gitleaks scan (`merge-base(origin/main)^..HEAD`) covering commits the incremental
  pre-push scan never saw.
- **In CI** (`.github/workflows/ci.yml`):
  - *unit job* (no DB): `fmt-check` + `vet` + full `make lint` + `tidy-check` + boundaries + unit tests + build.
  - *gate job* (authoritative): `make ci-container` (full suite against a real DB) + fuzz seeds.
  - *coverage job*: `make coverage-check` inside the toolbox (real DB) — enforces the interim 84.0% floor
    (see [Coverage floor](#coverage-floor)).
  - *reference-smoke job*: `make smoke-reference` — a scaffolded product behind the reference nginx (TLS),
    with the security headers smoke-tested through the proxy (B-7).

Bypass a hook only in a genuine emergency: `git commit --no-verify` / `git push --no-verify`.

## Full `make lint` is the gate; `lint-new` is the fast local pre-check

CI enforces the **full** `make lint` across the whole tree, on a **pinned** golangci-lint
(`GOLANGCI_VERSION`, so the gate is deterministic — a new upstream release can't fail CI until it's bumped).
This is safe because the pre-existing backlog (B-1) is closed (D-0087) — `make lint` is a clean 0.

`make lint-new` (`--new-from-merge-base=origin/main`) stays wired into the pre-commit/pre-push hooks as a
**fast local pre-check** — it lints only changed code, giving quick feedback before the full-tree gate runs in
CI. Historically (before B-1 was burned down) `lint-new` was itself the enforced gate, to avoid a large risky
cleanup or blanket `//nolint`; that history is recorded in [lint-backlog.md](lint-backlog.md).

## Adding/adjusting linters

The linter set and the import-law `depguard` rules live in `.golangci.yml`. Formatting is `gofumpt` +
`goimports` (v2 `formatters`). Keep the enforced set green on new code; widen it as the backlog shrinks.
