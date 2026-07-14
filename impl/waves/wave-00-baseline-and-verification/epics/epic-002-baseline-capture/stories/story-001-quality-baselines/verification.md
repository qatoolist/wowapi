---
id: VER-W00-E02-S001
type: verification-record
parent_story: W00-E02-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Verification record — W00-E02-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story. No row has been executed yet —
this table states the planned method only.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E02-S001-01 | Run `make coverage-check` (wraps `go test -coverprofile=coverage.out $(COVER_PKGS)` then `go tool cover -func=coverage.out`) against the real Postgres test DB; capture the `total:` line percentage and confirm it is at or above the committed `COVERAGE_FLOOR` (90.0, `Makefile:240`); record the exact percentage as a fresh measurement, not the prior ~92% figure. | Local dev environment or CI container with `docker compose -f deployments/compose.yaml up -d --wait postgres minio mailpit` running, `DATABASE_URL`/`TEST_DSN` reachable, `WOWAPI_REQUIRE_DB=1` set, pinned Go toolchain per `go.mod`/CI `GO_VERSION`. | Command completes; a numeric coverage percentage is printed; percentage is at or above 90.0% (if below, this is itself a finding to record, not a blocker to re-running the command). | Coverage report (`coverage.out` + `coverage.html`, plus the captured terminal summary) — evidence type "coverage report" per `evidence-policy.md`. | unassigned |
| AC-W00-E02-S001-02 | Build a throwaway `golangci-lint` config variant (starting from the committed `.golangci.yml`, preserving its `exclusions` block, additionally enabling all 25 analyzers named in MATRIX CS-23); run `golangci-lint run -c <throwaway-config> ./...` with the pinned v2.11.4 binary; tabulate hit counts per analyzer; compare against the MATRIX CS-23 snapshot (zero-hit set: sqlclosecheck, rowserrcheck, bodyclose, wastedassign, makezero, musttag, testifylint; near-zero: noctx 2, copyloopvar 1, gocritic exitAfterDefer 1; gosec 38 with named triage list G704/G120/G115/G304; nilerr/exhaustive/errorlint adjudicated as deliberate); flag any analyzer whose fresh count differs from its MATRIX-time count. | Local dev environment or CI container with `golangci-lint` v2.11.4 installed (pinned, matching `Makefile:16`/`ci.yml:62`), full repository checked out at the story's execution commit. | Command completes (non-zero exit from `golangci-lint run` is expected and acceptable — it reflects real findings, not a broken run); a per-analyzer hit-count table is produced; every one of the 25 analyzers has an explicit recorded count (including zero); every count is compared against its MATRIX CS-23 value with match/drift explicitly stated. | Static-analysis report, including the analyzer-by-analyzer drift table — evidence type "static-analysis report" per `evidence-policy.md`. | unassigned |
| AC-W00-E02-S001-03 | Run `make bench-budget` (wraps `go test -bench=. -benchmem -run=^$ $(BENCH_PKGS)` piped through `internal/tools/benchbudget bench-budgets.txt`) against the real Postgres test DB; capture the pass/fail result per budgeted benchmark; separately confirm `bench-budgets.txt`'s entry count equals the expected post-#25 value (43) by direct inspection of the file. | Local dev environment or CI container with Postgres reachable (`WOWAPI_REQUIRE_DB=1`), pinned Go toolchain, repository checked out at the story's execution commit. | Command completes; every budgeted benchmark reports pass or a specific violation (a violation is itself a valid, recordable result — not a blocker to completing the capture); `bench-budgets.txt` entry count is confirmed and stated explicitly (expected 43; any other count is drift to flag per RISK-W00-003). | Benchmark result, including the entry-count confirmation — evidence type "benchmark results" per `evidence-policy.md`. | unassigned |
| AC-W00-E02-S001-04 | Read the current `.github/workflows/ci.yml` and record its job/leg structure (`changes`, `workflow-lint`, `unit`, `gate` matrix `[test, race]`, `gate-bench`, `reference-smoke`, `coverage`); obtain per-leg wall-clock either from the most recent accessible GitHub Actions run history for this workflow, or, if run history is not accessible in the execution environment, from a fresh local timed run of the container-equivalent targets (`make ci-container-test`, `make ci-container-race`, `make ci-container-bench`) with the substitution explicitly noted as an approximation (local timing differs from hosted-runner timing due to cache/hardware differences). Explicitly record the SD-01 (3-leg parallelization, #23) and SD-02 (bench path-scoping, nightly schedule, #24) session-delta facts this baseline reflects. | GitHub Actions run-history access (preferred) OR local Docker environment with `docker compose -f deployments/compose.yaml` reachable (fallback), repository checked out at the story's execution commit. | A per-leg timing figure (or best-available approximation, explicitly labeled as such) is recorded for each of the jobs listed above; the CI-execution-record evidence explicitly states which data source (hosted run history vs. local approximation) was used. | CI execution record — evidence type "CI execution record" per `evidence-policy.md`. | unassigned |

## Post-execution record

Executed 2026-07-13. Per-AC actual results below; full detail in the linked evidence records.

| Acceptance criterion | Actual result | Pass/fail | Evidence ID |
|---|---|---|---|
| AC-W00-E02-S001-01 | `make coverage-check` at 0a31186 against real Postgres (`WOWAPI_REQUIRE_DB=1`): `total: 92.3%` vs floor 90.0%, exit 0. Fresh measurement; prior ~92% history figure reconfirmed, not assumed. | **pass** | EV-W00-E02-S001-001 |
| AC-W00-E02-S001-02 | Throwaway v2.11.4 config (committed config + the 18 analyzers MATRIX CS-23 names verbatim; committed `.golangci.yml` unmodified): 991 issues, per-analyzer counts recorded incl. zeros; analyzer-by-analyzer drift table vs MATRIX produced; drift explicitly flagged (noctx, exhaustive, errorlint, forcetypeassert, gosec, wrapcheck, revive), matches confirmed (zero-hit set ×7, copyloopvar, gocritic, nilerr, named adjudication sites). MATRIX's "25" headcount is unsubstantiated in the source — 18 names recoverable; recorded as DEV-W00-E02-S001-001, not silently absorbed. | **pass** (as a capture; the flagged drift is the recorded finding set) (pass-as-capture ratified by conductor 2026-07-13 per DEV-W00-E02-S001-001; 18-of-25 analyzer-name gap acknowledged; drift facts carried into W01-E01 briefs) | EV-W00-E02-S001-002 |
| AC-W00-E02-S001-03 | `bench-budgets.txt` entry count confirmed **43** by direct inspection (post-#25 expected value — no drift, RISK-W00-003 mitigated); `make bench-budget` at 0a31186 with real DB: 43/43 OK, 0 violations, exit 0. | **pass** | EV-W00-E02-S001-003 |
| AC-W00-E02-S001-04 | `ci.yml` job/leg structure confirmed by direct read at 0a31186 (RISK-W00-005 mitigated); per-leg wall-clock from hosted GH Actions run 29229288699 whose headSha equals the execution commit (data source explicitly stated; local fallback not used): total 3m24s; unit 2m14s, gate test/race/bench 1m39s/3m04s/3m12s, coverage 2m54s, smoke 1m00s, workflow-lint 25s, changes 6s. SD-01 (#23) and SD-02 (#24) facts explicitly noted; SD-03 (#25) visible in run history (14m23s pre-#25 vs ≤3m50s post). | **pass** | EV-W00-E02-S001-004 |

### Actual result

All four baselines captured and registered; see table above.

### Pass or fail

**Pass — 4/4 acceptance criteria.**

### Evidence identifier

EV-W00-E02-S001-001, EV-W00-E02-S001-002, EV-W00-E02-S001-003, EV-W00-E02-S001-004 (all in
`evidence/index.md`).

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (main) — all four evidence records pinned to this SHA;
the observed CI run's headSha is this same SHA.

### Environment

Local dev workstation (macOS Darwin 25.5.0, arm64, go1.26.5), real Postgres 16 + MinIO via
compose, golangci-lint v2.11.4 pinned; **concurrent sibling load present** (noted per record);
CI timing from hosted GitHub Actions runners.

### Reviewer

Unassigned — conductor review gate pending (story not self-marked accepted).

### Findings

1. Lint drift vs MATRIX CS-23, explicitly flagged per-analyzer in EV-W00-E02-S001-002 (incl.
   security-relevant gosec scoping drift and new un-triaged G204/G301/G306 classes) — candidate
   findings for FBL-05/FBL-07 disposition, not resolved here per story scope.
2. MATRIX names only 18 of its claimed 25 analyzers — DEV-W00-E02-S001-001.
3. Run-1 lint anomaly (11 nondeterministic staticcheck SA5011 hits under load) — disclosed and
   preserved, not absorbed.
4. No bench-budget or entry-count drift; no CI-shape drift.

### Retest status

Not required — no failed capture. (The lint run-1 anomaly was re-run twice cache-cleaned within
the same capture procedure; both re-runs agree at 991 issues.)

### Final conclusion

W00-E02-S001's four quantitative baselines exist as commit-pinned, field-complete evidence
records at `0a31186cada5c275a588c74081cf977adf346e61`; every AC has an actual-result row; no
production file was modified. Story ready for conductor review.
