---
id: W05-E01-S003-T002
type: task
title: Post-seal Context/registrar retention rejection
status: todo
parent_story: W05-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E01-S003-02
artifacts:
  - ART-W05-E01-S003-002
evidence:
  - EV-W05-E01-S003-002
---

# W05-E01-S003-T002 — Post-seal Context/registrar retention rejection

## Task Definition

### Task objective

Reject Context/registrar retention after `Register()` returns: a module retaining `ctx` or a
registrar post-boot gets an explicit error on mutation, never a silent no-op or a production panic,
validated specifically against wowsociety's `internal/modules/policy/pack.go:334-338`'s retained
`s.rulesReg` field as a named real-world consumer, without falsely rejecting the legitimately-used
`s.rulesStore`/`s.rulesResolver` pattern.

### Parent story

W05-E01-S003 — Snapshot immutability, post-seal rejection, model hash, and race safety.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T001 — disjoint concern); depends on W05-E01-S001 (T1,
T2, D-03's error-not-panic mechanism) at story scope.

### Detailed work

1. Extend S001's D-03 error-not-panic mechanism specifically to registrar retention: a module
   retaining a registrar or `ctx` past `Register()` returning gets an explicit error, never a silent
   no-op or production panic, on any post-boot mutation attempt.
2. Write `AR-01/post_seal_mutation_rejection_test.go`: a fixture module retains a registrar, calls
   it post-boot, and receives an explicit error.
3. Write a sub-test modeled on wowsociety's `s.rulesReg` pattern (retained, never read again),
   confirming rejection.
4. Write a sub-test modeled on wowsociety's `s.rulesStore`/`s.rulesResolver` pattern (built over the
   registry, used live in request handlers), confirming this pattern is NOT falsely rejected.
5. Document the retention-rejection contract and the two named patterns it must distinguish.

### Expected files or components affected

The `ApplicationModel`/`Compiler` from S001 (post-seal rejection enforcement point) — exact file
paths TBD per `plan.md`.

### Expected output

A retained registrar/ctx used post-boot receives an explicit error; the wowsociety-pattern-modeled
sub-tests confirm the dead-retention-vs-live-use distinction holds.

### Required artifacts

ART-W05-E01-S003-002.

### Required evidence

EV-W05-E01-S003-002.

### Related acceptance criteria

AC-W05-E01-S003-02.

### Completion criteria

The retention-rejection test passes for both the dead-retention and live-use sub-tests — proving the
mechanism is neither too permissive (silently no-ops) nor too aggressive (rejects legitimate live
use).

### Verification method

Direct execution of `AR-01/post_seal_mutation_rejection_test.go`, including both named sub-tests.

### Risks

Medium, per PLAN T8's own risk column — "wowsociety's `policy` module already retains `mc.Rules()`
today (harmlessly); this task has a direct named consumer to validate against." A false-positive
rejection of the live-use pattern would be a compatibility regression against a real,
already-shipping consumer.

### Rollback or recovery considerations

If the live-use sub-test fails (i.e. the mechanism incorrectly rejects `s.rulesStore`/
`s.rulesResolver`), do not ship — revise the retention-detection mechanism to correctly distinguish
retained-and-called-post-boot from built-over-and-used-live before proceeding.

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

*Not yet implemented — extends D-03's error-not-panic guarantee; recorded here once implemented.*

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
| AC-W05-E01-S003-02 | Run `AR-01/post_seal_mutation_rejection_test.go` (both sub-tests) | Local dev or CI, Go toolchain | Dead-retention pattern rejected; live-use pattern not falsely rejected | adversarial-test report | unassigned |

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
