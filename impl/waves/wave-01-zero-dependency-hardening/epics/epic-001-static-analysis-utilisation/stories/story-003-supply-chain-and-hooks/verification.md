---
id: VER-W01-E01-S003
type: verification-record
parent_story: W01-E01-S003
status: executed
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W01-E01-S003

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E01-S003-01 | Run the `go mod verify` step added to `ci.yml`, both directly (`go mod verify` on the shell) and as part of a full CI run | Local dev environment or CI, Go toolchain per `go.mod` | Exit code 0 against a clean module cache | CI execution record / command-execution log (evidence/) | W01ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 |
| AC-W01-E01-S003-02 | Trigger the license-scanning step (Trivy `license` scanner, or `go-licenses` if the choice is revised at implementation time) in CI | CI, Trivy pinned version per `security-scan.yml` (or `go-licenses` version, if revised) | Step executes and produces a license report or equivalent output; choice and rationale recorded in `implementation.md` | security-scan report (evidence/) | W01ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 |
| AC-W01-E01-S003-03 | Direct inspection of `ci.yml`'s `schedule:`/`cron:` trigger and job graph, plus an observed scheduled or manually-triggered run where feasible | CI (GitHub Actions), repository at implementation-time HEAD | Schedule confirmed present, genuinely nightly, and correctly reaching the fuzz-seed-corpus-replay step; `-fuzz=` coverage-guided-generation gap explicitly recorded as W07 (REL-04 T8 / PERF-06 T3/T4) scope, not silently closed or duplicated | CI execution record + confirmation/audit note (evidence/) | W01ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 |
| AC-W01-E01-S003-04 | Fail-before/pass-after demonstration of `.githooks/pre-push`'s DB-test-skip behavior, with and without `WOWAPI_REQUIRE_DB` set / DB reachable | Local dev environment, with and without a reachable local DB | Before fix: hook silently passes without a DB. After fix: hook fails loudly with an actionable message without a DB (unless explicitly configured to skip), and passes normally with a DB available | execution log (fail-before/pass-after pair) (evidence/) | W01ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 |

## Post-execution record

Executed 2026-07-13 by W01Lint. Revision: HEAD `0a31186cada5c275a588c74081cf977adf346e61` + this
story's uncommitted working change (diff-stat recorded in `implementation.md`; conductor owns the
wave commit). Environment: darwin/arm64 dev workstation, Go 1.26.5, golangci-lint v2.11.4,
actionlint (local), Trivy (local), Docker compose stack (`wowapi-postgres` on :5432) running.
Hook demonstrations ran in a pristine `git archive HEAD` copy (`/tmp/wowapi-head-w01lint`) because
sibling W01 workers were concurrently editing the shared working tree.

| AC | Actual result | Pass/fail | Evidence |
|---|---|---|---|
| AC-W01-E01-S003-01 | `go mod verify` step present in `ci.yml` unit job (fails job on failure — no continue-on-error); local execution: `all modules verified`, exit 0; `actionlint` clean over the edited workflow | **pass** (in-CI run log pending conductor push — explicit carry-forward note below) | EV-W01-E01-S003-001 → `evidence/logs/gomodverify-and-actionlint.log` |
| AC-W01-E01-S003-02 | `license` added to trivy `scanners:`; local Trivy license run against pristine HEAD enumerated 70 dependency licenses (all LOW: MIT/Apache-2.0/BSD-*); with the job's `severity: CRITICAL,HIGH` filter → 0 findings = clean forbidden/restricted signal; choice + rationale recorded in `implementation.md` | **pass** (same in-CI limitation note) | EV-W01-E01-S003-002 → `evidence/logs/trivy-license-local-report.txt` |
| AC-W01-E01-S003-03 | Schedule confirmed by file-chain inspection (`ci.yml:42-46` cron → `changes` fail-safe → `gate` test leg → `Makefile:324-326` seed replay) AND by an observed scheduled run (29229288699, 2026-07-13, event=schedule, success, log shows `-run "^Fuzz"` executing); `-fuzz=` gap explicitly recorded as W07 scope | **pass** | EV-W01-E01-S003-003 → `evidence/logs/nightly-fuzz-observed-run.log` + `artifacts/nightly-fuzz-confirmation.md` |
| AC-W01-E01-S003-04 | Fail-before: unmodified hook (restored from `git show HEAD:`), no DSN → hook prints `pre-push: OK`, exit 0, while kernel/database Integration tests all SKIP (skip-proof appended to the same log). After fix, unreachable DB (`localhost:59999`) → tests FAIL fast, hook prints actionable guidance block, exit 1. After fix, DB available → DB tests genuinely RAN (kernel/database 15.8s vs 7.8s skip-mode; RLS integration test passed), exit 0. Opt-out `WOWAPI_PREPUSH_SKIP_DB=1` → loud WARNING + pass. `pre-commit` untouched (`git diff --quiet` clean) | **pass** | EV-W01-E01-S003-004 → `evidence/logs/prepush-fail-before-silent-pass.log`, `prepush-after-nodb-loud-fail.log`, `prepush-after-withdb-pass.log`, `prepush-after-optout-pass.log` |

### Findings

1. The first unreachable-DB hook run appeared to hang (>30 min) — root-caused to a cold-cache
   compile storm in the fresh `/tmp` tree, not a hook defect: the reproducible rerun completed in
   <240 s with exit 1 (`prepush-after-nodb-loud-fail.log`). Recorded here so the anomaly is not
   silently dropped.
2. Observed scheduled runs start ~3 h after the 03:17 UTC cron mark (06:14Z/06:32Z) — GitHub's
   documented best-effort schedule delay; cadence is still genuinely nightly (consecutive daily
   runs observed).

### Carry-forward note (evidence-policy §"Revision-pinning")

AC-01/AC-02 cite local execution of the exact commands the new CI steps run, plus a clean
`actionlint` pass over the edited workflows. The in-CI run log cannot exist until the conductor
pushes the wave commit (workers are prohibited from pushing). Nothing material changes between the
local and CI execution of `go mod verify`/Trivy (same pinned tool invocations); the wave gate's
first CI run supersedes these records as `retested` evidence.

### Retest status

Not required pre-acceptance except the AC-01/AC-02 CI-run supersession above.

### Final conclusion

All four acceptance criteria verified; story ready for independent review (mandate §14).

> Carry-forward: first in-CI executions of go-mod-verify step and license scanner occur on next push (wave uncommitted at review time); register that CI run as retested evidence at next wave gate.
