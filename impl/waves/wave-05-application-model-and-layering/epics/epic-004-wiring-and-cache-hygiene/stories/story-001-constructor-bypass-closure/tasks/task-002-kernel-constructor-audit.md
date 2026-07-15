---
id: W05-E04-S001-T002
type: task
title: kernel/kernel.go audit
status: done
parent_story: W05-E04-S001
owner: task
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W05-E04-S001-02
artifacts:
  - ART-W05-E04-S001-002
evidence:
  - EV-W05-E04-S001-002
---

# W05-E04-S001-T002 — kernel/kernel.go audit

## Task Definition

### Task objective

Audit `kernel/kernel.go` for any other instance of the closure-captures-a-fresh-instance pattern
beyond the already-fixed `orgAncestry` line, producing an explicit confirm-or-refute report.

### Parent story

W05-E04-S001 — Constructor-boundary lint and kernel.go audit.

### Owner

task

### Status

done

### Dependencies

None (parallel-safe with T001).

### Detailed work

1. Read `kernel/kernel.go` line by line, specifically looking for any closure that constructs a
   fresh infrastructure instance instead of closing over the already-composed one.
2. Document the search methodology and scope explicitly, guarding against PLAN's own named risk of
   under-scoping to just the one already-fixed line.
3. Write `AR-06/kernel_constructor_audit.md`: explicit findings, confirming or refuting the pattern's
   isolation.

### Expected files or components affected

None (investigative task; `kernel/kernel.go` itself is read, not modified, unless the audit finds a
new instance requiring a fix — in which case that is a follow-up, recorded explicitly).

### Expected output

An explicit audit report.

### Required artifacts

ART-W05-E04-S001-002.

### Required evidence

EV-W05-E04-S001-002.

### Related acceptance criteria

AC-W05-E04-S001-02.

### Completion criteria

The audit report explicitly confirms or refutes the pattern's isolation, with documented search
methodology.

### Verification method

Documentation review of `AR-06/kernel_constructor_audit.md` for completeness and explicit findings.

### Risks

Low, per PLAN T3's own risk column — "mostly investigative; risk is under-scoping to just the one
cited line."

### Rollback or recovery considerations

If the audit finds a new instance of the pattern, record it explicitly and treat fixing it as a
follow-up item — do not silently omit the finding to keep the audit "clean."

## Implementation Record

Audited every executable cross-package `New*` call and every anonymous function literal
in `kernel/kernel.go`, rather than limiting inspection to the historical `orgAncestry`
line. The detailed method and findings are recorded in
`evidence/AR-06/kernel_constructor_audit.md`.

### What was actually implemented

The report distinguishes the 23 executable constructor calls from the non-executable
`authz.NewStore()` mention in the explanatory comment, and reviews all three anonymous
closures for fresh-instance construction.

### Components changed

Story evidence only; `kernel/kernel.go` required no code change.

### Files changed

- `evidence/AR-06/kernel_constructor_audit.md`
- `evidence/index.md`
- `artifacts/index.md`

### Tests added or modified

No runtime test applies to this investigative task. T001's permanent analyzer is the
regression guard for future constructor bypasses.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

Per AR-06 T3, the audit scope is specifically `kernel/kernel.go`; T001 enforces the
broader production-source boundary.

### Follow-up items

None. No additional bypass was found.

### Relationship to the approved plan

Implemented as planned, including the named under-scoping control.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W05-E04-S001-02 | inspect every constructor and closure recorded in `evidence/AR-06/kernel_constructor_audit.md` | source tree at baseline plus W05 diff | explicit confirm/refute with full-file scope | audit report | task |

### Actual result

All 23 executable cross-package constructors occur in composition code. Each of the
three anonymous closures reuses composed dependencies or constructs no infrastructure.
No remaining closure-captures-a-fresh-instance bypass was found.

### Pass or fail

Pass.

### Evidence identifier

EV-W05-E04-S001-002.

### Execution date

2026-07-13.

### Commit or revision

Baseline `733ef3e` plus W05 working-tree changes.

### Environment

Source audit on Darwin arm64.

### Reviewer

task.

### Findings

The historical bypass is isolated to the already-fixed `orgAncestry` site.

### Retest status

Not applicable; the permanent T001 analyzer supplies regression enforcement.

### Final conclusion

AC-W05-E04-S001-02 is satisfied.

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
