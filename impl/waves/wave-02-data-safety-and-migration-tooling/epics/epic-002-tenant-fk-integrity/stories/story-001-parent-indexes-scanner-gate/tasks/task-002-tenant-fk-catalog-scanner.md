---
id: W02-E02-S001-T002
type: task
title: Tenant-FK catalog scanner
status: todo
parent_story: W02-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W02-E02-S001-02
artifacts:
  - ART-W02-E02-S001-002
  - ART-W02-E02-S001-005
evidence:
  - EV-W02-E02-S001-002
---

# W02-E02-S001-T002 — Tenant-FK catalog scanner

## Task Definition

### Task objective

Build a tenant-FK catalog scanner that enumerates every tenant-scoped table's foreign keys and flags
any FK not composite on `(tenant_id, …)`, keyed off the existing RLS-tagged tenant-table matrix
rather than a hand-maintained list, per PLAN DATA-01 T2's own risk note.

### Parent story

W02-E02-S001 — Parent tenant-scoped unique indexes, FK catalog scanner, and CI gate.

### Owner

unassigned

### Status

todo

### Dependencies

None (parallel-safe with T001 — disjoint code surface; PLAN's own Depends-on column lists T2 as
depending on T1 as a source concept, but this task's own scanner implementation does not require
T1's migrations to have already landed to be written and tested against a fixture schema).

### Detailed work

1. Investigate and identify what constitutes "the existing RLS-tagged tenant-table matrix" (T2's own
   risk note) — the artifact or mechanism the scanner should key off, rather than inventing or
   hand-maintaining a list of tenant-scoped tables.
2. Implement the scanner: enumerate every tenant-scoped table's foreign keys, flag any FK not
   composite on `(tenant_id, …)`.
3. Write a fixture-schema test confirming the scanner enumerates exactly the 8 known FKs
   (persons/legal_entities/party_contacts/acting_capacities → parties; resources → organizations;
   document_versions/document_access_grants/attachments → documents/document_versions) with zero
   silent gaps.
4. Document the scanner's purpose and its matrix-keying mechanism.

### Expected files or components affected

A new tenant-FK catalog scanner tool (exact package location TBD per `plan.md`'s "Unresolved
questions," expected near W02-E01-S001's manifest-schema validator).

### Expected output

A scanner tool that enumerates exactly the 8 known FKs with zero silent gaps, proven by a
fixture-schema test.

### Required artifacts

ART-W02-E02-S001-002 (tenant-FK catalog scanner), ART-W02-E02-S001-005 (documentation, shared with
T003).

### Required evidence

EV-W02-E02-S001-002 (fixture-schema test report).

### Related acceptance criteria

AC-W02-E02-S001-02.

### Completion criteria

The scanner enumerates exactly the 8 known FKs with zero silent gaps against a fixture schema
mirroring the real schema's tenant-scoped tables.

### Verification method

Direct execution of the scanner against a fixture schema, comparing its output against the known
8-edge inventory.

### Risks

An incomplete or hand-maintained-list-based enumeration would be a security-relevant defect — a
future non-composite tenant FK could land undetected. T2's own risk note names this explicitly: "Must
key off the existing RLS-tagged tenant-table matrix, not a hand-maintained list."

### Rollback or recovery considerations

If the scanner's enumeration mechanism is found to have a gap after this task's own completion,
escalate as a correctness defect and re-verify against the full known inventory before re-closing —
do not silently patch the fixture test to match an incomplete scanner.

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

*Not applicable.*

### Schema or migration changes

*Not applicable — this task builds a scanner tool, it does not itself change any table schema.*

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
| AC-W02-E02-S001-02 | Run the scanner against a fixture schema mirroring the 8 known edges | Local dev or CI, Go toolchain | Scanner enumerates exactly 8 FKs, zero silent gaps | fixture-schema test report | unassigned |

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
