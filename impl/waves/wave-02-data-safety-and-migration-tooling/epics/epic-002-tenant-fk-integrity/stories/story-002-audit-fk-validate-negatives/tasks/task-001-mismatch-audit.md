---
id: W02-E02-S002-T001
type: task
title: Mismatch audit (DATA-01 T3)
status: todo
parent_story: W02-E02-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W02-E02-S002-01
artifacts:
  - ART-W02-E02-S002-001
evidence:
  - EV-W02-E02-S002-001
  - EV-W02-E02-S002-002
---

# W02-E02-S002-T001 — Mismatch audit (DATA-01 T3)

## Task Definition

### Task objective

Prove `child.tenant_id = parent.tenant_id` for every existing row across all 8 tenant-scoped FK
edges, using a platform-role connection to bypass RLS for the scan, against staging/prod-shaped
data. Fail deployment on any mismatch found. Write an integration test that seeds a deliberate
cross-tenant mismatch and confirms the audit tool detects it.

### Parent story

W02-E02-S002 — Cross-tenant mismatch audit, composite FK validation, and negative tests.

### Owner

unassigned

### Status

todo

### Dependencies

None hard. Soft: more useful once W02-E02-S001-T002's catalog scanner confirms the exact 8-edge FK
inventory is complete and current — an audit against an incomplete inventory could silently miss an
edge, per PLAN's own T3 Depends-on column ("T2").

### Detailed work

1. Confirm, at this task's actual start commit, the current 8-edge FK inventory via
   W02-E02-S001-T002's catalog scanner output — key off the scanner's output, not a hand-maintained
   list, per PLAN's own T2 risk note.
2. Build the mismatch-audit tool: a platform-role-connected scan across all 8 edges, producing a
   dated, inspectable report (edge identifier, row count scanned, mismatch count, timestamp,
   connection role used — exact schema TBD, informed by W02-E02-S001's own scanner reporting
   convention for consistency).
3. Write the integration test that seeds a deliberate cross-tenant mismatch via a platform-role
   connection and confirms the audit tool detects it — independent of, and not a substitute for, the
   real-data audit run.
4. Run the real-data mismatch audit against staging/prod-shaped data.
5. **Branch on outcome:**
   - Zero-mismatch: record the report as ART-W02-E02-S002-001, proceed to T002.
   - Mismatch found: halt, escalate to the acceptance authority (data/reliability lead) per
     RISK-W02-002's documented path, record the finding and its eventual resolution in
     `../deviations.md`, and do not signal T002/T003 as unblocked until a second zero-mismatch audit
     passes.

### Expected files or components affected

A new mismatch-audit tool (exact package location TBD, expected adjacent to W02-E02-S001's catalog
scanner tool).

### Expected output

A zero-mismatch report against staging/prod-shaped data for all 8 edges (or a documented,
escalated, and resolved remediation-decision record if a mismatch was found); the seeded-mismatch
integration test passing.

### Required artifacts

ART-W02-E02-S002-001 (mismatch-audit tool and its report).

### Required evidence

EV-W02-E02-S002-001 (zero-mismatch report, or resolved remediation-decision record),
EV-W02-E02-S002-002 (seeded-mismatch integration test output).

### Related acceptance criteria

AC-W02-E02-S002-01.

### Completion criteria

The audit tool produces a zero-mismatch report across all 8 edges (or the RISK-W02-002 escalation
path has run to a resolved, recorded conclusion); the seeded-mismatch integration test proves the
tool actually detects a mismatch when one exists, not merely that it reports clean by default.

### Verification method

Integration test execution against a seeded fixture; direct execution of the real-data audit against
staging/prod-shaped data, output retained as evidence.

### Risks

RISK-W02-002 (the audit may find real cross-tenant data requiring a remediation decision this task's
own scope cannot make unilaterally) — see `../../risks.md` for full detail.

### Rollback or recovery considerations

The audit tool is read-only; it has no rollback surface of its own. If a mismatch is found, the
remediation action taken against the data (once decided by the acceptance authority) carries its own
rollback considerations, not specified here per mandate §18 (genuinely undecidable in advance).

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

*Not yet implemented — anticipated: platform-role connection configuration for the audit tool.*

### Schema or migration changes

*Not applicable — this task is read-only.*

### Security changes

*Not applicable — this task audits existing data, it does not change access control.*

### Observability changes

*Not yet implemented — anticipated: the dated, inspectable audit report itself.*

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
| AC-W02-E02-S002-01 | Run the mismatch-audit tool via a platform-role connection against staging/prod-shaped data; run the seeded-mismatch integration test | Staging or prod-shaped environment, platform-role DB connection | Zero-mismatch report across all 8 edges (or a documented, resolved remediation decision); seeded-mismatch fixture correctly detected | audit report + integration-test report | unassigned |

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
