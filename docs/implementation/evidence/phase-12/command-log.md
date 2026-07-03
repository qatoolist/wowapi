# Phase 12 — Command Log

Exact commands with exit status; summarized output where long. Date: 2026-07-04.

| # | Command | Exit | Output summary |
|---|---|---|---|
| 1 | `go build ./...` (after wiring the init templates into real mains) | 0 | full framework tree builds |
| 2 | `go test ./internal/cli/` | 0 | scaffold golden tests still green — the wired templates render gofmt-clean, parseable Go (go/format.Source enforces it) |
| 3 | `DATABASE_URL=… go test -run E2E -v ./internal/e2e/` | 0 | scaffolds a repo via `wowapi init`, `replace`s wowapi with the local tree, `go mod tidy` + `go build ./...` (api/worker/migrate compile), `go vet`; then **ran the built migrate binary** (criterion #22 ✓) and **started the api binary → GET /healthz 200** (criterion #19 runtime ✓) — the startup log shows the Phase-11 config fingerprint + the canonical request access-log line |
| 4 | `go test ./testkit/ -run ScratchConsumer` | 0 | external scratch-consumer imports only public packages + runs the module contract (criterion #21) |
| 5 | `make ci` (host) | 0 | vet, boundary lint, unit, race, bench-budget, build green |
| 6 | `make ci-container` | 0 | green in the tools container — zero role/concurrent-update flakes (the Phase-11 root fix holds) |
| 7 | 28-criterion acceptance sweep | — | `acceptance-map.md`: all 28 framework acceptance criteria mapped to their delivering phase + concrete proof |
