---
id: W04-E01-S002-T003
type: task
title: Fenced reclaim with generation bump
status: done
parent_story: W04-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E01-S002-T001
acceptance_criteria:
  - AC-W04-E01-S002-03
artifacts:
  - ART-W04-E01-S002-003
  - ART-W04-E01-S002-004
evidence:
  - EV-W04-E01-S002-003
---

# W04-E01-S002-T003 — Fenced reclaim with generation bump

## Task Definition

### Task objective

Extend `ReclaimStalled` to bump `lease_generation` on every row it resets, so a reclaimed row
provably belongs to a new lease epoch, distinguishable from the epoch the stalled worker was
operating under.

### Parent story

W04-E01-S002 — Jobs lease columns, fenced finalize, and fenced reclaim.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E01-S002-T001 (lease columns and fenced claim SQL must exist before reclaim can bump a
generation that exists).

### Detailed work

1. Re-read `kernel/jobs`'s `ReclaimStalled` implementation at this task's actual start commit to
   confirm it currently performs a blind reset with no generation bump and no per-row fencing check.
2. Extend `ReclaimStalled` to bump `lease_generation` on every row it reclaims.
3. Reuse (per PLAN DATA-02 T4's own instruction, "Same test as T3") T002's stale-finalize test,
   extending its assertions to additionally confirm the reclaimed row's `lease_generation` delta —
   do not write a separate, disconnected test.
4. Document the reclaim generation-bump behavior (this task's share of ART-W04-E01-S002-004).

### Expected files or components affected

`kernel/jobs`'s `ReclaimStalled` implementation (exact file path TBD per `plan.md`).

### Expected output

`ReclaimStalled` bumping `lease_generation` on every reclaimed row, proven by the same test as
T002's stale-finalize test, extended with a generation-delta assertion.

### Required artifacts

ART-W04-E01-S002-003 (fenced `ReclaimStalled` code), ART-W04-E01-S002-004 (documentation, shared
with T001/T002).

### Required evidence

EV-W04-E01-S002-003 (reclaim generation-delta test report — same underlying test as
EV-W04-E01-S002-002, per T4's "Same test as T3" instruction).

### Related acceptance criteria

AC-W04-E01-S002-03.

### Completion criteria

`ReclaimStalled` bumps `lease_generation` on every reclaimed row, producing a provably new lease
epoch — proven by the shared test's generation-delta assertion.

### Verification method

Direct execution of the shared stale-finalize/generation-delta test, specifically inspecting the
generation-delta assertion.

### Risks

None distinct from T002's own risk framing — this task extends the same test rather than
introducing a separate risk surface. The only material risk is failing to actually bump the
generation on every reclaimed row (a partial or conditional bump would undermine the "provably new
lease epoch" guarantee).

### Rollback or recovery considerations

Revert the generation-bump change if it is found to bump generations on rows that were not actually
stalled/reclaimed (a false-positive reclaim) — escalate for redesign rather than silently narrowing
the reclaim condition without recording why.

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

*Not applicable.*

### Observability changes

*Not yet implemented.*

### Tests added or modified

*Not yet implemented — expected: an extension of T002's shared test, not a new standalone test.*

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
| AC-W04-E01-S002-03 | Run the shared stale-finalize/generation-delta test (same test as T002), asserting the reclaim generation delta | Local dev or CI, PostgreSQL instance | `ReclaimStalled` bumps `lease_generation`; delta is provable | integration-test report | unassigned |

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
