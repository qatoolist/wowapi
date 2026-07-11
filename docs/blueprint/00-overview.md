# 00 — Executive Recommendation, Principles, Architecture

> Blueprint for a reusable, domain-agnostic enterprise backend framework ("platform kernel") in Go.
> Source of truth: `Goal.md` (original vision prompt, archived in `wowapi2/archive/prompts-and-mandates/`; durable requirements in [../SRS.md](../SRS.md)). The housing society product is a *reference domain only*;
> nothing society-specific exists in the core. See [10-delivery.md](10-delivery.md) §Boundary Check.

## 1. Executive recommendation

Build a **modular monolith** with a **domain-neutral platform kernel**, **hexagonal boundaries at the
edges only** (HTTP, DB, object storage, providers), and a **module SDK** that lets product domains
(society, school, club, facility…) plug in resource types, relationship types, roles, permissions,
rule points, workflow definitions, events, jobs, seeds, and migrations — without touching kernel code.

**Distribution model:** `wowapi` (`github.com/qatoolist/wowapi`) is itself a versioned Go framework
dependency. Product applications live in **their own repositories**, add it with
`go get github.com/qatoolist/wowapi@vX.Y.Z`, import its public packages, and register their domain
modules through the module SDK. The framework repo ships no real product modules — only neutral
examples/fixtures. Consumer-facing contracts live in public packages (`wowapi/kernel`,
`wowapi/module`, `wowapi/app`, `wowapi/testkit`, `wowapi/adapters`), never only under Go
`internal/`. Full details: [11-framework-distribution-and-consumption.md](11-framework-distribution-and-consumption.md).

One binary set per product (`api`, `worker`, `migrate` — thin mains over `wowapi/app`), one
PostgreSQL database, tenant isolation via `tenant_id` + Row-Level Security enforced with
`SET LOCAL app.tenant_id` per transaction.

### Concrete stack (opinionated)

| Concern | Choice | Why / alternative rejected |
|---|---|---|
| Language / runtime | Go ≥ 1.23 | — |
| HTTP router | `chi` | stdlib-compatible `http.Handler`, zero magic. Gin/Echo rejected: custom context types leak everywhere. |
| DB driver | `pgx/v5` (pool) | Native Postgres features (`SET LOCAL`, LISTEN/NOTIFY, COPY), no `database/sql` lowest-common-denominator. |
| Query layer | `sqlc` per module | SQL stays visible and reviewable; no runtime ORM reflection. GORM rejected: hides SQL, fights RLS. |
| Migrations | `goose` with per-module embedded FS | Simple, embeddable, ordered. |
| Background jobs | **River** (Postgres-backed, pgx-native) behind a thin `jobs.Runner` interface | Transactional enqueue in the same tx as business writes; no extra infra. Redis/Kafka rejected for v1: operational cost, loses atomicity. |
| Workflow engine | **Custom Postgres-backed engine** in the kernel (see [02-workflow-rules.md](02-workflow-rules.md)) | Temporal is overkill for approval-style workflows and drags in a cluster; revisit if durable code-orchestration emerges. |
| Rules/config | Custom versioned rule registry, JSONB values validated by RuleValueSchema (a small closed grammar, not JSON Schema) | See [02-workflow-rules.md](02-workflow-rules.md). |
| AuthN | OIDC (external IdP: Zitadel/Keycloak/Auth0) + JWT middleware | Don't build identity; build *authorization*. |
| AuthZ | In-process evaluator: RBAC + scoped assignments + ReBAC relationships + policy conditions, deny-by-default | OPA/SpiceDB rejected for v1: network hop on the hottest path, second datastore to keep consistent. |
| Events | Transactional **outbox** table + relay worker; in-process handlers first, bus adapter later | Atomicity with business writes is non-negotiable. |
| IDs | **UUIDv7** everywhere (`google/uuid`) | Time-ordered → index-friendly; generated app-side via injectable `IDGen`. |
| Errors | RFC 9457 problem details envelope | See [04-project-and-primitives.md](04-project-and-primitives.md). |
| Logging | `log/slog` (JSON) | stdlib. |
| Observability | OpenTelemetry traces + Prometheus metrics | — |
| Config | layered typed config: compiled defaults → `base.yaml` → env overlay → env vars → secret refs; immutable after boot (see [12-configuration-and-deployment.md](12-configuration-and-deployment.md)) | No Viper-style runtime magic; runtime/tenant change goes through the rule engine. |
| Object storage | S3-compatible adapter (MinIO locally) | Presigned upload/download. |
| Validation | thin wrapper over `go-playground/validator` + explicit domain validation funcs | Tags for shape, code for rules. |
| OpenAPI | hand-authored fragments per module, merged by CLI; `oapi-codegen` for DTO/server stubs where it pays | Spec-first keeps API deliberate. |
| Testing | `testcontainers-go` Postgres + public `wowapi/testkit` | Real RLS tests, not mocks; usable from product repos. |
| DI | Manual composition root (`Kernel` + `App`), constructor injection | Wire optional later; no runtime container. |

## 2. Core framework principles

1. **Kernel is domain-blind.** The kernel compiles and ships with zero product concepts. If a word in
   kernel code would mean something to a housing-society lawyer, it's a bug (enforced by `wowapi lint boundaries`).
2. **Deny by default.** No tenant context → no tenant-scoped query. No permission metadata → route
   refuses to register. No matching allow → 403 + audited denial.
3. **Tenant isolation is structural, not disciplinary.** RLS at the DB + `TenantDB` type at the app
   layer. You cannot *reach* a connection for tenant data without a tenant-bound transaction.
4. **Everything important is a record.** Workflow state, rule versions, audit, outbox events, job runs,
   webhook deliveries — all Postgres rows, all queryable, all testable.
5. **Composition over inheritance.** Small embedded structs (`TenantScoped`, `Auditable`, `Versioned`),
   never a universal `BaseModel`.
6. **Interfaces at boundaries, structs inside.** Consumer-side small interfaces (`OutboxWriter`,
   `RuleResolver`); no interface for a thing with one implementation and no boundary.
7. **Explicit beats magical.** Constructor injection from a composition root; no service locator,
   no reflection container, no hidden hooks that mutate behavior silently.
8. **Async by outbox, never by fire-and-forget goroutine.** Side effects (notifications, webhooks,
   projections) ride the outbox → job runner with retries, idempotency, DLQ.
9. **Generate boilerplate, hand-write logic.** Codegen for module skeletons, sqlc wrappers, DTO
   mappers, seed stubs. Never generate business decisions.
10. **Fast by default on the hot path.** Auth+authz+tx+query for a simple GET budgeted at p95 < 50ms
    in-region; abstractions that cost allocations on this path must justify themselves.

## 3. Architecture style — comparison and choice

| Style | Verdict for v1 | Reason |
|---|---|---|
| **Modular monolith** | ✅ core choice | One deploy, real module boundaries, cheap refactors while the domain model is still moving. |
| Plugin/module architecture | ✅ core choice (compile-time) | Modules are Go packages registered at startup — not runtime `.so` plugins (fragile, unnecessary). |
| Hexagonal / ports & adapters | ✅ at the edges only | Adapters for storage, mail/SMS, IdP, payment callbacks. NOT five layers of indirection per module. |
| Clean/onion architecture | ⚠️ take the dependency rule, drop the ceremony | Dependency direction: `api → app → domain`, `store` implements `app` ports. No `usecase/interactor/presenter` taxonomy. |
| Layered architecture | ✅ inside each module | Thin, 3 layers (api/app/store around a domain core). |
| Vertical slice | ✅ modules *are* vertical slices | Each module owns its handlers→service→SQL for its resources. |
| Event-driven | ✅ for side effects only | Commands are synchronous; consequences are events via outbox. Not event-sourcing. |
| CQRS | ⚠️ lite | Separate command/query *objects and paths*; same database, no separate read store until a report proves the need. |
| Mediator | ❌ | A router already dispatches; an in-process mediator hides call graphs in Go. |
| Microservices | ❌ v1; extraction path preserved | Module boundaries + outbox events = future strangler seams. |

### Why start as a modular monolith
- Team is small; the tax of network boundaries (contracts, retries, tracing, deploy matrices) buys nothing yet.
- The kernel's guarantees (RLS in one DB, transactional outbox, one tx per request) are *easier to make true* in one process.
- Extraction later is mechanical if (a) modules never share tables, (b) cross-module calls go through
  declared Go ports, (c) async integration already uses events.

### How we avoid the failure modes
- **Big ball of mud:** module import rules enforced by lint (`modules/x` may import `kernel/*`,
  `shared/*`, and *declared ports* of other modules — never their `store` or `domain` packages).
- **Premature microservices:** the only distribution in v1 is `api` vs `worker` processes sharing one DB.
- **Over-engineering:** every abstraction in this blueprint names the concrete problem it solves;
  anything speculative is listed under "later only" in the pattern matrix ([09-patterns.md](09-patterns.md)).

## 4. Diagrams

### 4.1 High-level architecture

```text
                        ┌────────────────────────────────────────────────┐
   OIDC IdP ──tokens──▶ │                  /cmd/api                      │
                        │  chi router                                    │
   Clients ───HTTPS───▶ │   ├─ middleware: requestid → recover → otel    │
                        │   │   → authn(JWT) → tenant → capacity         │
                        │   │   → ratelimit → authz(route metadata)      │
                        │   ├─ kernel routes (tenants, users, roles,     │
                        │   │   rules, workflows, documents, webhooks…)  │
                        │   └─ module routes (registered via Module SDK) │
                        └───────────────┬────────────────────────────────┘
                                        │ pgx pool (app role, RLS FORCED)
                    ┌───────────────────▼────────────────────┐
                    │             PostgreSQL                  │
                    │  kernel tables + per-module tables      │
                    │  RLS: USING (tenant_id =                │
                    │        current_setting('app.tenant_id'))│
                    │  outbox / jobs / audit / idempotency    │
                    └───────────────────▲────────────────────┘
                                        │
                        ┌───────────────┴────────────────────────────────┐
   Email/SMS/WhatsApp ◀─│                 /cmd/worker                    │
   Outbound webhooks  ◀─│  outbox relay ─ job runner (River) ─ workflow  │
   Object storage    ◀──│  SLA timers ─ notification dispatch ─ DLQ      │
                        └────────────────────────────────────────────────┘
```

### 4.2 Framework layers

```text
┌───────────────────────────────────────────────────────────────┐
│ L3  Product Domain Modules      <product repo>/internal/      │
│     modules/* — separate repositories importing wowapi        │
│     society | school | club | facility | requests | assets    │
│     (business tables, business workflows, business rules)     │
├───────────────────────────────────────────────────────────────┤
│ L2  Domain Extension Layer      wowapi/module (public)        │
│     Module, Context, resource/relationship/role/              │
│     rule-point/workflow/event/job/seed registration           │
├───────────────────────────────────────────────────────────────┤
│ L1  Platform Kernel             wowapi/kernel/* (public)      │
│     tenant · auth · authz · policy · resource · relationship  │
│     workflow · rules · audit · outbox · jobs · document ·     │
│     notify · webhook · integration · httpx · errors ·         │
│     validation · pagination · database · config · o11y        │
│     (implementation guts live in wowapi/internal/*, private)  │
├───────────────────────────────────────────────────────────────┤
│ L0  Adapters                    wowapi/adapters/* (public)    │
│     postgres · s3 · smtp · sms · push · oidc · secrets        │
└───────────────────────────────────────────────────────────────┘
Import direction: L3 → L2 → L1;  L1 → L0 interfaces only.
L0–L2 can never import L3 — structurally: L3 lives in other repos,
and Go blocks anyone outside wowapi from importing wowapi/internal/*.
```

### 4.3 Module registration

```text
main() — the PRODUCT app's cmd/api, wired via wowapi/app
   │
   ├─▶ kernel.New(cfg) ──▶ builds pools, registries, services
   │
   ├─▶ app.Register(requests.Module{}, assets.Module{}, …)   // product-repo modules; order-independent
   │        │
   │        └─ for each module (topo-sorted by DependsOn):
   │             m.Register(ctx module.Context)
   │                ├─ ctx.Routes(...)        route + permission metadata
   │                ├─ ctx.Permissions(...)   catalog entries
   │                ├─ ctx.Roles(...)         role templates
   │                ├─ ctx.ResourceTypes(...) resource registry
   │                ├─ ctx.RelationshipTypes(...)
   │                ├─ ctx.RulePoints(...)    typed rule points
   │                ├─ ctx.Workflows(...)     workflow definitions
   │                ├─ ctx.Events(...)        event types + handlers
   │                ├─ ctx.Jobs(...)          job kinds + workers
   │                ├─ ctx.Migrations(fs)     embedded goose dir
   │                └─ ctx.Seeds(fs)          yaml seed bundles
   │
   ├─▶ app.Validate()   // startup fails hard: dup permissions, routes w/o
   │                    // metadata, unknown deps, unresolvable seeds
   └─▶ app.Start()      // migrate (cmd/migrate), seed sync, serve/work
```

### 4.4 Request lifecycle

```text
HTTP ─▶ requestID ─▶ recover ─▶ otel span ─▶ maxbytes/timeout
    ─▶ authn: verify JWT (OIDC JWKS) ─▶ principal
    ─▶ tenant resolve: path /t/{tenant} (+ header for APIs) ─▶ check user_tenant_access
    ─▶ acting capacity: X-Acting-Capacity or sole default ─▶ Actor{user, capacity}
    ─▶ rate limit (tenant+actor buckets)
    ─▶ route permission metadata ─▶ authz.Evaluate(actor, perm, scope)  [deny→403+audit]
    ─▶ handler:
         DecodeJSON → Validate → WithTenantTx(ctx, fn):
             BEGIN; SET LOCAL app.tenant_id = $1; SET LOCAL app.actor_id = $2;
             service.Do(cmd)         // domain logic
             repo writes             // RLS-checked
             outbox.Write(event)     // same tx
             audit.Write(action)     // same tx
             idempotency record      // same tx (if key present)
             COMMIT
    ─▶ WriteJSON(APIResponse) | WriteError(ProblemError)
```

### 4.5 Background job / event lifecycle

```text
COMMIT ─▶ events_outbox row (status=pending)
             │  (relay worker, batched, FOR UPDATE SKIP LOCKED)
             ▼
        publish to in-process dispatcher ─▶ for each registered handler:
             enqueue job (kind=event-handler, tenant_id, event_id, handler)
             mark outbox row dispatched
             ▼
        job runner (worker pool, bounded)
             ├─ SET LOCAL app.tenant_id from job payload
             ├─ inbox check: processed_events(handler, event_id)? skip
             ├─ run handler in tx; record inbox row on success
             ├─ failure: retry w/ exp backoff + jitter (policy per kind)
             └─ exhausted: dead-letter (job status=discarded + alert metric)
Scheduled jobs (cron): SLA escalation sweep, retention sweep, webhook retry,
notification digest — each tenant-iterating, each idempotent.
```

## 5. What lives where

| Belongs in… | Contents |
|---|---|
| **Kernel (L1)** | tenant context/resolver, authn middleware, authz+policy evaluator, relationship framework, resource registry, workflow runtime, rule registry/resolver, audit logger, outbox+dispatcher, job runner, notification dispatcher, document/file service, webhook service, integration registry, API/error/validation/pagination helpers, tx manager + RLS helpers, migration/seed loaders, o11y, testkit. |
| **Extension layer (L2)** | `Module` + `Context` contracts in public `wowapi/module`, registries the modules write into, seed schema, boundary lint rules. |
| **Domain modules (L3)** | product tables, product services, product workflows *as definitions*, product rule points, product roles/permissions, product API routes, product reports — **in the consuming product's repository**, never in wowapi. |
| **Contract fixtures** | `wowapi/internal/testmodules/*`: private neutral modules (`requests`) used by the framework's own contract suite; never part of the public API. |
| **Examples** | `wowapi/examples/*`: standalone sample product repos/apps, preferably with their own `go.mod`; non-contractual and never imported by `kernel`, `module`, or `app`. |
| **Adapters (L0)** | postgres, s3, smtp/sms/push providers, OIDC verifier, secret providers, malware-scan hook impl — public `wowapi/adapters/*`. |

### How domain logic is kept out of the kernel
1. Kernel entities reference domain things only as `resource_type` + `resource_id` (registry pattern).
2. Kernel vocabulary is closed: adding a noun to the kernel requires an ADR proving ≥2 unrelated domains need it.
3. `wowapi lint boundaries` greps kernel + extension packages for a denylist (building, wing, flat, member,
   committee, AGM, maintenance, defaulter, parking, visitor, conveyance, redevelopment, election…) and
   fails CI. Crude, effective.
4. Contract tests in the testkit instantiate the kernel with the neutral fixture module
   (`wowapi/internal/testmodules/requests`) — if the kernel only works with a specific module, tests break.
5. Product code physically cannot leak in: it lives in other repositories that *depend on* wowapi,
   and dependencies point one way.

### Future extraction path (strangler, later only)
A module becomes a service when: its write load needs independent scaling, or team ownership demands it.
Mechanics: its tables move behind its ports; sync calls become HTTP/gRPC on the same port interfaces;
async stays on events (outbox already in place); an anti-corruption layer wraps the remote calls.
Nothing else in this design needs to change — that's the payoff of the boundary rules.
