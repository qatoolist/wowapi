-- Data lifecycle: generalized legal hold + DSR request lifecycle (roadmap E2).
-- The document framework already had a per-document legal_hold flag; this
-- generalizes hold to ANY (entity_type, entity_id) so any retention sweep can
-- consult it, and adds a Data Subject Request ledger (export/erasure) with a
-- statutory-override reason for rejecting an erasure that a retention obligation
-- forbids. Per-record-class disposition over product tables is orchestrated by
-- the scheduler (00014) with product-supplied callbacks; these two tables are the
-- concrete, framework-owned primitives. Tenant-scoped under RLS.

-- +goose Up

CREATE TABLE legal_holds (
    id          uuid PRIMARY KEY,
    tenant_id   uuid NOT NULL,
    entity_type text NOT NULL,
    entity_id   uuid NOT NULL,
    reason      text NOT NULL,
    placed_at   timestamptz NOT NULL DEFAULT now(),
    placed_by   uuid,
    released_at timestamptz,
    released_by uuid
);
-- At most one ACTIVE hold per entity.
CREATE UNIQUE INDEX legal_holds_active ON legal_holds (tenant_id, entity_type, entity_id)
    WHERE released_at IS NULL;

ALTER TABLE legal_holds ENABLE ROW LEVEL SECURITY;
ALTER TABLE legal_holds FORCE ROW LEVEL SECURITY;
CREATE POLICY legal_holds_tenant_isolation ON legal_holds
    USING (tenant_id = app_tenant_id()) WITH CHECK (tenant_id = app_tenant_id());
GRANT SELECT, INSERT, UPDATE ON legal_holds TO app_rt;

CREATE TABLE dsr_requests (
    id              uuid PRIMARY KEY,
    tenant_id       uuid NOT NULL,
    subject_ref     text NOT NULL,             -- product subject identifier (party id, email, …)
    kind            text NOT NULL CHECK (kind IN ('export','erasure')),
    status          text NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','completed','rejected')),
    override_reason text,                       -- statutory reason an erasure was refused
    requested_at    timestamptz NOT NULL DEFAULT now(),
    requested_by    uuid,
    completed_at    timestamptz
);
CREATE INDEX dsr_requests_subject ON dsr_requests (tenant_id, subject_ref, requested_at DESC);

ALTER TABLE dsr_requests ENABLE ROW LEVEL SECURITY;
ALTER TABLE dsr_requests FORCE ROW LEVEL SECURITY;
CREATE POLICY dsr_requests_tenant_isolation ON dsr_requests
    USING (tenant_id = app_tenant_id()) WITH CHECK (tenant_id = app_tenant_id());
GRANT SELECT, INSERT, UPDATE ON dsr_requests TO app_rt;

-- +goose Down

DROP TABLE IF EXISTS dsr_requests;
DROP TABLE IF EXISTS legal_holds;
