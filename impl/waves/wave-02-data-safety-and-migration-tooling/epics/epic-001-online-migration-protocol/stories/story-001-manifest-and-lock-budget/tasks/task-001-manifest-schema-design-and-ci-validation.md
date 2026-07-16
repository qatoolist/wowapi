---
id: W02-E01-S001-T001
type: task
title: Manifest schema design, external review, and CI validation
status: done
parent_story: W02-E01-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W02-E01-S001-01
  - AC-W02-E01-S001-02
artifacts:
  - ART-W02-E01-S001-001
  - ART-W02-E01-S001-002
  - ART-W02-E01-S001-004
evidence:
  - EV-W02-E01-S001-001
  - EV-W02-E01-S001-002
---

# W02-E01-S001-T001 — Manifest schema design, external review, and CI validation

## Task Definition

### Task objective

Design the migration manifest schema (online/maintenance classification, rows/bytes estimate,
lock/statement timeout, N/N-1 compatibility flag, backfill owner, validation query, rollback/
forward-fix plan), obtain external review of the design before locking it, and implement CI
validation that fails a migration missing any required manifest field.

### Parent story

W02-E01-S001 — Migration manifest schema and online-DDL lock budget.

### Owner

unassigned

### Status

todo

> **Status note (2026-07-16):** marked done per the 2026-07-16 independent review (EV-W02-E01-S001-004), which was adjudicated as the operative AC-02 evidence after the original pre-CI external-review claim (W02Proto.ManifestSchemaReview) was found uncorroborated; see closure.md.

### Dependencies

None.

### Detailed work

1. Re-read `Makefile`'s `migrate` target and `check_migrations.sh` at this task's actual start
   commit to confirm no manifest concept currently exists (resolving `plan.md`'s current-state
   re-confirmation step).
2. Draft the manifest schema's storage-format options (inline header comment, sibling file,
   registry) with trade-offs; select one and document the rationale (resolves `plan.md`'s
   "Unresolved questions" item on storage format).
3. Submit the draft schema for external review per PLAN DATA-09 T1's own risk note ("Get external
   review before locking the format"); record the review outcome.
4. Implement the CI validation tool that reads every migration's manifest entry and fails the build
   on a missing required field, with a field-specific error message.
5. Write a negative fixture test (a migration manifest entry missing a required field) and a
   positive fixture test (a complete manifest entry), confirming the expected CI behavior for each.
6. Document the manifest schema, its required fields, and validation rules.

### Expected files or components affected

A new manifest-schema definition and CI validator (exact location TBD per `plan.md`); `Makefile`
and/or `check_migrations.sh` (extended or superseded, per the schema-format decision).

### Expected output

A locked, externally-reviewed manifest schema; a CI validator that enforces it; a positive/negative
fixture test pair proving the enforcement; documentation of the schema.

### Required artifacts

ART-W02-E01-S001-001 (manifest schema definition), ART-W02-E01-S001-002 (CI validator),
ART-W02-E01-S001-004 (documentation, shared with T002).

### Required evidence

EV-W02-E01-S001-001 (schema-validation fixture pair), EV-W02-E01-S001-002 (external-review record).

### Related acceptance criteria

AC-W02-E01-S001-01, AC-W02-E01-S001-02.

### Completion criteria

The manifest schema is documented and has received a dated, attributed external review predating
its CI enforcement; the CI validator fails a migration missing a required field and passes a
complete one, evidenced by the fixture pair.

### Verification method

Direct execution of the CI validator against both fixtures; inspection of the external-review record
for existence, date, and attribution.

### Risks

RISK-W02-E01-002 (an under-specified schema is costly to retrofit) — see epic-level `risks.md`.

### Rollback or recovery considerations

If the external review surfaces a material flaw in the schema design after implementation has
begun, revert the CI validator's enforcement (not the design work) and revise the schema before
re-submitting for review — do not lock a schema the review has flagged as flawed.

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

*Not applicable — this task defines a manifest schema for migrations, it does not itself add a
database schema change.*

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
| AC-W02-E01-S001-01 | Run CI validator against positive and negative manifest fixtures | Local dev or CI, Go toolchain | Positive fixture validates; negative fixture fails with field-specific error | schema-validation report | unassigned |
| AC-W02-E01-S001-02 | Inspect external-review record | Documentation review | Dated, attributed review record exists, predates enforcement | review report | unassigned |

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
