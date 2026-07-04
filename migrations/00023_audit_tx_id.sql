-- Add a forensic transaction-id column to audit_logs (roadmap E1/CA-11): audit
-- rows written inside the SAME database transaction can then be correlated by
-- tx_id. It is populated with pg_current_xact_id() at insert time by
-- audit.Writer.Record. tx_id is derived provenance, NOT part of the tamper-
-- evidence hash chain (like metadata), so it does not affect chain verification.

-- +goose Up

ALTER TABLE audit_logs ADD COLUMN tx_id text;

-- +goose Down

ALTER TABLE audit_logs DROP COLUMN tx_id;
