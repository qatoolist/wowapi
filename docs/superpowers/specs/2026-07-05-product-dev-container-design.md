# Design: `wowapi` product-dev container

- **Date:** 2026-07-05
- **Status:** Approved (design); spec under review
- **Goal:** Let a developer manually exercise the wowapi framework by dropping into a container whose
  working directory is a host path they provide, and building a real product there with the CLI + live
  backing services.

## Decisions (from brainstorming)
- **Framework link:** local mounted source. The container mounts this wowapi checkout **read-only** at
  `/wowapi`; the scaffolded product resolves the framework via a `replace` directive to that path. No
  publishing, no auth, no tags. Live framework edits (from the host) are picked up on the next product build.
- **First run:** empty shell — the developer drives (`wowapi init`, `new-module`, `gen crud`, migrate, run).
  No pre-scaffolded demo.
- **Read-only framework:** `/wowapi` is mounted `:ro`; the box cannot mutate the framework.
- **Example module path:** `github.com/qatoolist/wowproduct` in the cheatsheet (developer may use any path
  they own).
- **Working dir:** `/Users/qatoolist/wowtestdir` (created if missing), bind-mounted to `/workspace`.

## Non-goals (YAGNI)
Published-tag/private-auth consumption path; a pre-built demo module; Kubernetes; anything touching the
framework's public API. This is a dev harness under `deployments/` + `scripts/` only.

## Architecture

### Components
1. **`deployments/product-dev.yaml`** — a compose overlay (used with the existing `deployments/compose.yaml`)
   that adds one `devbox` service to the running stack (Postgres 16 + MinIO + Mailpit).
2. **`scripts/product-dev.sh <path>`** — host launcher: validates/creates `<path>`, brings up the services,
   runs the one-time DB/role bootstrap, then drops the developer into an interactive shell in `/workspace`.
3. **In-container helpers on `PATH`** (installed by the devbox image/entrypoint):
   - **`wow-link`** — after `wowapi init`, wires the local framework:
     `go mod edit -replace github.com/qatoolist/wowapi=/wowapi && go mod tidy`.
   - **`wow-dbinit`** — idempotent role + database bootstrap (see Service wiring).
   - the **`wowapi`** CLI itself, installed from `/wowapi`.

### The devbox service
- Builds from the existing `dev` Dockerfile stage (Go 1.26, git, make, bash; `safe.directory '*'`).
- **Mounts:** `../:/wowapi:ro` (this checkout, read-only), `${PRODUCT_DIR}:/workspace` (bind, read-write),
  and reuses the `gocache` + `gomod` named volumes so builds are cached.
- **Network:** joins the compose network — `postgres:5432`, `minio:9000`, `mailpit:1025` are reachable by
  service name.
- **Ports:** publishes `8080:8080` so the product API is reachable from the host (`curl localhost:8080`).
- **Env:** the framework-facing DSNs + object-store/mail endpoints (below).
- **Entrypoint:** `scripts/devbox/entrypoint.sh` (from the read-only `/wowapi` mount) — puts
  `/wowapi/scripts/devbox` on `PATH` (so `wow-link`/`wow-dbinit` are available), installs the CLI stamped
  `v0.0.0-dev`, prints the cheatsheet, then `exec bash` in `/workspace`. It does **not** scaffold, and it does
  **not** require any change to the framework `Dockerfile` (it reuses the `dev` stage as-is).

### Framework link (private + un-tagged bridge)
The CLI is installed with `-ldflags "-X …/internal/buildinfo.version=v0.0.0-dev"`, so `wowapi init` emits a
valid `require github.com/qatoolist/wowapi v0.0.0-dev`. `wow-link` then adds
`replace github.com/qatoolist/wowapi => /wowapi` and runs `go mod tidy`; the filesystem replace bypasses the
module proxy and sumdb, so the product builds directly against the read-only checkout.

### Service wiring (faithful RLS-enabled run)
The framework's roles matter: migration `00001` creates `app_rt`/`app_platform` **NOLOGIN** (login is an
out-of-band ops grant) and does **not** create `app_migrate` (the migrate runner just needs DDL rights). To
demonstrate tenant isolation, the API must run as a non-superuser `app_rt` login. `wow-dbinit` (run once by the
launcher, idempotent) connects to Postgres as the `wowapi` superuser and:
1. creates a fresh product database `wowproduct` (owner `wowapi`),
2. creates roles `app_rt` and `app_platform` **with LOGIN + a local password** *before* migrations run — so
   `00001`'s `CREATE ROLE IF NOT EXISTS … NOLOGIN` sees them present and preserves the LOGIN (the migration
   deliberately never re-asserts attributes on an existing role),
3. grants them `CONNECT` on `wowproduct`.

The devbox then exports the two DSNs the scaffolded `configs/local.yaml` references via `secretref://env/`:
- `MIGRATE_URL=postgres://wowapi:wowapi-local-only@postgres:5432/wowproduct?sslmode=disable` (superuser; runs
  DDL + creates the schema and grants app_rt its table privileges),
- `DATABASE_URL=postgres://app_rt:app-local-only@postgres:5432/wowproduct?sslmode=disable` (non-superuser
  runtime — RLS is enforced, not bypassed).

Object storage / mail (only needed once the product uses documents/notifications) are exported for
convenience: `S3_ENDPOINT=http://minio:9000` (+ the local MinIO creds) and `SMTP_ADDR=mailpit:1025`. The
minimal api + CRUD path does not require them.

### The developer flow (printed cheatsheet)
```
wowapi init --module github.com/qatoolist/wowproduct --name wowproduct
wow-link                                     # wire the local framework (replace → /wowapi)
wowapi config validate --dir configs --env local
go run ./cmd/migrate up                      # migrate product schema into Postgres (as MIGRATE_URL)
wowapi new-module --name tasks
wowapi gen crud --module tasks --resource task
go run ./cmd/migrate up
go run ./cmd/api                             # serves :8080 (mapped to host); runs as app_rt
# from the host: curl localhost:8080/healthz   → 200
```
Launch from the Mac: `scripts/product-dev.sh /Users/qatoolist/wowtestdir`.

## Error handling
- `<path>` missing → launcher creates it. Non-empty → allowed (developer scaffolds with `wowapi init --force`
  if needed; the shell doesn't force anything).
- Services not healthy → `docker compose up --wait` blocks until Postgres/MinIO/Mailpit report healthy.
- CLI build failure → surfaced by the entrypoint (non-zero exit before the shell).
- `wow-dbinit` is idempotent (DO-blocks + `CREATE DATABASE` guarded) — safe to re-run.

## Validation (the acceptance test)
Before handing over, run the whole flow **inside the container** end-to-end and assert:
`wowapi init` → `wow-link` → `go build ./...` compiles against `/wowapi` → `migrate up` applies kernel +
product migrations → `new-module` + `gen crud` → second `migrate up` → `go run ./cmd/api` starts →
`curl localhost:8080/healthz` returns 200. Tear down and confirm the scaffolded files persist on the host
under `/Users/qatoolist/wowtestdir`.

## Files
- `deployments/product-dev.yaml` (new)
- `scripts/product-dev.sh` (new)
- `scripts/devbox/entrypoint.sh`, `scripts/devbox/wow-link`, `scripts/devbox/wow-dbinit` (new helpers; live in
  the repo, ride into the devbox via the read-only `/wowapi` mount and are added to `PATH` — no framework
  Dockerfile change)
- `docs/operations/product-dev-container.md` (new — how to use it)
