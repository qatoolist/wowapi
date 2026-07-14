---
id: VER-W00-E01-S001
type: verification-record
parent_story: W00-E01-S001
status: executed
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W00-E01-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story. (Rows below were drafted
before execution; the AC-04 row was appended when task-004 was added to the story. Actual
results are in the post-execution record.)

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W00-E01-S001-01 | Re-run `go test ./kernel/workflow/... -race`; inspect for exit code 0 and no race warnings; confirm the `NewRuntime` nil-`ev` panic test and the `Override` fail-closed (unconditional authz check) test are present among, and passing within, the suite (`runtime_extra_test.go`, `runtime_lifecycle_test.go`, `runtime_test.go`, `testkit/workflowsim_cov_test.go`) | Local or CI Go toolchain per `go.mod`; no external DB required unless `testkit/workflowsim_cov_test.go` needs one — confirm during execution (see `plan.md` "Unresolved questions") | Exit code 0; all tests pass, including the specific nil-`ev` panic assertion and the fail-closed-`Override` assertion; no `-race` warnings | Test-execution log (race-detector log, `go test -v` output) | unassigned (framework architecture lead role) |
| AC-W00-E01-S001-02 | Re-run `go test ./app/... -run Boot` (or the specific boot-namespace-rejection test, name to be confirmed from `app/boot_extra_test.go` during execution); then re-run a full `go test ./...`; inspect both for exit code 0; confirm `app/boot.go` rejects an unknown `modules.<typo>` config namespace at boot with a deterministic named error | Local or CI Go toolchain per `go.mod`; environment requirements for the full-suite run to be confirmed during execution (some packages elsewhere in the tree may require Postgres/MinIO — out of this AC's direct scope but relevant to the full-suite green check) | Both commands exit 0; the boot-time test specifically demonstrates the named-error rejection behavior for an unknown namespace | Test-execution log (`go test -v` output) plus full-suite green-check log | unassigned (framework architecture lead role) |
| AC-W00-E01-S001-03 | Re-run `go test ./kernel/authz/... -race` and the test(s) covering `kernel_rules_test.go` (`go test ./kernel/... -run TestKernelRules -race` or equivalent, exact pattern to be confirmed during execution); inspect for exit code 0 and no race warnings; confirm via the sentinel-store-injection test that `kernel/kernel.go`'s `orgAncestry` closure (previously cited at lines 252-254) uses the composed `authzStore` instance rather than invoking a second `authz.NewStore()` | Local or CI Go toolchain per `go.mod`; DB requirement for `kernel/authz`/`kernel` test packages to be confirmed during execution | Exit code 0; all tests pass, including the sentinel-store-injection assertion demonstrating single-instance composition; no `-race` warnings | Test-execution log (race-detector log, `go test -v` output) | unassigned (framework architecture lead role) |
| AC-W00-E01-S001-04 | Run `grep -rn "RunAPI\|RunWorker\|RunMigrate" README.md docs/blueprint/` expecting zero phantom-API hits; extract the `Context` method list from `docs/blueprint/06-module-sdk.md` and diff against the live `module/module.go` interface expecting an empty diff | Local toolchain; grep + git; no DB | Zero grep hits; empty method-set diff | Doc-drift grep + diff log | unassigned (framework architecture lead role) |

## Post-execution record

Executed 2026-07-13 by worker W00E01S001. Only actually-observed results are recorded.

### Per-AC actual results

| Acceptance criterion | Actual result | Pass/fail | Evidence ID | Log file(s) |
|---|---|---|---|---|
| AC-W00-E01-S001-01 | `go test -v ./kernel/workflow/... -race` exit 0, no race warnings. `TestNewRuntimePanicsOnNilDeps` (`internal_extra_test.go:207`) PASS — `NewRuntime` panics on nil `ev` (source: `runtime.go:88-92`). `TestIntegrationOverrideAuthzGate` (`runtime_extra_test.go:440`) and `TestIntegrationOverrideFailsClosedWithoutPermission` (`runtime_extra_test.go:493`) PASS — `Override` unconditionally calls `rt.authz.Evaluate` (`runtime.go:285-306`, no nil-skip path); `TestIntegrationWorkflowOverride` (`runtime_lifecycle_test.go:171`) PASS. DB-backed tests executed (not skipped). | **pass** | EV-W00-E01-S001-01 | `evidence/tests/sec02-workflow-race.log` |
| AC-W00-E01-S001-02 | `go test -v ./app/... -run Boot` exit 0 — 16 tests PASS incl. `TestBootFailsOnUnknownConfigNamespace` (`app/boot_extra_test.go:255`), proving `app/boot.go:165-181` rejects an unknown `modules.<typo>` namespace with the deterministic named error `config: unknown module namespace(s) [...]: no registered module matches` (keys sorted). Full `go test ./...` exit 0 — 57 packages, zero FAIL. | **pass** | EV-W00-E01-S001-02 | `evidence/tests/ar04-boot-run-boot.log`; `evidence/tests/ar04-full-suite.log` |
| AC-W00-E01-S001-03 | `go test -v ./kernel/authz/... -race` exit 0, no race warnings — sentinel-store-injection test `TestCachingStoreOrgAncestorsRoutesToComposedInner` (`caching_internal_test.go:99`) PASS. `go test -v ./kernel/ -run 'TestIntegrationRulesResolverOrgAncestry' -race -count=1` (equivalent covering `kernel_rules_test.go`; planned `-run TestKernelRules` matched no tests — DEV-01) exit 0 — both org-ancestry integration tests PASS. Source confirms `kernel/kernel.go:254-256`'s `orgAncestry` closure delegates to the single composed `authzStore` (constructed once at line 230); no second `authz.NewStore()` call. | **pass** | EV-W00-E01-S001-03 | `evidence/tests/ar06-authz-race.log`; `evidence/tests/ar06-kernel-rules-race.log` |
| AC-W00-E01-S001-04 | Context method-set diff: **empty** (40/40 methods match between blueprint 06 and `module/module.go`) — T2 holds. Phantom-API grep: **7 hits** in `docs/blueprint/` (04:15,37-39; 06:207; 10:94; 12:171) instead of the expected zero. README.md and blueprint 11 (the files the executed T1 fix changed): zero hits; no such function exists in Go source; `git grep` at fix commit `345e4ce` shows the identical 7-hit set → no drift since the fix; remaining hits are future-state prose owned by AR-05 T5 (`W06-E04-S002`). See `deviations.md` DEV-02. | **fail** (as literally worded; executed T1/T2 slice itself verified intact) (adjudicated pass-on-executed-scope, conductor 2026-07-13) | EV-W00-E01-S001-04 (status `failed`, preserved) | `evidence/tests/ar05-doc-drift.log` |

### Execution date

2026-07-13 (12:13–12:21 local; full suite finished 12:17:33).

### Commit or revision

`0a31186cada5c275a588c74081cf977adf346e61` (branch `main`) — all four ACs executed against this
single pinned SHA.

### Environment

Local workstation, macOS 26.5.2 (Darwin 25.5.0), arm64; go1.26.5 darwin/arm64; Postgres + MinIO
via `make up` compose (`wowapi-postgres-1`/`wowapi-minio-1` healthy;
`DATABASE_URL=postgres://wowapi:***@localhost:5432/wowapi?sslmode=disable`); concurrent load
present (sibling W00 workers running test suites) — none of this story's checks is
timing-sensitive.

### Reviewer

unassigned — conductor review pending. Worker self-review only; per the story's Definition of
Done the independent-review checklist and acceptance remain the conductor's gate.

### Findings

1. SEC-02 T1-T3, AR-04 T1, and AR-06 T1 executed slices are all intact at
   `0a31186` — no regression (RISK-W00-001 did not materialize for these three).
2. AR-05: the executed T1/T2 slice is intact (README + blueprint 11 clean; Context diff empty;
   hit set unchanged since `345e4ce`), but AC-04's grep clause as worded fails — 7 pre-existing
   future-state `RunAPI/RunWorker/RunMigrate` references remain in blueprint 04/06/10/12,
   awaiting AR-05 T5. Conductor adjudication required (DEV-02).
3. Test infrastructure was available (RISK-W00-002 did not materialize); DB-backed tests
   executed rather than skipped.
4. Minor traceability corrections: nil-deps panic test lives in
   `kernel/workflow/internal_extra_test.go`; `testkit/workflowsim_cov_test.go` is covered by the
   full-suite run, not the `./kernel/workflow/...` pattern (DEV-03); `orgAncestry` closure now at
   `kernel/kernel.go:254-256`.

### Retest status

First execution under this programme. No retest performed; the AC-04 `failed` record awaits
adjudication — any later re-run must be a new `retested` record referencing EV-W00-E01-S001-04.

### Final conclusion

3 of 4 acceptance criteria **pass** with pinned, mandate-§10-conformant evidence; AC-04 is an
honest **fail as worded** with the underlying executed slice verified intact. Story moves to
`ready-for-review`; acceptance is blocked on the conductor's AC-04 adjudication per
`deviations.md` DEV-02.
