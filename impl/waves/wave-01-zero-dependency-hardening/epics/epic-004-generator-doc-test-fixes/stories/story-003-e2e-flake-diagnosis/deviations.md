---
id: DEV-W01-E04-S003
type: deviations-record
parent_story: W01-E04-S003
status: final
created_at: 2026-07-12
updated_at: 2026-07-13
---

# Deviations — W01-E04-S003

One deviation recorded (environment adaptation, not a protocol change):

## DEV-001 — Isolated-worktree re-execution after shared-tree contamination

- **What the plan said:** run `go test -count=N -parallel=P ./internal/e2e/` repeatedly; the
  plan did not specify *where* the tree under test lives (implicitly, the working tree).
- **What actually happened:** the first 4 primary invocations (runs 01-04, main working tree)
  produced 16 failures at the `go vet ./...` step — `adapters/tracing/otel/otel.go:58` failing
  to compile against an in-flight `observability.Span` interface change. Cause confirmed at
  failure time via `git status`: sibling wave worker W01Obs (D-08) had
  `adapters/tracing/otel/otel.go` and `kernel/observability/tracing.go` modified in the shared
  working tree. The e2e suite compiles the whole framework at run time through the scaffolded
  product's `replace` directive, so sibling edits become part of this suite's inputs.
- **Adaptation:** the identical protocol was re-executed in a detached `git worktree` pinned at
  `0a31186cada5c275a588c74081cf977adf346e61` (runs 05-08, stress, race) — 28/28 PASS. The
  worktree was removed afterwards. No git write to history; conductor still owns commits.
- **Why this is a deviation worth recording rather than silent:** the evidence-policy
  revision-pinning rule is better satisfied by the adaptation (runs against an exact SHA), but
  the plan's implicit "working tree" assumption proved unsound during a multi-worker wave; the
  contaminated runs are preserved as `failed` evidence rather than deleted (mandate §10).
- **T002 branch selection is NOT a deviation:** the monitoring-only outcome is illustrative
  branch 3 of `tasks/task-002-conditional-fix.md`, explicitly foreseen at planning time.

Per mandate §2.6: "The approved implementation plan must not be rewritten after implementation to make
it appear that the final implementation always matched the plan. Differences must be recorded in a
separate deviation record." Note this story's own special case: `plan.md`'s T002 content is explicitly
conditional, not fully specified in advance — T001 surfacing a branch not foreseen among the
illustrative options in `tasks/task-002-conditional-fix.md` is not automatically a "deviation" in the
traditional sense (the plan itself anticipated that possibility). A deviation record here would instead
apply if, e.g., T001's actual reproduction protocol departed materially from what `plan.md` describes
(different budget rationale, different methodology), or if T002 implemented something inconsistent with
what T001 actually found.

## Carry-forward addendum

Addendum (conductor, 2026-07-13): internal/e2e/e2e_test.go modified post-evidence (--local-framework flag added to integrate DX-01 fail-closed init); fresh runs TestE2EScaffoldedRepoBuild PASS (10.5s, 13.3s) + reviewer independent re-run PASS (11.4s). Diagnosis conclusions unaffected (change is harness wiring, not timing/DB behavior).
