---
id: VER-W00-E01-S003
type: verification-record
parent_story: W00-E01-S003
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W00-E01-S003

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story. Executed 2026-07-13; results in
the post-execution record below.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E01-S003-01 | Re-run `go test ./kernel/attachment/... ./kernel/notify/...` against testkit Postgres; confirm the fault-injection test in `kernel/attachment/coverage_test.go` proves rollback on the outbox-write error path, and confirm `kernel/notify/notify_test.go` proves the legal-delivery audit write succeeds via migration 00011's `events_outbox` INSERT grant | testkit Postgres via `make ci-container` or local `docker compose` | Exit code 0; fault-injection test proves rollback; audit write succeeds via migration 00011 grant | Test execution log (DB-gated) | unassigned |
| AC-W00-E01-S003-02 | Re-run the 20 S3-gated tests with `WOWAPI_REQUIRE_S3=1` against MinIO (via `make ci-container` or `docker compose` + `go test`); re-run the TOTP suite at 2 mocked clock/TZ settings; inspect `.github/workflows/ci.yml` for the 3-leg parallelized gate + toolbox image GHA-caching + docs-only skip (SD-01) and bench path-scoping/nightly/merge_group support (SD-02) | MinIO + Postgres via `make ci-container`; GitHub Actions workflow file inspection (`.github/workflows/ci.yml`, `deployments/compose.yaml`, `Makefile`) | 20/20 S3-gated tests pass; TOTP audit deterministic across both clock/TZ settings; `ci.yml` reflects SD-01/SD-02 state | S3-gated test-execution log + TOTP determinism test log + CI-configuration inspection note | unassigned |
| AC-W00-E01-S003-03 | Locate and re-run (or re-inspect, per the original verification basis) the test(s)/code path MATRIX cites for CS-03 (config fail-closed + fingerprint), CS-19 (i18n freeze + key-echo fallback), and CS-24 (SSRF dial-time guard); confirm each claim still holds at the story's closing commit | Environment per each claim's original verification basis (to be confirmed during Task 3 execution — Go toolchain at minimum; DB/network dependency per claim, not yet determined) | All three claims re-confirmed with an evidence pointer; any regression flagged as a new finding, not silently absorbed | Verify-outcome re-pin note with evidence pointers | unassigned |

## Post-execution record

Executed 2026-07-13, all against commit `0a31186cada5c275a588c74081cf977adf346e61` (branch `main`).

| Acceptance criterion | Result | Evidence ID | Raw evidence |
|---|---|---|---|
| AC-W00-E01-S003-01 | **pass** | EV-W00-E01-S003-01 | `evidence/logs/t001-db-gated-attachment-notify.log` |
| AC-W00-E01-S003-02 | **pass** (tests) / **confirmed, no drift** (CI inspection) | EV-W00-E01-S003-02 | `evidence/logs/t002-s3-gated-suite.log`, `t002-totp-tz-utc.log`, `t002-totp-tz-la.log`, `t002-ci-inspection-note.md` |
| AC-W00-E01-S003-03 | **pass — all three still hold** | EV-W00-E01-S003-03 | `evidence/logs/t003-cs-repin-note.md`, `t003-cs-repin-package-tests.log` |

### Actual result

- **AC-01 (DATA-08 W0):** `go test ./kernel/attachment/... ./kernel/notify/... -count=1 -v` with
  `WOWAPI_REQUIRE_DB=1` (skip-is-failure) — exit 0; 66 PASS / 0 FAIL / 0 SKIP.
  `TestAttachOutboxWriteErrorRollsBack` proves rollback on the outbox-write error path (W0-T1);
  `TestSendPendingLegalImportanceWritesAuditEvent` (+ negative control
  `TestSendPendingNonLegalImportanceWritesNoAuditEvent`) proves the legal-delivery audit write via
  migration 00011's `GRANT INSERT ON events_outbox TO app_platform` (W0-T2), which is present and
  unreverted at `migrations/00011_notify_webhook_integration.sql:178`.
- **AC-02 (REL-04 T1-T4 + SD-01/SD-02):** S3-gated suite with `WOWAPI_REQUIRE_S3=1` — exit 0,
  exactly 20 top-level tests PASS (count unchanged; no decrease), 0 SKIP. TOTP audit suite located
  at `kernel/mfa` (resolving the plan's open question); run at `TZ=UTC` and
  `TZ=America/Los_Angeles`, `-count=5` each — both exit 0, 245 PASS / 0 FAIL / 0 SKIP per TZ:
  deterministic. CI inspection (see `evidence/logs/t002-ci-inspection-note.md` for file:line
  citations): SD-01 (3 legs = gate matrix [test, race] + gate-bench; GHA-cached toolbox image;
  docs-only skip via the `changes` classifier) and SD-02 (bench path-scoped on PRs; nightly
  `cron "17 3 * * *"`; `merge_group` trigger) plus REL-04 T1 (Makefile/gate S3 env wiring), T2
  (minio `service_healthy`), T3 (canonical `S3_TEST_ENDPOINT` in all test wiring) — all CONFIRMED.
- **AC-03 (CS re-pins):** CS-03, CS-19, CS-24 re-confirmed against MATRIX's own citation basis
  (re-inspection of the cited file:line ranges in `kernel/config`, `kernel/i18n`,
  `kernel/httpclient`; majority verbatim, three single-line drifts with identical code, recorded);
  corroborating package tests exit 0, 199 PASS / 0 FAIL / 0 SKIP.

### Pass or fail

**Pass — all three acceptance criteria.** No `failed` evidence record exists; nothing was
retried-until-green.

### Evidence identifier

`EV-W00-E01-S003-01`, `EV-W00-E01-S003-02`, `EV-W00-E01-S003-03` — all registered with full
mandate-§10 fields in `evidence/index.md`.

### Execution date

2026-07-13, 12:07–12:15 +0530.

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`) — the same SHA for every run and
inspection; the repository did not advance during execution.

### Environment

Local macOS host (darwin/arm64, macOS 26.5.2, Darwin 25.5.0); go1.26.5 darwin/arm64; repo compose
stack (`postgres:16-alpine` at localhost:5432, `minio/minio:latest` at localhost:9000; Docker
29.4.0, Compose 5.3.1). `WOWAPI_REQUIRE_DB=1`/`WOWAPI_REQUIRE_S3=1` set so infrastructure
unavailability would fail loudly, never skip (ruling out the RISK-W00-002 false-negative mode —
both services were confirmed healthy). **Concurrent load present** (sibling W00 workers running
suites on the same machine); all evidence here is exit-code/functional, not timing-sensitive.

### Reviewer

Unassigned — acceptance is the conductor's review gate per the wave's status discipline; this
story does not self-assign a reviewer or self-mark `accepted`.

### Findings

No regression in any verified slice or matrix outcome. Two neutral observations, recorded rather
than silently absorbed: (1) the `kernel/mfa` suite grew from the 16 tests cited in the 2026-07-11
review to 49 top-level tests (an increase — only a decrease was flagged as potential regression);
(2) three MATRIX line citations drifted by one line each with identical code content (see
`evidence/logs/t003-cs-repin-note.md` "Line-drift note").

### Retest status

Not applicable — every command passed on its first and only execution.

### Final conclusion

All three acceptance criteria **satisfied** at `0a31186cada5c275a588c74081cf977adf346e61`. DATA-08
W0, REL-04 T1-T4, SD-01/SD-02, and the CS-03/CS-19/CS-24 verify-outcomes are all intact at current
HEAD; downstream waves (W04-E04-S001..S002, W07-E02-S002) may treat these slices as a proven
"before" state.
