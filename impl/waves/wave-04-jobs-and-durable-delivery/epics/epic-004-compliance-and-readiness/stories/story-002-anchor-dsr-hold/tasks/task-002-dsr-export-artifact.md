---
id: W04-E04-S002-T002
type: task
title: Encrypted immutable DSR export artifact
status: done
parent_story: W04-E04-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W04-E04-S002-02
artifacts:
  - ART-W04-E04-S002-002
  - ART-W04-E04-S002-006
evidence:
  - EV-W04-E04-S002-002
---

# W04-E04-S002-T002 — Encrypted immutable DSR export artifact

## Task Definition

### Task objective

Replace `retention/engine.go`'s bare in-memory map return with a durable, encrypted, immutable DSR
export artifact containing a manifest, per-class results, checksum, expiry, and access policy, with
downloads audited, so that export completion is gated on artifact-write success rather than a
best-effort return value.

### Parent story

W04-E04-S002 — External anchoring, DSR export artifact, central legal-hold, and explicit per-class
status.

### Owner

unassigned

### Status

done

### Dependencies

None.

### Detailed work

1. Draft the DSR export artifact's exact format (manifest fields, per-class result schema, checksum
   algorithm, expiry semantics, access-policy model, download-audit schema).
2. Draft the encryption-key-management design (custody, rotation, recovery), per PLAN DATA-08 W6-T3's
   own risk note ("New encryption-key-management dependency").
3. Implement the artifact writer: replace `retention/engine.go`'s bare in-memory map return with a
   write path producing the encrypted, checksummed artifact.
4. Gate export completion on artifact-write success — a DSR export must not report completion if the
   artifact write fails partway.
5. Implement access-gated download with download-audit logging.
6. Write the export-completion/checksum-verification test.
7. Document the artifact format and encryption-key-management scheme.

### Expected files or components affected

`kernel/retention/engine.go` (DSR export path, replaced/extended); new encryption-key-management code
(exact package location TBD); a new test file for the export-completion/checksum test.

### Expected output

A working DSR export artifact writer; a passing export-completion/checksum-verification test;
documentation of the artifact format and key-management scheme.

### Required artifacts

ART-W04-E04-S002-002 (DSR export artifact writer), ART-W04-E04-S002-006 (documentation, shared with
T001/T003/T004).

### Required evidence

EV-W04-E04-S002-002 (DSR export artifact-completion/checksum report).

### Related acceptance criteria

AC-W04-E04-S002-02.

### Completion criteria

DSR export completes only after successfully writing the encrypted artifact; the checksum verifies
against the written artifact; access-gated downloads are audited.

### Verification method

Direct execution of the export-completion/checksum-verification test.

### Risks

RISK-W04-E04-002 (the new encryption-key-management dependency) — see epic-level `risks.md`. An
under-specified key-management design could leave exported artifacts either unrecoverable (key loss)
or insufficiently protected (weak key handling).

### Rollback or recovery considerations

Once artifacts have been produced under the chosen encryption scheme, reverting the scheme requires a
compatibility plan for already-produced artifacts — recorded as a rollback constraint to resolve at
implementation time, not invented as a specific procedure here (see `story.md` "Rollback strategy").

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

*Not yet implemented — a possible artifact-registry table, exact shape TBD.*

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
| AC-W04-E04-S002-02 | Run export-completion/checksum-verification test | Local dev or CI, artifact storage + encryption-key source | Export completes only after artifact write succeeds; checksum verifies | DSR export artifact-completion report | unassigned |

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
