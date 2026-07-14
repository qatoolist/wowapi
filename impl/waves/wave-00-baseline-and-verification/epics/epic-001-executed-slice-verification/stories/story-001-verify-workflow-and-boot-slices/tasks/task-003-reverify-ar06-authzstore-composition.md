---
id: W00-E01-S001-T003
type: task
title: Re-verify AR-06 T1 authzStore composition (no duplicate authz.NewStore() call)
status: done
parent_story: W00-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria: [AC-W00-E01-S001-03]
artifacts: [ART-W00-E01-S001-003]
evidence: [EV-W00-E01-S001-03]
---

# W00-E01-S001-T003 — Re-verify AR-06 T1 authzStore composition

## Task Definition

### Task objective

Re-run `go test ./kernel/authz/... -race` and `go test ./kernel/... -run TestKernelRules -race`
(or the equivalent command covering `kernel_rules_test.go`) at the current repository HEAD, and
confirm that `kernel/kernel.go`'s `orgAncestry` closure (previously cited at lines 252-254) still
uses the composed `authzStore` instance rather than invoking `authz.NewStore()` a second time.
Register the result as a mandate-§10 evidence record.

### Parent story

W00-E01-S001 — Verify workflow and boot composition slices at current HEAD.

### Owner

unassigned.

### Status

done.

### Dependencies

None. This task is parallel-safe with W00-E01-S001-T001 and -T002 (disjoint package/file scope,
though it shares the `kernel` top-level package boundary with T001's `kernel/workflow` — no file
overlap).

### Detailed work

1. Confirm `kernel/kernel.go`'s `orgAncestry` closure still uses the single, composed `authzStore`
   instance built during kernel composition, and does not call `authz.NewStore()` a second time
   (the AR-06 T1 fix). Note: the previously cited line numbers (252-254) may have shifted since the
   review documents were written — confirm the closure's current location by reading the file, not
   by assuming the old line numbers still apply.
2. Confirm `kernel/authz/caching_internal_test.go` and `kernel/kernel_rules_test.go` still exist and
   contain a sentinel-store-injection test — a test that injects a distinguishable ("sentinel")
   store instance at composition time and asserts the same instance (not a freshly constructed one)
   is what `orgAncestry` observes.
3. Identify the exact test function name/`-run` pattern for the sentinel-store-injection assertion
   (not yet confirmed at plan time — see `plan.md` "Unresolved questions").
4. Run `go test ./kernel/authz/... -race` and capture full output.
5. Run `go test ./kernel/... -run TestKernelRules -race` (or the identified equivalent covering
   `kernel_rules_test.go`) and capture full output.
6. Inspect both outputs for exit code 0, no race warnings, and presence/pass of the
   sentinel-store-injection assertion.
7. If either test package requires a live Postgres instance, confirm test infrastructure
   availability before treating any failure as a genuine regression (RISK-W00-002).
8. Register the result as evidence per `evidence-policy.md`'s required-field list, citing
   `AR-06/sentinel_store_injection_test.go` as the evidence artifact reference.

### Expected files or components affected

None — read-only verification task. `kernel/kernel.go`, `kernel/authz/caching_internal_test.go`,
and `kernel/kernel_rules_test.go` are inspected and re-tested, not modified, unless a regression is
found (see "Rollback or recovery considerations").

### Expected output

Two test-execution logs (`go test -v ./kernel/authz/... -race` and
`go test -v ./kernel/... -run TestKernelRules -race` or equivalent), both showing exit code 0, no
race warnings, and the sentinel-store-injection assertion passing.

### Required artifacts

Test-execution log artifacts (one per command, or combined), registered in the story's
`artifacts/index.md` (lifecycle stage: post-implementation).

### Required evidence

One evidence record, planned ID `EV-W00-E01-S001-03`, evidence type "test-execution log
(race-detector)," referencing `AR-06/sentinel_store_injection_test.go` output as the evidence
artifact name.

### Related acceptance criteria

AC-W00-E01-S001-03.

### Completion criteria

This task is `done` when: both commands have actually been executed; the result (`pass` or
`failed`) is recorded in this task's `verification.md`; the evidence record is registered in the
story's `evidence/index.md` with all required fields; and, if the result is `failed`, a follow-up
remediation task has been opened under `W05-E04-S001` per `requirement-inventory.md`'s AR-06 target
(this task is not marked `done` while a regression is unresolved and unacknowledged).

### Verification method

`go test ./kernel/authz/... -race` and `go test ./kernel/... -run TestKernelRules -race` (or
equivalent), inspected for exit code 0 on both, no race warnings, and presence/pass of the
sentinel-store-injection assertion. See this task's own `verification.md` for the full
planned-procedure row.

### Risks

- RISK-W00-001 (inherited) — AR-06 T1's single-store-composition fix could have regressed since the
  reviewed SHA (e.g. a later unrelated change to `kernel/kernel.go` silently reintroducing a second
  `authz.NewStore()` call).
- RISK-W00-002 (inherited, conditional) — if `kernel/authz`/`kernel` test packages require a live
  Postgres instance and the environment lacks one, a false-negative failure could be mistaken for a
  genuine regression.

### Rollback or recovery considerations

If a regression is found (the `orgAncestry` closure calls `authz.NewStore()` a second time instead
of using the composed instance), this task does not fix it. Instead: record a `failed`-status
evidence record (preserved, never deleted); open a new remediation task under `W05-E04-S001`
(AR-06's canonical target story per `requirement-inventory.md`, noting AR-06 T2 lint and T3 audit
are already planned there); do not silently mark this task or its parent story `done`/`accepted`
while the regression is open.

## Implementation Record

Per mandate §8.7. Executed 2026-07-13. "Implementation" here means running the verification
commands and registering evidence, not writing code.

### What was actually implemented

Confirmed by direct source inspection that `kernel/kernel.go` composes exactly one
`authz.NewStore()` (line 230), optionally decorated by `authz.NewCachingStore` (lines 232-235),
and that the `orgAncestry` closure — now at lines 254-256, shifted from the previously cited
252-254 — delegates to that composed `authzStore` instance
(`return authzStore.OrgAncestors(ctx, db, orgID)`; source comment cites AR-06 T1). No second
`authz.NewStore()` call exists on this path. Identified the sentinel-injection test:
`TestCachingStoreOrgAncestorsRoutesToComposedInner` (`kernel/authz/caching_internal_test.go:99`),
plus the kernel-level integration tests `TestIntegrationRulesResolverOrgAncestry` and
`TestIntegrationRulesResolverOrgAncestryWithAuthzCache` (`kernel/kernel_rules_test.go:24/39`).
Ran `go test -v ./kernel/authz/... -race` and
`go test -v ./kernel/ -run 'TestIntegrationRulesResolverOrgAncestry' -race -count=1` at commit
`0a31186cada5c275a588c74081cf977adf346e61`. Both exit 0 — evidence EV-W00-E01-S001-03.

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

None — verification-only; no commit was produced by this task.

### Pull requests

None.

### Implementation dates

2026-07-13 (single session).

### Technical debt introduced

None expected.

### Known limitations

Point-in-time re-verification; the durable guard against a reintroduced second `authz.NewStore()` is AR-06 T2's lint rule (W05-E04-S001), not this task.

### Follow-up items

None.

### Relationship to the approved plan

Executed with one recorded deviation (story `deviations.md` DEV-01): the plan's suggested
`go test ./kernel/... -run TestKernelRules -race` matched **no tests** (every package reported
"no tests to run" — no test function is named `TestKernelRules*`). The plan explicitly allowed
"or the equivalent covering `kernel_rules_test.go`"; the equivalent actually run was
`go test -v ./kernel/ -run 'TestIntegrationRulesResolverOrgAncestry' -race -count=1`, which
executes both tests defined in `kernel/kernel_rules_test.go`. The `orgAncestry` closure's line
shift (252-254 → 254-256) was anticipated by the task definition and is noted, not a deviation.

## Verification Record

### Planned verification procedure

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E01-S001-03 | Re-run `go test ./kernel/authz/... -race` and `go test ./kernel/... -run TestKernelRules -race` (or equivalent, exact pattern to be confirmed during execution); inspect for exit code 0 and no race warnings; confirm via the sentinel-store-injection test that `kernel/kernel.go`'s `orgAncestry` closure uses the composed `authzStore` instance rather than a second `authz.NewStore()` call | Local or CI Go toolchain per `go.mod`; DB requirement for `kernel/authz`/`kernel` test packages to be confirmed during execution | Exit code 0 on both commands; all tests pass including the sentinel-store-injection assertion; no `-race` warnings | Test-execution log (race-detector log, `go test -v` output) | unassigned (framework architecture lead role) |

### Actual result

`go test -v ./kernel/authz/... -race`: exit 0, all tests PASS, no race warnings; sentinel test
`TestCachingStoreOrgAncestorsRoutesToComposedInner` PASS — injects a distinguishable counting
store as the composed inner instance, asserts `OrgAncestors` routes to it exactly once and that
an independently constructed second store observes zero calls.
`go test -v ./kernel/ -run 'TestIntegrationRulesResolverOrgAncestry' -race -count=1`: exit 0;
`TestIntegrationRulesResolverOrgAncestry` (0.23s) and
`TestIntegrationRulesResolverOrgAncestryWithAuthzCache` (0.14s) both PASS against local Postgres
(cache-off and cache-on composition paths); no race warnings.

### Pass or fail

pass.

### Evidence identifier

EV-W00-E01-S001-03 (`evidence/tests/ar06-authz-race.log`, sha256:b954cb0cbc1c15b0;
`evidence/tests/ar06-kernel-rules-race.log`, sha256:97441fa6cb69364c).

### Execution date

2026-07-13 (authz 12:14:12; kernel rules 12:15:13 local).

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### Environment

Local workstation, macOS 26.5.2 (Darwin 25.5.0) arm64; go1.26.5 darwin/arm64; local Postgres via
`make up` compose (`DATABASE_URL` set); concurrent load present (sibling W00 workers).

### Reviewer

unassigned — conductor review pending (worker self-review only; not self-marked accepted).

### Findings

AR-06 T1 single-instance `authzStore` composition intact at HEAD; no regression. `-run
TestKernelRules` naming in plan/story does not correspond to any real test function (DEV-01).

### Retest status

Not applicable — first execution under this programme; result pass.

### Final conclusion

AC-W00-E01-S001-03 satisfied. AR-06 executed slice (T1) re-proven at pinned HEAD.

## Deviations Record

One deviation, recorded at story level (`deviations.md` DEV-01): planned `-run TestKernelRules`
pattern matched no tests; the plan-sanctioned equivalent
(`-run 'TestIntegrationRulesResolverOrgAncestry'`) was used instead. No other deviation.
