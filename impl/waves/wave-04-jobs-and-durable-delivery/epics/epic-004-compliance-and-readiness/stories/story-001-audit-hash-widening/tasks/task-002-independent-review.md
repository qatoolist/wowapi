---
id: W04-E04-S001-T002
type: task
title: Independent review
status: done
parent_story: W04-E04-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E04-S001-T001
acceptance_criteria:
  - AC-W04-E04-S001-01
  - AC-W04-E04-S001-02
  - AC-W04-E04-S001-03
artifacts: []
evidence: []
---

# W04-E04-S001-T002 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: the
implementation matches the approved plan or deviations are documented; all three acceptance criteria
are proven with valid, revision-identified evidence; and — the review's story-specific focus per
epic-level `acceptance.md` AC-W04-E04-04 — the per-field tamper test genuinely covers every declared
field independently, and D-04's version-branch design was implemented exactly as ratified, not a
divergent interpretation silently substituted.

### Parent story

W04-E04-S001 — Audit hash-chain widening with hash_version discriminator.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E04-S001-T001 (review requires the implementation task completed first).

### Detailed work

1. Confirm the per-field tamper test's assertions, not its name — read the test to verify it
   independently mutates every declared field (metadata, tx_id, and each of the previously-covered
   15 fields), not a single combined mutation or a subset.
2. Confirm the metadata canonicalization function hashes a genuinely reproducible pre-serialization
   form, never the stored jsonb directly — re-derive this from the code, not from documentation
   claims alone.
3. Confirm the `hash_version` column was added in the same migration as the `chainHash` widening (per
   D-04's decision text), not as two separately-sequenced changes.
4. Confirm `Verify`'s version-branch dispatch correctly routes `hash_version = 1` rows to the
   original 15-field scheme and new rows to the widened scheme, and that a row with an unrecognized
   `hash_version` fails closed rather than silently defaulting to one branch.
5. Confirm the migration was genuinely classified and executed through W02-E01's protocol (manifest
   entry present and complete, lock-timeout budget honored), not shipped as an ad hoc one-off despite
   this story's stated dependency on W02-E01.
6. Confirm this story's acceptance criteria are not narrower than PLAN DATA-08 W6-T1's own
   acceptance-criteria and Tests columns, and that D-04's decision text was not silently reinterpreted
   or weakened during implementation.
7. Record findings; any issue must be resolved or explicitly accepted before this story moves to
   `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None — the review record is captured in this task's own Verification Record, consistent with the
pattern in W02-E01-S001-T003 and W02-E01-S003-T006.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W04-E04-S001-01 through -03 (confirms all three, does not itself prove any new one).

### Completion criteria

The review record confirms all three acceptance criteria are proven with valid evidence and D-04's
design was implemented exactly as ratified, or lists findings that must be resolved before this
story can close.

### Verification method

Manual review against mandate §14's checklist, cross-referenced with T001's evidence and with D-04's
own decision text.

### Risks

The review accepting a weakened or partial per-field tamper test (step 1's concern) — mitigated by
requiring the reviewer to read the test's assertions directly, field by field. Given this story's
confirmed highest-risk status in the epic, this review carries materially higher stakes than a
routine independent-review task; the reviewer must not treat it as a formality.

### Rollback or recovery considerations

Not applicable — a review task has no code to roll back; a failed review blocks story acceptance
until its findings are resolved.

## Implementation Record

*Not applicable — this is a review task, not an implementation task.*

### What was actually implemented

*Not applicable.*

### Components changed

*Not applicable.*

### Files changed

*Not applicable.*

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not applicable.*

### Commits

*Not applicable.*

### Pull requests

*Not applicable.*

### Implementation dates

*Not applicable.*

### Technical debt introduced

*Not applicable.*

### Known limitations

*Not applicable.*

### Follow-up items

*Not yet executed.*

### Relationship to the approved plan

*Not applicable.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W04-E04-S001-01 | Independent review against mandate §14 checklist | Test-assertion review | Confirmed: per-field tamper test genuinely covers every declared field independently | review report | unassigned |
| AC-W04-E04-S001-02 | Independent review against mandate §14 checklist | Code + test-assertion review | Confirmed: version-branch dispatch correct; unrecognized hash_version fails closed | review report | unassigned |
| AC-W04-E04-S001-03 | Independent review against mandate §14 checklist | Manifest + CI record inspection | Confirmed: migration classified and executed through W02-E01's protocol | review report | unassigned |

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
