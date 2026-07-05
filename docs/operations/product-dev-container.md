# Product-dev container — build a product on the framework, by hand

A throwaway dev box for manually exercising wowapi: it mounts a working directory you provide, wires a
scaffolded product to **this** framework checkout, and stands up the backing services (Postgres, MinIO,
Mailpit) so you can run `wowapi init`, generate modules, migrate, and run the API — end to end, locally.

The framework repo is private and un-tagged, so the product links to the framework via a filesystem
**`replace`** directive rather than a published `go get`. Your framework checkout is mounted **read-only** —
the box cannot change it.

## Prerequisites
- Docker (Desktop) running.
- This repository checked out (you're in it).

## Launch
```bash
scripts/product-dev.sh /Users/qatoolist/wowtestdir
```
The launcher:
1. creates the directory if missing and bind-mounts it at `/workspace`;
2. brings up `postgres`, `minio`, `mailpit` and waits for health;
3. bootstraps a product database (`wowproduct`) and the `app_rt` / `app_platform` **LOGIN** roles (so the API
   runs as a non-superuser and row-level security is enforced, not bypassed);
4. drops you into an interactive shell in `/workspace`.

Override defaults with env vars: `PRODUCT_DB=myapp APP_RT_PASSWORD=… scripts/product-dev.sh <dir>`.

## First-run flow (printed on entry)
```bash
wowapi init --module github.com/qatoolist/wowproduct --name wowproduct
wow-link                                  # link the product to /wowapi (adds the replace directive)
wowapi config validate --dir configs --env local
go run ./cmd/migrate up                   # migrate the product schema (runs as MIGRATE_URL / superuser)
wowapi new-module --name tasks
wowapi gen crud --module tasks --resource task
# register the new module in internal/wire/modules.go, then:
go run ./cmd/migrate up
go run ./cmd/api                          # serves :8080 (runs as app_rt — RLS enforced)
```
From your Mac:
```bash
curl localhost:8080/healthz   # {"status":"ok"}
curl localhost:8080/readyz    # {"checks":{"db":"ok"},...,"status":"ready"}
```

## How it's wired
| Piece | Value |
|---|---|
| Workspace | your dir → `/workspace` (bind mount; files persist on the host) |
| Framework | this checkout → `/wowapi` (**read-only**); product links via `replace … => /wowapi` |
| Runtime DSN (`DATABASE_URL`) | `app_rt@wowproduct` — non-superuser, RLS enforced |
| Platform DSN (`PLATFORM_URL`) | `app_platform@wowproduct` — **dedicated login** for cross-tenant catalog work (no role membership) |
| Migrate DSN (`MIGRATE_URL`) | `wowapi@wowproduct` — superuser, runs DDL |
| Config env (`APP_ENV`) | `local` → loads `configs/local.yaml` (which references the DSNs above) |
| Object storage | MinIO at `minio:9000` (`S3_ENDPOINT`) |
| Mail | Mailpit at `mailpit:1025` (`SMTP_ADDR`); web UI on the host at `localhost:8025` |
| API port | `8080` → host |

The CLI is installed inside the box from `/wowapi`, stamped `v0.0.0-dev`. Helpers on `PATH`: **`wow-link`**
(add/refresh the `replace` to the local framework — re-run after editing `go.mod`) and the **`wowapi`** CLI.

## Notes & caveats
- **Runtime/platform separation (CF-1).** `app_rt` (runtime) and `app_platform` (cross-tenant catalog) are
  **separate LOGIN roles**; the product's platform pool connects directly as `app_platform` via `db.platform_dsn`
  (`PLATFORM_URL`). The box does **not** grant `app_platform` membership to `app_rt` — that grant is
  cluster-global and would collapse privilege separation in every database on the cluster (including the test
  suite's). The api/worker **fail closed** if `db.platform_dsn` is unset, so this separation can't be skipped.
- **Local-only credentials.** Both roles use a throwaway local password (`app-local-only`); fine for a local box.
- **Read-only framework.** Edit the framework from your host (not inside the box); the product picks up the
  change on its next `go build`/`go run` (the `replace` points at the live checkout).
- The box runs commands non-interactively too (any compose command that includes the `product-dev.yaml`
  overlay needs `PRODUCT_DIR` set, since the `devbox` volume requires it):
  `PRODUCT_DIR=/path/to/product docker compose -f deployments/compose.yaml -f deployments/product-dev.yaml
  run --rm devbox -c '<cmds>'`.
- **Local / trusted networks only.** The base stack publishes Postgres on `0.0.0.0:5432` and the bootstrap
  gives `app_rt` a LOGIN with a well-known local password (`app-local-only`), so on an untrusted network the
  database would be reachable with known credentials. Run this on your own machine, not a shared host.

## Reset / teardown
Use the **base** compose file only for teardown — the `devbox` overlay requires `PRODUCT_DIR` at parse time,
and you don't need it just to remove volumes or run `psql`:
```bash
# stop services and wipe all data (incl. the product database):
docker compose -f deployments/compose.yaml down -v
# just drop the product database (keep other data):
docker compose -f deployments/compose.yaml exec -T postgres \
  psql -U wowapi -d postgres -c "DROP DATABASE IF EXISTS wowproduct;"
```
Your scaffolded product files remain in the working directory you provided.
