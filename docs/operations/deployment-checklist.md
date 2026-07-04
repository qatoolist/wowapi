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
      required security header is missing. Re-run quarterly as a drill.

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

## 3. Backup / restore

PITR + object-storage restore drill (O5) is documented in hardening phase H2 (`backup-restore.md`).
