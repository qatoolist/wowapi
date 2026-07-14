---
id: W01-E04-S001-EVIDENCE-INDEX
type: evidence-index
parent_story: W01-E04-S001
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E04-S001 — Evidence index

Per mandate §10. Structure adaptation per `governance/naming-conventions.md` "Adaptation 2": category
subdirectories under `evidence/` (e.g. `tests/`, `logs/`) are created on first real content, not
pre-populated empty. All five evidence items are produced. Evidence paths follow the naming given in
this story's own governing instruction: `DX-01/t1-flag-verify.json` through `DX-01/t5-e2e-temp-dir.json`,
and `DX-02/w0-t2-verb-fix.json`-equivalent naming.

| Evidence ID | Type | Task | Acceptance criteria proven | Execution command | Path | Commit SHA | Result | Status |
|---|---|---|---|---|---|---|---|---|
| EV-W01-E04-S001-001 | functional-test report (fail-first pairs, all resolution paths) | W01-E04-S001-T001 | AC-W01-E04-S001-01 | `go test ./internal/cli/ -run 'TestInitDevel\|TestInitFramework\|TestInitLocalFramework\|TestInitBoth\|TestInitStamped\|TestGoResolveModuleVersion' -count=1 -v`, plus manual CLI demos (go run, dirty-tree go build, --local-framework) | `DX-01/t1-flag-verify.json` (+ logs `DX-01/t1-t4-prefix-failfirst.log`, `DX-01/t1-t4-postfix.log`, `DX-01/t1-t4-tests-postfix.log`) — one consolidated report covers T1-T4 (shared suite + fail-first capture), superseding the anticipated per-sub-task filenames | 05dce5c8a548f7dce3222637ab2c82024236a2a0 (fix uncommitted on top; conductor commits) | PASS post-fix (all paths fail closed pre-write on invalid input, verified value on valid input); pre-fix baseline captured BOTH defect shapes (`v0.0.0` from go run AND the SF-7 `v…+dirty` stamp from go build) succeeding silently | resolved (failed pre-fix runs preserved) |
| EV-W01-E04-S001-002 | functional-test report (harness run log, both CLI paths) | W01-E04-S001-T002 | AC-W01-E04-S001-02 | `go test ./internal/cli/ -run 'TestE2EScaffold' -count=1 -v` (hermetic: file:// proxies, no network) | `DX-01/t5-e2e-temp-dir.json` (+ logs `DX-01/t5-e2e-both-paths.log`, `DX-01/pkg-internal-cli-full-3.log`) | 05dce5c8a548f7dce3222637ab2c82024236a2a0 (harness uncommitted on top; conductor commits) | PASS — all five pipeline steps (init, tidy, download, build, boot smoke) succeeded for BOTH the source-built and released CLI paths; source path also proves flag-less devel init fails closed pre-write | resolved |
| EV-W01-E04-S001-003 | unit-test report (fail-before/pass-after pair) | W01-E04-S001-T003 | AC-W01-E04-S001-03 | `go test ./internal/cli/ -run TestGenCRUDPermissionKeys -count=1 -v` | `DX-02/w0-t2-verb-fix.json` (+ logs `DX-02/t003-permkeys-*.log`) | 05dce5c8a548f7dce3222637ab2c82024236a2a0 (fix uncommitted on top; conductor commits) | PASS post-fix; pre-fix baseline PASSED on the buggy string, proving the RISK-W01-005 test-lock | resolved |
| EV-W01-E04-S001-004 | functional-test report (fail-before/pass-after pair) | W01-E04-S001-T004 | AC-W01-E04-S001-04 | `go test ./internal/cli/ -run TestGenCRUDOutputBoots -count=1 -v` (pre-T003 and post-T003 template state) | `DX-02/w0-t2-boots-test.json` (+ logs `DX-02/t004-boots-*.log`, `DX-02/pkg-internal-cli-full.log`) | 05dce5c8a548f7dce3222637ab2c82024236a2a0 (test + fix uncommitted on top; conductor commits) | PASS post-T003; pre-T003 run FAILED with the exact `kernel/authz/registry.go:88-90` closed-verb-set rejection (`... not in the closed verb set: widgets.widget.delete`), captured verbatim | resolved (failed pre-fix run preserved) |
| EV-W01-E04-S001-005 | functional-test report (fail-before/pass-after pair) | W01-E04-S001-T005 (scope addition, DEV-W01-E04-S001-03) | AC-W01-E04-S001-05 (added) | `go test ./internal/cli/ -run TestInitScaffoldConfigValidates -count=1 -v` | `DX-02/scaffold-config-validate-fix.json` (+ logs `DX-02/t005-scaffold-config-validate-*.log`, `DX-02/pkg-internal-cli-full-2.log`) | 05dce5c8a548f7dce3222637ab2c82024236a2a0 (fix uncommitted on top; conductor commits) | PASS post-fix; pre-fix run FAILED with `i18n.default_locale/go_bundles/locales_dir/supported_locales: unknown key`, captured verbatim | resolved (failed pre-fix run preserved) |

Evidence status vocabulary (per mandate §10): `not yet produced` is this programme's pre-execution
state, outside the mandate's own failed/superseded/retested/resolved/accepted-exception vocabulary,
which applies only once an evidence item has actually been produced at least once.

## Notes

- Every evidence item in this table is expected to include, once produced, a fail-before/pass-after (or
  fail-closed-vs-succeeds) pair rather than a single run — this is a deliberate consequence of this
  story's fail-first testing strategy (see `plan.md` "Testing strategy"), not an incidental choice.
- EV-W01-E04-S001-004's pre-T003 ("fail-before") run must capture the exact failure message and confirm
  it matches the closed-verb-set rejection at `kernel/authz/registry.go:88-90` specifically — a
  generically-failing pre-fix run would not be sufficient evidence that T003's fix (rather than some
  unrelated harness issue) is what makes the post-fix run pass (see
  `tasks/task-004-generator-output-boots-test.md` "Risks").
- EV-W01-E04-S001-004's pre-T003 requirement is satisfied: the fail-first log captures the verbatim
  rejection `permission action %q is not in the closed verb set: widgets.widget.delete` — see
  `DX-02/t004-boots-prefix-failfirst.log` and the run entries inside `DX-02/w0-t2-boots-test.json`.
