---
id: W05-E05-S002-T002
type: task
title: wowsociety identity-suite verification
status: todo
parent_story: W05-E05-S002
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W05-E05-S002-02
artifacts:
  - ART-W05-E05-S002-001
evidence:
  - EV-W05-E05-S002-002
---

# W05-E05-S002-T002 — wowsociety identity-suite verification

## Task Definition

### Task objective

Coordinate with wowsociety (PROD-02) to run its build and full identity/authz test suite against the
`kernel/mfa` shim (or `foundation/mfa` directly, if wowsociety has already migrated), recording both
repositories' commit SHAs.

### Parent story

W05-E05-S002 — Kernel package-count and wowsociety identity-suite verification.

### Owner

unassigned

### Status

todo

### Dependencies

None within this story (parallel-safe with T001); depends on W05-E05-S001 at story scope.

### Detailed work

1. Confirm W05-E05-S001's `kernel/mfa` shim has landed.
2. Coordinate with wowsociety (PROD-02) to pin a wowsociety commit against this wowapi commit (via
   the existing `replace`/`FRAMEWORK_VERSION` mechanism, consistent with PLAN's own AR-01
   wowsociety-verification pattern).
3. Run wowsociety's full build.
4. Run wowsociety's full identity/authz test suite — not a narrowed mfa-scoped subset, per REVIEW
   §P's own explicit instruction.
5. Record both repositories' commit SHAs and the full results in the verification-results document.

### Expected files or components affected

None in wowapi (verification-only); wowsociety's own repository is read/tested, not modified.

### Expected output

wowsociety's build and full identity/authz suite confirmed green against the shim, with both
repositories' commit SHAs recorded.

### Required artifacts

ART-W05-E05-S002-001 (shared with T001).

### Required evidence

EV-W05-E05-S002-002.

### Related acceptance criteria

AC-W05-E05-S002-02.

### Completion criteria

wowsociety's build and full identity/authz suite are confirmed green, with both commit SHAs
recorded.

### Verification method

Direct execution of wowsociety's build and test suite against the pinned wowapi commit.

### Risks

Medium-high — RISK-W05-004's own framing: a broken TOTP/OTP path is an authentication-availability
regression, and this task is the point at which that risk is either confirmed absent or surfaced.

### Rollback or recovery considerations

If wowsociety's suite reveals a failure, do not proceed to mark this story `accepted` — record the
failure as a finding requiring S001's own follow-up (a shim defect) or a wowsociety-side issue
(recorded for PROD-02's own coordination), whichever the root cause indicates.

## Implementation Record

*Not yet implemented.*

### What was actually implemented

*Not yet implemented.*

### Components changed

*Not applicable.*

### Files changed

*Not yet implemented.*

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

*Not applicable — this task runs wowsociety's existing suite.*

### Commits

*Not yet implemented.*

### Pull requests

*Not yet implemented.*

### Implementation dates

*Not yet implemented.*

### Technical debt introduced

*Not applicable.*

### Known limitations

*Not yet implemented.*

### Follow-up items

*Not yet implemented.*

### Relationship to the approved plan

*Not yet implemented.*

## Verification Record

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W05-E05-S002-02 | Run wowsociety's build and full identity/authz test suite | wowsociety repository, Go toolchain, CI, pinned wowapi commit | Build and full suite green | cross-repo test report | unassigned |

### Actual result

*Not yet executed.*

### Pass or fail

*Not yet executed.*

### Evidence identifier

*Not yet executed.*

### Execution date

*Not yet executed.*

### Commit or revision

*Not yet executed.*

### Environment

*Not yet executed.*

### Reviewer

*Not yet executed.*

### Findings

*Not yet executed.*

### Retest status

*Not yet executed.*

### Final conclusion

*Not yet executed.*

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
