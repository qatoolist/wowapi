---
id: W06-E01-S002-EVIDENCE-INDEX
type: evidence-index
parent_story: W06-E01-S002
status: current
derived: false
created_at: 2026-07-12
updated_at: 2026-07-14
---

# W06-E01-S002 — Evidence index

Per mandate §10. Failed and partial-state evidence is retained. Current passing records supersede or
resolve it explicitly; no earlier failure was deleted.

## Common execution context

- **Story:** W06-E01-S002.
- **Code revision:** worktree snapshot based on commit
  `733ef3e930cbb3f89f5bbc53d8f562c60e426513`; authoritative implementation/configuration files are
  content-pinned in `artifacts/index.md`.
- **Branch or tag:** `main` worktree; the N-1 replay additionally installs tagged `v1.1.0` and upgrades
  to locally packaged candidate `v1.2.0-w06e01s002.11`.
- **Execution environment:** Darwin arm64; real Docker Compose Postgres, MinIO, Mailpit, and Jaeger;
  `DATABASE_URL` set; `WOWAPI_REQUIRE_DB=1`; `WOWAPI_REQUIRE_S3=1`.
- **Relevant tool versions:** Go `go1.26.5 darwin/arm64`; Docker Compose `5.3.1`; Python `3.14.2`;
  actionlint `1.7.12`.

## EV-W06-E01-S002-001

- **Evidence type:** installation-log evidence.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T001.
- **Acceptance criteria proven:** AC-W06-E01-S002-01 at the earlier partial snapshot.
- **Execution command:** focused `TestGoldenConsumerInstalledBinaryTwoModules` invocation recorded in the file.
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T17:00:02Z.
- **Result:** PASS for the earlier installed-binary/two-CRUD slice.
- **File or URI:** `DX-04/t1-installed-two-module.log`.
- **Checksum:** SHA-256 `f76b3d403629b96c4d2ffa776593436c24b4dc4dbb8d26bca8a165473925ca25`.
- **Reviewer:** W06E01Impl.
- **Superseded evidence:** superseded by EV-W06-E01-S002-007 and EV-W06-E01-S002-010.
- **Status:** `superseded`.

## EV-W06-E01-S002-002

- **Evidence type:** failed subsystem-coverage report.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T002.
- **Acceptance criteria proven:** none; this record failed AC-W06-E01-S002-02 and exposed the original gap.
- **Execution command:** `go run ./cmd/wowapi gen rule`.
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T16:39:25Z.
- **Result:** FAIL: the then-current command surface reported `unknown subcommand`; later implementation resolved it.
- **File or URI:** `DX-04/t2-generator-surface-blocker.log`.
- **Checksum:** SHA-256 `d84e119bd43d23858d41d4eae93a0d70b49a3ebae823fcea96775d8575e5a26c`.
- **Reviewer:** W06E01Impl.
- **Superseded evidence:** underlying issue resolved by EV-W06-E01-S002-007 and EV-W06-E01-S002-010; this failed record remains.
- **Status:** `failed`.

## EV-W06-E01-S002-003

- **Evidence type:** real-infrastructure integration-test report.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T003.
- **Acceptance criteria proven:** AC-W06-E01-S002-03.
- **Execution command:** `make golden-consumer` (includes `TestGoldenConsumerRealInfrastructure`).
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T20:33:12Z.
- **Result:** PASS: generated API/worker booted; authenticated CRUD, tenant RLS, outbox dispatch, worker stop/restart recovery, and required MinIO/Mailpit/Jaeger service checks passed.
- **File or URI:** `DX-04/t1-t4-golden-consumer-retest2-2026-07-14.log`.
- **Checksum:** SHA-256 `20a491665dc4ca45f8e862c1f39eb8c4505eb5c68ad8941190d3c05274989ce2`.
- **Reviewer:** W06-E01-S002-Verify.
- **Superseded evidence:** replaces the former not-produced placeholder for EV-003.
- **Status:** `retested`.

## EV-W06-E01-S002-004

- **Evidence type:** two-pass compatibility integration-test report.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T004.
- **Acceptance criteria proven:** AC-W06-E01-S002-04.
- **Execution command:** `make golden-consumer` (includes `TestGoldenConsumerUpgradeReplay`).
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T20:33:12Z.
- **Result:** PASS: generated/build/boot contracts passed at tagged `v1.1.0`, then dependency and generated scaffold were upgraded to local candidate `v1.2.0-w06e01s002.11`; build/boot and real-infrastructure contracts passed again.
- **File or URI:** `DX-04/t1-t4-golden-consumer-retest2-2026-07-14.log`.
- **Checksum:** SHA-256 `20a491665dc4ca45f8e862c1f39eb8c4505eb5c68ad8941190d3c05274989ce2`.
- **Reviewer:** W06-E01-S002-Verify.
- **Superseded evidence:** supersedes the preflight-only record at `DX-04/t4-released-version-preflight.log`.
- **Status:** `retested`.

## EV-W06-E01-S002-005

- **Evidence type:** CI gate configuration and failure-injection report.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T005.
- **Acceptance criteria proven:** AC-W06-E01-S002-05.
- **Execution command:** `make actionlint`; release-gate manifest validation; focused `TestGoldenConsumerFailingFixture`; manifest/workflow assertions.
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T20:33:48Z.
- **Result:** PASS: workflow syntax and manifest schema are valid; the golden-consumer entry is required from Wave 4, requires services, invokes `make golden-consumer`, the exact-SHA runner provisions Jaeger, and an incomplete fixture is rejected.
- **File or URI:** `DX-04/t5-ci-gate-retest-2026-07-14.log`.
- **Checksum:** SHA-256 `8fa4f3196bff4da2a464c5c428e1905cb24ff1718122e9d6c37a5ea427e470c9`.
- **Reviewer:** W06-E01-S002-Verify.
- **Superseded evidence:** supersedes EV-W06-E01-S002-011.
- **Status:** `retested`.

## EV-W06-E01-S002-006

- **Evidence type:** independent partial-state document/code review.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T006.
- **Acceptance criteria proven:** none; it reviewed the earlier truthful blocked disposition.
- **Execution command:** not applicable (review-only).
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T17:48:52Z.
- **Result:** the partial-state disposition was accurate, but the story remained incomplete.
- **File or URI:** `DX-04/t6-independent-review.md`.
- **Checksum:** SHA-256 `f3f62cb6b11d86f5e3dd2d76218a27e2048163f43be319fad78f7d1057c539b9`.
- **Reviewer:** W06-E01-E04-Execution.W06E01ReviewR.
- **Superseded evidence:** to be superseded by the final independent acceptance review.
- **Status:** `superseded`.

## EV-W06-E01-S002-007

- **Evidence type:** installed-binary and subsystem-coverage retest.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T001 and W06-E01-S002-T002.
- **Acceptance criteria proven:** AC-W06-E01-S002-01 and AC-W06-E01-S002-02.
- **Execution command:** `make golden-consumer`.
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T20:33:12Z.
- **Result:** PASS: a versioned `go install` binary generated two automatically wired modules and all eight required subsystem types without a checkout replace or manual edits; artifact assertions, build, and boot test passed.
- **File or URI:** `DX-04/t1-t4-golden-consumer-retest2-2026-07-14.log`.
- **Checksum:** SHA-256 `20a491665dc4ca45f8e862c1f39eb8c4505eb5c68ad8941190d3c05274989ce2`.
- **Reviewer:** W06-E01-S002-Verify.
- **Superseded evidence:** supersedes EV-W06-E01-S002-001 and resolves the gap exposed by EV-W06-E01-S002-002.
- **Status:** `resolved`.

## EV-W06-E01-S002-008

- **Evidence type:** failed integration execution log.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T001 through T004.
- **Acceptance criteria proven:** none.
- **Execution command:** `make golden-consumer`.
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T20:31:25Z.
- **Result:** FAIL before fixture execution because the eval runtime's local file proxy lacked `github.com/prometheus/procfs@v0.21.1`.
- **File or URI:** `DX-04/t1-t4-golden-consumer-2026-07-14.log`.
- **Checksum:** SHA-256 `0c029808ab01d14689f8f6cfde19ee2542f294237faa431a20082c6c3374923f`.
- **Reviewer:** W06-Closure-Finalization.
- **Superseded evidence:** superseded by EV-W06-E01-S002-010.
- **Status:** `failed`.

## EV-W06-E01-S002-009

- **Evidence type:** failed integration retest log.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T001 through T004.
- **Acceptance criteria proven:** none.
- **Execution command:** host-shell cache seed followed by `make golden-consumer`.
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T20:31:48Z.
- **Result:** FAIL for the same missing module because the host shell and eval runner used different `GOMODCACHE` roots.
- **File or URI:** `DX-04/t1-t4-golden-consumer-retest-2026-07-14.log`.
- **Checksum:** SHA-256 `858a3ec071285df46a28b393a56c6e776e0423884bced5a0df926c6d11607a32`.
- **Reviewer:** W06-Closure-Finalization.
- **Superseded evidence:** superseded by EV-W06-E01-S002-010.
- **Status:** `failed`.

## EV-W06-E01-S002-010

- **Evidence type:** full golden-consumer retest.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T001 through T004.
- **Acceptance criteria proven:** AC-W06-E01-S002-01 through AC-W06-E01-S002-04.
- **Execution command:** seed the eval runner's local module cache, then `make golden-consumer`.
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T20:33:12Z.
- **Result:** PASS; all selected golden-consumer and RLS-census tests passed.
- **File or URI:** `DX-04/t1-t4-golden-consumer-retest2-2026-07-14.log`.
- **Checksum:** SHA-256 `20a491665dc4ca45f8e862c1f39eb8c4505eb5c68ad8941190d3c05274989ce2`.
- **Reviewer:** W06-E01-S002-Verify.
- **Superseded evidence:** supersedes EV-W06-E01-S002-008 and EV-W06-E01-S002-009.
- **Status:** `retested`.

## EV-W06-E01-S002-011

- **Evidence type:** failed CI-evidence assertion log.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T005.
- **Acceptance criteria proven:** none; all invoked commands passed, but the recorder's YAML-text assertion incorrectly expected YAML syntax in a JSON-formatted manifest.
- **Execution command:** `make actionlint`; release-gate manifest validation; focused failure-injection test; initial static assertions.
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T20:33:32Z.
- **Result:** FAIL in the evidence recorder's assertion only; product/workflow checks passed.
- **File or URI:** `DX-04/t5-ci-gate-2026-07-14.log`.
- **Checksum:** SHA-256 `bbd4e628149074643ba9b1e64a9ad5f97dc520086ffe2c159838a6ef9e24e988`.
- **Reviewer:** W06-Closure-Finalization.
- **Superseded evidence:** superseded by EV-W06-E01-S002-012.
- **Status:** `failed`.

## EV-W06-E01-S002-012

- **Evidence type:** CI-evidence assertion retest.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T005.
- **Acceptance criteria proven:** AC-W06-E01-S002-05.
- **Execution command:** parse `ci/release-gates.yaml` as JSON and assert the golden-consumer gate contract, using the immediately preceding successful actionlint, schema-validation, and failure-injection command results.
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T20:33:48Z.
- **Result:** PASS.
- **File or URI:** `DX-04/t5-ci-gate-retest-2026-07-14.log`.
- **Checksum:** SHA-256 `8fa4f3196bff4da2a464c5c428e1905cb24ff1718122e9d6c37a5ea427e470c9`.
- **Reviewer:** W06-E01-S002-Verify.
- **Superseded evidence:** supersedes EV-W06-E01-S002-011.
- **Status:** `retested`.

## EV-W06-E01-S002-013

- **Evidence type:** failed independent acceptance review.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T006.
- **Acceptance criteria proven:** technical commands for AC-01 through AC-05 passed, but the review did not authorize acceptance because evidence/artifact/register closure was incomplete.
- **Execution command:** independent `make golden-consumer`, `go test ./internal/cli -count=1`, `make ci`, `make actionlint`, release-gate manifest validation, and governance audit.
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T20:53:43Z.
- **Result:** FAIL: missing EV-013 record, incomplete ART-002 implementation pin, and stale programme registers.
- **File or URI:** `DX-04/t6-final-review-attempt-2026-07-14.md`; raw command output `artifact://2375`, `artifact://2377`, `artifact://2379`.
- **Checksum:** SHA-256 `8cc2139dccd5c9387f781964d02972a1815309ff38405d9ee0e5f2fec317e982`.
- **Reviewer:** W06-E01-S002-Verify.
- **Superseded evidence:** superseded by the passing fresh review EV-W06-E01-S002-014; this failed record remains.
- **Status:** `failed`.

## EV-W06-E01-S002-014

- **Evidence type:** final independent acceptance review.
- **Story and task:** W06-E01-S002 / W06-E01-S002-T006.
- **Acceptance criteria proven:** independently confirms AC-W06-E01-S002-01 through AC-W06-E01-S002-05 and the complete definition-of-done/governance transition.
- **Execution command:** independent `make golden-consumer`, `go test ./internal/cli -count=1`, `make ci`, `make actionlint`, release-gate manifest validation, aggregate-hash recomputation, and governance audit.
- **Code revision / branch or tag / environment / tools:** common context above.
- **Date and time:** 2026-07-13T20:58:28Z.
- **Result:** PASS: commands passed; Jaeger and upgrade prose are correct; EV-013 is preserved; programme registers align; ART-002/003/005 aggregate hashes recompute exactly; no open technical/evidence discrepancy.
- **File or URI:** `DX-04/t6-final-independent-review-2026-07-14.md`; raw command output `artifact://2375`, `artifact://2377`, `artifact://2379`.
- **Checksum:** SHA-256 `77a7a1dd5628e6f5266e001a4b3955b1867c8600e072c19f841f2c2c052cfe08`.
- **Reviewer:** W06-E01-S002-Verify.
- **Superseded evidence:** supersedes failed review EV-W06-E01-S002-013.
- **Status:** `retested`.
