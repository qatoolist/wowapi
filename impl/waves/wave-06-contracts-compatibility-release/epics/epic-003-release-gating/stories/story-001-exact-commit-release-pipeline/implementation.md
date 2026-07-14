---
id: IMPL-W06-E03-S001
type: implementation-record
parent_story: W06-E03-S001
status: implemented
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W06-E03-S001

This record aggregates the implemented exact-commit release path and its focused verification.

## What was actually implemented

Implemented a schema-validated Wave-6 gate catalog; reusable exact-SHA gate execution and attestation; CI/release callers; immutable candidate creation; blocking artifact/image scans; tamper-resistant manifest validation; draft exact-byte `gh`/ORAS publication; clean-runner post-publication verification; and verified alias promotion.

## Components changed

`.github/workflows/{required-gates,ci,release}.yml`, `ci/release-gates*`, and `scripts/validation/release_contract.py`.

## Files changed

See the story artifact index; focused fixtures and raw evidence are under `scripts/validation/tests/` and `evidence/tests/`.

## Interfaces introduced or changed

`release_contract.py` exposes fail-closed validate/run/describe/create/verify/publish/promote/verify-release/verify-tag commands. `required-gates.yml` accepts a full source SHA and emits attested gate results.

## Configuration changes

Release tags invoke the same reusable gates; build has no publish permissions; publication is protected, draft-first, exact-byte-only; mutable aliases and public release state move only after clean verification.

## Schema or migration changes

Added strict JSON Schemas for the required-gate catalog and immutable release manifest inputs. No database migration.

## Security changes

Gate, manifest, candidate, archive, image, SBOM, provenance, signature, security-report, platform, and version tampering fail closed.

## Observability changes

Gate outputs and release/security reports are retained as SHA-bound evidence artifacts and release subjects.

## Tests added or modified

Added 10 focused release-contract tests, including temporary Git repositories/registries and per-property golden failures; actionlint is clean.

## Commits

Workspace revision under test: `733ef3e930cbb3f89f5bbc53d8f562c60e426513` (dirty shared workspace; no commit created).

## Pull requests

None created.

## Implementation dates

2026-07-13.

## Technical debt introduced

The publisher deviation is explicitly recorded in `deviations.md`; it avoids relying on an unavailable GoReleaser Pro command.

## Known limitations

Real protected-environment execution remains DEC-Q10/S002's human-only blocker. The authored path was proven with scratch/throwaway fixtures, not falsely claimed live.

## Follow-up items

After an administrator activates DEC-Q10, run S002's live post-activation verification without changing the implementation.

## Relationship to the approved plan

Matched the plan except for ADR-005's implementation-time incompatibility. The release/security lead authorized the exact-byte `gh`/ORAS publisher; see `deviations.md`.
