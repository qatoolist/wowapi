---
id: W06-E02-S002-T001
type: task
title: Go public API diff
status: done
parent_story: W06-E02-S002
owner: W06E02Impl
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W06-E02-S002-01
artifacts:
  - ART-W06-E02-S002-001
evidence:
  - EV-W06-E02-S002-001
---

# W06-E02-S002-T001 — Go public API diff

## Task Definition

### Task objective

Wire golang.org/x/exp/apidiff/gorelease as a CI job classifying added/removed/changed exported symbols per DX-05's v1/N-1 policy.

### Parent story

W06-E02-S002

### Owner

W06E02Impl

### Status

done

### Dependencies

None.

### Detailed work

1. Wire apidiff/gorelease as a CI job.
2. Classify added/removed/changed exported symbols per DX-05's already-ratified v1/N-1 policy.
3. Write a seeded breaking-API fixture and confirm the gate fails it.

### Expected files or components affected

New CI workflow configuration for the API diff job.

### Expected output

A CI job correctly classifying API changes, failing a seeded breaking-API fixture.

### Required artifacts

ART-W06-E02-S002-001 (Go API diff CI job).

### Required evidence

EV-W06-E02-S002-001 (seeded breaking-API-fixture test report).

### Related acceptance criteria

AC-W06-E02-S002-01.

### Completion criteria

The seeded breaking-API fixture fails the gate.

### Verification method

Direct execution of the gate against the fixture.

### Risks

None beyond standard tooling-integration risk.

### Rollback or recovery considerations

If the gate produces false positives, revise the classification logic; do not silently disable the gate.

## Implementation Record

Implemented and focused-test verified.

### What was actually implemented

Pinned `golang.org/x/exp/cmd/apidiff` compares oldest-supported and current module exports and fails on incompatible reports.

### Components changed

Compatibility script, fixture modules, focused tests, and reusable CI job.

### Files changed

`scripts/check_go_api_compat.sh`; `internal/compat/go_api_gate_test.go`; `internal/compat/testdata/go-api/`; `.github/workflows/compatibility-gates.yml`.

### Interfaces introduced or changed

Adds the two-module-directory `check_go_api_compat.sh` gate interface.

### Configuration changes

Pinned `APIDIFF_VERSION`; workflow baseline ref defaults to `v1.0.0`.

### Schema or migration changes

*Not applicable.*

### Security changes

Tool version is pinned; script validates both module directories before execution.

### Observability changes

None.

### Tests added or modified

Identical and additive modules pass; removed method and changed exported type fail with classified symbol output.

### Commits

No commit; working tree based on `733ef3e`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

*None anticipated.*

### Known limitations

No known limitation; hosted CI remains authoritative for future refs and does not replace this evidence.

### Follow-up items

None.

### Relationship to the approved plan

Matches the approved mature-tooling and adversarial-fixture plan.

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W06-E02-S002-01 | Run the gate against the seeded fixture | CI | Breaking fixture fails the gate | CI gate test report | unassigned |

### Actual result

All four API fixture modules produced the expected pass/fail classification.

### Pass or fail

PASS.

### Evidence identifier

EV-W06-E02-S002-001.

### Execution date

2026-07-13.

### Commit or revision

Working tree based on `733ef3e`.

### Environment

Darwin arm64; Go 1.26.5.

### Reviewer

W06-E02-S002-Rerun — PASS.

### Findings

No open functional finding.

### Retest status

Focused retest PASS.

### Final conclusion

Implemented, independently verified, and accepted.

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
