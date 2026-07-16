---
id: W02-E01-S002-T002
type: task
title: Backfill-job harness and interim checkpoint-lease mechanism
status: done
parent_story: W02-E01-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W02-E01-S002-T001
acceptance_criteria:
  - AC-W02-E01-S002-02
artifacts:
  - ART-W02-E01-S002-002
  - ART-W02-E01-S002-004
evidence:
  - EV-W02-E01-S002-002
---

# W02-E01-S002-T002 — Backfill-job harness and interim checkpoint-lease mechanism

## Task Definition

### Task objective

Implement the backfill-job harness — resumable, tenant-scoped, keyset-paginated, checkpointed, with
bounded batch/tx time and rate controls — together with the interim checkpoint-lease mechanism that
substitutes for DATA-02 T1's not-yet-built shared lease primitive, and prove the harness with the
explicitly-required interrupted/resumed backfill test: no reprocessing, no skipping.

### Parent story

W02-E01-S002 — Expand-phase tooling, resumable backfill harness, and validation-phase tooling.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E01-S002-T001 (PLAN T4's own "Depends-on" column names T3). PLAN T4 also names "DATA-02 T1" as
a dependency — that dependency is unresolvable at this task's execution time (DATA-02 is W04 scope)
and is resolved by this task's own interim-lease design per the canonical allocation in
`impl/analysis/wave-allocation-detail.md`: "S002 builds a minimal checkpoint lease and W04-E01
replaces it — record as planned deviation-risk" (RISK-W02-001).

### Detailed work

1. **Design the interim checkpoint-lease's scope boundary first.** The mechanism provides exactly
   what the backfill harness needs — a checkpoint token and resumability semantics — and explicitly
   nothing more: no fencing generations, no heartbeats, no job-claim semantics (those are DATA-02
   T1's full-primitive scope, W04-E01-S001). Document this boundary in the mechanism's own code
   comments and this story's documentation, so W04-E01-S001's implementer and any interim reader
   cannot mistake it for the full primitive. This scope-bounding is the mitigation for RISK-W02-001
   and is what the story's independent review (T004) specifically checks.
2. Implement the interim checkpoint-lease mechanism, including its persistence (checkpoint-state
   record: backfill job ID, last-processed keyset position, checkpoint token — exact schema per
   `plan.md`'s "Unresolved questions," to be settled here).
3. Implement the backfill-job harness: resumable (via the checkpoint lease), tenant-scoped,
   keyset-paginated, checkpointed, with bounded batch/transaction time and configurable rate
   controls (the specific batch/rate/window values remain per-migration human decisions per PLAN
   T4's own classification column — this task builds the configuration surface, not the values).
4. Write the named interrupted/resumed backfill test (`DATA-09/backfill-interrupt-resume/` per
   PLAN's evidence column): start a backfill, interrupt it mid-run, resume it, and assert that no
   row is reprocessed and no row is skipped across the interruption boundary. PLAN's own "Tests"
   column: "This is the test" — the test is the acceptance criterion, not an accessory to it.
5. Consider (and document the outcome of) whether the interim lease's interface can anticipate a
   clean expansion path to DATA-02 T1's full primitive, minimizing W04-E01-S001's migration cost —
   per `plan.md`'s "Unresolved questions."

### Expected files or components affected

New backfill-harness package and interim checkpoint-lease mechanism (exact locations TBD per
`plan.md`); a new table or columns for checkpoint state (schema TBD).

### Expected output

A backfill harness that passes the interrupted/resumed test, built on a deliberately scope-bounded
interim checkpoint lease with a documented forward reference to W04-E01-S001.

### Required artifacts

ART-W02-E01-S002-002 (backfill harness + interim lease), ART-W02-E01-S002-004 (documentation
including the interim-lease scope-boundary note, shared with T001/T003).

### Required evidence

EV-W02-E01-S002-002 (interrupted/resumed backfill test output).

### Related acceptance criteria

AC-W02-E01-S002-02.

### Completion criteria

The interrupted/resumed backfill test passes (no reprocessing, no skipping), evidenced against a
named commit SHA; the interim lease's scope boundary and its W04-E01-S001 forward reference are
documented in code and story documentation.

### Verification method

Direct execution of the interrupted/resumed test against a live PostgreSQL instance; documentation
inspection for the scope-boundary note.

### Risks

PLAN T4's own risk column: "Largest risk surface in DATA-09." Plus RISK-W02-001 (the interim lease
is planned technical debt until W04-E01-S001 replaces it) — this task is where that risk's
mitigation (scope-bounding) is actually executed. If the minimal lease scope proves insufficient
for the interrupted/resumed test to pass, that is a deviation to record in the story's
`deviations.md`, not a silent scope expansion.

### Rollback or recovery considerations

The harness's own checkpoint semantics are its recovery mechanism — an interrupted backfill must
leave checkpoint state consistent and resumable (no partial-batch commit causing reprocessing or
skipping on resume). Reverting the harness itself is a code revert with no data impact, since this
task executes no real production backfill.

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

*Not yet implemented — a checkpoint-state table/columns are anticipated (see Detailed work step 2);
recorded here once actually implemented.*

### Security changes

*Not applicable.*

### Observability changes

*Not yet implemented — queryable checkpoint state is anticipated per the story's "Observability
considerations."*

### Tests added or modified

*Not yet implemented.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*Not yet implemented — the interim checkpoint lease is anticipated technical debt per RISK-W02-001,
to be formally recorded here (with the W04-E01-S001 pointer) once implemented.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E01-S002-02 | Named interrupted/resumed backfill test: interrupt mid-run, resume | Local dev or CI, PostgreSQL, backfill harness under test | No row reprocessed, no row skipped; interim-lease scope boundary documented | integration-test report + documentation inspection | unassigned |

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

*No deviations recorded yet. The interim checkpoint lease itself is the approved plan, not a
deviation (see the story's `deviations.md` for the distinction); a deviation would be, e.g., the
minimal lease scope proving insufficient and requiring unplanned expansion.*

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
