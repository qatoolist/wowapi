-- DATA-01 (W02-E02-S002-T5/T8): validate every composite tenant FK added in
-- 00035, then drop the redundant single-column FKs. The single-column FKs had
-- NO ACTION on both update and delete, so removing them does not change
-- cascade behavior; the composite FKs now enforce both parent existence and
-- tenant agreement.

-- +wowapi:manifest
-- classification: online
-- rows_estimate: 0
-- bytes_estimate: 0
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 300000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: SELECT count(*) FROM pg_constraint con JOIN pg_class child ON child.oid=con.conrelid JOIN pg_namespace n ON n.oid=child.relnamespace WHERE n.nspname='public' AND con.contype='f' AND con.conname LIKE '%_tenant_fkey' AND con.convalidated=false
-- rollback_plan: goose Down re-adds the 8 single-column FKs; the composite FKs remain validated (VALIDATE CONSTRAINT cannot be undone).
-- +wowapi:end

-- +goose Up

SET LOCAL lock_timeout = '2s';
SET LOCAL statement_timeout = '5min';

ALTER TABLE persons VALIDATE CONSTRAINT persons_party_id_tenant_fkey;
ALTER TABLE legal_entities VALIDATE CONSTRAINT legal_entities_party_id_tenant_fkey;
ALTER TABLE party_contacts VALIDATE CONSTRAINT party_contacts_party_id_tenant_fkey;
ALTER TABLE acting_capacities VALIDATE CONSTRAINT acting_capacities_party_id_tenant_fkey;
ALTER TABLE resources VALIDATE CONSTRAINT resources_org_id_tenant_fkey;
ALTER TABLE document_versions VALIDATE CONSTRAINT document_versions_document_id_tenant_fkey;
ALTER TABLE document_access_grants VALIDATE CONSTRAINT document_access_grants_document_id_tenant_fkey;
ALTER TABLE attachments VALIDATE CONSTRAINT attachments_document_version_id_tenant_fkey;

-- The old single-column FKs are now redundant: the composite FKs enforce a
-- stricter invariant (parent exists AND parent.tenant_id = child.tenant_id).
ALTER TABLE persons DROP CONSTRAINT IF EXISTS persons_party_id_fkey;
ALTER TABLE legal_entities DROP CONSTRAINT IF EXISTS legal_entities_party_id_fkey;
ALTER TABLE party_contacts DROP CONSTRAINT IF EXISTS party_contacts_party_id_fkey;
ALTER TABLE acting_capacities DROP CONSTRAINT IF EXISTS acting_capacities_party_id_fkey;
ALTER TABLE resources DROP CONSTRAINT IF EXISTS resources_org_id_fkey;
ALTER TABLE document_versions DROP CONSTRAINT IF EXISTS document_versions_document_id_fkey;
ALTER TABLE document_access_grants DROP CONSTRAINT IF EXISTS document_access_grants_document_id_fkey;
ALTER TABLE attachments DROP CONSTRAINT IF EXISTS attachments_document_version_id_fkey;

-- +goose Down

-- Re-create the redundant single-column FKs so the schema is structurally
-- symmetric after a rollback. Validation state of the composite FKs is
-- irreversible, but the resulting schema is still valid.
ALTER TABLE persons
  ADD CONSTRAINT persons_party_id_fkey FOREIGN KEY (party_id) REFERENCES parties (id);
ALTER TABLE legal_entities
  ADD CONSTRAINT legal_entities_party_id_fkey FOREIGN KEY (party_id) REFERENCES parties (id);
ALTER TABLE party_contacts
  ADD CONSTRAINT party_contacts_party_id_fkey FOREIGN KEY (party_id) REFERENCES parties (id);
ALTER TABLE acting_capacities
  ADD CONSTRAINT acting_capacities_party_id_fkey FOREIGN KEY (party_id) REFERENCES parties (id);
ALTER TABLE resources
  ADD CONSTRAINT resources_org_id_fkey FOREIGN KEY (org_id) REFERENCES organizations (id);
ALTER TABLE document_versions
  ADD CONSTRAINT document_versions_document_id_fkey FOREIGN KEY (document_id) REFERENCES documents (id);
ALTER TABLE document_access_grants
  ADD CONSTRAINT document_access_grants_document_id_fkey FOREIGN KEY (document_id) REFERENCES documents (id);
ALTER TABLE attachments
  ADD CONSTRAINT attachments_document_version_id_fkey FOREIGN KEY (document_version_id) REFERENCES document_versions (id);
