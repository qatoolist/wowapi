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

// UpsertAs writes (or updates) the resources mirror row attributing
// created_by (and, on update, updated_by) to actorID explicitly. It is the
// aggregate write helper's mirror stage (kernel/resource/aggregate, DATA-06
// T1/T2): the helper resolves the acting principal from context — rejecting a
// user-initiated write with no resolvable actor — before calling it, so every
// mirror row written through the enforced path carries real attribution.
//
// DATA-07 T3 note (single-owner fix surface): this method IS the nil-actor
// placeholder fix. Later actor-attribution work must route through it (or
// through the ctx-sourcing Upsert below) rather than reintroducing an
// independent created_by write.
func (r *PgRegistrar) UpsertAs(ctx context.Context, db database.TenantDB, actorID uuid.UUID, ref Ref, orgID *uuid.UUID, label, status string) error {
	return upsertMirror(ctx, db, actorID, ref, orgID, label, status)
}

// boundRegistrar implements Registrar over a single tenant tx's TenantDB.
type boundRegistrar struct{ db database.TenantDB }

var _ Registrar = (*boundRegistrar)(nil)

// Upsert writes (or updates) the kernel resources row mirroring a module
// aggregate: same id, its resource type, optional org, label and status.
// tenant_id is set from app_tenant_id() so the RLS WITH CHECK is satisfied and
// the row can never be written under the wrong tenant.
//
// created_by/updated_by are sourced from the actor bound in ctx
// (database.ActorIDFrom — the same binding TxManager mirrors into the
// app.actor_id GUC). Missing or zero actor attribution is rejected.
func (b *boundRegistrar) Upsert(ctx context.Context, ref Ref, orgID *uuid.UUID, label, status string) error {
	actorID, ok := database.ActorIDFrom(ctx)
	if !ok || actorID == uuid.Nil {
		return kerr.E(kerr.KindValidation, "actor_required", "resource mirror write requires an actor-bound context")
	}
	return upsertMirror(ctx, b.db, actorID, ref, orgID, label, status)
}

// upsertMirror is the single mirror-write statement shared by the attributed
// (UpsertAs) and ctx-sourced (Upsert) paths.
func upsertMirror(ctx context.Context, db database.TenantDB, actorID uuid.UUID, ref Ref, orgID *uuid.UUID, label, status string) error {
	const q = `
INSERT INTO resources (id, tenant_id, resource_type, org_id, label, status, version, created_at, created_by)
VALUES ($1, app_tenant_id(), $2, $3, $4, $5, 1, now(), $6)
ON CONFLICT (id) DO UPDATE SET
    label      = EXCLUDED.label,
    status     = EXCLUDED.status,
    org_id     = EXCLUDED.org_id,
    updated_at = now(),
    updated_by = EXCLUDED.created_by,
    version    = resources.version + 1`

	var org any
	if orgID != nil {
		org = *orgID
	}
	if _, err := db.Exec(ctx, q, ref.ID, ref.Type, org, label, status, actorID); err != nil {
		return kerr.Wrapf(err, "resource.Upsert", "upsert resources mirror for %s", ref.Type)
	}
	return nil
}
