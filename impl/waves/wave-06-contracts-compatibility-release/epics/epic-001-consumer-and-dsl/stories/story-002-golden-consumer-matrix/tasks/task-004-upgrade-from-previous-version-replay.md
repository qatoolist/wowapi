---
id: W06-E01-S002-T004
type: task
title: Upgrade-from-previous-version replay
status: done
parent_story: W06-E01-S002
owner: W06E01Impl
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W06-E01-S002-T003
acceptance_criteria:
  - AC-W06-E01-S002-04
artifacts:
  - ART-W06-E01-S002-004
evidence:
  - EV-W06-E01-S002-004
  - EV-W06-E01-S002-010
---

# W06-E01-S002-T004 — Upgrade-from-previous-version replay

## Task Definition

### Task objective

Replay an upgrade-from-previous-version cycle: fixture generated at N-1 (per DX-05's ratified v1/N-1 policy), upgraded to N, contracts rerun and passing.

### Parent story

W06-E01-S002

### Owner
W06E01Impl

### Status

done

### Dependencies

W06-E01-S002-T003 (the fixture must boot and pass its baseline exercises before an upgrade replay is meaningful).

### Detailed work

1. Generate the fixture at N-1, per DX-05's already-ratified v1/N-1 compatibility-class policy
   (W01-E04-S002, executed).
2. Upgrade the fixture to N.
3. Rerun T003's contracts (authenticated CRUD, async delivery, restart/retry, RLS isolation) against the
   upgraded fixture and confirm they pass.
4. Confirm this is a genuine two-pass test (N-1 state genuinely exercised before the upgrade, not a
   single-pass assertion dressed up as a replay).

### Expected files or components affected

An upgrade-replay test harness within the fixture.

### Expected output

A two-pass integration test proving the fixture survives an N-1-to-N upgrade with contracts intact.

### Required artifacts

ART-W06-E01-S002-004 (upgrade-replay harness).

### Required evidence

EV-W06-E01-S002-004 (two-pass integration-test report).

### Related acceptance criteria

AC-W06-E01-S002-04.

### Completion criteria

The upgrade replay is a genuine two-pass test and contracts pass after the upgrade.

### Verification method

Direct execution of the two-pass upgrade-replay test.

### Risks

RISK-W06-E01-002 (epic-scoped): a DX-05 policy-application ambiguity could stall this task on a question outside this story's own authority to resolve.

### Rollback or recovery considerations

If DX-05's policy proves ambiguous for this specific fixture, escalate to the developer-experience lead rather than silently inventing an interpretation.

## Implementation Record

Completed 2026-07-14 in `internal/cli/golden_consumer_test.go`.

`TestGoldenConsumerUpgradeReplay` installs the tagged `v1.1.0` CLI; scaffolds the consumer; generates
both modules and their subsystem content; and proves build/boot while `go.mod` pins v1.1.0 with no
checkout replace. It then installs the locally packaged candidate CLI
`v1.2.0-w06e01s002.11`, upgrades the framework dependency and scaffold, regenerates with `--force`,
repeats the build/boot contract against the candidate, and finally runs the upgraded fixture's
real-infrastructure contracts.

### Tests added or modified

`TestGoldenConsumerUpgradeReplay` plus shared generation and contract helpers.

### Implementation dates

2026-07-14.

### Known limitations

The N side is a content-pinned local candidate rather than a falsely claimed published tag.

### Relationship to the approved plan

This is a genuine two-pass N-1-to-N-candidate replay: the N-1 fixture is generated and exercised before
the upgrade, and the upgraded fixture is exercised again.
## Verification Record

### Actual result

PASS. Tagged `v1.1.0` generation/build/boot passed before upgrade. Candidate dependency/scaffold
upgrade, forced regeneration, candidate build/boot, and upgraded real-infrastructure contracts passed.

### Evidence identifier

EV-W06-E01-S002-004 and EV-W06-E01-S002-010.

### Execution date

2026-07-13T20:33:12Z.

### Commit or revision

Worktree snapshot based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`, content-pinned in
`artifacts/index.md`; tagged baseline `v1.1.0`; candidate `v1.2.0-w06e01s002.11`.

### Environment

Go 1.26.5; local candidate module proxy; real Docker Compose services.

### Reviewer

W06-E01-S002-Verify.

### Retest status

Retested after offline-module-cache environment failures; final two-pass run passed.

### Final conclusion

T004 is done and AC-04 is verified.
## Deviations Record

No task-local deviation. The replay documentation now states the matrix actually executed:
tagged v1.1.0 to the locally packaged release candidate.
