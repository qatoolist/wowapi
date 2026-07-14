---
id: W04-E04-S001-T001
type: task
title: Audit hash-chain widening, hash_version migration, and version-branched verification
status: done
parent_story: W04-E04-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W04-E04-S001-01
  - AC-W04-E04-S001-02
  - AC-W04-E04-S001-03
artifacts:
  - ART-W04-E04-S001-001
  - ART-W04-E04-S001-002
  - ART-W04-E04-S001-003
  - ART-W04-E04-S001-004
evidence:
  - EV-W04-E04-S001-001
  - EV-W04-E04-S001-002
  - EV-W04-E04-S001-003
---

# W04-E04-S001-T001 — Audit hash-chain widening, hash_version migration, and version-branched verification

## Task Definition

### Task objective

Widen `kernel/audit`'s `chainHash` to cover every persisted field, including canonicalized `metadata`
and `tx_id`; add a `hash_version smallint NOT NULL DEFAULT 1` column in the same migration per D-04;
implement version-branched verification so historical rows verify under v1 and new rows verify under
v2; ship the migration through W02-E01's online-migration protocol; and prove the fix with a per-field
tamper test mutating each declared field independently.

### Parent story

W04-E04-S001 — Audit hash-chain widening with hash_version discriminator.

### Owner

unassigned

### Status

todo

### Dependencies

None at task level. This story as a whole depends on W02-E01's exit gate (see `story.md`
"Dependencies").

### Detailed work

1. Re-read `kernel/audit/audit.go:130-311` at this task's actual start commit to confirm the
   current-state assessment (15-field `chainHash`, `metadata`/`tx_id` excluded, no `hash_version`
   column) still holds.
2. Design and implement the `metadata` canonicalization function: a deterministic pre-serialization
   form, never the stored jsonb directly, per `plan.md`'s canonicalization requirement.
3. Confirm and include `tx_id`'s field representation in the widened hash input.
4. Re-derive and confirm the complete field list for the widened `chainHash` from the actual code
   (all nullable fields, sequence, ID, timestamps, previous hash, plus metadata and tx_id) — not
   assumed from this task's own paraphrase.
5. Choose and document the exact `hash_version` value for the new scheme (D-04 reserves `1` for the
   historical scheme).
6. Author the `hash_version smallint NOT NULL DEFAULT 1` column migration and classify it via
   W02-E01-S001's manifest schema (online/maintenance classification, lock/statement timeout, N/N-1
   compatibility flag, backfill owner if applicable, validation query, rollback/forward-fix plan);
   run it through W02-E01's expand/backfill/validate/contract protocol as applicable.
7. Implement the widened `chainHash` and the version-branch dispatch in `Verify`, landing atomically
   with the migration per D-04's decision text.
8. Confirm whether `Anchor`/`CheckAnchor` require version-awareness of their own; implement if so.
9. Write the per-field tamper test: mutate `metadata`, `tx_id`, and every other declared field
   independently on a chained row; assert every one fails verification.
10. Write the version-branch verification test: a `hash_version = 1` row verifies under v1; a new row
    verifies under v2.
11. Document the widened field list, canonicalization approach, chosen `hash_version` value, and
    version-branch semantics.

### Expected files or components affected

`kernel/audit/audit.go` (chainHash, Verify, Anchor/CheckAnchor); a new migration file with its
manifest entry (via W02-E01's protocol tooling); new test files under `kernel/audit`'s test package.

### Expected output

A widened, canonicalized `chainHash`; a `hash_version` column shipped through W02-E01's protocol; a
version-branched `Verify`; a passing per-field tamper test and version-branch verification test;
documentation of the design.

### Required artifacts

ART-W04-E04-S001-001 (widened chainHash), ART-W04-E04-S001-002 (hash_version migration),
ART-W04-E04-S001-003 (version-branched Verify), ART-W04-E04-S001-004 (documentation).

### Required evidence

EV-W04-E04-S001-001 (per-field tamper-test report), EV-W04-E04-S001-002 (version-branch verification
report), EV-W04-E04-S001-003 (migration-classification report).

### Related acceptance criteria

AC-W04-E04-S001-01, AC-W04-E04-S001-02, AC-W04-E04-S001-03.

### Completion criteria

The widened `chainHash` and version-branched `Verify` are implemented and land atomically with the
`hash_version` migration; the per-field tamper test passes with every declared field independently
proven to break verification when mutated; the version-branch test passes for both the v1 and v2
branches; the migration has a complete, valid manifest entry and was executed within W02-E01's
lock-timeout budget.

### Verification method

Direct execution of the per-field tamper test and the version-branch verification test; inspection
of the migration's manifest entry and its execution record against W02-E01's protocol tooling.

### Risks

RISK-W04-002 (this task's confirmed highest-risk status in the epic's scope, breaking format change
hitting wowsociety's live audit rows) and RISK-W04-E04-S001-001 (an incorrect metadata
canonicalization could reintroduce non-reproducibility) — see epic-level `risks.md`.

### Rollback or recovery considerations

Given the migration and the widened hash land atomically, a rollback must be handled per W02-E01's
own protocol rollback/forward-fix discipline (the manifest's required field), not a bespoke code
revert — a naive revert alone would break verification for any new rows already written with the
widened scheme.

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

*Not yet implemented — the plan anticipates a `hash_version smallint NOT NULL DEFAULT 1` column
migration shipped through W02-E01's protocol; no migration has been authored or applied yet.*

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
| AC-W04-E04-S001-01 | Run per-field tamper test against a chained row | Local dev or CI, PostgreSQL instance | Every independent field mutation fails verification | tamper-test report | unassigned |
| AC-W04-E04-S001-02 | Run version-branch verification test (v1 historical row; v2 new row) | Local dev or CI, PostgreSQL instance with migration applied | Historical row verifies under v1; new row verifies under v2 | version-branch verification report | unassigned |
| AC-W04-E04-S001-03 | Inspect migration manifest entry and W02-E01 protocol execution record | Migration-manifest inspection + CI validation output | Manifest complete; migration executed through W02-E01's protocol within lock-timeout budget | migration-classification report | unassigned |

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
