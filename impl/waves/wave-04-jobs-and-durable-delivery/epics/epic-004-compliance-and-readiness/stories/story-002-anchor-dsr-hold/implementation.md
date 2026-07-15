---
id: IMPL-W04-E04-S002
type: implementation-record
parent_story: W04-E04-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W04-E04-S002

## What was actually implemented

- **W6-T2 external anchor verification:** `kernel/audit/external_anchor.go` adds `ExternalStore`,
  `FileStore`, `ExternalAnchor`, `AnchorNow`, `Verify`, and `ErrAnchorTampered`. `AnchorNow` reads
  the current chain head via `audit.Writer.Anchor`, persists it through `ExternalStore`, and records
  an audit entry. `Verify` fetches the latest external anchor and uses `audit.Writer.CheckAnchor` to
  detect tail truncation / head-hash rewinding.
- **W6-T3 DSR export artifact:** `kernel/retention/artifact.go` adds `ArtifactWriter`,
  `FileArtifactWriter`, `ArtifactManifest`, `ClassResult`, `ErasureResult`, AES-256-GCM encryption,
  SHA-256 checksum, and audit rows for artifact creation and download. `retention.TestKey()`
  supplies a deterministic test key.
- **W6-T4 central legal-hold wrapper:** `kernel/retention/engine.go` now requires `*Holds`,
  `ArtifactWriter`, and `*audit.Writer`. `SweepDisposition` checks `record_class` holds before each
  `Dispose`; `RunErasure` checks `dsr_subject` holds before erasing. Non-compliant callbacks are
  blocked with `ErrHeld`.
- **W6-T5 explicit per-class status:** `RunExport` returns an `*ArtifactManifest` with a
  `PerClassResults` entry for every registered class (`exported`, `not_applicable`, or `empty`).
  `RunErasure` returns an `*ErasureResult` with per-class statuses and a total.

## Components changed

- `kernel/audit` (external anchor)
- `kernel/retention` (artifact writer, engine wrapper, status reporting)
- `kernel` (composition root wiring)

## Files changed

- `kernel/audit/external_anchor.go` (new)
- `kernel/audit/external_anchor_test.go` (new)
- `kernel/retention/artifact.go` (new)
- `kernel/retention/anchor_dsr_test.go` (new)
- `kernel/retention/engine.go` (modified)
- `kernel/retention/engine_test.go` (modified)
- `kernel/retention/coverage_test.go` (modified)
- `kernel/kernel.go` (modified)

## Interfaces introduced or changed

- `audit.ExternalStore`
- `retention.ArtifactWriter`
- `retention.Engine` now constructed as
  `NewEngine(reg, dsr, holds, artifacts, audit)`.
- `Engine.RunExport` now returns `(*ArtifactManifest, error)`.
- `Engine.RunErasure` now returns `(*ErasureResult, error)`.

## Configuration changes

- `WOWAPI_DSR_ARTIFACT_KEY` (hex, 32 bytes) sources the production AES key.
- `WOWAPI_ARTIFACT_DIR` optionally overrides the artifact directory (defaults to
  `<os.TempDir()>/wowapi-artifacts`).

## Schema or migration changes

None. Artifacts are stored as encrypted files; no new table or column is introduced.

## Security changes

- External anchor closes the local-head-hash-compromise blind spot.
- DSR exports are encrypted at rest (AES-256-GCM) and checksummed.
- Legal-hold enforcement is centralized; a non-compliant callback cannot bypass it.

## Observability changes

- `audit.external_anchor` row written on anchor.
- `dsr.artifact.created` and `dsr.artifact.download` audit rows written by the artifact writer.
- `kernel.New` warns when `WOWAPI_DSR_ARTIFACT_KEY` is unset or invalid.

## Tests added or modified

- `TestIntegrationExternalAnchorTamperDetection` (AC-01)
- `TestIntegrationDSRArtifactWriteAndChecksum` (AC-02)
- `TestIntegrationDSRExportArtifactWriteFailure` (AC-02)
- `TestIntegrationCentralLegalHoldBlocksDisposeErase` (AC-03)
- `TestIntegrationExplicitPerClassExportStatus` (AC-04)
- `TestIntegrationExplicitPerClassErasureStatus` (AC-04)
- Plus audit/download/round-trip coverage tests.

## Commits

- Working tree at `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

## Pull requests

*To be opened by the conductor/reviewer workflow.*

## Implementation dates

- 2026-07-13

## Technical debt introduced

None identified. The env-var key sourcing should be replaced with a KMS-backed `ArtifactWriter` for
production hardening, but that is a follow-up, not debt introduced by this change.

## Known limitations

- The file-backed artifact writer is local-only. A replicated deployment needs a shared storage
  backend or an object-storage `ArtifactWriter` implementation.
- No scheduled external-anchor job is wired in `kernel.New`; the interface is ready for one.

## Follow-up items

- Production KMS-backed `ArtifactWriter`.
- Scheduled `ExternalAnchor.AnchorNow` job (leader-safe).

## Relationship to the approved plan

Implementation matches `plan.md` except for the four bounded design decisions recorded in
`deviations.md` (D-W04-E04-S002-001 through -004).
