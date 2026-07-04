// Package sequence provides a gap-free, race-free per-tenant numbered-series
// allocator for statutory documents (receipts, vouchers, certificates) — the
// primitive that keeps products off MAX()+1 (roadmap E3).
//
// Allocate runs INSIDE the caller's tenant transaction: the counter increment
// commits or rolls back atomically with the business write, so a number is
// consumed only on commit (gap-free) and concurrent allocations serialize on the
// counter row (race-free). This is deliberately NOT a Postgres sequence —
// nextval() does not roll back and therefore leaves gaps, which is unacceptable
// for statutory numbering. The cost is that allocations on one series serialize;
// that is inherent to gap-free numbering.
package sequence

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
)

// Allocator issues and voids numbers in per-tenant series. It is stateless; all
// state lives in the sequences / sequence_allocations tables under RLS.
type Allocator struct {
	idgen model.IDGen
}

// New builds an Allocator. idgen mints ledger row ids (UUIDv7).
func New(idgen model.IDGen) *Allocator {
	if idgen == nil {
		idgen = model.UUIDv7()
	}
	return &Allocator{idgen: idgen}
}

// Allocation is a single issued number and its ledger identity.
type Allocation struct {
	ID    uuid.UUID
	Value int64
}

// Allocate consumes the next value in seriesKey for the current tenant and writes
// a ledger row, all within db's transaction. Because the increment lives in the
// caller's tx, a rollback frees the number (gap-free) and parallel callers block
// on the counter row rather than colliding (race-free). seriesKey must be
// non-empty.
func (a *Allocator) Allocate(ctx context.Context, db database.TenantDB, seriesKey string) (Allocation, error) {
	if seriesKey == "" {
		return Allocation{}, kerr.E(kerr.KindValidation, "invalid_series", "sequence series key is required")
	}
	var value int64
	// INSERT … ON CONFLICT DO UPDATE takes the row lock and returns the new value
	// atomically; app_tenant_id() binds it to the caller's tenant (RLS-checked).
	if err := db.QueryRow(ctx,
		`INSERT INTO sequences (tenant_id, series_key, next_value)
		 VALUES (app_tenant_id(), $1, 1)
		 ON CONFLICT (tenant_id, series_key)
		 DO UPDATE SET next_value = sequences.next_value + 1
		 RETURNING next_value`, seriesKey).Scan(&value); err != nil {
		return Allocation{}, kerr.Wrapf(err, "sequence.Allocate", "advance series %q", seriesKey)
	}

	id := a.idgen.New()
	if _, err := db.Exec(ctx,
		`INSERT INTO sequence_allocations (id, tenant_id, series_key, value, allocated_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4)`,
		id, seriesKey, value, actorOrNil(ctx)); err != nil {
		return Allocation{}, kerr.Wrapf(err, "sequence.Allocate", "record allocation for series %q", seriesKey)
	}
	return Allocation{ID: id, Value: value}, nil
}

// Void marks an issued number voided with a reason, for audit. The number is NOT
// reissued — the gap is intentional and traceable (a voided statutory document
// leaves a hole, it is never silently renumbered). Returns KindNotFound if the
// value was never allocated in this series, KindConflict if already voided.
func (a *Allocator) Void(ctx context.Context, db database.TenantDB, seriesKey string, value int64, reason string) error {
	tag, err := db.Exec(ctx,
		`UPDATE sequence_allocations
		    SET voided_at = now(), void_reason = $3
		  WHERE tenant_id = app_tenant_id() AND series_key = $1 AND value = $2
		    AND voided_at IS NULL`,
		seriesKey, value, reason)
	if err != nil {
		return kerr.Wrapf(err, "sequence.Void", "void %q #%d", seriesKey, value)
	}
	if tag.RowsAffected() == 0 {
		// Distinguish "never allocated" from "already voided" for a clear error.
		var exists bool
		if err := db.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM sequence_allocations
			                 WHERE tenant_id = app_tenant_id() AND series_key = $1 AND value = $2)`,
			seriesKey, value).Scan(&exists); err != nil {
			return kerr.Wrapf(err, "sequence.Void", "check allocation")
		}
		if exists {
			return kerr.E(kerr.KindConflict, "already_voided", "that allocation is already voided")
		}
		return kerr.E(kerr.KindNotFound, "not_found", "no such allocation in the series")
	}
	return nil
}

// Peek returns the last value issued in a series without consuming one (0 if the
// series has never allocated). Read path — safe in a read-only tx.
func (a *Allocator) Peek(ctx context.Context, db database.TenantDB, seriesKey string) (int64, error) {
	var v int64
	err := db.QueryRow(ctx,
		`SELECT COALESCE(next_value, 0) FROM sequences
		  WHERE tenant_id = app_tenant_id() AND series_key = $1`, seriesKey).Scan(&v)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil // series never allocated
		}
		return 0, kerr.Wrapf(err, "sequence.Peek", "read series %q", seriesKey)
	}
	return v, nil
}

func actorOrNil(ctx context.Context) any {
	if id, ok := database.ActorIDFrom(ctx); ok {
		return id
	}
	return nil
}
