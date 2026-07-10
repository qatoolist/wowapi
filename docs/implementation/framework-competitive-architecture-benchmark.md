<!-- markdownlint-disable MD013 -->

# Competitive Framework Analysis And Architectural Benchmarking

Date: 2026-07-10

Repository reviewed: `/Users/qatoolist/go_home/src/github.com/qatoolist/wowapi`

Branch reviewed: `feat/wowsociety-framework-gaps`

Companion report: `docs/implementation/wowsociety-framework-gap-design-review.md`

## Objective

This document benchmarks `wowapi` against mature backend frameworks to turn the current post-`wowsociety` gaps into a better engineering program. The goal is not to copy another ecosystem's syntax or project layout. The goal is to identify the architectural shape of mature framework features: source of truth, lifecycle, safety defaults, performance discipline, operational controls, and developer workflow.

Frameworks reviewed:

- Laravel: developer workflow, container, middleware, localization-style feature completeness.
- Spring Boot and Spring Security: enterprise dependency scopes, filter chain, execution pools and virtual-thread posture.
- Gin: Go router internals and route-dispatch performance patterns.
- FastAPI: ASGI/async model, dependency injection, request validation workflow.
- Django: middleware model, security defaults, database connection lifecycle.
- Axum/Tower: typed state, extractor-driven request composition, service/layer model.

## Current `wowapi` Baseline

The current branch is not a naive implementation. It already has several strong framework instincts:

- Route registration requires metadata: every route is explicitly `Public` or permission-gated, and invalid/duplicate routes fail boot through `kernel/httpx/router.go`.
- Runtime route access is fail-fast: `kernel/httpx/authz_gate.go` authenticates, binds tenant and actor into context, evaluates authorization in a tenant transaction, applies step-up, and only then reaches the handler.
- Edge middleware exists centrally: secure headers, exact-match CORS, request body caps, timeout, request ID, panic recovery, rate limiting, locale binding, tracing, RED metrics, and access logging are wired by the scaffold.
- JSON input is strict: unknown fields, oversized bodies, empty/null bodies, trailing JSON, and parser internals are handled centrally in `kernel/httpx/decode.go`.
- Tenant data is sealed behind `database.TxManager`; tenant transactions bind `app.tenant_id` with `SET LOCAL`, reassert role, enforce RLS, and set server-side statement timeouts.
- Boot validates the module graph, RLS posture, registries, seed ownership, rule registry, document storage wiring, i18n ownership, and route-permission existence before serving.
- Jobs use a bounded worker pool with `FOR UPDATE SKIP LOCKED`, per-job timeouts, crash reclaim, and drain budget.
- The kernel is composed explicitly in Go rather than through a runtime global service locator.

The gaps are therefore more specific: several features were implemented as tactical APIs but not yet as productized framework subsystems with source ownership, loader contracts, performance benchmarks, lifecycle scopes, operational budgets, and scaffolded workflows.

## Comparative Matrix

| Subsystem | `wowapi` current baseline | Mature-framework benchmark | Gap / risk | RFC action |
| --- | --- | --- | --- | --- |
| Routing and request dispatch | Route registry validates metadata, then `SecureHandler` registers `method + pattern` on Go `net/http.ServeMux`. | Gin uses a priority-ordered radix tree with path segments, indices, wildcard state, and handler chains; Axum/Tower composes typed routers and services. | No route-dispatch benchmark or high-cardinality routing budget exists. Current `ServeMux` may be fine, but the framework has not proven it under product-scale route counts and param-heavy paths. | Add router benchmark gates before replacing anything. Keep `ServeMux` unless data shows route lookup or allocations matter. If needed, add a strategy interface backed by a radix tree. |
| Route safety | Mandatory `RouteMeta`; unknown route permissions fail boot; runtime gate is deny-by-default. | Spring Security's `FilterChainProxy` selects a security chain before application logic; Laravel/Spring/FastAPI let security concerns be attached to route/controller/dependency layers. | Strong baseline. Main missing piece is a generated manifest that proves middleware and authz phases are present in all binaries and products. | Generate and test a route/middleware/security manifest from the booted graph. |
| Middleware and context propagation | Standard Go middleware `func(http.Handler) http.Handler`; request values travel in `context.Context`; scaffold orders edge middleware intentionally. | Django documents onion-style middleware; Spring Security models a filter chain; Tower/Axum uses `Service`/`Layer`; Laravel resolves middleware through its container. | The pattern is idiomatic, but ordering is encoded in scaffold comments and tests, not a declarative phase model. Worker/migrate lifecycle middleware equivalents are uneven for cross-cutting features such as i18n loading. | Add middleware phases: edge, locale, telemetry, metering, authn/authz, body, timeout. Assert order in scaffold tests and expose a boot manifest. |
| Concurrency and I/O | Go `net/http` gives one goroutine per request; pgx pools cap DB work; jobs are bounded by pool size; timeouts exist. | Spring Boot exposes executor/scheduler builders and virtual-thread switches; FastAPI uses async coroutines for I/O; Django documents persistent-connection tradeoffs; Axum rides Tokio's async runtime. | No single framework-level concurrency budget ties HTTP inflight, DB pool max, worker pool, rate limits, job runner concurrency, and platform pool reservations. Local settings can fight each other. | Add `config.ConcurrencyProfile` and fail-closed validation for pool/inflight/worker relationships. Add backpressure middleware that returns 503/429 before exhausting DB pools. |
| Resource pooling | `config.DB.Pool` has `max_conns` and query timeout; API/worker/migrate use the config path in generated binaries. | Django explicitly documents persistent connection lifecycle and warns that ASGI should use backend pooling; Spring exposes executor/pool customization. | Pool sizing is local to DB config. There is no product-level capacity model for replicas, DB max connections, reserved migrate/platform capacity, and background workers. | Add a capacity calculator and `wowapi config capacity` lint: replicas * runtime pools + platform pools + worker pools must fit DB limits with reserve. |
| Security defaults | Secure headers, deny-by-default auth, strict JSON, RLS guard, tenant transaction sealing, CORS allowlist, rate limit, audit on sensitive/denied paths. | Django provides broad security guidance/default middleware; Spring Security places authn/authz in a central filter chain; Laravel ships CSRF/session/auth features for web workflows. | Good API-security baseline. Missing framework posture for cookie/session CSRF, outbound HTTP/SSRF-safe clients, HTML/template XSS if a product serves web pages, and security-profile selection by API vs web app. | Add `SecurityProfile`: API bearer default, optional browser/session mode with CSRF/SameSite, safe outbound client with allowlist/private-IP blocking, and HTML CSP templates. |
| Auth pipeline | Product supplies authenticators; framework gates every non-public route; composite auth hard-fails on non-authentication errors. | Spring Security's selected `SecurityFilterChain` can differ by request; FastAPI dependencies can enforce security before handler execution. | Good current path. Step-up remains boolean and hard-coded; AMR freshness cannot be evaluated because claims do not carry `auth_time`. | Move step-up to policy registry with accepted AMR, optional max age, challenge hints, and token freshness support. |
| Dependency injection / IoC | Explicit construction in `kernel.New` and `app.Boot`; module context exposes typed capabilities; no global service locator. | Laravel and Spring provide containers with singleton/request/scoped/transient lifecycles; FastAPI builds a dependency tree per path operation; Axum uses typed state/extractors. | Explicit Go wiring is safe but growing manually. `moduleContext` and `Kernel` have many fields, no generated dependency graph, and no first-class lifecycle scopes. | Add a static provider manifest/codegen path: compile-time descriptors for process, request, tenant-tx, job, and migrate scopes. Validate cycles and scope leaks at boot/CI. |
| Validation and schema | HTTP validation is centralized; rules use a focused hand-rolled schema subset. | FastAPI/Pydantic turns type models into runtime validation and OpenAPI; Spring has Bean Validation; Laravel has a mature validator workflow. | Rules still overclaim "JSON Schema" in public docs/comments and fail open on unknown type keywords. Defaults are not validated at registration; resolver returns stored/default values without validation. | Either adopt a real JSON Schema library or rename to strict `RuleValueSchema` and reject unknown keywords/types at registration and sync. |
| Localization / i18n | Locale negotiation, context binding, problem/validation lookup, and in-memory catalog exist. | Laravel treats localization as a complete workflow: file source conventions, fallback, placeholders/plurals, publish/override, runtime locale. | Critical productization gap: no first-class YAML/JSON/.go loaders, no framework YAML defaults, no publish path, no product override chain, no DB overlay contract, no coverage validation. | Reopen as `GAP-001B`: i18n source/loading/tooling. Embedded framework YAML defaults first; product YAML/JSON/.go overrides next; DB overlay last. |
| Lifecycle / generated operations | Generated migrate runs migrations, seed sync, and rules sync using composed product config. Standalone `wowapi seed sync` uses `DATABASE_URL` and default DB config. | Mature frameworks tend to make operational commands use the same app/container/config lifecycle as runtime commands. | Standalone seed sync can drift from product config and does not run `rules.SyncDefinitions`. It is an escape hatch but is not documented as one. | Generate product-local lifecycle commands or make CLI delegate into product config, and document low-level escape hatches clearly. |

## Low-Level Deep Dives

### Routing: Metadata Registry Plus Dispatch Strategy

`wowapi` currently uses a strong registration model and a standard dispatch model:

```text
module.Register
  -> ctx.Routes().Handle(method, pattern, RouteMeta, handler)
  -> app.Boot validates RouteMeta and route permissions
  -> httpx.SecureHandler registers each route on net/http.ServeMux
  -> request dispatch runs gateRoute before handler
```

Gin's routing source shows a different optimization point: a compact tree node stores the path fragment, child indices, wildcard state, priority, handlers, and full path. Priority increments move hot children forward. That design targets fast lookup and lower allocation in high-route-count APIs.

The decision for `wowapi` should be benchmark-driven:

```go
func BenchmarkDispatch(b *testing.B) {
    // Build N routes with a mix of static, path-param, and wildcard patterns.
    // Exercise steady-state dispatch through the full SecureHandler chain.
    // Track ns/op, allocs/op, and p99 under parallel load.
}
```

Acceptance threshold:

- Keep `ServeMux` if dispatch overhead stays below the middleware/authz/DB budget for realistic route counts.
- Add a radix-tree router only if it materially reduces request overhead or allocations and preserves the current `RouteMeta` safety contract.
- Do not let a performance router weaken boot validation, permission manifest generation, or OpenAPI/seed sync integration.

### Middleware: From Ordered Slice To Validated Phases

The current scaffold applies middleware in an intentionally ordered slice:

```text
RequestID
Recover
SecureHeaders
CORS
Locale
Trace
Metrics
AccessLog
RateLimit
BodyLimit
Timeout
SecureHandler/AuthzGate
Handler
```

Django's onion model, Spring Security's filter chain, and Tower's layer model all have the same architectural idea: each cross-cutting concern wraps the next concern, and the framework owns the sequence.

The current risk is not the Go implementation; it is that the order is scaffolded by convention. A better framework model is:

```go
type MiddlewarePhase string

const (
    PhaseCorrelation MiddlewarePhase = "correlation"
    PhaseRecovery    MiddlewarePhase = "recovery"
    PhaseEdge        MiddlewarePhase = "edge"
    PhaseLocale      MiddlewarePhase = "locale"
    PhaseTelemetry   MiddlewarePhase = "telemetry"
    PhaseMetering    MiddlewarePhase = "metering"
    PhaseLimits      MiddlewarePhase = "limits"
    PhaseAuth        MiddlewarePhase = "auth"
)

type MiddlewareSpec struct {
    Name  string
    Phase MiddlewarePhase
    Build func(BootedRuntime) httpx.Middleware
}
```

The scaffold can still emit plain Go, but tests should assert the phase manifest. That catches missing locale loaders, misplaced CORS, or product edits that put body parsing before rate limiting.

### Concurrency: Capacity Budget Instead Of Independent Knobs

The mature-framework lesson is not "make Go async." Go already gives request goroutines and cancellation. The lesson is to make concurrency explicit and bounded across every shared resource.

Current independent knobs:

- HTTP `request_timeout`, `read_header_timeout`, `max_body_bytes`, rate limit.
- DB `max_conns` and `query_timeout`.
- Job runner `poolSize`, `jobTimeout`, `drainTimeout`, `reclaimTimeout`.
- API and worker each create runtime and platform pools.

Missing framework model:

```yaml
concurrency:
  http_max_in_flight: 256
  db_reserved_for_admin: 4
  worker_max_jobs: 10
  platform_max_in_flight: 32
  overload:
    api_status: 503
    retry_after: 2s
```

Validation should reason about deployment shape:

```text
replicas * (runtime_pool_max + platform_pool_max)
  + migrate_pool_max
  + reserved_admin_connections
  <= database_max_connections
```

The framework should fail config validation when the declared product shape can exhaust the database before rate limits or backpressure engage.

### Security: Profiles, Not Handler-Level Advice

`wowapi` already has strong API defaults. The next enterprise gap is to make security posture explicit by product surface:

```text
SecurityProfileAPI
  bearer/API-key auth
  CSRF disabled by contract
  strict JSON
  CORS allowlist
  RLS guard
  safe outbound client optional

SecurityProfileBrowser
  cookie/session auth
  SameSite defaults
  CSRF tokens
  CSP profile for HTML
  XSS template guidance
  upload/content-domain isolation
```

Django and Spring Security are useful references here because they treat security as a framework layer. Product handlers should not each remember CSRF, SSRF, XSS, or header posture.

Immediate `wowapi` additions:

- `kernel/httpclient` safe outbound client with DNS/IP blocking for loopback, link-local, RFC1918, and metadata endpoints unless allowlisted.
- CSRF middleware available only when a browser/session auth profile is enabled.
- Security profile validation in `config validate` and generated scaffold tests.

### DI / IoC: Static Lifecycle Graph For Go

Laravel and Spring use containers with lifecycle scopes. FastAPI builds dependency trees. `wowapi` should not introduce a reflection-heavy runtime container, but it does need lifecycle discipline as the kernel grows.

Recommended Go-native shape:

```go
type Scope string

const (
    ScopeProcess  Scope = "process"
    ScopeRequest  Scope = "request"
    ScopeTenantTx Scope = "tenant_tx"
    ScopeJob      Scope = "job"
    ScopeMigrate  Scope = "migrate"
)

type ProviderDescriptor struct {
    Provides string
    Requires []string
    Scope    Scope
}
```

A generated manifest can validate:

- no process-scoped service depends on request-scoped state;
- no module receives raw pools when it should receive `TxManager`;
- no tenant-scoped service escapes its transaction;
- no migrate-only service is wired into API runtime;
- all declared dependencies have providers.

This keeps the current explicit Go style while adding the lifecycle guarantees mature containers provide.

## Actionable RFC

### RFC-COMP-001: Enterprise Framework Architecture Uplift

Goal: convert tactical gap closures into framework-quality subsystems with explicit source, lifecycle, safety defaults, benchmarks, and generated product workflows.

### Phase P0: Correctness And Truthfulness Gates

Priority: immediate before broad merge/announcement.

1. Reopen i18n as `GAP-001B`.
   - Add loader contract.
   - Ship embedded framework YAML defaults.
   - Add product YAML and JSON loaders.
   - Add first-class Go `.go` catalog bundle support.
   - Add optional DB overlay last.
   - Add publish/scaffold path and `wowapi i18n validate`.
   - Migrate `wowsociety` off manual Go-map registration once the loader exists.

2. Fix rules schema honesty and fail-closed behavior.
   - Either use a real JSON Schema implementation or rename the contract to a strict custom schema.
   - Reject unknown keywords and unknown type values during registration/sync.
   - Validate defaults against schemas at `Register` or boot.
   - Validate resolved/stored values before returning or document the read path as trusted only after write-time validation.

3. Align lifecycle commands.
   - Make standalone seed sync clearly documented as a low-level escape hatch, or generate a product-local command that uses `appcfg.Load`.
   - Ensure any production lifecycle path that syncs seeds also syncs `rules.SyncDefinitions`.

4. Add benchmark gates before router changes.
   - Dispatch benchmark with high route counts.
   - Middleware-chain allocation benchmark.
   - Authz gate benchmark with cached and uncached store paths.
   - JSON decode/body-limit benchmark.

### Phase P1: Productized Operational Controls

Priority: next engineering cycle after P0.

1. Add `ConcurrencyProfile`.
   - HTTP in-flight cap and overload response.
   - Worker pool cap tied to DB/platform pool budget.
   - Capacity validation across replicas.
   - Metrics for rejected-overload, DB pool wait, worker saturation, and queue lag.

2. Add static provider/lifecycle manifest.
   - Generate descriptors from kernel/app/module wiring.
   - Validate scope leaks and missing providers in CI.
   - Expose manifest in `wowapi doctor`.

3. Add `SecurityProfile`.
   - API profile remains bearer/API-key, CSRF-free by contract.
   - Browser/session profile wires CSRF/SameSite/CSP.
   - Safe outbound HTTP client prevents SSRF by default.
   - Scaffold tests prove the selected profile is actually wired.

4. Step-up policy.
   - Move hard-coded strong AMR values to policy/config.
   - Add `auth_time`/freshness support before `MaxAge` is exposed.
   - Keep `step_up: true` as shorthand for default policy.

### Phase P2: Performance Strategies Only If Data Requires Them

Priority: defer until P0/P1 prove actual bottlenecks.

1. Optional radix-tree router strategy.
   - Implement only if benchmarks show `ServeMux` route dispatch is a real cost.
   - Preserve current `RouteMeta`, OpenAPI, permission, and boot validation behavior.

2. Richer validation/OpenAPI integration.
   - Use typed schemas to feed validation, OpenAPI, and codegen from the same source.
   - Avoid duplicating rule schema, DTO validation, and OpenAPI models by hand.

3. Hot-reloadable operational sources.
   - DB-backed i18n/rules overlays can reload, but only after immutability, validation, metrics, and cache-invalidation semantics are defined.

## Acceptance Tests For Engineering

- `wowapi init` renders API, worker, and migrate binaries that load the same configured i18n catalog sources before boot completes.
- A product-local `locales/mr/kernel.yaml` overrides an embedded framework `kernel.*` key for Marathi while English still falls back to the embedded framework default.
- Product-local YAML, JSON, and Go catalog bundles all load through the same framework lifecycle.
- Optional DB i18n overlays win last and are visible only after validation.
- `wowapi i18n validate` fails on missing locale coverage, duplicate keys in the same layer, placeholder mismatch, and unauthorized namespace writes.
- Rules reject unknown schema keywords and unknown `type` values before a product can write rule values.
- Rule defaults are validated at registration or boot.
- Standalone seed sync is either documented as an escape hatch or replaced by a generated product-aware lifecycle command; production lifecycle syncs both seeds and rule definitions.
- Capacity validation fails when declared replicas and pools exceed the database budget.
- Middleware phase tests prove secure headers and CORS wrap all short-circuiting responses.
- Safe outbound HTTP client blocks loopback/private/metadata addresses unless explicitly allowlisted.
- Route dispatch benchmarks and allocation budgets are checked before any router replacement is accepted.

## Engineering Backlog Extract

1. `GAP-001B`: i18n source/loading/tooling and `wowsociety` migration to framework loaders.
2. Rules schema fail-closed plus public contract correction.
3. Seed/rule lifecycle CLI alignment.
4. Concurrency profile, capacity validator, overload/backpressure middleware.
5. Static provider/lifecycle manifest for kernel/app/module wiring.
6. Security profiles, including SSRF-safe outbound client and optional browser CSRF posture.
7. Step-up policy registry with `auth_time` freshness support.
8. Router strategy benchmark; implement radix tree only if benchmark data justifies it.

## Primary Sources Reviewed

- Laravel Service Container: <https://laravel.com/docs/13.x/container>
- Laravel Middleware: <https://laravel.com/docs/13.x/middleware>
- Laravel Localization: <https://laravel.com/docs/13.x/localization>
- Spring Security Servlet Architecture: <https://docs.spring.io/spring-security/reference/servlet/architecture.html>
- Spring Boot Task Execution and Scheduling: <https://docs.spring.io/spring-boot/reference/features/task-execution-and-scheduling.html>
- Spring Framework Bean Scopes: <https://docs.spring.io/spring-framework/reference/core/beans/factory-scopes.html>
- Gin router tree source: <https://github.com/gin-gonic/gin/blob/master/tree.go>
- FastAPI async/concurrency: <https://fastapi.tiangolo.com/async/>
- FastAPI dependencies: <https://fastapi.tiangolo.com/tutorial/dependencies/>
- Django middleware: <https://docs.djangoproject.com/en/6.0/topics/http/middleware/>
- Django security: <https://docs.djangoproject.com/en/6.0/topics/security/>
- Django database persistent connections: <https://docs.djangoproject.com/en/6.0/ref/databases/#persistent-connections>
- Axum crate documentation: <https://docs.rs/axum/latest/axum/>
- Go `net/http`: <https://pkg.go.dev/net/http>
