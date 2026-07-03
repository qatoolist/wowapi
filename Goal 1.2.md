------- GOAL ACCOMPLISHED - DO NOT REWORK --------

# Goal 1.2: Framework Configuration, Project Configuration, Deployment Robustness

You are the architect for `wowapi`.

Before making changes, read:

- `Goal.md`
- `Goal 1.1.md`
- all files under `docs/blueprint/`

## What Was Changed In The Previous Editing Session

The previous session refined the blueprint so `wowapi` is clearly a reusable third-party Go framework dependency, not a single product repository.

Important changes already made:

1. Added `Goal 1.1.md` to request framework-as-dependency refinements.
2. Added `docs/blueprint/11-framework-distribution-and-consumption.md`.
3. Updated the blueprint so product applications live in their own repositories and import `wowapi`.
4. Changed consumer-facing APIs from the old all-`internal` model to public packages:
   - `wowapi/kernel/...`
   - `wowapi/module`
   - `wowapi/app`
   - `wowapi/adapters/...`
   - `wowapi/testkit`
   - `wowapi/migrations`
   - `wowapi/cmd/wowapi`
5. Clarified that `wowapi/internal/...` is private implementation only and must not contain consumer-facing contracts.
6. Standardized the public module SDK naming to `module.Context`.
7. Added an installable CLI model:
   - `go install github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z`
   - `wowapi init --module example.com/acme-ops --wowapi-version vX.Y.Z`
   - `wowapi new-module`
   - `wowapi gen crud`
   - `wowapi seed validate`
   - `wowapi openapi merge`
   - `wowapi lint boundaries`
8. Made `go run github.com/qatoolist/wowapi/cmd/wowapi@vX.Y.Z ...` only a fallback, not the primary developer workflow.
9. Moved contract-test fixture guidance away from public examples:
   - private fixtures should live under `wowapi/internal/testmodules`
   - `wowapi/examples/*` should be standalone, non-contractual sample apps, preferably nested Go modules.
10. Added explicit acyclic package dependency rules:
    - `kernel` imports no `module`, `app`, `adapters`, `testkit`, examples, or product packages.
    - `module` imports kernel contracts.
    - `adapters` implement kernel ports.
    - `app` is the composition root and wires everything.
    - production packages must not import `testkit`.
11. Added delivery acceptance criteria that package dependency lint must prove the graph is acyclic.
12. Replaced stale Makefile-first wording with `wowapi` CLI commands where appropriate.

Preserve these decisions. Do not revert the framework-as-dependency model.

## New Requirement

We need one more robustness refinement.

`wowapi` must be heavily configurable, but framework configuration and project/product configuration are different things.

The blueprint must clearly define:

- framework-level configuration,
- product/project-level configuration,
- module-level configuration,
- deployment/environment configuration,
- tenant/runtime configuration,
- secrets and credential handling,
- validation and override rules,
- how the CLI and deployment workflows should use these configs safely.

This must improve robustness without compromising:

- security,
- tenant isolation,
- startup correctness,
- performance,
- operational simplicity,
- type safety,
- observability,
- maintainability.

Avoid building a magical or slow runtime configuration system.

## Core Design Problem

The current blueprint mentions typed env config, per-env compose files, rule engine, feature flags, and module config readers, but it needs a sharper model for configuration ownership and lifecycle.

Define the difference between:

1. **Framework configuration**
   - Configuration owned by `wowapi`.
   - Examples: DB pool settings, RLS/session settings, HTTP timeouts, auth/JWKS cache, outbox runner settings, job concurrency, workflow engine settings, upload limits defaults, audit retention defaults, observability/exporter settings, adapter defaults, CLI behavior defaults.

2. **Product/project configuration**
   - Configuration owned by the consuming application.
   - Examples: product module list, enabled adapters, service names, API base URL, product-specific deployment names, product-level default locale/timezone, selected notification providers, public URLs, project-specific operational defaults.

3. **Module configuration**
   - Namespaced configuration for each product module.
   - Examples: requests module SLA defaults, assets module behavior, module-specific job schedules, module feature toggles, module external provider keys by reference.

4. **Deployment/environment configuration**
   - Configuration that changes per local/dev/stage/prod deployment.
   - Examples: database DSN or secret reference, object storage endpoint, OIDC issuer, TLS settings, log level, metrics exporter, worker count, resource limits, deployment region, environment name.

5. **Tenant/runtime configuration**
   - Values controlled through the rule/config engine or tenant admin surfaces.
   - Examples: tenant-specific rule values, notification preferences, retention overrides, workflow template overrides, feature flags where explicitly allowed.

The blueprint must make these boundaries explicit and prevent accidental mixing.

## Required Configuration Principles

Add practical principles such as:

1. **Typed config only**
   - No unstructured global config maps for framework behavior.
   - Config structs must have defaults, validation, redaction tags, and documentation.

2. **Fail fast at boot**
   - Missing required config, invalid ranges, unknown keys, unsafe production defaults, and incompatible config versions fail startup.

3. **Explicit precedence**
   - Define exact order, such as:
     1. framework compiled defaults,
     2. product config file,
     3. environment-specific overlay,
     4. environment variables,
     5. secret manager references,
     6. CLI flags for local tooling only.
   - Explain which layers are allowed in production.

4. **Secrets are references, not values**
   - Config files should contain secret references, not raw secrets.
   - Secrets must be redacted in logs, errors, health output, config dumps, CLI diagnostics, and OpenAPI metadata.

5. **Immutable hot-path config**
   - Framework hot-path config is loaded and validated once at boot.
   - Request paths must read precomputed immutable structs, not parse config or hit stores.

6. **Safe runtime changes**
   - Runtime/tenant configuration belongs in the rule/config engine.
   - Runtime changes must be audited, versioned, validated, and optionally approval-gated.
   - Only explicitly safe values can be hot-reloaded.

7. **No security downgrade by config**
   - Production cannot disable tenant isolation, RLS, audit logging for sensitive actions, route metadata enforcement, authz enforcement, or secret redaction.
   - Any unsafe local/dev-only setting must be explicitly blocked in production.

8. **Module config isolation**
   - Modules can read only their namespaced config through `module.Context.Config()`.
   - Modules must not read arbitrary global framework config.

9. **Deployment clarity**
   - The blueprint should recommend how product apps deploy with framework and project config separated.
   - Include local Docker Compose, Kubernetes/Helm or manifests, and generic container deployments.

10. **Operational introspection without leakage**
    - Provide a safe `wowapi config doctor` or equivalent.
    - Show effective config only with secrets redacted.
    - Include config source/provenance for debugging.

## CLI Enhancements To Consider

Extend the installable `wowapi` CLI plan with configuration tooling.

Consider commands such as:

```text
wowapi config init
wowapi config validate
wowapi config doctor
wowapi config print --redacted
wowapi config diff --from dev --to prod
wowapi config schema
wowapi deploy render --env prod
```

The CLI should help product teams generate, validate, inspect, and render configuration without copying framework internals.

Do not require the CLI to become a deployment platform. Keep it practical: validation, scaffolding, rendering, and diagnostics are enough.

## Deployment Guidance To Add

Add or refine blueprint guidance for deployments.

Define how a product project should handle:

- `config/base.yaml` or equivalent product config,
- `config/dev.yaml`, `config/stage.yaml`, `config/prod.yaml` overlays,
- environment variables,
- secret manager references,
- Kubernetes ConfigMaps/Secrets or equivalent,
- local Docker Compose config,
- CI validation,
- container startup validation,
- migration-time config vs runtime config,
- API process config vs worker process config vs migration process config.

Be explicit that `cmd/api`, `cmd/worker`, and `cmd/migrate` may share a common product config schema but each process should receive only the needed effective config.

## Robustness Enhancements To Include If Useful

Add any additional refinements that improve robustness without harming security, optimization, or performance.

Consider:

- config schema versioning and migration strategy,
- config compatibility checks against the `wowapi` dependency version,
- startup config fingerprinting for observability,
- redacted config snapshot in diagnostics,
- drift detection between API and worker config,
- explicit resource/concurrency budgets,
- safe defaults for timeouts, retries, queue concurrency, and DB pools,
- per-environment production safety checks,
- config-driven adapter selection without service locators,
- contract tests for module config,
- CI gates for unsafe production config,
- docs for config anti-patterns.

Avoid:

- reflection-heavy config magic,
- dynamic config lookups on hot paths,
- untyped YAML blobs passed into services,
- secrets in config files,
- allowing modules to inspect all framework config,
- allowing config to disable core security controls in production,
- creating a low-code rules/config platform before real needs exist.

## Documentation Updates To Make

Update the smallest useful set of blueprint files. At minimum review:

- `docs/blueprint/00-overview.md`
- `docs/blueprint/04-project-and-primitives.md`
- `docs/blueprint/06-module-sdk.md`
- `docs/blueprint/07-platform-services.md`
- `docs/blueprint/08-testing-and-tooling.md`
- `docs/blueprint/10-delivery.md`
- `docs/blueprint/11-framework-distribution-and-consumption.md`
- `docs/blueprint/README.md`

Add a new file if useful, for example:

```text
docs/blueprint/12-configuration-and-deployment.md
```

## Acceptance Criteria

The refinement is successful when the blueprint clearly states:

1. Framework config, product config, module config, deployment config, and tenant/runtime config are separate concepts.
2. `wowapi` exposes typed public configuration contracts where product apps need them.
3. Product apps can define their own config schema without forking framework config.
4. Modules receive only namespaced config through `module.Context`.
5. Unsafe production config fails startup.
6. Secrets are handled by reference and redacted everywhere.
7. Hot-path framework code reads immutable validated config, not dynamic config stores.
8. Tenant/runtime changes are versioned, audited, validated, and handled through the rule/config engine.
9. API, worker, and migrate processes have clear config boundaries.
10. The CLI supports config validation and diagnostics.
11. Deployment guidance explains how to use framework config and project config together.
12. The configuration model does not introduce circular package dependencies.
13. The configuration model does not weaken security, tenant isolation, observability, or performance.

## Tone And Scope

Keep this as a focused architecture refinement.

Do not redesign the full framework.

Do not add a heavyweight centralized configuration platform.

Prefer typed Go structs, explicit validation, clear ownership boundaries, and safe deployment workflows.

The goal is production robustness with simple, predictable behavior.
