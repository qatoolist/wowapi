# Phase 0 — Command Log (2026-07-03)

All commands run from the repo root on the author machine (darwin/arm64, go1.26.4).

| # | Command | Exit | Result summary |
|---|---|---|---|
| 1 | `go version` | 0 | go1.26.4 darwin/arm64 |
| 2 | `docker version` | — | **daemon unavailable** (OrbStack socket missing). Container files created + statically validated; live `make up` deferred (risk R4). |
| 3 | `sh scripts/graphify_refresh.sh check` | 0 | Graphify check-update clean at phase start (Goal 2 §Graphify) |
| 4 | `go build ./...` | 0 | all packages compile |
| 5 | `go vet ./...` | 0 | clean |
| 6 | `go test ./...` (first run) | 1 | **FAIL**: app cycle test — module name regex required ≥2 chars, single-letter test modules rejected as invalid names before cycle detection. Fixed: `module/module.go` nameRE `{1,63}`→`{0,63}` (+ matching hint in app/app.go). Recorded as review finding self-caught by tests. |
| 7 | `go test ./...` (after fix) | 0 | ok: app, internal/cli, kernel/config, kernel/secrets |
| 8 | `go test -race ./...` | 0 | clean |
| 9 | `sh scripts/lint_boundaries.sh` | 0 | "boundary lint: OK" |
| 10 | `go run ./cmd/wowapi version` | 0 | `wowapi devel` + `context: wowapi framework repository` |
| 11 | `go run ./cmd/wowapi new-module x` | 2 | honest not-implemented message naming Phase 10 |
| 12 | `make help` | 0 | grouped target list renders |
| 13 | `make ci` | 0 | vet → lint-boundaries → test-unit → test-race → build (bin/wowapi) all green |
| 14 | `docker compose -f deployments/compose.yaml config --quiet` | 0 | compose file parses/validates (static; no daemon needed) |
| 15 | (post-review fixes) `gofmt -l .` + `go vet ./...` + `go test ./...` + `go test -race ./...` + `sh scripts/lint_boundaries.sh` | 0 | all green after fixing ARCH-1..4, SEC-2 (see review-findings.md) |
| 16 | two parallel review agents (architecture/boundaries; security/config) | — | findings + resolutions in review-findings.md; fixes verified by #15 |
| 17 | `sh scripts/graphify_refresh.sh update` | 0 | code graph rebuilt: 456 nodes, 440 edges, 40 communities (graphify-out/, not committed per policy) |
| 18 | `git commit` → `git show --stat` audit | 0 | **incident caught by post-commit stat audit:** user-global gitignore rule `**/secrets/` silently excluded `kernel/secrets/` source from the commit. Fix: repo `.gitignore` negations (`!kernel/secrets/`, `!adapters/secrets/`); commit amended (unpushed) → `0057fec`, 37 files, `git status` clean, `git ls-files kernel/secrets/` shows both files. Lesson encoded: pre-commit checklist now includes a created-files vs `git show --stat` diff. |

## Could not run (documented per evidence rules)
- `make up` / `make ci-container` / Dockerfile build — Docker daemon not running on the author
  machine at Phase 0 time. Residual risk: Dockerfile/compose only statically validated (cmd 14).
  Must run: first environment with a daemon (CI bring-up or next local session) before Phase 2
  integration tests, which require live Postgres.
- `scripts/graphify_refresh.sh extract` — no LLM backend key configured (docs/graphify.md); `check`
  ran clean instead (risk R11).
