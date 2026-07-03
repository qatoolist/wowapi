# Phase 3 — Command Log

Exact commands with exit status; summarized output where long. Date: 2026-07-03.

| # | Command | Exit | Output summary |
|---|---|---|---|
| 1 | `go test ./kernel/errors/ -count=1` | 0 | 10 tests (taxonomy closed-set mapping, 404-masked tenant isolation, msg-not-format-string, KindOf through wrapping, Wrapf preserves Kind) |
| 2 | `go get go-playground/validator/v10` + `go test ./kernel/validation/` (agent) | 0 | 11 tests (json field paths incl. nested, tag→code mapping, value never in message) |
| 3 | `go test ./kernel/pagination/ ./kernel/filtering/` (agent) | 0 | cursor round-trip, per_page clamping, injection payload lands only in args, unknown field/op rejected, placeholder numbering |
| 4 | `go test ./kernel/httpx/ ./app/... ./module/...` | 0 | route metadata enforcement, error→problem mapping, internal-never-leaks, strict decode, etag, recover-no-leak, envelope |
| 5 | `make test-integration` (idempotency store, live PG) | 0 | fresh/replay/conflict/in-flight + cross-tenant isolation; byte-exact replay (bytea, not jsonb) |
| 6 | `make test-integration` (WithIdempotency HTTP composition) | 0 | op runs exactly once across two same-key requests; reused key + different hash → 409 |
| 7 | `unset DATABASE_URL; make ci` | 0 | vet, boundary lint, unit (integration tests skip cleanly without DSN), race, build all green |
| 8 | `docker compose run --rm tools make test-integration` | 0 | Phase 3 integration suite (idempotency store + WithIdempotency) green inside the tools container |
| 9 | SEC-16/ARCH-27 reproduced (review agents, live PG) then fixed; `TestIntegrationIdempotencyConcurrent -count=5 -race` | 0 | 8-goroutine same-key race: operation runs exactly once, stored response preserved; ×5 under race detector |
| 10 | `golangci-lint run` (depguard import-law check) | n/a | **0 depguard violations** — kernel imports no upward layers (pre-existing errcheck/misspell in test files are outside `make ci`'s vet gate) |
| 11 | `gofmt -l . && sh scripts/lint_boundaries.sh && make ci && make test-integration` (after all review fixes) | 0 | boundary lint OK; host CI (vet/lint/unit/race/build) + integration all green |
