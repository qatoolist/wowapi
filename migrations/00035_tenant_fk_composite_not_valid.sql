-- DATA-01 (W02-E02-S002-T4): composite tenant FKs for the 8 confirmed edges.
-- Added NOT VALID so the metadata-only add stays within the online lock budget;
-- 00036 validates each constraint and removes the now-redundant single-column FKs.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 30000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM pg_constraint con JOIN pg_class child ON child.oid=con.conrelid JOIN pg_namespace n ON n.oid=child.relnamespace WHERE n.nspname='public' AND con.contype='f' AND con.conname LIKE '%_tenant_fkey' AND con.convalidated=false
-- rollback_plan: goose Down drops the 8 composite FK constraints; the old single-column FKs remain until 00036.
-- +wowapi:end

-- +goose Up

SET LOCAL lock_timeout = '2s';
SET LOCAL statement_timeout = '30s';

ALTER TABLE persons
  ADD CONSTRAINT persons_party_id_tenant_fkey
  FOREIGN KEY (tenant_id, party_id) REFERENCES parties (tenant_id, id) NOT VALID;

ALTER TABLE legal_entities
  ADD CONSTRAINT legal_entities_party_id_tenant_fkey
  FOREIGN KEY (tenant_id, party_id) REFERENCES parties (tenant_id, id) NOT VALID;

ALTER TABLE party_contacts
  ADD CONSTRAINT party_contacts_party_id_tenant_fkey
  FOREIGN KEY (tenant_id, party_id) REFERENCES parties (tenant_id, id) NOT VALID;

ALTER TABLE acting_capacities
  ADD CONSTRAINT acting_capacities_party_id_tenant_fkey
  FOREIGN KEY (tenant_id, party_id) REFERENCES parties (tenant_id, id) NOT VALID;

ALTER TABLE resources
  ADD CONSTRAINT resources_org_id_tenant_fkey
  FOREIGN KEY (tenant_id, org_id) REFERENCES organizations (tenant_id, id) NOT VALID;

ALTER TABLE document_versions
  ADD CONSTRAINT document_versions_document_id_tenant_fkey
  FOREIGN KEY (tenant_id, document_id) REFERENCES documents (tenant_id, id) NOT VALID;

ALTER TABLE document_access_grants
  ADD CONSTRAINT document_access_grants_document_id_tenant_fkey
  FOREIGN KEY (tenant_id, document_id) REFERENCES documents (tenant_id, id) NOT VALID;

ALTER TABLE attachments
  ADD CONSTRAINT attachments_document_version_id_tenant_fkey
  FOREIGN KEY (tenant_id, document_version_id) REFERENCES document_versions (tenant_id, id) NOT VALID;

-- +goose Down

ALTER TABLE persons DROP CONSTRAINT IF EXISTS persons_party_id_tenant_fkey;
ALTER TABLE legal_entities DROP CONSTRAINT IF EXISTS legal_entities_party_id_tenant_fkey;
ALTER TABLE party_contacts DROP CONSTRAINT IF EXISTS party_contacts_party_id_tenant_fkey;
ALTER TABLE acting_capacities DROP CONSTRAINT IF EXISTS acting_capacities_party_id_tenant_fkey;
ALTER TABLE resources DROP CONSTRAINT IF EXISTS resources_org_id_tenant_fkey;
ALTER TABLE document_versions DROP CONSTRAINT IF EXISTS document_versions_document_id_tenant_fkey;
ALTER TABLE document_access_grants DROP CONSTRAINT IF EXISTS document_access_grants_document_id_tenant_fkey;
ALTER TABLE attachments DROP CONSTRAINT IF EXISTS attachments_document_version_id_tenant_fkey;
