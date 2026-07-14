---
id: W04-E04-S003-T003
type: task
title: config doctor product-root discovery fix
status: done
parent_story: W04-E04-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W04-E04-S003-03
artifacts:
  - ART-W04-E04-S003-003
  - ART-W04-E04-S003-004
evidence:
  - EV-W04-E04-S003-003
---

# W04-E04-S003-T003 — config doctor product-root discovery fix

## Task Definition

### Task objective

Replace `config_delegate.go`'s CWD-relative `os.Stat` product-checker discovery with `go env GOMOD`/
`--project`-based discovery, so delegation works regardless of invocation directory and explicitly
reports whether product validation ran, instead of silently falling back to framework-only
validation.

### Parent story

W04-E04-S003 — Readiness and configuration diagnostics truthfulness.

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Re-read `config_delegate.go`'s current discovery logic at this task's actual start commit to
   confirm the CWD-relative `os.Stat` behavior still holds.
2. Implement discovery via `go env GOMOD` (and/or an explicit `--project` flag) so the product root
   is found regardless of the current working directory at invocation time.
3. Add explicit reporting of whether product validation ran — both the success case (product root
   found, product-aware validation engaged) and the fallback case (product root not found,
   framework-only validation ran) must be explicitly, visibly reported, not silently chosen.
4. Write the nested-subdirectory test: invoke `config doctor` from a subdirectory nested within the
   product repo, confirm discovery succeeds.
5. Write the outside-repo-with-`--project` test: invoke `config doctor` from outside the repo
   entirely, using an explicit `--project` flag, confirm discovery succeeds.
6. Document the new discovery mechanism and its explicit reporting behavior.

### Expected files or components affected

`config_delegate.go`; new unit test files for the nested-subdirectory and outside-repo-`--project`
discovery cases.

### Expected output

A working `go env GOMOD`/`--project`-based discovery mechanism; explicit product-validation-ran
reporting in both success and fallback cases; passing nested-subdirectory and outside-repo-
`--project` tests; documentation of the mechanism.

### Required artifacts

ART-W04-E04-S003-003 (config doctor discovery fix), ART-W04-E04-S003-004 (documentation, shared with
T001/T002).

### Required evidence

EV-W04-E04-S003-003 (config-doctor discovery test report).

### Related acceptance criteria

AC-W04-E04-S003-03.

### Completion criteria

Discovery succeeds regardless of invocation directory (nested subdirectory or outside-repo-with-
`--project`); both test cases pass; the tool explicitly reports whether product validation ran in
both the success and fallback cases.

### Verification method

Direct execution of the nested-subdirectory and outside-repo-`--project` discovery tests.

### Risks

Low-medium — per PLAN T3's own risk column. Confirmed non-breaking for wowsociety: its own
`tools/configcheck/main.go` already exists, so this fix is a wowapi-side-only improvement with no
required wowsociety-side action.

### Rollback or recovery considerations

Additive discovery-logic change, no schema or persistent-state impact — a code-level revert restores
the previous CWD-relative behavior if the new discovery mechanism proves incorrect in some
environment.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not yet implemented.*

### Configuration changes

*Not yet implemented.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not yet implemented.*

### Tests added or modified

*Not yet implemented.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*None anticipated.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E04-S003-03 | Run nested-subdirectory and outside-repo-`--project` discovery tests | Local dev or CI, Go toolchain | Discovery succeeds in both cases; explicit product-validation-ran reporting present | config-doctor discovery test report | unassigned |

### Actual result

*Not yet executed.*

### Pass or fail

*Not yet executed.*

### Evidence identifier

*Not yet executed.*

### Execution date

*Not yet executed.*

### Commit or revision

*Not yet executed.*

### Environment

*Not yet executed.*

### Reviewer

*Not yet executed.*

### Findings

*Not yet executed.*

### Retest status

*Not yet executed.*

### Final conclusion

*Not yet executed.*

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
