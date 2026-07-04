# Concepts & Architecture

wowapi is a **domain-neutral, modular-monolith backend framework** in Go. It gives a product team a
hardened kernel — multi-tenant PostgreSQL with row-level security, deny-by-default authorization, a
transactional async platform, and compliance primitives — behind a small module SDK. You write
**modules**; the framework composes and runs them.

This page is the mental model. For the *why* behind each decision, see the
[blueprint](../blueprint/README.md).

## The three layers

```
                      ┌─────────────────────────────────────────────┐
   product repo  ──▶  │  app/   composition root                     │
   (cmd/api,          │  ── wires []module.Module, builds the        │
    cmd/worker,       │     kernel services, starts HTTP + workers   │
    cmd/migrate)      └───────────────┬─────────────────────────────┘
                                      │ module.Context (capabilities)
                      ┌───────────────▼─────────────────────────────┐
   you write these ─▶ │  module/   the SDK contract                  │
                      │  ── Module{ Name, DependsOn, Register }      │
                      │  ── each module registers routes, migrations,│
                      │     seeds, health checks, OpenAPI, ports      │
                      └───────────────┬─────────────────────────────┘
                                      │ uses
                      ┌───────────────▼─────────────────────────────┐
   the hardened   ─▶  │  kernel/   services (no business logic)      │
   core               │  tx · authz · httpx · config · secrets ·     │
                      │  outbox · jobs · scheduler · audit ·         │
                      │  sequence · retention · bulk · artifact …    │
                      └───────────────┬─────────────────────────────┘
                                      │
                   adapters/  vendor bindings (pgx, object store, …)
                   migrations/ embedded goose SQL (kernel baseline)
                   testkit/   per-test isolated-DB harness
```

**Import law:** dependencies point *inward* — `app → module → kernel`, never the reverse; modules never
import each other directly (they talk through **ports**). This is mechanically enforced by
`make lint-boundaries`. Violating it fails CI.

| Layer | Directory | Responsibility | You edit it? |
|---|---|---|---|
| Composition root | `app/` | Build kernel, wire modules, run api/worker | Rarely (product `cmd/*` is thin) |
| Module SDK | `module/` | The `Module` + `Context` contract | No (it's the contract) |
| Your features | `internal/modules/*` (product) | Routes, handlers, migrations, seeds | **Yes** |
| Kernel | `kernel/*` | Tenancy, authz, http, async, compliance | No (framework) |
| Adapters | `adapters/*` | Vendor bindings behind kernel interfaces | Only new integrations |
| Migrations | `migrations/` | Embedded kernel baseline SQL | No (modules ship their own) |
| Test harness | `testkit/` | Isolated-DB integration tests | Use it in tests |

## Multi-tenancy & row-level security

Every tenant-scoped table is protected by PostgreSQL **row-level security (RLS)**, not by trusting
application `WHERE` clauses. The design is **fail-closed**:

- The runtime connects as a **non-superuser** role (`app_rt`) with `FORCE ROW LEVEL SECURITY`, so even the
  table owner can't bypass policies.
- Policies filter on `app_tenant_id()`, a function that **raises if no tenant is set** — a query with no
  tenant context returns an error, never "all rows".
- Each transaction issues `SET LOCAL app.tenant_id = $tenant` so the setting is scoped to that tx only.
- Three roles separate concerns: `app_rt` (runtime, least privilege), `app_platform` (deliberate
  cross-tenant operations), `app_migrate` (DDL/migrations).

You never write tenant filters by hand. You get a tenant-scoped DB handle from the **TxManager** and RLS
does the rest. See [Database & migrations](database-migrations.md).

## The request path

```
HTTP request
   │
   ▼
httpx server ── panic recovery, request ID, secure headers, body-size limit, rate limiting
   │
   ▼
Authenticator ── resolves the caller into an Actor  (default: DenyAllAuthenticator → 401)
   │
   ▼
Route gate (gateRoute / SecureHandler) ── reads the route's RouteMeta:
   │     • Public:true          → skip auth
   │     • Permission:"x.y.z"   → deny-by-default authz check (RBAC→ReBAC→ABAC + scopes + step-up)
   ▼
TxManager.WithTenant(...) ── opens a tx, SET LOCAL app.tenant_id, runs your handler under RLS
   │
   ▼
your handler ── decode+validate → do work → return; errors map to problem+json
```

**Deny-by-default** is the core posture: a route with no `RouteMeta.Public` and no satisfied `Permission`
is refused. The default `Authenticator` is `DenyAllAuthenticator`, so a freshly scaffolded product denies
every business route until you wire real authentication — safe by construction. See [Auth](auth.md).

## The async platform

Background work is **transactional and at-least-once**, never fire-and-forget:

- **Transactional outbox** — you write domain rows and outbox messages in the *same* tx. A **relay**
  (in `cmd/worker`) publishes them after commit, so a message is never lost or sent for a rolled-back
  change.
- **Jobs** — at-least-once execution with retries and a **dead-letter queue (DLQ)** for poison messages.
- **Scheduler** — leader-safe recurring work via `SELECT … FOR UPDATE SKIP LOCKED` on
  `next_run_at <= now()`, so multiple workers don't double-run a schedule.

All three run in the `cmd/worker` process. The api process serves HTTP; the worker drains async work.

## Compliance & platform primitives

The kernel ships primitives products usually have to build themselves:

| Primitive | Package | What it gives you |
|---|---|---|
| Audit log | `kernel/audit` | Hash-chained, tamper-evident event log |
| Sequences | `kernel/sequence` | Gap-free monotonic numbering (invoices, etc.) |
| Retention | `kernel/retention` | Legal holds, DSR erasure, disposition engine |
| Bulk ops | `kernel/bulk` | Batched, tenant-safe bulk mutations |
| Artifacts | `kernel/artifact` | Managed blob/object references |
| Idempotency | `kernel/httpx` | Idempotency-key dedupe on mutating routes |

Modules reach these through `module.Context` accessors — see [Building modules](modules.md).

## Configuration & secrets (in one line)

Config is layered: **defaults ← `base.yaml` ← `<env>.yaml` ← `WOWAPI__*` env vars ← `secretref://`
resolution**, producing a redaction-safe, fingerprinted config. Full detail in
[Configuration](configuration.md).

## Where to go next

- Build something: [Modules](modules.md)
- Understand the DB contract: [Database & migrations](database-migrations.md)
- Secure it: [Auth](auth.md)
- Test it: [Testing](testing.md)
- The design rationale in depth: [the blueprint](../blueprint/README.md)
