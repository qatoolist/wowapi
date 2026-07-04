# wowapi

[![CI](https://github.com/qatoolist/wowapi/actions/workflows/ci.yml/badge.svg)](https://github.com/qatoolist/wowapi/actions/workflows/ci.yml)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
<!-- These badges render once the repo is public (CodeQL/Scorecard need code
     scanning; pkg.go.dev needs a public module). Uncomment when going public:
[![CodeQL](https://github.com/qatoolist/wowapi/actions/workflows/codeql.yml/badge.svg)](https://github.com/qatoolist/wowapi/actions/workflows/codeql.yml)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/qatoolist/wowapi/badge)](https://securityscorecards.dev/viewer/?uri=github.com/qatoolist/wowapi)
[![Go Reference](https://pkg.go.dev/badge/github.com/qatoolist/wowapi.svg)](https://pkg.go.dev/github.com/qatoolist/wowapi)
-->


**A domain-neutral, reusable enterprise backend framework ("platform kernel") in Go.**

`wowapi` is a modular-monolith framework you consume as a versioned Go dependency
(`github.com/qatoolist/wowapi`). It gives a product team a production-grade spine — multi-tenant
PostgreSQL with row-level security, deny-by-default authorization, a transactional outbox, a job runner
and scheduler, HTTP primitives, configuration, observability, and a set of compliance/evidence primitives
— behind a small **module SDK**. Product domains (a housing society, a school, a clinic…) plug in their
resources, permissions, routes, workflows, events, jobs, seeds, and migrations **without touching kernel
code**.

> Status: **pre-1.0 (v0.1.0).** The public surface (`kernel` / `module` / `app` / `adapters` / `testkit`
> / `migrations` + `cmd/wowapi`) may still make breaking changes between minor versions. See
> [Versioning](#versioning--stability).

- 📘 **New here? Start with the [User Guide](docs/user-guide/README.md)** and
  [Getting Started](docs/user-guide/getting-started.md).
- 🏗️ **Designing on top of it?** Read [Concepts & Architecture](docs/user-guide/architecture.md) and the
  [Blueprint](docs/blueprint/README.md) (design rationale).
- 🤝 **Contributing to the framework itself?** Read the [working-capability layer](docs/working/README.md)
  (skills map, best practices, review gate).

---

## Table of contents
- [What it is & what problem it solves](#what-it-is--what-problem-it-solves)
- [Who should use it](#who-should-use-it)
- [Highlights](#highlights)
- [Architecture at a glance](#architecture-at-a-glance)
- [Repository layout](#repository-layout)
- [Prerequisites](#prerequisites)
- [Quick start A — build a product on wowapi](#quick-start-a--build-a-product-on-wowapi)
- [Quick start B — work on the framework repo](#quick-start-b--work-on-the-framework-repo)
- [Core concepts](#core-concepts)
- [Command reference](#command-reference)
- [Configuration & environment](#configuration--environment)
- [Testing & the regression gate](#testing--the-regression-gate)
- [Documentation map](#documentation-map)
- [Versioning & stability](#versioning--stability)
- [Known limitations & assumptions](#known-limitations--assumptions)
- [Contributing](#contributing)
- [FAQ](#faq)
- [License](#license)

---

## What it is & what problem it solves

Enterprise backends repeatedly re-implement the same unglamorous, easy-to-get-wrong plumbing: tenant
isolation, authorization, audit, idempotency, background jobs, outbox/eventing, config, migrations, and
the tests that prove they're safe. Getting any of these subtly wrong is a security or data-integrity
incident.

`wowapi` implements that plumbing **once**, domain-neutrally, with the safe defaults baked in
(deny-by-default authz, forced RLS, fail-closed tenancy, append-only audit, transactional outbox), and
exposes it through a small SDK so product teams write only their domain logic. The kernel carries **no**
product concepts; a product is a separate repo that depends on the framework and registers modules.

The reference domain used while designing it was a housing-society product, but **nothing
society-specific lives in the core** — it is a general enterprise backend kernel.

## Who should use it

- **Product teams** building a multi-tenant, PostgreSQL-backed enterprise backend in Go who want a
  vetted, secure spine instead of hand-rolling tenancy/authz/audit/jobs.
- Teams that value **safe-by-default** security and **compliance-grade evidence** (audit trails,
  gap-free numbering, retention/legal-hold, immutable artifacts) out of the box.
- Engineers comfortable with Go, PostgreSQL, and a modular-monolith deployment (one `api` + `worker` +
  `migrate` binary set per product, one database).

It is **not** a micro-framework, an ORM, or a frontend framework, and it is not (yet) a stable 1.0 API.

## Highlights

Everything below is implemented and tested in this repo (see the [User Guide](docs/user-guide/README.md)
for depth):

- **Multi-tenant PostgreSQL with Row-Level Security** — the runtime connects as a non-superuser role,
  RLS is `FORCE`d, and `app_tenant_id()` is fail-closed. Cross-tenant work runs under a separate platform
  role.
- **Deny-by-default authorization** — RBAC (scope-covering role assignments) → ReBAC (relationships) →
  ABAC (deny-first policies), plus machine-principal scopes, step-up/MFA hooks, and an opt-in decision
  cache. Enforced **per request** at the route gate.
- **Module SDK** — register resources, permissions, routes, seeds, migrations, jobs, event handlers,
  workflows, and more through a capability-scoped `module.Context`. Inter-module access via declared ports.
- **Async platform** — transactional outbox (event iff business commit) + relay, at-least-once job runner
  with retries/backoff/DLQ, and a leader-safe fixed-interval scheduler.
- **Compliance/evidence primitives** — durable field-level audit with hash-chaining, gap-free per-tenant
  sequence allocator, generalized legal hold + DSR (export/erasure) engine, bulk-operation framework, and
  immutable versioned artifacts.
- **HTTP primitives** — router with route metadata, RFC 9457 problem details, idempotency, ETag, keyset
  pagination, an injection-proof filter/sort DSL, and a fixed security-middleware chain.
- **Config, observability, ops** — layered config with `secretref://` secrets + fingerprint drift,
  Prometheus metrics + OpenTelemetry tracing adapters, rate limiting, and a container-first local stack.
- **A real regression gate** — `make ci` + `make ci-container` (DB/integration tests forced to run), a
  migration reversibility drill, boundary lint, fuzz targets, and perf budgets.

## Architecture at a glance

A modular monolith. Hexagonal boundaries only at the edges (HTTP, DB, object storage, providers). One
database per product; tenant isolation via `tenant_id` + RLS, applied with `SET LOCAL app.tenant_id` per
transaction.

```
                         ┌──────────────────────────────────────────────┐
   HTTP request ───────▶ │  api (thin main over wowapi/app)             │
                         │   httpx chain: RequestID → Recover → Trace →  │
                         │   metrics → SecureHeaders → CORS → BodyLimit →│
                         │   Timeout → [AuthN → tenant → AuthZ gate] →   │
                         │   module handler                              │
                         └───────────────┬──────────────────────────────┘
                                         │ module.Context (capability-scoped)
        ┌────────────────────────────────┼────────────────────────────────┐
        ▼                                 ▼                                 ▼
  ┌───────────┐                   ┌───────────────┐                 ┌──────────────┐
  │  modules  │  register into →  │  kernel        │  services      │  adapters    │
  │ (product) │                   │  authz, outbox,│                │ prometheus,  │
  │           │                   │  jobs, audit,  │                │ otel,        │
  │           │                   │  retention,…   │                │ secrets,…    │
  └───────────┘                   └──────┬────────┘                 └──────────────┘
                                         │ TxManager (SET LOCAL role + app.tenant_id)
                                         ▼
                                ┌──────────────────┐   worker (relay + job runner + scheduler)
                                │  PostgreSQL + RLS │◀──  migrate (goose migrations)
                                └──────────────────┘
```

Import law (enforced by `make lint-boundaries`): `kernel` ← `module`/`app` ← product modules; `adapters`
depend on kernel ports; modules reach the kernel **only** through `module.Context`. Full explanation:
[Concepts & Architecture](docs/user-guide/architecture.md).

## Repository layout

This is the **framework** repo. (A *product* repo scaffolded by `wowapi init` looks different — see
[Getting Started](docs/user-guide/getting-started.md).)

| Path | What it is |
|---|---|
| `kernel/` | The platform services (authz, database, httpx, outbox, jobs, audit, retention, sequence, bulk, artifact, config, observability, …). Public API. |
| `module/` | The product-facing SDK contract: `module.Module` + `module.Context`. Public API. |
| `app/` | The composition root: `kernel.New` + `App.Boot` + `StartWorker`. Public API. |
| `adapters/` | Vendor bindings behind kernel ports (`metrics/prometheus`, `tracing/otel`, `secrets/*`). |
| `testkit/` | Public test harness: isolated per-test DBs, fixtures, `RunModuleContract`. |
| `migrations/` | Embedded goose SQL migrations (`00001…`), kernel-owned. |
| `internal/` | CLI (`internal/cli`), the `wowapi` command wiring, neutral fixture modules, tools. |
| `cmd/wowapi/` | The `wowapi` CLI entry point. |
| `deployments/` | `compose.yaml` (local stack) + `reference/` (nginx + smoke). |
| `docs/` | [blueprint](docs/blueprint/README.md) (design), [user-guide](docs/user-guide/README.md), [operations](docs/operations/deployment-checklist.md), [implementation](docs/implementation/decisions.md) (decisions/evidence), [working](docs/working/README.md) (contributor standards). |
| `Makefile` | Developer + CI targets (see [Command reference](#command-reference)). |

## Prerequisites

- **Go 1.26+** (`go version`). The module targets `go 1.26`.
- **Docker + Docker Compose** — for the local stack (PostgreSQL, MinIO, Mailpit, Jaeger) and the
  authoritative container gate. `make up` / `make ci-container` need it.
- **PostgreSQL 16** — provided by the compose stack; a local `psql` client is optional (`make db-shell`).
- Optional: `golangci-lint` (installed by `make tools`; falls back to `go vet` if absent).

## Quick start A — build a product on wowapi

You consume `wowapi` from a **separate product repository**. The `wowapi` CLI scaffolds it.

**1. Get the `wowapi` CLI.** Either install a published version:

```bash
go install github.com/qatoolist/wowapi/cmd/wowapi@latest   # or @vX.Y.Z once tagged
```

…or build it from a clone of this repo (works even if no version is published yet):

```bash
git clone https://github.com/qatoolist/wowapi && cd wowapi
go build -o bin/wowapi ./cmd/wowapi     # then use ./bin/wowapi
```

**2. Scaffold a product repo:**

```bash
mkdir myapp && cd myapp
wowapi init --module github.com/acme/myapp --name myapp
# scaffolds: go.mod, Makefile, cmd/{api,worker,migrate}, configs/{base,local}.yaml,
#            internal/{wire,appcfg}, tools/configcheck, README.md, .gitignore
go mod tidy
```

**3. Start a database and set the env vars the local overlay expects** (`configs/local.yaml` uses
`secretref://env/DATABASE_URL` and `secretref://env/MIGRATE_URL`):

```bash
export APP_ENV=local
export DATABASE_URL="postgres://app_rt:secret@localhost:5432/myapp?sslmode=disable"
export MIGRATE_URL="postgres://app_migrate:secret@localhost:5432/myapp?sslmode=disable"
```

**4. Validate config, migrate, run:**

```bash
wowapi config validate --env local     # delegates to your tools/configcheck; exit 0 = OK
make migrate-up                        # go run ./cmd/migrate up
make build                             # builds bin/{api,worker,migrate}
go run ./cmd/api                       # listens on :8080 (config http.addr)
# in another shell:
go run ./cmd/worker                    # outbox relay + job runner + scheduler
```

Expected: the api logs a startup line with the config fingerprint and `listening addr=:8080`.
`GET /healthz` returns 200; `GET /readyz` returns readiness + the config fingerprint.

**5. Add your first module:**

```bash
wowapi new-module --name widgets                       # scaffolds internal/modules/widgets/
wowapi gen crud --module internal/modules/widgets --resource widget --fields "title:string,count:int"
```

Register it in `internal/wire/modules.go`, then re-run `make build`. Full walkthrough:
[Building & extending modules](docs/user-guide/modules.md).

## Quick start B — work on the framework repo

```bash
git clone https://github.com/qatoolist/wowapi && cd wowapi
make setup                  # install host tools + go mod download
make up                     # start postgres + minio + mailpit + jaeger + toolbox
make migrate                # apply kernel migrations to the local DB
make ci                     # host CI: vet, lint, boundaries, unit, race, perf budgets, build
make ci-container           # AUTHORITATIVE gate: runs `make ci` in the toolbox with DB tests FORCED
```

`make ci-container` is the gate that must pass before anything ships (it forces DB/integration tests to
run — a green host suite can hide skipped DB tests). See
[Testing & the regression gate](#testing--the-regression-gate).

## Core concepts

| Concept | One-liner | Learn more |
|---|---|---|
| **Kernel / Module / App** | Kernel = services; Module = your domain plugged in via `Register`; App = composition root that boots them | [architecture](docs/user-guide/architecture.md) |
| **Tenant + RLS** | Every tenant-scoped row is isolated by `tenant_id` + forced RLS; the runtime binds the tenant per transaction | [architecture](docs/user-guide/architecture.md), [database](docs/user-guide/database-migrations.md) |
| **Actor & authorization** | Users/machines act as an `Actor`; `RouteMeta` declares the permission; the gate enforces deny-by-default per request | [auth](docs/user-guide/auth.md) |
| **TxManager / TenantDB** | Do work in the caller's tenant transaction so business + kernel writes commit atomically | [database](docs/user-guide/database-migrations.md) |
| **Outbox / Jobs / Scheduler** | Events are written iff the business tx commits; jobs are at-least-once; the scheduler is leader-safe | [architecture](docs/user-guide/architecture.md) |
| **secretref config** | Config is layered; secrets are `secretref://env/VAR` references, never plaintext | [configuration](docs/user-guide/configuration.md) |

## Command reference

**`wowapi` CLI** (`wowapi help`, or per-command `wowapi <cmd> -h`):

| Command | Purpose |
|---|---|
| `wowapi version` | Print CLI version + check the go.mod dependency version |
| `wowapi init --module <path>` | Scaffold a product repository |
| `wowapi new-module --name <name>` | Scaffold a module package implementing `module.Module` |
| `wowapi gen crud --module <dir> --resource <name> --fields <...>` | Generate CRUD scaffolding for a resource |
| `wowapi migrate create --name <n>` | Create the next-numbered migration file |
| `wowapi config validate\|print\|doctor\|schema\|diff` | Validate/inspect config (delegates to product `tools/configcheck`) |
| `wowapi seed validate` | Validate a module's seed bundle |
| `wowapi openapi merge [fragments...]` | Merge OpenAPI fragments into one document |
| `wowapi lint boundaries` | Import-law + vocabulary boundary lint |
| `wowapi deploy render --env <env>` | Render deployment manifests (compose/env) |
| `wowapi dlq <jobs\|events> <list\|inspect\|replay\|discard>` | Inspect/operate the dead-letter queues (needs `DATABASE_URL`) |

**Make targets (framework repo)** — `make help` lists all. The essentials:

| Target | Purpose |
|---|---|
| `make up` / `make down` / `make reset` | Start / stop / wipe the local stack |
| `make migrate` | Apply kernel migrations to the local DB |
| `make ci` | Host CI (vet, lint, boundaries, unit, race, perf budgets, build) |
| `make ci-container` | **Authoritative** gate (DB tests forced) |
| `make test-integration` / `make test-security` / `make test-fuzz` | Focused suites |
| `make shell` / `make db-shell` | Toolbox shell / `psql` |

Full reference: [CLI reference](docs/user-guide/cli-reference.md).

## Configuration & environment

Config is layered: **compiled defaults ← `configs/base.yaml` ← `configs/<env>.yaml` ← `WOWAPI__*` env
vars ← secret references**. Secrets are `secretref://env/<VAR>` (or a cloud provider ref) — never
plaintext; `config.Secret` is compiler-redacted.

Key environment variables:

| Variable | Used by | Purpose |
|---|---|---|
| `APP_ENV` | product api/worker/migrate | Selects the `configs/<env>.yaml` overlay |
| `DATABASE_URL` | api/worker (`db.dsn`), local tooling | Runtime DSN (role `app_rt`) |
| `MIGRATE_URL` | migrate (`db.migrate_dsn`) | Migration DSN (role `app_migrate`) |
| `WOWAPI__*` | config loader | Env override for any config key (e.g. `WOWAPI__LOG__LEVEL=debug`) |
| `WOWAPI_TEST_DSN` / `DATABASE_URL` | `testkit` | DB for integration tests |
| `WOWAPI_REQUIRE_DB=1` | tests / `ci-container` | Makes DB tests **fail** instead of skip |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | otel tracing adapter | Collector endpoint (e.g. `http://jaeger:4318`) |

Details + samples: [Configuration](docs/user-guide/configuration.md).

## Testing & the regression gate

- **Unit** (`make test-unit`) — no external services.
- **Integration** (`make test-integration`) — real PostgreSQL via `testkit` (isolated per-test DBs).
- **Contract** (`make test-contract`) — module contract + external-consumer suite.
- **Security** (`make test-security`) — RLS isolation, deny-by-default, redaction, unsafe-config.
- **Fuzz** (`make test-fuzz`) — the filter DSL parser + cursor decoder.
- **Authoritative gate** (`make ci-container`) — runs everything in the toolbox with
  `WOWAPI_REQUIRE_DB=1`, so DB tests **must** run. This is the release gate.

A green host suite can be hollow if DB tests silently skip — always trust `make ci-container`. Full guide:
[Testing](docs/user-guide/testing.md).

## Documentation map

| You want to… | Read |
|---|---|
| Understand and start using it | [User Guide](docs/user-guide/README.md) → [Getting Started](docs/user-guide/getting-started.md) |
| Understand the design & concepts | [Concepts & Architecture](docs/user-guide/architecture.md), [Blueprint](docs/blueprint/README.md) |
| Configure it | [Configuration](docs/user-guide/configuration.md) |
| Build/extend a module | [Modules](docs/user-guide/modules.md) |
| Handle DB & migrations | [Database & Migrations](docs/user-guide/database-migrations.md) |
| Do auth / validation / errors | [Auth](docs/user-guide/auth.md), [Validation & Errors](docs/user-guide/validation-errors.md) |
| Test & regress | [Testing](docs/user-guide/testing.md) |
| Build & deploy | [Build & Deploy](docs/user-guide/build-deploy.md), [Deployment checklist](docs/operations/deployment-checklist.md) |
| Troubleshoot | [Troubleshooting & FAQ](docs/user-guide/troubleshooting-faq.md) |
| Contribute to the framework | [Working-capability layer](docs/working/README.md) |
| See design decisions / traceability | [Decisions](docs/implementation/decisions.md), [Evidence](docs/implementation/evidence/README.md) |

## Versioning & stability

- **Semantic-ish, pre-1.0.** Current: `v0.1.0`. Per [CHANGELOG.md](CHANGELOG.md) (Keep a Changelog), the
  public surface may make **breaking changes between minor versions** until 1.0.
- **Product pinning.** A product pins an exact `wowapi` version in its `go.mod`. The module contract-test
  suite (`testkit.RunModuleContract`) is the upgrade tripwire: run it in CI against a new framework
  version before upgrading.
- **Migrations are additive + reversible** — every migration ships an Up and a Down; the reversibility
  drill runs in `ci-container`.

## Known limitations & assumptions

Documented honestly rather than hidden:

- **Published version tags** — `go install …@vX.Y.Z` assumes the version is tagged/published on the
  module proxy; if not, build the CLI from a clone (Quick start A, step 1).
- **One PostgreSQL database per product**, modular monolith — not a microservice mesh.
- **Read-replica routing** for authz reads is a deployment seam (point `WithTenantRO` at a replica), not
  turnkey. Cross-process trace propagation through outbox events/job payloads is a documented follow-up.
- **Deferred/fail-closed features** (documented, not bugs): workflow vote/min-approval/self-approval are
  fail-closed at definition validation; `gen crud` emits honest TODO handler stubs to fill in.
- **A product must provide** its OIDC `Authenticator` (the generated api wires a fail-closed
  `DenyAllAuthenticator` until you do), object-storage adapter (if using documents), and secret provider
  for non-env secrets.

## Contributing

Framework contributors follow the **[working-capability layer](docs/working/README.md)**:
[best practices](docs/working/best-practices.md), [coding conventions & skills map](docs/working/skills-and-knowledge-map.md),
the [working persona](docs/working/working-persona.md), and the mandatory
[Independent Review & Quality Gate](docs/working/quality-gate-checklist.md) before anything ships. Run the
mechanical checks with `sh miscellaneous/review_gate.sh` (add `--full` for `make ci-container`). Every
design deviation is recorded in [decisions.md](docs/implementation/decisions.md) before the code.

## FAQ

**Is this an ORM / web framework / microservice toolkit?** No. It's a domain-neutral backend *kernel* +
module SDK for a modular monolith. You write SQL (parameterized) and HTTP handlers; it gives you tenancy,
authz, audit, jobs, config, and safe primitives.

**Do I fork wowapi to build my product?** No — you `go get` it as a dependency in your own repo and
register modules. `wowapi init` scaffolds that repo.

**Why one database + RLS instead of a DB per tenant?** Simplicity and correctness: RLS `FORCE` + a
fail-closed `app_tenant_id()` makes cross-tenant leakage impossible even on a coding mistake, without the
operational cost of N databases.

**How do I know an upgrade is safe?** Pin the version, run `testkit.RunModuleContract` and your suite in
CI against the new version. See [Versioning](#versioning--stability).

**Where do secrets go?** Never in config files. Config holds `secretref://env/VAR` references; the value
comes from the environment (or a cloud secret provider). See [Configuration](docs/user-guide/configuration.md).

**A DB test "skipped" — is that OK?** In local runs a DB test skips without a DSN. The **authoritative**
gate (`make ci-container`, `WOWAPI_REQUIRE_DB=1`) makes them fail instead, so nothing hides.

More: [Troubleshooting & FAQ](docs/user-guide/troubleshooting-faq.md).

## License

Licensed under the **Apache License 2.0** — see [LICENSE](LICENSE) and [NOTICE](NOTICE). Apache-2.0 is a
permissive license with an explicit patent grant, suitable for adopting wowapi as a dependency in commercial
and open-source products alike.
