---
id: W06-E03-S001-T008
type: task
title: verify_release.sh with golden failure tests
status: done
parent_story: W06-E03-S001
owner: W06E03Impl
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W06-E03-S001-T007
acceptance_criteria:
  - AC-W06-E03-S001-08
artifacts:
  - ART-W06-E03-S001-008
  - ART-W06-E03-S001-009
evidence:
  - EV-W06-E03-S001-008
  - EV-W06-E03-S001-009
---

# W06-E03-S001-T008 — verify_release.sh with golden failure tests

## Task Definition

### Task objective

Write scripts/validation/verify_release.sh <version> <source-sha> with golden failure tests, one per verified property, plus the verify-published job and SLSA-guarantee documentation.

### Parent story

W06-E03-S001

### Owner

unassigned

### Status

todo

### Dependencies

W06-E03-S001-T007 (the script verifies what T006/T007 produce).

### Detailed work

1. Write scripts/validation/verify_release.sh, running from a genuinely clean environment, exiting
   non-zero on any single mismatched field.
2. Write golden failure tests, one per verified property (wrong SHA, stripped signature, missing SBOM
   attestation, wrong platforms, tampered manifest hash) — per §13.1's explicit requirement that "a
   prose checklist is not sufficient."
3. Add a verify-published job invoking the script on a clean runner; failure marks the release failed,
   blocks latest promotion.
4. Run an end-to-end dry run against a disposable throwaway repo, never the real release pipeline.
5. Write SLSA 1.2 guarantee documentation stating exactly which build-track requirements are met, no
   over-claim.

### Expected files or components affected

scripts/validation/verify_release.sh; a verify-published job; SLSA-guarantee documentation.

### Expected output

A script with golden failure tests proving every verified property, plus a clean-runner verify-published job and honest SLSA documentation.

### Required artifacts

ART-W06-E03-S001-008 (verify_release.sh), ART-W06-E03-S001-009 (verify-published job + SLSA documentation).

### Required evidence

EV-W06-E03-S001-008 (golden-failure test output, one per property), EV-W06-E03-S001-009 (end-to-end dry-run output).

### Related acceptance criteria

AC-W06-E03-S001-08

### Completion criteria

Every golden failure test passes; the end-to-end dry run catches a corrupted publish and confirms latest is not moved; SLSA documentation makes no over-claim.

### Verification method

Direct execution of the golden failure tests and the end-to-end dry run against a disposable repo.

### Risks

Medium for the script (per PLAN T8); High for the end-to-end dry run (needs a disposable repo/registry for safe rehearsal, per PLAN T9); Low for documentation, but a false SLSA claim is itself a supply-chain trust defect.

### Rollback or recovery considerations

If a golden failure test cannot be made to pass, treat as a defect in verify_release.sh itself, not a reason to weaken the test.

## Implementation Record

Implemented `verify_release.sh`, the underlying clean verifier, per-property golden failures, clean-runner GitHub attestation checks, immutable image digest/platform smoke, and post-verification release/alias promotion. Evidence: EV-W06-E03-S001-008 and EV-W06-E03-S001-009.
## Verification Record

Pass — clean-verifier golden failures cover SHA, signature, SBOM, provenance, platforms, version, archives, images, and published hashes; temporary promotion dry run passed. Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; EV-W06-E03-S001-008/009.
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
