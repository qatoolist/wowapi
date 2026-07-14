---
id: W00-E01-S003-T002
type: task
title: Re-verify REL-04 T1-T4 (S3/TOTP wiring) and confirm SD-01/SD-02 CI-pipeline-state
status: done
parent_story: W00-E01-S003
owner: unassigned
created_at: 2026-07-12
updated_at: 2026-07-13
depends_on: []
acceptance_criteria: [AC-W00-E01-S003-02]
artifacts: [ART-W00-E01-S003-002, ART-W00-E01-S003-003, ART-W00-E01-S003-004]
evidence: [EV-W00-E01-S003-02]
---

# W00-E01-S003-T002 — Re-verify REL-04 T1-T4 (S3/TOTP wiring) and confirm SD-01/SD-02 CI-pipeline-state

## Task Definition

*Per mandate §8.6. This file defines the task before work begins. The implementation record,
verification record, and deviations record for this task are the `##`-level sections below in this
same file, per `governance/naming-conventions.md` "Adaptation 1" (flat single-file tasks).*

### Task objective

Re-run the 20 S3-gated tests with `WOWAPI_REQUIRE_S3=1` against MinIO, re-run the TOTP audit suite
at two distinct mocked clock/timezone settings, and inspect `.github/workflows/ci.yml`,
`deployments/compose.yaml`, and `Makefile` to confirm REL-04 T1-T4 and the SD-01/SD-02 CI-pipeline
state are still intact at the current repository HEAD, registering mandate-§10-conformant evidence
for the result.

### Parent story

W00-E01-S003 — Verify data-durability and CI-integration slices at current HEAD.

### Owner

unassigned

### Status

`done` — executed and evidenced 2026-07-13; awaiting the conductor's story-level review gate.

### Dependencies

None. This task targets `Makefile`, `deployments/compose.yaml`, `.github/workflows/ci.yml`, and the
S3-gated/TOTP test suites — disjoint from Task 1's `kernel/attachment`/`kernel/notify` scope. It may
execute in any order relative to T001 and T003, including fully in parallel. A soft, non-blocking
convenience overlap exists with T003: this task already opens `.github/workflows/ci.yml` for the
SD-01/SD-02 inspection, which is adjacent to (but does not satisfy or replace) T003's CS-24 SSRF
dial-time-guard inspection.

### Detailed work

- Confirm `Makefile`'s `ci-container` target and the hosted `gate` job wire `WOWAPI_REQUIRE_S3=1` and
  `S3_TEST_ENDPOINT` (REL-04 T1).
- Confirm `deployments/compose.yaml`'s minio service declares a `service_healthy` condition (REL-04
  T2).
- Confirm the canonical `S3_ENDPOINT` variable naming is consistent across the wiring (no leftover
  divergent variable name) (REL-04 T3).
- Run the 20 S3-gated tests via `make ci-container` (or `docker compose up` plus `go test` with
  `WOWAPI_REQUIRE_S3=1` set) against MinIO. Confirm the count is still 20 as named in the source
  material; if it differs, record the actual count found and treat a *decrease* as a potential
  regression to flag, not silently accept.
- Locate the TOTP audit test suite (exact path not yet pinned — confirm during execution) and run it
  at two distinct mocked clock/timezone settings, confirming the audit path is deterministic and not
  wall-clock-dependent (REL-04 T4).
- Inspect `.github/workflows/ci.yml` to confirm: the CI gate is parallelized into 3 legs; the toolbox
  image is GHA-cached; a docs-only-change skip exists (SD-01). Confirm: the benchmark job is
  path-scoped on pull requests; a nightly schedule exists; `merge_group` is supported (SD-02).
- Record the exact commit SHA the run/inspection was executed against, the exact commands, the
  environment, tool versions, and result for each sub-check.

### Expected files or components affected

None changed. Files read and re-tested/inspected: `Makefile`, `deployments/compose.yaml`,
`.github/workflows/ci.yml`, the S3-gated test suite files, and the TOTP audit test suite (path to be
confirmed during execution).

### Expected output

A `pass` or `failed` result for the S3-gated test suite (20/20) and the TOTP determinism check (2/2
clock-TZ settings), plus a `confirmed` or `drifted` result for the SD-01/SD-02 CI-pipeline-state
inspection, captured as test-execution logs and a CI-configuration inspection note, with a
corresponding evidence record.

### Required artifacts

S3-gated test-execution log (20 tests); TOTP determinism test log (2 clock/TZ settings);
CI-configuration inspection note (`.github/workflows/ci.yml` SD-01/SD-02 state) — see
`../../artifacts/index.md`.

### Required evidence

One evidence record, planned ID `EV-W00-E01-S003-02`, evidence type "S3-gated test-execution log +
TOTP determinism test log + CI-configuration inspection note" — see `../../evidence/index.md`.

### Related acceptance criteria

AC-W00-E01-S003-02.

### Completion criteria

This task is complete when: the S3-gated test suite and the TOTP suite (at both clock/TZ settings)
have actually been executed (not merely cited) against a confirmed commit SHA; the
`.github/workflows/ci.yml` inspection has actually been performed against that same commit; all
results are recorded in `verification.md` and in the story's `verification.md`; the evidence record
is registered in `../../evidence/index.md` with all required fields per `evidence-policy.md`; and,
if any result is `failed` or `drifted`, a follow-up remediation task has been opened under
`W07-E02-S002` (REL-04's canonical target per `requirement-inventory.md`) for a REL-04 regression, or
flagged as a new finding (not silently accepted) for an SD-01/SD-02 drift.

### Verification method

Direct re-execution of the 20 S3-gated tests and the TOTP suite (2 clock/TZ settings) against a live
MinIO + Postgres environment; direct file inspection of `.github/workflows/ci.yml`,
`deployments/compose.yaml`, and `Makefile` against the specific SD-01/SD-02 claims listed in
`requirement-inventory.md` §E.

### Risks

- RISK-W00-001 (inherited) — REL-04 T1-T4 fails to re-verify at current HEAD; would block
  W07-E02-S002's T5-T8 fuzz remainder, which assumes this S3/TOTP/CI wiring is the correct "before"
  state.
- RISK-W00-002 (inherited) — this task carries the epic's highest exposure to test-infrastructure
  unavailability, since it is the only task in the epic requiring both Postgres and MinIO
  simultaneously; must confirm both services' health before treating any failure as genuine.
- SD-01/SD-02 drift risk (story-specific, not yet assigned a RISK ID) — if the CI pipeline has
  changed since the session-delta facts were recorded (e.g. a leg was merged back, caching was
  removed), this is a baseline-accuracy risk for any later wave that references the SD-01/SD-02
  facts as "current state," not merely a REL-04 regression.

### Rollback or recovery considerations

Not applicable in the code sense (this task changes no code). If the re-verification fails, no
rollback occurs — the failure is recorded as `failed`-status evidence (preserved, not deleted per
`evidence-policy.md`) and a new remediation task is opened under `W07-E02-S002` for a REL-04
regression; an SD-01/SD-02 drift is recorded as a finding against the baseline claim, not folded
into the REL-04 remediation path.

## Implementation Record

*Per mandate §8.7.* Executed 2026-07-13 against commit
`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### What was actually implemented

Verification-only execution; no implementation. Sub-results:

- **REL-04 T1 (Makefile/gate wiring):** `Makefile:313-314` `ci-container` sets
  `WOWAPI_REQUIRE_DB=1 -e WOWAPI_REQUIRE_S3=1 -e S3_TEST_ENDPOINT=minio:9000`; all three parallel
  legs (`ci-container-test/race/bench`, `Makefile:324-333`) carry the identical posture; the
  hosted `gate` job invokes `make ci-container-${{ matrix.leg }}` (`ci.yml:227`). CONFIRMED.
- **REL-04 T2 (minio health):** `deployments/compose.yaml:33-36` minio healthcheck; `:95-99`
  `tools.depends_on.minio.condition: service_healthy`. CONFIRMED.
- **REL-04 T3 (canonical variable):** repo-wide grep — test wiring uses `S3_TEST_ENDPOINT`
  exclusively (Makefile, compose tools service, `s3_test.go:39-43`); no `S3_ENDPOINT` remains in
  `deployments/compose.yaml`; remaining `S3_ENDPOINT` occurrences are the distinct runtime product
  variable (`product-dev.yaml`, docs) exactly as the 2026-07-11 cleanup left them. CONFIRMED.
- **REL-04 T4 (TOTP determinism):** located the TOTP audit suite at `kernel/mfa` (the path
  unresolved in `plan.md` "Unresolved questions" — the suite the 2026-07-11 review ran; it has
  since grown from 16 to 49 top-level tests, an increase, not a regression). Ran it at `TZ=UTC`
  and `TZ=America/Los_Angeles`, `-count=5` each: both exit 0, 245/245 PASS per TZ, 0 SKIP —
  deterministic, no wall-clock dependence.
- **S3-gated suite:** exactly **20** top-level tests (count unchanged from the source material —
  no decrease to flag), all PASS, 0 SKIP, exit 0, under `WOWAPI_REQUIRE_S3=1` (skip-is-failure).
- **SD-01:** gate parallelized into 3 container legs (`gate` matrix `[test, race]`,
  `ci.yml:190-193`, + `gate-bench`, `:236-280`); toolbox image GHA-layer-cached
  (`cache-from/to: type=gha,scope=toolbox`, `ci.yml:217-218`); docs-only skip via the `changes`
  classifier (`code=false`, `ci.yml:114-117`) gating `gate`/`reference-smoke`/`coverage`
  (`:187,285,304`). CONFIRMED.
- **SD-02:** bench path-scoped on PRs (`bench_re`, `ci.yml:128-135`; `gate-bench`
  `if: …bench == 'true'`, `:245`); nightly `cron: "17 3 * * *"` (`:46`); `merge_group:` trigger
  (`:41`) with merge-group-aware diff base (`:85-91`). CONFIRMED.

Full citations in `../../evidence/logs/t002-ci-inspection-note.md`.

### Components changed

None — verification-only task, as planned.

### Files changed

None. Only this story directory's own governance/evidence files were written.

### Interfaces introduced or changed

None.

### Configuration changes

None — `.github/workflows/ci.yml`, `deployments/compose.yaml`, and `Makefile` were inspected,
not modified.

### Schema or migration changes

None.

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

Point-in-time re-verification at `0a31186`; TOTP determinism proven across two TZ settings ×5
iterations, not a formal proof over all clock states (same basis as the original verification).

### Follow-up items

None — all sub-checks passed/confirmed; no remediation task needed under W07-E02-S002 and no
SD-01/SD-02 drift finding to raise.

### Relationship to the approved plan

Executed per `../../plan.md` Testing strategy, using the plan's explicitly allowed alternative to
`make ci-container` (compose services + host-side `go test` with `WOWAPI_REQUIRE_S3=1` and
`S3_TEST_ENDPOINT` set). Both plan unresolved questions owned by this task are answered: the TOTP
suite path is `kernel/mfa`; the S3-gated test count is still exactly 20.

## Verification Record

*Per mandate §8.8. Table below is planned before execution; fields after it are filled after
execution.*

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E01-S003-02 | Re-run the 20 S3-gated tests with `WOWAPI_REQUIRE_S3=1` against MinIO; re-run the TOTP suite at 2 mocked clock/TZ settings; inspect `.github/workflows/ci.yml` for SD-01/SD-02 state | MinIO + Postgres via `make ci-container`; GitHub Actions workflow file inspection | 20/20 S3-gated tests pass; TOTP deterministic across both clock/TZ settings; `ci.yml` reflects SD-01/SD-02 state | S3-gated test-execution log + TOTP determinism test log + CI-configuration inspection note | unassigned |

### Actual result

S3-gated suite (`WOWAPI_REQUIRE_S3=1 WOWAPI_REQUIRE_DB=1 S3_TEST_ENDPOINT=localhost:9000 go test
github.com/qatoolist/wowapi/adapters/storage/s3 -count=1 -v`): exit 0, 20/20 top-level PASS, 0
SKIP. TOTP (`TZ=UTC` / `TZ=America/Los_Angeles`, `go test github.com/qatoolist/wowapi/kernel/mfa
-count=5 -v`): both exit 0, 245 PASS / 0 FAIL / 0 SKIP each — deterministic. CI inspection:
SD-01, SD-02, REL-04 T1/T2/T3 all CONFIRMED with file:line citations; no drift.

### Pass or fail

**Pass** (test re-runs) / **confirmed, no drift** (CI-pipeline-state inspection).

### Evidence identifier

`EV-W00-E01-S003-02` — registered in `../../evidence/index.md`; raw files
`../../evidence/logs/t002-s3-gated-suite.log`, `t002-totp-tz-utc.log`, `t002-totp-tz-la.log`,
`t002-ci-inspection-note.md`.

### Execution date

2026-07-13 12:10 +0530.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

### Environment

Local macOS host (darwin/arm64, macOS 26.5.2) against the repo compose MinIO
(`minio/minio:latest`, localhost:9000) and Postgres (`postgres:16-alpine`, localhost:5432),
Docker 29.4.0 / Compose 5.3.1; go1.26.5. Concurrent load present (sibling W00 workers) —
evidence is exit-code/functional, not timing-sensitive.

### Reviewer

Unassigned — acceptance is the conductor's review gate; not self-assigned.

### Findings

No regression, no drift. One neutral observation: the `kernel/mfa` suite has grown from the 16
tests the 2026-07-11 review cited to 49 top-level tests — an increase (more coverage), recorded
here per the "treat a decrease as a potential regression" instruction (its inverse needs no flag).

### Retest status

Not applicable — first runs passed; nothing retried.

### Final conclusion

AC-W00-E01-S003-02 **satisfied**: REL-04 T1-T4 and the SD-01/SD-02 CI-pipeline state re-verified
at current HEAD with mandate-§10-conformant evidence.

## Deviations Record

*Per mandate §8.9.*

**No deviations.** The S3/TOTP re-runs used the task definition's own named alternative
environment (`docker compose` services + `go test` with the required env), and the inspection
covered exactly the files and claims listed in "Detailed work."

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
