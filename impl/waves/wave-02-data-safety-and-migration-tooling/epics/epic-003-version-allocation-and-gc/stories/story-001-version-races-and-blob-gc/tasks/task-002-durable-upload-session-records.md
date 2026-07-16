---
id: W02-E03-S001-T002
type: task
title: Durable upload-session records
status: done
parent_story: W02-E03-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-16
depends_on:
  - W02-E03-S001-T001
acceptance_criteria:
  - AC-W02-E03-S001-02
artifacts:
  - ART-W02-E03-S001-002
  - ART-W02-E03-S001-005
evidence:
  - EV-W02-E03-S001-002
---

# W02-E03-S001-T002 — Durable upload-session records

## Task Definition

### Task objective

Add durable upload-session records to `kernel/document`: expiry, checksum/size, storage key, status,
and cleanup ownership, persisted before the presigned upload URL is returned to the caller.

### Parent story

W02-E03-S001 — Version-allocation races and upload-blob GC.

### Owner

unassigned

### Status

todo

### Dependencies

W02-E03-S001-T001 (per PLAN DATA-05 T2's own Depends-on column: "T1" — the session record's version
field is meaningless without T1's race-free allocation mechanism already in place).

### Detailed work

1. Design the upload-session table schema (expiry, checksum/size, storage key, status, cleanup
   ownership), following this programme's `<module>_<entity>` table-naming convention and applying
   RLS, per PLAN T2's own risk note: "New table needs RLS + `<module>_<entity>` naming."
2. Implement the migration introducing the upload-session table.
3. Implement session persistence in `kernel/document.InitiateUpload`: write the session row (status
   `pending`, with expiry set) before constructing and returning the presigned upload URL.
4. Write the crash-simulation test: initiate an upload, simulate a crash (i.e., do not proceed to
   confirmation), and assert the session row exists with `status='pending'` and a set expiry.
5. Document the upload-session schema and lifecycle (states, transitions, expiry).

### Expected files or components affected

`kernel/document`'s `InitiateUpload` code path; a new schema migration for the upload-session table
(exact file path TBD).

### Expected output

A durable upload-session record, persisted before URL issuance, proven by the crash-simulation test.

### Required artifacts

ART-W02-E03-S001-002 (the upload-session schema and table), ART-W02-E03-S001-005 (documentation,
shared with T001/T004).

### Required evidence

EV-W02-E03-S001-002 (crash-simulation test output).

### Related acceptance criteria

AC-W02-E03-S001-02.

### Completion criteria

A session row is persisted (status `pending`, expiry set) before the presigned upload URL is
returned; the crash-simulation test passes.

### Verification method

Direct execution of the crash-simulation test against a live PostgreSQL instance, asserting the
session row's existence and state after a simulated crash.

### Risks

New table requires RLS and correct `<module>_<entity>` naming — an RLS gap on this table would allow
cross-tenant visibility into upload-session metadata (storage keys, checksums), a real security
exposure if missed, per PLAN T2's own risk note.

### Rollback or recovery considerations

If session persistence introduces unacceptable latency before URL issuance, escalate for redesign
(e.g. async persistence with a compensating check) rather than silently reverting to the prior
no-session behavior, which would reintroduce the orphan-blob gap this story exists to close.

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

*Not yet implemented — expected: a new upload-session table.*

### Security changes

*Not yet implemented — expected: RLS on the new upload-session table.*

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
| AC-W02-E03-S001-02 | Run the crash-simulation upload-session test | Local dev or CI, PostgreSQL instance | Session row exists with `status='pending'` and a set expiry after simulated crash | integration-test report | unassigned |

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
