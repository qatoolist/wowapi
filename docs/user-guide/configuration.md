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
  read_timeout: "30s"             # max time to read the whole request incl. body (0 = unlimited; refused in prod)
  write_timeout: "60s"            # max time for response writes (0 = unlimited; refused in prod)
  idle_timeout: "120s"            # max keep-alive idle time between requests (0 = unlimited; refused in prod)
  request_timeout: "30s"          # per-request handler timeout
  max_body_bytes: 1048576         # max request body size (1 MiB)
  cors_allowed_origins: []        # exact-match CORS allowlist (empty = none)
  rate_limit:
    disabled: false               # set true to remove the default per-client rate limiter (enabled by default)
    requests_per_second: 20       # sustained requests/sec per client key
    burst: 40                     # burst capacity per client key

log:
  level: "info"                   # debug | info | warn | error
  format: "json"                  # json | text

db:
  max_conns: 16                   # pool size (range 2–200)
  query_timeout: "5s"             # per-query deadline (range 100ms–60s)
  max_conn_lifetime: "1h"         # max total age of a pooled connection (pgx default; 0 = pgx default)
  max_conn_idle_time: "30m"       # max idle age of a pooled connection (pgx default; 0 = pgx default)

webhook:
  outbound:
    ssrf_protection_disabled: false  # DANGEROUS; unsafe knob, refused in prod / warned in stage — see webhooks.md#outbound-ssrf-protection
    allowed_hosts: []                # exact-match hostname allowlist for intentional internal targets
    allowed_cidrs: []                # CIDR allowlist (e.g. "10.20.0.0/16") for resolved delivery addresses
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
| `http.read_timeout` | duration | `30s` | Connection-level whole-request read timeout; `0` = unlimited, refused in prod. |
| `http.write_timeout` | duration | `60s` | Connection-level response-write timeout; `0` = unlimited, refused in prod. |
| `http.idle_timeout` | duration | `120s` | Keep-alive idle timeout; `0` = unlimited, refused in prod. |
| `http.request_timeout` | duration | `30s` | Per-request handler timeout. |
| `http.max_body_bytes` | int64 | `1048576` | Request body cap (1 MiB). |
| `http.cors_allowed_origins` | []string | `[]` | Exact-match origins; empty = no cross-origin. |
| `http.rate_limit.disabled` | bool | `false` | Set `true` to remove the default per-client rate limiter (enabled by default as an opt-out guard). |
| `http.rate_limit.requests_per_second` | float64 | `20` | Sustained requests/sec per client key (per replica). |
| `http.rate_limit.burst` | int | `40` | Burst capacity per client key. |
| `log.level` | string | `info` | `debug`/`info`/`warn`/`error`. |
| `log.format` | string | `json` | `json` for prod; `text` for local. |
| `db.dsn` | secret | — | Runtime DSN (`app_rt`). |
| `db.migrate_dsn` | secret | — | Migration DSN (`app_migrate`). |
| `db.platform_dsn` | secret | — | Cross-tenant DSN (`app_platform`); **required** — api/worker fail closed without it. |
| `db.max_conns` | int | `16` | Pool size, clamped to 2–200. |
| `db.query_timeout` | duration | `5s` | Server-side statement ceiling, clamped 100ms–60s. |
| `db.max_conn_lifetime` | duration | `1h` | Max total age of a pooled connection before it is closed and replaced (pgx v5's own default). Bounds how long rotated credentials or drained LB backends linger on live connections. `0` = pgx default; non-zero values validated 1m–24h. |
| `db.max_conn_idle_time` | duration | `30m` | Max idle age of a pooled connection before it is closed (pgx v5's own default). `0` = pgx default; non-zero values validated 30s–24h. |
| `webhook.outbound.ssrf_protection_disabled` | bool | `false` | Disables outbound webhook SSRF protection entirely. `unsafe:"true"` — **refused in prod, warned in stage.** See [Webhooks](webhooks.md#outbound-ssrf-protection). |
| `webhook.outbound.allowed_hosts` | []string | `[]` | Exact-match hostname allowlist bypassing the address-class check for outbound webhook delivery. |
| `webhook.outbound.allowed_cidrs` | []string | `[]` | CIDR allowlist for RESOLVED outbound webhook delivery addresses (e.g. `10.20.0.0/16`). |
| `security.profile` | enum | `api` | `api` (default) or `browser` — see [Security profile](#security-profile-api-vs-browser) below. |
| `security.csrf.cookie_name` | string | `csrf_token` | Only consulted under `security.profile: browser`. |
| `security.csrf.header_name` | string | `X-CSRF-Token` | Only consulted under `security.profile: browser`. |
| `security.cookie.same_site` | enum | `lax` | `strict`/`lax`/`none`; only consulted under `security.profile: browser`. |
| `security.cookie.secure` | bool | `true` | Required `true` when `same_site: none`. |

> Modules read their own config namespace via `mc.Config().Decode(&cfg)` inside `Register` — see
> [Modules](modules.md). Unknown keys are **rejected**, so a typo'd module config key fails the boot.

## Security profile: `api` vs `browser`

`config.Security` (`kernel/config/security.go`) selects the framework's security posture **by profile**,
not by hand-assembling middleware per product (backlog B7; see the benchmark's "Security: Profiles, Not
Handler-Level Advice"). There are exactly two profiles:

```yaml
security:
  profile: api        # DEFAULT — leaving this section out is identical to profile: api
```

- **`api`** (the default): bearer/API-key auth, **no cookies**, CSRF disabled by contract (there is no
  cookie session to forge), strict JSON, CORS allowlist, RLS guard. This is exactly what wowapi does
  today — selecting it, or omitting `security:` entirely, changes **nothing**.
- **`browser`** (opt-in): additionally wires **CSRF token enforcement** on every state-changing request,
  **SameSite/Secure cookie defaults**, and a **CSP header profile** suited to HTML. No product gains any
  of this by doing anything other than explicitly selecting it.

```yaml
security:
  profile: browser
  csrf:
    cookie_name: "csrf_token"     # default shown
    header_name: "X-CSRF-Token"   # default shown
    field_name: "csrf_token"      # form-field fallback for classic HTML posts
  cookie:
    same_site: lax                # strict | lax | none
    secure: true                  # required when same_site: none
  csp: ""                        # empty uses the built-in HTML-safe default
```

**CSRF defense: double-submit cookie, not synchronizer tokens.** `kernel/httpx.CSRFProtect` needs no
server-side session store (backlog B7 deliberately scopes that out): a token is generated once, handed to
the browser as a (non-`HttpOnly`) cookie, and the client must echo it back via the configured header (or a
form field, for classic HTML form posts that can't set custom headers) on every state-changing request. A
request whose echoed token doesn't match the cookie is rejected with `403`. Safe methods (`GET`/`HEAD`/
`OPTIONS`/`TRACE`) are exempt and simply (re-)issue the cookie if one isn't already present.

`wowapi config validate` rejects an **incoherent** browser profile — e.g. a blank `csrf.cookie_name`, an
unrecognized `same_site`, or `same_site: none` without `secure: true` (browsers reject that combination
outright) — so a misconfigured browser profile fails the boot gate instead of silently shipping without
CSRF protection.

**The generated scaffold wires this, not just the config schema.** `wowapi init`'s `cmd/api/main.go`
appends `httpx.SecurityChain(cfg.Security)` to the middleware chain (innermost, right before the
auth-gated router), so switching `security.profile` to `browser` in config is enough — no product code
change needed. The generated `main.go` is identical either way: under the `api` profile the call appends
nothing (proven behavior-unchanged), under `browser` it activates CSP + CSRF at runtime.

The safe outbound HTTP client (DNS/IP-blocking SSRF guard for anything wowapi calls out to) is a separate,
already-shipped concern: `foundation/webhook.HTTPSender` (backlog B2). It is unaffected by, and unrelated to,
the security profile selected here.
## Concurrency: capacity budget + backpressure

`config.Concurrency` (`kernel/config/concurrency.go`) reasons about concurrency **across the whole
deployment shape**, not just one knob at a time: it bounds in-flight HTTP requests at the edge, and
validates that the declared shape (replica count × per-replica pool sizes + migration + admin reserve)
cannot exhaust the database before rate limits or backpressure engage.

```yaml
concurrency:
  http_max_in_flight: 0        # 0 = disabled (safe default); the backpressure limiter is off
  worker_max_jobs: 0           # worker pool size, for capacity bookkeeping only
  platform_max_in_flight: 0    # platform-pool in-flight cap, for bookkeeping only
  replicas: 0                  # 0 = deployment shape not declared; capacity check is a no-op
  runtime_pool_max: 0          # runtime (app_rt) pool max_conns per replica
  platform_pool_max: 0         # platform (app_platform) pool max_conns per replica
  migrate_pool_max: 0          # migrate process pool max_conns (counted once, not per replica)
  reserved_admin: 0            # connections reserved for admin/operator access
  capacity_mode: advisory      # advisory (warn only, default) | enforced (fail boot)
  overload:
    api_status: 503            # 503 or 429
    retry_after: 2s
```

**Capacity-budget formula** (checked whenever `concurrency.replicas` is non-zero):

```
replicas*(runtime_pool_max + platform_pool_max) + migrate_pool_max + reserved_admin <= db.max_conns
```

- `concurrency.replicas == 0` (the default): the shape is **not configured**, so the check is a
  deliberate no-op — it never passes or fails spuriously.
- `capacity_mode: advisory` (**the default**): an oversubscribed shape is reported as a warning
  (`wowapi config capacity`, boot logs) but **does not fail** `config validate` or process boot.
- `capacity_mode: enforced` (**opt-in**): the same oversubscribed shape **fails** `config validate`
  (and therefore boot) with a message citing the computed demand vs. `db.max_conns`.

Run `wowapi config capacity --dir configs --env <env>` to lint the budget independent of
`capacity_mode` — it always exits 1 on an oversubscribed shape, so CI can catch the problem early even
while production boot itself stays advisory.

**Migration path (advisory → enforced):** size `replicas`/`runtime_pool_max`/`platform_pool_max`/
`migrate_pool_max`/`reserved_admin` for your deployment, run `wowapi config capacity` until it reports
`capacity OK`, then flip `capacity_mode: enforced` in the environment overlay once you're confident the
shape is correct. Because the default is advisory, upgrading the framework never breaks an existing
deployment that hasn't set these fields yet.

**Backpressure middleware** (`httpx.Backpressure`, wired in the generated `cmd/api/main.go` when
`concurrency.http_max_in_flight > 0`): a bounded semaphore that rejects requests with `overload.api_status`
(503 default) + `Retry-After: overload.retry_after` **before** they reach auth or the database, once more
requests are concurrently in-flight than the configured cap. It is **disabled by default**
(`http_max_in_flight: 0`), so upgrading never causes an existing deployment to start returning the overload
status unexpectedly — size the cap using the capacity-budget knobs above, then set it explicitly. Rejected
requests increment the `http_overload_rejected_total` counter (labeled by route) and current occupancy is
exported as the `http_in_flight_requests` gauge.

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
| `wowapi config capacity --env <env>` | Check the concurrency capacity budget; exit 0 = within budget or shape not configured, exit 1 = oversubscribed. |

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
| `config: invalid configuration: …` (a list) | one or more invalid/out-of-range values | Read the joined list; fix each. Ranges: `db.max_conns` 2–200, `db.query_timeout` 100ms–60s, `db.max_conn_lifetime` 1m–24h (or 0), `db.max_conn_idle_time` 30s–24h (or 0). |
| `secretref resolution failed` | referenced env var not set | Export the var named after `secretref://env/<VAR>`. |
| Secret printed as `REDACTED` | working as designed | Use `wowapi config doctor` for provenance; values are intentionally never shown. |
| Flags rejected at boot | `--flag` overrides used with `environment: prod` | Remove local flag overrides in prod; use env vars/overlays. |
| Unknown key error | typo in a module config key | Fix the key; unknown keys are rejected on purpose. |

Next: [Modules](modules.md) · [Build & deploy](build-deploy.md).
