---
id: W07-E01-S002-T001
type: task
title: Index-definition audit (gap-fill)
status: done
parent_story: W07-E01-S002
owner: W07-Scoping-Dispatch.W07E01S002
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W07-E01-S002-01
artifacts:
  - ART-W07-E01-S002-001
evidence:
  - EV-W07-E01-S002-001
---

# W07-E01-S002-T001 — Index-definition audit (gap-fill)

## Task Definition

### Task objective

Verify current rule_versions index definitions before designing the new query.

### Parent story

W07-E01-S002

### Owner

unassigned

### Status

todo

### Dependencies

None. This is the story's own gap-fill task and must precede T002/T003.

### Detailed work

1. Run `grep "CREATE INDEX" migrations/*rules*.sql` (or equivalent) against actual migrations.
2. Confirm or refute the directive's own claim that current indexing favors active-only lookup.
3. Record the audit outcome.

### Expected files or components affected

An audit report (exact location TBD).

### Expected output

A confirmed-or-refuted answer to the directive's own indexing claim, based on actual migrations.

### Required artifacts

ART-W07-E01-S002-001 (index-definition audit report).

### Required evidence

EV-W07-E01-S002-001 (audit output).

### Related acceptance criteria

AC-W07-E01-S002-01.

### Completion criteria

The audit genuinely confirms or refutes the claim, based on actual migration inspection.

### Verification method

Direct grep-based inspection of migration files.

### Risks

Low, per PLAN T0's own risk classification — but must precede T2 (and, by extension, T1's design).

### Rollback or recovery considerations

Not applicable — an audit task has no code to roll back.

## Implementation Record

### What was actually implemented

Inspected the only `*rules*.sql` migration before any T1/T2 design or production edit and recorded
that `rule_versions` had an active-only exclusion constraint plus an explicit active-only lookup index
that omitted `scope_id` and `tenant_id`.

### Files changed

- `artifacts/implementation/ART-W07-E01-S002-001-index-audit.md`
- `evidence/audits/EV-W07-E01-S002-001.md`

### Implementation dates

2026-07-14.

### Relationship to the approved plan

Matched T0 exactly and completed before T1 began.
## Verification Record

| Acceptance criterion | Verification method | Required environment | Actual result | Evidence | Reviewer |
|---|---|---|---|---|---|
| AC-W07-E01-S002-01 | Migration glob plus line-numbered regex inspection | Repository at base `733ef3e` | PASS — claim confirmed before query design | EV-W07-E01-S002-001 | pending story independent review |

### Final conclusion

Passed on 2026-07-14. No retest was required.
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
