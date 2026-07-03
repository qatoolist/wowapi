# Phase 11 — Proof Bundle

Scope (phase-plan row 11, blueprint 07 §1–2/§9; AC #17/#18/#26/#27): observability wiring (metrics,
access log, health), performance budgets + gate, the security test suite, and cross-process config
fingerprint drift. Date: 2026-07-04.

## 1. Decision evidence
D-0058 (Phase 11: observability ports + Prometheus adapter, health endpoints, perf-budget gate,
security-suite gate, config shared-section drift).

## 2. Discussion evidence
- **Observability as ports + adapters:** the kernel defines a `Metrics` port + RED/access-log
  middleware + health endpoints and stays dependency-light; the Prometheus client lives ONLY in
  `adapters/metrics/prometheus`. A product wires a concrete exporter; the framework ships a NoOp so the
  default build has no metrics dependency. Full OTel span export is likewise a product adapter (the
  framework emits trace_id + provides the middleware).
- **Metric cardinality:** the RED middleware labels by the matched ROUTE PATTERN (Go 1.22+ `r.Pattern`),
  never the raw URL — bounded cardinality is a correctness/cost requirement for Prometheus.
- **Liveness vs readiness:** liveness runs NO checks (a failing dependency must not make an orchestrator
  kill a healthy process); readiness runs the checks and 503s. Both are framework-provided handlers the
  product mounts.
- **Config drift:** processes share env/schema/DB and legitimately differ on HTTP/Log; the shared
  fingerprint excludes the latter so only a genuine shared-config divergence is flagged. Secrets are
  redacted out of the fingerprint (rotating a value doesn't churn it; changing the ref does).
- **Perf budgets:** budgets live in `bench-budgets.txt` at ~10× measured values (arch-variance tolerant)
  and are enforced by a pure-Go gate wired into `make ci` — a regression fails the build. Config field
  reads at 0.3 ns/op, 0 allocs prove the hot path is reflection/lookup-free (#17/#27).

## 3. Critique/review evidence
`review-findings.md`: the security-suite agent audited the enforced guarantees (deny-by-default,
secret-reference-only, structural redaction, RLS, unsafe-config-fails-startup) and found NO gaps, with
tests for each; it flagged one dormant-mechanism maintenance note (enforceUnsafe has no tagged field
yet). Lead verified new-code correctness (health, middleware, drift, benchmarks) and integration — all
CI gates green, boundary lint clean.

## 4. Implementation evidence
Lead: `kernel/httpx/health.go` (+ test), `kernel/config/shared.go` (+ test), `app/health.go` readiness
assembler (+ test), the `bench-budget` gate added to the `ci` Makefile target. Agent A: `kernel/
observability` (Metrics port, NoOp, RED + AccessLog middleware) + `adapters/metrics/prometheus`.
Agent B: 24 hot-path benchmarks, `internal/tools/benchbudget` + `bench-budgets.txt`, `make bench`/
`bench-budget`/`test-security` targets, per-knob unsafe-config matrix + secret-redaction gap tests.

## 5. Verification evidence
`command-log.md`: health/readiness tests, config-drift tests, observability + prometheus tests, boundary
lint, `make bench-budget` (all within budget), `make test-security` (curated security gate), the
per-knob config matrix; host `make ci` (now including bench-budget) + `make ci-container` green.

## 6. Acceptance evidence
`acceptance-map.md`: all 16 Phase 11 exit criteria mapped (#17 budgets enforced in CI + hot paths
lookup-free; #18 middleware/metadata/test-enforced guardrails; #26 unsafe-config-fails + secret
redaction; #27 immutable boot-time config + shared-section drift). Carried forward: OTel span export +
pgx SQL tracer (product adapters), Grafana dashboards/alerts (ops). Graphify `extract` blocked (R11).
