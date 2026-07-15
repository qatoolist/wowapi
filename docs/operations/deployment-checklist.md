# Deployment checklist

Domain-neutral operational guardrails for any product built on wowapi. Covers the
proxy-delegated concerns the framework assumes (S7) and the monitoring conventions the
framework exposes hooks for but does not itself run (O4).

## 1. Edge / reverse proxy (S7)

The framework sets the application-layer security posture **in-process** — every response
carries `X-Content-Type-Options: nosniff`, `X-Frame-Options: DENY`,
`Content-Security-Policy: frame-ancestors 'none'`, `Referrer-Policy: no-referrer`, and HSTS
(see `kernel/httpx.SecureHeaders`, unit-tested in `kernel/httpx/edge_test.go`). Request bodies
are capped (`http.max_body_bytes`, default 1 MiB) and requests time out (`http.request_timeout`,
default 30s). CORS is deny-by-default; set `http.cors_allowed_origins` per environment.

The reverse proxy still owns TLS termination and edge-level limits. Use
[`deployments/reference/nginx.conf`](../../deployments/reference/nginx.conf) as the starting point:

- [ ] TLS terminated with managed certs; TLSv1.2+ only; plaintext 308-redirected to HTTPS.
- [ ] HSTS emitted at the edge (belt-and-suspenders with the app header).
- [ ] `client_max_body_size` matches `http.max_body_bytes`.
- [ ] Proxy read/send timeouts ≥ `http.request_timeout` (so the proxy does not cut first).
- [ ] `server_tokens off`; version banners suppressed.
- [ ] Run `BASE=https://<host> deployments/reference/smoke.sh` post-deploy — it fails if any
      required security header is missing. Re-run quarterly as a drill. (CI already runs this smoke against
      a scaffolded product behind the reference nginx over TLS via `make smoke-reference` — the `reference-smoke`
      job — so the config + header wiring is regression-gated; this checklist item verifies your *live* host.)

## 2. Config-drift alerting (O4)

`/readyz` returns `config_fingerprint` — a SHA-256 over the canonical, redacted config
(`kernel/config.FingerprintOf`). The fingerprint changes if and only if effective config changes.
The framework exposes it but does not alert on it; wire this convention in your monitoring:

- [ ] Scrape `config_fingerprint` from `/readyz` on every replica.
- [ ] **Alert when the fingerprint changes without an accompanying deploy** — this is
      unreviewed config drift (a hand-edited secret store, a stale overlay, a rolled-back replica).
- [ ] **Alert when replicas of the same process (api/api, worker/worker) disagree** — a partial
      rollout or a divergent secret provider. All replicas of one process must share a fingerprint.
- [ ] Record the expected fingerprint per release so the alert has a baseline to compare against.

Reference Prometheus rule (adapt to your fingerprint exporter):

```yaml
- alert: WowapiConfigDrift
  expr: changes(wowapi_config_fingerprint_info[10m]) > 0 unless on() wowapi_deploy_in_progress
  for: 5m
  labels: { severity: warning }
  annotations:
    summary: "config fingerprint changed without a deploy on {{ $labels.instance }}"
```

## 3. Distributed tracing (O1)

The kernel exposes a `Tracer` port with a zero-cost `NoOpTracer` default (tracing is off until wired).
To turn it on, wire the OpenTelemetry adapter in the api/worker mains and add the `Trace` middleware to
the chain:

```go
tr, _ := oteladapter.NewOTLP(ctx, cfg.TraceSampleRatio) // OTLP → OTEL_EXPORTER_OTLP_ENDPOINT
defer tr.Shutdown(ctx)
// api chain: httpx.Chain(mux, httpx.RequestID(), httpx.Recover(log), observability.Trace(tr), …)
```

- [ ] Run a tracing backend that speaks OTLP. Local dev: the compose stack includes **Jaeger**
      (`make up`); its UI is at http://localhost:16686 and it receives OTLP on `:4318`. Production: point
      `OTEL_EXPORTER_OTLP_ENDPOINT` at your managed collector (Tempo/Jaeger/vendor).
- [ ] Set the sample ratio per environment (1.0 in dev, a small fraction in high-traffic prod).
- [ ] Propagation is automatic: the `Trace` middleware continues an inbound `traceparent`, and
      `Tracer.Inject`/`Extract` carry trace context across process boundaries (embed the traceparent in
      outbox events / job payloads to connect API → relay → worker).

## 4. Backup / restore

PITR + object-storage restore procedure and the quarterly drill: [backup-restore.md](backup-restore.md).

## 5. Schema migrations

Zero-downtime expand/contract pattern and the CI reversibility drill: [migrations.md](migrations.md).

### Seed catalog sync

Migrating the schema is not enough — the authorization/resource **catalogs** (permissions, roles,
resource types, relationship types) must also be synced into the database, or every request denies and
resource writes fail their FK. See
[Database & Migrations § Seeds](../user-guide/database-migrations.md#seeds-declarative-yaml-catalogs).

- [ ] The generated `cmd/migrate up` (`make migrate-up`) runs `seeds.Apply` automatically after
      migrations — confirm your deploy pipeline still calls this and hasn't replaced it with a custom
      migrate step that drops the sync.
- [ ] If you run `wowapi seed sync` standalone instead, it's part of the deploy pipeline, not a one-time
      setup step — run it on every deploy (idempotent).
- [ ] Confirm the api process's `/readyz` includes the `seed_catalogs` check (`app.ReadinessWithCatalogs`)
      so a pod that skipped seed sync fails readiness with an actionable message instead of taking traffic.
- [ ] Confirm `/readyz` reports `details.seed_catalog_hash` after seed-sync has run, for drift correlation.

## 6. Rate limiting

The kernel ships an in-process token-bucket limiter (`kernel/httpx.RateLimit` + `NewTokenBucket`).
It is opt-in — limits are product-specific. Wire it in the api chain and pick a key strategy:

- [ ] Per-IP limit at the edge (`httpx.RateLimit(httpx.NewTokenBucket(rate, burst), httpx.KeyByIP)`)
      to blunt unauthenticated floods. Behind the reference proxy, supply a keyFn reading
      `X-Forwarded-For` rather than `RemoteAddr`.
- [ ] Per-actor limit after the authz gate (`httpx.KeyByActor`) for authenticated abuse.
- [ ] A tighter, dedicated bucket on expensive/PII-export routes (custom keyFn per permission).
- [ ] Note: limits are **per pod** (in-memory); size them per-replica, or plug a shared limiter behind
      the `httpx.RateLimiter` interface. Over-limit responses are `429` + `Retry-After` + RFC 7807.
