-- Audit tamper-evidence via hash-chaining (roadmap S6). Layers on the append-only
-- audit_logs table (00017): every row gets a per-tenant monotonic seq and a
-- row_hash = sha256(prev_hash || canonical(row)). Any mutation of a past row
-- changes its hash and breaks every subsequent link; any deletion leaves a seq
-- gap. A verification pass recomputes the chain and detects either. audit_chain
-- holds each tenant's head (next seq + latest hash), advanced atomically with the
-- row insert inside the caller's transaction. app_rt still cannot UPDATE/DELETE
-- audit_logs (00017) — the chain adds detection on top of that prevention.

-- +goose Up

ALTER TABLE audit_logs ADD COLUMN seq       bigint NOT NULL;
ALTER TABLE audit_logs ADD COLUMN row_hash  text   NOT NULL;
ALTER TABLE audit_logs ADD COLUMN prev_hash text   NOT NULL DEFAULT '';
CREATE UNIQUE INDEX audit_logs_chain ON audit_logs (tenant_id, seq);

CREATE TABLE audit_chain (
    tenant_id uuid PRIMARY KEY,
    next_seq  bigint NOT NULL DEFAULT 1,
    head_hash text   NOT NULL DEFAULT ''    -- genesis prev_hash is the empty string
);

ALTER TABLE audit_chain ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_chain FORCE ROW LEVEL SECURITY;
CREATE POLICY audit_chain_tenant_isolation ON audit_chain
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());
-- The head advances with each audit row; the runtime may read/insert/update the
-- head but the audit_logs rows themselves remain append-only (00017).
GRANT SELECT, INSERT, UPDATE ON audit_chain TO app_rt;

-- +goose Down

DROP TABLE IF EXISTS audit_chain;
DROP INDEX IF EXISTS audit_logs_chain;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS prev_hash;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS row_hash;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS seq;
