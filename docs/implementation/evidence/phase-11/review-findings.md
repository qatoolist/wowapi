# Phase 11 — Review Findings

Phase 11 is an additive observability / performance / security-hardening phase — no new mutable
domain surface. The review was structured to match that risk profile:

- **Security audit (primary):** the security-suite agent explicitly audited the enforced guarantees
  as part of building the `make test-security` gate and the per-knob unsafe-config matrix. It probed
  for real gaps and found **none** — see the findings below. This is the highest-value review for a
  hardening phase and it came back clean with evidence.
- **Correctness verification (integration):** the lead verified that the metrics middleware, health
  endpoints, config-drift primitives, and benchmark gate integrate and pass (all CI gates green,
  boundary lint clean, `make bench-budget` + `make test-security` exit 0).

## Security audit result — no gaps found

| Guarantee | Verified | Evidence |
|---|---|---|
| Deny-by-default authz has no config off-switch | ✓ | `evaluator.go` hardcodes `default_deny`; no `Framework` key can flip it; `TestDenyByDefault`, per-knob matrix `TestCoreSecurityGuaranteesHaveNoOffSwitch` |
| Secrets are references only (raw values rejected) | ✓ | `Secret.UnmarshalText` → `secrets.ParseRef` unconditionally; `TestSecretUnmarshalTextAcceptsOnlyRefs` |
| Secret redaction protects ANY key name (structural, not heuristic) | ✓ | `LogValuer` path; the new `TestSecretStructuralRedactionNonSensitiveKey` isolates the structural guarantee (the prior test used a key that also matched the suffix heuristic) |
| RLS enforcement has no disabling key | ✓ | the DB layer applies `SET LOCAL ROLE` unconditionally; escalation integration tests block self-grant at the DB privilege level |
| Unsafe production config fails startup | ✓ | `kernel/config` per-knob prod matrix — every `Validate()` rejection path exercised in `production` |

**Observational note (not a defect, flagged for awareness):** `bind.go`'s `enforceUnsafe()` reads
`unsafe:"true"` struct tags to gate risky knobs in production, but the current `Framework` struct has
no such tagged field — the mechanism is wired and dormant. A future unsafe knob added WITHOUT the tag
would bypass the binder's prod guard; the per-knob `Validate()` matrix is the compensating floor
today. Recorded as a maintenance guardrail (add the tag + a matrix row when introducing any such knob).

## New-code correctness (lead-verified)

| Area | Verification |
|---|---|
| `httpx.Health` liveness/readiness | Liveness never runs checks (a failing dep must not trip a liveness probe); Readiness → 503 on any failing check + reports the redacted config fingerprint; `TestLivenessAlwaysOK`, `TestReadiness*` |
| `observability` RED middleware + AccessLog | records route/method/status/dur + response bytes via a wrapping ResponseWriter; labels by the bounded route PATTERN (not raw URL) to keep metric cardinality bounded; `observability_test.go` (6 tests) |
| Prometheus adapter | implements `observability.Metrics` (compile-time assertion); `/metrics` handler serves the text format; `prometheus_test.go` (6 tests) |
| Config shared-section drift | `SharedFingerprint` excludes process-specific HTTP/Log and includes env/schema/DB; a HTTP-only difference does NOT change it, a DB/schema difference DOES; `CheckSharedDrift` catches divergence; `TestSharedFingerprint*`, `TestCheckSharedDrift` |
| Perf budgets | 24 hot-path benchmarks with `ReportAllocs`; config field read is 0.3 ns/op, 0 allocs (proves #17 — no reflection/map lookup on the hot path); `internal/tools/benchbudget` gate wired into `make ci` |
| Boundary integrity | `kernel/observability → kernel/httpx` is a kernel-internal import (no cycle); the Prometheus client dep lives only in `adapters/`; `scripts/lint_boundaries.sh` OK |

Reliability fix (bonus, resolved this phase): the recurring parallel-container CI flake that
intermittently hit Phases 8–11 (`FATAL: role "app_rt" is not permitted to log in` and
`tuple concurrently updated`) was root-caused and FIXED. Root cause: migration 00001's bootstrap
DO block re-asserted `ALTER ROLE app_rt NOLOGIN` on an EXISTING role on every (re-)run — flipping the
CLUSTER-GLOBAL role to NOLOGIN mid-run and fighting the test kit's out-of-band LOGIN grant, while
concurrent role DDL across parallel packages collided on the `pg_authid` tuple. Fix: (1) the bootstrap
migration now CREATEs each role NOLOGIN only when absent and never re-asserts attributes on an existing
role (the LOGIN attribute is owned out of band — ops in prod, testkit in tests), wrapped in a
benign-`tuple concurrently updated` EXCEPTION handler; (2) testkit retries the login-provisioning DDL
(`alterRoleWithRetry`) and the pool's first connection (`pingWithRoleRetry`) on the transient race.
Two consecutive green parallel `make ci-container` runs confirm it. This removes the "warm the template
serially first" workaround noted in earlier phases.

Residual / carried forward (honest): full OTel span export + a pgx SQL-span tracer are product adapters
(the framework emits trace_id in logs and ships the metrics/access-log middleware + the Metrics port);
Grafana dashboards + alert rules are ops artifacts; wiring the observability chain + /healthz//readyz//
metrics into a running server is the product's generated `cmd/api`/`cmd/worker` (the exact chain is in
the proof bundle). Graphify `extract` blocked on LLM key (R11).
