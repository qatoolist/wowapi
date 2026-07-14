---
id: W04-E02-S003-T001
type: task
title: Locate and replace both hand-rolled retry implementations
status: done
parent_story: W04-E02-S003
owner: W04-Rerun
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W04-E02-S003-01
artifacts:
  - ART-W04-E02-S003-001
  - ART-W04-E02-S003-003
evidence: []
---

# W04-E02-S003-T001 — Locate and replace both hand-rolled retry implementations

## Task Definition

### Task objective

Locate the framework's two duplicated hand-rolled retry implementations and replace both with
`cenkalti/backoff/v5`, configured to match (or documented-ly improve upon) each prior
implementation's own retry schedule.

### Parent story

W04-E02-S003 — Adopt cenkalti/backoff/v5 for duplicated retry logic.

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Search the codebase for retry-loop patterns (backoff/sleep/attempt-counting logic not already
   using a shared library) to locate both hand-rolled implementations REVIEW §K refers to.
2. For each, document its current retry schedule (attempt count, backoff timing/growth, jitter if
   any, terminal behavior) as the parity baseline for T002's test.
3. Add `cenkalti/backoff/v5` as a direct `go.mod` dependency.
4. Replace the first hand-rolled implementation with a `cenkalti/backoff/v5`-configured equivalent.
5. Replace the second hand-rolled implementation the same way.
6. If either location is found inside a call site W04-E02-S001 is simultaneously restructuring
   (`kernel/notify`/`kernel/webhook`'s effect stage), coordinate with that story's implementer and
   record the coordination outcome in `deviations.md` if it changes either story's own plan.
7. Document both call sites' new retry configuration.

### Expected files or components affected

Not yet determinable — the exact files are this task's own first-step discovery (see `plan.md`
"Expected file changes where determinable").

### Expected output

Both hand-rolled retry implementations replaced with `cenkalti/backoff/v5`, with no hand-rolled
retry logic remaining at either original call site.

### Required artifacts

ART-W04-E02-S003-001 (`cenkalti/backoff/v5` integration at both call sites), ART-W04-E02-S003-003
(retry-configuration documentation).

### Required evidence

None directly — T002's parity/fault-injection tests provide this task's evidence.

### Related acceptance criteria

AC-W04-E02-S003-01.

### Completion criteria

Both hand-rolled retry implementations are confirmed replaced by direct code inspection, with
`cenkalti/backoff/v5` as the sole retry mechanism at both call sites.

### Verification method

Code-level inspection confirming no hand-rolled retry loop remains at either original location;
confirmation `cenkalti/backoff/v5` is imported and used at both.

### Risks

Locating both implementations is this task's own open discovery step — the source text names the
duplication pattern without exact locations; a missed location would leave one hand-rolled
implementation undetected and unreplaced, defeating this story's own purpose.

### Rollback or recovery considerations

Revert either replacement independently if it destabilizes the affected call site's retry behavior;
because each is a library-configuration swap, reverting either is expected to be low-risk.

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

*Not applicable — tests are T002's scope.*

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
| AC-W04-E02-S003-01 | Code-level inspection of both original call sites | Local dev or CI, Go toolchain | No hand-rolled retry logic remains; `cenkalti/backoff/v5` used at both | code-inspection report | unassigned |

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
