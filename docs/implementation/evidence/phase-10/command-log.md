# Phase 10 — Command Log

Exact commands with exit status; summarized output where long. Date: 2026-07-04.

| # | Command | Exit | Output summary |
|---|---|---|---|
| 1 | `go test ./internal/cli/` (lead's 5 commands) | 0 | migrate create (next-number, empty-dir, bad-name), openapi merge (+dup-path fail), seed validate (+foreign-key fail, missing-module), lint checkBoundaries (isolation, clean, framework-layer), deploy render (compose, env, bad-format) |
| 2 | `go build -o /tmp/wowapi ./cmd/wowapi && wowapi lint boundaries` | 0 | `boundary lint: OK` on the framework repo — agrees with `scripts/lint_boundaries.sh` |
| 3 | `wowapi help` | 0 | lists all implemented commands; no "planned" stubs remain |
| 4 | `wowapi init --module github.com/acme/shop && new-module --name orders && gen crud --module … --resource order --fields "title:string,qty:int"` (in a temp dir) | 0 | scaffolds go.mod/cmd/api|worker|migrate/configs; module.go+migrations+seeds+openapi.json; order.go + migration |
| 5 | `gofmt -l` on generated Go | 0 | empty — generated `.go` is gofmt-clean (renderToFile runs go/format.Source on .go output; invalid Go fails generation loudly) |
| 6 | `go test ./internal/cli/` (full, incl. scaffold golden tests) | 0 | 64+ tests: init/new-module/gen golden (parse-check generated Go, correct module Name(), permission keys, RLS migration) + lead's command tests |
| 7 | `make ci` (host) | 0 | vet, boundary lint, unit, race, build green |
| 8 | `make ci-container` | 0 | green in the tools container |
