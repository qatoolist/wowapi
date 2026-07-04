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
pipeline. It does NOT exercise provider PITR/WAL; rehearse that against a real snapshot in staging at
least once per release train. Record each drill (date, RTO/RPO observed, issues) in your ops log.
