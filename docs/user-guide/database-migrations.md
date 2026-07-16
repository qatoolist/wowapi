# Database & Migrations

wowapi is PostgreSQL-first and treats the database as a **security boundary**, not just storage. This page
covers the role model, how migrations are structured and run, how to write a tenant-safe migration, and the
reversibility drill the framework enforces.

## Requirements

- **PostgreSQL 16.**
- Three roles: `app_rt` and `app_platform` are **created by the first kernel migration**
  (`migrations/00001_bootstrap.sql`); **`app_migrate` is not** — it's the role the migration *runner*
  connects as, so it must already exist.
- Migrations are **[goose](https://github.com/pressly/goose)** SQL files, embedded into the binaries.

## The role model

| Role | Used by | Created by | Privilege |
|---|---|---|---|
| `app_rt` | runtime (api + worker) | bootstrap migration | **Non-superuser, least privilege.** Subject to `FORCE ROW LEVEL SECURITY`. Cannot bypass RLS, cannot run DDL. |
| `app_platform` | deliberate cross-tenant operations | bootstrap migration | Can operate across tenants where a feature legitimately needs it (kept narrow). |
| `app_migrate` | migrations only | **you (pre-existing)** | The role the migration runner connects as; owns/alters schema (DDL). Not used to serve requests. |

The two DSNs your product provides map onto this split:

```bash
DATABASE_URL=postgres://app_rt@host:5432/db?sslmode=require        # runtime  → db.dsn
MIGRATE_URL=postgres://app_migrate@host:5432/db?sslmode=require    # DDL      → db.migrate_dsn
```

> **Bootstrapping:** the migration *runner* connects with whatever role `MIGRATE_URL` names, and the first
> kernel migration then creates `app_rt` + `app_platform` and the RLS scaffolding. So on a brand-new
> database, point `MIGRATE_URL` at a role that can create roles (the DB owner or a superuser, or a
> pre-created `app_migrate` with `CREATEROLE`) for the initial `migrate up`. `app_rt`/`app_platform` exist
> only *after* that first migration; provision `app_migrate` yourself beforehand.
>
> The migration creates `app_rt`/`app_platform` as `NOLOGIN` by design (no password ships in migrations) —
> ops must grant a login out-of-band before `DATABASE_URL`/`PLATFORM_URL` will connect: `ALTER ROLE app_rt
> LOGIN PASSWORD '…';` (same for `app_platform`). See `scripts/product-dev.sh` for the automated version of
> this step.

## Row-level security in practice

Tenant isolation is enforced in the database, so a bug in application code cannot leak another tenant's
rows:

- Tenant tables enable **and force** RLS. `app_rt` cannot see past a policy even by accident.
- Policies key off `app_tenant_id()`, which **raises when no tenant is set** — no tenant context means an
  error, never an unscoped read.
- The runtime sets the tenant per transaction with `SET LOCAL app.tenant_id = …`; `LOCAL` scopes it to
  that transaction so connections in the pool never leak tenant state.

In application code you never write `WHERE tenant_id = …`. The tenant identity travels in the
**`context.Context`** (set by the framework from the authenticated request via
`database.WithTenantID(ctx, id)`), and the `TxManager` binds it with `SET LOCAL app.tenant_id` for the
duration of the transaction. If no tenant is in the context, the call **fails closed** with
`ErrNoTenantContext` before any connection is used.

```go
// read-write, tenant-scoped (RLS active). Tenant comes from ctx, not an argument.
err := tx.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
    _, err := db.Exec(ctx, `INSERT INTO widgets (id, tenant_id, title) VALUES ($1, app_tenant_id(), $2)`, id, title)
    return err
})

// read-only: BEGIN READ ONLY — list/get paths cannot INSERT/UPDATE/DELETE
err = tx.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error { /* SELECT only */ })
```

The manager also re-asserts the RLS role transaction-locally (`SET LOCAL ROLE`) and, when
`WithRLSGuard()` is enabled, **refuses** to run a tenant transaction whose effective role is a superuser or
has `BYPASSRLS` — because `FORCE ROW LEVEL SECURITY` doesn't apply to those, deployed processes must enable
this guard. (`kernel/database/txmanager.go`.)

> `TxManager.Platform` (cross-tenant, no tenant binding) exists for **kernel services only** and is
> deliberately **not** exposed through `module.Context` — modules cannot escape tenant scope.

## Running migrations

From a product repo (generated `Makefile` + `cmd/migrate`):

```bash
make migrate-up          # go run ./cmd/migrate up    — apply all pending
make migrate-down        # go run ./cmd/migrate down  — FULL reset to version 0 (not a stepwise rollback); refuses outside local/dev
```

Migrations are embedded, so the compiled `migrate` binary is self-contained — no loose `.sql` files to
ship. In the **framework repo**, the same happens via the kernel `migrations/` package + the container
gate.

## Writing a module migration

Modules ship their own migrations and register them in `Register` via `mc.Migrations(migrationsFS)`:

```go
//go:embed migrations/*.sql
var migrationsFS embed.FS

func (m *Module) Register(mc module.Context) error {
    mc.Migrations(migrationsFS)   // discovered + applied in order with the kernel's
    // ...
}
```

A migration file is standard goose with **both** directions:

```sql
-- migrations/0001_widgets.sql
-- +goose Up
CREATE TABLE widgets (
    id          uuid PRIMARY KEY,
    tenant_id   uuid NOT NULL,
    title       text NOT NULL,
    count       int  NOT NULL DEFAULT 0,
    created_at  timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE widgets ENABLE ROW LEVEL SECURITY;
ALTER TABLE widgets FORCE ROW LEVEL SECURITY;

CREATE POLICY widgets_tenant_isolation ON widgets
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());

GRANT SELECT, INSERT, UPDATE, DELETE ON widgets TO app_rt;

-- +goose Down
DROP TABLE widgets;
```

**Checklist for every tenant table:**

- [ ] `tenant_id uuid NOT NULL`
- [ ] `ENABLE ROW LEVEL SECURITY` **and** `FORCE ROW LEVEL SECURITY`
- [ ] a policy `USING (tenant_id = app_tenant_id())` with a matching `WITH CHECK`
- [ ] `GRANT` only the needed privileges to `app_rt` (not the owner)
- [ ] a working `-- +goose Down` that actually reverses the `Up`

## The reversibility drill (do not skip)

The framework's migration gate runs each migration **up → down → up** to prove `Down` truly reverses `Up`.
A migration whose `Down` is a stub, or that can't re-apply cleanly, **fails the gate** — this has already
caught a real defect in the kernel's own migration set. Before you commit a migration:

```bash
make migrate-up
make migrate-down     # must succeed and leave the schema as before the migration
make migrate-up       # must re-apply cleanly
```

If `Down` can't be made truly reversible (e.g. a destructive data change), document why in the migration
and treat it as a forward-only migration explicitly rather than shipping a lying `Down`.

## Seeds (declarative YAML catalogs)

Seeds are **not** SQL. A module ships a declarative **YAML catalog** of authorization/registry data —
permissions, roles, resource types, and relationship types — which the kernel parses and syncs
idempotently at boot (`kernel/seeds/seeds.go`, parsed into a `Bundle`). `wowapi new-module` scaffolds a
`seeds/permissions.yaml`.

```go
//go:embed seeds/*.yaml
var seedsFS embed.FS

func (m *Module) Register(mc module.Context) error {
    mc.Seeds(seedsFS)   // embedded YAML catalog bundle
    // ...
}
```

```yaml
# seeds/permissions.yaml
permissions:
  - key: requests.request.create
    description: Create a request
  - key: requests.request.read
    description: Read a request
# a bundle may also declare: roles, resource_types, relationship_types
```

The catalog is for deterministic reference data the authz system needs — not row fixtures and not test
data. Syncing is idempotent (re-applying a bundle is a no-op), which the module contract suite asserts.
Row-level reference data still belongs in migrations.

### Seed sync is part of the production deploy lifecycle

Loading a bundle into memory at boot is **not** the same as syncing it to the database. Two lifecycle
steps must both happen before a deployed process serves traffic:

1. **Boot** parses every module's seeds and registers them into the in-memory registries the evaluator
   reads (`app.Boot`) — this always happens, on every process (api/worker/migrate).
2. **Sync** (`kernel/seeds.Apply`) upserts that merged bundle into the database's global catalog tables
   (`permissions`, `roles`, `role_permissions`, `resource_types`, `relationship_types`) on a
   platform-privileged (`app_platform`) connection — this does **not** happen automatically. Apply
   computes a canonical content hash of the parsed bundle, records every run in `seed_sync_runs`,
   and short-circuits to a true no-op when the database already reflects the same hash.

Skipping step 2 leaves the catalog tables empty on a fresh database: every authorization check denies
(deny-by-default with no seeded grants) and resource-mirror writes fail their foreign key against the
empty `resource_types` table. The failure is silent at deploy time — nothing warns you seeds were never
applied — and surfaces only as scattered 403s and FK errors once the process is already serving.

**The production path is the generated `cmd/migrate`.** It loads the composed product config via
`appcfg.Load()` (`configs/base.yaml` + `configs/<env>.yaml` + `secretref://` resolution), then runs
migrations, `seeds.Apply`, and `rules.SyncDefinitions` — in that order, on one privileged connection:

```bash
make migrate-up          # kernel migrations → module migrations → seeds.Apply → rules.SyncDefinitions
```

**`wowapi seed sync` is a low-level, standalone escape hatch** — e.g. to re-sync seed catalogs without a
full migrate run — not a substitute for the generated migrate on a fresh environment:

```bash
wowapi seed sync --module widgets=internal/modules/widgets/seeds \
                  --module billing=internal/modules/billing/seeds

wowapi seed sync --dry-run --module widgets=internal/modules/widgets/seeds
```

It has real limitations relative to the generated migrate path:

- **No product config.** It connects via a bare `DATABASE_URL` environment variable, not `appcfg.Load()` —
  no `configs/<env>.yaml` layering, no `secretref://` resolution, and hardcoded pool defaults
  (`config.Defaults().DB`) rather than the product's tuned `db.pool` settings.
- **Does not sync rule definitions.** Unlike the generated migrate, this command never calls
  `rules.SyncDefinitions` — see [Rule definitions](#rule-definitions) below for why: rule points exist only
  as Go declarations inside a booted product process, and a framework-only CLI binary has no product rule
  registry to read, so there is nothing to sync here even in principle. The command prints a warning to
  this effect on every run.

Both `seeds.Apply` calls are idempotent (safe to re-run every deploy). `wowapi seed sync` connects to
`DATABASE_URL` as `app_platform` (the same role/RLS-guard convention as `wowapi dlq`), loads and merges
every named module's seed directory the same way `app.Boot` does, then calls `seeds.Apply`. It has no
long-lived process, so it never holds an in-process authz cache to invalidate; a running api/worker with
`AuthzCacheTTL` set gets the equivalent `InvalidateAll()` call because `seeds.Apply` passes invalidators
through to `seeds.Sync` — pass it when calling `seeds.Apply` from a long-lived process (the generated
migrate/CLI paths don't need to; they exit immediately after).

A generated api process wires a `/readyz` check (`app.ReadinessWithCatalogs`, check name `seed_catalogs`)
that fails loudly — naming `wowapi seed sync` — if any module's declared seeds are missing from the
database, so a pod that skipped this step never reports ready instead of silently denying every request.
Once seed-sync has run, the readiness payload reports the seed/catalog hash as
`details.seed_catalog_hash` for drift correlation (MATRIX CS-21).

## Rule definitions

A module registers **rule points** in Go (`mc.Rules().Register(module, rules.Point{...})` — key, a
**RuleValueSchema**, default value, allowed scopes, whether changes require approval, description).
Registering a point only builds the in-memory registry; `rule_versions.rule_key` carries a foreign key
onto `rule_definitions`, so the framework must also persist a mirror of every registered point into that
table before any tenant/platform/org rule VALUE can be proposed for it (blueprint 02 §2.1: "Rule
definition row … persisted mirror of the registered point — makes points introspectable/auditable in the
DB").

### RuleValueSchema (not JSON Schema)

`Point.ValueSchema` is **not** JSON Schema, despite the historical name and older comments/docs that said
so. It is `RuleValueSchema`: a small, closed grammar the framework validates by hand — no JSON-Schema
library dependency (ratified Decision 2). The ONLY recognized top-level keywords are:

| Keyword | Applies to | Meaning |
|---|---|---|
| `type` | any | one of `integer`, `number`, `string`, `boolean`, `object`, `array`, `null` |
| `enum` | any | JSON array of allowed literal values |
| `minimum`, `maximum`, `exclusiveMinimum`, `exclusiveMaximum` | numbers | numeric bounds |
| `minLength`, `maxLength`, `pattern` (RE2) | strings | length/pattern constraints |
| `minItems`, `maxItems` | arrays | length bounds |
| `required` | objects | shallow presence check for the named keys — **not** recursive per-property validation |

There is no nested `properties`/`items` sub-schema evaluator, no `additionalProperties`, no
`multipleOf`, and no other JSON Schema keyword — a rule point needing per-property typing should
declare separate top-level rule points instead of one object-shaped point with nested constraints.

This is enforced, not just documented — `Registry.Register` **fails registration** (surfaced through
`Registry.Err()`, the same boot-error-accumulation gate `app.Boot` already calls) if:

- `type` is anything other than the seven recognized values (an unrecognized type used to be silently
  accepted — that was a real bug, now closed);
- the schema contains any keyword outside the table above (previously silently dropped by a lax JSON
  decode, so a typo'd or unsupported constraint was never enforced — now a strict decode rejects it);
- the `Default` value does not itself conform to `ValueSchema` (previously never checked at all — a
  point could ship with a default that violated its own schema).

A schema that fails any of these checks can never reach `rule_definitions` — `SyncDefinitions` only ever
sees points that already passed `Register`.

`Resolver.Resolve` additionally re-validates the winning **stored** value against the point's *current*
schema before returning it (cheap — pure in-memory, no extra I/O) — this catches the case where a
schema was tightened after a value was written (module upgrade) and the old stored value no longer
conforms; Resolve surfaces that drift as an error rather than silently handing back a non-conforming
value.

`kernel/rules.SyncDefinitions(ctx, db, registry)` is that lifecycle step — the rule-registry analogue of
`kernel/seeds.Sync` (GAP-007). It upserts every point the registry holds into `rule_definitions`,
idempotently converging `value_schema`, `default_value`, `allowed_scopes`, `requires_approval`, and
`description` on whatever the Go registry currently declares:

```bash
make migrate-up   # kernel migrations → module migrations → seeds.Sync → rules.SyncDefinitions
```

The generated `cmd/migrate` runs it automatically, immediately after `seeds.Sync`, on the same
`app_platform`-privileged connection — `rule_definitions` is `app_platform` SELECT/INSERT/UPDATE and
`app_rt` SELECT-only (migration `00008_rules.sql`), the same write posture as the seed catalogs.

Unlike seed catalogs (declarative YAML on disk, loadable by the framework CLI without any product Go
code), rule points exist only as Go declarations inside a booted product process — there is no
`rules.yaml` equivalent the framework can load from disk today. `wowapi rules sync` is therefore **not**
a framework CLI subcommand: a standalone framework binary has no way to import a product's registered
`rules.Point` values. The generated migrate main is the only sanctioned lifecycle path; a product with a
custom migrate main should call `rules.SyncDefinitions(ctx, pool, booted.Kernel.Rules)` itself, right
after `seeds.Sync`, following the same pattern.

This closes the gap that previously forced a product to hand-write a `rule_definitions` INSERT migration
(mirroring its Go declarations 1:1 in SQL) and a drift-guard test comparing the two — both become
unnecessary once `SyncDefinitions` keeps the database converged with the registry automatically.

## Common problems

| Symptom | Likely cause | Fix |
|---|---|---|
| `permission denied for table widgets` at runtime | missing `GRANT … TO app_rt`, or connected as wrong role | Add the grant in the migration; confirm `DATABASE_URL` uses `app_rt`. |
| Query returns error mentioning `app_tenant_id` | no tenant context set (queried outside `WithTenant`) | Run DB work inside `tx.WithTenant`/`WithTenantRO`. |
| `must be owner of table` during migrate | migrating as `app_rt` instead of `app_migrate` | Use `MIGRATE_URL` (`app_migrate`) for DDL. |
| Reversibility gate fails | `Down` doesn't reverse `Up` | Fix the `Down`, or declare the migration forward-only with a reason. |
| Rows from other tenants visible | RLS not forced, or policy missing `WITH CHECK` | Add `FORCE ROW LEVEL SECURITY` + a `WITH CHECK` clause. |

Next: [Auth](auth.md) · [Testing](testing.md).
