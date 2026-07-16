-- Backfill checkpoint identity is (job_id, tenant_id) — adversarial review
-- 2026-07-17 F-03: the job_id-only primary key made a second tenant using the
-- same stable JobID collide with the first tenant's checkpoint (reads missed,
-- writes overwrote), and runtime reads that filtered on (job_id, tenant_id)
-- could never find their own row. Global (tenant-less) jobs use the all-zeros
-- sentinel so the composite key stays NOT NULL. Mirrors the code-owned
-- EnsureCheckpointTable upgrade in kernel/migration/backfill.go.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 5000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM migration.backfill_checkpoint WHERE tenant_id IS NULL
-- rollback_plan: goose Down restores the job_id-only primary key; rows for more than one tenant per job_id must be resolved (keep one) before Down can succeed.
-- +wowapi:end

-- +goose Up

CREATE SCHEMA IF NOT EXISTS migration;

CREATE TABLE IF NOT EXISTS migration.backfill_checkpoint (
    job_id text NOT NULL,
    tenant_id uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    last_key bigint NOT NULL DEFAULT 0,
    updated_at timestamptz NOT NULL DEFAULT now(),
    lease_token text,
    lease_generation bigint NOT NULL DEFAULT 0,
    lease_expires_at timestamptz,
    PRIMARY KEY (job_id, tenant_id)
);

-- Upgrade a pre-existing single-key table in place (no-op on the new shape).
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint c
         JOIN pg_class r ON r.oid = c.conrelid
         JOIN pg_namespace n ON n.oid = r.relnamespace
        WHERE n.nspname = 'migration' AND r.relname = 'backfill_checkpoint'
          AND c.contype = 'p' AND array_length(c.conkey, 1) = 1
    ) THEN
        UPDATE migration.backfill_checkpoint
           SET tenant_id = '00000000-0000-0000-0000-000000000000'
         WHERE tenant_id IS NULL;
        ALTER TABLE migration.backfill_checkpoint
            ALTER COLUMN tenant_id SET DEFAULT '00000000-0000-0000-0000-000000000000',
            ALTER COLUMN tenant_id SET NOT NULL;
        ALTER TABLE migration.backfill_checkpoint
            DROP CONSTRAINT backfill_checkpoint_pkey;
        ALTER TABLE migration.backfill_checkpoint
            ADD PRIMARY KEY (job_id, tenant_id);
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down

-- Restores the pre-F-03 single-column identity. Requires at most one row per
-- job_id (see rollback_plan); the sentinel maps back to NULL.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_constraint c
         JOIN pg_class r ON r.oid = c.conrelid
         JOIN pg_namespace n ON n.oid = r.relnamespace
        WHERE n.nspname = 'migration' AND r.relname = 'backfill_checkpoint'
          AND c.contype = 'p' AND array_length(c.conkey, 1) = 2
    ) THEN
        ALTER TABLE migration.backfill_checkpoint
            DROP CONSTRAINT backfill_checkpoint_pkey;
        ALTER TABLE migration.backfill_checkpoint
            ALTER COLUMN tenant_id DROP NOT NULL,
            ALTER COLUMN tenant_id DROP DEFAULT;
        UPDATE migration.backfill_checkpoint
           SET tenant_id = NULL
         WHERE tenant_id = '00000000-0000-0000-0000-000000000000';
        ALTER TABLE migration.backfill_checkpoint
            ADD PRIMARY KEY (job_id);
    END IF;
END $$;
-- +goose StatementEnd
