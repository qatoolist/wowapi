# EV-W01-E03-S002-003 — waiver-exemption boot success

- **Evidence ID**: EV-W01-E03-S002-003
- **Evidence type**: unit-test report
- **Story / task**: W01-E03-S002 / W01-E03-S002-T001
- **Acceptance criteria proven**: AC-W01-E03-S002-04
- **Execution command**: `go test ./kernel/httpx/ -run 'RequestContract' -count=1 -v` (part of the ev-001 command's runs)
- **Code revision / commit SHA**: 0a31186cada5c275a588c74081cf977adf346e61 (working tree on top of this HEAD; conductor owns the wave commit — the diff is the uncommitted working change, `git diff --stat` recorded in the story's implementation.md)
- **Branch**: main
- **Execution environment**: darwin/arm64 workstation (local)
- **Tool versions**: go1.26.5; gosec (dev build)
- **Date/time**: 2026-07-13 13:06 IST
- **Reviewer**: pending — W01 wave review gate (conductor)
- **Result**: `TestRouterRequireRequestContractsWaiverExemptsBodylessMutation` — a POST route with `NoRequestBody: true` and no Request contract registers cleanly with enforcement ON. The companion adversarial guard `TestRouterRejectsContractWithWaiverContradiction` proves Request+NoRequestBody together is rejected unconditionally (the waiver cannot silently coexist with a declared contract). Logs: EV-W01-E03-S002-001 stage 3.
