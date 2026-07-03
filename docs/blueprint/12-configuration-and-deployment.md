# 12 — Configuration & Deployment

Configuration ownership and lifecycle for the framework-as-dependency model
([11-framework-distribution-and-consumption.md](11-framework-distribution-and-consumption.md)).
Goal: heavy configurability with simple, predictable behavior — typed structs, fail-fast boot,
immutable hot paths, secrets by reference, and no way to configure away a security guarantee.

## 1. The five configuration layers

| Layer | Owner | Examples | Defined in | Changes when | Mechanism |
|---|---|---|---|---|---|
| **Framework config** | wowapi | DB pool sizing, RLS session settings, HTTP timeouts/body limits, JWKS cache TTL, outbox relay batch/interval, job concurrency, workflow sweeper cadence, upload-limit *defaults*, audit retention *defaults*, otel exporter knobs, adapter defaults, CLI defaults | `kernel/config` typed structs with compiled defaults | wowapi release or deliberate product override | boot-time load, immutable |
| **Product/project config** | consuming app | module list (in code, not config), enabled adapters, service name, API base/public URLs, default locale/timezone, selected notification providers, project operational defaults | product-owned typed struct composed with `config.Framework` | product release | boot-time load, immutable |
| **Module config** | each product module | requests SLA defaults, assets behavior toggles, module job schedules, module provider keys *by reference* | `modules.<name>.*` namespace, decoded into a module-owned typed struct | product release | boot-time via `module.Context.Config()` only |
| **Deployment/environment config** | ops per env | DB DSN secret ref, object-storage endpoint, OIDC issuer, TLS, log level, exporters, worker counts, region, `environment` name | `configs/<env>.yaml` overlay + env vars + secret refs | per environment/rollout | boot-time load, immutable |
| **Tenant/runtime config** | tenant admins / platform ops | tenant rule values, notification preferences, retention overrides, workflow template overrides, allow-listed feature flags | **rule/config engine only** ([02-workflow-rules.md](02-workflow-rules.md)) | at runtime | versioned, validated, audited, optionally approval-gated; cached with event invalidation |

The boundary rule that prevents mixing: **if a value must change without a deploy, it is a rule
point, not config. If it changes per environment, it is deployment config. If a module needs it,
it lives under that module's namespace.** Framework structs never grow product fields; product
config never redefines framework keys (it overrides values through the same typed fields).

## 2. Typed contracts (`kernel/config`, public)

No global config maps, no `map[string]any` handed to services, no reflection-driven magic beyond
one well-audited struct binder.

```go
// kernel/config — consumer-facing contracts (importable; values flow only through constructors)
type Framework struct {
    Environment Env            `conf:"env"        validate:"oneof=local dev stage prod" doc:"deployment environment"`
    HTTP        HTTPConfig     `conf:"http"`
    DB          DBConfig       `conf:"db"`
    Auth        AuthConfig     `conf:"auth"`
    Outbox      OutboxConfig   `conf:"outbox"`
    Jobs        JobsConfig     `conf:"jobs"`
    Workflow    WorkflowConfig `conf:"workflow"`
    Uploads     UploadConfig   `conf:"uploads"`     // platform DEFAULTS; tenant values are rule points
    Audit       AuditConfig    `conf:"audit"`
    Obs         ObsConfig      `conf:"observability"`
    SchemaVersion int          `conf:"schema_version" validate:"required"` // config format version
}
type DBConfig struct {
    DSN          Secret        `conf:"dsn"       validate:"required" redact:"true"`
    MaxConns     int           `conf:"max_conns" default:"16" validate:"min=2,max=200"`
    QueryTimeout time.Duration `conf:"query_timeout" default:"5s" validate:"min=100ms,max=60s"`
}

// Secret: resolved value with structural redaction. String()/Format()/MarshalJSON/slog.LogValuer
// all emit "[redacted:<provider/key>]". The raw value is reachable only via Reveal(), which is
// lint-restricted to adapter packages.
type Secret struct{ /* unexported */ }

// Loader: files + env + secret resolution, strict by default (unknown keys are errors).
func Load[T any](opts Options) (T, Fingerprint, error)
func LoadDetailed[T any](opts Options) (Loaded[T], error) // + per-key provenance & warnings (doctor)
type Options struct {                                     // final shape per D-0014/D-0016
    BaseFile, EnvFile string            // configs/base.yaml, configs/<env>.yaml
    EnvPrefix         string            // "ACME__" → ACME__DB__MAX_CONNS=32
    Environ           []string          // env pairs; nil = os.Environ() (hermetic tests)
    Secrets           secrets.Provider
    Flags             map[string]string // local tooling only; refused when environment=prod
}

// ModuleView: the ONLY config surface modules see (returned by module.Context.Config()).
type ModuleView interface {
    Decode(out any) error      // strict-decodes modules.<name>.* into the module's typed struct
    // no Get(key), no parent traversal — a module cannot read framework or sibling config
}
```

- Every field carries: a `conf` key, a compiled **default** (or `validate:"required"`), validation
  tags, `redact` where applicable, and a `doc` string (feeds `wowapi config schema`).
- **Product config** composes rather than forks. Naming is fixed: the framework-owned struct is
  `config.Framework` (this package); the product-owned type is `Config` in the product's
  `internal/appcfg` package (scaffolded by `wowapi init`):
  `type Config struct { config.Framework; Product ProductConfig; Modules map[string]yaml.Node }`,
  loaded via `appcfg.Load(...)` / `appcfg.MustLoad()`. Product-owned fields live under `product.*`;
  module namespaces under `modules.*`. Framework keys keep framework meaning everywhere.
- **Module config isolation:** modules declare a typed struct with defaults/validation and decode
  it in `Register` via `ctx.Config().Decode(&cfg)`. Decode failure or leftover unknown keys =
  **boot failure**, reported per module. `RunModuleContract` asserts the module boots with an empty
  namespace (defaults must be complete) and rejects an invalid one.
- **No cycles:** `kernel/config` imports only stdlib + `kernel/secrets` types; every other package
  (including `kernel/rules`) *receives* its config sub-struct by value in its constructor. `app`
  is the only place that loads config and fans sub-structs out — same acyclic graph as
  [04-project-and-primitives.md](04-project-and-primitives.md) §1.

## 3. Precedence and environments

Effective config is computed once at boot, in exactly this order (later wins):

1. **Compiled framework defaults** (struct tags) — always present, always safe.
2. **`configs/base.yaml`** — product config file (committed).
3. **`configs/<env>.yaml`** — environment overlay (committed; `local|dev|stage|prod`).
4. **Environment variables** — `PREFIX__SECTION__FIELD` mapping; the deployment platform's knob.
5. **Secret resolution** — `secretref://` values resolved through the secret provider.
6. **CLI flags** — local tooling only (`--http.port` for a dev run); **refused in prod** (the
   loader errors if flags are set while `environment=prod`).

Production allows layers 1–5 only. Overlays are committed and reviewed; env vars are for
platform-injected values (region, replica counts, endpoints), not for smuggling business values.
Provenance is tracked per key (which layer set it) and shown by `wowapi config doctor`.
Two keys have narrowed layer rules: `environment` follows the trust rules in §4 (never from
flags; env var only when files are silent), and `modules.*` values come from config files only
until module decoding learns string coercion (D-0018 — env/flag strings would fail the modules'
strict typed decode confusingly).

## 4. Fail-fast boot & production safety

The loader fails startup — with the full list of problems, not just the first — on: missing
required fields, range/format violations, **unknown keys** (typo defense), cross-field
inconsistencies, unresolvable secret refs, `schema_version` newer than the wowapi dependency
supports (or older than its supported floor), and any **production safety violation**:

- **Not a knob (no config key exists):** RLS enforcement, `SET LOCAL` tenant binding, deny-by-default
  authz, route-metadata enforcement, audit writes for sensitive actions, secret redaction, webhook
  signature verification. These cannot be disabled by configuration in any environment — the safety
  comes from absence of the option, not from validation.
- **Dev-only knobs, env-gated:** the few legitimate dev conveniences (fake token issuer, seeded
  demo data, verbose SQL echo, public pprof, permissive CORS, `AllowFlags`) are declared with
  `unsafe:"true"` and the loader **refuses to start** when any is set and `environment=prod`
  (stage warns loudly). `wowapi config validate --env prod` applies the same check in CI.
- Sanity floors in prod: TLS-terminating base URL required, non-debug log level, DSN must be a
  secret ref (raw DSN string in a file = error), pool/timeout values inside safe ranges.
- **Environment is fail-closed:** deployed processes must set `environment` explicitly — the
  loader errors when it is absent from every layer. The compiled default (`local`) exists only for
  `Defaults()` in tests and local tooling; a production deploy can never silently validate under
  `local` rules because a missing/typoed env var left the field unset (phase-00 review finding SEC-1).
- **Environment is not downgradable (D-0017):** `environment` may never be set by CLI flags, and
  an environment variable may only *supply* it when no config file does — a mismatch with a
  committed file value is an error, not an override. Prod checks and the flag refusal key off the
  file-layer value, so a lower-trust layer cannot lower the gate it is checked against
  (phase-01 review finding SEC-5). Unsafe-knob refusal is evaluated against the **final bound
  values** — compiled defaults and non-scalar knobs included (findings SEC-3/SEC-4, D-0019).

## 5. Secrets: references, not values

Config files and env vars carry **references** — `secretref://<provider>/<path>` (env for local,
cloud secret manager in prod, k8s secrets where applicable) — resolved once at boot into
`config.Secret` values. Redaction is structural (the `Secret` type), so logs, errors, health
output, config dumps, CLI diagnostics, and OpenAPI metadata can never print a raw secret;
`testkit.AssertNoSecretsInLogs` and a CLI snapshot test keep it true. Rotation = rotate at the
provider and restart/roll pods (boot-time resolution keeps the hot path free of provider calls);
per-tenant provider credentials remain `credential_ref` columns resolved by adapters at use time
with short-lived caching ([07-platform-services.md](07-platform-services.md) §6).

## 6. Immutable hot paths, safe runtime change

- The validated config object is **frozen at boot**. Constructors receive their sub-struct by
  value; nothing re-reads files, env, or stores per request. There is no config watcher, no
  reload loop, no dynamic lookup on any request/job path.
- Anything that must vary at runtime or per tenant is a **rule point**: versioned, JSON-Schema
  validated, scope-resolved, audited, optionally approval-gated, cached with event invalidation
  ([02-workflow-rules.md](02-workflow-rules.md)). Feature flags are rule points; framework config
  holds only their platform *defaults* where needed.
- The single sanctioned live tweak: log level, via an authenticated admin endpoint — audited, and
  reset on restart. Everything else changes by rolling a new config (deploys are cheap; mystery
  state is not).

## 7. Process config boundaries (api / worker / migrate)

One product config schema, three narrowed effective views — each binary receives only what it needs:

| Process | Gets | Explicitly does not get |
|---|---|---|
| `cmd/api` | HTTP, auth, runtime DB (`app_rt` DSN), obs, module namespaces | migration DSN, provider send credentials it doesn't use |
| `cmd/worker` | runtime DB, jobs/outbox/workflow settings, provider credentials (notify, webhooks, storage), obs, module namespaces | HTTP server section, migration DSN |
| `cmd/migrate` | migration DSN (`app_migrate` secret ref), migration/seed settings | provider credentials, HTTP section |

`app.RunAPI/RunWorker/RunMigrate` perform the narrowing; unused sections are simply never wired.
**Fingerprinting:** at boot each process logs `config_fingerprint` = SHA-256 of its canonical
*redacted* effective config (also a metric label and a `/readyz` detail). Shared sections are
fingerprinted separately, so an api-vs-worker **drift** on shared config triggers a cheap alert —
catching half-rolled deploys.

## 8. CLI config & deploy tooling

```text
wowapi config init                     # scaffold configs/{base,local,dev,stage,prod}.yaml + typed Config stub
wowapi config validate [--env prod]    # full load+validation incl. unsafe-in-prod checks; CI gate
wowapi config doctor                   # effective config with per-key provenance, redacted; env sanity probes
wowapi config print --redacted [--env] # canonical effective config (redaction is not optional)
wowapi config diff --from dev --to prod# redacted effective diff between environments
wowapi config schema                   # JSON Schema from struct tags (framework + product + module specs)
wowapi deploy render --env prod        # render compose/k8s manifests from templates + effective config
```

**How the CLI sees product config types without importing product code:** the installed `wowapi`
binary never imports product packages (it can't — it's prebuilt). `wowapi init` scaffolds a tiny
**generated product-local checker** at `tools/configcheck/main.go` in the product repo: it imports
the product's `internal/appcfg` + `wowapi/kernel/config`, and emits the JSON Schema / validation
report / redacted effective config as JSON on stdout. The CLI's `config validate|doctor|print|diff|
schema` commands execute `go run ./tools/configcheck` inside the product repo and format its
output. In the framework repo (no product types), the same commands run against `config.Framework`
alone. The CLI is **not** a deployment platform: scaffolding, validation, rendering, and
diagnostics only; applying manifests stays with the team's normal tooling (kubectl/helm/CD).

## 9. Deployment guidance (product repos)

```text
configs/
  base.yaml            # product config, committed; secretrefs only, no secrets
  local.yaml dev.yaml stage.yaml prod.yaml   # overlays, committed
  config.schema.json   # generated by `wowapi config schema` (CI checks freshness)
deployments/
  compose.yaml         # local: pg+minio+mailpit + api/worker using configs/local.yaml
  k8s/                 # rendered by `wowapi deploy render`: Deployments (api, worker),
                       # migrate Job, ConfigMap (rendered non-secret config),
                       # Secret / ExternalSecret (secretref targets), HPA optional
```

- **Local:** compose mounts `configs/local.yaml`; `secretref://env/...` resolves from a
  git-ignored `.env`.
- **Kubernetes:** non-secret effective config → ConfigMap; secrets → k8s Secrets or
  external-secrets pointing at the cloud manager; migrate runs as a pre-rollout Job
  (expand-contract keeps N-1 pods compatible). Helm optional — rendered plain manifests are the
  supported baseline.
- **Generic containers (Cloud Run/ECS/Fly):** same images; overlay baked or mounted, env vars +
  platform secret store on top.
- **CI gates:** `wowapi config validate --env stage --env prod` + schema freshness + secret-scan
  (no raw secrets in configs/) run on every PR; a PR that flips an `unsafe` knob for prod cannot merge.
- **Container startup:** binaries run the same loader — invalid config exits non-zero before
  serving traffic (crash-loop is the alarm, `/readyz` never goes ready on bad config).
- **Migration-time vs runtime:** migrate consumes only its narrowed view (§7); runtime processes
  never hold `app_migrate` credentials.

## 10. Config anti-patterns (blocked or lint-checked)

Untyped YAML blobs passed into services · global config singletons / service-locator lookups ·
config reads on hot paths · secrets in files/env values (refs only) · modules reading framework or
sibling config · business values in config that belong in rule points (lint: magic business
numbers) · env-specific `if env == "prod"` branches in module code (behavior differences must be
typed config fields or rule points) · a runtime config platform before a real need exists.

## 11. Acceptance criteria (config & deployment)

1. The five layers in §1 are distinct in code: framework structs in `kernel/config`, product
   struct in the product repo, module namespaces via `ModuleView`, overlays per environment,
   tenant/runtime values only in the rule engine.
2. Modules receive only `modules.<name>.*` through `module.Context.Config()` — contract-tested;
   no API exists to read global config from a module.
3. Boot fails with a complete error list on invalid/unknown/unsafe config; `environment=prod`
   with any `unsafe` knob refuses to start (test matrix per knob).
4. Core security guarantees have no disabling configuration key at all.
5. Secrets appear only as references; redaction is structural and verified by tests and CLI output.
6. Hot paths read immutable boot-time config; zero config lookups in request/job flame graphs.
7. api/worker/migrate receive narrowed views; config fingerprints logged; shared-section drift alerts.
8. `wowapi config validate/doctor/print/diff/schema` and `wowapi deploy render` work from the
   installed CLI; validate runs as a CI gate.
9. The config model adds no package cycles: `kernel/config` sits at the base; only `app` loads and
   distributes; `rules` handles runtime change.
