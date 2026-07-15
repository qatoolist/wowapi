---
id: W00-E01-S003-T001
type: task
title: Re-verify DATA-08 W0 (attachment/notify durability)
status: done
parent_story: W00-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria: [AC-W00-E01-S003-01]
artifacts: [ART-W00-E01-S003-001]
evidence: [EV-W00-E01-S003-01]
---

# W00-E01-S003-T001 — Re-verify DATA-08 W0 (attachment/notify durability)

## Task Definition

*Per mandate §8.6. This file defines the task before work begins. The implementation record,
verification record, and deviations record for this task are the `##`-level sections below in this
same file, per `governance/naming-conventions.md` "Adaptation 1" (flat single-file tasks).*

### Task objective

Re-run `go test ./kernel/attachment/... ./kernel/notify/...` against testkit Postgres and confirm
DATA-08's W0 slice — W0-T1 (attachment outbox-write error no longer discarded) and W0-T2
(legal-delivery audit write via migration 00011's `events_outbox` INSERT grant) — is still intact at
the current repository HEAD, registering mandate-§10-conformant evidence for the result.

### Parent story

W00-E01-S003 — Verify data-durability and CI-integration slices at current HEAD.

### Owner

unassigned

### Status

`done` — executed and evidenced 2026-07-13; awaiting the conductor's story-level review gate.

### Dependencies

None. This task targets `kernel/attachment/` and `kernel/notify/`, disjoint from Task 2's
`kernel`-adjacent CI/S3/TOTP scope and Task 3's matrix-outcome scope. It may execute in any order
relative to T002 and T003, including fully in parallel.

### Detailed work

- Confirm `kernel/attachment/attachment.go`'s outbox-write error path still propagates the error
  rather than discarding it (DATA-08 W0-T1).
- Confirm `kernel/attachment/coverage_test.go` contains a fault-injection test that proves rollback
  occurs when the outbox write fails, and that this test currently passes.
- Confirm `kernel/notify/service.go`'s legal-delivery audit write still uses the `events_outbox`
  INSERT permission granted by migration `00011`, and that this permission is still granted in the
  current migration set (i.e. migration `00011` has not been reverted, renumbered, or superseded
  without an equivalent replacement).
- Confirm `kernel/notify/notify_test.go` still covers the legal-delivery audit write path and that
  this test currently passes.
- Run `go test ./kernel/attachment/... ./kernel/notify/...` against testkit Postgres (this is a
  DB-gated test run — confirm the DB dependency during execution per `plan.md` "Unresolved
  questions": whether the fault-injection test specifically requires Postgres or can run against a
  mocked outbox writer must be determined, not assumed).
- Record the exact commit SHA the run was executed against, the exact command, the environment
  (testkit Postgres via `make ci-container` or local `docker compose`), tool versions, and result.

### Expected files or components affected

None changed. Files read and re-tested: `kernel/attachment/attachment.go`,
`kernel/attachment/coverage_test.go`, `kernel/notify/service.go`, `kernel/notify/notify_test.go`, and
the migration file numbered `00011`.

### Expected output

A `pass` or `failed` result for `go test ./kernel/attachment/... ./kernel/notify/...`, with the test
output confirming or refuting the two specific behavioral claims (fault-injection rollback;
audit-write-via-grant), captured as a test-execution log artifact and a corresponding evidence
record.

### Required artifacts

DB-gated test-execution log (attachment/notify) — see `../../artifacts/index.md`.

### Required evidence

One evidence record, planned ID `EV-W00-E01-S003-01`, evidence type "test execution log (DB-gated)"
— see `../../evidence/index.md`.

### Related acceptance criteria

AC-W00-E01-S003-01.

### Completion criteria

This task is complete when: the test command has actually been executed (not merely cited) against
a confirmed commit SHA; the result (pass/fail) is recorded in `verification.md` and in the story's
`verification.md`; the evidence record is registered in `../../evidence/index.md` with all required
fields per `evidence-policy.md`; and, if the result is `failed`, a follow-up remediation task has
been opened under `W04-E04-S001..S002` (DATA-08's canonical target per `requirement-inventory.md`)
rather than silently retried until green.

### Verification method

Direct re-execution of `go test ./kernel/attachment/... ./kernel/notify/...` against testkit
Postgres; inspection of test source to confirm the fault-injection and audit-write assertions are
present in the current test files (not merely that the package compiles and some test passes).

### Risks

- RISK-W00-001 (inherited) — DATA-08 W0 fails to re-verify at current HEAD; would block
  W04-E04-S001..S002's hash-widening work, which assumes this durability fix is intact.
- RISK-W00-002 (inherited) — testkit Postgres unavailable or misconfigured in the execution
  environment, producing a false-negative regression; must be ruled out (by confirming Postgres
  health) before treating any failure as a genuine regression.

### Rollback or recovery considerations

Not applicable in the code sense (this task changes no code). If the re-verification fails, no
rollback occurs — the failure is recorded as `failed`-status evidence (preserved, not deleted per
`evidence-policy.md`) and a new remediation task is opened under `W04-E04-S001..S002`, distinct from
this task, which remains `done` in the sense of "executed and evidenced," not "regression silently
fixed here."

## Implementation Record

*Per mandate §8.7.* Executed 2026-07-13 against commit
`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### What was actually implemented

Verification-only execution; no implementation. Pre-run source inspection confirmed every claim
before testing: `kernel/attachment/attachment.go:82-93` propagates the outbox-write error
(`kerr.Wrapf(err, "attachment.Attach", "write outbox event")` — not discarded);
`kernel/attachment/coverage_test.go:19-26` defines the `failingOutboxWriter` fault-injection
double and `:325-367` `TestAttachOutboxWriteErrorRollsBack` asserts error propagation, KindInternal
wrapping, AND zero persisted attachment rows (whole-transaction rollback);
`kernel/notify/service.go:584-597` writes the `notify.legal_delivery` outbox audit event in the
same transaction as the 'sent' status update, with `migrations/00011_notify_webhook_integration.sql:178`
(`GRANT INSERT ON events_outbox TO app_platform`) still present, unreverted;
`kernel/notify/notify_test.go:646` `TestSendPendingLegalImportanceWritesAuditEvent` and `:730`
`TestSendPendingNonLegalImportanceWritesNoAuditEvent` cover the audit-write path both ways.
Then ran the full DB-gated command (see Verification Record).

### Components changed

None — verification-only task, as planned.

### Files changed

None. Only this story directory's own governance/evidence files were written.

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None. Migration 00011 was confirmed present and unmodified, not re-run (testkit applies the
migration set into its template database automatically).

### Security changes

None.

### Observability changes

None.

### Tests added or modified

None — existing tests re-run, not modified.

### Commits

None made by this task (read-only against `0a31186cada5c275a588c74081cf977adf346e61`).

### Pull requests

None.

### Implementation dates

2026-07-13 (single session).

### Technical debt introduced

None.

### Known limitations

Point-in-time re-verification: proves DATA-08 W0 intact at `0a31186`, not permanently.

### Follow-up items

None — result was `pass`; no remediation task needed under W04-E04-S001..S002.

### Relationship to the approved plan

Executed exactly as `../../plan.md` Testing strategy describes, with the plan's own allowed
alternative environment (compose services + host-side `go test`, rather than `make ci-container`).
The plan's unresolved question "does the fault-injection test require Postgres?" is answered: yes —
`TestAttachOutboxWriteErrorRollsBack` calls `testkit.NewDB(t)` (real Postgres; the *outbox writer*
is the injected fake, the transaction/rollback semantics are real DB behavior).

## Verification Record

*Per mandate §8.8. Table below is planned before execution; fields after it are filled after
execution.*

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E01-S003-01 | Re-run `go test ./kernel/attachment/... ./kernel/notify/...` against testkit Postgres; confirm fault-injection rollback test and legal-delivery audit-write test both pass | testkit Postgres via `make ci-container` or local `docker compose` | Exit code 0; fault-injection test proves rollback; audit write succeeds via migration 00011 grant | Test execution log (DB-gated) | unassigned |

### Actual result

`go test ./kernel/attachment/... ./kernel/notify/... -count=1 -v` (with `WOWAPI_REQUIRE_DB=1`,
`DATABASE_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable`) exited 0.
66 tests PASS, 0 FAIL, 0 SKIP (`WOWAPI_REQUIRE_DB=1` guarantees no silent skip).
`TestAttachOutboxWriteErrorRollsBack` PASS — fault-injection proves rollback on the outbox-write
error path (W0-T1). `TestSendPendingLegalImportanceWritesAuditEvent` PASS — legal-delivery audit
write succeeds via migration 00011's `events_outbox` INSERT grant (W0-T2); the negative control
`TestSendPendingNonLegalImportanceWritesNoAuditEvent` also PASS.

### Pass or fail

**Pass.**

### Evidence identifier

`EV-W00-E01-S003-01` — registered in `../../evidence/index.md`; raw log
`../../evidence/logs/t001-db-gated-attachment-notify.log`.

### Execution date

2026-07-13 12:07 +0530.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### Environment

Local macOS host (darwin/arm64, macOS 26.5.2) against the repo compose Postgres
(`postgres:16-alpine`, localhost:5432, Docker 29.4.0); go1.26.5. Concurrent load present (sibling
W00 workers) — evidence is exit-code/functional, not timing-sensitive.

### Reviewer

Unassigned — acceptance is the conductor's review gate; not self-assigned.

### Findings

No regression. DATA-08 W0-T1 and W0-T2 both intact at `0a31186`.

### Retest status

Not applicable — first run passed; nothing retried.

### Final conclusion

AC-W00-E01-S003-01 **satisfied**: DATA-08's W0 slice re-verified at current HEAD with
mandate-§10-conformant evidence.

## Deviations Record

*Per mandate §8.9.*

**No deviations.** The command, environment posture, and verification method match the task
definition; the compose-vs-ci-container environment choice is within the options the task
definition itself names ("via `make ci-container` or local `docker compose`").

### Deviation ID

Not applicable — no deviations occurred.

### Approved plan

Not applicable — executed as planned.

### Actual implementation

Not applicable — matches the plan.

### Reason

Not applicable.

### Impact

Not applicable.

### Risks

Not applicable.

### Approval

Not applicable.

### Compensating controls

Not applicable.

### Follow-up work

Not applicable.
