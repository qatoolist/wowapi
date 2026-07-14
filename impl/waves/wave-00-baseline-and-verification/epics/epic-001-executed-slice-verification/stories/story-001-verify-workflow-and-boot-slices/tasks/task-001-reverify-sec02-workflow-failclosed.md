---
id: W00-E01-S001-T001
type: task
title: Re-verify SEC-02 workflow fail-closed behavior
status: done
parent_story: W00-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria: [AC-W00-E01-S001-01]
artifacts: [ART-W00-E01-S001-001]
evidence: [EV-W00-E01-S001-01]
---

# W00-E01-S001-T001 — Re-verify SEC-02 workflow fail-closed behavior

## Task Definition

### Task objective

Re-run `go test ./kernel/workflow/... -race` at the current repository HEAD and confirm that
`kernel/workflow/runtime.go` still exhibits SEC-02's T1-T3 fail-closed behavior: `NewRuntime`
panics if `ev == nil`, and `Override` unconditionally checks authz (no `if rt.authz != nil` skip
path remains). Register the result as a mandate-§10 evidence record.

### Parent story

W00-E01-S001 — Verify workflow and boot composition slices at current HEAD.

### Owner

unassigned.

### Status

done.

### Dependencies

None. This task is parallel-safe with W00-E01-S001-T002 and -T003 (disjoint package/file scope).

### Detailed work

1. Confirm `kernel/workflow/runtime.go` still contains: (a) a nil-check on `ev` in `NewRuntime` that
   panics when `ev == nil`; (b) an `Override` implementation that unconditionally performs an authz
   check, with no conditional skip when `rt.authz` is nil.
2. Confirm the following test files still exist in `kernel/workflow/` and reference SEC-02's
   behavior: `runtime_extra_test.go`, `runtime_lifecycle_test.go`, `runtime_test.go`.
3. Confirm `testkit/workflowsim_cov_test.go` still exists and exercises the runtime under test.
4. Run `go test ./kernel/workflow/... -race` and capture full output.
5. Inspect the output for: exit code 0, no `-race` warnings, and the specific presence/pass of the
   nil-`ev` panic assertion and the fail-closed `Override` assertion among the suite's test names.
6. If `testkit/workflowsim_cov_test.go` requires a live Postgres instance, confirm test
   infrastructure availability (`make ci-container` / `docker compose`) before treating any failure
   as a genuine regression (RISK-W00-002).
7. Register the result as evidence per `evidence-policy.md`'s required-field list.

### Expected files or components affected

None — this is a read-only verification task. `kernel/workflow/runtime.go` and its test files are
inspected and re-tested, not modified, unless a regression is found (see "Rollback or recovery
considerations" below).

### Expected output

A test-execution log (full `go test -v ./kernel/workflow/... -race` output) showing exit code 0,
no race warnings, and the nil-`ev` panic and fail-closed-`Override` assertions present and passing.

### Required artifacts

Test-execution log artifact, registered in the story's `artifacts/index.md` (lifecycle stage:
post-implementation, since this task's "implementation" is running verification).

### Required evidence

One evidence record, planned ID `EV-W00-E01-S001-01`, evidence type "test-execution log
(race-detector)," referencing the three named test-file artifact names conventionally as
`SEC-02/mandatory-evaluator-tests.md`, `SEC-02/override-fail-closed-tests.md`, and
`SEC-02/test-constructor-migration.md` (evidence-path-convention names for this finding-slice, per
the story's evidence-path convention — these name the sub-aspects of SEC-02 T1-T3 this task's
single test-execution-log evidence record covers, not three separate evidence records).

### Related acceptance criteria

AC-W00-E01-S001-01.

### Completion criteria

This task is `done` when: the test command has actually been executed (not merely planned); the
result (`pass` or `failed`) is recorded in this task's `verification.md`; the evidence record is
registered in the story's `evidence/index.md` with all required fields; and, if the result is
`failed`, a follow-up remediation task has been opened under `W03-E05-S001` per
`requirement-inventory.md`'s SEC-02 target (this task itself is not marked `done` while a regression
is unresolved and unacknowledged).

### Verification method

`go test ./kernel/workflow/... -race`, inspected for exit code 0, absence of race warnings, and
presence/pass of the specific nil-`ev` and fail-closed-`Override` assertions. See this task's own
`verification.md` for the full planned-procedure row.

### Risks

- RISK-W00-001 (inherited) — SEC-02's fail-closed fix could have regressed since the reviewed SHA;
  high severity given this is a security-relevant control.
- RISK-W00-002 (inherited, conditional) — if `testkit/workflowsim_cov_test.go` requires a live
  Postgres instance and the environment lacks one, a false-negative failure could be mistaken for a
  genuine regression.

### Rollback or recovery considerations

If a regression is found (the nil-`ev` panic or the fail-closed `Override` check is missing or
weakened), this task does not fix it. Instead: record a `failed`-status evidence record (preserved,
never deleted, per `evidence-policy.md`); open a new remediation task under `W03-E05-S001` (SEC-02's
canonical target story per `requirement-inventory.md`); do not silently mark this task or its parent
story `done`/`accepted` while the regression is open.

## Implementation Record

Per mandate §8.7. Executed 2026-07-13. "Implementation" here means running the verification
command and registering evidence, not writing code.

### What was actually implemented

Re-ran `go test -v ./kernel/workflow/... -race` at commit
`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`). Before running, confirmed by direct
source inspection: (a) `kernel/workflow/runtime.go:88-92` — `NewRuntime` panics when any of
`txm/reg/ev/ob/idgen` is nil (including `ev == nil`); (b) `runtime.go:285-306` — `Override`
performs an **unconditional** `rt.authz.Evaluate(...)` (source comment: "unconditional: there is
no construction path that can bypass it"); no `if rt.authz != nil` skip exists on the Override
path (the nil-guard at `runtime.go:669` belongs to the separate task-decide `authorize` helper's
optional secondary role gate, not to Override). Result: pass — evidence EV-W00-E01-S001-01.

### Components changed

None expected.

### Files changed

None expected.

### Interfaces introduced or changed

None expected.

### Configuration changes

None expected.

### Schema or migration changes

None expected.

### Security changes

None expected.

### Observability changes

None expected.

### Tests added or modified

None — existing tests re-run, not modified.

### Commits

None — verification-only; no commit was produced by this task (evidence logs are registered inside the story directory).

### Pull requests

None.

### Implementation dates

2026-07-13 (single session).

### Technical debt introduced

None expected.

### Known limitations

Point-in-time re-verification only; ongoing regression protection is AR-06/SEC-02 later-wave scope (see story.md "Residual-risk expectations").

### Follow-up items

None.

### Relationship to the approved plan

Execution matched the planned command exactly (`go test ./kernel/workflow/... -race`, run with
`-v` for per-test visibility). One planning-time assumption corrected: the nil-deps panic test
lives in `kernel/workflow/internal_extra_test.go` (not one of the three files named in story.md);
recorded as DEV-03 in the story's `deviations.md`. `testkit/workflowsim_cov_test.go` is outside
the `./kernel/workflow/...` pattern (package `testkit`) and was exercised by T002's full-suite
run instead — also DEV-03.

## Verification Record

### Planned verification procedure

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E01-S001-01 | Re-run `go test ./kernel/workflow/... -race`; inspect for exit code 0 and no race warnings; confirm the `NewRuntime` nil-`ev` panic test and the `Override` fail-closed (unconditional authz check) test are present among, and passing within, the suite | Local or CI Go toolchain per `go.mod`; no external DB required unless `testkit/workflowsim_cov_test.go` needs it — confirm during execution | Exit code 0; all tests pass including the specific nil-`ev` and fail-closed-override assertions; no `-race` warnings | Test-execution log / `go test -v` output | unassigned (framework architecture lead role) |

### Actual result

Exit 0; all tests pass; no `-race` warnings. Named assertions confirmed present and passing:
`TestNewRuntimePanicsOnNilDeps` (`internal_extra_test.go:207`, nil-deps/nil-`ev` panic),
`TestIntegrationOverrideAuthzGate` (`runtime_extra_test.go:440`),
`TestIntegrationOverrideFailsClosedWithoutPermission` (`runtime_extra_test.go:493`, denies with
KindForbidden), `TestIntegrationWorkflowOverride` (`runtime_lifecycle_test.go:171`). DB-backed
integration tests executed (did not skip) against local Postgres via `make up` compose.

### Pass or fail

pass.

### Evidence identifier

EV-W00-E01-S001-01 (`evidence/tests/sec02-workflow-race.log`, sha256:0a17e85ea35ecdce).

### Execution date

2026-07-13 (12:13:43–12:13:54 local).

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### Environment

Local workstation, macOS 26.5.2 (Darwin 25.5.0) arm64; go1.26.5 darwin/arm64; local Postgres via
`make up` compose (`DATABASE_URL` set); concurrent load present (sibling W00 workers).

### Reviewer

unassigned — conductor review pending (worker self-review only; not self-marked accepted).

### Findings

SEC-02 T1–T3 fail-closed behavior intact at HEAD; no regression. Minor file-location correction
(nil-deps test in `internal_extra_test.go`) recorded as DEV-03.

### Retest status

Not applicable — first execution under this programme; result pass.

### Final conclusion

AC-W00-E01-S001-01 satisfied. SEC-02 executed slice (T1–T3) re-proven at pinned HEAD.

## Deviations Record

Two minor deviations, recorded at story level (`deviations.md` DEV-03): the nil-deps panic test
file differs from the story's named list, and `testkit/workflowsim_cov_test.go` is covered by the
full-suite run rather than the `./kernel/workflow/...` pattern. Command and scope otherwise
matched `plan.md` exactly.
