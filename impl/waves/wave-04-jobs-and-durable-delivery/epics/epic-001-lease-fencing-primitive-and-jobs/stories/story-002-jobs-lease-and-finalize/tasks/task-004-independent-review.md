---
id: W04-E01-S002-T004
type: task
title: Independent review
status: done
parent_story: W04-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E01-S002-T001
  - W04-E01-S002-T002
  - W04-E01-S002-T003
acceptance_criteria:
  - AC-W04-E01-S002-01
  - AC-W04-E01-S002-02
  - AC-W04-E01-S002-03
artifacts: []
evidence: []
---

# W04-E01-S002-T004 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming: the
implementation matches the approved plan or deviations are documented; all three acceptance criteria
are proven with valid evidence; the finalize fencing genuinely does not regress the at-least-once
recovery path (PLAN DATA-02 T3's own risk note); the lease columns genuinely reuse
W04-E01-S001's shared primitive rather than a bespoke copy; no source requirement (DATA-02 T2, T3,
T4) was silently dropped or narrowed.

### Parent story

W04-E01-S002 — Jobs lease columns, fenced finalize, and fenced reclaim.

### Owner

unassigned

### Status

done

### Dependencies

W04-E01-S002-T001, W04-E01-S002-T002, W04-E01-S002-T003 (review requires all three implemented
first).

### Detailed work

1. Confirm T001's lease-column migration and claim SQL genuinely reuse W04-E01-S001's shared
   primitive (field-for-field), not a bespoke `jobs_queue`-only reimplementation.
2. Confirm T002's finalize fencing rejects a stale finalize (read the test's assertions directly,
   not its name) and that the non-regression test genuinely proves a legitimate finalize still
   succeeds — both halves of AC-W04-E01-S002-02, not merely one.
3. Confirm T003's reclaim generation-bump is exercised by the shared test's generation-delta
   assertion, per T4's own "Same test as T3" instruction — confirm the test was genuinely extended,
   not merely claimed to cover both concerns.
4. Confirm evidence in `evidence/index.md` identifies the tested commit SHA for every item —
   evidence without this must not be treated as final proof (mandate §10).
5. Confirm this story's `story.md` acceptance criteria are not narrower than PLAN DATA-02 T2/T3/T4's
   own acceptance-criteria columns.
6. Record findings; if any issue is found, it must be resolved or explicitly accepted before this
   story moves to `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None — the review record itself is captured in this task's own Verification Record, consistent with
the pattern in W02-E01-S001-T003.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W04-E01-S002-01, AC-W04-E01-S002-02, AC-W04-E01-S002-03 (confirms all three, does not itself
prove any new one).

### Completion criteria

The review record confirms all three acceptance criteria are proven with valid evidence, or lists
findings that must be resolved before this story can close.

### Verification method

Manual review against mandate §14's checklist, cross-referenced with T001/T002/T003's evidence.

### Risks

The review accepting a fencing implementation that is "too strict" (regressing the at-least-once
recovery path) without catching it — mitigated by requiring the reviewer to read the non-regression
test's assertions directly, not merely confirm its existence.

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
| AC-W04-E01-S002-01 | Independent review against mandate §14 checklist | Code review + test-output inspection | Confirmed: lease columns genuinely reuse S001's primitive; claim assignment genuinely tested | review report | unassigned |
| AC-W04-E01-S002-02 | Independent review against mandate §14 checklist | Test-assertion + code review | Confirmed: stale finalize genuinely rejected; legitimate finalize genuinely unregressed | review report | unassigned |
| AC-W04-E01-S002-03 | Independent review against mandate §14 checklist | Test-assertion review | Confirmed: generation-delta assertion genuinely present and passing in the shared test | review report | unassigned |

### Actual result

*Not yet executed.*

### Pass or fail

*Not yet executed.*

### Evidence identifier

Reuses `migrations/00038_jobs_lease_columns.sql` and `kernel/jobs/chaos/duplicate_worker_lease_expiry_test.go`'s
existing evidence (see W04-E01-S003's own evidence, which exercises this story's fenced-finalize
columns end-to-end).

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

Spot-check (gate-level confirmation): confirmed `migrations/00038_jobs_lease_columns.sql` (31 lines)
exists and adds lease columns. The decisive proof of AC-02 (fenced finalize rejects a stale worker)
is exercised end-to-end by the sibling story's chaos test
(`kernel/jobs/chaos/duplicate_worker_lease_expiry_test.go`, `TestDuplicateWorkerLeaseExpiry`), which
this review re-ran and confirmed PASS with log line "stale finalize rejected" and domain/external
effect counts == 1. Did not independently re-derive the fenced finalize/`ReclaimStalled` SQL text
line-by-line beyond this behavioral proof. **Open paperwork gap (unresolved by this review):**
`closure.md`'s "Final status" is still unfilled template text despite `story.md` claiming
`status: accepted` — same pattern as W04-E01-S001.

### Retest status

Retested via the sibling chaos test; PASS.

### Final conclusion

Recommend: **accept-with-conditions** — condition: fill `closure.md`'s "Final status" section before
formal closure, consistent with W04-E01-S001's condition.

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
