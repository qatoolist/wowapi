---
id: IMPL-W06-E02-S002
type: implementation-record
parent_story: W06-E02-S002
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Implementation record — W06-E02-S002

REL-03a's six compatibility mechanisms are implemented, wired, locally exercised against their real
boundaries, independently rerun, and accepted.

## What was actually implemented

- Pinned `apidiff` exported-API gate with additive and breaking fixtures.
- Exact Go 1.26.0/1.26.5 compile matrix with explicit exclusions.
- Recursive generated config-schema compatibility gate and adversarial fixtures.
- v1.0.0 migration seed/upgrade/reverse/reconstruct drill against PostgreSQL.
- Exact-OCI-layout amd64/arm64 digest smoke inside the no-publish candidate job.
- Shared REL-01 clean verifier for GitHub/SBOM attestations, real cosign bundles, provenance, hashes,
  platforms, and CLI version.

## Components and files changed

Compatibility packages/CLI, migrations, `.github/workflows/compatibility-gates.yml`,
`.github/workflows/release.yml`, `scripts/smoke_candidate_arch.sh`,
`scripts/smoke_candidate_oci.sh`, and `scripts/validation`.

## Interfaces introduced or changed

Adds `compatcheck config`, `scripts/check_go_api_compat.sh`,
`scripts/smoke_candidate_arch.sh`, and
`scripts/smoke_candidate_oci.sh <candidate-oci.tar> <tag> <sha256:digest>`.
Production `verify-release` requires cosign; deterministic fixtures retain
`WOWAPI_OFFLINE_VERIFY=1`.

## Configuration, schema, security, and observability changes

The compatibility workflow produces baseline/current schemas and enumerates supported toolchains.
`MigrateTo` provides the controlled integration-test target. Candidate smoke and published signature
verification are immutable-digest, identity-bound, and fail closed. Gate output identifies failures.

## Tests added or modified

Adversarial Go/config fixtures, real PostgreSQL migration drill, exact-toolchain compile runs,
OCI orchestration/unit coverage, real multi-architecture archive smoke, and 12 release-contract golden
tests cover the six acceptance criteria.

## Revision, pull request, and implementation date

No commit or pull request created; shared working tree based on `733ef3e`; completed 2026-07-14.

## Technical debt, limitations, and follow-up

No compatibility shim or duplicate REL-03 release verifier was introduced. Hosted GitHub/Sigstore
execution remains authoritative for a real tagged release; local tests prove the fail-closed command
boundary and a real digest-selected multi-platform candidate.

## Relationship to the approved plan

Matches REL-03a T1/T2/T4/T6/T8/T9, including T9's deliberate fold-in to REL-01 T8/T9.
