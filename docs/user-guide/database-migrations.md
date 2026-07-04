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
make migrate-down        # go run ./cmd/migrate down  — roll back the last migration
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

## Common problems

| Symptom | Likely cause | Fix |
|---|---|---|
| `permission denied for table widgets` at runtime | missing `GRANT … TO app_rt`, or connected as wrong role | Add the grant in the migration; confirm `DATABASE_URL` uses `app_rt`. |
| Query returns error mentioning `app_tenant_id` | no tenant context set (queried outside `WithTenant`) | Run DB work inside `tx.WithTenant`/`WithTenantRO`. |
| `must be owner of table` during migrate | migrating as `app_rt` instead of `app_migrate` | Use `MIGRATE_URL` (`app_migrate`) for DDL. |
| Reversibility gate fails | `Down` doesn't reverse `Up` | Fix the `Down`, or declare the migration forward-only with a reason. |
| Rows from other tenants visible | RLS not forced, or policy missing `WITH CHECK` | Add `FORCE ROW LEVEL SECURITY` + a `WITH CHECK` clause. |

Next: [Auth](auth.md) · [Testing](testing.md).
