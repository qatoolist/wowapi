---
id: W00-E02-S001-T001
type: task
title: Coverage baseline
status: done
parent_story: W00-E02-S001
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria:
  - AC-W00-E02-S001-01
artifacts: []
evidence: []
---

# W00-E02-S001-T001 — Coverage baseline

## Task Definition

*Per mandate §8.6. Defines the task before work begins.*

### Task objective

Measure current unit-test coverage against the real Postgres test DB and register the result as a
fresh evidence record, explicitly distinguishing it from the prior ~92%-measured/90%-floor figure
recorded in project history (which must be reconfirmed, not assumed, for this story's execution
commit).

### Parent story

W00-E02-S001 — Quality baselines.

### Owner

Unassigned.

### Status

`todo` (per `impl/governance/status-model.md` §7.3).

### Dependencies

None. This task can run independently of T002 and T003.

### Detailed work

1. Confirm test infrastructure is reachable: `docker compose -f deployments/compose.yaml up -d
   --wait postgres minio mailpit` (or confirm an already-running equivalent), and confirm
   `DATABASE_URL`/`TEST_DSN` resolves.
2. Run `make coverage-check` (`Makefile:241-246`), which internally runs `make coverage`
   (`DATABASE_URL=... WOWAPI_REQUIRE_DB=1 go test -coverprofile=coverage.out $(COVER_PKGS)`, where
   `COVER_PKGS` excludes `/cmd/wowapi`, `/internal/tools/migrate`, `/internal/testmodules`, and
   `/module$` per `Makefile:233-234`), then `go tool cover -html=coverage.out -o coverage.html`,
   then computes the `total:` percentage via `go tool cover -func=coverage.out` and compares
   against `COVERAGE_FLOOR` (90.0).
3. Capture the exact commit SHA the command was run against (`git rev-parse HEAD`).
4. Capture the Go toolchain version used (`go version`).
5. Record the measured coverage percentage as a fresh fact. Do not cite the prior ~92% figure as
   this task's result — that figure is prior project history, cited in `story.md` "Current-state
   assessment" as something to be reconfirmed, not as this task's own measurement.
6. If the measured percentage differs materially from the prior ~92% figure, note this explicitly
   in the evidence record as an observation (not necessarily a defect — coverage naturally shifts
   with the codebase) rather than silently reporting only the new number without context.
7. Register the result as an evidence record per `impl/governance/evidence-policy.md`'s required
   fields, and add an entry to `../evidence/index.md`.
8. Register `coverage.out`/`coverage.html` as artifacts per `impl/governance/artifact-policy.md`
   (authoritative path, generation command — per the no-duplication rule, the large `coverage.out`
   file itself is not copied into the `impl/` tree, only referenced), and add an entry to
   `../artifacts/index.md`.

### Expected files or components affected

None in the source tree. Output files: `coverage.out`, `coverage.html` (at the repository root,
per `make coverage-check`'s default behavior) — these are build artifacts, not source changes, and
are registered by path/command per the no-duplication rule, not copied into `impl/`.

### Expected output

A coverage-baseline evidence record stating the measured percentage, the exact command, the commit
SHA, environment, tool versions, date, and result (pass/fail against the 90.0% floor).

### Required artifacts

Coverage report (`coverage.out` + `coverage.html`, referenced by path).

### Required evidence

Coverage-baseline evidence record (type: coverage report).

### Related acceptance criteria

AC-W00-E02-S001-01.

### Completion criteria

The coverage-baseline evidence record exists in `../evidence/index.md` with all required fields
populated (per `impl/governance/evidence-policy.md`), citing an actual executed command and a real
commit SHA — not a projected or assumed result.

### Verification method

Per `../verification.md`'s AC-01 row: re-run (or review the run of) `make coverage-check` against
the real DB, confirm the reported percentage and floor-pass/fail status match what is recorded in
the evidence record.

### Risks

If Postgres is unreachable in the execution environment, this task cannot produce a real-DB-measured
result. Per `../plan.md`'s error-handling strategy, this must be recorded as a blocker on the task
(status `blocked`), not worked around with a mocked/partial measurement.

### Rollback or recovery considerations

Not applicable — this task performs no write to the repository or any persistent system beyond
generating local coverage-report files.

## Implementation Record

*Per mandate §8.7. Not yet executed — no implementation claims are pre-populated.*

### What was actually implemented

Baseline capture executed as specified: compose infra confirmed up (postgres + minio healthy), `make coverage-check` run at `0a31186cada5c275a588c74081cf977adf346e61`, output preserved, evidence record EV-W00-E02-S001-001 registered.

### Components changed

None.

### Files changed

None in the committed source tree. Evidence/artifact files written under this story directory only.

### Interfaces introduced or changed

None.

### Configuration changes

None.

### Schema or migration changes

None.

### Security changes

None.

### Observability changes

None.

### Tests added or modified

None.

### Commits

None made by this task (verification-only; the conductor owns commits). Executed against `0a31186cada5c275a588c74081cf977adf346e61`.

### Pull requests

None.

### Implementation dates

2026-07-13.

### Technical debt introduced

None.

### Known limitations

Point-in-time baseline; drifts the moment HEAD moves (inherent, per story.md residual-risk note). Concurrent sibling load present during the run (does not affect the coverage percentage, which is deterministic).

### Follow-up items

None.

### Relationship to the approved plan

Followed `../plan.md` implementation sequence step 1 exactly; no deviation.

## Verification Record

*Per mandate §8.8. Table below is planned before execution; fields after it are filled after
execution.*

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E02-S001-01 | Run `make coverage-check` against the real Postgres test DB; capture the `total:` percentage from `go tool cover -func=coverage.out`. | Local dev or CI container with Postgres reachable (`WOWAPI_REQUIRE_DB=1`), pinned Go toolchain. | A numeric coverage percentage is printed and recorded, compared against the 90.0% floor. | Coverage report | unassigned |

### Actual result

`total: (statements) 92.3%` — floor 90.0%, exit 0; all packages passed. Fresh measurement, consistent with (and superseding as citable baseline) the prior ~92% history figure.

### Pass or fail

**PASS.**

### Evidence identifier

EV-W00-E02-S001-001.

### Execution date

2026-07-13.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (main).

### Environment

Local dev workstation, macOS (Darwin 25.5.0) arm64, go1.26.5; real Postgres 16 via compose; concurrent sibling load present.

### Reviewer

Unassigned (conductor review gate pending).

### Findings

None — floor comfortably met.

### Retest status

Not required — first capture, no failed run to retest.

### Final conclusion

AC-W00-E02-S001-01 satisfied; coverage baseline registered.

## Deviations Record

*Per mandate §8.9. No deviations recorded yet.*

### Deviation ID

*Assign a stable deviation ID (`DEV-W00-E02-S001-T001-NNN`) if a deviation occurs.*

### Approved plan

*State what `../plan.md` said.*

### Actual implementation

*State what was actually implemented.*

### Reason

*State the reason for the deviation.*

### Impact

*State the impact of the deviation.*

### Risks

*State risks introduced by the deviation.*

### Approval

*State who approved the deviation and when.*

### Compensating controls

*State any compensating controls put in place.*

### Follow-up work

*State any follow-up work arising from the deviation.*
