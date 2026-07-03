# Phase 11 — Acceptance Map

Phase 11 exit criteria (Goal 2 Phase 11 + phase-plan row 11 + blueprint 07 §1–2/§9; AC #17/#18/#26/#27) → proof.

| # | Criterion | Proof |
|---|---|---|
| 1 | Structured JSON logs (request_id/tenant_id/actor_id/trace_id/module) + canonical per-request line | `kernel/logging` (Phase 1) + `observability.AccessLog` middleware (one structured line/request) |
| 2 | **Metrics port (RED per route, counters, gauges) + Prometheus adapter** | `kernel/observability/metrics.go` `Metrics` + NoOp; `observability.Requests` RED middleware; `adapters/metrics/prometheus` + `/metrics` handler |
| 3 | Metrics names mirror 07 §9 (outbox_pending, dispatch_lag, job depth, workflow tasks, authz denials, breaker state, delivery failures, rate-limit drops) | `SetGauge`/`IncCounter` by name; composition root records them from the relevant subsystems |
| 4 | **Liveness `/healthz` + readiness `/readyz`** | `kernel/httpx/health.go` (Liveness always 200; Readiness runs checks → 200/503 + config_fingerprint); `TestLivenessAlwaysOK`, `TestReadiness*` |
| 5 | Readiness checks: DB ping, migrations current, registries validated, config valid + fingerprint, module checks | `app.Readiness` assembles module `ctx.Health` + framework checks + fingerprint; `TestReadinessWiresModuleAndFrameworkChecks` |
| 6 | **Performance budgets defined + enforced in CI (#17)** | hot-path `Benchmark*` (authz/router/config-read/filtering/pagination) + `internal/tools/benchbudget` gate + `bench-budgets.txt`; `make bench-budget` |
| 7 | **Hot paths free of reflection/registry lookups (#17/#27)** | config value read is a struct field (benchmarked ~ns, zero-alloc); hot paths read immutable boot-time config |
| 8 | **Race gate** | `make test-race` (go test -race) in `make ci` |
| 9 | **Security guardrails enforced by middleware + route metadata + tests (#18)** | `make test-security` curated gate over RLS/authz/privilege/secret tests across the repo |
| 10 | **Unsafe production config fails startup — per-knob matrix (#26)** | `kernel/config` unsafe-config matrix test (each unsafe prod knob fails validation) |
| 11 | Core security guarantees have no disabling config key (#26) | documented: RLS enforcement, deny-by-default authz, secret-reference-only have no off-switch (asserted/noted) |
| 12 | Secrets only as references; redaction verified in logs, dumps, CLI (#26) | secret-redaction tests (slog line, `wowapi config print --redacted`, schema/doctor) |
| 13 | **Config fingerprint + shared-section drift alert (#27)** | `kernel/config/shared.go` `SharedFingerprint`/`CheckSharedDrift` (HTTP/Log excluded; DB/env/schema shared); `TestSharedFingerprint*`, `TestCheckSharedDrift` |
| 14 | Tenant/runtime changes flow only through the versioned/audited rule engine (#27) | rule engine (Phase 7) is the only runtime-config path; modules get only their `modules.<name>` config namespace (Phase 5 contract) |
| 15 | Container-first verification | host `make ci`; `make ci-container` |
| 16 | Evidence bundle + review | this directory; review-findings.md |

Carried forward: full OTel span export (the framework emits trace_id in logs + provides the metrics/
access-log middleware; a concrete OTel exporter is a product adapter); pgx SQL-span tracer; Grafana
dashboards + alert rules (ops artifacts). Graphify `extract` blocked on LLM key (R11).
