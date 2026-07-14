---
id: W02-E04-S001-T002
type: task
title: Actor-attribution fix (single owner, shared with DATA-07 T3)
status: todo
parent_story: W02-E04-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W02-E04-S001-T001
acceptance_criteria:
  - AC-W02-E04-S001-02
artifacts:
  - ART-W02-E04-S001-002
evidence:
  - EV-W02-E04-S001-002
---

# W02-E04-S001-T002 — Actor-attribution fix (single owner, shared with DATA-07 T3)

## Task Definition

### Task objective

Source `created_by` from context inside the T1 helper, replacing `registrar_pg.go`'s current
`uuid.Nil` placeholder; reject a missing actor for a user-initiated write; leave system-actor paths
unaffected.

**Single-owner note (must not be silently dropped or reimplemented elsewhere):** per PLAN's own
PF-DATA cross-cutting note (2), this exact fix ("`kernel/resource/registrar_pg.go`'s nil-actor
placeholder") is claimed by two findings: this task (DATA-06 T2) and DATA-07 T3, "one owner, not two
PRs." This task is that one owner. When W03-E04-S001 (DATA-07 T3, out of this story's scope) is
later implemented, it must consume this task's mechanism directly rather than reimplementing an
independent, possibly-divergent nil-actor fix at the same file. This task's implementation record,
once populated, should reference the exact final location/shape of the fix precisely enough for a
W03 implementer to find and reuse it without re-reading this entire task file.

### Parent story

W02-E04-S001 — Typed aggregate write contract with mandatory mirror, audit, and outbox.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E04-S001-T001 (the actor-attribution fix lives inside the same helper T001 builds).

### Detailed work

1. Re-read `registrar_pg.go` at this task's actual start commit to re-confirm the exact current
   line range of the `uuid.Nil` placeholder (PLAN cites `:38-58` as of the source document's
   writing).
2. Implement context-sourced actor resolution inside the T1 helper, replacing the `uuid.Nil`
   placeholder with a real, resolved actor.
3. Implement fail-fast rejection of a missing actor specifically for user-initiated writes — do not
   reject a system-actor path (e.g. a scheduled job) that legitimately has no user-initiated actor.
4. Write a test covering: (a) a user-initiated write with an actor present, succeeding with the real
   `created_by`; (b) a user-initiated write with no actor, failing fast; (c) a system-actor path,
   succeeding unaffected.
5. Confirm no existing legitimate system-actor call site is broken by this change (PLAN T2's own
   named risk: "Must not break legitimate system-actor call sites") — audit existing system-actor
   call sites as part of this task's verification, not merely trust the new test in isolation.
6. Record the fix's final location and shape clearly in this task's Implementation Record, so
   DATA-07 T3 (W03-E04-S001) can consume it directly.

### Expected files or components affected

`kernel/resource/registrar_pg.go` (exact line range to be re-confirmed at implementation time).

### Expected output

A working actor-attribution fix inside the T1 helper, proven by the with/without-actor and
system-actor-path test, with the fix's location documented for DATA-07 T3's future consumption.

### Required artifacts

ART-W02-E04-S001-002 (`registrar_pg.go` actor-attribution fix).

### Required evidence

EV-W02-E04-S001-002 (actor-attribution unit-test report).

### Related acceptance criteria

AC-W02-E04-S001-02.

### Completion criteria

The with-actor case succeeds with real attribution, the without-actor user-initiated case fails
fast, and the system-actor path is confirmed unaffected — all three proven by a passing test suite
against a named commit SHA, with no regression in any existing system-actor call site.

### Verification method

Direct execution of the actor-attribution test suite; audit of existing system-actor call sites for
regression; logged output retained as evidence.

### Risks

PLAN T2's own named risk: "Must not break legitimate system-actor call sites." Secondary risk: if
this task's fix location or shape is not clearly documented, DATA-07 T3 (W03-E04-S001) may
reimplement it independently, violating the single-owner intent of PLAN's cross-cutting note (2).

### Rollback or recovery considerations

Revert this fix if it is found to break any legitimate system-actor call site not covered by this
task's own audit; escalate for a corrected fail-fast condition rather than silently loosening the
check to avoid the regression.

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

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not yet implemented — this task's own fix is itself the security-relevant change (real actor
attribution replacing an unattributed placeholder); recorded here once implemented.*

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

*Not yet implemented — expected follow-up: DATA-07 T3 (W03-E04-S001) consumes this fix once W03
begins; this is a forward reference, not a follow-up item owned by this task itself.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W02-E04-S001-02 | Run the actor-attribution test suite (with actor, without actor, system-actor path) | Local dev or CI | With-actor succeeds with real `created_by`; without-actor user-initiated write fails fast; system-actor path unaffected | unit-test report | unassigned |

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
