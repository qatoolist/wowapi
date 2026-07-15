---
id: W05-E01-S001-T001
type: task
title: ApplicationModel/Compiler lifecycle skeleton and D-03 error/panic behavior
status: todo
parent_story: W05-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E01-S001-01
  - AC-W05-E01-S001-02
artifacts:
  - ART-W05-E01-S001-001
  - ART-W05-E01-S001-003
evidence:
  - EV-W05-E01-S001-001
  - EV-W05-E01-S001-002
---

# W05-E01-S001-T001 — ApplicationModel/Compiler lifecycle skeleton and D-03 error/panic behavior

## Task Definition

### Task objective

Define the `ApplicationModel` type and `Compiler`'s `collect → validate → seal → expose read-only
snapshot` lifecycle skeleton, and implement post-seal mutation behavior per D-03: error in
production builds, panic only under an explicit dev/test build tag.

### Parent story

W05-E01-S001 — ApplicationModel lifecycle skeleton and Registrar capability type.

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Re-read `kernel/module`'s current registration surface at this task's actual start commit to
   confirm no lifecycle skeleton currently exists.
2. Define the `ApplicationModel` type: an immutable, read-only-snapshot-exposing type.
3. Define the `Compiler` type: accumulates declarations via owner-bound calls; exposes `Compile()`,
   which validates then seals the accumulated state into an `ApplicationModel`.
4. Implement the state-machine transitions (collecting → validating → sealed) with clear, typed
   errors on invalid transitions.
5. Implement post-seal mutation behavior per D-03: a typed error in production builds; a panic only
   under an explicit dev/test build tag (Go build constraint, not a runtime environment check).
6. Write state-machine transition unit tests.
7. Write the build-tag-scoped test proving the production build errors (never panics) and the
   explicit dev/test-tagged build panics post-seal.
8. Document the lifecycle and the D-03 decision this task enacts.

### Expected files or components affected

A new `ApplicationModel`/`Compiler` lifecycle skeleton (exact location TBD per `plan.md`).

### Expected output

A working `Compiler` that validates-then-seals into an immutable `ApplicationModel`, with
D-03-compliant post-seal mutation behavior, proven by state-machine and build-tag-scoped tests.

### Required artifacts

ART-W05-E01-S001-001 (lifecycle skeleton), ART-W05-E01-S001-003 (documentation, shared with T002).

### Required evidence

EV-W05-E01-S001-001 (state-machine transition test output), EV-W05-E01-S001-002 (build-tag matrix
test output).

### Related acceptance criteria

AC-W05-E01-S001-01, AC-W05-E01-S001-02.

### Completion criteria

`Compile()` validates then seals; post-seal calls error in production builds and panic only under
the explicit dev/test build tag — proven by both test suites passing.

### Verification method

Direct execution of the state-machine unit tests and the build-tag matrix test (both build-tag
configurations run in CI).

### Risks

None beyond the epic-level risks already recorded — this task is foundational but not itself flagged
High risk in the source (T1's own PLAN risk column: "Medium — load-bearing type every other AR-0x
task depends on").

### Rollback or recovery considerations

If the build-tag-scoped test reveals the error/panic split does not correctly separate by build
configuration, revert and redesign the build-constraint mechanism before proceeding — do not ship a
lifecycle skeleton where a production build could panic.

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
| AC-W05-E01-S001-01 | Run state-machine transition unit tests | Local dev or CI, Go toolchain | `Compile()` validates then seals; post-seal errors in production build | unit-test report | unassigned |
| AC-W05-E01-S001-02 | Run build-tag matrix test (default vs. dev/test tag) | Local dev or CI, Go toolchain (build-tag matrix) | Production errors, never panics; dev/test-tagged build panics post-seal | unit-test report | unassigned |

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
