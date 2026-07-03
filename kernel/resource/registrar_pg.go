package resource

import (
	"context"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// PgRegistrar is the Postgres-backed resource Registrar factory. It carries no
// state of its own: a Registrar must run inside the caller's tenant transaction
// so the mirror row and the business write commit atomically, and the tx's
// TenantDB is only available per unit of work. Bind supplies that db and yields
// a Registrar bound to it.
type PgRegistrar struct{}

// NewRegistrar returns the Postgres resource registrar factory.
func NewRegistrar() *PgRegistrar { return &PgRegistrar{} }

// Bind returns a Registrar that writes the resources mirror through db (the
// caller's open tenant transaction). Module code obtains db from the TxManager
// and binds the registrar for the duration of the write.
func (r *PgRegistrar) Bind(db database.TenantDB) Registrar {
	return &boundRegistrar{db: db}
}

// boundRegistrar implements Registrar over a single tenant tx's TenantDB.
type boundRegistrar struct{ db database.TenantDB }

var _ Registrar = (*boundRegistrar)(nil)

// Upsert writes (or updates) the kernel resources row mirroring a module
// aggregate: same id, its resource type, optional org, label and status.
// tenant_id is set from app_tenant_id() so the RLS WITH CHECK is satisfied and
// the row can never be written under the wrong tenant.
func (b *boundRegistrar) Upsert(ctx context.Context, ref Ref, orgID *uuid.UUID, label, status string) error {
	// created_by: NIL uuid placeholder for now. Full actor attribution wires in
	// from ctx (database.ActorIDFrom) in a later refinement; the column is
	// NOT NULL so a zero uuid keeps the insert legal without blocking Phase 4.
	// TODO(phase-later): source created_by/updated_by from the request actor.
	const q = `
INSERT INTO resources (id, tenant_id, resource_type, org_id, label, status, version, created_at, created_by)
VALUES ($1, app_tenant_id(), $2, $3, $4, $5, 1, now(), $6)
ON CONFLICT (id) DO UPDATE SET
    label      = EXCLUDED.label,
    status     = EXCLUDED.status,
    org_id     = EXCLUDED.org_id,
    updated_at = now(),
    version    = resources.version + 1`

	var org any
	if orgID != nil {
		org = *orgID
	}
	if _, err := b.db.Exec(ctx, q, ref.ID, ref.Type, org, label, status, uuid.Nil); err != nil {
		return kerr.Wrapf(err, "resource.Upsert", "upsert resources mirror for %s", ref.Type)
	}
	return nil
}
