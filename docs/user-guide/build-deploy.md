# Build & Deploy

How to build the binaries, render deployment manifests, run the local stack, and what to check before
production. Grounded in the root `Makefile`, `deployments/compose.yaml`, `Dockerfile`, and
`wowapi deploy render`.

## The processes

A wowapi product runs as (at least) two processes plus a migration step. In a **product repo** these are
the thin `cmd/*` mains that `wowapi init` scaffolds:

| Process | Command | Role |
|---|---|---|
| API | `go run ./cmd/api` → `bin/api` | Serves HTTP on `http.addr`. Runs as `app_rt` (RLS enforced). |
| Worker | `go run ./cmd/worker` → `bin/worker` | Drains the outbox relay, runs jobs, ticks the scheduler. |
| Migrate | `go run ./cmd/migrate up` → `bin/migrate` | Applies embedded migrations. Runs as `app_migrate`. Run **before** rolling out api/worker. |

Migrations are embedded, so each binary is self-contained — nothing loose to ship alongside it.

> The **framework repo** itself only builds `cmd/wowapi` (the CLI). The api/worker/migrate mains live in
> the product repo the CLI scaffolds.

## Building

Product repo (generated `Makefile`):

```bash
make build          # builds bin/{api,worker,migrate}
```

Framework repo:

```bash
make build          # builds all packages + the wowapi CLI
```

For a container image, the repo ships a root `Dockerfile`; build your product image the same way and inject
config via `WOWAPI__*` env vars and `secretref://env/…` secrets (see [Configuration](configuration.md)).

## Rendering deployment manifests

`wowapi deploy render` emits a manifest for your product. It **validates `--env`** against the config
loader's accepted set, so it never emits a manifest that can't boot (`internal/cli/deploy_cmd.go`):

```bash
# docker-compose manifest for prod
wowapi deploy render --format compose --name myapp --image myapp:1.2.3 --env prod --out deploy/compose.yaml

# just the environment-variable block (for k8s/systemd/etc.)
wowapi deploy render --format env --name myapp --env prod --out deploy/myapp.env
```

Flags: `--format` (`compose`|`env`, default `compose`), `--name` (default `app`), `--image`
(default `app:latest`), `--env` (`local`|`dev`|`stage`|`prod`, default `prod`), `--out` (default stdout).

The rendered manifest **never inlines the DSN** — `config.DB.DSN` is a `Secret`, so the manifest references
`WOWAPI_DB_DSN` / `WOWAPI_MIGRATE_DSN` env vars and the real DSNs live only in the environment. The api and
worker receive the runtime DSN; migrate receives the migrate DSN.

## The local development stack

The framework's `deployments/compose.yaml` (via `make up`) brings up everything a product needs locally:

| Service | Image | Purpose |
|---|---|---|
| `postgres` | `postgres:16-alpine` | The database (user/db `wowapi`, password local-only). |
| `minio` | `minio/minio` | S3-compatible object store (artifacts/documents). |
| `mailpit` | `axllent/mailpit` | SMTP sink for notification testing. |
| `jaeger` | `jaegertracing/all-in-one:1.57` | Tracing; OTLP HTTP receiver on `:4318`. |
| `neo4j` | `neo4j:5-community` | Graph database for Graphify exports and bridge analysis. |
| `tools` | (repo toolbox) | Containerized runner — where `make ci-container` executes. |

```bash
make up            # start the stack
make db-shell      # psql into local postgres
make logs          # tail infra logs
make down          # stop (keep volumes)
make reset         # stop and DELETE volumes
```

The `tools` service is preconfigured with `S3_ENDPOINT`, `SMTP_ADDR`, and
`OTEL_EXPORTER_OTLP_ENDPOINT=http://jaeger:4318`, so the container gate exercises object storage, email, and
tracing against real services — not mocks.

## Object storage (S3/MinIO)

The document framework (`kernel/document`, exposed via `module.Context.DocumentClasses()`) talks to blob
storage through the `storage.Adapter` port (`kernel/storage/storage.go`), never directly to S3/MinIO. Two
adapters satisfy that port:

| Adapter | Package | Use |
|---|---|---|
| In-memory | `kernel/storage` (`storage.NewMemory()`) | Tests and local dev without a real object store. |
| S3/MinIO | `github.com/qatoolist/wowapi/adapters/storage/s3` | Production and any local dev pointed at real MinIO. |

If your module registers a document class but the kernel has no storage adapter wired, `app.Boot` fails
closed: *"document classes are registered (…) but no storage adapter is wired: pass kernel.Deps.Storage"*.

### Wiring the S3/MinIO adapter (generated scaffold)

`wowapi init` scaffolds the S3/MinIO wiring for you: `internal/appcfg.Config` carries a `Storage` section
(`StorageConfig` — endpoint, bucket, region, secretref-only credentials, `use_ssl`, `presign_ttl`,
`create_bucket`), and the generated `cmd/api`/`cmd/worker` mains construct the adapter and pass it into
`kernel.Deps.Storage` automatically, gated on `cfg.Storage.Enabled()` (true once `storage.endpoint` is set):

```go
// Generated in cmd/api/main.go and cmd/worker/main.go — no product edits needed.
var store storage.Adapter
if cfg.Storage.Enabled() {
    s3a, serr := s3adapter.New(ctx, s3adapter.Config{
        Endpoint:     cfg.Storage.Endpoint,
        Bucket:       cfg.Storage.Bucket,
        Region:       cfg.Storage.Region,
        AccessKey:    cfg.Storage.AccessKey.Reveal(),
        SecretKey:    cfg.Storage.SecretKey.Reveal(),
        UseSSL:       cfg.Storage.UseSSL,
        PresignTTL:   cfg.Storage.PresignTTL,
        CreateBucket: cfg.Storage.CreateBucket,
    })
    if serr != nil {
        return fmt.Errorf("storage: %w", serr)
    }
    store = s3a
}

k, err := kernel.New(cfg.Framework, log, kernel.Deps{
    Pool: pool, Platform: platformPool, Tx: txm, Storage: store, /* … */
})
```

Leave `storage.endpoint` unset (the default in the generated `configs/base.yaml`, commented out) and
`Deps.Storage` stays nil — `app.Boot` still fails closed if a module registers a document class, exactly as
before. Set the section in an env overlay to enable it:

```yaml
storage:
  endpoint: "localhost:9000"
  bucket: "myapp-docs"
  access_key: "secretref://env/S3_ACCESS_KEY"
  secret_key: "secretref://env/S3_SECRET_KEY"
  presign_ttl: 15m
  create_bucket: true   # local/dev overlays only
```

`New` fails closed at boot if the bucket doesn't exist and `CreateBucket` is false — the same fail-fast
posture as the DB pool, so a missing bucket is a boot error, not a 500 on the first upload. Set
`CreateBucket: true` only for local/dev overlays; production buckets are provisioned out of band.

The adapter mirrors `storage.NewMemory`'s semantics exactly (same `KindNotFound` mapping, same idempotent
`Delete`, same checksum-verified `Stat`), so swapping it in changes nothing about how your document classes
behave — only where the bytes live.

### Local development against the compose MinIO

`make up` already starts a MinIO at `localhost:9000` (root user/password `wowapi` / `wowapi-local-only`; see
the table above), so pointing `storage.endpoint` at `localhost:9000` (or `minio:9000` from inside the
`tools`/api/worker containers) with `create_bucket: true` gets you a working local object store with no
additional infrastructure.

## Health & readiness

Every deployment exposes two infrastructure endpoints (public):

```bash
curl -s localhost:8080/healthz   # liveness — process is up
curl -s localhost:8080/readyz    # readiness — checks the DB + reports config_fingerprint
```

Wire `healthz` to your liveness probe and `readyz` to your readiness probe. The `config_fingerprint` on
`/readyz` lets you confirm every replica booted the same configuration.

## Pre-production checklist

The authoritative runbook is [`docs/operations/deployment-checklist.md`](../operations/deployment-checklist.md).
The essentials:

- [ ] Distinct least-privilege roles: `app_rt` (runtime), `app_migrate` (DDL), `app_platform` (cross-tenant).
- [ ] **RLS guard enabled** on the runtime `TxManager` (`WithRLSGuard`) — refuses to run tenant work as a
      superuser/`BYPASSRLS` role. Deployed processes **must** enable this.
- [ ] DSNs supplied via `secretref://env/…`, never inlined; `sslmode=require` (or stricter).
- [ ] `environment: prod`; **no** local flag overrides (the loader refuses them in prod).
- [ ] `log.format: json`.
- [ ] Migrations applied (`migrate up`) **before** api/worker roll out.
- [ ] A **real `Authenticator`** wired (not `DenyAllAuthenticator`) — see [Auth](auth.md).
- [ ] `make ci-container` green on the release commit.
- [ ] Worker running (outbox relay + jobs + scheduler), not just the api.
- [ ] Backups configured — see [`backup-restore.md`](../operations/backup-restore.md).

## Common problems

| Symptom | Cause | Fix |
|---|---|---|
| App refuses to boot in prod | local flag overrides present | Remove flags; use env vars/overlays. |
| `deploy render` exits non-zero | invalid `--env` | Use `local`/`dev`/`stage`/`prod`. |
| Tenant queries run with RLS off | RLS guard not enabled / over-privileged DSN | Enable `WithRLSGuard`; connect as `app_rt`. |
| Events/jobs never process | worker not running | Run `cmd/worker` alongside `cmd/api`. |
| `readyz` 503 | DB unreachable or migrations pending | Check DSN; run `migrate up`. |

Next: [CLI reference](cli-reference.md) · [Troubleshooting & FAQ](troubleshooting-faq.md).
