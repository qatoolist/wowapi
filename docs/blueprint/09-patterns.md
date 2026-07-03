# 09 — Pattern Catalog, Anti-Patterns, Decision Matrix, Recommended Stack

Every pattern below states where it's used, where it's not, and the Go approach. Patterns already
specified in detail elsewhere reference their section instead of repeating it.

## 1. Architectural
- **Modular monolith + compile-time plugin modules + edge hexagonal** — the core (00 §3). Not used: runtime plugins, microservices, BFF (one API serves all clients until a client proves divergent needs), mediator (hides call graphs).
- **Shared kernel** — deliberately: the kernel *is* the shared kernel, kept domain-blind by the boundary rules. Anti-pattern avoided: shared kernel accreting domain nouns (ADR gate + lint).
- **Anti-corruption layer** — at integration adapters only (07 §6): provider types die at the adapter. Not between internal modules (ports are enough; ACL there is ceremony).
- **Strangler** — dormant seam for future extraction (00 §5). Do nothing now beyond keeping boundaries clean.
- **CQRS-lite** — command/query object separation, shared DB (05 §3). Full CQRS/read stores: only per proven reporting pain. Event sourcing: avoided; audit log + outbox give the compliance trail without rebuild-from-events complexity.

## 2. Domain & modeling
- **Entity / Aggregate** — module aggregates with invariant methods (`Request.Approve`). Use for anything with state transitions; skip aggregate ceremony for pure reference data (a row is a row).
- **Value object** — `Money`, `TimeRange`, `ResourceRef`, `ActorRef` (04 §3): immutable, compared by value. Don't wrap every string.
- **Domain service vs application service** — app services orchestrate (05 §3); a separate domain service only when logic spans aggregates AND has no natural home (rare — resist).
- **Anemic model caution** — simple CRUD resources may be legitimately anemic (fine); workflow/financial aggregates must carry behavior or invariants scatter into services.
- **Repository** — per aggregate, SQL inside (05 §2). NOT generic `Repository[T]` hiding SQL — Postgres features (RLS, FOR UPDATE, exclusion constraints) are the design.
- **Domain event** — outbox rows (07 §7). Not for intra-service control flow (call the function).
- **Specification** — only as the filter-allowlist builders (05 §2). A full composable spec DSL in Go reads worse than SQL — avoided.
- **Policy pattern** — authz policy conditions (01 §3): data-driven, closed operator set. Not a general expression VM in v1.
- **State machine** — workflow engine (02) for configurable flows; plain status enums + guarded transition funcs for simple lifecycles. Don't run trivial lifecycles through the workflow engine (overhead without benefit).
- **Temporal modeling** — `valid_from/to` where "as of" is a real question (01, 03). Not everywhere: temporal tables tax every query.
- **Resource registry / relationship graph / actor-capacity** — the kernel's three novel primitives (01). Registry: use for anything needing kernel services; don't mirror rows that never need them. Relationship graph: authz + structural edges; not a general graph DB (no traversal queries beyond 1–2 hops). Metadata JSONB extension bag — pressure valve only (04 §3).

## 3. Persistence & transactions
Specified in 03/05: repository, tx-manager-as-unit-of-work, optimistic locking, idempotency keys,
outbox + inbox, append-only audit/event log, status lifecycle over delete, temporal validity, RLS,
keyset pagination, expand-contract migrations. Additional judgments:
- **Transaction script** — fine for simple module operations; not everything needs an aggregate dance.
- **Read model / materialized view** — later only, per specific slow report; refresh via event handler.
- **Inbox** — mandatory for event handlers (processed_events); skip for jobs that are naturally idempotent (sweeps).

## 4. API & integration
REST resource pattern, DTO separation, problem-details, envelope, webhook receiver + signature,
retry/backoff, circuit breaker, bulk-as-async-job, LRO, presigned transfer, provider adapter,
versioning — all specified in 04/07. Judgments: retries only on idempotent operations; breakers only
on *outbound* calls (never internal); backward compatibility = additive-only within a version +
contract tests on DTO JSON.

## 5. Security
Secure/deny by default, defense in depth (middleware + service check + RLS), least privilege (scoped
time-boxed assignments), tenant isolation, record-level authz, policy authz, capacity selection,
break-glass, audited impersonation, revocation hook, route permission metadata, sensitive-action
audit, upload pipeline, secret provider, rate limiting, replay protection, boundary validation, safe
errors — all specified in 01/07 with enforcement points. The pattern-level rule: **every security
pattern has a structural enforcement (type system, middleware, DB) plus a test in testkit** — policy
by convention alone is a bug.

## 6. Concurrency & async
Worker pool, bounded concurrency, backpressure, retry policy, DLQ, scheduled jobs, outbox relay,
idempotent consumer, progress tracking, context cancellation, graceful shutdown — specified in 07 §3.
Judgments: fan-out/fan-in only inside workers over independent items (`errgroup.WithContext` +
`SetLimit`); singleflight for shared cache fills only; advisory locks for per-aggregate handler
serialization and singleton crons; **parallelism never inside a business tx, never unbounded, never
fire-and-forget**.

## 7. Go-specific
Composition/embedding (04 §3), consumer-side small interfaces, constructor injection + composition
root (06 §3), context propagation, error wrapping (04 §5), table-driven tests + fakes (08), no global
state, no reflection containers. Additional:
- **Functional options** — only for the kernel server/runner constructors with genuinely optional
  knobs; module constructors take plain required args (options hide requiredness).
- **Generics** — where they remove real duplication with one type parameter: `APIResponse[T]`,
  `CursorPage[T]`, `DecodeJSON[T]`, `Statused[S]`, typed job workers. Not for repositories/services.
- **Code generation** — sqlc, moq, module scaffolds (08 §3); never for logic.

## 8. Developer-experience patterns
Starter template, registration contract, seed/migration/permission/workflow/rule/event/job
registration, testkit fixtures, handler/repo/response helpers, codegen, make targets, boundary
linting, OpenAPI generation — specified in 05/06/08. Net effect: a new module = `make new-module`,
fill in domain + service + SQL, seeds declare its catalog — no kernel edits, boilerplate ≈ zero,
logic 100% visible.

## 9. Anti-patterns → safer alternative (explicit list)

| Anti-pattern | Safer framework alternative |
|---|---|
| God service / kernel façade object | capability-scoped interfaces on ModuleContext; services own one aggregate family |
| Fat controller | 10–25-line handlers + httpx helpers; logic in services |
| Harmful anemic model | invariant methods on workflow/financial aggregates |
| Generic repository hiding SQL | per-aggregate repos + sqlc; SQL is a feature |
| Universal BaseModel | small opt-in embeds (04 §3) |
| Service locator / hidden globals | composition root; ModuleContext passed explicitly at Register only |
| Circular module imports | DependsOn topo-sort + ports + lint |
| Cross-module SQL joins | ports or exported read views; lint |
| Tenant/audit bypass | TenantDB is the only door; audit writers ride the tx; RLS backstop |
| Route without permission metadata | registration refuses; startup validation |
| Business logic in middleware | middleware = context + guards only; logic in services |
| Workflow logic hard-coded in handlers | workflow definitions + runtime (02) |
| Rule constants in code | rule points + resolver (02); lint for magic business numbers |
| Unbounded goroutines | jobs framework; `go` keyword lint outside kernel async packages |
| Reflection-heavy runtime magic | codegen at build time; explicit wiring |
| Premature microservices / event sourcing / CQRS-everywhere / low-code CRUD engine | modular monolith; audit+outbox; CQRS-lite; codegen scaffolds with owned code |

## 10. Decision matrix

| Pattern | Core | Modules | Later only | Avoid | Note |
|---|:-:|:-:|:-:|:-:|---|
| Modular monolith | ✅ | — | | | extraction seams preserved |
| Hexagonal (edges) | ✅ | | | | adapters only |
| Full clean-arch layering | | | | ✅ | dependency rule yes, ceremony no |
| Vertical slice | | ✅ | | | module structure |
| Outbox / inbox | ✅ | use | | | atomicity spine |
| Event sourcing | | | | ✅ | audit log suffices |
| CQRS (separate stores) | | | ✅ | | per proven report |
| Mediator | | | | ✅ | |
| Aggregate w/ behavior | ✅ | ✅ | | | where transitions exist |
| Specification DSL | | | | ✅ | allowlist builders instead |
| State machine (engine) | ✅ | definitions | | | simple enums stay enums |
| Temporal validity | ✅ targeted | targeted | | | not universal |
| Optimistic locking | ✅ | ✅ | | | editable aggregates |
| Generic repository | | | | ✅ | |
| Keyset pagination | ✅ | ✅ | | | offset admin-only |
| RLS | ✅ | ✅ | | | FORCED, tested |
| Circuit breaker | ✅ outbound | | | | never internal |
| Saga/process manager | | | ✅ | | workflow engine covers v1 |
| Worker pool / DLQ | ✅ | use | | | River |
| Functional options | kernel ctors | | | modules | |
| Generics | primitives | | | services/repos | |
| Wire DI | | | optional | | manual root first |
| Runtime DI container | | | | ✅ | |
| Low-code CRUD runtime | | | | ✅ | codegen instead |

## 11. Recommended pattern stack (final)

- **Core architecture:** modular monolith; kernel + module SDK; hexagonal at edges.
- **Module boundary:** vertical slices; ports for cross-module; lint-enforced imports; DependsOn graph.
- **Data access:** per-aggregate repositories; sqlc/pgx; allowlisted dynamic filters; keyset pagination.
- **Transaction:** TxManager-as-UoW; one tx per command; `SET LOCAL` tenant+actor; optimistic locking.
- **Authorization:** deny-default layered RBAC → ReBAC → ABAC policies; capacity-based actors; route metadata.
- **Workflow:** custom Postgres declarative engine; versioned definitions; tenant overrides.
- **Rules:** typed rule points; JSON-Schema values; scope-resolution with temporal versions; flags on top.
- **Events:** transactional outbox → in-process dispatcher → idempotent inbox consumers.
- **Audit:** append-only, partitioned, same-tx writers, DB-enforced immutability.
- **Background jobs:** River worker pools; retry+backoff+DLQ; tenant-aware payloads; progress tracking.
- **Integrations:** provider adapters + ACL; signed idempotent webhooks; circuit breakers.
- **API:** REST + problem details + envelopes; ETag/If-Match; Idempotency-Key; cursor pagination; OpenAPI from fragments.
- **Testing:** testcontainers integration-first; contract suite per module; RLS/authz/audit assertions; fakes at process boundaries only.
- **Developer scaffolding:** module template + codegen CLI + seed registration + boundary lint + make targets.
