---
id: W00-E01-DEPS
type: epic-dependencies
epic: W00-E01
wave: W00
status: planned
created_at: 2026-07-12
updated_at: 2026-07-12
---

# W00-E01 — Dependencies

## Upstream (this epic depends on)

None. W00-E01 has no dependency on W00-E02 or on any other wave — see `../../dependencies.md`
(wave-level): "W00 does not itself depend on any later wave's output." Within W00, E01 and E02 are
independent workstreams (re-verification vs. baseline-capture/ADR-ification) and can proceed in
parallel.

## Downstream (items that depend on this epic)

See `../../dependencies.md` (wave-level dependency table) for the full downstream list. The rows
specific to W00-E01 (as opposed to W00-E02):

| Downstream item | Depends on | Why |
|---|---|---|
| W01's AR-04/AR-06 remainder tasks | W00-E01-S001 | Re-verification confirms AR-04 T1/AR-06 T1's current state before T2-T5/T2-T3 build on it |
| W03's SEC-02 T4/T5 (ratification) | W00-E01-S001 | Confirms SEC-02's Wave-0 fail-closed fix (T1-T3) is genuinely intact before layering ratification on top |
| W07's PERF-02..06 remainder | W00-E01-S002 | Confirms PERF-01/PERF-06's fixes and the #25-recalibrated budgets are the correct "before" baseline for W07's relative-comparison work |
| W02/W04's DATA-08 W6 tasks | W00-E01-S003 | Confirms the DATA-08 W0 slice (attachment outbox propagation, legal-delivery audit) is intact before W6 widens the hash contract over it |
| W03's SEC-01/DATA-07 | W00-E01-S003 | REL-04 T1-T4's S3/TOTP wiring underpins the parallel-CI pipeline state later waves' test infrastructure assumes is stable |

## Internal (cross-story) dependencies within this epic

None load-bearing. S001, S002, and S003 target disjoint packages and disjoint test commands:

| Story | Packages/files touched | Test commands |
|---|---|---|
| S001 | `kernel/workflow/`, `app/`, `kernel/kernel.go`, `kernel/authz/` | `go test ./kernel/workflow/... -race`, `go test ./app/... -run Boot`, `go test ./kernel/... -race` (kernel_rules_test.go, authz caching) |
| S002 | `kernel/httpx/ratelimit.go`, `internal/tools/benchbudget/` | `go test ./kernel/httpx/... -race`, `make bench-budget`, `go test ./internal/tools/benchbudget/...` |
| S003 | `kernel/attachment/`, `kernel/notify/`, `.github/workflows/ci.yml`, `deployments/compose.yaml`, `Makefile` | `go test ./kernel/attachment/... ./kernel/notify/...`, `make ci-container` / S3-gated suite |

No story's task reads or depends on another story's output artifact or evidence. They may execute in
any order, or fully in parallel, within this epic.

## External dependencies

- Postgres via `testkit` (`make ci-container` or local `docker compose`) — required for S001's
  `kernel/kernel_rules_test.go`/`kernel/authz/caching_internal_test.go` if those are DB-backed, and
  for S003's `kernel/attachment`/`kernel/notify` DB-gated tests.
- MinIO (S3-gated tests) — required for S003's REL-04 T1-T4 re-verification (20 S3-gated tests named
  in the task-issuer content).
- No external dependency for S002 (rate-limiter and benchbudget tests are pure in-process/subprocess,
  per the source content — `ratelimit_test.go`, `bench_test.go`, `coverage_test.go` subprocess
  exit-code assertions).

## Tooling dependencies

- `go test -race` (Go toolchain, version per `go.mod`).
- `make bench-budget` / `internal/tools/benchbudget/main.go`.
- `docker compose` (for `make ci-container`).

## Decision dependencies

None. This epic consumes no unresolved decision from `../../dependencies.md`'s D-01..D-09 table —
those ADRs unblock later-wave *implementation* stories, not this epic's re-verification work.
