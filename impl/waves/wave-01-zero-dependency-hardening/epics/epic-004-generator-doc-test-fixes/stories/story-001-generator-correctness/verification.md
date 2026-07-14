---
id: VER-W01-E04-S001
type: verification-record
parent_story: W01-E04-S001
status: complete
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Verification record — W01-E04-S001

## Planned verification procedure

Per mandate §8.8. One row per acceptance criterion for this story.

| Acceptance criterion | Verification method | Required environment | Expected result | Evidence type | Reviewer |
|---|---|---|---|---|---|
| AC-W01-E04-S001-01 | Fail-first: run `wowapi init` with an unresolvable `--framework-version`, an invalid `--local-framework` path, and a dirty/unreachable-commit default (no flags) — confirm each fails before any file write, with a remediation command; then run a clean/reachable-commit default and confirm a real VCS pseudo-version is derived | Local dev environment or CI, Go toolchain per `go.mod`, a git checkout with controllable clean/dirty/reachable/unreachable commit states | All three failure paths fail closed pre-write with remediation; the success path writes a real, non-`v0.0.0` version | functional-test report + execution log | W01ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 |
| AC-W01-E04-S001-02 | Run the isolated-temp-dir generate→build→boot→smoke harness (T002) end to end, for both the released-CLI and source-built-CLI invocation paths | Isolated temp dir, Go toolchain, network access for `go mod download` | `init` → `go mod download` → `go build` → contract/smoke tests all succeed, for both CLI paths | functional-test report (harness run log) | W01ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 |
| AC-W01-E04-S001-03 | Fail-before/pass-after run of `TestGenCRUDPermissionKeys`; direct inspection of `resource.go.tmpl:54`'s emitted permission string | Local dev environment or CI, `go test ./internal/cli/...` | Test asserts and passes against `"widgets.widget.deactivate"`, not `"widgets.widget.delete"` | unit-test report | W01ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 |
| AC-W01-E04-S001-04 | Fail-before/pass-after run of the generator-output-boots test (T004), reusing T002's harness, against the pre-T003 and post-T003 template | Isolated temp dir, Go toolchain, `kernel/authz` boot path exercised | Fails with closed-verb-set rejection at `kernel/authz/registry.go:88-90` before T003; passes after | functional-test report (fail-before/pass-after pair) | W01ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 |
| AC-W01-E04-S001-05 (added — DEV-W01-E04-S001-03 scope addition) | Fail-before/pass-after run of `TestInitScaffoldConfigValidates` (pristine scaffold, framework-only `config validate --env local`) | Local dev or CI, `go test ./internal/cli/` | Fails pre-fix with `i18n.*: unknown key`; passes post-fix with `config OK` | functional-test report (fail-before/pass-after pair) | W01ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13 |

## Post-execution record

Recorded 2026-07-13. AC-03/-04/-05 executed by W01Gen (07:26–07:45 UTC); AC-01/-02 executed by
W01GenDX01 (07:50–08:05 UTC, the DEV-W01-E04-S001-01 follow-up). All five criteria are now verified.

### Actual result

- **AC-W01-E04-S001-01**: fail-first baseline captured BOTH live defect shapes silently succeeding
  (`v0.0.0` from an unstamped `go run` CLI; the SF-7 `v1.0.1-…+dirty` stamp from a dirty-tree
  `go build` CLI — DEV-W01-E04-S001-04), with `go mod download` failing only afterwards. Post-fix:
  unresolvable `--framework-version` fails closed pre-write with the exact `go list -m -versions`
  remediation (proven through the REAL `go list` subprocess under GOPROXY=off); invalid
  `--local-framework` (relative / nonexistent / non-framework) fails closed pre-write; flag-less
  devel/dirty/unreachable builds fail closed pre-write with remediation; clean/reachable derives the go
  tool's canonical version; every failure case asserts ZERO files written; the `v0.0.0` fallback no
  longer exists in `init_cmd.go` (grep: 0 occurrences). 15 tests + 2 subtests PASS.
- **AC-W01-E04-S001-02**: the T002 harness ran `init` → `go mod tidy` → `go mod download` →
  `go build ./...` → boot-smoke to success for BOTH the source-built (`--local-framework`) and
  released (v0.1.0-ldflags-stamped, hermetic file:// module proxy) CLI paths, each step individually
  logged; the source path additionally proves the flag-less devel pipeline failure mode is now
  impossible (init fails closed before the pipeline starts). Reuse clause satisfied: T004 consumes the
  same underlying scaffold primitive (`buildRenderedProduct`, now built on `init --local-framework`).
- **AC-W01-E04-S001-03**: `TestGenCRUDPermissionKeys` PASSED pre-fix on the buggy
  `"widgets.widget.delete"` (documenting the RISK-W01-005 test-lock), and PASSED post-fix on
  `"widgets.widget.deactivate"`; `resource.go.tmpl:54` inspected — emits `deactivate`.
- **AC-W01-E04-S001-04**: `TestGenCRUDOutputBoots` FAILED pre-T003 with the verbatim
  `kernel/authz/registry.go:88-90` rejection (`permission action %q is not in the closed verb set:
  widgets.widget.delete`) and PASSED post-T003. CI-wired via `make test-unit` (no `-short`).
- **AC-W01-E04-S001-05**: `TestInitScaffoldConfigValidates` FAILED pre-fix with four
  `i18n.*: unknown key` rejections and PASSED post-fix (`config OK`).
- Full touched-package regression: `go test ./internal/cli/ -count=1` → ok (DX-02 session: 15.2s,
  13.0s; DX-01 session: 24.4s, post-T001/T002); `go test ./internal/buildinfo/` → ok.

### Pass or fail

PASS for all five acceptance criteria (AC-01 through AC-05), each with its fail-first pair.

### Evidence identifier

EV-W01-E04-S001-001 through EV-W01-E04-S001-005 (see `evidence/index.md`, `evidence/DX-01/`,
`evidence/DX-02/`).

### Execution date

2026-07-13 (07:26–08:05 UTC).

### Commit or revision

HEAD `05dce5c8a548f7dce3222637ab2c82024236a2a0`; all changes uncommitted on top (conductor commits).

### Environment

macOS Darwin 25.5.0 arm64, go1.26.5, local dev workstation; isolated `t.TempDir()` scaffolds.

### Reviewer

W01ReviewGate (independent reviewer agent); accepted by conductor 2026-07-13.

### Findings

1. Line drift vs story citations (DEV-W01-E04-S001-02, cosmetic).
2. T004's first boot attempt failed for the wrong reason (nil TxManager panic in `kernel.New`) and was
   caught by the test's failure-message discrimination exactly as task-004's risk section intended;
   fixed with a no-op stub before the genuine fail-first capture.
3. DX-01's dominant live shape on Go 1.24+ is the stamped `+dirty` pseudo-version (SF-7), not the
   story's cited `devel`→`v0.0.0` arm — both shapes captured fail-first, both now fail closed
   (DEV-W01-E04-S001-04). The former residual T005 risk (in-scaffold `config validate` delegation leg
   broken by the go.mod defect) is cleared: scaffolds created via any of the three resolution paths
   carry a resolvable framework requirement.
4. Workstation `GOPRIVATE=github.com/qatoolist/*` bypasses module proxies for the framework —
   neutralized inside the T002 harness env (task-002 Findings 1).

### Retest status

Not required — all criteria verified first-pass at the pinned revision.

### Final conclusion

All five acceptance criteria verified with preserved fail-first evidence. Story-level verification is
COMPLETE; story ready for the wave review gate and `verified` status.
