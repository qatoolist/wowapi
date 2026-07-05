# Backup & restore (O5)

Domain-neutral procedure for backing up and restoring a wowapi-backed product. The framework holds no
backup primitives — durability is the database/object-store provider's job — but a product must
*rehearse* restore, not just configure backups. Run the drill quarterly.

## What to back up

1. **PostgreSQL** — the system of record. Use the managed provider's **PITR**: daily base snapshot +
   continuous WAL archiving, retention ≥ your legal minimum. RLS, roles, and the per-source
   `goose_version_*` tables are all in the database, so a physical/PITR restore brings the schema and
   tenancy model back intact.
2. **Object storage** (documents/attachments blobs, `kernel/storage`) — enable bucket **versioning** +
   cross-region replication. Blob keys are tenant-prefixed and referenced by `document_versions.storage_key`;
   a DB restore is only consistent if the bucket is restored to the *same or later* point in time (never
   earlier — a DB row must never point at a missing blob).
3. **Secrets** — provider-managed (`secretref://`); back up per your secret manager's own procedure. The
   config holds only references, never values.

## Restore order (point-in-time)

1. Restore Postgres to target timestamp T (provider PITR).
2. Restore/roll object storage to ≥ T (so every `storage_key` resolves).
3. Rotate any credentials that may have leaked in the incident window.
4. Boot `cmd/migrate` (no-op if the restore is already at head; it reconciles `goose_version_*`).
5. Verify `/readyz` is green and the config fingerprint matches the expected release
   (see [deployment-checklist.md](deployment-checklist.md) §2).

## The drill

[`scripts/backup_restore_drill.sh`](../../scripts/backup_restore_drill.sh) proves the dump→restore
round-trip against a seeded instance: it seeds a marker, `pg_dump`s the source, restores into a scratch
database, and asserts the schema and marker survived.

```sh
# against the local compose DB (make up && make migrate first)
SRC_URL=postgres://wowapi:wowapi-local-only@localhost:5432/wowapi ./scripts/backup_restore_drill.sh
```

It is a **logical** dump/restore round-trip — enough to validate the procedure and catch a broken dump
pipeline. Record each drill (date, RTO/RPO observed, issues) in your ops log.

## PITR & object-storage restore legs (B-5)

Two further drills rehearse the mechanisms a real restore relies on — both self-contained and runnable
against the local stack, so the *procedure* is proven in CI/local rather than only trusted from this
runbook:

1. **PITR** — [`scripts/pitr_restore_drill.sh`](../../scripts/pitr_restore_drill.sh) (`make drill-pitr`).
   Spins up its OWN throwaway `postgres:16-alpine` primary with `archive_mode=on`, takes a physical
   `pg_basebackup`, writes a pre-target row + marks timestamp T, writes a post-target row, then restores
   the base backup into a fresh server with `restore_command` + `recovery_target_time=T`, replays WAL and
   promotes. It asserts the pre-target row survives and the post-target row does **not** — i.e. recovery
   stopped exactly at T. This proves we can execute the recovery procedure against real WAL. **Production
   PITR itself stays a managed-provider capability** (continuous WAL archiving + retention live in the
   provider, not the ephemeral compose DB) — see decision **D-0080**.

2. **Object storage** — [`scripts/object_storage_restore_drill.sh`](../../scripts/object_storage_restore_drill.sh)
   (`make drill-object-storage`). Against the compose MinIO: writes a tenant-prefixed blob, backs it up
   (mirror to a backup bucket), simulates loss, restores (mirror back), and asserts the blob returns
   byte-identical and the `storage_key` resolves — the referential invariant that a restored DB row must
   never point at a missing blob.

```sh
make drill-pitr            # needs docker (no compose stack required)
make up && make drill-object-storage   # needs the compose MinIO
```

Provider PITR/WAL and cross-region object replication are still worth rehearsing against a real snapshot
in staging at least once per release train; these drills prove the local, in-repo half of the procedure.
