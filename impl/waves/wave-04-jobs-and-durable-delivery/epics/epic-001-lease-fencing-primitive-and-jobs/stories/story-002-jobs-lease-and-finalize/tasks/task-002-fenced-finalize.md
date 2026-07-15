---
id: W04-E01-S002-T002
type: task
title: Fenced finalize paths
status: done
parent_story: W04-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E01-S002-T001
acceptance_criteria:
  - AC-W04-E01-S002-02
artifacts:
  - ART-W04-E01-S002-002
  - ART-W04-E01-S002-004
evidence:
  - EV-W04-E01-S002-002
---

# W04-E01-S002-T002 — Fenced finalize paths

## Task Definition

### Task objective

Extend the `complete`/`fail` finalize code paths to compare the caller's lease token/generation
against the row's current lease state and reject a mismatch, so a stale finalize affects zero rows
and is observably rejected, without regressing the existing at-least-once recovery path for a
legitimate (non-superseded) worker.

### Parent story

W04-E01-S002 — Jobs lease columns, fenced finalize, and fenced reclaim.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E01-S002-T001 (lease columns and fenced claim SQL must exist before finalize can compare
against lease state).

### Detailed work

1. Re-read `kernel/jobs`'s `complete`/`fail` finalize code paths at this task's actual start commit
   to confirm they currently match only by `id`, with no lease comparison.
2. Extend the finalize function signatures to accept the caller's lease context (token/generation).
3. Implement the comparison: a mismatch is rejected observably (choose and document the rejection-
   surfacing mechanism — returned error, logged event, or both, per `plan.md`'s "Unresolved
   questions").
4. Write a test proving a stale finalize (simulating a since-reclaimed lease epoch) affects zero
   rows and is observably rejected.
5. Write a positive-case test proving a legitimate, non-superseded finalize still succeeds exactly
   as before fencing was introduced — the at-least-once recovery path must not regress (PLAN DATA-02
   T3's own risk note).
6. Document the fencing/rejection behavior (this task's share of ART-W04-E01-S002-004).

### Expected files or components affected

`kernel/jobs`'s `complete`/`fail` finalize code paths (exact file path TBD per `plan.md`).

### Expected output

Fenced finalize paths rejecting a stale token/generation mismatch observably, with the legitimate
recovery path unregressed.

### Required artifacts

ART-W04-E01-S002-002 (fenced finalize code), ART-W04-E01-S002-004 (documentation, shared with
T001/T003).

### Required evidence

EV-W04-E01-S002-002 (stale-finalize rejection + non-regression test report).

### Related acceptance criteria

AC-W04-E01-S002-02.

### Completion criteria

A stale finalize affects zero rows and is observably rejected; a legitimate finalize still succeeds
unregressed — both proven by passing tests.

### Verification method

Direct execution of the stale-finalize test and the non-regression positive-case test.

### Risks

Must not regress the at-least-once recovery path — PLAN DATA-02 T3's own risk note. A fencing
implementation that is "too strict" and rejects legitimate, non-superseded finalizes would be a
correctness defect as serious as the race this task closes.

### Rollback or recovery considerations

Revert the finalize fencing if it produces false-positive rejections against legitimate finalizes
under real load; escalate for redesign rather than silently loosening the comparison without
recording why.

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

*Not yet implemented — the fenced finalize path is itself the security control; recorded here once
implemented.*

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
| AC-W04-E01-S002-02 | Run stale-finalize rejection test + legitimate-finalize non-regression test | Local dev or CI, PostgreSQL instance | Stale finalize affects 0 rows, observably rejected; legitimate finalize succeeds unregressed | integration-test report | unassigned |

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
