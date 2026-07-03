# Phase 11 — Command Log

Exact commands with exit status; summarized output where long. Date: 2026-07-04.

| # | Command | Exit | Output summary |
|---|---|---|---|
| 1 | `go test ./kernel/httpx/ -run Health\|Readiness\|Liveness` | 0 | liveness always 200 (never runs checks); readiness 200 all-pass / 503 on a failing check + reports config_fingerprint |
| 2 | `go test ./kernel/config/ -run Shared\|Drift` | 0 | shared fingerprint ignores HTTP/Log; changes with DB/schema; CheckSharedDrift catches divergence, empty-expected disables |
| 3 | `go test ./app/ -run Readiness` | 0 | app.Readiness wires module `ctx.Health` + framework checks + fingerprint into /readyz |
| 4 | `go mod tidy` (prometheus client dep) | 0 | `github.com/prometheus/client_golang` + transitive deps added; go.sum complete |
| 5 | `go test ./kernel/observability/ ./adapters/metrics/...` | 0 | Metrics port + NoOp; RED middleware records route/method/status/dur/bytes; AccessLog canonical line; Prometheus adapter + /metrics handler (12 tests) |
| 6 | `sh scripts/lint_boundaries.sh` | 0 | OK — kernel/observability→kernel/httpx internal import (no cycle); prometheus dep only in adapters/ |
| 7 | `make bench-budget` | 0 | 24 hot-path benchmarks within budget; config field read 0.3 ns/op 0 allocs (criterion #17) |
| 8 | `make test-security` | 0 | curated gate: RLS/privilege-escalation, deny-by-default authz, secret redaction (logs/CLI/dumps), unsafe-config prod rejection, DSN non-echo, env-mismatch |
| 9 | `go test ./kernel/config/ -run Unsafe\|Prod\|OffSwitch` | 0 | per-knob unsafe-config matrix (every prod Validate() rejection); core guarantees have no off-switch |
| 10 | `make ci` (host, with bench-budget added to the ci target) | 0 | vet, boundary lint, unit, race, **bench-budget**, build green |
| 11 | `make ci-container` (1st attempts) | FAIL→0 | surfaced the long-standing testkit flake (`FATAL: role "app_rt" is not permitted to log in` / `tuple concurrently updated`) under parallel packages |
| 12 | root-caused + fixed the flake: migration 00001 no longer resets LOGIN→NOLOGIN on an EXISTING role (it fought testkit's LOGIN grant, flipping the cluster-global role mid-run) + a benign-`tuple concurrently updated` EXCEPTION handler; testkit `alterRoleWithRetry` + pool-connect `pingWithRoleRetry` on SQLSTATE 28000 | 0 | `make migrate` clean |
| 13 | `make ci-container` ×2 (post-fix) | 0, 0 | **two consecutive green parallel runs** — the flake that intermittently hit Phases 8–11 is resolved at the root |
| 14 | `make ci` (host, post-fix) | 0 | still green |
