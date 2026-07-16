# Getting Started

Go from zero to a running wowapi-based product. (To work on the framework itself instead, see
[README → Quick start B](../../README.md#quick-start-b--work-on-the-framework-repo).)

## Prerequisites

| Tool | Version | Why |
|---|---|---|
| Go | **1.26+** | The module targets `go 1.26`. Check: `go version`. |
| Docker + Compose | recent | Local PostgreSQL/MinIO/Mailpit/Jaeger stack + the container gate. |
| PostgreSQL client (`psql`) | 16 (optional) | `make db-shell`; not required to run. |
| `git` | any | Clone the framework if installing the CLI from source. |

## Step 1 — Get the `wowapi` CLI

**Option A (published):**
```bash
go install github.com/qatoolist/wowapi/cmd/wowapi@latest    # or @vX.Y.Z once a version is tagged
wowapi version
```

**Option B (from source — always works, even pre-publish):**
```bash
git clone https://github.com/qatoolist/wowapi && cd wowapi
go build -o bin/wowapi ./cmd/wowapi
./bin/wowapi version     # add ./bin to PATH, or use the full path below
```

> **Gap:** `go install …@vX.Y.Z` requires the version to be tagged/published on the Go module proxy. If
> that isn't set up yet, use Option B.

## Step 2 — Scaffold a product repository

`wowapi init <name>` creates a *new, separate* repo — the directory `<name>` — that depends on the framework:

```bash
wowapi init myapp --module github.com/acme/myapp   # creates ./myapp/
cd myapp
go mod tidy
```

The positional `<name>` sets both the new directory and the product name. Flags: `--module` (required, your
Go module path), `--name` (override the product name), `--dir` (base directory — the product goes in
`<dir>/<name>`), `--force` (scaffold into a non-empty directory). Omit `<name>` to scaffold directly into the
current directory (`--dir .`).

What it scaffolds:

```
myapp/
├── go.mod                     # requires github.com/qatoolist/wowapi
├── Makefile                   # build / test / lint / migrate-up / migrate-down
├── .gitignore
├── README.md
├── cmd/
│   ├── api/main.go            # HTTP server: auth (API key + optional OIDC/JWT), optional
│   │                          # S3/MinIO storage, OTel tracing, Prometheus metrics, i18n
│   │                          # locale middleware, seed-catalog readiness check
│   ├── worker/main.go         # outbox relay + job runner + scheduler + optional storage
│   └── migrate/main.go        # migrations → seeds.Sync → rules.SyncDefinitions
├── configs/
│   ├── base.yaml              # shared defaults (http, log, db pool, commented-out auth/storage)
│   └── local.yaml             # local overlay (environment: local, secretref DSNs)
├── internal/
│   ├── wire/modules.go        # returns []module.Module — register your modules here
│   └── appcfg/config.go       # product Config: config.Framework + Auth (OIDC) + Storage (S3/MinIO)
│                              # + module namespaces
└── tools/
    └── configcheck/main.go    # product-local checker the `wowapi config` CLI delegates to
                                # (links the COMPOSED appcfg.Config — validate/print/doctor/schema/diff)
```

### Generated vs. product-owned: what you edit, what wowapi regenerates

Every file above is committed to your product repo and yours to edit — `wowapi init` does not manage a
"do not touch" zone. But some files are thin framework boilerplate you'll rarely need to change, while
others are where your product's logic actually lives. Knowing which is which tells you what's safe to
regenerate (diff and reconcile) versus what always needs your own review:

| File | Nature | Typical edits |
|---|---|---|
| `cmd/api/main.go`, `cmd/worker/main.go`, `cmd/migrate/main.go` | Framework process shell | Rarely edited directly; extend by registering modules (`internal/wire`) and turning on config sections (`auth.oidc`, `storage`). Re-running `wowapi init --force` regenerates these — diff before overwriting if you *did* hand-edit one. |
| `tools/configcheck/main.go` | Framework process shell | Same as above — parameterized by `appcfg.Config`, no product logic lives here. |
| `internal/appcfg/config.go` | **Mixed** — framework sections generated, product sections yours | `Config.Auth`/`Config.Storage` are the framework's standard sections (regenerable); add your own top-level fields alongside them for product-specific settings. |
| `configs/base.yaml`, `configs/*.yaml` | Product-owned | Values only, never regenerated — `auth`/`storage` ship commented out; uncomment and fill in per environment. |
| `internal/wire/modules.go` | Product-owned | This is where you register every module (`wowapi new-module` scaffolds modules; you list them here). Never overwritten by `--force` behavior you'd want repeated. |
| `internal/modules/**` | Product-owned | Your business logic — routes, handlers, migrations, seeds. |

The standard adapter wiring (storage, OIDC/JWT auth, OTel tracing, Prometheus metrics, i18n locale
negotiation, seed + rule-definition sync) is generated so no product hand-writes it; product-specific
config values, module registration, and business logic stay entirely product-owned.

## Step 3 — Provide a database

The generated `configs/local.yaml` references three secretref env vars — `DATABASE_URL`, `MIGRATE_URL`, and
`PLATFORM_URL` — all required (api/worker fail closed without `platform_dsn`):

```bash
export APP_ENV=local
export DATABASE_URL="postgres://app_rt:secret@localhost:5432/myapp?sslmode=disable"
export MIGRATE_URL="postgres://app_migrate:secret@localhost:5432/myapp?sslmode=disable"
export PLATFORM_URL="postgres://app_platform:secret@localhost:5432/myapp?sslmode=disable"
```

You need a PostgreSQL 16 instance. The **first kernel migration creates `app_rt` and `app_platform`**
(the runtime + cross-tenant roles); **`app_migrate` is not created for you** — it's the role the migration
runner connects as, so point `MIGRATE_URL` at a role that can create roles (a superuser/owner, or a
pre-created `app_migrate` with `CREATEROLE`) for the first `migrate up`. For pure local experimentation you
can also point all three DSNs at a single dev superuser; production must use the dedicated least-privilege roles
(see
[Database & migrations](database-migrations.md) and the
[deployment checklist](../operations/deployment-checklist.md)).

> **One-time role provisioning:** the bootstrap migration creates `app_rt`/`app_platform` as `NOLOGIN` —
> ops must grant a login out-of-band before the DSNs above will connect: `ALTER ROLE app_rt LOGIN PASSWORD
> '…';` (same for `app_platform`). See `scripts/product-dev.sh` for the automated version of this step.

> Tip: the **framework** repo ships a ready local stack (`make up` starts PostgreSQL etc.). In your
> product repo you bring your own database or reuse that compose file as a starting point.

## Step 4 — Validate config, migrate, build, run

```bash
wowapi config validate --env local     # loads + validates config; exit 0 = OK (lists every problem on failure)
make migrate-up                        # go run ./cmd/migrate up  (applies kernel migrations)
make build                             # builds bin/{api,worker,migrate}
go run ./cmd/api                       # HTTP server on :8080 (config http.addr)
```

In a second terminal, start the background worker:

```bash
go run ./cmd/worker
```

**Expected output (api):** a JSON (or, in `local`, text) startup log line naming the environment and the
config **fingerprint**, then `listening addr=:8080`.

Verify it's alive:

```bash
curl -s localhost:8080/healthz   # 200, liveness
curl -s localhost:8080/readyz    # 200 + readiness incl. config_fingerprint (checks the DB)
```

> Business routes return **401** until you wire a real authenticator — the generated api uses a
> fail-closed `DenyAllAuthenticator` by design. See [Auth](auth.md).

## Step 5 — Add your first module

```bash
wowapi new-module --name widgets
# scaffolds internal/modules/widgets/ (module.go, migrations/, seeds/, openapi.json)

wowapi gen crud --module internal/modules/widgets --resource widget \
  --fields "title:string,count:int"
```

Then register it in `internal/wire/modules.go`:

```go
package wire

import (
    "github.com/qatoolist/wowapi/module"
    "github.com/acme/myapp/internal/modules/widgets"
)

func Modules() []module.Module {
    return []module.Module{
        &widgets.Module{},
    }
}
```

Re-run `make build`. The full module walkthrough — routes, permissions, handlers, migrations, seeds — is
in [Building & extending modules](modules.md).

## Common first-run mistakes

| Symptom | Cause | Fix |
|---|---|---|
| `db.dsn required: …` on start | `DATABASE_URL`/`MIGRATE_URL` not exported, or wrong `APP_ENV` | Export both vars; set `APP_ENV=local`. |
| `config: invalid configuration` | overlay declares the wrong `environment`, or an unsafe-in-prod value | `wowapi config validate --env <env>`; read the accumulated error list. |
| Every business route returns 401 | no real `Authenticator` wired (default is `DenyAllAuthenticator`) | Wire your OIDC/API-key authenticator — [Auth](auth.md). |
| `permission denied for table …` at runtime | connected as the wrong role, or migrations not applied | Ensure `app_rt` DSN + `make migrate-up`. |

Next: [Concepts & Architecture](architecture.md).
