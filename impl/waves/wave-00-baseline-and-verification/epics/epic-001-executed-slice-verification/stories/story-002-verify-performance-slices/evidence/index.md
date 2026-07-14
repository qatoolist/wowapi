---
id: EVIDENCE-INDEX-W00-E01-S002
type: evidence-index
parent_story: W00-E01-S002
status: recorded
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Evidence index — W00-E01-S002

Per `evidence-policy.md` required fields. All three evidence records were produced on 2026-07-13 at
commit `0a31186cada5c275a588c74081cf977adf346e61` (branch `main`) — which is itself PR #25's merge
commit. Full records: `evidence/tests/EV-W00-E01-S002-01.md`, `evidence/tests/EV-W00-E01-S002-02.md`,
`evidence/baselines/EV-W00-E01-S002-03.md`. Raw logs live under this story's `artifacts/`.

| Evidence ID | Evidence type | Story and task | Acceptance criteria proven | Execution command | Code revision / commit SHA | Branch or tag | Execution environment | Tool versions | Date and time | Result | File or URI | Checksum | Reviewer | Superseded evidence |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| EV-W00-E01-S002-01 | test-execution log + bench-budget tool output | W00-E01-S002 / T001 | AC-W00-E01-S002-01 | `go test ./kernel/httpx/... -race` then `make bench-budget` | `0a31186cada5c275a588c74081cf977adf346e61` | main | Local dev machine — macOS 26.5.2 (Darwin 25.5.0), arm64 Apple M3 Max; concurrent load present (sibling W00 workers running tests) | go1.26.5 darwin/arm64; GNU Make 3.81; git 2.55.0 | 2026-07-13 12:11:56–12:13:43 +05:30 | pass (both exit 0; no races; 43/43 budgets OK) | `evidence/tests/EV-W00-E01-S002-01.md`; logs: `artifacts/T001-go-test-kernel-httpx-race.log`, `artifacts/T001-make-bench-budget.log` | sha256 `7887dcc4…7150fd` / `5fdd45bf…880745` (full values in record) | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |
| EV-W00-E01-S002-02 | test-execution log + fail-first gate-check log | W00-E01-S002 / T002 | AC-W00-E01-S002-02 | `go test ./internal/tools/benchbudget/... -run TestMainMissingBenchmarkFails -v` (name reconfirmed at coverage_test.go:103 before running) + scratch-budgets ghost-entry fail-first check via `go run ./internal/tools/benchbudget` | `0a31186cada5c275a588c74081cf977adf346e61` | main | Local dev machine — macOS 26.5.2 (Darwin 25.5.0), arm64 Apple M3 Max; concurrent load present (sibling W00 workers running tests) | go1.26.5 darwin/arm64; GNU Make 3.81; git 2.55.0 | 2026-07-13 12:14:40–12:15:54 +05:30 | pass (test exit 0, PASS; fail-first check: gate exit 1, `budgeted but not found in bench output`) | `evidence/tests/EV-W00-E01-S002-02.md`; logs: `artifacts/T002-go-test-benchbudget-missingbenchmark.log`, `artifacts/T002-failfirst-ghost-check.log`, `artifacts/T002-scratch-budgets-ghost.txt` | sha256 `91836074…f35d1` / `438bf32f…5303c8` / `45d7f4e1…2eca43` (full values in record) | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |
| EV-W00-E01-S002-03 | inspection note | W00-E01-S002 / T003 | AC-W00-E01-S002-03 | manual inspection of `bench-budgets.txt` + `git ls-files '*bench-budgets*'` + `git show 0a31186 -- bench-budgets.txt` + `git diff 0a31186 -- bench-budgets.txt` + `git status --porcelain bench-budgets.txt` | `0a31186cada5c275a588c74081cf977adf346e61` | main | Local dev machine — macOS 26.5.2 (Darwin 25.5.0), arm64 Apple M3 Max; inspection only, load-insensitive | go1.26.5 darwin/arm64; git 2.55.0 | 2026-07-13 ~12:16 +05:30 | pass (43 entries, both counting methods agree; file byte-identical to #25 state; 3-entry spot-check matches; RISK-W00-003 does not materialize) | `evidence/baselines/EV-W00-E01-S002-03.md`; note: `artifacts/T003-bench-budgets-inspection-note.md` | sha256 `649fadfb…cb16c3` (full value in record) | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |

Per `evidence-policy.md` revision-pinning rule, each record above cites the exact commit SHA it was
captured against. No run failed, so no `failed` record exists; the failed-evidence-preservation rule
was not triggered.
