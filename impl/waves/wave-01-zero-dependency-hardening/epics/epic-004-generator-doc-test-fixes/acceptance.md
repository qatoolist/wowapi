---
id: W01-E04-ACCEPTANCE
type: epic-acceptance
epic: W01-E04
wave: W01
status: accepted
created_at: 2026-07-12
updated_at: 2026-07-13
---

# W01-E04 — Epic-level acceptance

Reproduces and elaborates `epic.md`'s "Epic acceptance criteria" section as a standalone,
independently-referenceable acceptance record, consistent with the wave-level `../../acceptance.md`
pattern (AC-W01-08, AC-W01-09, AC-W01-10 there map onto this epic).

## AC-W01-E04-01 — Version-resolution flags fail closed correctly

`wowapi init --framework-version vX.Y.Z` verifies the version via `go list -m` before any file write;
an unresolvable version fails before writes with an exact remediation command.
`wowapi init --local-framework /absolute/path` emits a `replace` directive and a visible dev-mode
warning; a non-absolute or nonexistent path is rejected. With neither flag, a clean/reachable-commit
`init` derives an exact VCS pseudo-version; a dirty/unreachable-commit `init` fails closed with
remediation. The `v0.0.0` fallback path is deleted from the codebase, not merely bypassed. Traces to
W01-E04-S001, task T001.

## AC-W01-E04-02 — Isolated-temp-dir E2E harness proves boot-safety

The DX-01 T5 harness performs a real `init` → `go mod download` → `go build` → contract/smoke-test →
success cycle in an isolated temporary directory, for both the released-CLI and source-built-CLI
paths. The harness is built as a reusable primitive (not story-specific glue) and is reused, not
reimplemented, by AC-W01-E04-03's generator-output-boots test. Traces to W01-E04-S001, task T002.

## AC-W01-E04-03 — Generator emits a valid, boot-passing permission verb

`internal/cli/templates/crud/resource.go.tmpl` emits `RouteMeta{Permission: "{{.PermPrefix}}.deactivate"}`,
not `.delete`. `TestGenCRUDPermissionKeys` is updated to assert the corrected verb — the bug is no
longer test-locked. A generator-output-boots CI test (reusing task T002's harness) fails before this
fix (closed-verb-set rejection at `kernel/authz/registry.go:88-90`) and passes after. The kernel's
closed authorization-verb set itself is not widened. Traces to W01-E04-S001, tasks T003 and T004.

## AC-W01-E04-04 — Plan document traceability reconciled

`docs/implementation/premier-framework-implementation-plan.md`'s §6 traceability-matrix row for DX-05
is corrected to show T1/T2 as `EXECUTED`, matching §9's execution record. This criterion is satisfied
by a task file that precisely describes the required edit (the exact table row and correction) as
planning documentation — the edit itself is deferred to when the task moves from `todo` to
`in-progress`, per this epic's own scoping instruction that "Do not modify existing files" applies to
files outside `impl/waves/`. Traces to W01-E04-S002, task T001.

## AC-W01-E04-05 — DX-05 residual items planned or explicitly deferred

DX-05 T3's blueprint-11 CLI examples each have a recorded implement-or-delete decision against
`internal/cli/cli.go`'s real commands/flags. DX-05 T4's version-compatibility gate on mutating
generator commands is planned with its dependency on S001's DX-01 version-verification plumbing
stated explicitly. DX-05 T5 (public API/config/event compatibility gates, shared with REL-03) is
recorded as deferred-to-W06 with the cross-reference in the story's out-of-scope section, not silently
dropped. Traces to W01-E04-S002, task T002.

## AC-W01-E04-06 — FBL-03 register reconciliation planned as PROD-level coordination

FBL-03's target register plan states precisely: PF-2 is closeable contingent on S001's DX-02 task
landing (with the cross-story dependency recorded); PF-6/RFF-001 are corrected to already-resolved
status per REVIEW Answer 18. Because the register is a wowsociety-repository artifact, this criterion
is satisfied by a documented, precise PROD-level coordination recommendation — following the pattern
in `requirement-inventory.md` §D's product-items table — not a direct edit to a file this repository
does not own. Traces to W01-E04-S002, task T003.

## AC-W01-E04-07 — e2e flake reproduction attempted and diagnosis recorded honestly

T-TEST-01's reproduction is attempted under `-count`+parallel full-suite runs (task T001); a
determination is recorded of whether `internal/e2e` uses `testkit.NewDB` cloning or its own DB wiring.
The resulting diagnosis — confirmed cause, or an honest "not reproducible, downgraded to monitoring"
— is recorded without re-asserting the withdrawn "shared-DB concurrency" cause. Task T002's fix is
implemented strictly according to what T001's investigation determines, recorded as conditional in
`plan.md` and `verification.md` rather than invented in advance. Traces to W01-E04-S003.

## AC-W01-E04-08 — Independent review passed

All three stories (S001, S002, S003) have passed independent review per mandate §14. S001's review
specifically confirms `TestGenCRUDPermissionKeys` was actually updated (RISK-W01-005), not merely the
template. S003's review specifically confirms no fix mechanism was pre-committed before T001's
reproduction step actually completed (RISK-W01-004 and the story's own investigation-first framing).

## Acceptance authority

Developer-experience lead (role-based per `../../wave.md`'s split — this epic sits on the
"DX-04/generator side" of the wave's owner split, distinct from the ARCH-adjacent linter/
observability/HTTP epics).

## Acceptance record — 2026-07-13

Satisfied 2026-07-13. All acceptance criteria for W01-E04 are met; independent review passed
(W01ReviewGate); accepted by conductor.
