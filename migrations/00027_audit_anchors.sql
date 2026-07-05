-- Durable audit-chain anchors (roadmap CA-11). The hash chain (00018) makes any
-- mutation or deletion of a PAST row detectable by Verify — but a *tail*
-- truncation (drop the last k rows AND rewind audit_chain.head_hash to match) is
-- undetectable by Verify alone, because the shortened chain is still internally
-- consistent. An anchor closes that gap: an append-only, immutable snapshot of a
-- tenant's chain head (last seq + its row_hash) taken periodically. An offline
-- verifier detects truncation by checking the live chain still contains the last
-- anchored (seq, hash); if the head has rewound below the anchored seq, or the
-- anchored row_hash is gone, the tail was truncated after the anchor was taken.
--
-- Writes come from the leader-safe anchor-export sweep running cross-tenant as
-- app_platform (one INSERT..SELECT for all tenants, like the idempotency sweep in
-- 00012). app_rt is read-only and tenant-scoped: a tenant can read its own
-- anchors as evidence but can never forge or delete one.

-- +goose Up

CREATE TABLE audit_anchors (
    id              bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id       uuid        NOT NULL,
    anchor_seq      bigint      NOT NULL,   -- last seq in the chain at snapshot time
    chain_head_hash text        NOT NULL,   -- row_hash of the row at anchor_seq
    row_count       bigint      NOT NULL,   -- rows covered (== anchor_seq; seq is gap-free)
    created_at      timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX audit_anchors_tenant_seq ON audit_anchors (tenant_id, anchor_seq);

ALTER TABLE audit_anchors ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_anchors FORCE ROW LEVEL SECURITY;

-- app_rt: tenant-scoped, read-only. No INSERT/UPDATE/DELETE grant — anchors are
-- immutable evidence the runtime may consult but never author or alter.
CREATE POLICY audit_anchors_tenant_read ON audit_anchors
    FOR SELECT TO app_rt USING (tenant_id = app_tenant_id());
GRANT SELECT ON audit_anchors TO app_rt;

-- app_platform: cross-tenant append (the scheduled anchor-export). SELECT+INSERT
-- only (no UPDATE/DELETE) keeps the table append-only even for the writer.
CREATE POLICY audit_anchors_platform_write ON audit_anchors
    TO app_platform USING (true) WITH CHECK (true);
GRANT SELECT, INSERT ON audit_anchors TO app_platform;

-- The anchor-export must read every tenant's chain head cross-tenant. audit_chain
-- (00018) FORCEs RLS with a tenant-isolation policy; add a permissive read policy
-- + SELECT grant for the platform role (mirrors the idempotency sweep in 00012).
CREATE POLICY audit_chain_platform_read ON audit_chain
    FOR SELECT TO app_platform USING (true);
GRANT SELECT ON audit_chain TO app_platform;

-- +goose Down

REVOKE SELECT ON audit_chain FROM app_platform;
DROP POLICY IF EXISTS audit_chain_platform_read ON audit_chain;
DROP TABLE IF EXISTS audit_anchors;
