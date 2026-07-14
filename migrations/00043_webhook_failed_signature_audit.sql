-- Failed-signature audit table for inbound webhook verification (DATA-03 T5).
-- Rows are body-free by construction: the schema has no payload/body column.
-- Each failed verification writes one row in its own short transaction.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 5000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM webhook_failed_signature_audit
-- rollback_plan: goose Down drops webhook_failed_signature_audit and its index.
-- +wowapi:end

-- +goose Up

CREATE TABLE webhook_failed_signature_audit (
    id          uuid PRIMARY KEY,
    tenant_id   uuid NOT NULL,
    endpoint_id uuid NOT NULL REFERENCES webhook_endpoints(id),
    event_type  text NOT NULL,
    signature_header text,               -- the value that failed verification, if present
    failure_reason text NOT NULL,
    received_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX wfsa_endpoint ON webhook_failed_signature_audit (tenant_id, endpoint_id, received_at DESC);

ALTER TABLE webhook_failed_signature_audit ENABLE ROW LEVEL SECURITY;
ALTER TABLE webhook_failed_signature_audit FORCE ROW LEVEL SECURITY;
CREATE POLICY webhook_failed_signature_audit_tenant_isolation ON webhook_failed_signature_audit
    USING (tenant_id = app_tenant_id())
    WITH CHECK (tenant_id = app_tenant_id());

GRANT SELECT, INSERT ON webhook_failed_signature_audit TO app_rt;
GRANT SELECT, INSERT ON webhook_failed_signature_audit TO app_platform;

-- +goose Down

DROP TABLE IF EXISTS webhook_failed_signature_audit;
