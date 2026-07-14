---
id: IMPL-W06-E03-S003
type: implementation-record
parent_story: W06-E03-S003
status: implemented
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Implementation record — W06-E03-S003

This record aggregates the implemented REL-02 fail-closed scanner and waiver controls.

## What was actually implemented

Implemented blocking Trivy vulnerability/secret/misconfiguration/license scans, expiring scoped waivers, hosted public exact-SHA scanner verification, private local SAST/posture fallback, required-gate cross-references, tag triggers, and blocking scans of release archives and OCI candidates.

## Components changed

Security workflows, schemas/registries, validation commands, seeded adversarial fixtures, and release-manifest security subjects.

## Files changed

`.github/workflows/{security-scan,codeql,scorecard,required-gates,release}.yml`, `.trivyignore.yaml`, `ci/security-waivers*`, `scripts/validation/security_contract.py`, and focused fixtures/tests.

## Interfaces introduced or changed

`security_contract.py` validates waivers/ignore sync, scanner results, visibility policy, local fallback, workflow posture, and one-to-one gate catalog coverage.

## Configuration changes

All scanner failures block. Public exact-SHA runs require hosted security workflow success; non-public repositories automatically execute explicit local substitutes without claiming CodeQL/Scorecard parity.

## Schema or migration changes

Added strict waiver JSON Schema. No database migration.

## Security changes

Every required scanner class is catalogued; artifact and image Trivy JSON reports are immutable required release-manifest subjects.

## Observability changes

Scanner results and raw focused reports are retained as evidence and release subjects.

## Tests added or modified

Added 8 security-contract tests plus executable seeded Trivy fail/remove test; all passed.

## Commits

Workspace revision under test: `733ef3e930cbb3f89f5bbc53d8f562c60e426513` (no commit created).

## Pull requests

None created.

## Implementation dates

2026-07-13.

## Technical debt introduced

None.

## Known limitations

Hosted CodeQL/Scorecard parity is inherently unavailable on private repositories; the fallback is documented as narrower and fails closed.

## Follow-up items

No implementation follow-up. Repository-admin protection activation remains separate S002 scope.

## Relationship to the approved plan

Matched `plan.md`; no S003 deviation was required.
