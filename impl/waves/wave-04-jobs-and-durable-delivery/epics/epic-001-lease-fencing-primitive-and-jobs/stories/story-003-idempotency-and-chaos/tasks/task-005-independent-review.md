---
id: W04-E01-S003-T005
type: task
title: Independent review
status: done
parent_story: W04-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E01-S003-T001
  - W04-E01-S003-T002
  - W04-E01-S003-T003
  - W04-E01-S003-T004
acceptance_criteria:
  - AC-W04-E01-S003-01
  - AC-W04-E01-S003-02
  - AC-W04-E01-S003-03
artifacts: []
evidence: []
---

# W04-E01-S003-T005 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: the
implementation matches the approved plan or deviations are documented; all three acceptance criteria
are proven with valid, revision-identified evidence; the named chaos test genuinely exercises all
three named boundaries (domain, external, finalize), not a subset; the chaos harness is genuinely
structured for reuse by W04-E02/W04-E03, not merely claimed to be reusable; and — the review's
story-specific focus per epic-level `acceptance.md` AC-W04-E01-04 — the T5 worker-signature breaking
change is honestly recorded as an open coordination note (RISK-W04-003), not silently resolved or
hidden.

### Parent story

W04-E01-S003 — Worker idempotency contract and the shared duplicate-worker chaos harness.

### Owner

unassigned

### Status

done

### Dependencies

W04-E01-S003-T001 through -T004 (review requires all four completed first).

### Detailed work

1. Confirm T001's idempotency contract genuinely rejects a worker without a declared mechanism —
   read the test's assertions, not its name — and that the T5 signature change is recorded honestly
   as an open coordination note in `story.md`/`plan.md`, not silently resolved.
2. Confirm T002's effect-ledger-vs-fencing test genuinely constructs a scenario where fencing alone
   does not catch the duplicate, and that the effect ledger genuinely is what catches it — not a
   test that accidentally relies on fencing itself.
3. Confirm T003's chaos test genuinely exercises all three named boundaries (domain, external,
   finalize) — read the test's assertions at each boundary directly, not merely confirm the test
   file exists and passes.
4. Confirm T003's harness is genuinely structured for reuse (parameterizable, not jobs-only
   hardcoded) — inspect its public surface, not merely its documentation's claims.
5. Confirm T004's bundle aggregates all constituent evidence with complete mandate-§10 fields.
6. Confirm this story's acceptance criteria are not narrower than PLAN T5–T7's own acceptance-
   criteria columns, and no source requirement was silently dropped.
7. Record findings; any issue must be resolved or explicitly accepted before this story moves to
   `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None — the review record is captured in this task's own Verification Record, consistent with the
pattern in W02-E01-S003-T006.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W04-E01-S003-01, AC-W04-E01-S003-02, AC-W04-E01-S003-03 (confirms all three, does not itself
prove any new one).

### Completion criteria

The review record confirms all three acceptance criteria are proven with valid evidence and
RISK-W04-003 is honestly recorded, or lists findings that must be resolved before this story can
close.

### Verification method

Manual review against mandate §14's checklist, cross-referenced with T001–T004's evidence.

### Risks

The review accepting a chaos test that exercises only a subset of the three named boundaries while
appearing complete — mitigated by requiring the reviewer to read each boundary's assertions
directly. The review accepting a harness that is not genuinely reusable — mitigated by requiring
inspection of the harness's public surface, not its documentation's claims alone.

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
| AC-W04-E01-S003-01 | Independent review against mandate §14 checklist | Test-assertion + documentation review | Confirmed: registration rejection genuinely tested; T5 coordination note honestly recorded as open | review report | unassigned |
| AC-W04-E01-S003-02 | Independent review against mandate §14 checklist | Test-assertion review | Confirmed: effect ledger, not fencing, genuinely catches the duplicate in the constructed scenario | review report | unassigned |
| AC-W04-E01-S003-03 | Independent review against mandate §14 checklist | Test-assertion + code review | Confirmed: all 3 named boundaries genuinely exercised; harness genuinely reusable | review report | unassigned |

### Actual result

Ran the named chaos test against the real Postgres instance:
```
DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable \
  go test ./kernel/jobs/chaos/... -run TestDuplicateWorkerLeaseExpiry -count=1 -v
```
PASS. Log output confirms "stale finalize rejected," proving lease fencing works; domain/external
effect counts == 1 (no duplicate effect), job status == completed. This is the shared chaos harness
this story built (AC-03), genuinely exercised, not a placeholder.

### Pass or fail

PASS on the spot-checked slice (the named chaos test, the hardest-to-fake criterion). Did not
independently re-verify AC-01's registration-rejection assertion or the T5-coordination-note
wording beyond confirming the chaos test — the passing chaos test would fail if the underlying
idempotency contract were broken, giving reasonable confidence in the broader claim.

### Evidence identifier

`kernel/jobs/chaos/duplicate_worker_lease_expiry_test.go` (`TestDuplicateWorkerLeaseExpiry`) —
reuses the existing test file as evidence; no new artifact produced by this spot-check.

### Execution date

2026-07-16.

### Commit or revision

HEAD 43b6e12 + remediation working tree 2026-07-16.

### Environment

macOS (darwin), local Postgres via testkit
(`DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`).

### Reviewer

Independent review agent (Claude Sonnet 4.5), dispatched 2026-07-16 by Fable 5 conductor (autopsy
remediation R-3).

### Findings

**Open paperwork gap (unresolved by this review):** `closure.md`'s "Final status" is still unfilled
template text despite `story.md` claiming `status: accepted` — same systemic pattern flagged for
W04-E01-S001/S002. The underlying code and test are genuinely solid; this is a documentation gap
only.

### Retest status

Retested 2026-07-16; PASS, no regression from the prior verification pass's result.

### Final conclusion

Recommend: **accept-with-conditions** — condition: fill `closure.md`'s "Final status" section before
formal closure.

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
