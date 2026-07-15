---
id: PLAN-W00-E01-S003
type: plan
parent_story: W00-E01-S003
status: draft
created_at: 2026-07-12
updated_at: 2026-07-12
derived: false
---

# Plan — W00-E01-S003

Per mandate §8.5. This plan describes the proposed *verification* approach — no code changes are
expected from this story. Per mandate §8.5, verbatim: "Do not invent precise code changes where the
repository does not yet provide enough information. Clearly distinguish confirmed facts, planned
changes, and implementation assumptions."

## Proposed architecture

Not applicable in the code-change sense — this story makes no architectural change. The "approach
architecture" here is the verification structure itself: three tasks, each producing one
mandate-§10-conformant evidence record against one bounded slice of claimed-executed work (DATA-08
W0; REL-04 T1-T4 + CI-pipeline-state; CS-03/CS-19/CS-24 re-pin), each independently executable and
independently reviewable.

## Implementation strategy

Re-run the exact test commands named in the source material against the current repository HEAD,
inspect the exact configuration files named for the CI-pipeline-state and matrix-outcome checks, and
register one evidence record per acceptance criterion. Where the exact file, line, or test name
cited by the source documents cannot be confirmed without inspection (e.g. the TOTP audit suite's
exact path), the task records what was found during execution rather than assuming the prior
citation is still accurate.

## Expected package or module changes

None. This is a verification-only story; no package or module is expected to change.

## Expected file changes where determinable

None. No file is expected to change as a result of this story's execution. The files this story
*reads and re-tests* (not changes) are listed in `story.md` "Affected packages or components."

## Contracts and interfaces

Not applicable. No interface is introduced or changed.

## Data structures

Not applicable. No data structure is introduced or changed.

## APIs

Not applicable. No API is introduced or changed.

## Configuration changes

None expected. Task 2 inspects (does not modify) `.github/workflows/ci.yml`, `deployments/
compose.yaml`, and `Makefile` for the SD-01/SD-02 CI-pipeline state and the REL-04 T1-T3 wiring.

## Persistence changes

None. Task 1 depends on migration `00011` already being applied in the test environment (a confirmed
precondition, not a change this story makes) but performs no new migration.

## Migration strategy

Not applicable. No migration is created, modified, or re-run by this story.

## Concurrency implications

None expected beyond whatever concurrency behavior the pre-existing tests already exercise (e.g. the
fault-injection test in `kernel/attachment/coverage_test.go`, if it exercises concurrent writers —
to be confirmed during Task 1 execution, not assumed).

## Error-handling strategy

Not applicable to the story's own execution. Task 1 specifically re-verifies an error-handling
property of production code (that the outbox-write error is no longer discarded) — see Task 1 detail
below.

## Security controls

No new security control is introduced. Task 2 re-verifies REL-04 T4's TOTP-audit determinism (a
verification, not an introduction, of a security-relevant property) and Task 3 re-pins three
security-relevant matrix verify-outcomes (CS-03, CS-19, CS-24) without modifying their underlying
controls.

## Observability changes

None.

## Testing strategy

- **Task 1 (DATA-08 W0)**: `go test ./kernel/attachment/... ./kernel/notify/...` against testkit
  Postgres (DB-gated). Confirms the fault-injection rollback test and the legal-delivery
  audit-write test both pass.
- **Task 2 (REL-04 T1-T4 + CI-pipeline-state)**: the 20 S3-gated tests via `make ci-container` (or
  `docker compose up` + `go test ... ` with `WOWAPI_REQUIRE_S3=1` and `S3_TEST_ENDPOINT` set) against
  MinIO; the TOTP audit suite run twice, once per each of two distinct mocked clock/timezone
  settings; a file-inspection pass over `.github/workflows/ci.yml`, `deployments/compose.yaml`, and
  `Makefile` for the REL-04 T1-T3 wiring and the SD-01/SD-02 CI state.
- **Task 3 (CS-03/CS-19/CS-24 re-pin)**: locate and re-run (or re-inspect, if the original
  verification was inspection-based) whatever test(s) or code path MATRIX cites as the basis for each
  of the three `INV→verified` claims, and record the evidence pointer confirming each still holds.

No new tests are written by this story — mandate §13: "Do not create tests merely to increase
numerical coverage," and this story's entire purpose is to prove existing tests still hold, not to
add coverage.

## Regression strategy

If any task's re-run fails, the evidence record is registered with status `failed` (not deleted or
silently retried), and a follow-up remediation task is opened under that finding's canonical target
story (`W04-E04-S001..S002` for DATA-08, `W07-E02-S002` for REL-04) per `requirement-inventory.md`.
A CS-03/19/24 regression is escalated immediately as a new finding (see `story.md` "Risks") rather
than folded into routine follow-up handling, given its security nature.

## Compatibility strategy

Not applicable. No code change occurs.

## Rollout strategy

Not applicable. This story produces evidence records, not a deployable change.

## Rollback strategy

Not applicable in the code-rollback sense. If a re-verification fails, the "rollback" is simply: the
story's affected acceptance criterion is not marked `pass`, the story does not move to `accepted`,
and the failure is handled per "Regression strategy" above.

## Implementation sequence

1. Confirm testkit Postgres and MinIO are both available in the execution environment (this story's
   one story-specific precondition — see `story.md` "Assumptions"; it is the only story in the epic
   requiring both simultaneously).
2. Execute Task 1 (DATA-08 W0) — no dependency on Task 2 or 3.
3. Execute Task 2 (REL-04 T1-T4 + CI-pipeline-state) — no dependency on Task 1 or 3.
4. Execute Task 3 (CS-03/CS-19/CS-24 re-pin) — no hard dependency on Task 1 or 2; a soft ordering
   convenience exists in that Task 2 already opens `.github/workflows/ci.yml` for the SD-01/SD-02
   check, which is adjacent to (but not required by) CS-24's SSRF dial-time guard inspection.
5. Register evidence for all three tasks in `evidence/index.md`; update `verification.md`'s
   post-execution record; update `story.md` front-matter status per the outcome.

Tasks 1-3 may equally be executed in parallel, in any order, or interleaved — none blocks another.

## Task breakdown

- **W00-E01-S003-T001** — Re-verify DATA-08 W0 (attachment/notify durability). Related AC:
  AC-W00-E01-S003-01.
- **W00-E01-S003-T002** — Re-verify REL-04 T1-T4 (S3/TOTP wiring) and confirm the SD-01/SD-02
  CI-pipeline-state. Related AC: AC-W00-E01-S003-02.
- **W00-E01-S003-T003** — Re-pin CS-03/CS-19/CS-24 matrix verify-outcomes. Related AC:
  AC-W00-E01-S003-03.

## Expected artifacts

- DB-gated test-execution log (attachment/notify) — Task 1.
- S3-gated test-execution log (20 tests) — Task 2.
- TOTP determinism test log (2 clock/TZ settings) — Task 2.
- CI-configuration inspection note (`.github/workflows/ci.yml` SD-01/SD-02 state) — Task 2.
- CS-03/CS-19/CS-24 verify-outcome re-pin note — Task 3.

## Expected evidence

- `EV-W00-E01-S003-01` (planned) — AC-01, DB-gated test-execution log.
- `EV-W00-E01-S003-02` (planned) — AC-02, S3-gated + TOTP test-execution logs plus CI-inspection
  note.
- `EV-W00-E01-S003-03` (planned) — AC-03, verify-outcome re-pin note.

## Unresolved questions

- The exact file path of the TOTP audit test suite is not yet pinned in the source material cited
  for this story — Task 2 must locate it during execution and record the actual path found, not
  assume the prior citation without confirming it.
- Whether `kernel/attachment/coverage_test.go`'s fault-injection test requires testkit Postgres, or
  can run with a mocked outbox writer, is not yet confirmed — Task 1 must determine this and record
  the actual environment used, not assume DB-gating without confirming it.
- The exact test(s) or inspection basis MATRIX cites for CS-03, CS-19, and CS-24 individually is not
  restated in full here — Task 3 must locate MATRIX's specific citation for each and follow it, not
  invent a new verification method.
- Whether the S3-gated test count is still exactly 20 at this story's execution commit (as named in
  the source material) must be confirmed, not assumed; if the count differs, Task 2 records the
  actual count found and treats a *decrease* as a potential regression to flag.

## Approval conditions

This plan is considered approved and ready for implementation once: an owner is assigned to the
story; the Postgres + MinIO test-environment precondition (Implementation sequence, step 1) is
confirmed available; and no reviewer objection is raised to the three-task breakdown (in particular,
to Task 3 being split out as its own task rather than folded into Task 2 — see `story.md`'s judgment
call, restated in this story directory's governing instructions, on how the CS-03/CS-19/CS-24 re-pin
is organized within this story's task structure).
