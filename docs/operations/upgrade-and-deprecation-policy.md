# Upgrade & Deprecation Policy (O3 / CA-12)

wowapi is distributed as a versioned Go dependency. The target `v1.2.0` release is the first
supported release of the clean baseline at `github.com/qatoolist/wowapi`. This document is the
published upgrade discipline and deprecation policy the roadmap (O3) requires — the contract a product
depends on when it pins a framework version.

## Versioning

- **Semantic versioning from the clean baseline.** The previously published `v1.0.0` and `v1.1.0`
  identities are abandoned practice releases and carry no compatibility or data-migration
  obligation. Public Go symbols, generated contracts, config semantics, event compatibility, and
  migrations remain backward-compatible from `v1.2.0` through the supported v1 line. An
  incompatible future release requires the `/v2` module path and a migration guide.
- The `module.Context` interface is not widened within the supported v1 line: new capabilities are added via narrow
  interfaces/packages and adapters.
- **Support window.** The current and immediately previous supported minor lines are supported. The previous
  minor receives critical security/data-integrity fixes for at least six months after its successor ships.
  The window starts at `v1.2`; abandoned `v1.0`/`v1.1` are explicitly outside it.
- **Generator/CLI version pairing.** A generated product records the framework major/minor and
  manifest-schema version. Mutating CLI generators require the same framework major/minor as the product;
  patch differences are allowed only when generated-template compatibility tests pass. `wowapi version`
  fails mutating commands on an incompatible pairing rather than merely warning.
- **Rolling-deployment compatibility is N/N-1 supported minor:** N code must run on the expanded N schema, and N-1
  code must continue to run during the N rollout until the contract phase. Direct upgrades older than N-1
  run the intervening upgrade steps in order.
- **Release-blocking after v1.2.0:** OpenAPI request requirements, response removals/narrowing, security
  weakening, config removals/semantic changes, and incompatible event schema changes are release-blocking.
  Additive optional fields and new operations remain allowed.
- Products **pin an exact version** in `go.mod` (`require github.com/qatoolist/wowapi vX.Y.Z`) and upgrade
  deliberately, never via `@latest`.

## The upgrade tripwire (mandatory)

Every product runs the **module contract suite** as its upgrade gate:

```go
func TestWidgetsContract(t *testing.T) { testkit.RunModuleContract(t, &widgets.Module{}) }
```

`testkit.RunModuleContract` boots the module against the new framework version and asserts the invariants
a module must uphold (boot+validate, idempotent migrations/seeds, RLS enforced, config-key strictness).
**A green contract suite is the signal that an upgrade is safe;** a red one localizes the break to a
specific module before it reaches production. Run it in CI against the target version before merging a
framework bump.

## What may change within a v1 minor (additive only)

Declared fair game within the supported v1 line after `v1.2.0` (call these out in the CHANGELOG when they change) — all backward-compatible,
additive changes, never breaking:

- **Error kinds** (`kernel/errors.Kind`) — new kinds may be appended; the HTTP status/code mapping is
  stable per kind. (Kinds are appended at the end of the enum so existing values never shift.)
- **Config schema** (`config.Framework`) — new keys may appear; `schema_version` gates incompatibilities;
  the loader rejects a config declaring a newer `schema_version` than the binary supports.
- **`module.Context`** — not widened again within the supported v1 line (see Versioning above); new capabilities land as
  narrow interfaces/packages and adapters instead. A product recompiles and the contract suite catches any
  real break.
- **`testkit`** helpers — test-only surface, may gain additive helpers to match kernel changes.

Anything incompatible with the above — narrowing, removal, or breaking behavior change to any of these —
requires a `/v2` module path, per the Versioning section.

## Deprecation process

After the `v1.2.0` baseline, a public symbol is never removed or changed incompatibly within v1 — that is itself a breaking change and
requires a `/v2` module path (see Versioning above). A symbol may still be superseded in place:

1. **Announce in the CHANGELOG** under a `### Deprecated` heading in the release that introduces the
   replacement, naming the old symbol and the replacement.
2. **Keep the old symbol working** for the remainder of v1, with a doc comment
   `// Deprecated: use X instead (removed in /v2).` so `go vet`/staticcheck surface it at call sites.
3. **Removal happens only in the future `/v2` module path**, under that release's `### Removed` CHANGELOG
   heading — never as an in-place v1 removal.
4. **Migrations are forward-only** in spirit: a released migration is never edited; corrections ship as a
   new migration. `Down` legs exist for the reversibility drill and local rollback, not for production
   downgrade of applied data changes.

## Framework-purity gate

Every issue shaken out of a product lands upstream as a **domain-neutral** fix in wowapi or in
the framework backlog (`docs/implementation/premier-framework-implementation-plan.md` §5 plus the
fable5 review's task register; historical backlogs `ROADMAP-wowapi.md` and
`framework-engineering-backlog.md` are archived in `wowapi2/archive/`) — never as a product-side
workaround.
Society/product-specific concepts do not enter
the framework (blueprint 13 §3).

## Before Phase 2 of any product

Per O3, the framework publishes this policy (this document) before a product's post-MVP phases depend on
it. Products should: pin an exact version, wire the contract suite into CI as the tripwire, and read the
CHANGELOG `Deprecated`/`Removed` headings on every bump.
