---
id: W00-E01-S003-EVIDENCE-INDEX
type: evidence-index
parent_story: W00-E01-S003
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W00-E01-S003 — Evidence index

Per `governance/evidence-policy.md` required fields. All three records below were produced by
actual execution on 2026-07-13 against commit `0a31186cada5c275a588c74081cf977adf346e61`
(branch `main`). Raw logs and inspection notes live in `logs/` (created on first real content per
`naming-conventions.md` Adaptation 2).

Shared environment for all three records: local macOS host (darwin/arm64, macOS 26.5.2,
Darwin kernel 25.5.0); Postgres + MinIO from the repo compose stack
(`deployments/compose.yaml`, services already up and healthy: `postgres:16-alpine`,
`minio/minio:latest`, via Docker 29.4.0 / Compose 5.3.1), reached host-side at
`localhost:5432` / `localhost:9000`. **Concurrent load present** (sibling W00 workers running
test suites on the same machine) — all evidence in this story is functional/exit-code based, not
timing-sensitive. Tool versions: go1.26.5 darwin/arm64; Docker 29.4.0; Docker Compose 5.3.1.

| Evidence ID | Evidence type | Story / task | Acceptance criteria proven | Execution command | Code revision / commit SHA | Branch / tag | Execution environment | Tool versions | Date / time | Result | File / URI | Checksum | Reviewer | Superseded evidence |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| EV-W00-E01-S003-01 | Test execution log (DB-gated) | W00-E01-S003 / W00-E01-S003-T001 | AC-W00-E01-S003-01 | `WOWAPI_REQUIRE_DB=1 DATABASE_URL='postgres://wowapi:wowapi-local-only@localhost:5432/wowapi?sslmode=disable' go test ./kernel/attachment/... ./kernel/notify/... -count=1 -v` | 0a31186cada5c275a588c74081cf977adf346e61 | main | Local host vs compose testkit Postgres (localhost:5432); WOWAPI_REQUIRE_DB=1 so a missing DB fails, never skips; concurrent load present | go1.26.5 darwin/arm64; postgres:16-alpine; Docker 29.4.0 | 2026-07-13 12:07 +0530 | **pass** — exit 0; 66 PASS / 0 FAIL / 0 SKIP; `TestAttachOutboxWriteErrorRollsBack` PASS (DATA-08 W0-T1 fault-injection rollback); `TestSendPendingLegalImportanceWritesAuditEvent` + `TestSendPendingNonLegalImportanceWritesNoAuditEvent` PASS (DATA-08 W0-T2 audit write via migration 00011 grant) | `logs/t001-db-gated-attachment-notify.log` | n/a (text log) | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |
| EV-W00-E01-S003-02 | S3-gated test-execution log + TOTP determinism test log + CI-configuration inspection note | W00-E01-S003 / W00-E01-S003-T002 | AC-W00-E01-S003-02 | `WOWAPI_REQUIRE_S3=1 WOWAPI_REQUIRE_DB=1 DATABASE_URL=… S3_TEST_ENDPOINT=localhost:9000 go test github.com/qatoolist/wowapi/adapters/storage/s3 -count=1 -v`; `TZ=UTC go test github.com/qatoolist/wowapi/kernel/mfa -count=5 -v`; `TZ=America/Los_Angeles go test github.com/qatoolist/wowapi/kernel/mfa -count=5 -v`; file inspection of `.github/workflows/ci.yml`, `deployments/compose.yaml`, `Makefile` | 0a31186cada5c275a588c74081cf977adf346e61 | main | Local host vs compose MinIO (localhost:9000) + Postgres; WOWAPI_REQUIRE_S3=1 so an unreachable store fails, never skips; concurrent load present | go1.26.5 darwin/arm64; minio/minio:latest; Docker 29.4.0; Compose 5.3.1 | 2026-07-13 12:10 +0530 | **pass** — S3 suite exit 0, exactly 20 top-level tests PASS (count unchanged from source material), 0 SKIP; TOTP: both TZ runs exit 0, 49 distinct top-level tests × 5 iterations each, 245 PASS / 0 FAIL / 0 SKIP per TZ (deterministic); CI inspection: SD-01 + SD-02 + REL-04 T1/T2/T3 all CONFIRMED, no drift | `logs/t002-s3-gated-suite.log`; `logs/t002-totp-tz-utc.log`; `logs/t002-totp-tz-la.log`; `logs/t002-ci-inspection-note.md` | n/a (text logs/note) | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |
| EV-W00-E01-S003-03 | Verify-outcome re-pin note with evidence pointers | W00-E01-S003 / W00-E01-S003-T003 | AC-W00-E01-S003-03 | Re-inspection of MATRIX's own citations (fable5-closure-depth-matrix-2026-07-11.md CS bodies) at HEAD: kernel/config/{load.go,config.go,fingerprint.go}, kernel/i18n/{embed.go,catalog.go}, kernel/httpclient/{client.go,guard.go}; corroborated by `go test github.com/qatoolist/wowapi/kernel/config github.com/qatoolist/wowapi/kernel/i18n github.com/qatoolist/wowapi/kernel/httpclient -count=1 -v` | 0a31186cada5c275a588c74081cf977adf346e61 | main | Local host; inspection is environment-independent; package tests need no DB/S3; concurrent load present | go1.26.5 darwin/arm64 | 2026-07-13 12:14 +0530 | **pass / still holds** — CS-03, CS-19, CS-24 each re-confirmed against MATRIX's original citation basis (majority of line citations verbatim; three single-line drifts with identical code, recorded in the note); package tests exit 0, 199 PASS / 0 FAIL / 0 SKIP; no regression to escalate | `logs/t003-cs-repin-note.md`; `logs/t003-cs-repin-package-tests.log` | n/a (text note/log) | W00ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 | none |

## Notes

- Per `evidence-policy.md`'s failed-evidence preservation rule: no run failed, so no `failed`
  record exists; nothing was retried-until-green (every command above was executed exactly once,
  first result recorded).
- No checksum is recorded — all evidence files are text execution logs or inspection notes, not
  binary artifacts ("checksum where appropriate").
- Reviewer field: acceptance is the conductor's review gate; this story does not self-assign a
  reviewer.
