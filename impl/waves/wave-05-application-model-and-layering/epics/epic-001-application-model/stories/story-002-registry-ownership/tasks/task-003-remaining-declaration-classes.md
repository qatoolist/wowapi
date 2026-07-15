---
id: W05-E01-S002-T003
type: task
title: Owner-bound wrappers for remaining ~9+ declaration classes
status: todo
parent_story: W05-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E01-S002-03
artifacts:
  - ART-W05-E01-S002-004
  - ART-W05-E01-S002-005
evidence:
  - EV-W05-E01-S002-003
---

# W05-E01-S002-T003 — Owner-bound wrappers for remaining ~9+ declaration classes

## Task Definition

### Task objective

Audit the framework's actual registration surface to confirm the full list of declaration classes
beyond the three headline registries (starting from PLAN's own named list: events, jobs, workflow
actions, providers, templates, health checks, migrations, seeds, OpenAPI), then build an owner-bound
registrar wrapper for each, following the T001/T002 pattern, so that every declaration class in
AR-01's own acceptance gate is ownership-checked.

### Parent story

W05-E01-S002 — Owner-bound registry wrappers across all declaration classes.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T001, T002 — disjoint registration surfaces); depends on
W05-E01-S001 (T1, T2, and ideally the T001/T002 reference pattern) at story scope. PLAN's own T6
dependency row: "T1, T2, T3-T5 pattern" — this task benefits from T001/T002 landing first as a
reference implementation, though it is not strictly blocked on their completion.

### Detailed work

1. Enumerate the framework's actual registration surface, confirming or correcting PLAN's own named
   list (events, jobs, workflow actions, providers, templates, health checks, migrations, seeds,
   OpenAPI) against the real codebase at this task's start commit — record the confirmed list as
   ART-W05-E01-S002-005, resolving PLAN's own "~9+" estimate to an exact, explicit count.
2. For each confirmed declaration class, implement an owner-bound registrar wrapper following the
   T001/T002 pattern (structural ownership via S001's `Registrar` type, not string comparison).
3. Write the table-driven adversarial suite `AR-01/full_declaration_class_matrix_test.go`: one
   fixture per confirmed declaration class, each proving a cross-module claim attempt is rejected.
4. Document the full declaration-class enumeration and each wrapper.

### Expected files or components affected

The packages backing each confirmed declaration class (exact list TBD by step 1's audit).

### Expected output

An explicit, audited declaration-class enumeration; an owner-bound wrapper for each; a table-driven
adversarial suite proving every class is ownership-checked.

### Required artifacts

ART-W05-E01-S002-004 (wrappers), ART-W05-E01-S002-005 (enumeration/audit record).

### Required evidence

EV-W05-E01-S002-003 (table-driven adversarial-test report).

### Related acceptance criteria

AC-W05-E01-S002-03.

### Completion criteria

Every declaration class in AR-01's acceptance gate is ownership-checked, proven by the table-driven
adversarial suite passing with one fixture per confirmed class — and the confirmed class count is
explicitly documented, not left as PLAN's own approximate "~9+."

### Verification method

Direct execution of `AR-01/full_declaration_class_matrix_test.go`; independent cross-check of the
audit's declaration-class list against the framework's actual registration call sites (not merely
trusting PLAN's own named list without re-confirming it against the real codebase).

### Risks

Medium — PLAN's own explicit risk note: "easy to under-scope." See RISK-W05-002 in epic-level
`risks.md`.

### Rollback or recovery considerations

If a declaration class is found missing from the enumeration after this task's own completion (e.g.
at independent-review time), treat as a task-scope correction recorded in `deviations.md`, extending
the table-driven suite with the missing fixture — not a silent narrowing of AR-01's acceptance gate.

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

*Not yet implemented.*

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
| AC-W05-E01-S002-03 | Run `AR-01/full_declaration_class_matrix_test.go` | Local dev or CI, Go toolchain | Every declaration-class fixture rejects a cross-module claim | adversarial-test report (table-driven) | unassigned |

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
