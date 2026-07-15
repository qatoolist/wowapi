---
id: W06-E01-S002-T003
type: task
title: Boot and exercise against real infrastructure
status: done
parent_story: W06-E01-S002
owner: W06E01Impl
created_at: 2026-07-12
updated_at: 2026-07-14
depends_on:
  - W06-E01-S002-T002
acceptance_criteria:
  - AC-W06-E01-S002-03
artifacts:
  - ART-W06-E01-S002-003
evidence:
  - EV-W06-E01-S002-003
  - EV-W06-E01-S002-010
---

# W06-E01-S002-T003 — Boot and exercise against real infrastructure

## Task Definition

### Task objective

Boot the generated API and worker processes against real Postgres/MinIO/Mailpit/OTel; exercise authenticated CRUD, async delivery, restart/retry, and RLS isolation.

### Parent story

W06-E01-S002

### Owner
W06E01Impl

### Status

done

### Dependencies

W06-E01-S002-T002 (the generated modules must exist before they can be booted).

### Detailed work

1. Boot the generated API and worker processes against real Postgres, MinIO, Mailpit, and OTel.
2. Exercise authenticated CRUD against the generated resource(s).
3. Exercise async delivery (via the generated notification/webhook/recurring-job paths).
4. Exercise restart/retry (kill and restart the worker process mid-job).
5. Exercise RLS isolation (confirm tenant boundaries are enforced across the generated modules).

### Expected files or components affected

A boot-and-exercise integration test harness within the fixture.

### Expected output

All four exercise paths (authenticated CRUD, async delivery, restart/retry, RLS isolation) passing against real infrastructure.

### Required artifacts

ART-W06-E01-S002-003 (boot-and-exercise harness).

### Required evidence

EV-W06-E01-S002-003 (integration-test report).

### Related acceptance criteria

AC-W06-E01-S002-03.

### Completion criteria

All four exercise paths pass against real infrastructure.

### Verification method

Direct execution of the boot-and-exercise integration test.

### Risks

None beyond standard integration-test flakiness risk against real infrastructure.

### Rollback or recovery considerations

If a specific exercise path proves unreliable in CI, diagnose and fix root cause per `superpowers:systematic-debugging` discipline rather than silently removing the exercise.

## Implementation Record

Completed 2026-07-14 in `internal/cli/golden_consumer_infra_test.go`.

The harness provisions a disposable generated-consumer database; configures MinIO and OTLP; checks
MinIO, Mailpit, and Jaeger; migrates the generated consumer; starts its API and worker; exercises
API-key-authenticated create/read/update/list/delete; proves tenant B cannot read tenant A's row; drains
an event committed before worker startup; stops the worker; commits another event; and proves a
restarted worker dispatches the pending outbox row.

### Tests added or modified

`TestGoldenConsumerRealInfrastructure`; the same helper is invoked after the upgrade replay.

### Implementation dates

2026-07-14.

### Known limitations

None against AC-W06-E01-S002-03.

### Relationship to the approved plan

Matches the approved real-infrastructure CRUD, delivery, restart/retry, RLS, and OTel scope.
## Verification Record

### Actual result

PASS. Generated API and worker processes booted against real Postgres/MinIO/Mailpit/Jaeger. CRUD,
cross-tenant denial, outbox dispatch, worker stop/restart recovery, and final delete all passed.

### Evidence identifier

EV-W06-E01-S002-003 and EV-W06-E01-S002-010.

### Execution date

2026-07-13T20:33:12Z.

### Commit or revision

Worktree snapshot based on `733ef3e930cbb3f89f5bbc53d8f562c60e426513`, content-pinned in
`artifacts/index.md`.

### Environment

Go 1.26.5; Docker Compose Postgres, MinIO, Mailpit, and Jaeger.

### Reviewer

W06-E01-S002-Verify.

### Retest status

Retested after two offline-module-cache environment failures; final run passed.

### Final conclusion

T003 is done and AC-03 is verified.
## Deviations Record

No task-local deviation. The earlier story blocker was resolved before this task's passing run.
