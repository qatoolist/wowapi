-- Gap-free per-tenant sequence allocator (roadmap E3). Statutory numbered series
-- (receipts, vouchers, certificates) must be gap-free and race-free — exactly
-- what MAX()+1 is not. The allocator increments a per-(tenant,series) counter row
-- INSIDE the caller's transaction, so the number is consumed only if the business
-- write commits (a rollback frees it — gap-free) and concurrent allocations
-- serialize on the row lock (race-free). Voids are recorded, never renumbered.

-- +goose Up

CREATE TABLE sequences (
    tenant_id   uuid   NOT NULL,
    series_key  text   NOT NULL,        -- e.g. 'receipt', 'voucher:2026'
    next_value  bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (tenant_id, series_key)
);

ALTER TABLE sequences ENABLE ROW LEVEL SECURITY;
ALTER TABLE sequences FORCE ROW LEVEL SECURITY;
CREATE POLICY sequences_tenant_isolation ON sequences
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());
GRANT SELECT, INSERT, UPDATE ON sequences TO app_rt;

-- The allocation ledger: one immutable row per issued number, with an audited
-- void path. UNIQUE(tenant, series, value) is the second guard against a
-- duplicated statutory number.
CREATE TABLE sequence_allocations (
    id           uuid PRIMARY KEY,
    tenant_id    uuid   NOT NULL,
    series_key   text   NOT NULL,
    value        bigint NOT NULL,
    allocated_at timestamptz NOT NULL DEFAULT now(),
    allocated_by uuid,                  -- app.actor_id when present
    voided_at    timestamptz,
    void_reason  text,
    UNIQUE (tenant_id, series_key, value)
);

ALTER TABLE sequence_allocations ENABLE ROW LEVEL SECURITY;
ALTER TABLE sequence_allocations FORCE ROW LEVEL SECURITY;
CREATE POLICY sequence_allocations_tenant_isolation ON sequence_allocations
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());
GRANT SELECT, INSERT, UPDATE ON sequence_allocations TO app_rt;

-- +goose Down

DROP TABLE IF EXISTS sequence_allocations;
DROP TABLE IF EXISTS sequences;
