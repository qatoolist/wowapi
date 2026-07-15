---
id: W07-E02-S002-T001
type: task
title: Fail-not-skip E2E prerequisites
status: done
parent_story: W07-E02-S002
owner: W07-E02-S002 executor
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on: []
acceptance_criteria:
  - AC-W07-E02-S002-01
artifacts:
  - ART-W07-E02-S002-001
evidence:
  - EV-W07-E02-S002-001
---

# W07-E02-S002-T001 — Fail-not-skip E2E prerequisites

## Task Definition

### Task objective

Make E2E prerequisite failures fail, not skip, in the authoritative E2E job.

### Parent story

W07-E02-S002

### Owner

unassigned

### Status

todo

### Dependencies

None.

### Detailed work

1. Classify each of the 22 inventoried skip sites as legitimately optional or masking required
   coverage.
2. Convert masking-required-coverage cases from skip to fail.
3. Kill a required E2E dependency and confirm the job now fails.

### Expected files or components affected

The authoritative E2E job's own workflow configuration.

### Expected output

Unmet prerequisite exits non-zero, not '0 tests ran, green.'

### Required artifacts

ART-W07-E02-S002-001 (fail-not-skip E2E job + classification record).

### Required evidence

EV-W07-E02-S002-001 (kill-a-required-dependency test output).

### Related acceptance criteria

AC-W07-E02-S002-01.

### Completion criteria

Killing a required E2E dependency produces a non-zero exit.

### Verification method

Direct execution: kill a required dependency, confirm failure.

### Risks

Medium — requires classifying which of the 22 inventoried skip sites are legitimately optional vs. mask required coverage, per PLAN T5's own risk note.

### Rollback or recovery considerations

If a legitimately-optional skip is incorrectly converted to fail, causing false CI failures, revert that specific site's classification and record why.

## Implementation Record

### What was actually implemented

Required DB/S3/E2E branches now fail before their local-only skip when the authoritative requirement
flag is set. The execution-time inventory expanded to 39; the flaky TOTP skip was removed and all 38
remaining sites classified in ART-W07-E02-S002-001.

### Components and files changed

E2E/DB/S3 prerequisite test helpers, `.github/workflows/ci.yml`, `Makefile`,
`miscellaneous/check_required_test_prerequisites.sh`, and the classification artifact.

### Interfaces and configuration

The authoritative workflow continues to set `WOWAPI_REQUIRE_DB=1` and `WOWAPI_REQUIRE_S3=1`; these
flags now close every classified required branch with an actionable fatal diagnosis.

### Schema, security, and observability

No schema change. Security/integration coverage can no longer disappear behind a green skip. Negative
fixtures print the exact missing dependency and requirement flag.

### Tests

`make check-required-test-prerequisites`; deterministic TOTP wrong-code regression.

### Revision, date, debt, and plan relationship

Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus scoped shared-worktree provenance; implemented
2026-07-14; no debt. DEV-W07-E02-S002-001 records the larger execution-time inventory.

## Verification Record

| Acceptance criterion | Actual result | Result | Evidence | Reviewer |
|---|---|---|---|---|
| AC-W07-E02-S002-01 | Missing DB and S3 each caused non-zero inner test exits with actionable diagnoses; wrapper passed only after observing both. | PASS | EV-W07-E02-S002-001 | W05ReviewGateFinal: PASS |

Environment: Darwin arm64, Go 1.26.5. Retest passed on 2026-07-14. Final task conclusion: verified and
artifact/evidence registered; independent story review remains the story-level gate.

## Deviations Record

See DEV-W07-E02-S002-001 for the execution-time 39-site inventory versus the historical 22-site plan.
