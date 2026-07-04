-- Bulk-operation framework (roadmap E6). Chunked processing of large item sets
-- with progress reporting, a partial-failure ledger, and resumability. A bulk
-- operation owns N item rows; a processor claims pending items in chunks
-- (FOR UPDATE SKIP LOCKED — safe across replicas), runs each, and records
-- done/failed per item so a crash resumes from the remaining pending items and
-- one bad item never fails the whole run. Tenant-scoped under RLS.

-- +goose Up

CREATE TABLE bulk_operations (
    id           uuid PRIMARY KEY,
    tenant_id    uuid NOT NULL,
    kind         text NOT NULL,                 -- what the items are (product-defined)
    total_items  int  NOT NULL DEFAULT 0,
    status       text NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending','running','completed')),
    created_at   timestamptz NOT NULL DEFAULT now(),
    created_by   uuid,
    updated_at   timestamptz
);

ALTER TABLE bulk_operations ENABLE ROW LEVEL SECURITY;
ALTER TABLE bulk_operations FORCE ROW LEVEL SECURITY;
CREATE POLICY bulk_operations_tenant_isolation ON bulk_operations
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());
GRANT SELECT, INSERT, UPDATE ON bulk_operations TO app_rt;

CREATE TABLE bulk_items (
    id           uuid PRIMARY KEY,
    bulk_id      uuid NOT NULL REFERENCES bulk_operations(id),
    tenant_id    uuid NOT NULL,
    seq          int  NOT NULL,                 -- position in the original set
    payload      jsonb NOT NULL DEFAULT '{}',
    status       text NOT NULL DEFAULT 'pending'
                     CHECK (status IN ('pending','done','failed')),
    attempts     int  NOT NULL DEFAULT 0,
    last_error   text,
    processed_at timestamptz,
    UNIQUE (bulk_id, seq)
);

-- Claim index: the processor polls pending items per operation.
CREATE INDEX bulk_items_pending ON bulk_items (bulk_id) WHERE status = 'pending';

ALTER TABLE bulk_items ENABLE ROW LEVEL SECURITY;
ALTER TABLE bulk_items FORCE ROW LEVEL SECURITY;
CREATE POLICY bulk_items_tenant_isolation ON bulk_items
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());
GRANT SELECT, INSERT, UPDATE ON bulk_items TO app_rt;

-- +goose Down

DROP TABLE IF EXISTS bulk_items;
DROP TABLE IF EXISTS bulk_operations;
