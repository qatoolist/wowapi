---
id: W06-E01-S001-T002
type: task
title: Formalize design into a labeled ADR-style decision record
status: done
parent_story: W06-E01-S001
owner: W06E01Impl
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W06-E01-S001-T001
acceptance_criteria:
  - AC-W06-E01-S001-02
artifacts:
  - ART-W06-E01-S001-002
evidence: []
---

# W06-E01-S001-T002 — Formalize design into a labeled ADR-style decision record

## Task Definition

### Task objective

Formalize T001's selected design into an ADR-style decision record, explicitly and visibly labeled 'target, not implemented' per AR-05's labeling convention, and confirm the labeling is correct such that AR-05 T5's future lint would not flag it.

### Parent story

W06-E01-S001

### Owner

W06E01Impl

### Status

done

### Dependencies

W06-E01-S001-T001 (the design document must exist before it can be formalized into a decision record).

### Detailed work

1. Take T001's selected design and formalize it into an ADR-style decision record, following the
   same shape as the D-01..D-09 decision records already produced at W00
   (`impl/waves/wave-00-baseline-and-verification/epics/epic-002-baseline-capture/stories/
   story-003-adr-ification/decisions/`).
2. Explicitly and visibly label the decision record "target, not implemented," per AR-05's labeling
   convention (PLAN DX-03-T0's own instruction: "explicitly labeled 'target, not implemented' per
   AR-05").
3. Confirm no implementation code, compiler, or runtime type-system change accompanies either document.
4. Cross-check the labeling against AR-05 T5's stated lint requirement (a future-state design block that
   is not labeled fails the lint) — even though the lint itself does not yet exist (W06-E04-S002), this
   task's own labeling must be correct in anticipation of it.

### Expected files or components affected

A new ADR-style decision record (exact path TBD at implementation time).

### Expected output

A decision record formalizing the design, visibly labeled 'target, not implemented,' with no accompanying code.

### Required artifacts

ART-W06-E01-S001-002 (ADR-style decision record).

### Required evidence

None beyond the decision record itself.

### Related acceptance criteria

AC-W06-E01-S001-02.

### Completion criteria

The decision record exists, formalizes T001's design, and is visibly and correctly labeled 'target, not implemented'; no code accompanies it.

### Verification method

Direct inspection of the decision record for the presence and correct placement of the label, and confirmation no code was produced.

### Risks

The primary risk is an incorrectly-placed or missing label, which would cause this story's own output to fail AR-05 T5's lint once that lint exists — mitigated by this task's own explicit cross-check step.

### Rollback or recovery considerations

If the label is found missing or misplaced after this task is marked complete, correct it directly — a documentation labeling fix does not require a formal rollback process.

## Implementation Record

Implemented 2026-07-13 by W06E01Impl.

### What was actually implemented

Created story-local `decisions.md` in the established ADR shape, selecting typed immutable manifests
and typed operations compiled into the landed `ApplicationModel`.

### Components changed

Documentation only.

### Files changed

- `decisions.md`
- story artifact/evidence/lifecycle records

### Interfaces introduced or changed

None. The ADR is explicitly a future-state decision.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None.

### Observability changes

None.

### Tests added or modified

None by this task. W06E04Impl independently ran the AR-05 documentation gate against the design record.

### Commits

None; uncommitted shared working tree at base revision
`733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

The ADR intentionally decides a target contract without implementing it.

### Follow-up items

DX-03 implementation remains out of scope.

### Relationship to the approved plan

Matched `plan.md` and followed the W00 ADR structure. The exact record location was resolved to the
story-local `decisions.md`; the design record lives in `docs/implementation/` so AR-05 can lint it.
## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E01-S001-02 | Direct inspection of label placement and code absence, plus AR-05 gate | Documentation review | ADR formalizes the design, visibly labels it target-not-implemented, and adds no code | review report | W06E04Impl |

### Actual result

The ADR places `> **Target, not implemented.**` immediately after its future-state heading and repeats
the status in the decision body. The design record uses the same exact marker. W06E04Impl reported that
`go test ./internal/tools/docexamples -run TestRepositoryDocumentationPassesAllGates` included
`docs/implementation/module-dsl-target-design.md` and passed.

### Pass or fail

Pass.

### Evidence identifier

EV-W06-E01-S001-002 (`evidence/labeling-correctness-review.md`).

### Execution date

2026-07-13.

### Commit or revision

Base `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus uncommitted story artifact.

### Environment

Documentation inspection and focused AR-05 documentation gate.

### Reviewer

W06E04Impl (independent AR-05 label check).

### Findings

No open label or code-absence findings. Story-local `decisions.md` is intentionally outside the AR-05
scan path but carries the exact marker independently.

### Retest status

Passed on the focused AR-05 gate.

### Final conclusion

AC-W06-E01-S001-02 is verified; no DX-03 implementation code exists.
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
