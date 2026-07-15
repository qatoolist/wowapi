---
id: W06-E01-S001-T001
type: task
title: Draft module-DSL design options and trade-offs
status: done
parent_story: W06-E01-S001
owner: W06E01Impl
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W06-E01-S001-01
artifacts:
  - ART-W06-E01-S001-001
evidence: []
---

# W06-E01-S001-T001 — Draft module-DSL design options and trade-offs

## Task Definition

### Task objective

Draft the module-DSL design's options and trade-offs for `port`, `Manifest[T]`, and `Operation[Request,Response]`, grounded in AR-01's Registrar capability and AR-02's typed port.Key[T] mechanism, and select a design with documented rationale.

### Parent story

W06-E01-S001

### Owner

W06E01Impl

### Status

done

### Dependencies

None (this story's own dependency on W05's AR-01/AR-02 acceptance is a story-level entry gate, not a task-to-task dependency within this story).

### Detailed work

1. Re-read the directive's own module-DSL prose and AR-01/AR-02's `accepted` implementation
   (W05-E01, W05-E02) to ground the design in the framework's actual typed-registration mechanisms.
2. Draft storage/typing options for `port` (a typed capability descriptor), `Manifest[T]` (a typed
   per-module declaration surface), and `Operation[Request,Response]` (a typed request/response
   contract), with trade-offs for each.
3. Select a design and document the rationale, explicitly describing how it relates to and builds on
   AR-01's `Registrar` and AR-02's `port.Key[T]` rather than replacing them.
4. Write the design document.

### Expected files or components affected

A new design document (exact path TBD at implementation time).

### Expected output

A design document covering all three DSL elements with documented trade-offs and a selected design grounded in AR-01/AR-02's actual shape.

### Required artifacts

ART-W06-E01-S001-001 (module-DSL design document).

### Required evidence

None beyond the design document itself.

### Related acceptance criteria

AC-W06-E01-S001-01.

### Completion criteria

The design document exists, covers `port`, `Manifest[T]`, and `Operation[Request,Response]` at implementer-actionable detail, and explicitly references AR-01/AR-02.

### Verification method

Direct inspection of the design document's completeness and grounding.

### Risks

None beyond the general design-before-implementation risk recorded at story scope.

### Rollback or recovery considerations

If the draft is found materially incomplete during T002's labeling/formalization step, revise the draft directly — no formal rollback process required for a pre-formalization draft.

## Implementation Record

Implemented 2026-07-13 by W06E01Impl.

### What was actually implemented

Created `docs/implementation/module-dsl-target-design.md`, an implementer-actionable future-state
design covering typed ports, `Manifest[TConfig]`, and `Operation[Request,Response]`.

### Components changed

Documentation only.

### Files changed

- `docs/implementation/module-dsl-target-design.md`
- story artifact/evidence/lifecycle records

### Interfaces introduced or changed

None. Every proposed interface is visibly labeled target-not-implemented.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None at runtime. The design records fail-closed authentication/authorization invariants for a future
compiler.

### Observability changes

None at runtime. The design records operation observability policy as a future compiler projection.

### Tests added or modified

None; the acceptance method is direct documentation inspection.

### Commits

None; uncommitted shared working tree at base revision
`733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None. Actual DSL/compiler/runtime implementation remains explicitly out of scope rather than being
introduced as partial code.

### Known limitations

This is a future-state design, not an available module-authoring API.

### Follow-up items

DX-03 T1..Tn implementation remains outside this programme.

### Relationship to the approved plan

Matched `plan.md`: the design was grounded in the landed W05 `ApplicationModel`,
`Registrar[T]`, and `port.Key[T]` APIs; alternatives and a selected shape are documented.
## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E01-S001-01 | Direct inspection | Documentation review | Design document exists and covers all three DSL elements with AR-01/AR-02 grounding | review report | W06E01Impl |

### Actual result

The design specifies the author model, compiler phases, invariants, runtime boundary, diagnostics,
compatibility policy, options, and future implementation sequence for all three required concepts.
It explicitly builds on the landed W05 APIs and prohibits a parallel model/port graph.

### Pass or fail

Pass.

### Evidence identifier

EV-W06-E01-S001-001 (`evidence/design-completeness-review.md`).

### Execution date

2026-07-13.

### Commit or revision

Base `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus uncommitted story artifact.

### Environment

Documentation inspection.

### Reviewer

W06E01Impl.

### Findings

No open design-completeness findings.

### Retest status

Not applicable; no executable implementation exists.

### Final conclusion

AC-W06-E01-S001-01 is verified.
## Deviations Record

*No task-local deviation; story-level entry-gate deviation DEV-W06-E01-S001-001 remains pending.*

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
