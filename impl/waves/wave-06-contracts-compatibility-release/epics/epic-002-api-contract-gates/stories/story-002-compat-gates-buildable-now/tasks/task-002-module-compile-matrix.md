---
id: W06-E02-S002-T002
type: task
title: Module compile matrix
status: done
parent_story: W06-E02-S002
owner: W06E02Impl
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W06-E02-S002-02
artifacts:
  - ART-W06-E02-S002-002
evidence:
  - EV-W06-E02-S002-002
---

# W06-E02-S002-T002 — Module compile matrix

## Task Definition

### Task objective

Build a module compile matrix across supported Go/dependency versions with explicit exclusions.

### Parent story

W06-E02-S002

### Owner

W06E02Impl

### Status

done

### Dependencies

None.

### Detailed work

1. Determine the set of supported Go/dependency versions.
2. Build a CI matrix compiling the module against each.
3. Document any excluded version explicitly in CI configuration.

### Expected files or components affected

New CI workflow configuration for the compile matrix.

### Expected output

A CI matrix job compiling across supported versions with explicit exclusions.

### Required artifacts

ART-W06-E02-S002-002 (module compile matrix CI configuration).

### Required evidence

EV-W06-E02-S002-002 (compile-matrix CI run output).

### Related acceptance criteria

AC-W06-E02-S002-02.

### Completion criteria

The matrix runs across supported versions; exclusions are explicit in CI config.

### Verification method

Direct execution of the matrix job; inspection of CI config for exclusions.

### Risks

None beyond CI wall-clock cost.

### Rollback or recovery considerations

If a version proves untenable to support, exclude it explicitly with a documented reason.

## Implementation Record

### What was actually implemented

`.github/workflows/compatibility-gates.yml` runs a fail-independent compile matrix at the
supported Go floor (`1.26.0`) and current supported patch (`1.26.5`). Unsupported older/newer
toolchains and a synthetic minimum-dependency set are excluded with explicit reasons in the workflow.

### Components and files changed

`.github/workflows/compatibility-gates.yml`; no runtime interface or schema changed.

### Configuration, security, and observability changes

CI-only toolchain/dependency-set configuration. No security or observability change.

### Tests added or modified

The matrix compiles every package with `go test -run '^$' ./...`.

### Implementation date and revision

2026-07-14; shared working tree based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`.

### Technical debt, limitations, and follow-up

No new debt. Hosted CI remains the authoritative release run, while both exact toolchains were also
executed locally for acceptance. Future support-window changes must edit the explicit matrix rationale.

### Relationship to the approved plan

Matches REL-03a T2: supported versions compile and all exclusions are visible.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E02-S002-02 | Execute both exact toolchains and inspect explicit exclusions | Go toolchain download/cache | Every package compiles under 1.26.0 and 1.26.5 | command output | W06-E02-S002-Rerun-2 |

### Actual result

`GOTOOLCHAIN=go1.26.0 go test -run '^$' ./...` and the corresponding `go1.26.5` run each
passed: 68 packages compiled, 8 packages had no tests. `actionlint` accepted the workflow.

### Pass or fail

PASS.

### Evidence identifier

EV-W06-E02-S002-002.

### Execution date, revision, and environment

2026-07-14; working tree based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`;
Darwin arm64 with downloaded Go 1.26.0 and 1.26.5 toolchains and the supplied DB/S3-required environment.

### Reviewer and findings

Executor: W06-E02-S002-Rerun-2. Independent verifier: W06-E02-S002-Rerun — PASS.
No compile failure or silent matrix exclusion was found.

### Retest status and final conclusion

PASS on fresh sequential runs. AC-02 is satisfied.

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
