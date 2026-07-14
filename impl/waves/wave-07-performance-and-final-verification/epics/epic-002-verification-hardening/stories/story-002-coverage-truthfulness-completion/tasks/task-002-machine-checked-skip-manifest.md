---
id: W07-E02-S002-T002
type: task
title: Machine-checked skip manifest
status: done
parent_story: W07-E02-S002
owner: W07-E02-S002 executor
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W07-E02-S002-T001
acceptance_criteria:
  - AC-W07-E02-S002-02
artifacts:
  - ART-W07-E02-S002-002
evidence:
  - EV-W07-E02-S002-002
---

# W07-E02-S002-T002 — Machine-checked skip manifest

## Task Definition

### Task objective

Build a machine-checked skip manifest extending check_test_skips.sh; new/unapproved skip fails CI, approved skip with rationale passes.

### Parent story

W07-E02-S002

### Owner

unassigned

### Status

todo

### Dependencies

W07-E02-S002-T001 ('only meaningful once known-bad skips are fixed,' per PLAN T6's own dependency note).

### Detailed work

1. Extend check_test_skips.sh with a machine-checked skip manifest.
2. Write a fixture adding an unguarded t.Skip() and confirm it fails CI.
3. Confirm an approved skip with documented rationale passes.

### Expected files or components affected

check_test_skips.sh (extended); a new skip-manifest file.

### Expected output

New/unapproved skip fails CI; approved skip with rationale passes.

### Required artifacts

ART-W07-E02-S002-002 (machine-checked skip manifest).

### Required evidence

EV-W07-E02-S002-002 (unguarded t.Skip() fixture fail-test output).

### Related acceptance criteria

AC-W07-E02-S002-02.

### Completion criteria

The unguarded-skip fixture fails; the approved-skip fixture passes.

### Verification method

Direct execution of both fixture tests.

### Risks

Medium, per PLAN T6's own risk classification.

### Rollback or recovery considerations

If a legitimate skip is incorrectly rejected by the manifest, add its approval entry with rationale rather than silently reverting the manifest's own enforcement.

## Implementation Record

### What was actually implemented

`internal/tools/testskipmanifest` parses every non-testdata `*_test.go` AST and exactly reconciles
`t.Skip`, `t.Skipf`, and `t.SkipNow` sites with manifest version 1. It rejects unapproved sites, stale
approvals, duplicate IDs/sites, missing owner/rationale, invalid classification, and required entries
without a named guard.

### Components and files changed

`miscellaneous/test-skip-manifest.json`, `miscellaneous/check_test_skips.sh`,
`miscellaneous/check_test_skip_fixtures.sh`, `internal/tools/testskipmanifest/`, `Makefile`, and the CI
unit job.

### Interfaces and configuration

`make check-test-skips` is the machine gate and includes isolated approved/unapproved fixtures.

### Tests

Unit tests cover seed rejection, approval acceptance, and incomplete metadata. The command fixture
proves an unapproved call exits 1 while a fully owned/rationalized approval exits 0.

### Revision, date, debt, and plan relationship

Revision `733ef3e930cbb3f89f5bbc53d8f562c60e426513` plus scoped shared-worktree provenance; implemented
2026-07-14; no debt; matches the plan.

## Verification Record

| Acceptance criterion | Actual result | Result | Evidence | Reviewer |
|---|---|---|---|---|
| AC-W07-E02-S002-02 | Repository 38-site manifest passed; unapproved fixture rejected with exact diagnosis; approved owner+rationale fixture passed. | PASS | EV-W07-E02-S002-002 | W05ReviewGateFinal: PASS |

Environment: Darwin arm64, Go 1.26.5. Retest passed on 2026-07-14. Final task conclusion: verified,
artifact/evidence registered.

## Deviations Record

No task-specific deviation beyond story-level DEV-W07-E02-S002-001.
