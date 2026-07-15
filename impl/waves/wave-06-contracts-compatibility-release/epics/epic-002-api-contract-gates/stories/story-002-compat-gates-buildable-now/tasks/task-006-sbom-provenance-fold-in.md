---
id: W06-E02-S002-T006
type: task
title: SBOM/provenance/signature verification fold-in
status: done
parent_story: W06-E02-S002
owner: W06E02Impl
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W06-E02-S002-06
artifacts:
  - ART-W06-E02-S002-006
evidence:
  - EV-W06-E02-S002-006
---

# W06-E02-S002-T006 — SBOM/provenance/signature verification fold-in

## Task Definition

### Task objective

Cross-reference REL-01 T8/T9's existing SBOM/provenance/signature verification as satisfying REL-03's own naming of the same property, per PLAN's own fold-in framing.

### Parent story

W06-E02-S002

### Owner

W06E02Impl

### Status

done

### Dependencies

Shares evidence with, but does not require completion of, W06-E03-S001's REL-01 T8/T9 work before this task's own cross-reference can be drafted (the cross-reference itself can be written once REL-01 T8/T9's evidence exists).

### Detailed work

1. Confirm REL-01 T8/T9's golden-failure tests (W06-E03-S001) cover the same acceptance bar
   REL-03 T9 names.
2. Write the cross-reference documentation, per PLAN's own citation ("Same acceptance as REL-01 T8/T9 |
   Same golden-failure tests | REL-01/verify_release/ (shared)").
3. Confirm no separate implementation is silently introduced that duplicates REL-01's own work.

### Expected files or components affected

A documentation cross-reference; no new source-code file.

### Expected output

A confirmed, accurate cross-reference from REL-03 T9 to REL-01 T8/T9's shared evidence.

### Required artifacts

ART-W06-E02-S002-006 (REL-03 T9 cross-reference).

### Required evidence

EV-W06-E02-S002-006 (shared REL-01 T8/T9 evidence reference).

### Related acceptance criteria

AC-W06-E02-S002-06.

### Completion criteria

The cross-reference is accurate and no duplicate implementation exists.

### Verification method

Documentation review, cross-checked against REL-01 T8/T9's actual evidence.

### Risks

The primary risk is accidentally re-implementing REL-01 T8/T9's work under a different name — mitigated by this task's own explicit no-duplication check.

### Rollback or recovery considerations

If REL-01 T8/T9's evidence does not actually cover REL-03 T9's acceptance bar, escalate rather than silently writing an inaccurate cross-reference.

## Implementation Record

### What was actually implemented

REL-03 reuses REL-01's clean published-release verifier. Offline fixtures retain their deterministic
JSON receipt, while production mode now invokes `cosign verify-blob` against the real Sigstore bundle
with exact tag-workflow identity and GitHub OIDC issuer. The clean workflow installs pinned cosign,
verifies GitHub build/SBOM attestations, checks both provenance records, hashes, platforms and CLI
version, and promotes aliases only after that verifier passes.

### Components and files changed

`.github/workflows/release.yml`, `scripts/validation/release_contract.py`, and
`scripts/validation/tests/test_release_contracts.py`.

### Interfaces and configuration changed

Production `verify-release` now requires `cosign`; `WOWAPI_OFFLINE_VERIFY=1` remains restricted to
deterministic local fixtures. No application runtime or schema interface changed.

### Security and observability changes

Signature verification is cryptographic and fail-closed in production rather than parsing a synthetic
receipt. Verification output names rejected bundles and missing/tampered SBOM or provenance evidence.

### Tests added or modified

Added real-bundle-schema acceptance, exact cosign argument/identity assertions, fail-closed cosign
rejection, missing provenance, and tampered provenance golden failures.

### Implementation date and revision

2026-07-14; shared working tree based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Technical debt, limitations, and follow-up

No duplicate REL-03 verifier was created. Keyless verification itself is hosted-only because it
requires the real Sigstore certificate and GitHub OIDC identity; the production invocation is covered
with a fail-closed tool boundary and exact-argument tests.

### Relationship to the approved plan

Matches REL-03a T9's explicit fold-in to REL-01 T8/T9 and uses the same shared evidence.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E02-S002-06 | Execute shared verifier golden suite and review production workflow | Python, fake cosign boundary; hosted GitHub/Sigstore for release | Accurate shared reference; signature/SBOM/provenance failures reject | test and workflow review | W06-E02-S002-Rerun-2 |

### Actual result

All 12 release-contract tests pass. The suite proves real Sigstore bundle schema is handed to
`cosign verify-blob`, cosign rejection fails closed, and stripped signature, missing SBOM, missing
provenance, tampered provenance, wrong platforms, wrong SHA, and tampered manifest are rejected.
`actionlint` accepted both release and compatibility workflows.

### Pass or fail

PASS.

### Evidence identifier

EV-W06-E02-S002-006 (shared REL-01 T8/T9 verification evidence).

### Execution date, revision, and environment

2026-07-14; working tree based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`;
Darwin arm64 with Python 3, supplied DB/S3-required environment, and workflow review.

### Reviewer and findings

Executor: W06-E02-S002-Rerun-2. Independent verifier: W06-E02-S002-Rerun — PASS.
The prior synthetic-versus-real cosign gap and provenance golden-test gap are closed.

### Retest status and final conclusion

PASS. AC-06 is satisfied through the single shared REL-01 verifier; no duplicate implementation exists.

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
