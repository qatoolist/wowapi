---
id: W01-E04-S001-T003
type: task
title: Generator verb fix (DX-02)
status: done
parent_story: W01-E04-S001
owner: W01Gen
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W01-E04-S001-03
artifacts:
  - ART-W01-E04-S001-003
  - ART-W01-E04-S001-004
evidence:
  - EV-W01-E04-S001-003
---

# W01-E04-S001-T003 — Generator verb fix (DX-02)

## Task Definition

### Task objective

Fix `internal/cli/templates/crud/resource.go.tmpl:54`'s generated DELETE-route permission from
`"{{.PermPrefix}}.delete"` (an action outside the kernel's closed authorization-verb set) to
`"{{.PermPrefix}}.deactivate"` (matching both the closed set and the template's own existing soft-delete
TODO at `resource.go.tmpl:146`), **and** fix `internal/cli/scaffold_test.go:937-953`'s
`TestGenCRUDPermissionKeys`, which currently asserts the buggy `"widgets.widget.delete"` string as
correct — both changes must land together, or the bug is test-locked and will resurface.

### Parent story

W01-E04-S001 — Generator correctness — source-built CLI path validity and boot-safe CRUD generation.

### Owner

W01Gen (wave-01 generator worker)

### Status

done

### Dependencies

None — independent of T001/T002 (disjoint files: `resource.go.tmpl`/`scaffold_test.go` vs.
`init_cmd.go`/the harness).

### Detailed work

1. Run `TestGenCRUDPermissionKeys` against the current (unfixed) template and confirm it passes with
   the string `"widgets.widget.delete"` (line 949) — this is the fail-first evidence that the test
   currently locks in the bug as correct behavior, motivating why the test itself, not just the
   template, must change.
2. Edit `internal/cli/templates/crud/resource.go.tmpl:54`: change
   `"{{.PermPrefix}}.delete"` to `"{{.PermPrefix}}.deactivate"`.
3. Edit `internal/cli/scaffold_test.go:949`: change `"widgets.widget.delete"` to
   `"widgets.widget.deactivate"`.
4. Re-run `TestGenCRUDPermissionKeys` and confirm it now passes against the corrected string.
5. Confirm no other test, template, or documentation reference to the `.delete` permission key exists
   elsewhere in `internal/cli/` (a search for the literal string `.delete` scoped to
   permission-key-shaped strings, distinguishing it from the unrelated `h.delete` handler-method
   reference and the `DELETE` HTTP method string, both of which are correctly left unchanged — this
   task fixes the permission-action string only, not the HTTP verb or the handler name).
6. Do **not** widen `kernel/authz/registry.go`'s closed verb set — confirm by inspection that this task's
   diff touches only `internal/cli/`, not `kernel/authz/registry.go`.

### Expected files or components affected

`internal/cli/templates/crud/resource.go.tmpl`; `internal/cli/scaffold_test.go`.

### Expected output

A template emitting `deactivate` instead of `delete`, and a test asserting the corrected string, both
verified fail-before/pass-after.

### Required artifacts

ART-W01-E04-S001-003 (updated `resource.go.tmpl`), ART-W01-E04-S001-004 (updated `scaffold_test.go`).

### Required evidence

EV-W01-E04-S001-003 (unit-test report, fail-before/pass-after pair for `TestGenCRUDPermissionKeys`;
recorded under `DX-02/w0-t2-verb-fix.json` per `story.md` "Required evidence").

### Related acceptance criteria

AC-W01-E04-S001-03.

### Completion criteria

`TestGenCRUDPermissionKeys` passes against `"widgets.widget.deactivate"`; the pre-fix run (asserting the
old string) is recorded as the fail-first baseline; `resource.go.tmpl:54`'s permission string is
confirmed, by direct inspection, to read `deactivate`.

### Verification method

`go test ./internal/cli/... -run TestGenCRUDPermissionKeys -v`, logged output retained as evidence
before and after the fix.

### Risks

RISK-W01-005 (see `story.md` "Risks" and epic-level `../../risks.md`) — the generator fix must also fix
the generator's own test-locking assertion, or the bug re-surfaces immediately. This task's own
"Detailed work" step 1 (confirming the test currently passes with the buggy string before any edit) is
the concrete mechanism by which this task addresses RISK-W01-005 directly: it makes the test-lock
visible and provable before it is removed, rather than assuming it away.

### Rollback or recovery considerations

If reverted, this task's two file changes must be reverted together — reverting only the template change
while leaving the test's corrected assertion in place would immediately fail the test (correctly, since
the bug would have been reintroduced); reverting only the test while leaving the template fixed would
silently mask a regression if the template were later broken again. Revert as a single unit.

## Implementation Record

### What was actually implemented

Exactly the two-part fix the task defines, in one working change: (1) the generated DELETE route's
permission action in `internal/cli/templates/crud/resource.go.tmpl:54` changed from
`"{{.PermPrefix}}.delete"` to `"{{.PermPrefix}}.deactivate"` (one token); (2) the test-locking
assertion in `TestGenCRUDPermissionKeys` changed from `"widgets.widget.delete"` to
`"widgets.widget.deactivate"` (one line). Fail-first order was observed: the pre-fix
`TestGenCRUDPermissionKeys` run (PASS on the buggy string, proving the RISK-W01-005 test-lock) was
captured before either edit. Detailed-work step 5's residual search confirms the only remaining
`.delete` references in `internal/cli/` are the `h.delete` handler-method name and the `DELETE` HTTP
method string (both correctly unchanged) plus historical-context text in the new boots test. Step 6
confirmed: `git diff --stat kernel/authz/` is empty — the closed verb set was not touched.

### Components changed

`internal/cli` (generator template + its test). No kernel package modified.

### Files changed

- `internal/cli/templates/crud/resource.go.tmpl` — 1 insertion, 1 deletion (line 54).
- `internal/cli/scaffold_test.go` — 1 line (assertion at line 997; the task's cited line 949 drifted
  to 997 because sibling story W01-E03-S002 added tests earlier in the same file in the shared working
  tree — content matched the citation exactly, only line numbers moved).

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

### Tests added or modified

`TestGenCRUDPermissionKeys`'s assertion string corrected (no new test added, an existing test's
assertion fixed), exactly as anticipated.

### Commits

None yet — the change is an uncommitted working-tree delta on top of HEAD
`05dce5c8a548f7dce3222637ab2c82024236a2a0`; the wave conductor owns commits.

### Pull requests

*Not yet implemented.*

### Implementation dates

2026-07-13 (implemented and verified same day).

### Technical debt introduced

*None anticipated.*

### Known limitations

None. The generated DELETE handler body remains the documented soft-delete TODO stub — that is DX-02's
P1/Wave-4 scope, explicitly out of this task (story.md "Out of scope").

### Follow-up items

None for this task. (The wowsociety upstream-register note is sibling story W01-E04-S002/FBL-03's
concern; its owner was notified the fix has landed in the working tree.)

### Relationship to the approved plan

Matches `plan.md` step 6 exactly (template token + test assertion + fail-first run order). No deviation.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E04-S001-03 | `TestGenCRUDPermissionKeys` run before and after the fix | Local dev or CI, `go test ./internal/cli/...` | Passes against `"widgets.widget.delete"` before (documenting the test-lock); passes against `"widgets.widget.deactivate"` after | unit-test report (fail-before/pass-after pair) | unassigned |

### Actual result

Pre-fix: `TestGenCRUDPermissionKeys` PASSED asserting `"widgets.widget.delete"` — the test-lock
documented and preserved in `evidence/DX-02/t003-permkeys-prefix-failfirst.log`. Post-fix: the same
test PASSED asserting `"widgets.widget.deactivate"` (`evidence/DX-02/t003-permkeys-postfix.log`).
`resource.go.tmpl:54` inspected directly post-fix: emits `{{.PermPrefix}}.deactivate`. Full
`go test ./internal/cli/ -count=1` passes (15.2s).

### Pass or fail

PASS (fail-before/pass-after pair complete).

### Evidence identifier

EV-W01-E04-S001-003 (`evidence/DX-02/w0-t2-verb-fix.json`).

### Execution date

2026-07-13 (07:26 UTC).

### Commit or revision

HEAD `05dce5c8a548f7dce3222637ab2c82024236a2a0`; fix uncommitted on top (conductor commits).

### Environment

macOS Darwin 25.5.0 arm64, go1.26.5, local dev workstation.

### Reviewer

Pending — wave-level review gate (conductor assigns).

### Findings

One environmental finding, no defect: line drift in `scaffold_test.go` (task cites 949; actual 997)
caused by sibling-story additions in the shared working tree. Re-derived against the live file per the
wave constraint; recorded here and in the story `deviations.md`.

### Retest status

Not required — first pass verification succeeded at the pinned revision.

### Final conclusion

AC-W01-E04-S001-03 satisfied. Template and test corrected together as one unit (RISK-W01-005 closed for
this change); kernel verb set untouched.

## Deviations Record

*No deviations recorded yet.*

### Deviation ID

*Not applicable.*

### Approved plan

*Not applicable.*

### Actual implementation

*Not applicable.*

### Reason

*Not applicable.*

### Impact

*Not applicable.*

### Risks

*Not applicable.*

### Approval

*Not applicable.*

### Compensating controls

*Not applicable.*

### Follow-up work

*Not applicable.*
