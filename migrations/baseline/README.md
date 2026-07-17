# Clean kernel baseline proof

`00001_baseline.sql` is the first supported database state for `v1.2.0`. It is
not an upgrade from the abandoned `v1.0.0`/`v1.1.0` databases.

The baseline was constructed from PostgreSQL 16's catalog after applying the
proven historical `00001..00050` state to an empty database. Database-owned Up
DDL and dependency-ordered Down DDL came from `pg_dump --schema-only`; the
cluster-global `app_rt` and `app_platform` role bootstrap was carried explicitly
because roles are outside a database dump. The goose history relation is owned
by the runner and excluded from generated DDL. Database-scoped extensions are
retained on Down because other schemas or product modules in the same database
may use them.

There are no intentional logical schema deltas between the reference head and
the clean baseline. The census compares named objects and complete semantics.
It deliberately does not compare `pg_attribute.attnum` gaps: those encode
dropped-column history rather than a named-column contract, and preserving them
would require replaying abandoned physical DDL. Production, generated, and
testkit queries name their projected columns; the clean line does not support
old `SELECT *` consumers.

The executable proof is:

```sh
make baseline-census-check
make baseline-census-discriminates
DATABASE_URL=postgres://... ./scripts/migration_reversibility_drill.sh
make tenantfk-gate
make golden-consumer
```

The first command requires every semantic line to equal
`census-reference.txt`; the second proves the oracle notices same-count catalog
mutations. The remaining gates prove version-1 up/idempotent-up/down/up,
tenant-FK/RLS/grant behavior, and generated kernel-plus-module migration
compatibility.
