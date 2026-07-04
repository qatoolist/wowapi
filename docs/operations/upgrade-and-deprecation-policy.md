# Upgrade & Deprecation Policy (O3 / CA-12)

wowapi is distributed as a versioned Go dependency and is **pre-1.0 (v0)**. This document is the published
upgrade discipline and deprecation policy the roadmap (O3) requires — the contract a product depends on
when it pins a framework version.

## Versioning

- Semantic-ish, v0 rules: `v0.MINOR.PATCH`. While `MAJOR` is 0 the **minor** bumps may carry breaking
  changes to the public surface (`kernel` / `module` / `app` / `adapters` / `testkit` / `migrations` +
  `cmd/wowapi`); **patch** bumps never do.
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

## What may break between v0 minors

Declared fair game while v0 (call these out in the CHANGELOG when they change):

- **Error kinds** (`kernel/errors.Kind`) — new kinds may be appended; the HTTP status/code mapping is
  stable per kind. (Kinds are appended at the end of the enum so existing values never shift.)
- **Config schema** (`config.Framework`) — new keys may appear; `schema_version` gates incompatibilities;
  the loader rejects a config declaring a newer `schema_version` than the binary supports.
- **`module.Context`** — the interface widens as kernel capabilities land (new accessors). Widening is an
  accepted breaking change while v0; a product recompiles and the contract suite catches any real break.
- **`testkit`** helpers — test-only surface, may change to match kernel changes.

## Deprecation process

When a public symbol must be removed or changed incompatibly:

1. **Announce in the CHANGELOG** under a `### Deprecated` heading in the release that introduces the
   replacement, naming the old symbol, the replacement, and the earliest version it may be removed in
   (at least one minor later).
2. **Keep the old symbol working** for that overlap window where feasible, with a doc comment
   `// Deprecated: use X instead (removed in v0.N).` so `go vet`/staticcheck surface it at call sites.
3. **Remove** no earlier than the announced version, under a `### Removed` CHANGELOG heading.
4. **Migrations are forward-only** in spirit: a released migration is never edited; corrections ship as a
   new migration. `Down` legs exist for the reversibility drill and local rollback, not for production
   downgrade of applied data changes.

## Framework-purity gate

Every issue shaken out of a product lands upstream as a **domain-neutral** fix in wowapi or in
`ROADMAP-wowapi.md` — never as a product-side workaround. Society/product-specific concepts do not enter
the framework (blueprint 13 §3).

## Before Phase 2 of any product

Per O3, the framework publishes this policy (this document) before a product's post-MVP phases depend on
it. Products should: pin an exact version, wire the contract suite into CI as the tripwire, and read the
CHANGELOG `Deprecated`/`Removed` headings on every bump.
