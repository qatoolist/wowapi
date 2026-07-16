---
id: W02-E04-S001-T001
type: task
title: Typed aggregate repository/unit-of-work helper
status: done
parent_story: W02-E04-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on: []
acceptance_criteria:
  - AC-W02-E04-S001-01
artifacts:
  - ART-W02-E04-S001-001
evidence:
  - EV-W02-E04-S001-001
---

# W02-E04-S001-T001 — Typed aggregate repository/unit-of-work helper

## Task Definition

### Task objective

Build a typed aggregate repository/unit-of-work helper in `kernel/resource` that bundles a module's
business-row write with the resource-mirror upsert, an audit-row write, and an outbox-entry write
atomically in one transaction, so a module cannot write its business row without the framework also
writing the mirror.

### Parent story

W02-E04-S001 — Typed aggregate write contract with mandatory mirror, audit, and outbox.

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Re-read `kernel/resource`'s current package documentation and the existing registrar `Upsert`
   API at this task's actual start commit to confirm the current-state assessment (manual,
   comment-only contract; no framework enforcement).
2. Design the helper's interface: what it accepts from the calling module (the business-row write
   operation) and what it performs internally (mirror upsert, audit write, outbox write), all
   within one transaction.
3. Implement the helper, keeping the existing low-level `Upsert` API available and unchanged
   alongside it (per PLAN's own wowsociety-compatibility note).
4. Write a fault-injection test suite that independently injects a failure at each of the 4 stages
   (business write, mirror upsert, audit write, outbox write) and confirms full transaction
   rollback at every one of the 4 fault points.

### Expected files or components affected

A new helper within `kernel/resource` (exact file path TBD per `plan.md`'s "Unresolved questions").

### Expected output

A working, atomicity-proven aggregate repository/unit-of-work helper, plus its fault-injection test
suite as evidence.

### Required artifacts

ART-W02-E04-S001-001 (the helper).

### Required evidence

EV-W02-E04-S001-001 (fault-injection test report).

### Related acceptance criteria

AC-W02-E04-S001-01.

### Completion criteria

Fault injection at each of the 4 stages independently causes full rollback at every stage, proven by
a passing test suite against a named commit SHA.

### Verification method

Direct execution of the fault-injection test suite, logged output retained as evidence.

### Risks

RISK-W02-E04-001 (overlap with AR-03's future work, W05-E03) — see epic-level `risks.md`. This
task's design decisions should be documented explicitly (not left implicit in code) so a future
AR-03 implementer can evaluate compatibility deliberately.

### Rollback or recovery considerations

Revert the helper if the fault-injection suite surfaces a partial-rollback defect that cannot be
resolved within this task's bounded scope; escalate for redesign rather than shipping a helper that
claims atomicity it does not actually provide.

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

*Not applicable — the security-relevant fix is T002's scope, not this task's.*

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
| AC-W02-E04-S001-01 | Run the fault-injection test suite against each of the 4 stages independently | Local dev or CI, PostgreSQL instance | Full rollback at every one of the 4 fault points | fault-injection test report | unassigned |

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
