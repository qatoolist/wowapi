---
id: W01-E01-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W01-E01-S002
status: produced
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E01-S002 — Evidence index

Per mandate §10. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2":
category subdirectories under `evidence/` (e.g. `static-analysis/`, `tests/`) are created on first
real content, not pre-populated empty. All entries below were produced 2026-07-13 (W01Lint).

Historical citation note: the source review material (the earlier architecture-review pass that
produced FBL-07) uses the evidence path convention `evidence/premier/FBL-07/` for its own cited
counts (the "38 hits" gosec figure and the named sites in `story.md` "Current-state assessment"). That
path is external provenance for the *historical* citation only — it is not a location this programme
writes to. This story's own verification evidence, produced fresh at this story's execution commit,
is registered exclusively in this file and stored under this story's own `evidence/` subdirectories
once produced.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|
| EV-W01-E01-S002-001 | static-analysis report (fresh-run baseline, all five judged analyzers) | W01-E01-S002-T001 (fresh-run step, shared baseline for T001-T006) | AC-W01-E01-S002-01 | `golangci-lint run --enable=gosec,errorlint,exhaustive,forcetypeassert,usestdlibvars ./...` | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | fresh-run baselines at HEAD and Phase-2 state (static-analysis/judged-set-enumeration.txt, phase2-fail-before-enumeration.txt, triage-fresh-run-*.json) | produced |
| EV-W01-E01-S002-002 | static-analysis report (fail-before/pass-after pair) | W01-E01-S002-T001 | AC-W01-E01-S002-02 | `golangci-lint run --enable=gosec ./kernel/auth/...` | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | G704 x2 annotated w/ SEC-06; gosec run exit 0 (per-linter log) | produced |
| EV-W01-E01-S002-003 | static-analysis report + per-site triage record | W01-E01-S002-T002 | AC-W01-E01-S002-02 | `golangci-lint run --enable=gosec ./...` (filtered to G115 hits) | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | G115: 7 sites enumerated+dispositioned (implementation.md table); cursor fix fail-first proven at HEAD | produced |
| EV-W01-E01-S002-004 | static-analysis report (fail-before/pass-after pair) | W01-E01-S002-T003 | AC-W01-E01-S002-02 | `golangci-lint run --enable=gosec ./...` (filtered to G304 hit) | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | G304 sites annotated; gosec run exit 0 | produced |
| EV-W01-E01-S002-005 | static-analysis report (fail-before/pass-after pair) | W01-E01-S002-T004 | AC-W01-E01-S002-03 | `golangci-lint run --enable=errorlint ./kernel/httpx/...` | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | errorlint fail-before both triages; exit 0 after fix | produced |
| EV-W01-E01-S002-006 | static-analysis report (fail-before/pass-after pair) + review note | W01-E01-S002-T005 | AC-W01-E01-S002-04 | `golangci-lint run --enable=exhaustive ./kernel/workflow/...` | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | exhaustive 4 sites annotated w/ fail-closed comments; exit 0 after | produced |
| EV-W01-E01-S002-007 | static-analysis report (fail-before/pass-after pair) + unit-test report | W01-E01-S002-T006 | AC-W01-E01-S002-05 | `golangci-lint run --enable=forcetypeassert ./kernel/auth/... ./kernel/config/...` && `go test ./kernel/auth/... ./kernel/config/... -run TestForceTypeAssert -v` | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | forcetypeassert 3 sites fixed comma-ok; exit 0 after; package suites pass (see verification.md finding 3 on false-path tests) | produced |
| EV-W01-E01-S002-008 | static-analysis report (fail-before/pass-after pair) | W01-E01-S002-T007 | AC-W01-E01-S002-06 | `golangci-lint run --enable=usestdlibvars ./...` | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | usestdlibvars 9 sites fixed; exit 0 after | produced |
| EV-W01-E01-S002-009 | static-analysis report + review note (nilerr annotation, wrapcheck/revive absence) | W01-E01-S002-T007 | AC-W01-E01-S002-07 | Manual review of `kernel/policy/policy.go:166` annotation and final `.golangci.yml` `enable:` list | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | nilerr annotation in policy.go; wrapcheck/revive absence machine-checked (static-analysis/wrapcheck-revive-absence.txt) | produced |
| EV-W01-E01-S002-010 | static-analysis report (final combined confirmation run) | W01-E01-S002-T007 | AC-W01-E01-S002-01 | `golangci-lint run ./...` (full module tree, updated `.golangci.yml`) | `0a31186cada5c275a588c74081cf977adf346e61` + wave diff | final full-tree run exit 0 (static-analysis/final-full-tree-lint-pass.txt) | produced |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.
