---
id: DEV-W00-E01-S001
type: deviation-record
parent_story: W00-E01-S001
status: executed
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations record — W00-E01-S001

Per mandate §8.9. Story executed 2026-07-13 at commit
`0a31186cada5c275a588c74081cf977adf346e61`. Three deviations from `plan.md`/`story.md` occurred;
none was silently absorbed. The approved plan text was not rewritten.

## DEV-01 — planned `-run TestKernelRules` pattern matches no tests (T003)

- **What deviated:** `plan.md`/`story.md` name `go test ./kernel/... -run TestKernelRules -race`
  as the command covering `kernel/kernel_rules_test.go`. Actually running it yields
  "no tests to run" in every package — no test function is named `TestKernelRules*`.
- **Reason:** the real test functions are `TestIntegrationRulesResolverOrgAncestry` and
  `TestIntegrationRulesResolverOrgAncestryWithAuthzCache` (`kernel/kernel_rules_test.go:24/39`).
  The plan anticipated exactly this ("or the equivalent covering `kernel_rules_test.go`" — listed
  under "Unresolved questions").
- **Actual command:** `go test -v ./kernel/ -run 'TestIntegrationRulesResolverOrgAncestry' -race -count=1`.
- **Impact:** none on evidence validity — both tests in the named file executed and passed with
  `-race` against a live Postgres.
- **Approval:** pre-authorized by the plan's own "or equivalent" clause; recorded here per
  mandate §2.6 rather than rewriting the plan.
- **Compensating control:** the exact command used is pinned in `EV-W00-E01-S001-03` and in the
  log header, so the record is reproducible.

## DEV-02 — AC-W00-E01-S001-04's grep clause fails as literally worded (T004)

- **What deviated:** the AC expects
  `grep -rn "RunAPI\|RunWorker\|RunMigrate" README.md docs/blueprint/` to return **zero** hits.
  It returned **7 hits**, all in `docs/blueprint/` (04:15,37-39; 06:207; 10:94; 12:171).
- **Analysis (why this is an AC-scoping artifact, not a code/doc regression):**
  (1) `README.md` and blueprint 11 — the two files the executed AR-05 T1 fix actually changed,
  per `premier-framework-implementation-plan.md` §AR-05 T1 and the final review §C rows 29/31 —
  have zero hits; (2) no `RunAPI`/`RunWorker`/`RunMigrate` function exists anywhere in Go source;
  (3) `git grep` at the fix commit `345e4ce` shows the byte-identical 7-hit set, so nothing has
  drifted since the reviewed fix; (4) the remaining hits are unlabeled future-state design prose,
  whose labeling is AR-05 **T5** — explicitly planned, not executed, tracked at `W06-E04-S002`.
- **Impact:** AC-W00-E01-S001-04 is recorded **fail** in `verification.md`;
  `EV-W00-E01-S001-04` is preserved with status `failed` (not retried, not deleted). The story
  cannot self-declare all-ACs-pass and is handed to the conductor at `ready-for-review`.
- **Approval:** conductor, 2026-07-13 — AC-04 re-scoped to executed T1/T2 slice (README +
  blueprint 11 + Context diff, all clean); the 7 future-state blueprint hits routed to AR-05 T5
  (W06-E04-S002); see impl/tracking/deviation-register.md row DEV-W00-E01-S001-002.
- **Compensating control:** the full grep output, the code-absence proof, and the `345e4ce`
  cross-check are all captured in `evidence/tests/ar05-doc-drift.log` so the adjudication can be
  made from evidence alone.

## DEV-03 — SEC-02 test-file locations differ from story.md's named list (T001)

- **What deviated:** `story.md` names `runtime_extra_test.go`, `runtime_lifecycle_test.go`,
  `runtime_test.go`, and `testkit/workflowsim_cov_test.go` as the SEC-02 coverage set. In fact
  (a) the nil-deps panic assertion lives in `kernel/workflow/internal_extra_test.go:207`
  (`TestNewRuntimePanicsOnNilDeps`), a file not on the list; (b) `testkit/workflowsim_cov_test.go`
  is in package `testkit` and is therefore **not** executed by `go test ./kernel/workflow/...` —
  it was exercised by T002's full-suite run (`go test ./...`, green) instead.
- **Impact:** none on evidence validity — every named assertion was located, executed, and passed;
  the coverage claim is now pinned to the correct files.
- **Approval:** recorded for reviewer awareness; no plan rewrite.
- **Compensating control:** exact test names + file:line are cited in T001's verification record
  and in `verification.md`.

No other deviation occurred. No production file (Go source, Makefile, configs, docs outside
`impl/`) was modified; all writes are inside this story directory.
