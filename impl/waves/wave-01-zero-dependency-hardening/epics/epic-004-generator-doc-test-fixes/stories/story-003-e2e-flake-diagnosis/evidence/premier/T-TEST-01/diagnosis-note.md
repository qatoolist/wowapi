---
id: ART-W01-E04-S003-002
type: decision-record
title: E2E flake diagnosis note — bounded non-reproduction + DB-wiring determination
parent_story: W01-E04-S003
parent_task: W01-E04-S003-T001
source_requirement: T-TEST-01
status: final
commit_sha: 0a31186cada5c275a588c74081cf977adf346e61
created_at: 2026-07-13
---

# Diagnosis note — intermittent `internal/e2e` full-suite failure (T-TEST-01)

## Verdict

**Bounded non-reproduction.** 29 consecutive clean executions of `TestE2EScaffoldedRepoBuild`
(the entirety of `internal/e2e`) at pinned commit `0a31186cada5c275a588c74081cf977adf346e61`,
under and beyond the story's prescribed `-count`+parallel protocol — including a race-detector
variant and a targeted stress phase that deliberately recreated full-suite-style concurrent DB
contention. Zero failures attributable to the suite itself. The historical single failure is
**downgraded to a monitoring item** per the story's own residual-risk expectations.

The withdrawn "shared-DB concurrency" cause is **not re-asserted**. No new evidence supports it:
even under deliberate concurrent template-clone/DDL contention on the shared base database
(stress phase below), the suite passed 6/6.

## 1. Reproduction protocol executed (T-TEST-01 step 1)

Environment: darwin/arm64, go1.26.5, PostgreSQL 16.14 (`wowapi-postgres-1` compose container,
`make up` stack), MinIO healthy. DSN `postgres://wowapi:...@localhost:5432/wowapi`,
`WOWAPI_REQUIRE_DB=1` set on every run so a silent skip could not masquerade as a pass.
Full per-run logs: `logs/` (16 files, every run preserved, pass or fail).

| Phase | Runs | Flags | Tree | Executions | Result |
|---|---|---|---|---|---|
| Preflight | run-00 | `-count=1` | main working tree | 1 | PASS (11.6s) |
| Primary (contaminated) | run-01..04 | `-count=5 -parallel=4` | main working tree | 20 | 4 PASS, **16 FAIL — environment contamination, see §3; not the historical flake** |
| Primary (isolated) | run-05..08 | `-count=5 -parallel=4` | worktree pinned @0a31186 | 20 | **20/20 PASS** |
| Stress | run-stress-1..3 | e2e `-count=2` concurrent with `go test ./testkit/ ./internal/cli/` on the same base DB | worktree @0a31186 | 6 (e2e) | **6/6 PASS** (companion suites 6/6 ok) |
| Race | run-09 | `-race -count=2` | worktree @0a31186 | 2 | **2/2 PASS** |

Clean-conditions total: **29/29 PASS** (preflight + isolated primary + stress + race).

Protocol facts recorded honestly:

- `-parallel=4` is **inert** for this package: `internal/e2e` contains exactly one test function
  and it does not call `t.Parallel()`; `-count=N` iterations of the same test run serially by
  design of `go test`. The flag was passed as prescribed by the story; its inertness is a
  property of the suite, not a protocol shortcut. Cross-test parallelism was instead supplied
  by the stress phase (concurrent DB-heavy sibling packages), which is the condition the
  original observation ("full-suite run") actually involved.
- The stress phase's companion packages (`testkit`, `internal/cli`) exercise the exact shared
  resources a full-suite run contends on: the base database (testkit connects to it as admin to
  `CREATE DATABASE ... TEMPLATE` clones and ALTERs cluster-global roles), the go build cache,
  and host CPU. This recreates the original failure's condition class without running the full
  tree, which was prohibited for this wave (conductor owns the wave-level gate) and would have
  been meaningless anyway while sibling W01 workers mutate other packages (§3).

## 2. DB-wiring determination (T-TEST-01 step 2) — determined by direct code reading

**`internal/e2e` does NOT use `testkit.NewDB`.** It has its own, separate DB wiring:

- `internal/e2e/` contains a single file, `e2e_test.go`, whose only reference to testkit is a
  comment (line 16, citing the offline-skip pattern). No import, no call.
- The suite consumes the raw `DATABASE_URL` env var directly (`e2e_test.go:113`). The scaffolded
  product's `migrate` binary applies kernel migrations **directly to the base database named in
  the DSN** (`e2e_test.go:127-137`), and the product `api` binary connects to that same base
  database (`e2e_test.go:207-210`). There is no per-test clone, no template, no `t.Cleanup` drop.
- Isolation posture that follows from this: within one `go test` invocation the suite is a
  single serial test, so it never races **itself**; but its database state is shared with any
  concurrent user of the same base DSN (testkit-based packages use that DSN as the admin DSN
  for template/clone DDL). The kernel migrations it re-applies are idempotent (goose), and the
  known cluster-global role race is already handled inside `migrations/00001_bootstrap.sql:47-59`
  (the "tuple concurrently updated" exception swallow) and `testkit/db.go` `alterRoleWithRetry`.
- This determination is a standalone fact independent of the reproduction outcome, as the plan
  required. It does **not** by itself establish a failure cause: the stress phase hammered
  exactly this shared surface and produced zero failures.

## 3. The 16 contaminated failures (run-01..04) — real failures, fully explained, not the flake

Run-01 iteration 5 through run-04 failed at the `go vet ./...` step with a compile error in
`adapters/tracing/otel/otel.go:58` (`otelSpan does not implement observability.Span (missing
method SpanID)`). Cause, confirmed at the moment of failure via `git status`: sibling wave
worker **W01Obs** (D-08 observability work) had `adapters/tracing/otel/otel.go` and
`kernel/observability/tracing.go` modified in-flight in the shared working tree. The e2e suite
builds the CLI from the local tree and `replace`s the scaffolded product's wowapi dependency
with the local tree, so `go build ./...`/`go vet ./...` compile the framework **as it exists at
run time**. The investigation moved to a detached worktree pinned at `0a31186` (run-05 onward),
after which the identical protocol passed 20/20.

These 16 failures are preserved per the failed-evidence rule (mandate §10) with status `failed`,
superseded by run-05..08 (`retested` at the pinned SHA). They are **not** evidence of the
historical flake — but they are direct, demonstrated evidence of a mechanism class worth
recording (§4).

## 4. Demonstrated failure-domain insight (new, evidence-backed)

`internal/e2e`'s failure domain is the **entire repository tree state at run time**, plus the
module/build caches and network, plus the shared base database — because it rebuilds and
re-vets the whole framework through the product's `replace` directive on every execution. This
investigation *demonstrated* (16 real failures) that any transient tree inconsistency fails the
suite in a way that later in-isolation reruns cannot reproduce — exactly the signature of the
original observation (one full-suite failure, 4/4 isolated passes). Whether the original
failure was of this class is **unknowable**: its failure log was not preserved, so no cause is
assigned. That log loss is itself the actionable lesson (§6).

Candidate load-sensitivity mechanisms in the suite, listed as **unconfirmed hypotheses** for
future triage (none reproduced in 29 clean executions):

- `assert429CarriesEdgeHeaders` (`e2e_test.go:252-279`) must exhaust a 40-token/20rps bucket
  within 120 requests; a severely loaded host spacing requests >~50ms average never triggers a
  429. Request errors are silently `continue`d, consuming attempts.
- `pollHealthz` 30s budget (`e2e_test.go:230`) — api cold-start under full-suite CPU contention.
- Subprocess `go build`/`go mod tidy` steps under toolchain/network hiccups (partially guarded
  by the `isOfflineErr` skip).

## 5. Decision for T002 (T-TEST-01 step 3)

**Branch taken: monitoring-only, no code fix** (task-002 illustrative branch 3 — "cannot
reproduce after the planned budget"). Rationale: 29/29 clean executions at pinned HEAD under
the prescribed protocol, a race variant, and deliberate shared-DB stress; no confirmed root
cause exists to fix, and the story prohibits inventing one. The residual risk (one unexplained,
unreproduced historical failure) is accepted and downgraded to a programme-level monitoring
item per the story's "Residual-risk expectations".

## 6. Monitoring item (programme-level)

If `internal/e2e` fails again in any full-suite run:

1. **Preserve the complete failure log before any rerun** — the original failure is
   undiagnosable solely because its log was discarded.
2. Classify the failing step first: a tree-compilation step (`go build`/`go vet`/`tidy`)
   points at tree/cache/network state (§4's demonstrated class); a runtime step
   (migrate / healthz / 429-burst) points at DB or load sensitivity (§4's hypotheses).
3. Only then rerun in isolation; attach both logs to a new T-TEST-01 evidence record.
