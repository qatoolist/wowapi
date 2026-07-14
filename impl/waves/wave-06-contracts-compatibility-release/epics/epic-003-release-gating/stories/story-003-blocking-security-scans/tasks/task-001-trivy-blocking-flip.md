---
id: W06-E03-S003-T001
type: task
title: Trivy blocking flip with reviewed allowlist scoping
status: done
parent_story: W06-E03-S003
owner: W06E03Impl
created_at: 2026-07-12
updated_at: 2026-07-12
depends_on: []
acceptance_criteria:
  - AC-W06-E03-S003-01
artifacts:
  - ART-W06-E03-S003-001
evidence:
  - EV-W06-E03-S003-001
---

# W06-E03-S003-T001 — Trivy blocking flip with reviewed allowlist scoping

## Task Definition

### Task objective

Flip Trivy to blocking (exit-code: "1"), scoping ignore-unfixed to a reviewed allowlist only, after a report-only baseline run.

### Parent story

W06-E03-S003

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Run Trivy report-only first to baseline current findings, per PLAN T1's own risk note.
2. Review the baseline findings.
3. Flip Trivy to exit-code: "1", scoping ignore-unfixed to a reviewed allowlist only (not blanket).
4. Write a seeded-vulnerability fixture proving fail-then-pass-after-removal-or-waiver.

### Expected files or components affected

The Trivy scanning workflow configuration (likely .github/workflows/security-scan.yml).

### Expected output

Trivy blocks on CRITICAL/HIGH findings with an available fix; ignore-unfixed scoped to a reviewed allowlist.

### Required artifacts

ART-W06-E03-S003-001 (Trivy blocking-flip configuration).

### Required evidence

EV-W06-E03-S003-001 (seeded-vulnerability fail-then-pass test report).

### Related acceptance criteria

AC-W06-E03-S003-01.

### Completion criteria

The seeded fixture fails then passes after removal/waiver; ignore-unfixed is scoped, not blanket.

### Verification method

Direct execution of the seeded-vulnerability fixture test.

### Risks

Medium — run once report-only to baseline before flipping, or it can immediately break main on latent findings, per PLAN T1's own risk note.

### Rollback or recovery considerations

If the flip breaks main on a latent finding not caught by the report-only baseline, add a properly-documented waiver (T2) rather than reverting to the unscoped soft-fail posture.

## Implementation Record

Performed a report-only baseline, added one active path-scoped `AVD-DS-0002` waiver, and flipped source plus release-artifact/image Trivy scans to blocking exit code 1. The real seeded lodash fixture proves fail-then-pass. Evidence: EV-W06-E03-S003-001.
## Verification Record

Pass — `scripts/validation/tests/test_trivy_seed.sh` rejected the seeded vulnerable lock and accepted it after removal. Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; EV-W06-E03-S003-001.
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
