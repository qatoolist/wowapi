---
id: W06-E01-S002-T002
type: task
title: Cross-module, cross-subsystem generation
status: done
parent_story: W06-E01-S002
owner: W06E01Impl
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W06-E01-S002-T001
acceptance_criteria:
  - AC-W06-E01-S002-02
artifacts:
  - ART-W06-E01-S002-002
evidence:
  - EV-W06-E01-S002-002
  - EV-W06-E01-S002-007
  - EV-W06-E01-S002-010
---

# W06-E01-S002-T002 — Cross-module, cross-subsystem generation

## Task Definition

### Task objective

Generate a resource, rule, workflow, event handler, recurring job, document flow, notification, and webhook across two modules, each subsystem exercised at least once.

### Parent story

W06-E01-S002

### Owner
W06E01Impl

### Status

done

### Dependencies

W06-E01-S002-T001 (the scaffold job's installed CLI is the tool this task invokes).

### Detailed work

1. Choose domain content for two modules (exact names/fields TBD at implementation time).
2. Generate a resource, rule, workflow, event handler, recurring job, document flow, notification, and
   webhook, distributed across the two modules such that each subsystem is exercised at least once.
3. Confirm generation succeeds for each subsystem with no manual post-generation edits required.

### Expected files or components affected

Generated module content within the golden-consumer fixture directory.

### Expected output

Two generated modules collectively exercising all 8 named subsystems at least once each.

### Required artifacts

ART-W06-E01-S002-002 (generated fixture content).

### Required evidence

EV-W06-E01-S002-002 (subsystem-coverage report).

### Related acceptance criteria

AC-W06-E01-S002-02.

### Completion criteria

All 8 subsystems are generated and present in the fixture's two modules.

### Verification method

Run the generation step and inspect output for each subsystem's presence.

### Risks

PLAN's own risk note: 'High — broad surface, many kernel subsystems must be generator-reachable' (RISK-W06-E01-001).

### Rollback or recovery considerations

If a subsystem is not generator-reachable within this task's bounded scope, record the gap in `deviations.md` rather than silently narrowing coverage without saying so.

## Implementation Record

Completed 2026-07-14. The installed CLI generates `catalog` and `fulfillment`, CRUD resources, rule,
workflow, event handler, recurring job, document flow, notification, and webhook output. The fixture
asserts every expected file, automatic module registration, tidy/build, and generated boot without
manual post-generation edits.

### Files changed

- `internal/cli/golden_consumer_test.go`
- generator implementation delivered by the wider W06 execution
- current story evidence and lifecycle records

### Interfaces introduced or changed

No task-local production interface; this task consumes the installed CLI surface.

### Tests added or modified

`TestGoldenConsumerInstalledBinaryTwoModules` and the generation phases shared by
`TestGoldenConsumerUpgradeReplay`.

### Implementation dates

2026-07-13 through 2026-07-14.

### Known limitations

None against AC-W06-E01-S002-02.

### Relationship to the approved plan

Matches the approved eight-subsystem, two-module scope. The earlier generator-surface blocker is
retained in story deviations/evidence as resolved history.
## Verification Record

### Actual result

PASS. The versioned installed CLI generated and exercised all eight named subsystem types across two
modules; all artifact and module-wiring assertions, build, and boot checks passed.

### Evidence identifier

EV-W06-E01-S002-007 and EV-W06-E01-S002-010.

### Execution date

2026-07-13T20:33:12Z.

### Commit or revision

Worktree snapshot based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`, content-pinned in
`artifacts/index.md`.

### Environment

Go 1.26.5; versioned local module proxy.

### Reviewer

W06-E01-S002-Verify.

### Retest status

Retested after the earlier failed generator-surface record.

### Final conclusion

T002 is done and AC-02 is verified.
## Deviations Record

DEV-W06-E01-S002-001 is resolved. No scope reduction was accepted; the completed implementation
matches the approved plan. See story `deviations.md`.
