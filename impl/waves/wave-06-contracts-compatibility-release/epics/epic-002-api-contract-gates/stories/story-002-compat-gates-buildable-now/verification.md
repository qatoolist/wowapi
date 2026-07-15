---
id: VER-W06-E02-S002
type: verification-record
parent_story: W06-E02-S002
status: verified
created_at: 2026-07-12
updated_at: 2026-07-14
---

# Verification record — W06-E02-S002

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E02-S002-01 | Run the API diff gate against a seeded breaking-API fixture | CI | The breaking fixture fails the gate | CI gate test report | unassigned |
| AC-W06-E02-S002-02 | Run the compile matrix and inspect CI configuration for explicit version exclusions | CI | Matrix runs across supported versions; exclusions are explicit, not silent | CI run report | unassigned |
| AC-W06-E02-S002-03 | Run the config-compat gate against seeded breaking and additive fixtures | CI | Breaking fixture fails; additive optional-field fixture passes | CI gate test report | unassigned |
| AC-W06-E02-S002-04 | Run the migration upgrade-drill test | CI, real Postgres | Seed at oldest supported version, migrate forward, reverse on disposable data, all succeed | integration-test report | unassigned |
| AC-W06-E02-S002-05 | Run the architecture-smoke job against the REL-01 candidate image for each published architecture | CI, candidate image from build-candidate stage | Every architecture boots and passes minimal smoke | CI job report | unassigned |
| AC-W06-E02-S002-06 | Cross-check REL-01 T8/T9's evidence for SBOM/provenance/signature verification | Documentation + evidence review | REL-01 T8/T9's golden-failure tests are correctly cross-referenced, not re-implemented | review report | unassigned |

## Post-execution record

All six acceptance criteria have executor and independent-verifier evidence.

### Actual result

- AC-01: API additive/breaking fixtures PASS.
- AC-02: Go 1.26.0 and 1.26.5 compile-only runs each PASS (68 packages; 8 no-test packages).
- AC-03: config additive/breaking/required-direction fixtures PASS.
- AC-04: real-PostgreSQL oldest-supported migration drill PASS.
- AC-05: exact OCI digest boots PASS on amd64 and arm64 before publish.
- AC-06: shared 12-test release verifier PASS; production cosign/SBOM/provenance wiring is actionlint-valid.

### Pass or fail

PASS.

### Evidence identifier

EV-W06-E02-S002-001 through EV-W06-E02-S002-007.

### Execution date and revision

2026-07-14; working tree based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Environment

Darwin arm64; Go 1.26.0/1.26.5; real local PostgreSQL; OrbStack Docker 29.4.0; ORAS 1.2.3;
amd64 emulation and native arm64; supplied DB/S3-required environment.

### Reviewer

W06-E02-S002-Rerun — PASS, no production findings.

### Findings

The remainder closed three acceptance-critical gaps: pre-publish smoke now runs against the exact OCI
candidate, production verifies real cosign bundles instead of synthetic JSON, and provenance has
explicit golden failures.

### Retest status and final conclusion

Focused compatibility, both compile toolchains, release-contract suite, actionlint, real PostgreSQL,
and an independently built multi-platform OCI digest smoke all PASS. W06-E02-S002 is verified.
