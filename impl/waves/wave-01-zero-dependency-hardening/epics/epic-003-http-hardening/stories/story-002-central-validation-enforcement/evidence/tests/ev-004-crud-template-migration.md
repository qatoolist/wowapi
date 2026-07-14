# EV-W01-E03-S002-004 — crud template migration proof

- **Evidence ID**: EV-W01-E03-S002-004
- **Evidence type**: unit-test report (generator output)
- **Story / task**: W01-E03-S002 / W01-E03-S002-T003
- **Acceptance criteria proven**: AC-W01-E03-S002-03 (transitively — template-correctness proof per the task's own framing)
- **Execution command**: `go test ./internal/cli/ -run 'TestGenCRUDMutatingRoutesDeclareContractsAndUseValidatedHandler|TestGenCRUDPermissionKeys|TestGenCRUDResourceGoParsable' -count=1 -v`
- **Code revision / commit SHA**: 0a31186cada5c275a588c74081cf977adf346e61 (working tree on top of this HEAD; conductor owns the wave commit — the diff is the uncommitted working change, `git diff --stat` recorded in the story's implementation.md)
- **Branch**: main
- **Execution environment**: darwin/arm64 workstation (local)
- **Tool versions**: go1.26.5; gosec (dev build)
- **Date/time**: 2026-07-13 13:06 IST
- **Reviewer**: pending — W01 wave review gate (conductor)
- **Result**: PASSED — `wowapi gen crud` output now declares `Create<R>Request`/`Update<R>Request` contracts (with `validate:"required"` starter tags) on POST/PUT `RouteMeta.Request` and wires both handlers through `httpx.ValidatedHandler(v, 1<<20, ...)`; generated file parses; W01Gen's `.deactivate` DELETE permission preserved (coordination via irc, no overlap). Full `go test -race ./internal/cli/` also green at the same SHA (includes the rendered-product compile tests).

```
=== RUN   TestGenCRUDResourceGoParsable
--- PASS: TestGenCRUDResourceGoParsable (0.00s)
=== RUN   TestGenCRUDMutatingRoutesDeclareContractsAndUseValidatedHandler
--- PASS: TestGenCRUDMutatingRoutesDeclareContractsAndUseValidatedHandler (0.00s)
=== RUN   TestGenCRUDPermissionKeys
--- PASS: TestGenCRUDPermissionKeys (0.00s)
ok  	github.com/qatoolist/wowapi/internal/cli	0.255s
```
