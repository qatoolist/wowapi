---
id: W06-E02-S002-T007
type: task
title: Independent review
status: done
parent_story: W06-E02-S002
owner: W06-E02-S002-Rerun
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W06-E02-S002-T001
  - W06-E02-S002-T002
  - W06-E02-S002-T003
  - W06-E02-S002-T004
  - W06-E02-S002-T005
  - W06-E02-S002-T006
acceptance_criteria:
  - AC-W06-E02-S002-01
  - AC-W06-E02-S002-02
  - AC-W06-E02-S002-03
  - AC-W06-E02-S002-04
  - AC-W06-E02-S002-05
  - AC-W06-E02-S002-06
artifacts: []
evidence: []
---

# W06-E02-S002-T007 — Independent review

## Task Definition

### Task objective

Perform an independent review of this story's implementation per mandate §14, confirming each of the six REL-03a gates genuinely enforces its stated acceptance bar, and that T9's cross-reference to REL-01 T8/T9 is accurate, not a disguised duplicate implementation.

### Parent story

W06-E02-S002

### Owner

W06-E02-S002-Rerun

### Status

done

### Dependencies

T001 through T006 (review requires all prior tasks implemented first).

### Detailed work

1. Confirm T001's API diff gate genuinely fails the seeded breaking-API fixture.
2. Confirm T002's compile matrix genuinely runs across the claimed version set with genuinely explicit
   exclusions.
3. Confirm T003's config-compat gate genuinely fails the seeded breaking-config fixture and passes the
   additive one.
4. Confirm T004's migration upgrade-drill genuinely seeds at the oldest supported version and genuinely
   reverses.
5. Confirm T005's architecture-smoke job genuinely runs against the candidate image, not an
   already-published one.
6. Confirm T006's cross-reference to REL-01 T8/T9 is accurate and does not silently duplicate work.
7. Record findings; resolve or explicitly accept before this story moves to `accepted`.

### Expected files or components affected

None (review-only task; no code change).

### Expected output

An independent-review record confirming or rejecting this story's readiness for `accepted` status.

### Required artifacts

None.

### Required evidence

None beyond this task's own review record.

### Related acceptance criteria

AC-W06-E02-S002-01, AC-W06-E02-S002-02, AC-W06-E02-S002-03, AC-W06-E02-S002-04, AC-W06-E02-S002-05, AC-W06-E02-S002-06 (confirms all six, does not itself prove any new one).

### Completion criteria

The review record confirms all six acceptance criteria are proven with valid evidence, or lists findings that must be resolved before this story can close.

### Verification method

Manual review against mandate §14's checklist, cross-referenced with T001-T006's evidence.

### Risks

None beyond the review itself missing a genuine gap.

### Rollback or recovery considerations

Not applicable — a review task has no code to roll back; a failed review blocks story acceptance until its findings are resolved.

## Implementation Record

*Not applicable — this is a review task, not an implementation task.*

### What was actually implemented

*Not applicable.*

### Components changed

*Not applicable.*

### Files changed

*Not applicable.*

### Interfaces introduced or changed

*Not applicable.*

### Configuration changes

*Not applicable.*

### Schema or migration changes

*Not applicable.*

### Security changes

*Not applicable.*

### Observability changes

*Not applicable.*

### Tests added or modified

*Not applicable.*

### Commits

*Not applicable.*

### Pull requests

*Not applicable.*

### Implementation dates

*Not applicable.*

### Technical debt introduced

*Not applicable.*

### Known limitations

*Not applicable.*

### Follow-up items

None.

### Relationship to the approved plan

*Not applicable.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E02-S002-01 | Fresh compatibility fixture run | Go | Breaking API fixture rejected | command output | W06-E02-S002-Rerun |
| AC-W06-E02-S002-02 | Both exact toolchain compile runs | Go 1.26.0/1.26.5 | Every supported package compiles | command output | W06-E02-S002-Rerun |
| AC-W06-E02-S002-03 | Fresh config fixture run | Go | Breaking rejected; additive accepted | command output | W06-E02-S002-Rerun |
| AC-W06-E02-S002-04 | Fresh migration drill | Real PostgreSQL | Seed/forward/reverse succeeds | command output | W06-E02-S002-Rerun |
| AC-W06-E02-S002-05 | Independent OCI build/copy/digest boot | Docker, ORAS, amd64/arm64 | Exact candidate boots before publish | command output | W06-E02-S002-Rerun |
| AC-W06-E02-S002-06 | Release suite and workflow review | Python, actionlint | Shared verifier is real, fail-closed, and non-duplicated | review output | W06-E02-S002-Rerun |

### Actual result

All six criteria independently passed. The reviewer reran actionlint, both smoke tests, all 12 release
contract tests, both exact compile toolchains, full compatibility fixtures, and the real PostgreSQL
migration drill. An independent multi-platform OCI candidate was built, copied with pinned ORAS 1.2.3
to a fresh registry, and booted by digest on amd64 and arm64; both returned `wowapi verify-s002`.

### Pass or fail

PASS — no production findings.

### Evidence identifier

EV-W06-E02-S002-007.

### Execution date, revision, and environment

2026-07-14; shared working tree based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`;
Darwin arm64, Go 1.26.0/1.26.5, Docker/QEMU, pinned ORAS 1.2.3, real PostgreSQL, supplied DB/S3-required environment.

### Reviewer

W06-E02-S002-Rerun.

### Findings

No open issues. Review confirmed pre-publish ordering, exact-layout digest selection, pinned
registry/ORAS/cosign, online real-bundle verification with exact tag identity and OIDC issuer,
isolated offline fixtures, and explicit SBOM/provenance golden failures.

### Retest status and final conclusion

PASS. W06-E02-S002 is accepted.

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
