---
id: W04-E02-S001-T001
type: task
title: Shared-primitive reuse — claim-row lease-column migration
status: done
parent_story: W04-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on:
  - W04-E01-S001
acceptance_criteria:
  - AC-W04-E02-S001-01
artifacts:
  - ART-W04-E02-S001-001
evidence:
  - EV-W04-E02-S001-001
---

# W04-E02-S001-T001 — Shared-primitive reuse — claim-row lease-column migration

## Task Definition

### Task objective

Migrate `kernel/notify` and `kernel/webhook` claim rows onto W04-E01's shared lease/fencing
primitive's columns (`lease_token`, `lease_generation`, `lease_expires_at`), not a bespoke copy, so
that both packages' three-stage protocols (T002, T003) have a common fencing mechanism identical to
jobs (W04-E01-S002).

### Parent story

W04-E02-S001 — Notify and webhook three-stage remote-I/O protocol.

### Owner

unassigned

### Status

todo

### Dependencies

W04-E01-S001 (the shared lease/fencing primitive must exist and expose a stable claim/finalize API
before this task can integrate against it).

### Detailed work

1. Confirm W04-E01-S001's shared primitive's finalized column set and claim/finalize API surface.
2. Inspect the current schema of notify/webhook's delivery-tracking tables to determine what, if
   any, prior claim/status tracking exists that this migration supersedes.
3. Design and write the migration adding the shared primitive's lease columns to notify/webhook's
   delivery-tracking tables, additive (not destructive) to any existing delivery-tracking data.
4. Write a migration test confirming the new columns exist, default correctly, and do not disrupt
   any existing row.
5. Document that notify/webhook now use the same lease-column schema as jobs (W04-E01-S002),
   explicitly citing the shared primitive rather than describing a new, parallel mechanism.

### Expected files or components affected

A new migration file (exact path TBD); notify/webhook's delivery-tracking table definitions.

### Expected output

A schema migration adding the shared primitive's lease columns to notify/webhook's delivery-tracking
tables, proven additive and non-disruptive by a migration test.

### Required artifacts

ART-W04-E02-S001-001 (claim-row lease-column migration).

### Required evidence

EV-W04-E02-S001-001 (migration test report).

### Related acceptance criteria

AC-W04-E02-S001-01.

### Completion criteria

The migration adds the shared primitive's lease columns to both notify's and webhook's
delivery-tracking tables, proven by a passing migration test; a code-level inspection confirms the
same column set and semantics as W04-E01's shared primitive, not a parallel/bespoke implementation.

### Verification method

Direct execution of the migration test; code-level inspection comparing notify/webhook's new
columns against W04-E01's shared primitive's own schema definition.

### Risks

None beyond W04-E01's own risk (per DATA-03 T1's own risk column: "None beyond DATA-02's own
risk") — this task inherits, not adds to, the shared primitive's risk profile.

### Rollback or recovery considerations

Revert the migration if it disrupts in-flight notify/webhook deliveries; because the migration is
additive, a revert should not lose any pre-existing delivery-tracking data.

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

*Not yet implemented — the lease-column migration is planned, not yet executed.*

### Security changes

*Not applicable.*

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
| AC-W04-E02-S001-01 | Run migration test; inspect column set against W04-E01's shared primitive | Local dev or CI, PostgreSQL instance | New columns added, additive, match shared primitive's schema exactly | migration test report | unassigned |

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
