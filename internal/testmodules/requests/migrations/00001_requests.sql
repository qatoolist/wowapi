-- Module: requests — fixture test module migration (blueprint 08 §2, 11 §4).
-- This is a MODULE migration source "requests"; numbering is independent of
-- the kernel source "wowapi". Goose histories are keyed by source name so
-- 00001 here does not collide with 00001_bootstrap.sql in the kernel.
--
-- Table: requests_request — tenant-scoped, RLS-enforced, audit columns per
-- blueprint 03 §1 convention (id, tenant_id, created_at/by, updated_at/by,
-- version for optimistic locking).

-- +goose Up

CREATE TABLE requests_request (
    id         uuid        PRIMARY KEY,
    tenant_id  uuid        NOT NULL REFERENCES tenants(id),
    title      text        NOT NULL,
    status     text        NOT NULL DEFAULT 'open'
                               CHECK (status IN ('open', 'in_progress', 'closed')),
    version    int         NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    created_by uuid        NOT NULL,
    updated_at timestamptz,
    updated_by uuid
);

-- RLS: every tenant-scoped table MUST enable AND force RLS so that even a
-- superuser login cannot bypass isolation (blueprint 03 §1, SEC-11/SEC-12).
ALTER TABLE requests_request ENABLE ROW LEVEL SECURITY;
ALTER TABLE requests_request FORCE ROW LEVEL SECURITY;

-- Policy: rows are visible/writable only when tenant_id matches the
-- transaction-local app.tenant_id setting (set by TxManager.WithTenant).
CREATE POLICY requests_request_tenant ON requests_request
    USING      (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());

-- Grant least-privilege to the runtime role. No DELETE: lifecycle via status.
GRANT SELECT, INSERT, UPDATE ON requests_request TO app_rt;

-- +goose Down

DROP TABLE IF EXISTS requests_request;
