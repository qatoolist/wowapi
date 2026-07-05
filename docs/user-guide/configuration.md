# Configuration

Configuration is computed **once, at boot**, from layered sources into a single validated, fingerprinted,
redaction-safe struct. If anything is invalid the process refuses to start and reports **all** problems at
once — never just the first. (`kernel/config/load.go`.)

## The layering order

From lowest to highest precedence (**later layers override earlier ones**):

```
compiled defaults
   └─◀ configs/base.yaml            committed, shared across environments
        └─◀ configs/<env>.yaml      the environment overlay (local|dev|stage|prod)
             └─◀ WOWAPI__* env vars environment-variable layer (prefix configurable)
                  └─◀ flags         local-tooling overrides — REFUSED when environment=prod
   then: secretref:// resolution  →  validation  →  fingerprint
```

- **Defaults** are compiled into the binary (`config.Defaults()`), so a minimal config still boots.
- **`base.yaml`** holds settings common to every environment.
- **`<env>.yaml`** overlays per-environment values; the file must declare a matching `environment:`.
- **Env vars** (`WOWAPI__…`) override files — ideal for containerized deploys and CI.
- **Flags** are for local tooling only; the loader **refuses to start when they're set and
  `environment: prod`**, so an operator can't hand-patch production config.

## The config structure

Top-level struct: `config.Framework` (`kernel/config/config.go`). A product embeds this and adds its own
module namespaces (see [the generated `internal/appcfg`](getting-started.md#step-2--scaffold-a-product-repository)).

```yaml
# configs/base.yaml — shared defaults
schema_version: 1                 # config file format version (default 1)

http:
  addr: ":8080"                   # listen address (default ":8080")
  read_header_timeout: "5s"       # max time to read request headers
  request_timeout: "30s"          # per-request handler timeout
  max_body_bytes: 1048576         # max request body size (1 MiB)
  cors_allowed_origins: []        # exact-match CORS allowlist (empty = none)

log:
  level: "info"                   # debug | info | warn | error
  format: "json"                  # json | text

db:
  max_conns: 16                   # pool size (range 2–200)
  query_timeout: "5s"             # per-query deadline (range 100ms–60s)
```

```yaml
# configs/local.yaml — environment overlay
environment: local                # local | dev | stage | prod  (REQUIRED, no default — fail-closed)

db:
  dsn: secretref://env/DATABASE_URL          # runtime role (app_rt)
  migrate_dsn: secretref://env/MIGRATE_URL   # migration role (app_migrate)
  platform_dsn: secretref://env/PLATFORM_URL # cross-tenant role (app_platform) — REQUIRED; api/worker fail closed without it
```

| Key | Type | Default | Notes |
|---|---|---|---|
| `environment` | enum | **none** | `local`/`dev`/`stage`/`prod`. No default by design — an unset env fails closed. |
| `schema_version` | int | `1` | Config format version. |
| `http.addr` | string | `:8080` | Listen address. |
| `http.read_header_timeout` | duration | `5s` | Slow-header (Slowloris) guard. |
| `http.request_timeout` | duration | `30s` | Per-request handler timeout. |
| `http.max_body_bytes` | int64 | `1048576` | Request body cap (1 MiB). |
| `http.cors_allowed_origins` | []string | `[]` | Exact-match origins; empty = no cross-origin. |
| `log.level` | string | `info` | `debug`/`info`/`warn`/`error`. |
| `log.format` | string | `json` | `json` for prod; `text` for local. |
| `db.dsn` | secret | — | Runtime DSN (`app_rt`). |
| `db.migrate_dsn` | secret | — | Migration DSN (`app_migrate`). |
| `db.platform_dsn` | secret | — | Cross-tenant DSN (`app_platform`); **required** — api/worker fail closed without it. |
| `db.max_conns` | int | `16` | Pool size, clamped to 2–200. |
| `db.query_timeout` | duration | `5s` | Server-side statement ceiling, clamped 100ms–60s. |

> Modules read their own config namespace via `mc.Config().Decode(&cfg)` inside `Register` — see
> [Modules](modules.md). Unknown keys are **rejected**, so a typo'd module config key fails the boot.

## Environment variables (`WOWAPI__*`)

The env layer maps `PREFIX__SECTION__FIELD` onto the dotted config key, splitting on the **double
underscore** `__` and lowercasing segments:

```bash
export WOWAPI__DB__MAX_CONNS=32        # → db.max_conns = 32
export WOWAPI__HTTP__ADDR=":9090"      # → http.addr    = ":9090"
export WOWAPI__LOG__LEVEL="debug"      # → log.level    = "debug"
```

The prefix defaults to `WOWAPI__` and is configurable (`--env-prefix`, or `Options.EnvPrefix`). This layer
sits above the file overlays, so it's the clean way to tune a containerized deployment without editing YAML.

## Secrets: `secretref://`

Secrets never sit in YAML as plaintext. A **`config.Secret`** field accepts only a reference of the form:

```
secretref://<provider>/<path>
```

The only provider wired by default is **`env`**, which reads an environment variable:

```yaml
db:
  dsn: secretref://env/DATABASE_URL
```

- References are **resolved once at boot** by the `secrets.Provider`; a resolution failure fails the load.
- `Secret` fields **reject raw values** — you cannot accidentally inline a password; it must be a
  `secretref://`. (`kernel/config/secret.go`.)
- `Secret` is **redaction-safe**: it never renders its value in logs, `String()`, JSON, or
  `wowapi config print`. This is why config printing requires the explicit `--redacted` flag.

> **Adding providers is a seam, not a shipped feature.** Only `env` is wired today. A Vault/SSM/GCP-Secret-
> Manager provider would implement `secrets.Provider` and be passed via `Options.Secrets`. Documented as a
> gap, not invented.

## Inspecting & validating config (CLI)

| Command | What it does |
|---|---|
| `wowapi config validate --env <env>` | Load + validate; exit 0 = OK, exit 1 = invalid (prints every problem). |
| `wowapi config print --env <env> --redacted` | Print the effective config as JSON (secrets redacted; `--redacted` is required). |
| `wowapi config doctor --env <env>` | Per-key **provenance** table (which layer set each key) + the fingerprint. |
| `wowapi config schema` | Emit the JSON Schema derived from struct tags. |

Common flags: `--dir` (config dir, default `configs`), `--base` (base file), `--env` (overlay/environment
name), `--env-prefix` (default `WOWAPI__`).

```bash
$ wowapi config doctor --env local
# → a two-column table:  KEY  LAYER
#   where LAYER is one of: default | base-file | env-file | env | flag | secret
# → fingerprint=<hash>
```

(The table shows which layer supplied each key — not the values themselves; use `config print --redacted`
if you need the effective values.)

The **fingerprint** is a stable hash of the effective config (secrets excluded). It appears in the startup
log and `/readyz`, so you can confirm two processes booted the *same* configuration.

## Common mistakes

| Symptom | Cause | Fix |
|---|---|---|
| `environment must be set` | overlay missing `environment:` | Add `environment: <env>` to `configs/<env>.yaml`. |
| `config: invalid configuration: …` (a list) | one or more invalid/out-of-range values | Read the joined list; fix each. Ranges: `db.max_conns` 2–200, `db.query_timeout` 100ms–60s. |
| `secretref resolution failed` | referenced env var not set | Export the var named after `secretref://env/<VAR>`. |
| Secret printed as `REDACTED` | working as designed | Use `wowapi config doctor` for provenance; values are intentionally never shown. |
| Flags rejected at boot | `--flag` overrides used with `environment: prod` | Remove local flag overrides in prod; use env vars/overlays. |
| Unknown key error | typo in a module config key | Fix the key; unknown keys are rejected on purpose. |

Next: [Modules](modules.md) · [Build & deploy](build-deploy.md).
