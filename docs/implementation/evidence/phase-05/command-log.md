# Phase 5 — Command Log

Exact commands with exit status; summarized output where long. Date: 2026-07-03.

| # | Command | Exit | Output summary |
|---|---|---|---|
| 1 | `go test ./kernel/seeds/` | 0 | seed loader: merge, module-prefix ownership rejection, strict unknown-key rejection, empty-fs |
| 2 | `go build ./...` (kernel + app boot + testkit + testmodule) | 0 | Kernel composition root, app.Boot lifecycle, contract suite, requests fixture compile |
| 3 | `sh scripts/lint_boundaries.sh` | 0 | OK — testkit composes app/kernel (allowed); internal/testmodules domain-neutral, no forbidden imports |
| 4 | `DATABASE_URL=… go test -run TestIntegrationRequestsModuleContract ./testkit/` | 0 | contract suite green: boots on empty namespace, migrate+seed idempotent, RLS on module table, rejects invalid config key |
| 5 | `DATABASE_URL=… go test -run TestIntegrationScratchConsumer ./testkit/` | 0 | **external repo** (tmpdir, `replace` wowapi) imports public packages, defines a `widgets` module, passes RunModuleContract — zero framework edits |
| 6 | `make ci` (host, no DSN) | 0 | vet, boundary lint, unit (seed loader + boot), race, build green (integration tests skip without DSN) |
| 7 | `make test-integration` (all packages) | 0 | authz/relationship/resource/testkit integration green incl. contract + consumer |
| 8 | `docker compose run --rm tools make test-contract` | 0 | contract + scratch-consumer suite green INSIDE the tools container — proves the external-consumer flow works without host Go (container-first) |
