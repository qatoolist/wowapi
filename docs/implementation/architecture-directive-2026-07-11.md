# wowapi Premier-Framework Architecture Directive

- **Review date:** 2026-07-11
- **Reviewed revision:** `d3c2640dbe1a0fe27e826cdf053945c4f49bc034`
- **Scope:** framework architecture, module DSL, security, tenancy, persistence, asynchronous execution, performance, observability, developer experience, testing, release engineering, operations, compliance, and compatibility
- **Status:** implementation directive; no production code was changed by this review

## 1. Executive directive

wowapi already has a stronger foundation than most pre-product backend frameworks: forced PostgreSQL RLS, fail-closed transaction scoping, deny-by-default route authorization, strict configuration loading, transactional outbox primitives, jobs and scheduling, migration reversibility, structured errors, SSRF protection, signed cursors, audit chaining, observability adapters, and an unusually broad test suite. Those strengths are real and should be preserved.

It is not yet safe to call the repository a premier, works-out-of-the-box framework. Green CI currently coexists with framework-level integrity defects that ordinary package tests do not model:

1. The module registration surface is a large shared mutable service locator. A module receives raw shared registries, can claim another module's identity on several registration APIs, can retain those registries after boot, and can mutate most of them after validation.

2. Inter-module ports are stringly typed `any` values stored in a last-writer-wins map. The documented boot validation and ownership rule do not exist in the implementation.

3. The authentication boundary accepts tenant, break-glass, and impersonation state from JWT claims more readily than the repository's own normative tenancy model permits.

4. Workflow override can skip authorization entirely when the runtime is constructed with a nil evaluator.

5. The in-memory token-bucket limiter's eviction algorithm cannot evict one-shot identities. After the map reaches 10,000 entries, each new identity triggers a full scan that deletes nothing. This is both an unbounded-memory defect and an attacker-controlled CPU-amplification path.

6. Notification and webhook delivery perform network I/O while database transactions and row locks remain open. Job completion lacks a fencing token. Bulk processing is documented as replica-safe while the implementation explicitly assumes one processor. These are reliability boundaries, not tuning details.

7. Tenant-local foreign keys generally reference only globally unique `id` columns. RLS protects direct access but the database does not prove that parent and child rows carry the same `tenant_id`.

8. A source-built CLI can generate a syntactically valid but unresolvable or unrelated `v0.0.0` dependency, the CRUD generator emits a permission verb rejected by the authorization registry, most CRUD handlers are TODOs, and release/version documentation simultaneously claims both pre-1.0 instability and post-1.0 stability.

9. A `v*` tag can directly start a signing and publishing workflow without proving that the exact tagged commit passed the authoritative CI/security/compatibility gates. Several security scanners are deliberately informational or skipped in the repository's current private configuration.

The architectural response is not to add more unrelated features. Freeze breadth-first feature growth and execute the foundation program in this directive. The defining target is:

> A module declares one immutable, ownership-bound, typed application model. The framework compiles that model once, proves all cross-cutting invariants, derives every runtime and documentation projection from it, seals it, and then serves requests and workers through bounded, observable, fail-closed runtime components.

The current router replacement (B11), standalone schema-unification project (B12), and hot-reload overlays (B13) should remain parked. The router benchmark is flat; a separate generator is not the right first move while the module declaration model is fragmented; and mutable i18n would work against the required seal-at-boot invariant. Schema derivation should later emerge from the typed application model, not from a parallel metadata system.

## 2. Review method, evidence, and limits

### 2.1 Repository coverage

The existing Graphify corpus was checked for freshness before use. Its report covers 611 files and approximately 696,733 words, with 5,062 nodes, 10,425 edges, and 366 communities (`graphify-out/GRAPH_REPORT.md:3-8`). The graph was used for navigation and relationship discovery; source files, tests, migrations, templates, and workflows were then read directly before a finding was accepted.

Mechanical gathering was split into bounded audits for architecture/DSL, security, persistence/reliability, performance, CI/release quality, and developer experience/operations. Findings were then cross-checked against live source by the lead review rather than copied from existing decision documents.

### 2.2 Executable evidence

The following gates passed during this review against the reviewed revision. Revision-addressed commands, environment, exits, skip caveats, and benchmark output are recorded in `docs/implementation/evidence/architecture-review-2026-07-11/command-log.md` and `evidence.json`:

- `scripts/graphify_refresh.sh check`
- `make ci`
- `make ci-container`
- `make lint-new`
- `miscellaneous/check_migrations.sh` (30 registered/contiguous migrations with Up/Down markers)
- `miscellaneous/check_test_skips.sh` (22 explicit skip sites inventoried)
- uncached `TestIntegrationMigrationsReversible` through the DB-backed container test run (disposable down/up reconstruction)
- uncached MinIO-backed S3 contract and document round-trip tests with `WOWAPI_REQUIRE_S3=1`
- `go build ./...`
- `go vet ./...`
- Markdown/JSON formatting, JSON parse, source-citation resolution, and trailing-whitespace checks

The reviewed SHA identifies the source evaluated by the review; it is not a claim that these newly produced review artifacts already exist in that SHA. At finalization the artifacts remain working-tree deliverables because this review did not authorize a commit or publication workflow. They become durable repository evidence only when a later authorized change preserves them and records the artifact commit/digest.

The authoritative container gate did force DB-backed tests through `WOWAPI_REQUIRE_DB=1`, but it did not set `WOWAPI_REQUIRE_S3=1` (`Makefile:312-314`). Consequently, MinIO-backed tests may skip even though hosted CI starts MinIO (`.github/workflows/ci.yml:92-97`, `adapters/storage/s3/s3_test.go:66-77`). An explicit forced run also proved that the toolbox's `S3_ENDPOINT` variable is not the test harness's `S3_TEST_ENDPOINT` (`deployments/compose.yaml:86-97`, `adapters/storage/s3/s3_test.go:35-40`): without the latter, tests fail closed against the container-local `localhost:9000`; with `S3_TEST_ENDPOINT=minio:9000`, the complete S3 suite passed uncached. The repository contains 22 skip sites across optional integration and environment-dependent tests; a green general-purpose run is therefore not proof that every advertised adapter path executed.

The live dispatch benchmark remained effectively flat:

| Environment        |   50 routes |  500 routes | 2,000 routes |  Allocations |
| ------------------ | ----------: | ----------: | -----------: | -----------: |
| Host, 3-run median | 588.3 ns/op | 612.7 ns/op |  641.1 ns/op | 14 allocs/op |
| Toolbox container  | 1,000 ns/op | 990.3 ns/op |  933.0 ns/op | 14 allocs/op |

This is why B11 remains parked. Route count increased 40-fold while the host median moved about 9%; the small non-monotonic container variation is normal benchmark noise, not evidence for a new routing data structure. The container row is the single budget invocation from the authoritative gate; the host row is the median of three uncached benchmark iterations.

### 2.3 Confidence labels

Every finding below uses one of these labels:

- **Verified defect:** the problematic behavior follows directly from current source or generated output.
- **Conditional security risk:** the unsafe outcome depends on deployment state that is outside this repository, such as IdP claim policy or GitHub tag protection.
- **Measured bottleneck:** a benchmark or concrete query/lock shape demonstrates the cost.
- **Scalability risk:** the source has an adverse complexity or contention shape, but a representative production trace is still required to rank its real budget share.
- **Strategic requirement:** a target capability required for the desired framework class, not a claim that the current code is broken.

### 2.4 External state not verified

This review did not and should not pretend to know:

- GitHub organization rules, tag protection, environment protection, or who can push release tags.
- The production IdP's exact rules for issuing `tenant_id`, `capacity_id`, `break_glass`, `impersonator_user_id`, `amr`, `acr`, or `auth_time` claims.
- Which external webhook providers sign timestamps and event IDs as part of their canonical payload.
- Production request distributions, tenant counts, queue depths, database plans, provider latency, object sizes, or failure rates.
- The state of a consuming product repository except where the framework's own documents explicitly refer to it.

Those unknowns reduce certainty around exploitability or priority; they do not erase the framework's obligation to make its contract explicit and fail closed.

## 3. Strengths that must survive the modernization

The modernization must be evolutionary around these assets, not a rewrite that discards them.

### 3.1 Tenancy and database boundary

- `TxManager.WithTenant` fails before using a connection when tenant context is absent and sets transaction-local role, tenant, and actor state (`kernel/database/txmanager.go:84-123`).
- Runtime and platform roles are separate, tenant-scoped tables force RLS, and boot can assert that RLS is active (`app/boot.go:120-126`).
- Modules receive `TenantDB` only inside transaction callbacks rather than a raw pool (`module/module.go:100-103`).
- The public testkit includes broad RLS isolation coverage, including a catalog-driven table matrix.

### 3.2 Authorization and HTTP safety

- Route metadata is validated at boot and enforced per request through authenticate → bind tenant/actor → evaluate (`kernel/httpx/authz_gate.go:68-141`).
- Permission evaluation is deny-by-default and has explicit RBAC, ReBAC, ABAC, step-up, and audit seams.
- Request decoding rejects unknown fields; filtering and sorting use allowlisted columns; cursors carry integrity protection.
- Outbound HTTP has DNS-resolution-time SSRF checks and revalidates redirects.

### 3.3 Configuration and secrets

- Configuration has deterministic layering, explicit environment selection, production override controls, secret references, redaction, provenance, fingerprints, and aggregated validation (`kernel/config/load.go:57-202`).
- Module config views cannot traverse to framework or sibling configuration and decode with unknown-field rejection (`kernel/config/moduleview.go:9-44`).
- Config JSON Schema is derived from the same tags as the binder and uses JSON Schema 2020-12 (`kernel/config/schema.go:11-37`).

### 3.4 Durable platform primitives

- Business events can be written transactionally with domain changes.
- Job claims are committed before worker execution, have bounded concurrency, retry/backoff, DLQ handling, and stalled-job recovery (`kernel/jobs/runner.go:302-368`).
- Audit rows have a per-tenant ordered hash chain; sequence allocation and immutable artifacts exist.
- Migrations are embedded, ordered, reversible, and tested.

### 3.5 Testing and supply-chain groundwork

- Hosted actions are SHA-pinned.
- The repository runs vet, boundary lint, lifecycle lint, unit tests, race tests, DB tests, performance budgets, secret scanning, vulnerability scanning, SBOM generation, keyless signing, and provenance attestation.
- The generated runtime exposes health, readiness, metrics, configuration fingerprinting, request metrics, and optional OpenTelemetry traces.

These are meaningful differentiators. The directive below turns them into a coherent framework contract.

## 4. Severity model and release posture

| Priority | Meaning                                                                                                                                                      | Required response                                                        |
| -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------ |
| P0       | Can violate authorization, tenant integrity, durable execution, release trust, or bounded-resource guarantees; or makes the advertised default path unusable | Block the next stability/release claim until fixed and regression-tested |
| P1       | Deep architecture/DX flaw that makes extension unsafe, drift-prone, or operationally unreliable                                                              | Complete before adding another major framework subsystem                 |
| P2       | Scalability, completeness, or operability work whose urgency depends on measured use                                                                         | Implement after P0/P1 foundations or when its explicit trigger fires     |
| Parked   | Evidence says the current implementation is adequate or the need is not demonstrated                                                                         | Keep measured; do not implement speculatively                            |

The repository may continue internal development while P0 items are open, but it should not publish a new “stable,” “production-grade,” or premier-framework claim. Existing tags are not retroactively invalidated; the point is to make the next contract truthful.

## 5. Must-refactors: verified foundation defects

### AR-01 — Replace the mutable mega-context with an ownership-bound application model (P1 architecture; P0 for untrusted extensions)

**Evidence.** `module.Context` exposes nearly every registry and runtime service through one 147-line interface (`module/module.go:66-216`). App boot constructs one shared set of registry pointers and passes them to each module (`app/boot.go:128-158`). Several registration methods accept a caller-supplied module name, for example `resource.Registry.Register(module, spec)` and `rules.Registry.Register(module, point)` (`kernel/resource/resource.go:53-72`, `kernel/rules/rules.go:106-139`). Because the raw registry is shared, a module can pass another module's name and a matching key. Permission registration does not receive or verify module identity at all (`kernel/authz/registry.go:79-99`).

Only the i18n catalog is sealed after boot (`app/boot.go:264-268`). Other registries remain mutable and a module can retain its context after `Register` returns. `resource.Registry.Specs` and `rules.Registry.Points` return their backing maps directly (`kernel/resource/resource.go:74-75`, `kernel/rules/rules.go:154-155`).

This is an integrity flaw between trusted in-process extensions, not a remote exploit by itself. It still invalidates the claimed capability/ownership model and makes whole-graph validation a one-time observation of mutable state.

**Directive.** Introduce an immutable `ApplicationModel` compiled from ownership-bound module declarations. A module identity must be captured by the registrar; registration APIs must never accept an arbitrary owner string from module code. Every collector follows `collect → validate → seal → expose read-only snapshot`. Post-seal mutation returns an explicit error or panics in developer builds; it must never silently succeed or silently no-op.

**Acceptance gates.**

- A malicious test module cannot register another module's permission, resource, rule, event, job, workflow action, provider, template, health check, migration, seed bundle, or OpenAPI operation.
- A retained registrar cannot mutate state after compilation.
- Every snapshot returns cloned/immutable data, not a backing map or slice.
- One deterministic model hash covers all declarations and is emitted at startup/readiness.
- Parallel read tests and race tests prove no runtime mutation.

### AR-02 — Replace string/`any` ports with typed keys and a compiled provider graph (P1 architecture; P0 for untrusted extensions)

**Evidence.** The public contract is `ProvidePort(name string, impl any)` and `Port(name string) (any, error)` (`module/module.go:129-134`). The implementation writes directly to a map without ownership, duplicate, dependency, or type checks (`app/context.go:337-351`). A duplicate silently replaces the prior value. The comments claim the name must be provider-prefixed and dependencies are rechecked at validation, but neither behavior exists. The lifecycle manifest is explicitly hand-maintained rather than derived from wiring (`kernel/lifecycle/lifecycle.go:1-17`, `kernel/lifecycle/manifest.go:3-10`).

**Directive.** Use typed port keys with generic free functions, because Go does not support type-parameterized methods. A target API can have this shape (this is proposed API, not current source):

```go
package port

// Registrar is an owner-bound declaration capability. The unexported method
// prevents modules from implementing or fabricating one.
type Registrar interface {
	registrarSeal()
}

type Key[T any] struct {
	ownerID string
	name    string
}

func Define[T any](r Registrar, name string) (Key[T], error)
func Provide[T any](r Registrar, key Key[T], build func(Resolver) (T, error)) error
func Require[T any](r Registrar, key Key[T]) error
func Resolve[T any](r Resolver, key Key[T]) (T, error)
```

An internal compiler factory creates each registrar with immutable owner identity and a reference to the one model catalog; module code receives only the sealed capability and cannot manufacture another owner from a string. `Define` obtains owner and catalog from that registrar, atomically registers `(owner, local-name, T)`, and rejects invalid names or duplicates before returning the key. `Provide` rejects a provider whose registrar does not own the key; `Require` records a typed cross-module dependency without transferring ownership. The internal compiler may type-erase providers into a heterogeneous graph, but it records `reflect.Type` once at compile time and never reflects on request hot paths. It must reject duplicate providers, missing requirements, type mismatches, undeclared dependencies, cycles, and invalid scope/lifetime edges before starting any process.

**Acceptance gates.**

- A consumer cannot compile a `port.Key[Notifier]` resolution as `port.Key[Ledger]`.
- Duplicate provider keys and undeclared cross-module edges fail compilation/boot with both owners named.
- API, worker, and migrate profiles are projections of one graph, not three copied wiring templates.
- The lifecycle lint consumes the generated graph; no hand-maintained duplicate manifest remains.

### AR-03 — Make one declaration authoritative and derive all projections (P1 architecture/DSL)

**Evidence.** The same feature is currently represented independently in route calls, permission seeds, resource/rule registries, workflow YAML, OpenAPI fragments, migration files, lifecycle descriptors, worker wiring, and documentation. The OpenAPI command only merges `paths` and `components.schemas`; valid top-level fragment content is silently dropped (`internal/cli/openapi_cmd.go:71-87`, `internal/cli/openapi_cmd.go:127-160`). API/worker/migrate composition is repeated in separate templates. Graph communities were used only to navigate these areas; direct source duplication, not clustering, establishes the finding.

**Directive.** A module manifest becomes the authoritative declaration. From it, deterministic build tooling derives:

- route registration and route metadata;
- permission and resource catalogs;
- request/response schema references and OpenAPI operations;
- event/job/workflow/rule identifiers and ownership;
- module dependency and provider graphs;
- migration/seed/i18n/OpenAPI bundle inventory;
- required runtime capabilities by API/worker/migrate profile;
- conformance tests, documentation tables, and a machine-readable manifest.

SQL and business behavior remain explicit Go/SQL. The model should eliminate duplicate identity and contract metadata, not become a reflection-heavy ORM or hide transactions.

### AR-04 — Eliminate configuration and boot-time silent behavior (P1 architecture)

**Evidence.** Module namespace contents are strict once a module decodes them, but a top-level `modules.<typo>` namespace for no registered module is retained as opaque data and never rejected. Boot simply looks up namespaces for known modules (`app/boot.go:136-143`). Migrations, seeds, OpenAPI fragments, health checks, and ports are stored with last-writer-wins assignments (`app/context.go:322-340`). `Catalog.Add` silently returns after freeze (`kernel/i18n/catalog.go:74-91`).

**Directive.** Reject unknown module namespaces, duplicate collectors, empty required fragments, and post-seal writes. “Optional” capabilities must be explicit in the compiled profile. A no-op metrics/tracing adapter may be allowed in `local`, but `prod` must either require a real adapter or record an explicit, policy-approved waiver in readiness. Silent fallback is not a production profile.

### AR-05 — Remove composition/documentation drift (P1 architecture)

**Evidence.** `App` currently holds and orders modules (`app/app.go:19-45`); products construct `kernel.Kernel` and pools themselves before calling `App.Boot`. README and blueprint text still describe `app` as the composition root containing `kernel.New` and advertise `RunAPI`/`RunWorker`/`RunMigrate` in ways that do not match the public surface (`README.md:148-153`, `docs/blueprint/11-framework-distribution-and-consumption.md:32`). The blueprint's `Context` includes `Clock()` and hooks not present in the live interface, while live services are absent or have different signatures (`docs/blueprint/06-module-sdk.md:65-98`, `module/module.go:66-216`).

**Directive.** Generate reference/API documentation from the authoritative manifest and compile examples in CI. Design documents may describe future state only when visibly labeled “target, not implemented.” No normative code block is allowed without a compile/test owner.

### AR-06 — Remove hidden constructor bypasses from kernel wiring (P1 architecture)

**Evidence.** `kernel.New` constructs the authorization store, optionally decorates it, and gives that instance to the evaluator (`kernel/kernel.go:227-245`). The rules ancestry closure then constructs a fresh `authz.NewStore()` for every call instead of using the composed store (`kernel/kernel.go:249-255`). The current cache deliberately passes ancestry through, so this does not presently change query results; it does bypass any future instrumentation, policy, replica routing, or store decorator and makes the lifecycle description false by construction.

**Directive.** Constructors receive and retain their exact dependencies from the compiled provider graph. No runtime path may instantiate an infrastructure dependency ad hoc when a composed instance exists. Add an AST/lifecycle lint for forbidden constructors outside composition packages and a wiring test that injects a sentinel store to prove every consumer uses it.

## 6. Must-refactors: security and trust boundaries

### SEC-01 — Resolve tenant membership and privileged session state server-side (P0 security)

**Evidence.** `Verifier.Actor` resolves the IdP subject, but validates membership only when `CapacityID` is non-zero. It then copies `TenantID`, `ImpersonatorUserID`, and `BreakGlass` directly from claims (`kernel/auth/auth.go:166-207`). The PostgreSQL principal adapter implements subject lookup and capacity validation but has no active tenant-membership method (`adapters/auth/pgprincipal/pgprincipal.go:25-80`). The HTTP gate binds the resulting tenant directly into RLS context (`kernel/httpx/authz_gate.go:89-112`).

The repository's own normative model says a user may act in a tenant only with active `user_tenant_access`, and tenant selection is resolved from path/header then checked server-side (`docs/blueprint/01-domain-model.md:35-50`, `docs/blueprint/01-domain-model.md:143-159`).

**Risk classification.** The break-glass and impersonation issue is conditional on IdP trust and claim issuance. Even with a perfectly controlled IdP, encoding all privileged-session validity into a bearer token makes revocation and time-boxing harder than a server-side activation record.

**Directive.** Replace `PrincipalStore` with a resolver that receives verified identity plus requested tenant/capacity and returns an authoritative actor. It must:

- reject zero/unknown tenant IDs before opening a tenant transaction;
- require active `user_tenant_access` for every user actor, including capacity-less actors;
- select or validate capacity server-side and require explicit choice when more than one is active;
- load impersonation and break-glass activation records by opaque session/grant ID;
- verify grant status, tenant, actor, impersonated user, approver, reason, activation time, expiry, and revocation on every privileged request or through a short, revocation-aware cache;
- bind `auth_time`, `acr`, and `amr` into actor assurance and enforce freshness for sensitive permissions;
- distinguish user, API key, webhook, and internal system principals through explicit credential schemes.

JWTs may carry cryptographically verified hints and stable IDs, but tenant membership and privileged authorization state are not authoritative solely because they appear in signed claims; the server must independently authorize them.

### SEC-02 — Make workflow privileged operations fail closed (P0 security)

**Evidence.** `workflow.NewRuntime` explicitly allows a nil evaluator (`kernel/workflow/runtime.go:72-90`). `Override` performs its permission check only under `if rt.authz != nil` and otherwise proceeds (`kernel/workflow/runtime.go:276-333`). Ratification is a documented TODO rather than an implemented control (`kernel/workflow/runtime.go:280-282`).

**Directive.** Make an evaluator mandatory for every public runtime. If tests or internal migrations need an ungated engine, expose an unexported test constructor or an explicitly named privileged internal service whose caller must present a capability token. Implement ratification as a real definition field and state transition, or reject definitions/operations that request it. A privileged override must record actor, impersonator, grant ID, source/target states, reason, and ratification outcome in durable audit.

### SEC-03 — Bind webhook replay controls to provider-authenticated data (P1 security)

**Evidence.** The default inbound `HMACVerifier` authenticates only the body (`kernel/webhook/verifier.go:12-58`). `HandleInbound` separately trusts `InboundIn.Timestamp` and `ExternalEventID` supplied by the HTTP adapter for replay-window and dedup decisions (`kernel/webhook/service.go:22-113`, `kernel/webhook/webhook.go:82-100`). A provider-specific verifier may correctly sign these fields, but the interface cannot communicate which canonical timestamp/event ID it authenticated.

**Directive.** Change the verifier contract from `error` to a verified envelope, for example `{CanonicalBody, EventID, OccurredAt, SignatureVersion, KeyID}`. Only values returned by the verifier may drive replay controls. Keep provider-specific canonicalization inside the provider adapter. The generic body-only HMAC verifier must either synthesize replay identity solely from the authenticated body and use receipt time, or be labeled unsuitable for timestamped provider protocols.

### SEC-04 — Bound authorization staleness and memory (P1; P0 when the cache is enabled in production)

**Evidence.** The opt-in authorization cache uses one global mutex and an unbounded map. Expired entries are overwritten only if the same actor is read again; dormant entries are never swept. Concurrent misses for one key duplicate the database load (`kernel/authz/caching.go:29-88`). Revocation invalidation is local to one process and the TTL is the cross-pod stale-allow window (`kernel/authz/caching.go:15-24`).

**Directive.** Use a bounded, sharded cache with admission/eviction metrics and collapsed fills. Add a per-tenant/global authorization epoch or invalidation stream so revocations become visible across pods without accepting an opaque TTL window. The authorization decision must expose whether cached data was used and the epoch/version observed. Production config must set an explicit maximum size and stale-allow bound.

### SEC-05 — Establish a versioned security verification profile (P1 security)

Adopt an auditable mapping to [OWASP ASVS 5.0.0](https://owasp.org/www-project-application-security-verification-standard/) and the [OWASP API Security Top 10 (2023)](https://owasp.org/API-Security/). Target ASVS Level 2 for the default production profile, with Level 3 controls selected for tenant isolation, privileged access, audit/evidence, secrets, cryptography, and high-value workflows. This is a verification map, not a certification claim.

Authentication assurance should align with [NIST SP 800-63-4](https://pages.nist.gov/800-63-4/) and its [authenticator requirements](https://pages.nist.gov/800-63-4/sp800-63b/authenticators/). The framework should consume authoritative IdP assurance (`acr`, `amr`, `auth_time`) and support phishing-resistant step-up/passkeys; it should not become a password store.

Required security test classes include:

- cross-tenant read, write, FK, cache, async, storage-prefix, and presigned-URL isolation;
- object-level authorization for every generated operation;
- privilege escalation across module ownership, ports, relationships, rules, break-glass, and impersonation;
- token substitution, zero-tenant, stale membership, revoked capacity, expired step-up, issuer/audience/key rotation, and JWKS failure;
- webhook canonicalization, signed replay fields, duplicate and reordered events, SSRF rebinding, redirects, and secret rotation;
- resource exhaustion across body size, pagination, filter complexity, route count, rate-limit keys, cache keys, queue depth, provider latency, and object size;
- sensitive-data redaction in every error/log/audit/trace path.

### SEC-06 — Govern explicit outbound-security escape hatches (P1 security/operations)

**Evidence.** The default JWKS client disables ambient proxies and constrains URLs, but callers may inject an arbitrary `*http.Client`; private-IP guarding is intentionally not applied because internal IdPs are supported (`kernel/auth/jwks.go:47-110`). The SSRF-safe HTTP client deliberately lets an exact allowed hostname bypass resolved-IP restrictions (`kernel/httpclient/client.go:41-58`, `kernel/httpclient/client.go:135-175`). These are legitimate operator escape hatches, not defects, but their security depends on immutable trusted configuration.

**Directive.** Production configuration must distinguish trusted internal destinations from arbitrary URLs, validate them at boot, include them in the model/config fingerprint, and audit changes. Prefer injecting a constrained transport policy rather than an unconstrained client. Never allow tenant/user-controlled data to populate host/CIDR allowlists or JWKS clients. Add a startup report that names enabled egress exceptions without exposing credentials.

## 7. Must-refactors: bounded resources and performance

### PERF-01 — Fix the token-bucket map before further rate-limit features (P0 correctness/security)

**Evidence.** A new bucket starts with `burst` tokens and the first request immediately consumes one (`kernel/httpx/ratelimit.go:204-216`). Sweeping deletes only buckets whose stored token count is already at least `burst` (`kernel/httpx/ratelimit.go:222-229`). Refill is calculated only when that same key is used. A one-shot key therefore remains at `burst-1` forever and is never eligible for eviction. Once the map has 10,000 keys, every new request invokes an O(N) sweep (`kernel/httpx/ratelimit.go:200-202`). Existing tests cover limiting and tenant independence but not sweeping or a cardinality attack (`kernel/httpx/ratelimit_test.go:46-215`).

**Directive.** In the immediate patch, compute effective refill during sweep and evict any idle bucket that would now be full. Add a hard capacity and deterministic overflow behavior so correctness does not depend on cleanup. For production scale, expose a bounded sharded implementation and a distributed adapter where per-pod limits are insufficient.

**Acceptance gates.**

- Insert more than 10,000 one-shot keys, advance the fake clock beyond TTL, trigger a sweep, and prove the map returns below the configured bound.
- Benchmark sweep cost at 10k, 100k, and the hard limit.
- Fuzz/rapid-test invalid rates, burst sizes, clock movement, and concurrent keys.
- Emit current entries, evictions, rejected admissions, and sweep duration.

### PERF-02 — Measure complete requests against real PostgreSQL (P1 performance)

**Evidence.** `BenchmarkDispatch` intentionally fakes authentication, evaluation, and transaction work. It proves route-count behavior but not the real request budget (`kernel/httpx/bench_test.go:156` onward). Every tenant transaction issues role/tenant/actor setup statements and, with the RLS guard, queries `pg_roles` (`kernel/database/txmanager.go:84-123`). No benchmark attributes pool wait, transaction setup, authz queries, handler queries, serialization, or middleware separately.

**Directive.** Add reproducible DB-backed benchmarks and traces for representative public, authenticated-read, authenticated-write, resource-authz, idempotent-write, and async-enqueue requests. Record p50/p95/p99 latency, allocations, SQL count, bytes, pool wait, transaction duration, lock wait, and query-plan hash. Test cold/warm cache and 1/10/100 concurrent tenants.

Do not remove RLS guards merely to win a microbenchmark. Optimize only after proving an equivalent connection/role invariant, such as boot-time role verification plus connection reset hooks and continuous health assertions.

### PERF-03 — Collapse rules resolution into bounded SQL work (P1 performance)

**Evidence.** Rule resolution loads ancestry, then executes one lookup per ancestor before tenant and platform fallbacks (`kernel/rules/resolver.go:73-100`). Historical lookup includes active and superseded versions, while migration indexing is concentrated on active lookup. Deep organizations therefore amplify round trips and historical reads may not match the best index shape.

**Directive.** Use one set-based query over an ordered ancestry relation, tenant, and platform fallback, returning the nearest effective version. Add indexes matching both current and historical predicates. Put `EXPLAIN (ANALYZE, BUFFERS)` fixtures around representative depth and history cardinality. Preserve live per-request rule updates; B13 is not needed for rules.

### PERF-04 — Remove N+1 and unbounded materialization from sweepers/workers (P1 performance)

**Evidence.** Workflow SLA sweeping materializes all due tasks, then updates and loads each instance/definition individually (`kernel/workflow/sweeper.go:27-138`). The reminder query has no matching partial `remind_after` index; only `due_at` is indexed (`migrations/00009_workflow.sql:40-55`). Webhook retries load endpoints per delivery (`kernel/webhook/service.go:248-300`). Outbox dispatch performs an inbox insert per subscriber and keeps the outer claim transaction open while tenant handlers run (`kernel/outbox/relay.go:85-160`, `kernel/outbox/relay.go:163-229`).

**Directive.** Claim bounded batches, use set-based updates/joins, cache immutable definitions/endpoints by version, add predicate-matching indexes, and expose queue lag plus batch duration. Rework outbox claim/dispatch as a leased state machine while preserving per-aggregate ordering.

### PERF-05 — Make object checksum behavior explicit (P2 performance)

**Evidence.** S3 `Stat` downloads and hashes the entire object when checksum metadata is absent (`adapters/storage/s3/s3.go:213-245`). This is correct but can turn a metadata check into unbounded bandwidth and latency.

**Directive.** Require framework uploads to persist a canonical checksum in object metadata and document the fallback as an import/repair path with size limits and dedicated metrics. For legacy objects, run asynchronous checksum backfill rather than repeatedly hashing on download confirmation.

### PERF-06 — Make performance gates fail closed (P1 quality)

**Evidence.** A budgeted benchmark absent from output produces only a warning, so renaming or no longer executing a benchmark silently removes the gate (`internal/tools/benchbudget/main.go:49-55`). Hosted CI runs only fuzz seed corpora, not time-bounded coverage-guided fuzzing (`.github/workflows/ci.yml:98-101`).

**Directive.** Missing budget entries fail CI. Run short time-bounded fuzzing on PRs and longer scheduled fuzzing with corpus retention. Track benchmark baselines statistically; require an explicit reviewed budget update for regressions.

## 8. Must-refactors: persistence and durable execution

### DATA-01 — Encode tenant equality in foreign keys (P0 data integrity)

**Evidence.** Tenant-scoped child tables carry `tenant_id` but reference only a parent's `id`, for example persons/legal entities/contacts/capacities → parties (`migrations/00004_org_party_capacity.sql:32-74`), resources → organizations (`migrations/00005_resource_relationship.sql:24-34`), and document versions/grants/attachments → documents or versions (`migrations/00010_documents.sql:40-98`). RLS validates the child row's tenant but does not prove the referenced parent has the same tenant.

Globally unique UUIDs make accidental mismatch unlikely; they do not make the invariant database-enforced. Platform-role bugs, migrations, imports, and future privileged services can create inconsistent graphs that ordinary tenant reads may then partially hide.

**Directive.** Every tenant-local relationship uses a composite FK:

```sql
FOREIGN KEY (tenant_id, party_id)
    REFERENCES parties (tenant_id, id)
```

Migration sequence:

1. Add/confirm unique parent indexes on `(tenant_id, id)`.
2. Audit every existing child-parent pair for tenant mismatch and fail deployment if any exist.
3. Add composite constraints `NOT VALID` where lock duration matters.
4. Validate constraints in a controlled migration.
5. Update generators/testkit so every new tenant table and FK is catalog-checked.
6. Remove redundant single-column FKs only after all consumers and rollback paths are verified.

### DATA-02 — Add lease generations/fencing and effect idempotency to jobs (P0 reliability)

**Evidence.** Job claim changes status to running and sets `locked_at`, but no claim token/generation is returned (`kernel/jobs/runner.go:291-319`). Completion/failure updates match only `id` (`kernel/jobs/runner.go:421-485`). Reclaim resets every sufficiently old running row (`kernel/jobs/runner.go:522-537`). The configured timeout floor reduces overlap but cannot prevent a delayed or partitioned old worker from committing an outcome after another worker reclaims the job.

**Directive.** Add `lease_token`, monotonically increasing `lease_generation`, `lease_expires_at`, and optional heartbeat. Every final queue update must compare the token/generation. Pass a stable job idempotency key and lease context to workers. For domain side effects, require one of:

- an inbox/effect ledger unique on `(job_id, effect_name)` in the same transaction;
- a domain CAS that incorporates the expected generation/state;
- a provider idempotency key for external effects.

Fencing the queue row alone does not undo a stale worker's already-committed domain transaction; the worker contract and testkit must make that explicit.

### DATA-03 — Move remote provider/secret I/O outside database transactions (P0 reliability/performance)

**Evidence.** Notification delivery claims rows with `FOR UPDATE SKIP LOCKED`, invokes the channel sender, then updates status inside one tenant transaction (`kernel/notify/service.go:446-585`). Webhook dispatch and retry call `Sender.Post` through `deliverToEndpoint` inside tenant transactions (`kernel/webhook/service.go:203-238`, `kernel/webhook/service.go:248-453`). Webhook paths also resolve endpoint secrets while a transaction is open (`kernel/webhook/service.go:32-54`, `kernel/webhook/service.go:364-392`); whether that is remote I/O depends on the configured resolver. Locks and connections remain held for provider DNS, connect, TLS, response, timeout, retries, and potentially remote secret resolution. A crash after provider success but before DB commit can duplicate delivery.

**Directive.** Standardize a three-stage durable-delivery protocol:

1. **Claim:** in a short transaction, atomically move due rows to `leased`, assign lease token/generation/expiry, and commit.
2. **Execute:** resolve any remote secret and perform network I/O with no database transaction open, using delivery ID as the stable provider idempotency key. Local, bounded, non-I/O secret lookups may remain in a transaction only when the provider contract proves that property. An adapter must declare whether the receiver honors idempotency and may not present a non-idempotent high-impact operation as duplicate-safe.
3. **Finalize:** in a short transaction, compare the lease token and move to sent/failed/dead; persist provider receipt, bounded error, attempt, and next time.

Transport attempts remain at-least-once; one logical external effect requires provider idempotency or an independently proven duplicate-safe operation. Add chaos tests at every boundary: before send, during send, after success/before finalize, lease expiry, duplicate workers, and provider timeout.

Inbound verification uses a separate two-phase protocol. First, a short read transaction loads endpoint ID, tenant, status, verifier/provider key, secret reference, and endpoint/secret version, then closes. Resolve any remote secret and authenticate the canonical envelope outside a transaction. Second, a short write transaction compares the endpoint status and version/secret reference with the verified snapshot; if rotation or deactivation occurred, discard the result and retry verification under the new version rather than persisting under stale policy. Persist the verified envelope, replay identity, authenticated timestamp, signature/key version, and dedup result atomically. A failed signature may write a body-free audit row in its own short transaction after the same endpoint-version check. Tests rotate/deactivate the endpoint between both phases and prove no event is accepted under a mismatched secret/config version.

### DATA-04 — Reconcile bulk-processing concurrency and make multi-worker mode safe (P1; P0 before advertising/enabling multi-worker execution)

**Evidence.** Migration comments claim `FOR UPDATE SKIP LOCKED` replica-safe processing (`migrations/00016_bulk_operations.sql:1-6`), but `Service.next` explicitly performs an unlocked read and assumes one processor per operation (`kernel/bulk/bulk.go:123-144`). Two workers can execute the same item concurrently before either marks it done.

**Directive.** Either narrow the public contract to a single-owner local processor and enforce an operation lock, or implement atomic leased claims with `UPDATE ... FROM (SELECT ... FOR UPDATE SKIP LOCKED) RETURNING`. For a premier framework, implement the latter with fencing, retry policy, item idempotency keys, cancellation, pause/resume, and bounded batch claims.

### DATA-05 — Allocate immutable versions without `MAX()+1` races and clean orphan blobs (P1 reliability)

**Evidence.** Artifact generation uses `MAX(version)+1` and intentionally returns a conflict for concurrent generation (`kernel/artifact/artifact.go:64-106`). Document upload reserves `MAX(version_no)+1`; concurrent sessions can reserve the same version, and the loser can leave a randomly keyed object orphaned (`kernel/document/service.go:159-194`). Tests confirm this behavior rather than eliminating it.

**Directive.** Allocate versions by locking/updating a parent counter or a dedicated per-aggregate sequence row. Upload sessions become durable records with expiry, intended checksum/size, storage key, status, and cleanup ownership. Confirmation CASes the session and version; a scheduled garbage collector removes expired/unreferenced objects with metrics and audit.

### DATA-06 — Integrate the resource mirror into the aggregate write contract (P1 architecture/data)

**Evidence.** Modules own business tables and must separately upsert the generic `resources` mirror (`kernel/resource/resource.go:1-6`). The registrar inserts nil actor placeholders (`kernel/resource/registrar_pg.go:34-58`). Even the reference request handler manually performs the business insert and mirror upsert (`internal/testmodules/requests/handlers.go:43-54`). A forgotten mirror write breaks resource authz, comments, documents, workflow, or relationships without necessarily failing the business write.

**Directive.** Make resource projection part of the typed aggregate repository/unit-of-work contract generated from the module model. The same transaction must write the aggregate, mirror, audit change, and outbox event. The framework should provide a complete helper/generator, not rely on a comment. Source actor from context and reject missing actor for user-initiated writes.

### DATA-07 — Complete relationship semantics and actor attribution (P1 correctness)

**Evidence.** The ReBAC checker ignores party-subject edges and only evaluates capacity edges (`kernel/relationship/relationship.go:42-66`). Relationship inserts and resource mirror writes use nil actor placeholders (`kernel/relationship/relationship.go:69-88`, `kernel/resource/registrar_pg.go:38-58`).

**Directive.** Resolve actor → active capacity → optional party through the authoritative principal model and evaluate explicitly supported subject kinds. Every authorization-input mutation must be ownership-checked, attributed, audited, versioned, and invalidate relevant caches.

### DATA-08 — Make compliance evidence complete, durable, and centrally enforced (P0/P1 compliance)

**Evidence.** Audit metadata is stored but intentionally excluded from the row hash (`kernel/audit/audit.go:99-159`), so the full row is not tamper-evident. Attachment creation ignores an outbox write error after inserting the attachment (`kernel/attachment/attachment.go:72-89`). Legal notification delivery audit is marked deferred because the worker supposedly lacks outbox INSERT, but migration 00011 explicitly grants it (`kernel/notify/service.go:451-453`, `kernel/notify/service.go:546-559`, `migrations/00011_notify_webhook_integration.sql:172-178`). The blueprint requires a legal delivery audit with provider receipt (`docs/blueprint/07-platform-services.md:83-94`).

DSR export aggregates results in memory, marks the request complete, and returns the map; it does not persist a signed/encrypted artifact, manifest, checksum, expiry, or delivery record (`kernel/retention/engine.go:119-149`). Erasure and disposition callbacks are individually responsible for consulting legal holds (`kernel/retention/engine.go:22-45`, `kernel/retention/engine.go:151-180`).

The ignored required outbox error and missing legal-delivery receipt are P0 and belong to Wave 0. Full declared-row audit integrity, durable DSR delivery, and central hold enforcement are P1 premier-release requirements in Wave 6.

**Directive.**

- Define the immutable audit-row integrity contract and hash every persisted field in it, including tenant identity, canonicalized metadata, `tx_id`, all nullable core fields, sequence, ID, timestamps, and previous hash. Obtain transaction identity before hash construction or compute/verify the canonical hash in the database; do not call the result “full-row integrity” while any mutable/evidentiary field is excluded.
- Treat required outbox/audit writes as part of the transaction; never discard their errors.
- Write legal-delivery evidence with provider receipt after successful delivery.
- Persist DSR exports as encrypted immutable artifacts with manifest, per-class results, checksum, creation/expiry, access policy, and download audit before completing the request.
- Move hold enforcement into a framework wrapper that each dispose/erase callback must pass through; callbacks return candidates/effects, not unilateral authority to ignore holds.
- Emit explicit partial/not-applicable results for classes without export/erase callbacks rather than silently skipping them.

### DATA-09 — Adopt an online expand/backfill/validate/contract protocol (P0 release/data)

The composite-FK and lease-state changes affect live, shared tables and must support mixed binary versions. The v1 migration contract is:

1. **Classify and preflight.** Every migration declares `online` or `maintenance`, estimated rows/bytes, lock/statement timeout, retry behavior, N/N-1 compatibility, backfill owner, validation query, and rollback/forward-fix plan in a machine-readable manifest. The default online DDL lock timeout is 2 seconds; exceeding it aborts and retries rather than waiting behind production traffic. A migration needing stronger locks is `maintenance` and cannot run in the normal rolling path.
2. **Expand.** Add nullable/default-safe columns, new tables, indexes, triggers/compatibility views, and `NOT VALID` constraints without removing or reinterpreting old fields. Use non-transactional concurrent index creation where PostgreSQL requires it. Queue/outbox state additions are accepted by old readers and new readers; new producers continue writing the old representation until the compatibility reader is deployed.
3. **Backfill.** Run resumable, tenant-scoped, keyset-paginated jobs with checkpoints, bounded batch/transaction time, sleep/rate controls, row/error counters, and safe restart. Never perform an unbounded all-tenant rewrite in a release transaction. Reconciliation queries must prove old/new representations agree.
4. **Validate.** Validate constraints and invariants separately after backfill. For tenant FKs, record zero mismatch count and successful `VALIDATE CONSTRAINT`. For lease/queue changes, prove every active row has a valid state/token invariant. Validation artifacts include query text, plan, duration, and row counts.
5. **Deploy N.** Canary API and worker N against the expanded schema while N-1 remains active. Observe errors, lock/pool/queue metrics, old/new read parity, and model/schema hashes for a defined soak. Rollback means returning application traffic to N-1 while retaining additive schema; production rollback does not run destructive `Down` migrations.
6. **Switch.** Enable new writes/reads through a separately controlled, observable compatibility flag only after canary parity. For queues/events, consumers accept both schema versions during the whole window and producers change versions only after all consumers are compatible.
7. **Contract.** Remove old columns/states/views and compatibility code no sooner than the next minor release, after telemetry proves no N-1 process or old producer/consumer remains. Contract is a new forward migration with its own backup/restore and maintenance analysis.

Required CI/deployment drills are N-1 code on expanded N schema, N code before/after backfill, interrupted/resumed backfill, partial fleet rollout, application rollback after switch, and forward recovery from every failed phase. Disposable `Down` testing remains useful but is not the production rollback strategy.

## 9. Must-refactors: developer experience and the DSL

### DX-01 — Make the source-built CLI path valid (P0 DX)

**Evidence.** README explicitly offers building the CLI from a repository clone (`README.md:174-195`). When build info is `devel`, `wowapi init` changes it to `v0.0.0` and writes that into the generated `go.mod` (`internal/cli/init_cmd.go:121-132`, `internal/cli/templates/init/go.mod.tmpl:1-7`). `v0.0.0` is valid module-version syntax, but this revision has no matching tag and the value is not guaranteed to resolve to the checked-out source; the generated dependency can therefore be unresolvable or unrelated.

**Directive.** Never invent a version. A development CLI must do one of these explicitly:

- require `--framework-version vX.Y.Z` and verify that it resolves;
- accept `--local-framework /absolute/path` and emit a deliberate `replace` directive plus a visible development warning;
- derive the exact pseudo-version from VCS metadata when the commit is reachable and verify it with `go list -m`.

If none applies, fail before writing files with a command the user can run. The generated repository must pass `go mod download`, `go build ./...`, and its contract tests in an isolated temp directory.

### DX-02 — Replace the TODO generator with a tested vertical-slice generator (P0/P1 DX)

**Evidence.** Generated create, get, update, and delete handlers are TODOs returning empty success objects (`internal/cli/templates/crud/resource.go.tmpl:46-65`, `internal/cli/templates/crud/resource.go.tmpl:140-147`). The route uses permission `.delete`, but the closed authorization verb set contains `deactivate`, not `delete` (`internal/cli/templates/crud/resource.go.tmpl:50-54`, `kernel/authz/registry.go:13-19`). The generated migration has no status column even though delete claims to perform a status transition (`internal/cli/templates/crud/migration.sql.tmpl:7-18`). The generator emits only Go and SQL, not permissions, resource declaration, OpenAPI, tests, or module registration (`internal/cli/gen_cmd.go:135-159`). Module generation itself leaves migration and registration TODOs.

Emitting an authorization-invalid route and false-success handlers from an advertised command is the P0 portion: Wave 0 must either disable that path with an explicit experimental/unsupported error or implement a minimally correct slice. The complete cross-projection generator below is P1 and belongs to Wave 4.

**Directive.** `wowapi gen crud` must either generate a complete, secure simple-resource slice or stop calling itself CRUD. The complete slice includes:

- typed create/update/input/output models with validation;
- create/read/list/update/deactivate/restore operations using valid verbs;
- optimistic concurrency and ETag behavior;
- tenant-safe migration with status, composite FK conventions, indexes, and actor fields;
- atomic resource mirror, audit, and outbox integration;
- permission/resource manifest entries;
- OpenAPI operations and schemas derived from the same operation declarations;
- unit, contract, RLS, authorization, idempotency, pagination, and migration tests;
- automatic module registration with no manual TODO required to boot.

Generated UUID/time fields use `uuid.UUID`/`time.Time`, not wire strings paired with UUID/timestamptz SQL unless an explicit conversion layer is emitted. Every generator test compiles and boots the generated product; substring assertions are insufficient.

### DX-03 — Define the state-of-the-art module DSL (P1 architecture/DX)

The target DSL must be declarative about contracts and explicit about business behavior. Proposed concepts below are future design, not claims about current packages.

#### Module identity and manifest

```go
type ModuleID string

type Manifest struct {
	ID           ModuleID
	Version      string
	Dependencies []ModuleID
	Config       ConfigContract
	Capabilities []CapabilityRequirement
	Migrations   MigrationBundle
	Seeds        SeedBundle
	Locales      LocaleBundle
}
```

The compiler creates an owner-bound registrar from `Manifest.ID`; no nested declaration accepts a free-form owner. `Version` is the module contract/schema version, separate from the framework release.

#### Typed operations

```go
type OperationKind string

const (
	OperationSync   OperationKind = "sync"
	OperationAsync  OperationKind = "async"
	OperationStream OperationKind = "stream"
)

type TenantScope string

const (
	TenantNone     TenantScope = "none"
	TenantCurrent  TenantScope = "current_tenant"
	TenantPlatform TenantScope = "platform"
)

type CredentialScheme string

const (
	CredentialAnonymous CredentialScheme = "anonymous"
	CredentialUser      CredentialScheme = "user"
	CredentialAPIKey    CredentialScheme = "api_key"
	CredentialWebhook   CredentialScheme = "webhook"
	CredentialInternal  CredentialScheme = "internal"
)

type AuthenticationPolicy struct {
	Allowed            []CredentialScheme
	MinimumACR         string
	RequiredAMR        []string
	MaximumAuthAge     time.Duration
}

type CompatibilityMode string

const (
	CompatibilityBackward CompatibilityMode = "backward"
	CompatibilityExact    CompatibilityMode = "exact"
)

type Schema[T any] struct {
	ID          string
	Version     uint32
	JSONSchema  json.RawMessage
	Fingerprint string
	Validate    func(T) error
}

type Authorization[Request any] struct {
	Permission PermissionKey
	Scope      AuthorizationScope
	Target     func(Request) (resource.Ref, error)
}

type ErrorContract struct {
	Code       string
	Kind       errors.Kind
	HTTPStatus int
}

type EmittedEvent struct {
	Key           EventKey
	SchemaVersion uint32
	Compatibility CompatibilityMode
}

type AsyncPolicy[Request any] struct {
	Job             JobKey
	IdempotencyKey  func(RequestContext, Request) string
	MaxAttempts     int
	ExecutionBudget time.Duration
}

type StreamPolicy struct {
	MaxRecords      int64
	MaxBytes        int64
	MaxDuration     time.Duration
	Backpressure    BackpressurePolicy
}

type ExecutionPolicy[Request any] struct {
	Kind   OperationKind
	Async  *AsyncPolicy[Request]
	Stream *StreamPolicy
}

type Operation[Request, Response any] struct {
	ID            string
	Method        string
	Path          string
	Tenant        TenantScope
	Authentication AuthenticationPolicy
	Authorization Authorization[Request]
	Input         Schema[Request]
	Output        Schema[Response]
	Errors        []ErrorContract
	Emits         []EmittedEvent
	Execution     ExecutionPolicy[Request]
	Idempotency   IdempotencyPolicy
	Concurrency   ConcurrencyPolicy
	Audit         AuditPolicy
	RateLimit     RateLimitPolicy
	Observability ObservabilityPolicy
	Handler       func(context.Context, RequestContext, Request) (Response, error)
}
```

These names describe the required contract; final package names can change. `JSONSchema`, `Fingerprint`, and the validation executable in `Schema` are compiled projections of one canonical typed schema declaration, not three independent inputs that module authors keep synchronized. Anonymous operations use `TenantNone`, allow only `CredentialAnonymous`, and declare no authorization permission. `TenantCurrent` disallows anonymous credentials and requires a permission; `TenantPlatform` allows only explicitly approved service/privileged schemes and platform authorization. Webhook credentials resolve to the verified-envelope contract from SEC-03. The compiler also validates that resource-scoped authorization supplies a non-zero target resolver, assurance requirements are satisfiable by every allowed credential scheme, error codes are globally stable, every schema ID/version/fingerprint is unique and valid, emitted event versions obey compatibility policy, async mode has exactly one async policy, stream mode has exactly one bounded stream policy, and sync mode has neither. Generic free functions register heterogeneous operations into the model. The operation is the source for route metadata, authentication/authorization wrapping, OpenAPI security declarations, compatibility checks, and conformance tests. Request-time dispatch remains normal Go with precompiled descriptors; no reflection belongs in the hot path.

#### Typed owned identifiers

Use distinct key types for permissions, resources, events, jobs, rules, workflow actions, notification templates, integration providers, and ports. Construction binds the owner. Raw strings may exist at serialization boundaries, but module code should not repeatedly assemble identity strings.

#### Narrow runtime capabilities

Replace the mega-interface with small immutable handles such as `Runtime`, `Operations`, `Transactions`, `Events`, and explicitly required typed ports. Registration receives declaration capability only; request handlers receive runtime capability only. No handler should retain or even see a mutable boot registry.

#### Open workflow extension

Workflow definitions currently use a closed set of step types plus stringly auto-action/resolver keys. Keep built-ins for approvals/votes/todos/terminals, but let modules register owned typed action/assignee/condition descriptors with input/output schemas, timeout/idempotency policy, and compensation semantics. Unknown action versions fail model compilation.

### DX-04 — Create one golden consumer and upgrade matrix (P1 DX/compatibility)

The framework repo needs a generated, non-internal consumer fixture that is treated like a third-party product:

1. Install the built CLI.
2. Scaffold a product and at least two interacting modules.
3. Generate a resource, rule, workflow, event handler, recurring job, document flow, notification, and webhook.
4. Run migrations/seeds/rule sync.
5. Boot API and worker against Postgres/MinIO/Mailpit/OTel.
6. Exercise authenticated CRUD, async delivery, restart/retry, and RLS isolation.
7. Upgrade from the previous supported framework version and rerun all contracts.

This catches public API, templates, docs, generated code, and runtime wiring as one product experience.

### DX-05 — Make CLI/docs/version identity singular (P1 documentation/release hygiene)

**Evidence.** README and upgrade policy say pre-1.0/v0 (`README.md:23-25`, `docs/operations/upgrade-and-deprecation-policy.md:1-13`), while CHANGELOG says the public surface is stable as of v1.0.0 and records v1.1.0 (`CHANGELOG.md:7-14`). README recommends `@latest` while policy requires exact pins. Blueprint CLI examples advertise unsupported commands and flags (`docs/blueprint/11-framework-distribution-and-consumption.md:112-137`, `internal/cli/cli.go:92-112`).

**Decision.** The repository is on the stable v1 line. The reviewed Git history contains `v1.0.0` and `v1.1.0` tags, and CHANGELOG records the v1 stability promise; the pre-1.0 README and upgrade-policy text is stale. Apply these rules:

- Public Go symbols, generated contracts, config semantics, event compatibility, and migrations remain backward-compatible throughout v1. An incompatible public change requires a `/v2` module path and a v2 migration guide.
- Do not widen the existing `module.Context` interface again in v1. Add narrow interfaces/packages and adapters; remove the legacy surface only in v2.
- Support the current and immediately previous v1 minor lines. The previous minor receives critical security/data-integrity fixes for at least six months after its successor. At the reviewed state, that means v1.1 and v1.0.
- A generated product records the framework major/minor and manifest-schema version. Mutating CLI generators require the same framework major/minor; patch differences are allowed only when generated-template compatibility tests pass. `wowapi version` fails mutating commands on an incompatible pairing rather than merely warning.
- Rolling deployment compatibility is N/N-1 minor: N code must run on the expanded N schema, and N-1 code must continue to run during the N rollout until the contract phase. Direct upgrades older than N-1 run the intervening upgrade steps in order.
- Within v1, OpenAPI request requirements, response removals/narrowing, security weakening, config removals/semantic changes, and incompatible event schema changes are release-blocking. Additive optional fields and new operations remain allowed.

Update README, upgrade policy, command examples, generated `go.mod`, and release automation to this decision. Generate command-reference docs from parser definitions and execute every shell example in CI where practical.

### DX-06 — Make OpenAPI merge complete or fail loudly (P1 API governance)

Preserve all supported OpenAPI 3.1 fields with explicit merge policies, or reject fragments containing unsupported fields. Silently discarding `security`, `tags`, `servers`, `webhooks`, callbacks, parameters, responses, and non-schema components is unacceptable. Validate the final document against the [OpenAPI 3.1.1 specification](https://spec.openapis.org/oas/v3.1.1.html); schema objects should follow [JSON Schema 2020-12](https://json-schema.org/draft/2020-12).

Add semantic API diffing that classifies breaking request/response/security/path changes and gates v1 releases according to DX-05.

### DX-07 — Make readiness and configuration diagnostics truthful (P1 operations/DX)

**Evidence.** The health contract describes readiness as including migration currency (`kernel/httpx/health.go:10-15`, `app/health.go:8-14`), but the generated API supplies DB ping and seed-catalog checks, not a migration-current check (`internal/cli/templates/init/cmd_api_main.go.tmpl:207-219`). Product-aware config validation is delegated only when `tools/configcheck/main.go` is found relative to the current working directory; otherwise the installed CLI silently performs framework-only validation (`internal/cli/config_delegate.go:25-45`). Capacity validation defaults to advisory, may be skipped when deployment shape is unset, and HTTP backpressure defaults off (`kernel/config/concurrency.go:32-52`, `kernel/config/concurrency.go:217-246`).

**Directive.** Readiness claims only checks it actually executes and must include migration version, seed/rule/model hash, required adapters, and critical queue/storage dependencies according to the process profile. `config doctor` discovers the product root through `go env GOMOD`/explicit `--project`, reports whether product validation ran, and fails a production check if only framework validation is available. The production profile requires a declared/enforced capacity shape and an intentional backpressure policy; advisory/unset state is a visible readiness failure or approved waiver.

## 10. Must-haves for a premier framework

These are definitive product requirements, not an invitation to build every fashionable subsystem. Each item closes a repeated class of product-side manual work or makes a critical invariant enforceable.

### 10.1 Immutable application compiler

- Ownership-bound module manifests.
- Typed provider/requirement graph with scopes.
- Whole-model deterministic validation and model hash.
- API/worker/migrate runtime profiles derived from one graph.
- Read-only snapshots and explicit seal semantics.
- Machine-readable manifest and human-readable `wowapi inspect` output.

### 10.2 First-class operation model

- Typed request/response/error contracts.
- Route, authorization, idempotency, concurrency, rate-limit, audit, and observability policies in one descriptor.
- OpenAPI and test contracts derived from the operation.
- Resource-level authorization hook that cannot be forgotten for object operations.
- Streaming and asynchronous operation variants with explicit limits.

### 10.3 Tenancy invariant compiler

- Catalog of every tenant-scoped table and policy.
- Composite tenant FK generation/verification.
- RLS enable/force/policy/grant verification.
- Storage prefix, cache key, queue row, audit row, and trace/log tenant propagation tests.
- Explicit platform/global tables and approved crossing services.

### 10.4 Authoritative identity and privileged-session service

- Server-side tenant membership and capacity selection.
- API key/webhook/system/user credential separation.
- Revocable, time-boxed break-glass and impersonation grants.
- Assurance/freshness-aware step-up.
- Phishing-resistant authentication support through IdP/WebAuthn claims.
- Complete actor attribution and audit.

### 10.5 Durable execution substrate

- Shared lease/fencing state machine for jobs, outbox, notifications, webhooks, bulk work, and maintenance sweeps.
- Stable idempotency keys and effect ledger.
- Pause/resume/cancel/replay with audit.
- DLQ inspection and safe replay policies.
- Backpressure, quotas, fairness, and per-tenant concurrency.
- Chaos-tested crash semantics.

### 10.6 Versioned event and schema registry

- Owned event keys and immutable schema versions.
- Compatibility policy and upcaster/downcaster strategy where needed.
- Subscriber declarations and “no subscribers” policy (allowed, warning, or error per event).
- Payload size/sensitivity classification and redaction.
- OpenTelemetry messaging semantic conventions.

### 10.7 Compliance/evidence plane

- Full-row canonical audit integrity with external anchor verification.
- Legal hold centrally enforced across dispose/erase/document flows.
- Durable DSR export artifact and delivery lifecycle.
- Evidence bundles for privileged access, workflow override/ratification, notification receipts, and data erasure.
- Retention classes that must explicitly declare export/erase/dispose support.
- Cryptographic agility and key-rotation metadata.

### 10.8 Production capability profiles

- Named `local`, `test`, `stage`, and `prod` profiles.
- Required/optional adapter matrix for metrics, traces, secrets, object storage, IdP, mail/SMS/push, malware scanning, and audit anchoring.
- Boot/readiness failure for missing required adapters.
- Capacity planning and unsafe-config policy integrated into `config doctor`.
- Unknown module/config namespace rejection.

### 10.9 Observability and SLO kit

- OpenTelemetry semantic conventions for HTTP, DB, messaging, jobs, and object storage, tracking the current stable conventions rather than bespoke naming. The current reference is [OpenTelemetry semantic conventions 1.43.0](https://opentelemetry.io/docs/specs/semconv/).
- RED metrics for HTTP and workers; pool wait, transaction time, lock wait, queue age, lease expiry, retries, DLQ growth, provider latency, cache size/eviction, and rate-limit cardinality.
- Low-cardinality metric dimensions; exact tenant in logs/traces, tenant tier in metrics.
- Trace propagation over every async boundary and links for retries/replays.
- Default SLOs and alerts with documented tuning, not hard-coded universal thresholds.

### 10.10 Operator control plane

- Read-only topology/model inspection.
- Migration/seed/rule/schema drift status.
- Queue/DLQ state, replay preview, and audited execution.
- Privileged-grant activation/revocation and live visibility.
- Key/secret/certificate rotation status.
- Readiness that verifies DB migration currency, seed/rule catalogs, required adapters, and model hash.

### 10.11 Compatibility and lifecycle discipline

- One truthful semantic-version policy.
- Public API inventory and automated API diff.
- Generated-code version compatibility check.
- Module manifest schema version and migration tool.
- Deprecation annotations, overlap window, and tested upgrade path.
- Previous-version golden consumer in CI.

### 10.12 Premier out-of-box experience

- A source-built or released CLI always generates a resolvable project.
- Generated projects compile, migrate, boot, and pass a smoke test without TODOs.
- One command starts the generated API/worker/dependencies for local development.
- Examples are executable and version-aligned.
- Error messages contain the failed invariant and exact remediation command without leaking secrets.
- The framework clearly distinguishes trusted in-process modules from any future untrusted plugin model. In-process Go modules are not a security sandbox; untrusted extensions require an out-of-process or sandboxed/Wasm contract.

### 10.13 Elite engineering discipline

- Public interfaces are consumer-owned and narrow; optional security collaborators use explicit option/result types, not nil meaning “skip the control.”
- Constructors validate complete invariants and return errors; panic is reserved for impossible programmer invariants in generated/static declarations.
- Every ignored error needs a proof that the operation is genuinely best-effort plus a metric/log; audit, outbox, persistence, close/flush, and finalization errors are never discarded by default.
- `context.Context` is the first parameter for I/O, deadlines propagate, background work has explicit ownership, and goroutines have bounded lifetimes plus shutdown tests.
- Maps/slices crossing package boundaries are cloned or immutable. Every identity-indexed runtime map declares cardinality, eviction, and metrics. This includes the webhook breaker registry, which currently creates per-endpoint state without eviction (`kernel/webhook/breaker.go:85-108`).
- SQL is parameterized, query shape is bounded, tenant equality is encoded in constraints, concurrency semantics are documented beside the statement, and representative plans are regression-tested.
- State machines enumerate valid states/transitions and use compare-and-set updates. String constants remain serialization details behind typed APIs.
- Errors have stable machine codes, safe public messages, wrapped causes, operation context, and redaction tests. No secret/provider body is copied into logs, traces, metrics, or DLQ text.
- Direct dependencies and build tools have explicit owners, version policy, vulnerability/license review, and reproducible pins. Release images and service dependencies use immutable digests where reproducibility matters; floating `latest` and broad tool ranges are confined to documented local convenience.
- Tests assert externally meaningful contracts and failure modes, not merely coverage. Generated artifacts are compiled and executed. Concurrency work includes race, duplicate-worker, cancellation, timeout, and fault-injection tests.
- Documentation statements are classified as implemented, target, experimental, deprecated, or unsupported and are checked against executable examples/manifests.

## 11. Release engineering and quality directive

### REL-01 — Gate release on the exact commit being published (P0 supply chain)

**Evidence.** `.github/workflows/release.yml` triggers on every `v*` tag and immediately builds, signs, attests, pushes images, and creates a release (`.github/workflows/release.yml:11-17`, `.github/workflows/release.yml:25-128`). There is no dependency on authoritative CI, security scans, API compatibility, migration checks, or a protected environment in the workflow itself.

**Selected design.** Use one reusable, repository-owned `required-gates` workflow and make the tag workflow rerun it for the exact tag SHA before any publish permission is available.

1. Add `.github/workflows/required-gates.yml` with `workflow_call`. It checks out the caller's immutable SHA and executes a versioned `ci/release-gates.yaml` manifest. Each entry has an ID, command/reusable job, owner, `required_from_wave`, timeout, and evidence artifact. The Wave 0 manifest contains every extant authoritative/container/DB/S3, lint/action-lint, blocking security, migration/reversibility, fuzz, build, and release-config check. Generated-consumer and OpenAPI/Go/config/event compatibility checks are added to the manifest in the same change that implements them and become mandatory at their declared Wave 4 boundary. A `premier` release requires `completed_wave: 6`, forbids skipped/not-applicable entries, and therefore runs the full set; an earlier maintenance release carries its lower completed-wave value and must not claim premier readiness.
2. Both PR/main CI and release call that workflow. Release does not trust a similarly named check on another SHA or branch; its `verify` job passes `${{ github.sha }}` from the tag event and emits an attested `gate-results.json` containing source SHA, workflow/run IDs, completed-wave value, each required check, tool versions, results, and evidence hashes.
3. `build-candidate` has `needs: verify`, `contents: read`, `id-token: write`, and `attestations: write`, but no `packages` or `contents` write permission. It verifies `gate-results.json`, builds each release artifact exactly once, and emits archives/checksums/SBOMs plus a multi-architecture OCI image-layout archive as immutable Actions artifacts. It then creates and attests `release-manifest.json`, binding the gate-results digest, source SHA, artifact IDs/hashes, image digest/platforms, and builder identity. Candidate build does not push an image or create a release.
4. `publish` has `needs: build-candidate`, `contents: write`, `packages: write`, `id-token: write`, and `attestations: write`; it is the only job allowed to mutate release/package state and uses a protected `release` environment. With pinned tooling, it publishes the exact archived bytes and copies all platforms from the OCI layout to the registry under semantic-version/digest tags; it does not run a build. The protected-environment transition is the promotion boundary.
5. `verify-published` downloads by version/digest and runs a repository script that verifies archive checksums, keyless signatures, SBOM/provenance attestations, subject repository/workflow identity, source SHA, image platforms, CLI `version`, and the equality of published hashes with `release-manifest.json`. A failed verification marks the release failed and prevents/withdraws `latest` promotion.

Machine acceptance is: a deliberately failing required check prevents `build-candidate`; changing the tag target changes both manifest SHAs; tampering with gate results or candidate bytes is detected; the publish job rejects any artifact/digest absent from `release-manifest.json`; and post-publish verification succeeds from a clean runner with no build workspace.

Protect release tags and the environment at the repository/organization level. Target and document SLSA guarantees using the [SLSA 1.2 build track](https://slsa.dev/spec/v1.2/build-track-basics) and [provenance model](https://slsa.dev/spec/v1.2/provenance), without claiming a level the builder configuration has not been assessed to meet.

### REL-02 — Make security checks blocking or replace them (P0/P1 supply chain)

**Evidence.** Trivy ignores unfixed findings and exits zero (`.github/workflows/security-scan.yml:63-75`). Dependency review runs only for public repositories (`.github/workflows/security-scan.yml:77-93`). CodeQL and Scorecard skip private repositories without GHAS (`.github/workflows/codeql.yml:3-8`, `.github/workflows/scorecard.yml:3-8`).

**Directive.** The production branch must have blocking SAST, dependency/license review, secret scanning, reachable Go vulnerability scanning, container/filesystem/config scanning, and action/workflow lint regardless of repository visibility. Where GitHub-hosted features are unavailable, run local alternatives and retain artifacts. Maintain a reviewed allowlist with owner, rationale, expiry, and remediation issue; `ignore-unfixed` is not a universal policy.

### REL-03 — Expand compatibility gates (P1 quality)

Required release checks:

- Go public API diff and module compile matrix.
- OpenAPI semantic diff.
- config schema compatibility and generated config fixture migration.
- event/schema compatibility.
- migration upgrade from the oldest supported version plus reversibility on disposable data.
- generated consumer upgrade.
- container architecture smoke on every published architecture.
- SBOM/provenance/signature verification after publish.

### REL-04 — Make integration coverage truthful (P1 quality)

- Make the toolbox depend on healthy MinIO and set both `WOWAPI_REQUIRE_S3=1` and `S3_TEST_ENDPOINT=minio:9000` in the authoritative container gate; use one canonical endpoint variable in a follow-up cleanup.
- Make E2E prerequisites fail in the authoritative E2E job, not skip.
- Replace time-sensitive TOTP skip behavior with deterministic clocks/codes.
- Report executed/skipped integration suites as a machine-checked manifest.
- Run race tests over integration-relevant packages where feasible.
- Run actual time-bounded fuzzing. Go's default `go test` execution only runs seed corpora; coverage-guided fuzzing requires `-fuzz`, as documented by [Go fuzzing](https://go.dev/doc/security/fuzz/).

## 12. Phased implementation blueprint

The phases are dependency-ordered. Teams may parallelize work inside a wave only when write sets and invariants are independent.

### Wave 0 — Stop release and correctness hazards

Deliver these small, high-leverage fixes first:

1. Fix token-bucket eviction; add hard bound, sweep/cardinality tests, and metrics.
2. Require workflow authorization for override; reject or implement ratification declarations.
3. Propagate attachment outbox errors and implement legal notification audit using the already-granted outbox permission.
4. Fix development CLI dependency handling. Make `gen crud` fail honestly as unsupported/experimental or replace its false-success TODO handlers with a minimally correct slice; fix generated permission verb/status semantics and make generated-output compile/boot tests authoritative.
5. Apply the DX-05 v1/N-1 compatibility decision across README, CHANGELOG, policy, tags, CLI, generators, and diff gates.
6. Implement the REL-01 reusable exact-SHA gate → immutable candidate → protected digest promotion → clean-runner verification workflow.
7. Make benchmark absence and required S3/E2E execution fail closed.
8. Add targeted regression tests for each defect before refactoring adjacent code.

**Exit gate:** no P0 defect assigned to Wave 0 remains; a generated product from both released and development CLI paths resolves, builds, migrates, boots, and runs its smoke suite. A release dry-run proves a failing SHA cannot reach candidate build, an unmanifested byte/digest cannot publish, and a clean runner verifies `gate-results.json`, `release-manifest.json`, and the promoted artifacts. No stability or premier-framework release is allowed until all remaining P0 defects assigned to Waves 2–3 are also closed and regression-tested.

### Wave 1 — Compile and seal the application model

1. Define the versioned manifest and immutable internal `ApplicationModel`.
2. Add owner-bound registrars and read-only snapshots around existing registries.
3. Add typed port keys and provider graph while retaining a legacy adapter.
4. Compile API/worker/migrate profiles from one graph.
5. Derive lifecycle lint and model inspection from that graph.
6. Reject duplicate collectors, unknown module config namespaces, unowned keys, missing providers, cycles, and post-seal writes.
7. Add model hash to logs/readiness and a deterministic JSON export.

**Compatibility strategy:** introduce the new declaration surface in additive v1 packages alongside the current `module.Module`; “v2 declaration” describes the DSL generation, not a premature module-path break. Adapt legacy modules into a restricted legacy node and ship an automated inspection/migration command. Do not widen or remove the current v1 interfaces. Removal happens only in the future `/v2` Go module after the DX-05 support/deprecation window and golden-consumer migration pass.

**Exit gate:** all built-in/reference modules compile through the model; adversarial ownership/lifecycle tests pass; no registry backing map is externally mutable; all process profiles are graph-validated.

### Wave 2 — Establish authoritative identity and tenant data invariants

1. Implement principal/tenant/capacity resolution and privileged-session records.
2. Add assurance freshness and server-side break-glass/impersonation checks.
3. Inventory all tenant-local FKs and migrate to composites.
4. Integrate resource mirror/audit/outbox into aggregate writes.
5. Complete relationship subject semantics and actor attribution.
6. Generate tenancy catalog checks and negative test matrices.
7. Execute DATA-09 expand/backfill/validate/canary/switch choreography with N/N-1 compatibility and application rollback drills.

**Exit gate:** a database catalog query proves every tenant-local FK includes tenant equality or has an explicit reviewed exemption; every user request proves active membership; privileged sessions can be revoked and expire independently of JWT lifetime. N and N-1 binaries pass against the expanded schema, interrupted backfills resume, validation artifacts show zero mismatches, canary rollback leaves the additive schema usable, and no contract migration runs while an N-1 process/producer/consumer remains.

### Wave 3 — Standardize durable execution

1. Implement shared lease/generation/finalization primitives.
2. Migrate jobs, notifications, webhooks, bulk items, and the complete outbox claim/dispatch/finalize boundary to short leased transactions.
3. Move every provider/network call outside transactions.
4. Add mandatory effect idempotency/inbox APIs, domain CAS helpers, and provider idempotency adapters. Each worker/effect declaration selects one mechanism at model compilation.
5. Add per-tenant fairness, concurrency, backpressure, queue-age metrics, and operator replay.
6. Add deterministic crash/partition/duplicate-worker tests.

**Exit gate:** every worker documents and tests crash points; stale workers cannot finalize leases or begin a new effect after lease loss; cancellation propagates on lease loss. Every domain effect is protected by an atomic effect ledger or domain CAS. Every external effect reuses one stable idempotency key across retries; an adapter whose provider lacks idempotency may not be used for a non-idempotent high-impact operation and must pass an explicit duplicate-safety contract. A test pauses worker A through lease expiry, lets worker B reclaim, resumes A at every effect boundary, and proves at most one logical effect. No provider call occurs while a DB transaction is open. Merely documenting possible duplicate effects does not satisfy this gate.

### Wave 4 — Build the operation DSL and golden product

1. Add typed operation/resource/event/job/rule/workflow declarations.
2. Derive route metadata, schemas, OpenAPI, catalogs, and contract tests.
3. Replace any disabled/minimal CRUD scaffolding with the complete vertical slice specified by DX-02.
4. Build the two-module golden consumer and previous-version upgrade matrix.
5. Generate command/API docs and execute examples.
6. Add API/config/event compatibility gates.

**Exit gate:** one declaration change produces deterministic, reviewable diffs across runtime model, OpenAPI, permissions, and tests; no duplicate hand-maintained identity is needed; the golden product survives upgrade and rollback drills.

### Wave 5 — Optimize from production-shaped evidence

1. Add real-Postgres end-to-end benchmarks and query-plan fixtures.
2. Replace per-ancestor rule lookups with set-based resolution.
3. Bound/shard caches and rate-limit maps.
4. Batch sweepers and retry workers; add matching indexes.
5. Optimize the already-leased outbox's batch/query/index behavior without changing its Wave 3 correctness contract.
6. Remove S3 full-read checksum fallback from normal paths.
7. Establish SLOs and capacity envelopes.

**Exit gate:** budgets cover complete request/worker paths, missing measurements fail, and each optimization has before/after evidence without weakening security invariants.

### Wave 6 — Complete evidence, operations, and premier release

1. Canonical integrity over every field in the declared immutable audit-row contract, plus external anchor verification.
2. Durable DSR export/delivery and central legal-hold enforcement.
3. Operator control plane and privileged replay/grant workflows.
4. ASVS/API-security verification matrix and independent penetration test.
5. Blocking supply-chain controls and verified provenance promotion.
6. Verify DX-05 public API/version-policy enforcement across the full compatibility matrix and complete the release-candidate soak.

**Exit gate:** all Must-Haves have owners and executable evidence; no Critical/High review finding is open; release artifacts are promoted from a verified SHA/digest; documentation and generated product behavior agree.

## 13. Required acceptance matrix

| Domain                 | Minimum proof before premier release                                                                                                                |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| Module model           | ownership attack tests, duplicate/missing/cycle/scope tests, seal/race tests, deterministic model hash                                              |
| HTTP/API               | generated operation contracts, object-authz negatives, malformed/oversize/filter/pagination/idempotency tests, OpenAPI validation/diff              |
| Identity               | issuer/audience/key rotation, tenant membership, capacity selection, zero tenant, revocation, assurance freshness, break-glass/impersonation expiry |
| Tenancy                | catalog-driven RLS and composite-FK proof, platform-role negatives, cache/async/storage cross-tenant tests                                          |
| Persistence            | migration upgrade/down drill, constraint validation, optimistic concurrency, resource mirror atomicity                                              |
| Jobs/outbox            | duplicate workers, stale lease, crash matrix, retry/DLQ, per-aggregate ordering, idempotent effect ledger                                           |
| Notifications/webhooks | signed canonical replay fields, provider idempotency, network-outside-tx assertion, receipt audit, SSRF/redirect/secret rotation                    |
| Compliance             | audit tamper tests including metadata, anchor verification, legal hold enforcement, DSR artifact lifecycle and access audit                         |
| Performance            | end-to-end DB benchmarks, query counts/plans, p50/p95/p99, cache/cardinality bounds, queue/provider latency                                         |
| Operations             | migration-current readiness, required adapter profile, config/model fingerprint drift, alerts/runbooks, safe replay                                 |
| DX                     | released and source CLI scaffold, clean build/migrate/boot, two-module golden app, exact-version upgrade                                            |
| Supply chain           | blocking scans, action lint, SBOM, signed provenance, post-publish verification, exact-SHA promotion                                                |

### 13.1 Work packages, owners, and evidence protocol

These IDs are directive work packages, not claims that matching tracker issues already exist. Create one epic per row before implementation. A work package cannot move to `in_progress` until a named human DRI and reviewer replace the role owner.

| Work package | Findings/scope                                                                  | Default accountable role          | Required evidence root                          |
| ------------ | ------------------------------------------------------------------------------- | --------------------------------- | ----------------------------------------------- |
| PF-ARCH      | AR-01 through AR-06, application model, ports, sealing, DSL                     | framework architecture lead       | `docs/implementation/evidence/premier/PF-ARCH/` |
| PF-SEC       | SEC-01 through SEC-06, identity, privileged sessions, security profile          | product-security lead             | `docs/implementation/evidence/premier/PF-SEC/`  |
| PF-DATA      | DATA-01 through DATA-09, migrations, jobs, outbox, delivery, compliance storage | data/reliability lead             | `docs/implementation/evidence/premier/PF-DATA/` |
| PF-DX        | DX-01 through DX-07, generators, golden consumer, docs, API contracts           | developer-experience lead         | `docs/implementation/evidence/premier/PF-DX/`   |
| PF-PERF      | PERF-01 through PERF-06, reference environment, SLOs, capacity                  | performance/SRE lead              | `docs/implementation/evidence/premier/PF-PERF/` |
| PF-REL       | REL-01 through REL-04, compatibility, security gates, publication               | release/security-engineering lead | `docs/implementation/evidence/premier/PF-REL/`  |

Each root contains a machine-readable `evidence.json` with work-package ID, source/base SHA, DRI/reviewer, requirement IDs, commands, start/end time, tool/image versions, exit status, artifact hashes, environment fingerprint, known skips, findings/waivers, and links to raw outputs. A schema and CI validator reject missing fields, unapproved skips, stale SHAs, or evidence generated before the relevant change. Human proof bundles summarize; they do not replace raw evidence.

The independent security assessment scope includes tenant/RLS/FK isolation, object/function authorization, identity and privileged sessions, module ownership, API keys/JWKS, SSRF/webhooks, async replay/idempotency, secrets/redaction, compliance evidence, resource exhaustion, generated product defaults, CI/release trust, and the deployed reference stack. Premier release requires zero open Critical or High findings. A Medium needs a named owner, compensating control, expiry no later than the next supported minor, and security-lead acceptance in `evidence.json`; Low findings enter the normal backlog.

Required operational evidence includes an alert/runbook catalog whose tests inject each alert condition, migration/rollback and backup/restore drills, queue/DLQ replay drills, privileged-grant revocation, IdP/JWKS/provider outage, storage loss/latency, and capacity exhaustion. Required post-publish evidence is produced by a proposed `scripts/validation/verify_release.sh <version> <source-sha>` running from a clean environment and verifying all identities/hashes/attestations specified in REL-01. The script and its golden failure tests are part of PF-REL; a prose checklist is not sufficient.

### 13.2 Finding-level closure contracts

A finding closes only when every proof in its row is attached to the applicable work-package evidence bundle at the implementation SHA. Passing a broad package suite, adding a type, or updating prose is not a substitute for the named adversarial, integration, or operational proof.

#### PF-ARCH

| Finding | Machine-closeable proof                                                                                                                                                                                                                                                             |
| ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| AR-01   | Adversarial modules fail to claim every foreign declaration class; retained registrars reject post-seal writes; snapshot-mutation and race tests pass; two identical compiles emit the same complete model hash.                                                                    |
| AR-02   | Compile-fail fixtures reject key-type misuse; boot fixtures reject duplicate providers, missing requirements, undeclared edges, cycles, and invalid scope/lifetime edges; a foreign registrar cannot provide an owned key; API/worker/migrate graphs derive from the same manifest. |
| AR-03   | A golden declaration delta deterministically changes the expected route, permission, resource, schema/OpenAPI, lifecycle, profile, test, and documentation projections; a lint fails on a hand-maintained duplicate identity or omitted projection.                                 |
| AR-04   | Negative fixtures reject unknown module namespaces, duplicate collectors, empty required fragments, last-writer replacement, and every post-seal write; a production profile with a required no-op/missing adapter fails readiness unless a policy-valid waiver is present.         |
| AR-05   | Every normative example compiles/runs against the supported version matrix; generated reference tables match the model export byte-for-byte; CI fails a deliberately stale blueprint/README surface.                                                                                |
| AR-06   | Sentinel dependency injection proves every rules/authz path uses the composed store; the constructor-boundary lint fails a forbidden runtime infrastructure constructor and allows only declared composition packages.                                                              |

#### PF-SEC

| Finding | Machine-closeable proof                                                                                                                                                                                                                                                          |
| ------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| SEC-01  | Integration negatives reject zero/foreign tenant, inactive or capacity-less membership, revoked capacity, stale assurance, expired/revoked break-glass, and mismatched impersonation grants even when signed claims request them; revocation meets its declared propagation SLO. |
| SEC-02  | Public runtime construction without authorization fails; denied override and missing ratification tests fail closed; successful override/ratification persists actor, impersonator, grant, reason, transition, and outcome evidence.                                             |
| SEC-03  | Tampering independently with body, timestamp, event ID, key ID, or signature version fails; replay/reorder and key-rotation matrices pass; body-only HMAC cannot populate unauthenticated provider replay fields.                                                                |
| SEC-04  | Cardinality never exceeds the configured bound under adversarial identities; concurrent misses collapse to one load; cross-pod revocation/epoch tests meet the stale-allow bound; cache metrics account for hits, fills, evictions, invalidations, and rejected admissions.      |
| SEC-05  | A version-pinned ASVS/API-security control map links every applicable requirement to an executable test or approved time-bounded waiver; the independent assessment leaves zero Critical/High findings.                                                                          |
| SEC-06  | Boot rejects untrusted or tenant-controlled egress exceptions; allowlist changes alter the config/model fingerprint and audit record; DNS rebinding, redirect, proxy, private-address, and injected-client tests preserve the declared exception boundary.                       |

#### PF-PERF

| Finding | Machine-closeable proof                                                                                                                                                                                                                                             |
| ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| PERF-01 | More than the hard-limit number of one-shot keys cannot grow the map beyond policy; fake-clock TTL sweep returns it below the bound; 10k/100k/boundary benchmarks and concurrent/fuzz cases pass with cardinality, eviction, rejection, and sweep metrics asserted. |
| PERF-02 | The pinned reference runner publishes raw public/read/write/resource-authz/idempotent/enqueue workloads with p50/p95/p99, allocations, SQL count, pool/transaction/lock wait, bytes, and plan hashes across cold/warm and 1/10/100-tenant concurrency.              |
| PERF-03 | Result-parity tests cover organization depth and active/historical fallback; SQL-count assertions stay constant with depth; representative `EXPLAIN (ANALYZE, BUFFERS)` artifacts use the intended indexes and meet the reference budget.                           |
| PERF-04 | Bounded-batch tests prove fixed query counts and memory across due-row cardinalities; reminder/retry queries use predicate-matching indexes; no outer transaction spans tenant handlers; queue-lag/batch metrics and cancellation tests pass.                       |
| PERF-05 | Framework uploads always persist and verify canonical checksum metadata; normal `Stat` performs no body download; legacy fallback enforces size/time bounds, emits dedicated metrics, and a resumable backfill eliminates repeated hashing.                         |
| PERF-06 | Removing/renaming a budgeted benchmark fails CI; PR and scheduled jobs execute time-bounded `-fuzz` runs and retain corpora; a seeded performance regression fails the statistical/absolute gate unless an independently reviewed budget change is attached.        |

#### PF-DATA

| Finding | Machine-closeable proof                                                                                                                                                                                                                                                                                             |
| ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| DATA-01 | A catalog query proves every tenant-local FK includes tenant equality or a named reviewed exemption; seeded cross-tenant parent/child inserts fail under both runtime and platform paths; backfill/validation reports zero existing mismatches.                                                                     |
| DATA-02 | Deterministic tests pause worker A, expire/reclaim with worker B, and resume A at every domain/external/finalize boundary; stale token/generation writes fail and exactly one logical effect is recorded through ledger, CAS, or provider idempotency.                                                              |
| DATA-03 | Transaction instrumentation proves no DNS/secret/provider call occurs while a DB transaction is open; outbound crash points and inbound secret rotation/deactivation between read/verify/write phases cannot duplicate or accept under stale policy.                                                                |
| DATA-04 | Two or more processors concurrently claim, retry, pause, resume, and cancel the same operation without duplicate item effects or stale finalization; SQL proves atomic leased `SKIP LOCKED` claims and bounded batches.                                                                                             |
| DATA-05 | High-concurrency allocation yields unique monotonic artifact/document versions without expected conflicts; failures at upload/reserve/confirm leave durable sessions; expiry cleanup removes every unreferenced object and never a referenced one.                                                                  |
| DATA-06 | Fault injection at aggregate, resource mirror, audit, and outbox stages proves all-or-nothing commit; generated repositories cannot omit the projection; user writes reject missing actor and preserve actor attribution.                                                                                           |
| DATA-07 | Capacity- and party-subject authorization matrices cover allow/deny/revocation; every relationship/resource mutation is owner-checked, attributed, audited, versioned, and invalidates all affected cache entries.                                                                                                  |
| DATA-08 | Mutating any field in the declared immutable audit contract breaks verification; a required outbox/audit failure rolls back attachment/domain writes; provider receipts are durable; DSR completion requires encrypted checksummed artifact/delivery evidence; hold tests block every erase/dispose path centrally. |
| DATA-09 | CI runs N-1 and N binaries against expanded schema, interrupted/resumed backfill, zero-mismatch validation, lock-timeout retry, canary switch, application rollback without destructive Down, and delayed contract only after old-process/producer/consumer absence is proven.                                      |

#### PF-DX

| Finding | Machine-closeable proof                                                                                                                                                                                                                                                                                                                       |
| ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| DX-01   | Released and source-built CLIs generate isolated projects whose framework version resolves to the intended artifact/source, then download, build, test, migrate, boot, and smoke successfully; an unavailable version fails before writing files with an exact remediation command.                                                           |
| DX-02   | Wave 0 proves the command cannot emit authorization-invalid or false-success CRUD; final closure compiles/boots the generated vertical slice and passes permission, validation, RLS, object-authz, idempotency, ETag/concurrency, pagination, audit/outbox/resource, migration, and OpenAPI contracts.                                        |
| DX-03   | Compiler negatives cover ownership, credential/tenant incompatibility, missing object target, duplicate schema/error/event identity, incompatible event change, invalid sync/async/stream policy, unbounded stream, and provider-graph errors; generated projections are deterministic and dispatch profiles show no request-path reflection. |
| DX-04   | The clean two-module golden product exercises every declared subsystem, upgrades from each supported previous version, rolls the application back on expanded schema, and runs against released artifacts rather than repository-private imports.                                                                                             |
| DX-05   | README, policy, changelog, tags, CLI version, generated `go.mod`, module metadata, and support matrix agree; public/API/config/event diffs permit supported additive changes and reject an intentional v1 breaking fixture.                                                                                                                   |
| DX-06   | Merge fixtures preserve or deliberately reject every OpenAPI 3.1 top-level field, reference, security declaration, parameter, response, callback, and webhook; structural validation and semantic diff fail malformed or breaking output.                                                                                                     |
| DX-07   | Readiness fails on stale migration/model/seed/rule state or a missing required adapter; `config doctor` finds the product independently of CWD and proves product validation ran; production rejects undeclared capacity/backpressure posture.                                                                                                |

#### PF-REL

| Finding | Machine-closeable proof                                                                                                                                                                                                                                                                                                                 |
| ------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| REL-01  | A failing exact-tag-SHA gate prevents candidate build; altered gate results, tag target, manifest, archive, or image digest are rejected; publish accepts only candidate bytes named by the attested manifest; clean-runner verification matches source SHA, signatures, SBOM/provenance, platforms, CLI version, and published hashes. |
| REL-02  | Private-repository CI executes blocking replacements for unavailable hosted scanners; seeded secret, reachable vulnerable dependency, disallowed license, workflow defect, and Critical/High image/config finding each fail; every waiver has owner, rationale, expiry, and remediation link.                                           |
| REL-03  | Intentional breaking Go/OpenAPI/config/event fixtures fail; oldest-supported-version upgrade and migration compatibility pass; generated consumers compile across the support window; every published architecture boots and verifies.                                                                                                  |
| REL-04  | A machine suite manifest proves required DB, MinIO/S3, E2E, race, and fuzz suites executed with zero unauthorized skips; toolbox MinIO health plus canonical endpoint wiring pass; TOTP is deterministic; fuzz artifacts prove non-zero time beyond seed execution.                                                                     |

## 14. Measurement budgets and operating envelopes

Avoid universal numbers that are disconnected from a deployment. Establish a dedicated Linux amd64 reference runner and commit `perf/reference-v1.json` plus its workload fixtures. The artifact records CPU/runner image digest, Go version, Postgres image/config, pool size, loopback/network shaping, dataset cardinality, tenant distribution, object sizes, provider latency model, workload seed, warm-up/measurement durations, and every absolute ceiling. Changing environment metadata or a ceiling is a separately reviewed performance-policy change, not an incidental code diff.

The initial baseline procedure is machine-defined:

1. Build the exact commit with a clean module/build cache policy recorded in the artifact.
2. Seed the fixed reference dataset and verify its checksum.
3. Run Go microbenchmarks 10 times and compare with `benchstat` at alpha 0.01.
4. Warm each end-to-end workload for 5 minutes, measure for 15 minutes, and repeat three times; use the median run for p50/p95/p99 and the worst run for errors/resource ceilings.
5. Record raw benchmark/load-test output, SQL/query-plan hashes, profiles, and reference JSON as CI artifacts addressed by source SHA.

Until the first reference artifact is approved, performance work is evidence-gathering and cannot claim the premier-release performance gate. After approval, a PR fails on any statistically significant microbenchmark regression greater than 10% ns/op, any increase over one allocation/op unless explicitly budgeted, any end-to-end p95 regression greater than 10%, p99 regression greater than 15%, SQL-statement count increase, error rate above 0.01%, or violation of an absolute resource/SLO ceiling. A flaky result reruns the complete workload once; a second failure is final. Budget changes require before/after profiles, explanation, and performance-owner approval.

At minimum, the artifact gates:

- request latency and allocations by public/read/write/resource-authz/idempotent profiles;
- SQL statements, pool wait, and transaction duration per request;
- rules resolution by org depth/history;
- authz cache hit/miss/fill/eviction and revocation propagation;
- rate-limit entry count and sweep/admission cost;
- job/outbox/notification/webhook queue age, claim time, execution time, finalize time, retries, lease expiry, DLQ rate;
- workflow SLA sweep by due-row count;
- audit write and chain verification by tenant concurrency;
- document confirmation by object size/checksum path;
- memory under adversarial tenant/actor/endpoint/route cardinality.

Missing workloads, samples, environment fields, or budget entries fail. B11 reopens only if route-count growth is both statistically and practically non-flat—initially, more than a 15% median increase from 50 to 2,000 routes in two consecutive reference-run comparisons, or a fitted curve that breaches the route budget at the supported maximum—or if dispatch exceeds 10% of measured p95 authenticated-request latency. The current 9% host delta does not meet that trigger. A reopened decision uses a profile before selecting a new router.

## 15. Explicit non-goals and parked work

- **No router rewrite now.** Current dispatch is flat through 2,000 routes and dwarfed by real authz/DB work.
- **No general hot-reload registry.** Immutable boot snapshots are a safety property. Rules already resolve active versions per request. If an operational i18n need is later demonstrated, implement validated copy-on-write snapshot replacement with version/audit/rollback, not mutation of the live map.
- **No standalone schema generator before the application model.** OpenAPI/schema derivation belongs to typed operations. B12's original module-count/drift trigger remains unmet.
- **No ORM or reflection DI container.** Preserve explicit SQL, transaction boundaries, and Go constructors. Compilation may use reflection/AST/codegen at build/boot time; request hot paths should use compiled descriptors.
- **No in-process untrusted plugin claim.** Go modules linked into the process are trusted code. A future untrusted extension system requires process isolation or a sandboxed ABI.
- **No “exactly once” marketing.** Database transitions can be fenced; external effects are at-least-once unless the receiver provides idempotency/transactional guarantees.
- **No security-through-no-op adapters.** Local convenience defaults must not silently become production posture.

## 16. Standards baseline

The implementation program should pin standards versions in its verification matrix and review upgrades deliberately:

- [OWASP ASVS 5.0.0](https://owasp.org/www-project-application-security-verification-standard/) for application security requirements.
- [OWASP API Security Top 10 (2023)](https://owasp.org/API-Security/) for API threat coverage.
- [NIST SP 800-63-4](https://pages.nist.gov/800-63-4/) for digital identity and assurance.
- [OpenAPI 3.1.1](https://spec.openapis.org/oas/v3.1.1.html) and [JSON Schema 2020-12](https://json-schema.org/draft/2020-12) for API/schema contracts.
- [OpenTelemetry semantic conventions](https://opentelemetry.io/docs/specs/semconv/) for telemetry names and meanings.
- [SLSA 1.2](https://slsa.dev/spec/v1.2/) for supply-chain provenance goals.
- [Go fuzzing](https://go.dev/doc/security/fuzz/) for actual coverage-guided fuzz execution.
- [Go module versioning](https://go.dev/doc/modules/version-numbers) for dependency/release semantics.

Standards are a baseline and vocabulary. The repository must retain concrete threat models, tests, and operating evidence for its own tenant, workflow, evidence, and async semantics.

## 17. Final architectural decision

wowapi should evolve into a compiled modular-monolith framework, not accumulate more shared registries and manually synchronized metadata. The future framework contract is:

1. **Declare once:** owned, typed module and operation contracts.
2. **Compile once:** dependency, scope, security, schema, tenancy, lifecycle, and capability validation.
3. **Seal once:** immutable runtime model and deterministic fingerprint.
4. **Execute safely:** authoritative principals, tenant-bound transactions, fenced durable work, bounded memory/concurrency, external I/O outside transactions.
5. **Derive consistently:** runtime routes, catalogs, OpenAPI, lifecycle graph, conformance tests, docs, and operational manifest.
6. **Prove continuously:** adversarial tests, real-DB performance, compatibility matrices, blocking security gates, and verified release provenance.

That program addresses the same root failure pattern that made historical i18n work manual: one concern represented in too many places, mutable ownership not encoded in the API, wiring not derived, and completion inferred from component presence rather than end-to-end use. Fixing that pattern at the application-model boundary will prevent the next generation of i18n-like refactors across permissions, resources, rules, workflows, events, jobs, OpenAPI, lifecycle, configuration, and generated products.
