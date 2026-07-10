<!-- markdownlint-disable MD013 -->

# Post-Implementation Gap Analysis: Framework Capability Design Review

Date: 2026-07-10

Framework repository: `/Users/qatoolist/go_home/src/github.com/qatoolist/wowapi`

Product repository reviewed for consumption: `/Users/qatoolist/go_home/src/github.com/qatoolist/wowsociety`

Framework branch reviewed: `feat/wowsociety-framework-gaps` at `be84ee2`

Product branch reviewed: `feat/consume-framework-gap-apis` at `4d07067`

Competitive benchmark/RFC companion: `docs/implementation/framework-competitive-architecture-benchmark.md`

## Scope and Review Baseline

This is a post-implementation design review of the eight `wowsociety` framework gaps after the framework branch added the missing APIs and the product branch consumed them.

The latest engineering update was reviewed locally before this document was revised:

- `wowapi` branch `feat/wowsociety-framework-gaps` is at `be84ee2`.
- `wowsociety` branch `feat/consume-framework-gap-apis` is at `4d07067`.
- `wowsociety/FRAMEWORK_VERSION` pins `be84ee2`.
- `ProblemError.Detail` now resolves through `kernel.detail.<code>` keys and falls back byte-identically to the producer's `Msg`.
- Internal-kind errors still expose no `Detail`.
- The stale `wowsociety` comments called out in the prior review have been updated.

Those three review findings are treated as resolved here. This report does not reopen them. Where this report discusses i18n `detail`, it is about the broader source-of-truth, loader, and ownership model for translation catalogs.

## Executive Summary

The branch materially improves `wowapi`: the immediate `wowsociety` workarounds are mostly removed, and several features are implemented with solid engineering discipline. However, the program also shows a process smell: some gaps were filled at the level of "remove this product workaround" rather than "design the framework capability a future product would expect."

The sharpest example is i18n. The framework now has locale negotiation, request-context binding, localized problem titles/details, and validation-message lookup. But it does not define where translations live, how they are loaded, how product teams validate missing keys, how framework strings are overridden safely, or how new products get a translation directory from the scaffold. That is not a full i18n subsystem; it is localization plumbing over an in-memory Go map.

This distinction matters. A framework capability should first "feel the gap" by understanding source of truth, authoring workflow, lifecycle, override model, validation, operations, and scaffold ergonomics. Only then should it fill the gap with APIs.

## External Benchmark: Laravel Localization

Laravel is a useful comparison because it treats localization as a complete developer workflow, not only as a runtime lookup function. The lesson to copy is the product experience: scaffolded source locations, multiple source styles, configured fallback, runtime locale selection, placeholders/pluralization, and safe package/framework overrides. `wowapi` should not copy Laravel's language syntax or project setup; the Go-native equivalent should be framework-shipped language-specific YAML defaults, product YAML/JSON catalog files, optional `.go` catalog bundles, and optional overlays.

Laravel's official localization documentation describes:

- A publish/scaffold step for language files.
- File-backed short-key translations under locale-specific directories.
- JSON translation files for applications with many strings or text-as-key usage.
- Configured default and fallback locales through application config and environment variables.
- Request/runtime locale selection.
- Placeholder replacement, pluralization, and pluralization ranges.
- Package/framework translation overrides.

Sources:

- Laravel 13.x Localization: <https://laravel.com/docs/13.x/localization>
- Specifically: language files and JSON files are documented in the introduction, locale/fallback configuration in "Configuring the Locale", placeholders in "Replacing Parameters", pluralization in "Pluralization", and package overrides in "Overriding Package Language Files". The benchmark is the completeness of the workflow, not the framework's implementation language.

The design lesson for `wowapi`: a framework's i18n story is not complete when it can look up `(locale, key)`. It needs a first-class convention for source files, fallback, publication/scaffold, package/framework overrides, runtime negotiation, validation, and tests.

For `wowapi`, the expected override model is:

1. The framework ships its own default localized keys as language-specific YAML files.
2. `wowapi init` or a dedicated publish command makes those framework keys visible in the product repository.
3. Product-local framework-key files override the embedded framework defaults.
4. Product and module keys load through the same catalog lifecycle before the product's API, worker, or migrate binaries finish booting.
5. Optional DB overlays apply last for tenant/admin editable text.

## Capability Design Checklist

Before a feature is called "filled", the framework should answer these questions:

1. Source of truth: where does the data/config live, and who owns it?
2. Loading lifecycle: when and how is it loaded, validated, cached, and refreshed?
3. Override model: how do product/module/framework namespaces interact safely?
4. Failure mode: does a missing or malformed entry fail boot, fail CI, or fall back deterministically?
5. Scaffold ergonomics: does `wowapi init` generate the expected structure and config?
6. Tooling: is there a CLI/testkit path to validate drift and missing coverage?
7. Extensibility: can a product choose a source or provider without reimplementing framework plumbing?
8. Operations: are metrics, readiness, reload, and deployment behavior defined where relevant?
9. Migration: does the consuming product delete the workaround or only move it somewhere else?

## Finding FG-POST-001: i18n Is Runtime Plumbing, Not a Complete Localization Subsystem

Severity: Critical for claiming GAP-001 is fully closed.

Evidence:

- `kernel/i18n.Catalog` is an in-process `map[locale][key]message` with `Add` and `Lookup`; it has no loader abstraction or persistence model.
- `kernel/i18n.Registry.Register` only accepts a Go `i18n.Bundle` and enforces module-prefix ownership.
- Framework English strings are compiled into Go maps in `kernel/i18n/framework_catalog.go`; they are not maintained as language-specific YAML defaults that a product can publish and override.
- `wowsociety/internal/i18n/messages.go` still keeps product translations as Go maps and manually adds them to `booted.I18n`.
- The generated scaffold wires `httpx.Locale(booted.I18n)` but does not scaffold a `locales/` or `translations/` directory and does not expose an `i18n` config section.

Impact:

- Product teams cannot choose a translation source such as framework-default YAML, product YAML, JSON, Go bundles, TOML, embedded files, or database without writing their own loader.
- There is no `wowapi i18n validate` equivalent to catch missing keys, malformed catalog files, placeholder drift, or unsupported locales in CI.
- New products get locale negotiation but no authoring workflow.
- Framework-owned `kernel.*` translations for non-English locales require direct `Catalog.Add` calls after boot, bypassing the guarded registry lifecycle.
- A product cannot publish the framework's default localized keys into its own repository and override them through a documented precedence chain.
- The implementation removes the exact `wowsociety` parser workaround but leaves the next product to invent its own source/loading convention.

Required framework design:

- Add a first-class loader contract, for example `i18n.Source` or `i18n.Loader`, that loads bundles into a catalog from framework-embedded YAML defaults, compiled Go bundles, `fs.FS`, directories, or other sources.
- Move framework-owned strings out of hardcoded Go maps into language-specific YAML files shipped by the framework, for example `kernel/i18n/locales/en/kernel.yaml`. These files should be embedded into `wowapi` so zero-config products still work.
- Provide a publish/scaffold path that copies the framework's default locale files into the product repository when the product wants to customize them.
- Provide built-in loaders for four first-class source types:
  - Framework default YAML files embedded in `wowapi`.
  - Product-local YAML catalog files for framework overrides, product keys, and module keys.
  - Product-local JSON catalog files for large text-as-key catalogs and tooling interoperability.
  - Go-native `.go` catalog bundles for generated catalogs and products that want compile-time ownership checks.
- Keep TOML optional unless the framework wants config-style translation catalogs, but the source contract should not prevent it.
- Treat database-backed translations as an optional overlay source, not the only source. Static framework/module/product strings should come from framework YAML defaults, product YAML/JSON files, and/or compiled Go bundles; DB storage is better suited for tenant/admin editable overrides.
- Define load precedence explicitly:
  1. embedded framework YAML defaults;
  2. product-published framework override files;
  3. product/module catalog files;
  4. compiled Go catalog bundles, if configured at this layer;
  5. optional DB overlay.
- Define a canonical layout, for example:

  ```text
  locales/
    en/
      kernel.yaml
      product.yaml
      modules/
        identity.yaml
    en.json
    mr/
      kernel.yaml
      product.yaml
      modules/
        identity.yaml
    mr.json
  internal/i18n/catalogs/
    en.go
    mr.go
  ```

- Define collision rules: product-local `kernel.*` keys intentionally override embedded framework defaults; product/module keys must stay in owned namespaces; duplicate keys inside the same precedence layer fail validation.
- Define the Go catalog shape explicitly, for example a package that exports `i18n.Bundle` values or a generated `Register(reg *i18n.Registry) error` function. This is the Go-native equivalent of code-backed language files and must remain idiomatic Go.
- Add `wowapi i18n validate` to check locale coverage, key namespace ownership, placeholder compatibility, duplicate keys, and missing fallback entries.
- Add a scaffolded `locales/en` directory, sample product/module YAML or JSON catalog files, and an optional `internal/i18n/catalogs/en.go` sample for compiled catalogs.
- Add an explicit override model for framework strings: product-provided `kernel.*` locale bundles should be accepted through a controlled framework/system registration path, not through unguarded direct `Catalog.Add` calls.
- Add interpolation and pluralization support, or explicitly document that the first version supports only static strings and cannot be called complete i18n.

Merge implication:

GAP-001 should not be marked fully closed until the source-of-truth and loader story exists. The current implementation is useful plumbing, but it is not a framework-quality localization feature.

## Finding FG-POST-002: The Scaffold Turns i18n On Without Providing an i18n Product Workflow

Severity: High.

Evidence:

- `internal/cli/templates/init/cmd_api_main.go.tmpl` wires `httpx.Locale(booted.I18n)` as an always-on concern.
- `internal/cli/templates/init/internal_appcfg_config.go.tmpl` has framework, auth, storage, and modules config, but no i18n config.
- `internal/cli/templates/init/configs_base.yaml.tmpl` documents auth and storage sections, but no i18n source, fallback locale, supported locales, or catalog path.
- `internal/cli/templates/init/cmd_worker_main.go.tmpl` and `internal/cli/templates/init/cmd_migrate_main.go.tmpl` have no i18n source/load references, so API, worker, and migrate do not share a single localization lifecycle.
- The scaffold tests assert that locale middleware is wired, not that a product can place translation files somewhere and have them loaded.

Impact:

This creates a false sense of completeness. A new product has the middleware but must still invent:

- translation-file layout;
- loader code;
- product/framework override path;
- missing-key checks;
- CI validation.

Required framework design:

- Generate a translation directory in `wowapi init`.
- Generate or publish framework default locale files into the product repository when requested, so product teams can inspect and override framework-owned `kernel.*` keys without editing framework code.
- Generate an optional Go catalog package for teams that prefer compiled translation bundles.
- Generate config such as:

  ```yaml
  i18n:
    default_locale: en
    supported_locales: [en]
    sources:
      - kind: framework_defaults
      - kind: fs
        path: locales
        formats: [yaml, json]
        overrides_framework: true
      - kind: go
        package: internal/i18n/catalogs
        enabled: false
      - kind: db_overlay
        enabled: false
  ```

- Load those catalogs before API, worker, and migrate boot finishes, and bind the request locale before the HTTP middleware is assembled.
- Add scaffold tests that render a product with:
  - a product-local override for a framework `kernel.*` key;
  - a product-owned key in YAML or JSON;
  - a compiled Go catalog.
- The tests must prove `Accept-Language` returns those locale strings without product-authored loader code and that product-local framework overrides win over embedded framework defaults.

## Finding FG-POST-003: The i18n Namespace Model Is Internally Inconsistent

Severity: High.

Evidence:

- `Registry.Register` correctly rejects `kernel.*` writes from modules.
- The documented current path for new `kernel.*` locale strings is direct `booted.I18n.Add(...)`.
- Direct `Catalog.Add` is intentionally unguarded and can write any locale/key pair after boot.

Impact:

The framework has two different safety models:

- module bundles are boot-validated and namespace-guarded;
- framework/product additions made directly to `Catalog` bypass ownership validation.

This is manageable for a test, but it is not a robust product extension path. It also makes it unclear whether the source of truth is the module registry, the app composition root, or arbitrary post-boot catalog mutation.

Required framework design:

- Split registration roles explicitly:
  - framework catalogs under `kernel.*`;
  - module catalogs under `<module>.*`;
  - product-global catalogs under a reserved product namespace such as `product.*` or `app.*`.
- Add a controlled framework-override path for product-local `kernel.*` files. This path should be allowed only for configured product override sources, should validate that keys already exist or are explicitly declared overrideable, and should beat embedded framework defaults during load.
- Add a controlled `RegisterFrameworkLocale` or `RegisterSystemBundle` path that is available to the product composition root but still validates locale/key ownership.
- Freeze or seal the catalog after boot for request-time use, or document and test safe concurrent mutation if post-boot changes are supported.

## Finding FG-POST-004: Rules Public Surface Still Overclaims JSON Schema

Severity: High.

Evidence:

- `kernel/rules/schema.go` is honest internally: it says the validator is focused, not a full JSON Schema implementation, and lists the supported subset.
- The public surface is still stale: `rules.Point.ValueSchema` is documented as `JSON Schema`, and migration `00008_rules.sql` labels `value_schema` as `JSON Schema for values`.
- The focused validator explicitly excludes nested `properties`, `additionalProperties`, `items`, and other JSON Schema keywords.
- Unknown `type` values currently return true in `typeMatches`, which means malformed or unsupported schemas can fail open instead of failing registration.
- `Registry.Register` checks that a schema and default are present, but it does not validate that the default satisfies the schema.
- `Resolver.Resolve` returns stored values or code defaults without revalidating them, despite the public field comment saying schemas are validated at write and resolve.

Impact:

This is a contract honesty problem. The implementation source narrows the contract, but the API and database comments still say `JSON Schema`. Product engineers will reasonably assume JSON Schema semantics from those public surfaces. A partial hand-rolled validator is acceptable only if it is named and documented as a strict custom grammar that fails closed.

Required framework design:

- Either use a proven JSON Schema implementation and support the declared contract, or rename the contract to a deliberately limited `RuleValueSchema` with its own strict grammar.
- If the framework keeps a limited schema, reject unknown keywords and unknown types at rule registration or `SyncDefinitions`, not at proposal time.
- Add validation for defaults against the schema during `Registry.Register` or boot.
- Either validate resolved values before returning from `Resolver.Resolve`, or update the public contract so it does not claim resolve-time validation.
- Add `wowapi rules validate` or include this in boot validation so a product cannot ship a silently under-enforced rule point.

## Finding FG-POST-005: MFA Is a Primitives Package, Not a Complete Framework MFA Capability

Severity: Medium.

Evidence:

- `kernel/mfa` provides correct TOTP/HOTP/OTP primitives, challenge-policy helpers, and sender ports.
- Its package documentation deliberately leaves enrollment UX, factor storage schema, challenge rows, attempt counters, delivery-provider selection, retry/rate limiting, and factor-to-permission policy to products.
- `docs/user-guide/auth.md` documents that products must verify a factor and then append an AMR value themselves.

Impact:

This is better than product teams hand-rolling crypto, but a product still has to design security-sensitive pieces:

- factor enrollment;
- secret storage and rotation;
- backup/recovery codes;
- challenge persistence and replay protection;
- attempt accounting;
- delivery throttling;
- AMR freshness.

Required framework design:

- If the framework intends out-of-box MFA, add a service-level package around the primitives: factor repository port, challenge repository port, DB migrations or storage contract, enrollment APIs, recovery-code support, and rate-limit hooks.
- Keep the closure claim as "MFA primitives" and keep product-level MFA lifecycle as an explicit non-goal unless a service layer is added. This is preventive precision, not a correction of an existing overclaim.

## Finding FG-POST-006: Step-Up Policy Is Boolean and Hard-Coded

Severity: Medium.

Evidence:

- `seeds.PermissionSeed.StepUp` and `authz.Permission.StepUp` are booleans.
- `authz.evaluator` hard-codes the AMR values considered strong: `mfa`, `otp`, `totp`, `hwk`, `sms`, `fpt`, `face`.
- That hard-coded strong-factor set includes `sms`. NIST SP 800-63B treats PSTN out-of-band authentication, including SMS/voice delivery, as a restricted authenticator that requires risk acceptance and alternatives.
- `auth.Claims` carries `amr`, but it does not carry an `auth_time` or equivalent authentication-time claim, so a `StepUpPolicy.MaxAge` cannot be evaluated today.
- The HTTP challenge always advertises `step_up="mfa"`.

Impact:

The implementation is enough for `wowsociety`'s current impersonation permission, but it is not a flexible framework policy:

- a product cannot require hardware-backed MFA for one permission and OTP for another;
- a product cannot express freshness, such as "MFA within last 10 minutes";
- a product cannot configure accepted AMR values without framework code changes;
- a product cannot downgrade or exclude restricted factors such as SMS without framework code changes;
- the challenge cannot tell the client which factor is acceptable beyond generic `mfa`.

Required framework design:

- Replace or extend the boolean with a policy shape, for example:

  ```go
  type StepUpPolicy struct {
      RequiredAMR []string
      MaxAge      time.Duration
      Challenge   string
  }
  ```

- Keep `step_up: true` as shorthand for the default policy, but allow richer seed syntax when needed.
- Move the strong-factor set to configuration or a policy registry.
- Add `auth_time` or equivalent freshness propagation before exposing `MaxAge`; without it, the field would be decorative.
- Treat restricted factors, including SMS/PSTN out-of-band, as policy choices that products must explicitly accept rather than framework defaults.

Source: NIST SP 800-63B, "Restricted Authenticators" and "Authentication Using the Public Switched Telephone Network": <https://pages.nist.gov/800-63-4/sp800-63b.html>

## Finding FG-POST-007: Standalone Seed Sync Is Less Product-Aware Than the Generated Migrate Path

Severity: Medium.

Evidence:

- Generated `cmd/migrate` now uses the composed product config and runs migrations, `seeds.Sync`, and `rules.SyncDefinitions`.
- The standalone `wowapi seed sync` CLI uses `DATABASE_URL` and `config.Defaults().DB` directly.
- The standalone `wowapi seed sync` CLI runs `seeds.Sync` only. It does not run `rules.SyncDefinitions`, so it is not equivalent to the generated migrate lifecycle for products that register rule points.

Impact:

The generated migrate path is good. The standalone CLI is less aligned with the product's real deployment configuration:

- it does not use the product's `configs/<env>.yaml`;
- it does not use secretref resolution;
- it does not share timeout/pool settings from the product config;
- it does not converge rule definitions;
- it can drift from how api/worker/migrate actually connect.

Required framework design:

- Either document `wowapi seed sync` as a low-level escape hatch, not the production-default path, or scaffold a product-local lifecycle command that uses `appcfg.Load` and runs both `seeds.Sync` and `rules.SyncDefinitions`.
- Prefer a generated lifecycle command for production operations, the same way `tools/configcheck` delegates product-specific config.

## Finding FG-POST-008: Privileged Services Are Well Designed, But Allow-List Extension Is Not Scaffolded

Severity: Low to Medium.

Evidence:

- `module.Context.Privileged()` constructs `privileged.New(c.name, ..., privileged.Config{})`.
- The framework supports `AllowRelTypes` and `AllowRuleKeys`, but the default module context always uses an empty config.
- Documentation says a product that needs widened ownership should construct its own `privileged.Services`.

Impact:

The default is safe and the core service is well thought through. The gap is ergonomic: a product that legitimately needs cross-namespace privileged operations must step outside the standard `module.Context` path, which weakens the "use the framework surface, not custom wiring" story.

Required framework design:

- Add a product config section for privileged allow-lists, boot-validate it, and pass it into `module.Context.Privileged()`.
- Keep the default prefix-only behavior.
- Add scaffold docs/tests for a module that is explicitly allow-listed to manage a framework/kernel-owned key.

## Positive Findings

Not every gap is under-designed. These pieces show the right framework instincts:

- S3/MinIO storage is close to framework-quality: it implements a real port, validates boot-time config, maps storage errors to framework errors, handles checksum reality, and has scaffold wiring.
- Seed sync in the generated migrate lifecycle is the right shape: it is idempotent, privileged, and readiness-backed by `app.CatalogsSeeded`.
- Privileged services are structurally sound: tenant binding, ownership, resource/subject existence checks, audit, and DB backstops are in the framework rather than in product SQL bridges.
- Step-up's narrow seed-to-authz path is connected: seed YAML, DB persistence, boot propagation, JWT `amr`, evaluator, HTTP challenge, and testkit all line up.

These are the patterns the weaker features should follow: source of truth, lifecycle, guardrails, scaffold, tests, and operational failure modes all designed together.

## Required Next Actions Before Calling the Program Complete

1. Re-open GAP-001 as `GAP-001B: i18n source/loading/tooling`. Treat current i18n as plumbing, not complete localization.
2. Move framework-owned default translations into language-specific YAML files embedded in `wowapi`.
3. Require framework-owned `kernel.*` translations for every product-declared locale. For example, if a product declares `mr`, framework strings must have Marathi coverage or a deliberate validated fallback, not silent English drift.
4. Decide the translation source contract: embedded framework YAML defaults, product-local YAML/JSON loaders, first-class Go catalog bundles, and optional DB overlay support.
5. Add scaffolded translation directories, a publish path for framework locale files, a sample Go catalog package, config, loaders, and `wowapi i18n validate`.
6. Define safe load precedence: embedded framework defaults first, product-local framework overrides next, product/module keys next, optional DB overlays last.
7. Define a safe system/framework translation registration path for `kernel.*` instead of direct unguarded `Catalog.Add`.
8. Migrate `wowsociety` product translations out of manual Go-map registration into the new framework-supported locale sources.
9. Tighten rules schema semantics: use a JSON Schema library or make the custom schema format explicit and fail closed.
10. Keep GAP-005 closure wording as "MFA primitives" unless a real challenge/enrollment/persistence service is added.
11. Decide whether step-up needs richer per-permission policy now or in a tracked follow-up before broad product adoption.
12. Clarify standalone `wowapi seed sync` as an escape hatch, or provide a product-config-aware generated lifecycle command that also runs rule-definition sync.
13. Use the competitive benchmark/RFC companion to add capacity, middleware phase, security profile, static lifecycle-graph, and benchmark gates to the engineering backlog rather than treating these as ad hoc follow-ups.

## Bottom Line

The branch is valuable, but the standard for a framework cannot be "the product workaround is gone." The standard must be "a new product can use the capability in the obvious way, with source-of-truth, lifecycle, validation, override, and scaffold support already designed."

By that standard, storage, seed lifecycle, privileged services, and the seed-to-authz step-up path are mostly on track. i18n is not. Rules schema validation and MFA also need clearer capability boundaries before the framework team claims these gaps are fully solved.
