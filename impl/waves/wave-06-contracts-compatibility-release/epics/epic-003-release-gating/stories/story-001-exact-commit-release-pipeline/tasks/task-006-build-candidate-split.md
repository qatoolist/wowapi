---
id: W06-E03-S001-T006
type: task
title: build-candidate split via GoReleaser --skip=publish
status: done-with-deviation
parent_story: W06-E03-S001
owner: W06E03Impl
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on:
  - W06-E03-S001-T005
acceptance_criteria:
  - AC-W06-E03-S001-06
artifacts:
  - ART-W06-E03-S001-006
evidence:
  - EV-W06-E03-S001-006
---

# W06-E03-S001-T006 — build-candidate split via GoReleaser --skip=publish

## Task Definition

### Task objective

Split into build-candidate (no publish permissions) using GoReleaser --skip=publish per D-05
(`ADR-W00-E02-S003-005`, consumed by reference — this story mints no new ADR), emitting
archives/checksums/SBOMs/OCI layout plus an attested release-manifest.json.

### Parent story

W06-E03-S001

### Owner

unassigned

### Status

todo

### Dependencies

W06-E03-S001-T005 (build-candidate runs only after verify passes).

### Detailed work

1. Confirm ADR-005's GoReleaser split-mode support against the repository's pinned GoReleaser
   version, per ADR-005's own stated responsibility.
2. Implement build-candidate with permissions scoped to contents:read, id-token:write,
   attestations:write only — no write.
3. Use GoReleaser release --skip=publish to emit archives/checksums/SBOMs/OCI layout.
4. Emit an attested release-manifest.json.
5. Write a tamper test: hand-edit one artifact byte, prove mismatch detected.

### Expected files or components affected

.github/workflows/release.yml (build-candidate job); .goreleaser.yaml or equivalent.

### Expected output

A no-write-permission build-candidate job producing attested, tamper-evident release artifacts.

### Required artifacts

ART-W06-E03-S001-006 (build-candidate job).

### Required evidence

EV-W06-E03-S001-006 (tamper-test output).

### Related acceptance criteria

AC-W06-E03-S001-06

### Completion criteria

The tamper test proves mismatch detection; the job's token cannot push or release.

### Verification method

Direct execution of the tamper test against a scratch/throwaway repo.

### Risks

High — needs a design spike per PLAN T6's own risk note; RISK-W06-E03-001 (ADR-005's version-support caveat).

### Rollback or recovery considerations

If the pinned GoReleaser version does not support --skip=publish as ADR-005 assumed, escalate to the release/security-engineering lead rather than silently hand-rolling a substitute pipeline.

## Implementation Record

Implemented a no-publish-permission candidate job using pinned GoReleaser v2.17.0 `release --clean --skip=publish`, a single multi-platform OCI export, SBOM/provenance/signature creation, blocking artifact/image scans, and an immutable release manifest. The separate publisher mechanism follows accepted deviation DEV-W06-E03-S001-001. Evidence: EV-W06-E03-S001-006.
## Verification Record

Pass with accepted deviation — tampered gate/tag/manifest/archive/image/candidate inputs and missing security reports rejected. GoReleaser OSS caveat independently reproduced. Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; EV-W06-E03-S001-006.
## Deviations Record

DEV-W06-E03-S001-001 applies: GoReleaser OSS v2.17.0 has no separate `publish` command and Pro Split & Merge can push during its split phase. Main authorized the manifest-verified draft `gh`/ORAS exact-byte publisher. See the story-level `deviations.md` and caveat evidence.
