---
id: EVIDENCE-W00-E02-S001
type: evidence-index
parent_story: W00-E02-S001
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Evidence index — W00-E02-S001

Per mandate §10. All four expected evidence records were produced on 2026-07-13 at commit
`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`). Category subdirectories were created
on first real content per Adaptation 2 (`impl/governance/naming-conventions.md`). Full
field-complete records live in the linked files; the table summarizes.

| Evidence ID | Evidence type | Story and task | Acceptance criteria proven | Execution command | Code revision or commit SHA | Branch or tag | Execution environment | Relevant tool versions | Date and time | Result | File or URI | Checksum | Reviewer | Superseded evidence |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| EV-W00-E02-S001-001 | coverage report | W00-E02-S001 / T001 | AC-W00-E02-S001-01 | `make coverage-check` (wraps `go test -coverprofile=coverage.out $(COVER_PKGS)` + `go tool cover -func`) | 0a31186cada5c275a588c74081cf977adf346e61 | main | local (macOS arm64), real Postgres 16 via compose, `WOWAPI_REQUIRE_DB=1`; concurrent load present | go1.26.5 darwin/arm64; postgres:16-alpine | 2026-07-13 | **PASS — total 92.3% vs 90.0% floor, exit 0** | [coverage/EV-W00-E02-S001-001.md](coverage/EV-W00-E02-S001-001.md); raw: `../artifacts/coverage/coverage-check-output.txt` | n/a | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |
| EV-W00-E02-S001-002 | static-analysis report | W00-E02-S001 / T002 | AC-W00-E02-S001-02 | `golangci-lint run -c .golangci.matrix-cs23.yml ./...` (throwaway config; committed `.golangci.yml` untouched) + committed-config control run | 0a31186cada5c275a588c74081cf977adf346e61 | main | local (macOS arm64), full checkout; concurrent load present (see run-1 anomaly note in record) | golangci-lint v2.11.4 (pinned); go1.26.5 | 2026-07-13 | **Captured — 991 issues (throwaway) / 0 issues (committed control); zero-hit set MATCH; drift flagged per-analyzer (noctx, exhaustive, errorlint, gosec, forcetypeassert, wrapcheck, revive); MATRIX names 18 of claimed 25 analyzers — flagged** | [static-analysis/EV-W00-E02-S001-002.md](static-analysis/EV-W00-E02-S001-002.md); raw: `../artifacts/static-analysis/` | n/a | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none (run-1 anomalous output preserved as artifact) |
| EV-W00-E02-S001-003 | benchmark results | W00-E02-S001 / T003 | AC-W00-E02-S001-03 | `grep -vc '^#' bench-budgets.txt`; `make bench-budget` | 0a31186cada5c275a588c74081cf977adf346e61 | main | local (macOS arm64), real Postgres via compose; concurrent load present — timed window serialized with the one bench-active sibling (W00E01S002, confirmed idle via IRC) | go1.26.5; `internal/tools/benchbudget` at HEAD | 2026-07-13 | **PASS — 43 entries confirmed (post-#25, no drift); 43/43 OK, 0 violations, exit 0** | [benchmarks/EV-W00-E02-S001-003.md](benchmarks/EV-W00-E02-S001-003.md); raw: `../artifacts/benchmarks/` | n/a | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |
| EV-W00-E02-S001-004 | CI execution record | W00-E02-S001 / T003 | AC-W00-E02-S001-04 | `gh run view 29229288699 --json jobs` + direct read of `.github/workflows/ci.yml` | 0a31186cada5c275a588c74081cf977adf346e61 (checkout **and** observed run headSha) | main | hosted GitHub Actions runners (data source: hosted run history — preferred, not local fallback) | gh CLI; workflow pins golangci-lint v2.11.4, actionlint v1.7.12 | 2026-07-13 (run 06:32–06:35 UTC) | **Captured — total 3m24s; gate legs test 1m39s / race 3m04s / bench 3m12s; unit 2m14s; coverage 2m54s; SD-01/SD-02 shape confirmed** | [logs/EV-W00-E02-S001-004.md](logs/EV-W00-E02-S001-004.md); raw: `../artifacts/ci-timing/ci-wallclock-run-29229288699.txt` | n/a | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |

## Failed-evidence preservation note

Per `impl/governance/evidence-policy.md`, no capture attempt failed to run. One anomalous first
lint run (11 nondeterministic staticcheck SA5011 hits under concurrent load, not reproducible in
two cache-cleaned re-runs) is fully disclosed in EV-W00-E02-S001-002 and its raw output preserved
at `../artifacts/static-analysis/lint-25-analyzer-run1-anomalous.{txt,json}` — not deleted.

## Cross-reference — coverage floor superseded (2026-07-16)

EV-W00-E02-S001-001's captured baseline (92.3% coverage vs. a 90.0% floor, pinned to
`0a31186`) has been superseded by the coverage-floor reduction landed in `e8cda6b` ("finalize
wowapi implementation programme"), which lowers the operative floor to 84.0% (current measured
coverage 84.5% as of HEAD `43b6e12`). This baseline-capture work itself remains sound and
honestly self-reported; the floor it measured against is simply no longer the operative gate. See
`impl/tracking/programme-deviations.md` **DEV-PROG-001** (coverage floor lowered 90.0% → 84.0%
without a deviation/decision record at the time) and **DEC-PROG-001** (interim floor ratification,
status: proposed — human ratification pending) for the full traceability chain.

— dated 2026-07-16, conductor adjudication (Fable 5), per review-gate-2026-07-16.md records
