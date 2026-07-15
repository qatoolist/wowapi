---
id: W05-E02-S003-T002
type: task
title: Legacy port adapter
status: todo
parent_story: W05-E02-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E02-S003-02
artifacts:
  - ART-W05-E02-S003-002
evidence:
  - EV-W05-E02-S003-002
---

# W05-E02-S003-T002 — Legacy port adapter

## Task Definition

### Task objective

Build a legacy port adapter (`ProvidePort`/`Port` shim onto the typed graph) so any existing caller
(confirmed zero in wowsociety; possibly wowapi-internal fixtures) compiles/resolves unchanged,
re-confirming the zero-external-caller finding at this task's own start commit.

### Parent story

W05-E02-S003 — Lifecycle manifest retirement and legacy port adapter.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T001); depends on W05-E02-S002 at story scope.

### Detailed work

1. Re-run a repo-wide search (wowapi-internal and wowsociety) for `ProvidePort`/`Port(` call sites,
   confirming PLAN's own "zero external callers" finding still holds.
2. Implement the legacy port adapter shimming any confirmed caller onto the typed graph.
3. Write `AR-02/legacy_port_adapter_compat_test_output.txt`'s producing integration test.

### Expected files or components affected

A new legacy port adapter package (exact location TBD).

### Expected output

Existing calls (if any) compile/resolve unchanged through the adapter.

### Required artifacts

ART-W05-E02-S003-002.

### Required evidence

EV-W05-E02-S003-002.

### Related acceptance criteria

AC-W05-E02-S003-02.

### Completion criteria

The integration test confirms unchanged compile/resolve behavior for any existing caller, and the
repo-wide search result (zero-callers or otherwise) is explicitly recorded.

### Verification method

Direct execution of the integration test; the repo-wide search result is retained as part of this
task's own evidence.

### Risks

Low, per PLAN T7's own risk column — "confirmed zero external callers."

### Rollback or recovery considerations

If the repo-wide search reveals a caller PLAN's own finding did not anticipate, treat as a
deviation — do not silently narrow the adapter's scope to exclude that caller.

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

*Not applicable.*

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
| AC-W05-E02-S003-02 | Run the legacy port adapter integration test | Local dev or CI, Go toolchain | Existing calls compile/resolve unchanged | integration-test report | unassigned |

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
