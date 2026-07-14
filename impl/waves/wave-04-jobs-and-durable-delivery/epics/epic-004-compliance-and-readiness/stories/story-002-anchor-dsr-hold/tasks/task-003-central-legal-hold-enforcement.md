---
id: W04-E04-S002-T003
type: task
title: Central legal-hold enforcement wrapper
status: done
parent_story: W04-E04-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W04-E04-S002-03
  - AC-W04-E04-S002-04
artifacts:
  - ART-W04-E04-S002-003
  - ART-W04-E04-S002-004
  - ART-W04-E04-S002-006
evidence:
  - EV-W04-E04-S002-003
  - EV-W04-E04-S002-004
---

# W04-E04-S002-T003 — Central legal-hold enforcement wrapper

## Task Definition

### Task objective

Enumerate every currently-registered `RecordClass`/`Dispose`/`Erase` callback in both wowapi and
wowsociety, then build a central legal-hold enforcement wrapper every registered callback must pass
through, replacing today's per-callback responsibility, and prove with a negative test that a
deliberately non-compliant callback is still blocked.

### Parent story

W04-E04-S002 — External anchoring, DSR export artifact, central legal-hold, and explicit per-class
status.

### Owner

unassigned

### Status

done

### Dependencies

None at task level, but internally sequenced: the enumeration step (below) must complete before the
wrapper implementation begins, per PLAN DATA-08 W6-T4's own risk note.

### Detailed work

1. Enumerate every currently-registered `RecordClass` and its `Dispose`/`Erase` callback in both
   wowapi and wowsociety, per PLAN's own risk note: "Breaking change to the `DisposeFunc`/`EraseFunc`
   contract — enumerate every registered `RecordClass` in both repos first." This enumeration must
   complete and be reviewed before step 3 begins.
2. Determine the exact new `DisposeFunc`/`EraseFunc` contract shape (wrapper-injected parameter, a
   required registration-time declaration, or fully transparent interposition) that minimizes
   unnecessary breakage while guaranteeing the wrapper cannot be bypassed.
3. Implement the central legal-hold enforcement wrapper, interposed so every registered callback
   passes through it regardless of the callback's own internal hold-check correctness.
4. Write the negative test: register a deliberately non-compliant callback (one with no internal hold
   check) and confirm the wrapper still blocks it.
5. Document the wrapper's contract and how a `Dispose`/`Erase` callback registers against it.

### Expected files or components affected

The `Dispose`/`Erase` callback registration mechanism's own source file(s) (exact location TBD,
expected near the existing registration code); a new test file for the legal-hold negative test.

### Expected output

A completed `RecordClass` enumeration record (both repos); a working central legal-hold enforcement
wrapper; a passing negative test; documentation of the wrapper's contract.

### Required artifacts

ART-W04-E04-S002-003 (central legal-hold enforcement wrapper), ART-W04-E04-S002-004 (RecordClass
callback enumeration record), ART-W04-E04-S002-006 (documentation, shared with T001/T002/T004).

### Required evidence

EV-W04-E04-S002-003 (legal-hold negative-test report), EV-W04-E04-S002-004 (RecordClass enumeration
record).

### Related acceptance criteria

AC-W04-E04-S002-03, AC-W04-E04-S002-04 (enumeration half).

### Completion criteria

The enumeration record is complete and predates the wrapper's implementation; every registered
callback passes through the wrapper; the negative test confirms a deliberately non-compliant callback
is still blocked.

### Verification method

Direct execution of the negative test; inspection of the enumeration record's completeness and its
timestamp/commit relative to the wrapper's own implementation commit.

### Risks

RISK-W04-E04-001 (the breaking `DisposeFunc`/`EraseFunc` contract change) — see epic-level
`risks.md`. An incomplete enumeration risks the wrapper silently missing a currently-registered
callback, either failing closed (blocking a legitimate dispose/erase) or failing open (defeating the
wrapper's purpose).

### Rollback or recovery considerations

The wrapper is additive interposition, not a destructive schema change — a code-level revert is
expected to be sufficient if it produces false positives blocking a legitimate callback (see
`story.md` "Rollback strategy").

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

*Not applicable unless the wrapper requires its own audit trail table — TBD at implementation time.*

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
| AC-W04-E04-S002-03 | Run legal-hold negative test | Local dev or CI | Non-compliant callback is still blocked by the wrapper | legal-hold negative-test report | unassigned |
| AC-W04-E04-S002-04 (enumeration half) | Inspect the RecordClass enumeration record's completeness and timing | Documentation review | Enumeration complete across both repos, predates wrapper implementation | enumeration record | unassigned |

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
