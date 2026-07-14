---
id: W05-E01-S001-T002
type: task
title: Registrar capability type and typed-key mechanism per D-02
status: todo
parent_story: W05-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E01-S001-03
artifacts:
  - ART-W05-E01-S001-002
  - ART-W05-E01-S001-003
evidence:
  - EV-W05-E01-S001-003
---

# W05-E01-S001-T002 — Registrar capability type and typed-key mechanism per D-02

## Task Definition

### Task objective

Define the owner-bound `Registrar` capability type — a single generic type per D-02, with an
unexported seal method, minted only by the `Compiler` from `Manifest.ID`/`Module.Name()` — and the
per-subsystem typed-key mechanism that prevents capability confusion across subsystems sharing the
one `Registrar` type. Prove, via a compile-fail fixture, that module code cannot construct or
type-assert a `Registrar` for another owner.

### Parent story

W05-E01-S001 — ApplicationModel lifecycle skeleton and Registrar capability type.

### Owner

unassigned

### Status

todo

### Dependencies

None (parallel-safe with T001 — disjoint code surface within the same story's shared design
context; the `Registrar` type does not require the `Compiler`'s full lifecycle implementation to be
complete, only its minting entry point to be designed consistently).

### Detailed work

1. Design the `Registrar` capability type per D-02's resolution: one generic type (not
   per-subsystem distinct types), with an unexported seal method.
2. Implement the compiler-only minting path: the `Registrar` is constructible only from within the
   `Compiler`'s own package, derived from `Manifest.ID`/`Module.Name()`.
3. Design the per-subsystem typed-key mechanism (`Key[T]`-shaped) that binds to the `Registrar`,
   preventing capability confusion across subsystems without multiplying `Registrar` types — per
   D-02's explicit "capability confusion is prevented by the key's phantom type + owner binding, not
   by multiplying registrar types."
4. Write the compile-fail fixture: simulated module code attempting to construct a `Registrar`
   directly (bypassing the compiler) or type-assert one issued to another owner — this fixture must
   fail to compile.
5. Document the `Registrar` capability type, the typed-key mechanism, and the D-02 decision this
   task enacts.

### Expected files or components affected

A new `Registrar` capability type and typed-key mechanism (exact location TBD per `plan.md`); a new
compile-fail fixture (exact mechanism TBD).

### Expected output

A `Registrar` capability type mintable only by the compiler, with a per-subsystem typed-key
mechanism, proven secure by a compile-fail fixture.

### Required artifacts

ART-W05-E01-S001-002 (Registrar capability type), ART-W05-E01-S001-003 (documentation, shared with
T001).

### Required evidence

EV-W05-E01-S001-003 (compile-fail fixture output, `AR-01/registrar_capability_test_output.txt`).

### Related acceptance criteria

AC-W05-E01-S001-03.

### Completion criteria

Module code cannot construct or type-assert a `Registrar` for another owner — proven by the
compile-fail fixture genuinely failing to compile, not merely by code review.

### Verification method

Direct compilation attempt of the fixture; confirmed compile failure retained as evidence (e.g. the
build log showing the expected compile error).

### Risks

RISK-W05-E01-S001-001 and RISK-W05-001 (epic/story-level) — this is, per PLAN T2's own risk column,
"the actual security boundary." An incorrectly-scoped seal method (e.g. accidentally exported, or a
reflection-based bypass left open) would defeat this task's entire purpose.

### Rollback or recovery considerations

If the compile-fail fixture is found to actually compile (i.e. the security boundary is bypassable),
treat as a blocking defect — do not ship T002 until the fixture genuinely fails to compile.

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

*Not yet implemented — the Registrar's seal-method boundary is the security change; recorded here
once implemented.*

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
| AC-W05-E01-S001-03 | Attempt to compile the fixture that fabricates a Registrar for another owner | Local dev or CI, Go toolchain | Fixture fails to compile | compile-fail fixture report | unassigned |

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
