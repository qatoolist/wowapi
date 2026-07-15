---
id: W06-E03-S001-T007
type: task
title: publish job scaffolding tested against a stub environment
status: done
parent_story: W06-E03-S001
owner: W06E03Impl
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W06-E03-S001-T006
acceptance_criteria:
  - AC-W06-E03-S001-07
artifacts:
  - ART-W06-E03-S001-007
evidence:
  - EV-W06-E03-S001-007
---

# W06-E03-S001-T007 — publish job scaffolding tested against a stub environment

## Task Definition

### Task objective

Add a publish job (needs: build-candidate, protected release environment) copying only manifested artifacts, never rebuilding — tested against a stub environment since the real protected environment does not yet exist.

### Parent story

W06-E03-S001

### Owner

unassigned

### Status

todo

### Dependencies

W06-E03-S001-T006 (publish consumes build-candidate's manifested artifacts).

### Detailed work

1. Add the publish job, scoped to needs: build-candidate.
2. Implement artifact-diffing logic: the job must refuse any artifact/digest absent from
   release-manifest.json.
3. Test against a stub environment (exact mechanism TBD — see `plan.md` Unresolved questions), since
   the real protected release environment does not yet exist (W06-E03-S002's scope).
4. Write an unmanifested-artifact test: inject an extra artifact, prove rejection.

### Expected files or components affected

.github/workflows/release.yml (publish job scaffolding).

### Expected output

A publish job that rejects any unmanifested artifact, proven against a stub environment.

### Required artifacts

ART-W06-E03-S001-007 (publish job scaffolding).

### Required evidence

EV-W06-E03-S001-007 (unmanifested-artifact test output).

### Related acceptance criteria

AC-W06-E03-S001-07

### Completion criteria

The unmanifested-artifact test proves rejection, tested against a stub environment.

### Verification method

Direct execution of the unmanifested-artifact test against the stub environment.

### Risks

High — blocked on human environment setup for full end-to-end proof; code can be tested against a stub environment but not proven end-to-end without the real protected environment (W06-E03-S002).

### Rollback or recovery considerations

This job runs unprotected-in-scratch until W06-E03-S002's DEC-Q10 resolution creates the real protected environment — this is the expected interim state, not a defect, per REVIEW §G's own framing.

## Implementation Record

Implemented protected, draft-first `gh`/ORAS publication after GitHub attestation, gate-manifest hash, and every manifested byte are re-verified. It copies only existing candidate bytes and never rebuilds; unmanifested input is rejected. Evidence: EV-W06-E03-S001-007.
## Verification Record

Pass for scratch/stub scope — unmanifested file rejected; only manifest-declared, hash-verified bytes copied; failed candidate never moved immutable/latest pointers. Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; EV-W06-E03-S001-007.
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
