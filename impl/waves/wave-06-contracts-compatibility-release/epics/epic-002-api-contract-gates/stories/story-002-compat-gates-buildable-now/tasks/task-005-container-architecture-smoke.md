---
id: W06-E02-S002-T005
type: task
title: Container architecture smoke
status: done
parent_story: W06-E02-S002
owner: W06E02Impl
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W06-E02-S002-05
artifacts:
  - ART-W06-E02-S002-005
evidence:
  - EV-W06-E02-S002-005
---

# W06-E02-S002-T005 — Container architecture smoke

## Task Definition

### Task objective

Build a container architecture smoke CI job running against the REL-01 candidate image for every published architecture.

### Parent story

W06-E02-S002

### Owner

W06E02Impl

### Status

done

### Dependencies

Cross-story: coordinates with W06-E03-S001's REL-01 T6/T7 build-candidate split, which produces the candidate image this task smoke-tests against (see `plan.md` Unresolved questions for the exact coordination point).

### Detailed work

1. Coordinate with W06-E03-S001 on the candidate image's availability and shape.
2. Build a smoke-test job running against the candidate image for every published architecture.
3. Confirm each architecture boots and passes minimal smoke before `publish`.

### Expected files or components affected

New CI workflow configuration for the architecture-smoke job.

### Expected output

A smoke-test job proving every published architecture boots from the candidate image.

### Required artifacts

ART-W06-E02-S002-005 (container architecture smoke CI job).

### Required evidence

EV-W06-E02-S002-005 (architecture-smoke test report).

### Related acceptance criteria

AC-W06-E02-S002-05.

### Completion criteria

Every published architecture boots and passes minimal smoke against the candidate image.

### Verification method

Direct execution of the smoke job per architecture.

### Risks

arm64 via QEMU is slow per PLAN's own risk note — native runners should be considered if CI budget becomes a concern.

### Rollback or recovery considerations

If a specific architecture proves unreliable in CI, diagnose root cause per systematic-debugging discipline; do not silently drop that architecture from the published set without escalating.

## Implementation Record

### What was actually implemented

The release `build-candidate` job now exports its one multi-platform OCI archive, copies that exact
layout into an ephemeral loopback registry, and boots both `linux/amd64` and `linux/arm64` by the
BuildKit-produced immutable digest before any publish-capable job begins. The registry image, ORAS,
QEMU, BuildKit, and workflow actions are pinned. Cleanup is fail-safe.

### Components and files changed

`.github/workflows/release.yml`, `scripts/smoke_candidate_oci.sh`,
`scripts/smoke_candidate_arch.sh`, and `internal/compat/architecture_smoke_test.go`.

### Interfaces and configuration changed

New internal validation command:
`scripts/smoke_candidate_oci.sh <candidate-oci.tar> <tag> <sha256:digest>`.
No application runtime or schema interface changed.

### Security and observability changes

The smoke is inside the no-publish `build-candidate` trust boundary and addresses the artifact by
digest. It emits per-architecture pass/fail output; failures block the candidate.

### Tests added or modified

Added orchestration coverage proving exact-layout copy, pinned registry startup, digest use, both
platform invocations, and cleanup. Existing digest/platform behavior coverage remains.

### Implementation date and revision

2026-07-14; shared working tree based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Technical debt, limitations, and follow-up

No new debt. The hosted release remains authoritative; local verification used the same OCI archive,
ORAS version, registry digest, and architecture script.

### Relationship to the approved plan

Matches REL-03a T8 and closes the earlier ordering defect: smoke now occurs before `publish`.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E02-S002-05 | Copy exact OCI candidate to ephemeral registry and boot both digest-selected platforms | Docker/BuildKit, ORAS 1.2.3, QEMU | amd64 and arm64 return the version response before publish | command output | W06-E02-S002-Rerun-2 |

### Actual result

An OCI archive with root digest
`sha256:bc5a6bc43418d523947d0145491d14f2ffefb683421699907c693f9c668f480c`
was copied without rebuild and booted successfully as amd64 and arm64. Both returned
`wowapi verify-733ef3e`. Focused Go smoke tests and `actionlint` also passed.

### Pass or fail

PASS.

### Evidence identifier

EV-W06-E02-S002-005.

### Execution date, revision, and environment

2026-07-14; working tree based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`;
Darwin arm64, OrbStack Docker 29.4.0, ORAS 1.2.3, amd64 emulation and native arm64.

### Reviewer and findings

Executor: W06-E02-S002-Rerun-2. Independent verifier: W06-E02-S002-Rerun — PASS.
The previous post-publish-only smoke gap is closed.

### Retest status and final conclusion

PASS on a real multi-platform OCI archive. AC-05 is satisfied.

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
