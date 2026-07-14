---
id: W00-E01-S001-T002
type: task
title: Re-verify AR-04 T1 boot-time unknown-namespace rejection
status: done
parent_story: W00-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria: [AC-W00-E01-S001-02]
artifacts: [ART-W00-E01-S001-002]
evidence: [EV-W00-E01-S001-02]
---

# W00-E01-S001-T002 — Re-verify AR-04 T1 boot-time unknown-namespace rejection

## Task Definition

### Task objective

Re-run `go test ./app/... -run Boot` (or the specific boot-namespace-rejection test) plus a full
`go test ./...` at the current repository HEAD, and confirm that `app/boot.go` still rejects
unknown `modules.<typo>` config namespaces at boot with a deterministic named error. Register the
result as a mandate-§10 evidence record.

### Parent story

W00-E01-S001 — Verify workflow and boot composition slices at current HEAD.

### Owner

unassigned.

### Status

done.

### Dependencies

None. This task is parallel-safe with W00-E01-S001-T001 and -T003 (disjoint package/file scope).

### Detailed work

1. Confirm `app/boot.go` still contains logic that rejects unknown `modules.<typo>` config
   namespaces at boot with a deterministic named error (not a silent no-op or generic panic).
2. Confirm `app/boot_extra_test.go` still exists and exercises this rejection behavior; identify the
   exact test function name/`-run` pattern that isolates it (not yet confirmed at plan time — see
   `plan.md` "Unresolved questions").
3. Run `go test ./app/... -run Boot` (or the identified specific test).
4. Run a full `go test ./...` to confirm no unrelated regression accompanies the boot-namespace
   behavior; capture full output.
5. Inspect both outputs for exit code 0 and confirm the unknown-namespace rejection assertion is
   present and passing.
6. Register the result as evidence per `evidence-policy.md`'s required-field list, citing
   `AR-04/unknown_namespace_rejection_test.go` as the evidence artifact reference.

### Expected files or components affected

None — read-only verification task. `app/boot.go` and `app/boot_extra_test.go` are inspected and
re-tested, not modified, unless a regression is found (see "Rollback or recovery considerations").

### Expected output

A test-execution log (`go test -v ./app/... -run Boot` or equivalent output) plus a full-suite
`go test ./...` green-check log, both showing exit code 0 and the unknown-namespace rejection
assertion passing.

### Required artifacts

Test-execution log artifact plus full-suite green-check log artifact, registered in the story's
`artifacts/index.md` (lifecycle stage: post-implementation).

### Required evidence

One evidence record, planned ID `EV-W00-E01-S001-02`, evidence type "test-execution log + full-suite
green check," referencing `AR-04/unknown_namespace_rejection_test.go` output as the evidence
artifact name.

### Related acceptance criteria

AC-W00-E01-S001-02.

### Completion criteria

This task is `done` when: both commands have actually been executed; the result (`pass` or
`failed`) is recorded in this task's `verification.md`; the evidence record is registered in the
story's `evidence/index.md` with all required fields; and, if the result is `failed`, a follow-up
remediation task has been opened under `W05-E03-S002` per `requirement-inventory.md`'s AR-04 target
(this task is not marked `done` while a regression is unresolved and unacknowledged).

### Verification method

`go test ./app/... -run Boot` (or the identified specific test) plus a full `go test ./...`,
inspected for exit code 0 on both and presence/pass of the unknown-namespace rejection assertion.
See this task's own `verification.md` for the full planned-procedure row.

### Risks

- RISK-W00-001 (inherited) — AR-04 T1's boot-time rejection behavior could have regressed since the
  reviewed SHA.
- Full-suite `go test ./...` may surface unrelated failures outside this task's direct scope (e.g.
  from packages requiring Postgres/MinIO not available in the execution environment) — such
  failures must be triaged as environment issues (RISK-W00-002-class) versus genuine regressions
  before being attributed to this task's finding.

### Rollback or recovery considerations

If a regression is found (unknown namespaces are no longer rejected, or the error is no longer
deterministic/named), this task does not fix it. Instead: record a `failed`-status evidence record
(preserved, never deleted); open a new remediation task under `W05-E03-S002` (AR-04's canonical
target story per `requirement-inventory.md`, noting AR-04 T2-T5 already depend on AR-01); do not
silently mark this task or its parent story `done`/`accepted` while the regression is open.

## Implementation Record

Per mandate §8.7. Executed 2026-07-13. "Implementation" here means running the verification
commands and registering evidence, not writing code.

### What was actually implemented

Confirmed by direct source inspection that `app/boot.go:165-181` still rejects unknown
`modules.<name>` config namespaces at boot with a deterministic named error (offending keys
sorted; error text `config: unknown module namespace(s) %v: no registered module matches`,
accumulated into `regErrs` so boot fails before serving). Identified the specific test:
`TestBootFailsOnUnknownConfigNamespace` (`app/boot_extra_test.go:255`) — this resolves the
plan-time unresolved question about the exact test name; the `-run Boot` pattern covers it. Then
ran `go test -v ./app/... -run Boot` and a full `go test ./...` at commit
`0a31186cada5c275a588c74081cf977adf346e61`. Both exit 0 — evidence EV-W00-E01-S001-02.

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

Full-suite green check ran once on a machine with concurrent sibling-worker load; correctness-only (no timing assertions), so load does not affect validity.

### Follow-up items

None.

### Relationship to the approved plan

Execution matched the planned commands (`go test ./app/... -run Boot` with `-v`, plus full
`go test ./...`). The boot-test name assumed open at plan time was resolved during execution to
`TestBootFailsOnUnknownConfigNamespace`; the planned `-run Boot` pattern covered it, so no
command deviation occurred.

## Verification Record

### Planned verification procedure

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E01-S001-02 | Re-run `go test ./app/... -run Boot` (or the specific boot namespace test, name to be confirmed during execution) plus a full `go test ./...`; inspect both for exit code 0; confirm `app/boot.go` rejects an unknown `modules.<typo>` config namespace with a deterministic named error | Local or CI Go toolchain per `go.mod`; full-suite environment requirements to be confirmed during execution | Both commands exit 0; the boot-time test demonstrates the named-error rejection behavior for an unknown namespace | Test-execution log / `go test -v` output plus full-suite green-check log | unassigned (framework architecture lead role) |

### Actual result

`go test -v ./app/... -run Boot`: exit 0; 16 Boot-matching tests all PASS, including
`TestBootFailsOnUnknownConfigNamespace` (PASS, 0.08s) which feeds a phantom `modules.<typo>`
namespace and asserts the deterministic named boot error. Full `go test ./...`: exit 0, 57
packages, zero FAIL (3 packages report `[no test files]`); DB-backed tests executed against local
Postgres. No unrelated regression accompanies the boot-namespace behavior.

### Pass or fail

pass.

### Evidence identifier

EV-W00-E01-S001-02 (`evidence/tests/ar04-boot-run-boot.log`, sha256:d04aec5132af0008;
`evidence/tests/ar04-full-suite.log`, sha256:91427e58ded80d82).

### Execution date

2026-07-13 (boot tests 12:14:09; full suite 12:15:25–12:17:33 local).

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### Environment

Local workstation, macOS 26.5.2 (Darwin 25.5.0) arm64; go1.26.5 darwin/arm64; local Postgres +
MinIO via `make up` compose (`DATABASE_URL` set); concurrent load present (sibling W00 workers).

### Reviewer

unassigned — conductor review pending (worker self-review only; not self-marked accepted).

### Findings

AR-04 T1 unknown-namespace rejection intact at HEAD; no regression; full suite green.

### Retest status

Not applicable — first execution under this programme; result pass.

### Final conclusion

AC-W00-E01-S001-02 satisfied. AR-04 executed slice (T1) re-proven at pinned HEAD.

## Deviations Record

No deviations. Commands, scope, and environment matched `plan.md`; the plan-time open question
(exact boot-test name) was resolved during execution as the plan itself directed.
