---
id: W02-E05-S001-T002
type: task
title: Seed-sync command and manifest implementation
status: todo
parent_story: W02-E05-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W02-E05-S001-T001
acceptance_criteria:
  - AC-W02-E05-S001-02
artifacts:
  - ART-W02-E05-S001-002
  - ART-W02-E05-S001-003
evidence:
  - EV-W02-E05-S001-002
---

# W02-E05-S001-T002 — Seed-sync command and manifest implementation

## Task Definition

### Task objective

Implement the seed-sync command/path and the catalog manifest schema/loader per T001's resolved
design, proving the two properties CS-21's acceptance bar names first: idempotent and RLS-respecting.

### Parent story

W02-E05-S001 — Production seed-sync path — design investigation and implementation.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E05-S001-T001 (this task implements against T001's resolved design; it cannot begin before T001's
decision record exists).

### Detailed work

1. Implement the catalog manifest schema/loader per T001's resolved format and versioning scheme.
2. Implement the seed-sync command/path (CLI shape per T001's resolved design, sketched by CS-21 as
   `wowapi seed sync --env prod`).
3. Implement the idempotency mechanism T001 resolved (e.g. content-hash comparison against an
   applied-version tracking table).
4. Implement the RLS/role posture T001 resolved, ensuring the sync's actual runtime behavior matches
   T001's documented safety rationale.
5. Write the idempotency test: run seed-sync twice against the same manifest version, confirm no
   duplicate or conflicting state results.
6. Write the RLS/role-posture test: confirm the sync's actual role and its interaction with RLS
   policies matches T001's documented rationale — this test must be adversarial enough to actually
   exercise the RLS boundary, not merely assert the role name is as expected.
7. If T001 resolved that concurrent invocation is in-scope, write a test proving safe behavior under
   concurrent seed-sync attempts; if it was explicitly assumed away as single-invoker, confirm no
   code path silently assumes multi-invoker safety it does not have.

### Expected files or components affected

A new seed-sync command/path (exact file path TBD by T001); a new catalog manifest schema/loader
(exact file path TBD by T001); possibly a new schema migration for an applied-manifest-version
tracking table, if T001's idempotency mechanism requires one (exact determination deferred to T001's
output, not invented here).

### Expected output

A working seed-sync command/path that is idempotent and RLS-respecting per T001's resolved design,
proven by the idempotency test and the RLS/role-posture test.

### Required artifacts

ART-W02-E05-S001-002 (catalog manifest schema definition), ART-W02-E05-S001-003 (seed-sync
command/path).

### Required evidence

EV-W02-E05-S001-002 (idempotency + RLS-posture integration-test report).

### Related acceptance criteria

AC-W02-E05-S001-02.

### Completion criteria

The idempotency test and the RLS/role-posture test both pass, proving the seed-sync path's behavior
genuinely matches T001's documented design and safety rationale — not merely that the code compiles
or that a happy-path run completes.

### Verification method

Direct execution of both tests against a live PostgreSQL instance with RLS policies active, logged
output retained as evidence.

### Risks

RISK-W02-E05-001 (the RLS-respecting bootstrap tension) is this task's central risk — an
incorrectly-implemented role posture, even if T001's design was sound on paper, would be a
security-relevant defect, not merely a quality gap. This task's RLS/role-posture test exists
specifically to catch a design-to-implementation gap, not only a design gap.

### Rollback or recovery considerations

If the RLS/role-posture test fails against T001's resolved design (i.e. the design itself, not just
its implementation, proves unsafe), escalate back to T001 for redesign rather than silently loosening
the test or the RLS posture to make it pass — record this as a deviation, per `deviations.md`.

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

*Not yet implemented — whether this task requires a schema migration depends on T001's resolved
idempotency mechanism, not yet known.*

### Security changes

*Not yet implemented — the RLS/role posture is itself the security control; recorded here once
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
| AC-W02-E05-S001-02 | Run the idempotency test and the RLS/role-posture test | Local dev or CI, PostgreSQL instance with RLS policies active | Repeated run produces no duplicate/conflicting state; role posture matches T001's documented rationale | integration-test report | unassigned |

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
