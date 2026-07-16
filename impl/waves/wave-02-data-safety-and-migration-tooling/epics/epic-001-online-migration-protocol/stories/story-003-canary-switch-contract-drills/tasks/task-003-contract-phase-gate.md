---
id: W02-E01-S003-T003
type: task
title: Contract-phase gate
status: done
parent_story: W02-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W02-E01-S003-T002
acceptance_criteria:
  - AC-W02-E01-S003-03
artifacts:
  - ART-W02-E01-S003-003
  - ART-W02-E01-S003-006
evidence:
  - EV-W02-E01-S003-003
---

# W02-E01-S003-T003 — Contract-phase gate

## Task Definition

### Task objective

Implement contract-phase tooling gated on an evidenced no-N-1-remains precondition, and prove, via
the named contract-gate test, both explicitly-required properties: forward recovery from every
failed phase, and delayed-contract-only-after-old-process-absence-proven (PLAN DATA-09 T8's bolded
acceptance criterion — "Most safety-critical piece — running contract too early is destructive and
hard to detect pre-outage" per T8's own risk column).

### Parent story

W02-E01-S003 — Canary, switch, and contract-phase tooling with the full CI drill pipeline.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E01-S003-T002 (PLAN T8's "Depends-on" column names T7 — contract follows switch).

### Detailed work

1. Define what constitutes admissible evidence of N-1 absence (version-tagged connection/process
   registry, deploy-system attestation, or another mechanism — `plan.md`'s "Unresolved questions").
2. Implement the contract gate: the contract phase refuses to run unless that evidence is present
   and unambiguous — fail closed on missing or ambiguous evidence, never default-allow.
3. Implement/verify forward-recovery paths from every failed phase (expand, backfill, validate,
   canary, switch, contract) — a failure at any phase leaves a recoverable, well-defined state with
   a documented forward path.
4. Write the named contract-gate test (`DATA-09/contract-gate/`) covering both required properties:
   forward recovery exercised from each failed phase, and contract provably blocked until N-1
   absence is evidenced.
5. Document the gate's evidence requirements and the human boundary: "human sign-off strongly
   advisable even with the gate passing" (PLAN T8's own classification column).

### Expected files or components affected

New contract-gate package (location TBD per `plan.md`).

### Expected output

A fail-closed contract gate with proven forward recovery, passing the named contract-gate test on
both required properties.

### Required artifacts

ART-W02-E01-S003-003 (contract gate), ART-W02-E01-S003-006 (documentation, shared).

### Required evidence

EV-W02-E01-S003-003 (named contract-gate test output, both properties).

### Related acceptance criteria

AC-W02-E01-S003-03.

### Completion criteria

The named contract-gate test passes both required properties, evidenced against a named commit SHA;
the gate demonstrably fails closed on missing/ambiguous evidence.

### Verification method

Direct execution of the named contract-gate test, including negative cases (missing evidence,
ambiguous evidence) proving the fail-closed posture.

### Risks

PLAN T8's own risk column — a gate that passes on ambiguous evidence is worse than no gate, because
it launders a destructive decision through apparent mechanical approval. The fail-closed negative
cases in step 4's test are the control; the independent review (T006) checks them specifically.

### Rollback or recovery considerations

Forward recovery is this task's own deliverable (contract is deliberately not reversible — that is
why the gate exists). For the task's own code: plain revert, no data impact.

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

*Not applicable — the gate tooling adds no application schema.*

### Security changes

*Not yet implemented — the fail-closed evidence check is this task's own safety control; recorded
here once implemented.*

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
| AC-W02-E01-S003-03 | Named contract-gate test incl. negative (missing/ambiguous evidence) cases | Local dev or CI, PostgreSQL | Forward recovery from every failed phase; contract blocked until absence evidenced; fail-closed proven | integration-test report | unassigned |

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
