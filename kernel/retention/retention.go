// Package retention is the data-lifecycle layer (roadmap E2): a generalized legal
// hold over any entity (not just documents) and a Data Subject Request ledger
// (export/erasure) with a statutory-override reason. Per-record-class disposition
// over product tables is orchestrated by the scheduler with product-supplied
// callbacks; these are the concrete, framework-owned primitives a compliance
// product would otherwise hand-roll. All operations run in the caller's tenant
// transaction (RLS-scoped).
package retention

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
)

// Holds manages generalized legal holds. A hold on (entityType, entityID) blocks
// disposition of that entity until released — retention sweeps consult IsHeld.
type Holds struct {
	idgen model.IDGen
}

// NewHolds builds the legal-hold service.
func NewHolds(idgen model.IDGen) *Holds {
	if idgen == nil {
		idgen = model.UUIDv7()
	}
	return &Holds{idgen: idgen}
}

// Hold is an active legal hold.
type Hold struct {
	ID         uuid.UUID
	EntityType string
	EntityID   uuid.UUID
	Reason     string
}

// Place puts an entity under legal hold. Reason is required. A second active hold
// on the same entity is a KindConflict (there is at most one active hold).
func (h *Holds) Place(ctx context.Context, db database.TenantDB, entityType string, entityID uuid.UUID, reason string) (uuid.UUID, error) {
	if entityType == "" || entityID == uuid.Nil || reason == "" {
		return uuid.Nil, kerr.E(kerr.KindValidation, "invalid_hold", "entity type, id, and reason are required")
	}
	id := h.idgen.New()
	_, err := db.Exec(ctx,
		`INSERT INTO legal_holds (id, tenant_id, entity_type, entity_id, reason, placed_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5)`,
		id, entityType, entityID, reason, actorOrNil(ctx))
	if err != nil {
		if isUniqueViolation(err) {
			return uuid.Nil, kerr.E(kerr.KindConflict, "already_held", "entity is already under an active legal hold")
		}
		return uuid.Nil, kerr.Wrapf(err, "retention.Place", "insert hold")
	}
	return id, nil
}

// Release lifts an active hold by id. KindNotFound if it is not an active hold.
func (h *Holds) Release(ctx context.Context, db database.TenantDB, id uuid.UUID) error {
	tag, err := db.Exec(ctx,
		`UPDATE legal_holds SET released_at = now(), released_by = $2
		  WHERE id = $1 AND released_at IS NULL`, id, actorOrNil(ctx))
	if err != nil {
		return kerr.Wrapf(err, "retention.Release", "release hold")
	}
	if tag.RowsAffected() == 0 {
		return kerr.E(kerr.KindNotFound, "not_found", "no active hold with that id")
	}
	return nil
}

// IsHeld reports whether the entity has an active legal hold.
func (h *Holds) IsHeld(ctx context.Context, db database.TenantDB, entityType string, entityID uuid.UUID) (bool, error) {
	var held bool
	if err := db.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM legal_holds
		                 WHERE entity_type = $1 AND entity_id = $2 AND released_at IS NULL)`,
		entityType, entityID).Scan(&held); err != nil {
		return false, kerr.Wrapf(err, "retention.IsHeld", "check hold")
	}
	return held, nil
}

// List returns the tenant's active legal holds.
func (h *Holds) List(ctx context.Context, db database.TenantDB) ([]Hold, error) {
	rows, err := db.Query(ctx,
		`SELECT id, entity_type, entity_id, reason FROM legal_holds
		  WHERE released_at IS NULL ORDER BY placed_at DESC`)
	if err != nil {
		return nil, kerr.Wrapf(err, "retention.List", "query holds")
	}
	defer rows.Close()
	var out []Hold
	for rows.Next() {
		var hd Hold
		if err := rows.Scan(&hd.ID, &hd.EntityType, &hd.EntityID, &hd.Reason); err != nil {
			return nil, kerr.Wrapf(err, "retention.List", "scan hold")
		}
		out = append(out, hd)
	}
	return out, rows.Err()
}

func actorOrNil(ctx context.Context) any {
	if id, ok := database.ActorIDFrom(ctx); ok {
		return id
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return false
}
