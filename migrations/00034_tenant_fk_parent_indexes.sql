-- DATA-01 (W02-E02-S001-T1): parent-side UNIQUE (tenant_id, id) indexes for the
-- 8 tenant-scoped child-table composite FKs added in 00035. Built CONCURRENTLY
-- so concurrent DML on the parent tables is not blocked.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 600000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM pg_indexes WHERE schemaname='public' AND tablename IN ('parties','organizations','documents','document_versions') AND indexdef LIKE '%UNIQUE%tenant_id, id%'
-- rollback_plan: goose Down drops the four unique indexes; additive-only index creation, no data loss.
-- +wowapi:end

-- +goose NO TRANSACTION

-- +goose Up

-- CONCURRENTLY must run outside a transaction. Use session-level timeouts and
-- reset them at the end so the connection is left in a predictable state.
SET lock_timeout = '2s';
SET statement_timeout = '10min';

CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS parties_tenant_id_id_uidx ON parties (tenant_id, id);
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS organizations_tenant_id_id_uidx ON organizations (tenant_id, id);
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS documents_tenant_id_id_uidx ON documents (tenant_id, id);
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS document_versions_tenant_id_id_uidx ON document_versions (tenant_id, id);

RESET lock_timeout;
RESET statement_timeout;

-- +goose Down

DROP INDEX IF EXISTS parties_tenant_id_id_uidx;
DROP INDEX IF EXISTS organizations_tenant_id_id_uidx;
DROP INDEX IF EXISTS documents_tenant_id_id_uidx;
DROP INDEX IF EXISTS document_versions_tenant_id_id_uidx;
