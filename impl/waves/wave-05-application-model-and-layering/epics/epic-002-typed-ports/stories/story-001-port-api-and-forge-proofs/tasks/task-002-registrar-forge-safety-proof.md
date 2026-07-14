---
id: W05-E02-S001-T002
type: task
title: Compiler factory extension and registrar-forge compile-fail fixture
status: todo
parent_story: W05-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E02-S001-02
artifacts:
  - ART-W05-E02-S001-002
  - ART-W05-E02-S001-003
evidence:
  - EV-W05-E02-S001-002
---

# W05-E02-S001-T002 — Compiler factory extension and registrar-forge compile-fail fixture

## Task Definition

### Task objective

Extend the internal compiler factory to mint registrars with immutable owner identity for AR-02's
own port-key registration flow (reusing W05-E01-S001's minting mechanism), and prove — via an
adversarial compile-fail fixture — that module code cannot manufacture a `Registrar` from a bare
string, specifically verifying capability confusion is impossible given AR-01 and AR-02 share the one
`Registrar` type by design (D-02).

### Parent story

W05-E02-S001 — Typed port-key API and registrar-forge safety proof.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T001); depends on W05-E01-S001 at story scope.

### Detailed work

1. Extend the internal compiler factory (from W05-E01-S001) to mint registrars for AR-02's own
   port-key registration flow, reusing the same minting mechanism.
2. Write `AR-02/registrar_forge_compile_fail_fixture/`: simulated module code attempting to
   manufacture a `Registrar` from a bare string, specifically probing the shared-`Registrar`-type
   scenario — e.g. a capability minted in an AR-01 registration context being misused in an AR-02
   port-registration context, or vice versa.
3. Document the safety proof, referencing D-02's shared-type design.

### Expected files or components affected

The compiler factory (from W05-E01-S001, extended here); a new compile-fail fixture directory.

### Expected output

Module code cannot manufacture a `Registrar` from a bare string, and capability confusion across
AR-01/AR-02's shared `Registrar` type is proven impossible — by the named compile-fail fixture.

### Required artifacts

ART-W05-E02-S001-002, ART-W05-E02-S001-003 (documentation, shared with T001).

### Required evidence

EV-W05-E02-S001-002.

### Related acceptance criteria

AC-W05-E02-S001-02.

### Completion criteria

The compile-fail fixture genuinely fails to compile, specifically covering the cross-subsystem
capability-confusion scenario, not merely a bare-string-construction attempt.

### Verification method

Direct compilation attempt of the fixture; confirmed compile failure retained as evidence.

### Risks

High, per PLAN T2's own risk column — "verify capability confusion is impossible if AR-01/AR-02
share one `Registrar` type." See RISK-W05-E02-001 in epic-level `risks.md`.

### Rollback or recovery considerations

If the fixture is found to compile (i.e. capability confusion is possible), treat as a blocking
security defect — do not ship until genuinely closed, and escalate for redesign of the minting
mechanism if needed.

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

*Not yet implemented — this task's entire purpose is a security proof; recorded here once
implemented.*

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
| AC-W05-E02-S001-02 | Attempt to compile `AR-02/registrar_forge_compile_fail_fixture/` | Local dev or CI, Go toolchain | Fixture fails to compile, including the cross-subsystem capability-confusion scenario | compile-fail fixture report | unassigned |

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
