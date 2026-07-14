---
id: PLAN-W06-E03-S001
type: plan
parent_story: W06-E03-S001
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
---

# Plan — W06-E03-S001

Per mandate §8.5. This plan follows PLAN REL-01's own directive-selected 5-step design summary exactly.
Confirmed facts, planned changes, and assumptions are distinguished explicitly below.

## Proposed architecture

PLAN's own directive-selected design (5-step summary), reproduced as this story's architecture: (1) a
reusable `required-gates.yml` (`workflow_call`) runs a versioned `ci/release-gates.yaml` manifest
against an exact SHA; (2) both PR/main CI and release call it, release passing `${{ github.sha }}`
from the tag event, emitting an attested `gate-results.json`; (3) `build-candidate` (no publish
permissions) verifies gate results, builds once, emits artifacts + `release-manifest.json`, never
pushes; (4) `publish` (protected `release` environment, the only job with write permissions) copies
exactly the manifested bytes, never rebuilds; (5) `verify-published` re-verifies everything from a clean
runner post-publish.

## Implementation strategy

1. Design and implement `ci/release-gates.yaml`'s schema + JSON Schema validator (T1).
2. Populate Wave-0 manifest entries (T2), cross-checked against every job currently required for green
   `ci-container`.
3. Build `required-gates.yml` (T3), parameterized on SHA, emitting attested `gate-results.json`.
4. Update `ci.yml` to call `required-gates.yml` (T4), confirming byte-identical results for the same SHA
   through both paths.
5. Add the `verify` job to `release.yml` (T5), with a seeded-failure fixture.
6. Implement the `build-candidate`/`publish` split (T6), using GoReleaser `--skip=publish` per D-05
   (`ADR-W00-E02-S003-005`, consumed by reference), confirming the version-support caveat first.
7. Add the `publish` job scaffolding (T7), tested against a stub environment.
8. Write `verify_release.sh` (T8) with golden failure tests, one per verified property.
9. Add the `verify-published` job (T9's buildable half) with an end-to-end dry run against a disposable
   repo; write the SLSA-guarantee documentation.

## Expected package or module changes

New: `ci/release-gates.yaml`, a JSON Schema validator, `.github/workflows/required-gates.yml`,
`scripts/validation/verify_release.sh`. Extended: `.github/workflows/ci.yml`, `.github/workflows/
release.yml`, the GoReleaser configuration.

## Expected file changes where determinable

- `ci/release-gates.yaml` (new).
- A JSON Schema validator for the manifest (new, exact location TBD).
- `.github/workflows/required-gates.yml` (new).
- `.github/workflows/ci.yml` (extended to call `required-gates.yml`).
- `.github/workflows/release.yml` (extended with `verify`, `build-candidate`, `publish` scaffolding,
  `verify-published`).
- `.goreleaser.yaml` or equivalent (extended for `--skip=publish` support).
- `scripts/validation/verify_release.sh` (new).
- `docs/` — SLSA-guarantee documentation (new or extended).

## Contracts and interfaces

`ci/release-gates.yaml`'s own schema (ID, command/job ref, owner, `required_from_wave`, timeout,
evidence-artifact path) is the primary new contract. `gate-results.json` and `release-manifest.json`
are the primary new attested-output contracts.

## Data structures

The manifest-entry schema itself; the `gate-results.json` and `release-manifest.json` output schemas.

## APIs

None affected — this is CI/release-tooling-internal.

## Configuration changes

None to runtime configuration; CI/workflow configuration changes only.

## Persistence changes

None.

## Migration strategy

Not applicable.

## Concurrency implications

None beyond standard CI-job concurrency already managed by GitHub Actions.

## Error-handling strategy

Every gate/job must fail clearly and attestably — `gate-results.json` reports each manifest entry
individually, not an aggregate pass/fail only.

## Security controls

`build-candidate`'s no-write-permission scoping (T6) and `publish`'s manifested-artifact-only copying
(T7) are both required security controls, not optional hardening.

## Observability changes

`gate-results.json` and `release-manifest.json` are themselves the primary observability/audit
artifacts this story adds.

## Testing strategy

- T1: malformed-manifest-fixture unit test.
- T2: diff-review confirming entry count matches existing required-check count.
- T3: seeded-failure fixture through the workflow, confirming attested failure.
- T4: diff-based test confirming byte-identical results (excluding run ID/timestamp) for the same SHA.
- T5: seeded-failure fixture (a deliberately broken test on a tagged commit), confirming `verify` fails
  and `build-candidate` never runs.
- T6: tamper test (hand-edit one artifact byte), confirming mismatch detection.
- T7: unmanifested-artifact test (against a stub environment), confirming rejection.
- T8: golden failure tests, one per verified property.
- T9 (buildable half): end-to-end dry run against a disposable throwaway repo.

## Regression strategy

Once wired into CI, `required-gates.yml`'s manifest becomes the ongoing regression guard for every
future required check — a check removed or renamed without a corresponding manifest update fails CI.

## Compatibility strategy

T4's own constraint (no PR CI latency regression) is the primary compatibility concern; the manifest
schema itself is additive infrastructure, not a breaking change to any existing workflow's own
semantics.

## Rollout strategy

T1-T8 (and T9's buildable half) land as this story's own reviewable unit; T7's `publish` job remains
unprotected-in-scratch (per REVIEW §G's own interim-default framing) until W06-E03-S002's DEC-Q10
resolution creates the real protected environment.

## Rollback strategy

If any gate/job proves to have a false-positive failure mode once wired into real CI, it can be
temporarily demoted to advisory while the false positive is diagnosed — standard CI-gate rollback
pattern. T6's GoReleaser split-mode choice, if ADR-005's caveat surfaces a real incompatibility, would
require escalation per RISK-W06-E03-001's own contingency, not a silent workaround.

## Implementation sequence

T1 → T2 → T3 → T4 → T5 → T6 → T7 → T8 → T9 (buildable half), matching PLAN REL-01's own T1-T9
dependency chain exactly (each task's own "Depends-on" column in PLAN's task table).

## Task breakdown

- **W06-E03-S001-T001** — Manifest schema + validator (T1).
- **W06-E03-S001-T002** — Wave-0 manifest entries (T2).
- **W06-E03-S001-T003** — `required-gates.yml` (T3).
- **W06-E03-S001-T004** — `ci.yml` wiring (T4).
- **W06-E03-S001-T005** — `release.yml` `verify` job + seeded-failure fixture (T5).
- **W06-E03-S001-T006** — `build-candidate` split via GoReleaser `--skip=publish` (T6).
- **W06-E03-S001-T007** — `publish` job scaffolding tested against a stub environment (T7).
- **W06-E03-S001-T008** — `verify_release.sh` + golden failure tests (T8).
- **W06-E03-S001-T009** — Independent review.

## Expected artifacts

`ci/release-gates.yaml` + validator; Wave-0 manifest entries; `required-gates.yml`; `ci.yml` wiring;
`release.yml`'s `verify` job; `build-candidate` job; `publish` job scaffolding; `verify_release.sh`;
`verify-published` job + SLSA documentation.

## Expected evidence

Malformed-manifest-fixture test output; manifest-entry-count diff review; seeded-failure gate-results
attestation; diff-based same-SHA-both-paths test; seeded-failure tag test; tamper test; golden failure
tests (one per verified property); end-to-end dry-run output.

## Unresolved questions

- T7's exact stub-environment testing mechanism.
- Whether T9's `verify-published` job and SLSA documentation land as one task or are further split — this
  story folds them together as T008/documentation within T008's own scope, per the task brief's own
  framing.
- ADR-005's GoReleaser-version-support confirmation outcome (T006's own first step).

## Approval conditions

This plan is approved for implementation once: (a) ADR-005 is confirmed still valid against the
repository's pinned GoReleaser version (T006's own first step, or escalated if not), and (b) the owner
and reviewer are assigned.
