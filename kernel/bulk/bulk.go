// Package bulk is the chunked bulk-operation framework (roadmap E6): start a set
// of items, process them in caller-sized chunks with per-item isolation, record
// a partial-failure ledger, and resume after an interruption.
//
// Each item is processed in its own transaction, and its success status commits
// ATOMICALLY with the item's work — so a crash re-processes only the items not
// yet marked done (resumable), and one item's failure neither rolls back the
// others nor stops the run (partial-failure ledger). Item work must be
// idempotent, like a job worker (a re-run after a crash may repeat the last
// in-flight item).
package bulk

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/model"
)

// ItemFunc processes one item's payload inside the item's own tenant transaction.
// Returning an error records the item as failed (with the error) and rolls back
// any work it did; returning nil commits the work together with the done mark.
type ItemFunc func(ctx context.Context, db database.TenantDB, payload []byte) error

// Progress is a bulk operation's live counts.
type Progress struct {
	Total   int
	Done    int
	Failed  int
	Pending int
	Status  string // pending | running | completed
}

// Service starts and drives bulk operations.
type Service struct {
	idgen model.IDGen
}

// New builds the service. idgen mints operation and item ids.
func New(idgen model.IDGen) *Service {
	if idgen == nil {
		idgen = model.UUIDv7()
	}
	return &Service{idgen: idgen}
}

// Start creates a bulk operation of kind with one pending item per payload, in
// the caller's tenant transaction (so the operation and its items commit with any
// business write that spawned them). Returns the operation id. An empty item set
// is allowed (a completed no-op operation).
func (s *Service) Start(ctx context.Context, db database.TenantDB, kind string, items []json.RawMessage) (uuid.UUID, error) {
	if kind == "" {
		return uuid.Nil, kerr.E(kerr.KindValidation, "invalid_kind", "bulk operation kind is required")
	}
	opID := s.idgen.New()
	if _, err := db.Exec(ctx,
		`INSERT INTO bulk_operations (id, tenant_id, kind, total_items, status, created_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5)`,
		opID, kind, len(items), statusFor(len(items)), actorOrNil(ctx)); err != nil {
		return uuid.Nil, kerr.Wrapf(err, "bulk.Start", "insert operation")
	}
	for i, payload := range items {
		if len(payload) == 0 {
			payload = json.RawMessage("{}")
		}
		if _, err := db.Exec(ctx,
			`INSERT INTO bulk_items (id, bulk_id, tenant_id, seq, payload)
			 VALUES ($1, $2, app_tenant_id(), $3, $4)`,
			s.idgen.New(), opID, i, []byte(payload)); err != nil {
			return uuid.Nil, kerr.Wrapf(err, "bulk.Start", "insert item %d", i)
		}
	}
	return opID, nil
}

// Process runs up to `limit` pending items of the operation (limit <= 0 = all
// remaining), each in its own tenant transaction, and returns how many it
// processed this call. Re-run to continue — Process is resumable and picks up
// only still-pending items. When no pending items remain it marks the operation
// completed. txm is the tenant runtime TxManager (app_rt); a worker drives this
// per operation, binding tenantID for each item's transaction.
func (s *Service) Process(ctx context.Context, txm database.TxManager, tenantID, bulkID uuid.UUID, limit int, fn ItemFunc) (int, error) {
	tctx := database.WithTenantID(ctx, tenantID)
	if err := s.mark(tctx, txm, bulkID, "running"); err != nil {
		return 0, err
	}

	processed := 0
	for limit <= 0 || processed < limit {
		if ctx.Err() != nil {
			return processed, ctx.Err()
		}
		item, ok, err := s.next(tctx, txm, bulkID)
		if err != nil {
			return processed, err
		}
		if !ok {
			// Nothing pending — the operation is complete.
			if err := s.mark(tctx, txm, bulkID, "completed"); err != nil {
				return processed, err
			}
			break
		}
		if err := s.runItem(tctx, txm, item, fn); err != nil {
			return processed, err // an infrastructure error (not an item failure)
		}
		processed++
	}
	return processed, nil
}

type claimedItem struct {
	id      uuid.UUID
	payload []byte
}

// next reads the lowest-seq pending item (no lock — single processor per
// operation; add SKIP LOCKED here to fan out across workers later).
func (s *Service) next(ctx context.Context, txm database.TxManager, bulkID uuid.UUID) (claimedItem, bool, error) {
	var it claimedItem
	found := false
	err := txm.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		row := db.QueryRow(ctx,
			`SELECT id, payload FROM bulk_items
			  WHERE bulk_id = $1 AND status = 'pending'
			  ORDER BY seq
			  LIMIT 1`, bulkID)
		if err := row.Scan(&it.id, &it.payload); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return kerr.Wrapf(err, "bulk.next", "read pending item")
		}
		found = true
		return nil
	})
	return it, found, err
}

// runItem executes one item. On success, fn's work and the 'done' mark commit in
// ONE transaction (atomic, resumable). On failure, fn's transaction is rolled
// back and a SECOND transaction records 'failed' + the error — so a partial write
// never lingers but the failure is durably ledgered.
func (s *Service) runItem(ctx context.Context, txm database.TxManager, item claimedItem, fn ItemFunc) error {
	fnErr := txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		if err := fn(ctx, db, item.payload); err != nil {
			return err
		}
		// Mark done in the SAME tx as the work; the WHERE guard makes a re-run a
		// no-op if the item was already completed.
		if _, err := db.Exec(ctx,
			`UPDATE bulk_items SET status = 'done', attempts = attempts + 1, processed_at = now()
			  WHERE id = $1 AND status = 'pending'`, item.id); err != nil {
			return err
		}
		return nil
	})
	if fnErr == nil {
		return nil
	}
	// The item failed: fn's writes rolled back with the tx above. Record the
	// failure in its own transaction so it survives.
	return txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		if _, err := db.Exec(ctx,
			`UPDATE bulk_items
			    SET status = 'failed', attempts = attempts + 1,
			        last_error = left($2, 1000), processed_at = now()
			  WHERE id = $1 AND status = 'pending'`, item.id, fnErr.Error()); err != nil {
			return kerr.Wrapf(err, "bulk.runItem", "record failure")
		}
		return nil
	})
}

// mark sets the operation status (idempotent).
func (s *Service) mark(ctx context.Context, txm database.TxManager, bulkID uuid.UUID, status string) error {
	return txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		if _, err := db.Exec(ctx,
			`UPDATE bulk_operations SET status = $2, updated_at = now() WHERE id = $1`, bulkID, status); err != nil {
			return kerr.Wrapf(err, "bulk.mark", "set status %s", status)
		}
		return nil
	})
}

// Progress reports live counts for an operation, in the caller's tenant tx.
func (s *Service) Progress(ctx context.Context, db database.TenantDB, bulkID uuid.UUID) (Progress, error) {
	var p Progress
	if err := db.QueryRow(ctx,
		`SELECT total_items, status FROM bulk_operations WHERE id = $1`, bulkID).
		Scan(&p.Total, &p.Status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Progress{}, kerr.E(kerr.KindNotFound, "not_found", "no such bulk operation")
		}
		return Progress{}, kerr.Wrapf(err, "bulk.Progress", "read operation")
	}
	if err := db.QueryRow(ctx,
		`SELECT
		    count(*) FILTER (WHERE status = 'done'),
		    count(*) FILTER (WHERE status = 'failed'),
		    count(*) FILTER (WHERE status = 'pending')
		  FROM bulk_items WHERE bulk_id = $1`, bulkID).
		Scan(&p.Done, &p.Failed, &p.Pending); err != nil {
		return Progress{}, kerr.Wrapf(err, "bulk.Progress", "count items")
	}
	return p, nil
}

func statusFor(n int) string {
	if n == 0 {
		return "completed"
	}
	return "pending"
}

func actorOrNil(ctx context.Context) any {
	if id, ok := database.ActorIDFrom(ctx); ok {
		return id
	}
	return nil
}
