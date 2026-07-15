-- Negative fixture: a migration that adds a single-column tenant FK on a
-- tenant-scoped table. The tenantfk gate must reject this.

-- +goose Up
ALTER TABLE persons
  ADD CONSTRAINT persons_party_id_bad_fkey
  FOREIGN KEY (party_id) REFERENCES parties (id);

-- +goose Down
ALTER TABLE persons DROP CONSTRAINT IF EXISTS persons_party_id_bad_fkey;
