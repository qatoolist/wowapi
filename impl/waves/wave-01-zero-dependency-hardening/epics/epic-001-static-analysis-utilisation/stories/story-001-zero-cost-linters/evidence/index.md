---
id: W01-E01-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W01-E01-S001
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E01-S001 — Evidence index

Per mandate §10. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
category subdirectories under `evidence/` (e.g. `static-analysis/`, `tests/`) are created on first
real content, not pre-populated empty. All entries below were produced 2026-07-13 (W01Lint).

**Phase-1 fresh-run baseline (produced 2026-07-13, W01Lint)** — the story-local fail-first re-run
required by `plan.md`, captured at HEAD `0a31186cada5c275a588c74081cf977adf346e61` (clean tracked
tree at run time), golangci-lint v2.11.4, W00 throwaway matrix config:
`static-analysis/zero-cost-and-nearzero-enumeration.txt` (raw run shared at
`../../story-002-judged-linter-set/evidence/static-analysis/triage-fresh-run-raw.{json,txt}`).
Result: all seven zero-cost analyzers at **0 hits** (cited claim CONFIRMED); `copyloopvar` 1 prod +
6 test hits (named prod site confirmed); `noctx` **drift** — the 2 named CLI `exec.Command` sites are
NOT reported by noctx v2.11.4 (it flags net/http request construction only): 146 hits, 145 in
`_test.go`, 1 in `testkit/i18n.go:33`. Fail-before halves of EV-002/EV-003 are contained in this
baseline; pass-after halves land in Phase 2 with the fixes + enablement.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W01-E01-S001-001 | static-analysis report | W01-E01-S001-T001 | AC-W01-E01-S001-01 | `golangci-lint run --enable=sqlclosecheck,rowserrcheck,bodyclose,wastedassign,makezero,musttag,testifylint ./...` | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | all 7 at 0 hits (per-linter runs exit 0; full tree exit 0) (static-analysis/per-linter-enablement-pass-after.txt) | produced |
| EV-W01-E01-S001-002 | static-analysis report (fail-before/pass-after pair) | W01-E01-S001-T002 | AC-W01-E01-S001-02 | `golangci-lint run --enable=noctx ./internal/cli/...` | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | sites use CommandContext; gosec G204 fail-before + code diff; noctx run exit 0 (static-analysis/noctx-copyloopvar-site-fix.diff) | produced (mechanism substituted per DEV-001) |
| EV-W01-E01-S001-003 | static-analysis report (fail-before/pass-after pair) | W01-E01-S001-T003 | AC-W01-E01-S001-03 | `golangci-lint run --enable=copyloopvar ./app/...` | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | fail-before in both triages; copyloopvar run exit 0 after fix (static-analysis/ files as above) | produced |
| EV-W01-E01-S001-004 | unit-test report | W01-E01-S001-T004 | AC-W01-E01-S001-04 | `go test ./kernel/config/... -run TestMaxConn -v` | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | config+database tests pass incl. TestPoolLifetimeKeysValidate / TestIntegrationPoolLifetimeConfigWiring (tests/touched-package-test-sweep.log) | produced |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.
