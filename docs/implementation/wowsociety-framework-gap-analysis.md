<!-- markdownlint-disable MD013 MD024 -->

# GAP Analysis: wowsociety Framework Utilization

Date: 2026-07-09

Framework repository: `/Users/qatoolist/go_home/src/github.com/qatoolist/wowapi`

Product repository used as evidence: `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety`

Framework commit evaluated: `287abc38df8a`

Product commit inspected: `6ac94049ee73`

Post-implementation design review: after the gap-closure branch and the `wowsociety`
consumption branch were built, a second review found that some work closed the
observed product workaround without fully designing the framework capability. See
[`wowsociety-framework-gap-design-review.md`](wowsociety-framework-gap-design-review.md),
especially the i18n source/loading/tooling follow-up.

Competitive architecture benchmark: the follow-up review was broadened into an
RFC-style benchmark against Laravel, Spring Boot/Security, Gin, FastAPI, Django,
and Axum/Tower. See
[`framework-competitive-architecture-benchmark.md`](framework-competitive-architecture-benchmark.md)
for the comparative matrix, low-level deep dives, and prioritized engineering
backlog.

## Purpose

This report identifies components that `wowsociety` had to implement product-side but that should be provided out of the box by `wowapi` as framework capabilities, production adapters, lifecycle hooks, or scaffolded integration code.

The goal is twofold:

1. Implement or move the generic capabilities from `wowsociety` into `wowapi`.
2. Change `wowsociety` to consume those framework capabilities instead of carrying local implementations and workarounds.

This report intentionally separates framework gaps from domain-specific product logic. Maharashtra housing society policy content, domain roles, citations, committee semantics, and product workflows should remain in `wowsociety`. Generic infrastructure, adapters, security plumbing, lifecycle hooks, and reusable framework services should move into `wowapi`.

## Framework Utilization Snapshot

`wowsociety` consumes `wowapi` as its framework dependency:

- `wowsociety/go.mod` requires `github.com/qatoolist/wowapi v1.0.0`.
- `wowsociety/go.mod` replaces `github.com/qatoolist/wowapi => ../wowapi`.
- `wowsociety/FRAMEWORK_VERSION` pins the consumed framework commit to `287abc3`, matching the local `wowapi` checkout inspected here.
- `wowsociety` directly depends on `github.com/minio/minio-go/v7`, largely because the framework has a storage port but no production S3/MinIO adapter.

Relevant product areas inspected:

- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/i18n`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/adapters/storage/s3`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/appcfg`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/identity`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/policy`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/cmd/api`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/cmd/migrate`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/cmd/worker`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/tools/configcheck`

## Executive Summary

The product had to implement several generic framework concerns locally:

| ID | Gap | Severity | Product workaround | Framework target |
| --- | --- | --- | --- | --- |
| GAP-001 | Locale negotiation and API i18n | P0 | `internal/i18n` catalog and manual `Content-Language` handling | `kernel/i18n`, locale-aware `httpx`, localized validation/errors |
| GAP-002 | Production object storage adapter | P0 | `internal/adapters/storage/s3` | `adapters/storage/s3` plus scaffold/config wiring |
| GAP-003 | Production seed synchronization lifecycle | P0 | Manual `seeds.Sync` in product migrate command | Generated migrate/CLI seed sync path |
| GAP-004 | Step-up/MFA seedability and AMR propagation | P0 | Direct permission registration plus JWT reparse wrapper | `PermissionSeed.StepUp`, `auth.Claims.AMR`, testkit support |
| GAP-005 | MFA factor primitives and sender ports | P1 | Product TOTP, OTP hashing, SMS port | Reusable `kernel/mfa` or `kernel/authn` primitives and ports |
| GAP-006 | Scoped privileged framework service APIs | P0 | Product `SECURITY DEFINER` SQL bridges | Module-safe ReBAC edge and rules activation services |
| GAP-007 | Rules registry lifecycle and schema validation | P1 | SQL mirror of rule definitions and product bounds validation | Registry-to-table sync and fuller schema enforcement |
| GAP-008 | Product-aware scaffold/config tooling | P1 | Product `tools/configcheck` and composition boilerplate | Generated product config tooling and standard adapter wiring |

## GAP-001: Locale Negotiation And API i18n

### Current product implementation

`wowsociety` implements an in-process message catalog and locale negotiation:

- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/i18n/catalog.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/i18n/negotiate.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/i18n/messages.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/cmd/api/main.go`

The package comment in `internal/i18n/catalog.go` states that the framework has no `kernel/i18n`, no locale-aware HTTP response helper, and only notification-template locale fallback for async notification bodies. The product also wires locale negotiation and `Content-Language` manually in the API process.

### Framework evidence

Framework API errors are English-only at the response layer:

- `kernel/httpx/errors.go` hardcodes problem-detail titles in `titles`.
- `kernel/httpx/errors.go` copies `errors.Error.Msg` directly into `ProblemError.Detail`.
- `kernel/validation/validation.go` hardcodes field validation messages in `messageForTag`.

The only locale-aware framework mechanism found is notification-template lookup, which is not usable for synchronous API responses.

### Why this belongs in `wowapi`

i18n is a cross-cutting API concern. Every product built on the framework will need consistent handling of:

- `Accept-Language` parsing and fallback.
- Response `Content-Language`.
- Localized problem-detail titles and details.
- Localized field validation messages.
- Stable machine codes independent of translated user text.
- Testkit helpers for asserting localization behavior.

Leaving this to each product creates inconsistent API behavior and duplicates risky translation plumbing.

### Required framework implementation

Implement:

- `kernel/i18n` with `Catalog`, `Bundle`, `Lookup`, fallback policy, and supported locale registration.
- `kernel/httpx` locale middleware that parses `Accept-Language`, stores locale in request context, and sets `Content-Language`.
- Locale-aware `httpx.WriteError` or an injectable/localized problem renderer.
- Locale-aware validation message provider in `kernel/validation`.
- A standard catalog loading convention for framework messages and module/product messages.
- Testkit helpers for locale negotiation and localized error assertions.

Keep:

- Machine error codes stable and untranslated.
- Internal logs in stable technical English.
- Product-specific translations in product-owned catalogs.

### `wowsociety` migration after framework support

Remove or shrink:

- `internal/i18n/catalog.go`
- `internal/i18n/negotiate.go`
- Product-owned response-localization wiring in `cmd/api/main.go`

Replace with:

- Framework locale middleware.
- Product message bundles registered through framework i18n APIs.
- Localized `httpx` and validation responses.

### Acceptance criteria

- `Accept-Language: mr-IN,mr;q=0.9,en;q=0.8` resolves to Marathi when the catalog supports it.
- Unsupported locales fall back deterministically to the default locale.
- Problem-details responses include localized `title` and `detail` while preserving stable `code`.
- Validation field messages localize without changing `field` or `code`.
- `wowsociety` no longer owns request-locale parsing.

## GAP-002: Production Object Storage Adapter

### Current product implementation

`wowsociety` implements a production S3/MinIO adapter:

- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/adapters/storage/s3/s3.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/appcfg/config.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/cmd/api/main.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/cmd/worker/main.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/docs/rff/RFF-001-object-storage-adapter.md`

The adapter comment says it is a product-side RFF-001 workaround because `wowapi` ships only `storage.NewMemory()`. It implements `storage.Adapter` against S3-compatible endpoints, including presigned PUT/GET, `Stat`, `Peek`, and idempotent `Delete`.

### Framework evidence

`kernel/storage/storage.go` defines the framework object-storage port and explicitly describes production S3/MinIO semantics. `kernel/storage/memory.go` only provides an in-memory adapter for tests and local development.

No production object-storage adapter exists in `wowapi/adapters`.

### Why this belongs in `wowapi`

The storage adapter is domain-neutral. The framework owns the document storage port and document service semantics, so it should also provide at least one production adapter that satisfies that port.

Without it, every product must reimplement the same S3 edge cases:

- Path-style versus virtual-host addressing.
- Presign TTL bounds.
- Bucket existence validation.
- Missing object mapping to `KindNotFound`.
- S3 checksum behavior for plain presigned uploads.
- Ranged GET for MIME sniffing.
- Idempotent deletes.

### Required framework implementation

Move or reimplement `wowsociety/internal/adapters/storage/s3` into `wowapi`, likely under:

- `adapters/storage/s3`

Add:

- Adapter config struct with endpoint, bucket, region, credentials, TLS, presign TTL, and optional dev bucket creation.
- Unit tests using the memory adapter semantics as the contract.
- Integration tests gated behind env vars or local MinIO.
- Documentation in the user guide.
- Generated scaffold wiring for API and worker processes.

Decide whether the framework's base `config.Framework` should include a standard optional storage section, or whether scaffolded product config should embed an adapter-specific config block.

### `wowsociety` migration after framework support

Remove:

- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/adapters/storage/s3`
- Direct `github.com/minio/minio-go/v7` dependency from product `go.mod` if no longer used elsewhere.

Replace with:

- `github.com/qatoolist/wowapi/adapters/storage/s3`
- Framework-generated config and wiring where possible.

### Acceptance criteria

- A new product can enable S3/MinIO document storage without writing adapter code.
- `wowsociety` uses the framework adapter unchanged.
- Document upload confirm behavior remains checksum-compatible with current product behavior.
- Missing objects map to `KindNotFound`.
- Local MinIO can be used in development without product-specific adapter code.

## GAP-003: Production Seed Synchronization Lifecycle

### Current product implementation

`wowsociety` manually calls `seeds.Sync` from its migrate command:

- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/cmd/migrate/main.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/docs/PILOT-FINDINGS.md`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/docs/upstream/02-pf-9-no-production-seed-sync-path.md`

The product comment states that generated mains omit seed sync and that without this step the DB catalogs stay empty, causing authorization denial and resource foreign-key failures.

### Framework evidence

`kernel/seeds/seeds.go` implements `seeds.Sync`, and it must run on a platform-privileged connection. The framework also loads module seeds during boot and registers seed-declared permissions into the in-memory registry in `app/boot.go`.

The missing piece is the production lifecycle path that syncs the merged seed bundle into the database.

### Why this belongs in `wowapi`

Seeds define framework authorization and resource catalogs. A product should not need to remember a framework-internal lifecycle requirement in its own migrate command.

The current split is unsafe:

- Boot loads seeds into memory.
- DB migrations create tables.
- But generated production commands do not guarantee DB catalog sync.

This creates a deployment path where the app boots against empty authorization catalog tables.

### Required framework implementation

Implement one or both:

- Generated `cmd/migrate` templates that run kernel migrations, module migrations, then `seeds.Sync`.
- A first-class CLI command such as `wowapi seed sync` or `wowapi migrate --sync-seeds`.

Also add:

- Idempotency tests.
- Cache invalidation support where `authz.CachingStore` is active.
- Documentation that seed sync is part of the production deploy lifecycle.
- A failure mode that clearly reports missing seed sync instead of surfacing as scattered authorization failures.

### `wowsociety` migration after framework support

Replace the manual seed-sync block in `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/cmd/migrate/main.go` with the generated or framework-provided lifecycle call.

### Acceptance criteria

- Running the generated migrate command fully prepares framework catalog tables.
- A fresh database has permissions, roles, resource types, and relationship types after migration.
- Re-running migration and seed sync is idempotent.
- `wowsociety` no longer carries custom comments or code for the seed lifecycle workaround.

## GAP-004: Step-up/MFA Seedability And AMR Propagation

### Current product implementation

`wowsociety` works around two separate framework gaps.

First, step-up permissions cannot be declared in seed YAML:

- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/identity/seeds/permissions.yaml`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/identity/module.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/identity/migrations/00004_stepup_permission.sql`

The product registers `identity.impersonation.assign` directly with `StepUp: true` and inserts the DB catalog row manually because `seeds.PermissionSeed` has no `step_up` field.

Second, the JWT `amr` claim is not propagated into `authz.Actor.AMR`:

- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/identity/authenticator.go`

The product wraps the framework authenticator and reparses the already-validated JWT to recover `amr`, adding an extra signature verification per authenticated request.

### Framework evidence

The framework already has the core step-up model:

- `kernel/authz/authz.go` has `Actor.AMR`.
- `kernel/authz/registry.go` has `Permission.StepUp`.
- `kernel/authz/evaluator.go` checks strong factors for `StepUp` permissions.
- `kernel/httpx/authz_gate.go` emits the step-up challenge.

But the framework does not connect all pieces:

- `kernel/seeds/seeds.go` has no `PermissionSeed.StepUp`.
- `app/boot.go` registers seed permissions with key, sensitivity, and `granted_via`, but not step-up.
- `migrations/00006_authz.sql` stores `permissions.sensitive`, but not `step_up`.
- `kernel/auth/auth.go` `Claims` has no `AMR`.
- `kernel/auth/auth.go` `Verifier.Actor` does not set `Actor.AMR`.
- `testkit/auth.go` has no `WithAMR` token option.

### Why this belongs in `wowapi`

Step-up is a framework authorization capability. A product should not need to:

- Register step-up permissions outside the seed catalog.
- Manually insert permission catalog rows.
- Reparse JWTs to move `amr` from authentication to authorization.
- Duplicate bearer token extraction.

The current state makes step-up appear supported but leaves product teams to connect security-critical plumbing themselves.

### Required framework implementation

Implement:

- `StepUp bool` on `seeds.PermissionSeed`.
- Seed strict decoding and validation for `step_up`.
- `app.Boot` propagation from `PermissionSeed.StepUp` to `authz.Permission.StepUp`.
- Database support if the permission catalog is intended to persist this flag. Add a migration for `permissions.step_up boolean NOT NULL DEFAULT false`, and update `seeds.Sync`.
- `AMR []string` on `auth.Claims`.
- `Verifier.Actor` propagation from `Claims.AMR` to `authz.Actor.AMR`.
- `testkit.WithAMR(...string)` and auth tests proving step-up works through issued JWTs.

### `wowsociety` migration after framework support

Remove:

- Direct step-up permission registration in `identity/module.go`.
- Manual step-up catalog migration if framework DB schema persists `step_up`.
- `internal/modules/identity/authenticator.go` AMR wrapper.

Replace with:

- `step_up: true` in identity permission seed YAML.
- Framework authenticator that surfaces `Actor.AMR`.
- Testkit-issued tokens with AMR values.

### Acceptance criteria

- A seed-declared permission can require step-up.
- A JWT with `amr: ["pwd","mfa"]` produces an actor with matching `Actor.AMR`.
- A JWT without a strong factor receives a step-up HTTP challenge for step-up permissions.
- `wowsociety` no longer reparses bearer tokens.

## GAP-005: MFA Factor Primitives And Sender Ports

### Current product implementation

`wowsociety` implements MFA factor mechanics locally:

- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/identity/totp.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/identity/otp.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/identity/sms.go`

The product implements:

- RFC 6238 TOTP generation and verification.
- HOTP calculation.
- Phone OTP code generation.
- Salted OTP hashing.
- Expiry and attempt enforcement.
- SMS delivery behind a product port.

### Framework evidence

The framework has step-up authorization semantics but no reusable TOTP/HOTP/OTP primitives or delivery port. The product code comments explicitly state that no TOTP/HOTP capability exists at this framework pin.

### Why this may belong in `wowapi`

This should move into the framework if `wowapi` intends to provide out-of-box MFA/step-up support, not merely consume an external IdP's AMR claim.

The framework does not need to own every identity workflow, but it should provide reusable primitives so products do not implement crypto-sensitive MFA mechanics from scratch.

### Required framework implementation

Implement a small reusable package, for example:

- `kernel/mfa`
- or `kernel/authn/mfa`

Capabilities:

- TOTP secret generation.
- TOTP verification with configurable step, digits, algorithm, and skew window.
- HOTP primitive if needed.
- Numeric OTP generation.
- OTP hash/verify helpers with constant-time comparison.
- Challenge policy helpers for TTL and attempt limits.
- Sender interfaces for SMS/email delivery, with test/log adapters.

Keep product-owned:

- Enrollment UX.
- Factor storage schema if product-specific.
- Delivery provider selection.
- Which actions require which factor.

### `wowsociety` migration after framework support

Replace local TOTP and OTP helper implementations with framework primitives. Keep the product service and persistence model unless the framework intentionally adds a full identity/MFA module.

### Acceptance criteria

- `wowsociety` no longer owns raw TOTP/HOTP implementation.
- MFA tests use framework primitives.
- Security-sensitive comparisons remain constant-time.
- The framework has test vectors for RFC 4226/6238 behavior.

## GAP-006: Scoped Privileged Framework Service APIs

### Current product implementation

`wowsociety` creates `SECURITY DEFINER` SQL bridges because modules cannot use a sanctioned framework service for valid privileged operations:

- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/identity/migrations/00003_committee_seat.sql`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/policy/migrations/00003_activation.sql`

The identity module uses a bridge to grant/revoke ReBAC relationship edges. The policy module uses a bridge to activate rule versions.

### Framework evidence

The framework deliberately protects the relevant tables:

- `migrations/00005_resource_relationship.sql` grants app runtime only `SELECT` on `relationships`; writes are platform-only.
- `kernel/database/txmanager.go` keeps platform access internal.
- `module.Context` exposes tenant transaction access but no platform write door.
- `kernel/rules/store.go` requires platform privilege for `rules.Store.Activate`.

The security posture is correct. The gap is the missing framework service surface that lets modules perform valid tenant-scoped privileged operations without writing their own `SECURITY DEFINER` functions.

### Why this belongs in `wowapi`

Relationship edges and rule activation are framework concepts. The framework should provide narrowly scoped, audited services that preserve its own invariants.

Product-authored `SECURITY DEFINER` bridges are risky because each product must re-implement:

- Tenant binding checks.
- Resource and capacity existence checks.
- Relationship-type ownership.
- Scope restrictions.
- Audit propagation.
- Race behavior.
- Permission boundaries.

### Required framework implementation

Expose module-safe services through `module.Context`, for example:

- `mc.Relationships().Grant(ctx, spec)`
- `mc.Relationships().Revoke(ctx, id, actor)`
- `mc.Rules().ActivateTenant(ctx, versionID, approvedBy, options)`

The services should:

- Run with controlled platform privilege internally.
- Require the caller's tenant context.
- Restrict relationship types and rule keys to the owning module or declared allow-list.
- Validate subject/object resource existence.
- Write audit metadata.
- Preserve existing table protections.
- Be testable through testkit without product SQL bridges.

### `wowsociety` migration after framework support

Remove:

- `identity_grant_committee_seat`
- `identity_revoke_committee_seat`
- `policy_activate_rule_version`

Replace product SQL calls with framework service calls from module code.

### Acceptance criteria

- Modules can create/revoke owned relationship edges without direct app runtime writes to `relationships`.
- Modules can activate owned tenant-scope rule versions through a framework API.
- No product-owned `SECURITY DEFINER` functions are needed for these framework concepts.
- Tests prove tenant isolation and ownership checks.

## GAP-007: Rules Registry Lifecycle And Schema Validation

### Current product implementation

`wowsociety` mirrors rule definitions into SQL manually and adds product-side bounds validation:

- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/policy/migrations/00004_mh_pack_v1.sql`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/policy/rulemirror_test.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/modules/policy/rulepoints.go`

The product migration says framework docs describe registry sync to `rule_definitions`, but no boot path performs it. Because `rule_versions.rule_key` references `rule_definitions`, product code must mirror the Go registry declarations in SQL. A drift guard test then compares the Go registry and the SQL mirror.

The product also enforces numeric minimum/maximum bounds because the framework's schema validator only checks top-level type and enum.

### Framework evidence

- `kernel/rules/store.go` validates values through `validateAgainstSchema` on propose.
- `kernel/rules/schema.go` says the validator is focused and enforces only top-level type and enum.
- No inspected framework lifecycle path syncs registered rule points into `rule_definitions`.

### Why this belongs in `wowapi`

The rules registry and rules database schema are framework-owned. Products should register rule points once in Go and let the framework persist or sync definitions consistently.

Duplicating rule declarations in SQL causes:

- Drift risk between registry and DB.
- More product migrations for framework metadata.
- Repeated tests to guard framework lifecycle gaps.
- Product-side validation duplication.

### Required framework implementation

Implement:

- `rules.SyncDefinitions(ctx, db, registry)` or equivalent lifecycle hook.
- Generated migrate/seed path that syncs rule definitions after migrations and before rule versions are inserted.
- Ownership/module checks for rule keys.
- Idempotent updates for schema, default value, allowed scopes, approval requirement, and description.
- Expanded schema validation for common JSON Schema keywords used by rule points:
  - `minimum`
  - `maximum`
  - `exclusiveMinimum`
  - `exclusiveMaximum`
  - `minLength`
  - `maxLength`
  - `pattern`
  - `minItems`
  - `maxItems`
  - `required`
  - object property schemas, if the framework advertises object schema support.

If full JSON Schema is out of scope, explicitly narrow the framework contract and provide first-class typed bounds fields instead of accepting broader JSON Schema documents.

### `wowsociety` migration after framework support

Remove:

- Manual `rule_definitions` insert block from `00004_mh_pack_v1.sql`.
- `rulemirror_test.go`, or reduce it to a framework-level contract test.
- Product-side duplicate bounds validation where the framework schema can enforce it.

Keep:

- Maharashtra rule definitions and statutory citations as product/domain data.
- Product-specific citation verification and activation gates.

### Acceptance criteria

- Registering a rule point in Go creates or updates its `rule_definitions` row through framework lifecycle code.
- `rules.Store.Propose` rejects out-of-bounds numeric values when schema declares `minimum` or `maximum`.
- `wowsociety` rule declarations are not duplicated in SQL.

## GAP-008: Product-Aware Scaffold, Config, And Deployment Tooling

### Current product implementation

`wowsociety` owns a config checker and substantial process composition code:

- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/tools/configcheck/main.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/internal/appcfg/config.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/cmd/api/main.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/cmd/worker/main.go`
- `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety/cmd/migrate/main.go`

The config checker exists because the prebuilt `wowapi` CLI cannot link the composed product config. Product composition also wires standard concerns such as storage, OIDC/JWT auth, API keys, OpenTelemetry, Prometheus, and seed sync.

### Why this belongs partly in `wowapi`

Some product composition must remain product-owned. However, framework-generated scaffold should cover standard framework concerns so each product does not rebuild the same process shell.

The framework should generate or provide:

- Product-aware config validation command.
- Config schema generation for composed product config.
- Standard environment overlay loading.
- Standard adapter wiring points.
- Standard storage/auth/seed/rules lifecycle hooks.
- Deployment posture checks.

### Required framework implementation

Implement scaffold support that can generate product-local code but keep the template maintained by `wowapi`:

- `cmd/api` template with standard adapter hooks.
- `cmd/worker` template with standard adapter hooks.
- `cmd/migrate` template with migrations, seed sync, and rule-definition sync.
- `tools/configcheck` template parameterized by the product config type.
- Optional generated `internal/appcfg` helpers that embed `config.Framework` and add product namespaces cleanly.
- Documentation for the expected generated/custom boundary.

This does not mean `wowapi` must know every product's config. It means `wowapi` should generate the boilerplate that links product config into framework tooling.

### `wowsociety` migration after framework support

Regenerate or simplify:

- `tools/configcheck`
- `cmd/api`
- `cmd/worker`
- `cmd/migrate`

Keep:

- Product-specific config sections.
- Product module registration.
- Product-specific deployment overlays.

### Acceptance criteria

- A new product can run config validation against its composed config without hand-writing a configcheck command.
- Generated process shells include seed sync and rule-definition sync.
- Standard storage and OIDC/JWT configuration can be wired without product-specific boilerplate.

## Explicit Non-Gaps: Product Logic That Should Stay In wowsociety

The following should not move into `wowapi`:

- Maharashtra cooperative housing society statutory rule content.
- Rule citations, legal verification status, and policy-pack rows.
- Domain roles such as member, chairman, secretary, treasurer, auditor, and society manager.
- Committee-seat business semantics, except for the generic relationship-edge service needed to implement them.
- Product API route shape and workflow choices.
- Product module composition in `internal/wire`.
- Product-specific configuration sections, although scaffold support should improve how they are loaded and validated.

## Migration Sequence

Recommended implementation order:

1. **P0 storage adapter**: move S3/MinIO adapter first because it is isolated and domain-neutral.
2. **P0 seed sync lifecycle**: wire generated migrate/CLI support because it blocks fresh deployments.
3. **P0 step-up/AMR**: add seedability and JWT AMR propagation before products rely on step-up broadly.
4. **P0 privileged service APIs**: replace product `SECURITY DEFINER` bridges with audited framework services.
5. **P0/P1 i18n**: add request locale, catalog, and localized `httpx`/validation support.
6. **P1 rules lifecycle/schema**: sync rule definitions and complete schema enforcement.
7. **P1 MFA primitives**: move TOTP/OTP primitives if the framework intends to support MFA beyond consuming external IdP AMR.
8. **P1 scaffold/config**: regenerate `wowsociety` process and config tooling after the above hooks exist.

## Cross-Repository Refactor Checklist

For each upstreamed capability:

1. Add or move the implementation into `wowapi`.
2. Add framework unit and integration tests.
3. Update `wowapi` docs and scaffold templates.
4. Update `wowsociety` to consume the new framework API.
5. Delete the product workaround.
6. Remove product-only dependencies that were introduced only for the workaround.
7. Keep or add compatibility tests proving product behavior did not regress.

## Definition Of Done

This GAP analysis is closed when:

- `wowsociety` no longer carries domain-neutral framework adapters or lifecycle workarounds.
- New products can get storage, seed sync, rule definition sync, locale handling, step-up/AMR propagation, and config tooling from the framework or generated scaffold.
- Product code contains product behavior, not framework infrastructure.
- Framework tests cover the moved capabilities at the `wowapi` layer.
- `wowsociety` tests prove the product still works while depending on framework-provided implementations.
