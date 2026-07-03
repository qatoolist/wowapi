# Phase 1 — Command Log

Exact commands with exit status; summarized output where long. Date: 2026-07-03.

| # | Command | Exit | Output summary |
|---|---|---|---|
| 1 | `go get gopkg.in/yaml.v3@latest` | 0 | added gopkg.in/yaml.v3 v3.0.1 (first third-party dep, D-0011) |
| 2 | `go test ./kernel/config/` (before loader impl) | ≠0 | compile failure — TDD red state for load_test.go (undefined Load/Options/…) |
| 3 | `go mod tidy && go test ./kernel/config/` | 0 | loader implemented; full config suite green on first run |
| 4 | `go vet ./... && go test ./... -count=1 && sh scripts/lint_boundaries.sh` | 0 | all packages green; `boundary lint: OK` |
| 5 | `make up` | 0 | compose stack started: postgres/minio/mailpit healthy, tools up — **R4 gate 1 closed** (Docker available this session) |
| 6 | `sh scripts/graphify_refresh.sh check` | 0 | graph fresh at phase start (Goal 2 Graphify duty) |
| 7 | `go test ./kernel/logging/ -count=1` (agent B) | 0 | 11 tests pass (formats, level filter, redaction, LogStartup) |
| 8 | `go test ./app/... ./adapters/... -count=1` + `gofmt -l` + `go vet` + boundary lint (agent C) | 0 | 28 tests pass (envprovider, process views narrowing, module context isolation, RunHooks lifecycle); `boundary lint: OK` |
| 9 | `go test ./internal/cli/ -count=1` (agent D) | 0 | 19 tests pass (validate/print/schema/doctor, exit codes, redaction gate) |
| 10 | `gofmt -l .` + `make ci` (integration) | 0 | vet, boundary lint, unit, race, build all green across 7 packages |
| 11 | CLI smoke in scratch dir: `wowapi config validate` / `doctor` / `print --redacted` on good config; `validate` on bad prod config | 0/0/0/1 | good: `config OK fingerprint=2d8db6f0bd12`, provenance table, redacted JSON; bad prod config: all 3 errors accumulated (unknown key `htp.x`, prod format, prod debug level), exit 1 |
| 12 | `make ci-container` (first attempt) | 2 | vet/boundary-lint/unit green in container; **race failed**: golang alpine has no C toolchain → cgo auto-disabled. Root cause fixed in Dockerfile dev stage (gcc, musl-dev, CGO_ENABLED=1) |
| 13 | `docker compose build tools && make ci-container` (after fix) | 0 | full CI green inside the container (vet, boundaries, unit, race, build) — **R4 closed** |
| 14 | `gofmt -l . && make ci && make ci-container` (after review-fix pass) | 0 | host + container CI both fully green with all 19 review fixes and their regression tests |
| 15 | `sh scripts/graphify_refresh.sh update` | 0 | graph updated post-phase (extract still blocked on LLM key, R11) |
