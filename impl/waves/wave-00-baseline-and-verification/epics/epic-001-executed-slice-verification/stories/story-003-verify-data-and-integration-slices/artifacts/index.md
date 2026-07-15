---
id: W00-E01-S003-ARTIFACTS-INDEX
type: artifacts-index
parent_story: W00-E01-S003
status: complete
derived: false
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W00-E01-S003 — Artifacts index

Per `governance/artifact-policy.md` §9.2 required fields. All five artifacts were produced by story
execution on 2026-07-13 at commit `0a31186cada5c275a588c74081cf977adf346e61`. They live under
`../evidence/logs/` (evidence-log artifacts co-located with their evidence records; directory
created on first real content per `naming-conventions.md` Adaptation 2).

| Artifact ID | Title | Type | Lifecycle stage | Description | Source requirement | Producing task | Repository path / storage location | Version | Checksum | Status | Reviewer | Retention requirement |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| ART-W00-E01-S003-001 | Attachment/notify DB-gated test log | Test-execution log | pre-implementation (baseline re-verification evidence) | Output of `go test ./kernel/attachment/... ./kernel/notify/... -count=1 -v` against compose testkit Postgres (WOWAPI_REQUIRE_DB=1), confirming DATA-08 W0-T1/W0-T2: 66 PASS / 0 FAIL / 0 SKIP, exit 0 | DATA-08 | W00-E01-S003-T001 | `../evidence/logs/t001-db-gated-attachment-notify.log` | 0a31186 | n/a | produced | unassigned | Retain per `evidence-policy.md` failed-evidence preservation rule; not superseded until re-run at a later commit |
| ART-W00-E01-S003-002 | S3-gated test suite log (20 tests) | Test-execution log | pre-implementation | Output of the 20 S3-gated tests run with `WOWAPI_REQUIRE_S3=1` against MinIO: exactly 20 top-level tests, all PASS, 0 SKIP, exit 0, confirming REL-04 T1-T3 | REL-04 | W00-E01-S003-T002 | `../evidence/logs/t002-s3-gated-suite.log` | 0a31186 | n/a | produced | unassigned | Retain per `evidence-policy.md` |
| ART-W00-E01-S003-003 | TOTP determinism test log (2 clock/TZ settings) | Test-execution log | pre-implementation | Output of the `kernel/mfa` suite at `TZ=UTC` and `TZ=America/Los_Angeles`, `-count=5` each: 49 distinct tests × 5 iterations, 245 PASS / 0 FAIL / 0 SKIP per TZ, both exit 0, confirming REL-04 T4 | REL-04 | W00-E01-S003-T002 | `../evidence/logs/t002-totp-tz-utc.log`; `../evidence/logs/t002-totp-tz-la.log` | 0a31186 | n/a | produced | unassigned | Retain per `evidence-policy.md` |
| ART-W00-E01-S003-004 | ci.yml SD-01/SD-02 state confirmation note | Inspection note | pre-implementation | Written confirmation with file:line citations that `.github/workflows/ci.yml` reflects the 3-leg parallelized gate, GHA-cached toolbox image, docs-only skip (SD-01) and bench path-scoping/nightly/merge_group support (SD-02), plus REL-04 T1/T2/T3 wiring — all CONFIRMED, no drift | SD-01, SD-02 | W00-E01-S003-T002 | `../evidence/logs/t002-ci-inspection-note.md` | 0a31186 | n/a | produced | unassigned | Retain per `evidence-policy.md` |
| ART-W00-E01-S003-005 | CS-03/CS-19/CS-24 verify-outcome re-pin note | Inspection/re-pin note | pre-implementation | Written confirmation, with evidence pointers to MATRIX's own citation basis, that CS-03 (config fail-closed + fingerprint), CS-19 (i18n freeze + key-echo fallback), and CS-24 (SSRF dial-time guard) still hold at `0a31186`; corroborating package-test log `../evidence/logs/t003-cs-repin-package-tests.log` | CS-03, CS-19, CS-24 | W00-E01-S003-T003 | `../evidence/logs/t003-cs-repin-note.md` | 0a31186 | n/a | produced | unassigned | Retain per `evidence-policy.md`; treat any regression as a new finding, not a silent update |

## Notes

No artifact in this story requires a checksum ("where appropriate" per `artifact-policy.md` §9.2) —
all are test-execution logs or inspection notes, not generated/binary artifacts. No large generated
artifact is expected; the no-duplication rule (mandate §9.3) does not apply here since nothing large
enough to warrant authoritative-path-only registration is produced by this story.
