---
id: W03-E04-S001-T003
type: task
title: Mutation governance - ownership, attribution, audit, versioning (DATA-07 T4)
status: done
parent_story: W03-E04-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W03-E04-S001-T001
  - W03-E04-S001-T002
acceptance_criteria:
  - AC-W03-E04-S001-03
artifacts:
  - ART-W03-E04-S001-003
evidence:
  - EV-W03-E04-S001-003
---

# W03-E04-S001-T003 — Mutation governance: ownership, attribution, audit, versioning (DATA-07 T4)

## Task Definition

### Task objective

Ensure every authorization-input mutation (edge create/revoke) is ownership-checked, attributed (via
DATA-06 T2's mechanism, consumed not reimplemented), audited, and versioned. The cache-invalidation
sub-criterion is deferred-linked to W05-E04-S002 (SEC-04's epoch table, D-06).

### Parent story

W03-E04-S001 — Relationship semantics — party-subject evaluation, full subject-kind matrix, mutation
governance.

### Owner

unassigned

### Status

todo

### Dependencies

W03-E04-S001-T001, W03-E04-S001-T002 — PLAN's own Depends-on column for T4: "T1-T3" (T3 itself is
out of this story's scope; T001/T002 here are this story's T1/T2 equivalents). **Soft external
dependency: DATA-06 T2 (W02-E04-S001)'s actor-attribution mechanism in `registrar_pg.go` must be
available to consume — if not yet landed, this task is blocked on that specific input, not worked
around by an independent reimplementation.** **Deferred-link: W05-E04-S002 (SEC-04's epoch table,
D-06) — this task's cache-invalidation sub-criterion depends on it, per PLAN's own T4 Depends-on
column: "also depends on SEC-04's cache-epoch work."** Both dependencies are restated here per this
story's own design goal that a task-level reader cannot miss either by reading only this file.

### Detailed work

1. Confirm DATA-06 T2 (W02-E04-S001) has landed and its actor-attribution mechanism in
   `registrar_pg.go` is available to consume. If not, halt this task and record the blocking
   dependency explicitly rather than reimplementing an independent attribution mechanism.
2. Identify the relationship-edge mutation call site(s) (`Relate`/revoke) at this task's actual start
   commit.
3. Implement an ownership check on edge create/revoke.
4. Wire attribution via DATA-06 T2's consumed mechanism (a call into the shared `registrar_pg.go`
   mechanism, not a reimplementation).
5. Implement an audit-row write for every edge create/revoke, using the existing `kernel/audit`
   convention (or a small, confirmed extension to it).
6. Implement version-bump logic on mutation, with a concurrency-safe mechanism against concurrent
   mutation attempts on the same edge.
7. Check whether W05-E04-S002 has landed. If yes, implement and test cache-invalidation triggering
   against it. If no, explicitly record the cache-invalidation sub-criterion as deferred-linked in
   `../story.md`/`../closure.md` — not silently dropped, not silently assumed complete.
8. Write the mutation-governance test proving ownership-check enforcement, correct attribution, an
   audit row written, and a version bump on mutation.

### Expected files or components affected

The relationship-edge mutation call site(s) (`Relate`/revoke, exact file TBD at implementation time);
a call into DATA-06 T2's consumed attribution mechanism in `registrar_pg.go` (no modification to that
file itself); possibly a small audit-table extension if the existing `kernel/audit` convention does
not already cover relationship-edge mutations.

### Expected output

Every relationship-edge create/revoke mutation is ownership-checked, attributed, audited, and
versioned. The cache-invalidation sub-criterion is either implemented and tested (if W05-E04-S002 has
landed) or explicitly recorded as deferred-linked (if not).

### Required artifacts

ART-W03-E04-S001-003 (the mutation-governance implementation).

### Required evidence

EV-W03-E04-S001-003 (mutation-governance test output; cache-invalidation test output if applicable).

### Related acceptance criteria

AC-W03-E04-S001-03.

### Completion criteria

The mutation-governance test proves ownership-check enforcement, correct attribution, an audit row
written, and a version bump on mutation. The cache-invalidation sub-criterion's disposition
(implemented-and-tested, or explicitly deferred-linked) is recorded, not left ambiguous.

### Verification method

Direct test execution against a testkit DB, logged output retained as evidence.

### Risks

RISK-W03-003 (the cache-invalidation sub-criterion depends on W05-E04-S002, which may not land on
this story's timeline) — see epic-level `risks.md`. Additionally: if DATA-06 T2 has not landed, this
task is blocked, which is itself a scheduling risk to this story's closure timeline, though not a
technical risk requiring workaround (the correct response is to wait or escalate, not to
reimplement).

### Rollback or recovery considerations

The ownership-check/audit/versioning logic is additive to the existing mutation call sites —
independently revertible if a legitimate mutation is found to be incorrectly blocked post-rollout.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not yet implemented.*

### Files changed

*Not yet implemented.*

### Interfaces introduced or changed

*Not yet implemented — anticipated: none new, unless the cache-invalidation portion (once
W05-E04-S002 exists) introduces a consumed interface.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not yet implemented — anticipated: possible small audit-table extension, to be confirmed.*

### Security changes

*Not yet implemented — this task's entire output is a mutation-governance security control.*

### Observability changes

*Not yet implemented — anticipated: the audit-row write itself.*

### Tests added or modified

*Not yet implemented.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*None anticipated, unless the cache-invalidation sub-criterion is deferred, in which case it is
tracked as a follow-up item (not technical debt in the negative sense) once W05-E04-S002 lands.*

### Known limitations

*Not yet implemented. If W05-E04-S002 has not landed, recorded here as the explicit, deferred-linked
limitation for the cache-invalidation sub-criterion.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W03-E04-S001-03 | Run the mutation-governance test: create/revoke a relationship edge, assert ownership check, attribution (via DATA-06 T2's mechanism), audit-row write, and version bump | Local dev or CI, testkit DB, DATA-06 T2 (W02-E04-S001) landed | Ownership-checked, attributed, audited, and versioned mutation confirmed; cache-invalidation sub-criterion tested if W05-E04-S002 has landed, otherwise recorded as deferred-linked | mutation-governance test report (+ cache-invalidation test report if applicable) | unassigned |

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
