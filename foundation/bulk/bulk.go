// Package bulk is the chunked bulk-operation framework (roadmap E6): start a set
// of items, process them in caller-sized chunks with per-item isolation, record
// a partial-failure ledger, and resume after an interruption.
//
// Items are claimed in bounded batches via an atomic UPDATE ... FROM
// (SELECT ... FOR UPDATE SKIP LOCKED LIMIT $batch) statement, so multiple
// workers can safely process the same operation concurrently. Each item is
// processed in its own transaction, and its success/failure status commits
// ATOMICALLY with the item's work. A crash re-processes only the items not yet
// marked done (resumable), and one item's failure neither rolls back the others
// nor stops the run (partial-failure ledger). Item work must be idempotent.
package bulk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/lease"
	"github.com/qatoolist/wowapi/kernel/model"
)

// Item is one claimed bulk item passed to the worker function.
type Item struct {
	ID             uuid.UUID
	Seq            int
	Payload        []byte
	Lease          lease.Lease
	IdempotencyKey uuid.UUID
	Attempts       int // total claim attempts so far
}

// ItemFunc processes one item's payload inside the item's own tenant transaction.
// Returning an error records the item as failed (or retried, if attempts remain)
// and rolls back any work it did; returning nil commits the work together with
// the done mark.
type ItemFunc func(ctx context.Context, db database.TenantDB, item Item) error

// Progress is a bulk operation's live counts.
type Progress struct {
	Total     int
	Done      int
	Failed    int
	Pending   int
	Cancelled int
	Status    string // pending | running | paused | completed | cancelled
	Running   int    // items currently leased
}

// Option configures a Service.
type Option func(*Service)

// WithLogger replaces the logger used by the service. The default logger is
// slog.Default().
func WithLogger(log *slog.Logger) Option {
	return func(s *Service) { s.log = log }
}

// WithBatchSize sets the number of items claimed in one atomic leased-claim
// statement. It must be positive; non-positive values are ignored.
func WithBatchSize(n int) Option {
	return func(s *Service) {
		if n > 0 {
			s.batchSize = n
		}
	}
}

// WithLeaseTTL sets how long a claimed item lease remains valid. It must be
// positive; non-positive values are ignored.
func WithLeaseTTL(d time.Duration) Option {
	return func(s *Service) {
		if d > 0 {
			s.leaseTTL = d
		}
	}
}

// WithMaxAttempts sets the default per-operation retry budget for failed items.
// It must be positive; non-positive values are ignored.
func WithMaxAttempts(n int) Option {
	return func(s *Service) {
		if n > 0 {
			s.maxAttempts = n
		}
	}
}

// Service starts and drives bulk operations.
type Service struct {
	idgen       model.IDGen
	log         *slog.Logger
	batchSize   int
	leaseTTL    time.Duration
	maxAttempts int
}

// New builds the service. idgen mints operation and item ids.
func New(idgen model.IDGen, opts ...Option) *Service {
	if idgen == nil {
		idgen = model.UUIDv7()
	}
	s := &Service{
		idgen:       idgen,
		log:         slog.Default(),
		batchSize:   10,
		leaseTTL:    time.Minute,
		maxAttempts: 3,
	}
	for _, o := range opts {
		o(s)
	}
	return s
}

// ErrLeaseMismatch is returned when a finalize write is rejected because the
// caller's lease token/generation no longer matches the row (a stale or fenced
// worker).
var ErrLeaseMismatch = kerr.E(kerr.KindConflict, "lease_mismatch", "stale finalize rejected: lease token or generation mismatch")

// ErrOperationNotRunning is returned when Process is called against an operation
// that is paused, cancelled, or already completed.
var ErrOperationNotRunning = kerr.E(kerr.KindConflict, "operation_not_running", "bulk operation is not in a runnable state")

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
		`INSERT INTO bulk_operations (id, tenant_id, kind, total_items, status, max_attempts, created_by)
		 VALUES ($1, app_tenant_id(), $2, $3, $4, $5, $6)`,
		opID, kind, len(items), statusFor(len(items)), s.maxAttempts, actorOrNil(ctx)); err != nil {
		return uuid.Nil, kerr.Wrapf(err, "bulk.Start", "insert operation")
	}
	for i, payload := range items {
		if len(payload) == 0 {
			payload = json.RawMessage("{}")
		}
		if _, err := db.Exec(ctx,
			`INSERT INTO bulk_items (id, bulk_id, tenant_id, seq, payload, idempotency_key)
			 VALUES ($1, $2, app_tenant_id(), $3, $4, $5)`,
			s.idgen.New(), opID, i, []byte(payload), s.idgen.New()); err != nil {
			return uuid.Nil, kerr.Wrapf(err, "bulk.Start", "insert item %d", i)
		}
	}
	return opID, nil
}

// Process runs up to `limit` pending items of the operation (limit <= 0 = all
// remaining), each in its own tenant transaction, and returns how many it
// processed this call. Re-run to continue — Process is resumable and picks up
// only still-pending items. When no pending items remain it marks the operation
// completed. txm is the tenant runtime TxManager (app_rt); workers may drive
// this per operation, binding tenantID for each item's transaction.
//
// Multiple workers may call Process concurrently against the same bulkID; the
// leased-claim SQL guarantees disjoint batches. If the operation is paused or
// cancelled mid-run, Process returns after finishing in-flight items and
// respecting the lifecycle state.
func (s *Service) Process(ctx context.Context, txm database.TxManager, tenantID, bulkID uuid.UUID, limit int, fn ItemFunc) (int, error) {
	tctx := database.WithTenantID(ctx, tenantID)

	// Respect lifecycle state before claiming. A pending operation transitions to
	// running; a running operation resumes; paused/cancelled/completed operations
	// are left alone.
	status, err := s.operationStatus(tctx, txm, bulkID)
	if err != nil {
		return 0, err
	}
	switch status {
	case "completed", "cancelled":
		return 0, nil
	case "paused":
		return 0, nil
	}
	if status == "pending" {
		// Race-tolerant: a peer may transition first; any outcome other than
		// pending/running is re-checked by the loop below.
		if err := s.transition(tctx, txm, bulkID, "running", "pending"); err != nil && kerr.KindOf(err) != kerr.KindConflict {
			return 0, err
		}
	}

	processed := 0
	for limit <= 0 || processed < limit {
		if ctx.Err() != nil {
			return processed, ctx.Err()
		}

		status, err := s.operationStatus(tctx, txm, bulkID)
		if err != nil {
			return processed, err
		}
		switch status {
		case "paused":
			// Stop claiming new batches; in-flight items finish. Caller resumes via Resume.
			return processed, nil
		case "cancelled":
			// Cancel any remaining pending items and stop.
			if err := s.cancelPending(tctx, txm, bulkID); err != nil {
				return processed, err
			}
			return processed, nil
		case "completed":
			return processed, nil
		}

		batch, err := s.claimBatch(tctx, txm, bulkID, s.effectiveBatchSize(limit, processed))
		if err != nil {
			return processed, err
		}
		if len(batch) == 0 {
			// Complete ONLY when no nonterminal item exists (F-04): an empty
			// claim while a peer worker holds a live running item must return
			// without completing — that item may yet fail retryably and go back
			// to pending. The peer (or a later Process) completes the aggregate.
			completed, err := s.completeIfDrained(tctx, txm, bulkID)
			if err != nil {
				return processed, err
			}
			_ = completed
			break
		}

		for _, item := range batch {
			if ctx.Err() != nil {
				return processed, ctx.Err()
			}
			if err := s.runItem(tctx, txm, item, fn); err != nil {
				return processed, err // infrastructure error, not an item failure
			}
			processed++
			if limit > 0 && processed >= limit {
				break
			}
		}
	}
	return processed, nil
}

// Pause suspends a pending or running operation. In-flight items are allowed
// to finish; new batches will not be claimed until Resume is called. Terminal
// or already-paused operations are an invalid transition (F-04: lifecycle
// writes are compare-and-swap over legal source states, never labels).
func (s *Service) Pause(ctx context.Context, txm database.TxManager, tenantID, bulkID uuid.UUID) error {
	return s.transition(database.WithTenantID(ctx, tenantID), txm, bulkID, "paused", "pending", "running")
}

// Resume transitions a paused operation back to running. Only paused
// operations may resume — completed/cancelled are terminal.
func (s *Service) Resume(ctx context.Context, txm database.TxManager, tenantID, bulkID uuid.UUID) error {
	return s.transition(database.WithTenantID(ctx, tenantID), txm, bulkID, "running", "paused")
}

// Cancel stops a pending, running, or paused operation and marks all pending
// items as cancelled. In-flight items finish under their existing leases;
// subsequent Process calls will not claim new items. Terminal operations
// cannot be re-cancelled or reopened.
func (s *Service) Cancel(ctx context.Context, txm database.TxManager, tenantID, bulkID uuid.UUID) error {
	tctx := database.WithTenantID(ctx, tenantID)
	if err := s.transition(tctx, txm, bulkID, "cancelled", "pending", "running", "paused"); err != nil {
		return err
	}
	return txm.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
		if _, err := db.Exec(ctx,
			`UPDATE bulk_items SET status = 'cancelled' WHERE bulk_id = $1 AND status = 'pending'`, bulkID); err != nil {
			return kerr.Wrapf(err, "bulk.Cancel", "cancel pending items")
		}
		return nil
	})
}

// ReclaimStalled resets items whose leases have expired back to pending so other
// workers can claim them. It returns the number of items reclaimed.
func (s *Service) ReclaimStalled(ctx context.Context, txm database.TxManager, tenantID uuid.UUID, bulkID uuid.UUID) (int, error) {
	tctx := database.WithTenantID(ctx, tenantID)
	var n int64
	err := txm.WithTenant(tctx, func(ctx context.Context, db database.TenantDB) error {
		res, err := db.Exec(ctx,
			`UPDATE bulk_items
			    SET status = 'pending',
			        lease_token = NULL,
			        lease_generation = 0,
			        lease_expires_at = NULL,
			        attempts = attempts + 1
			  WHERE bulk_id = $1
			    AND status = 'running'
			    AND lease_expires_at <= now()`,
			bulkID)
		if err != nil {
			return kerr.Wrapf(err, "bulk.ReclaimStalled", "reclaim items for bulk %s", bulkID)
		}
		n = res.RowsAffected()
		return nil
	})
	if err != nil {
		return 0, err
	}
	return int(n), nil
}

const claimSQL = `WITH claimed AS (
    SELECT id
      FROM bulk_items
     WHERE bulk_id = $1
       AND (status = 'pending' OR (status = 'running' AND lease_expires_at <= now()))
     ORDER BY seq
     FOR UPDATE SKIP LOCKED
     LIMIT $5
)
UPDATE bulk_items bi
   SET status = 'running',
       lease_token = $2,
       lease_generation = $3,
       lease_expires_at = $4
  FROM claimed
 WHERE bi.id = claimed.id
RETURNING bi.id, bi.seq, bi.payload, bi.idempotency_key, bi.attempts, bi.lease_token, bi.lease_generation, bi.lease_expires_at`

// claimBatch atomically leases up to n claimable items for bulkID. Claimable
// items are pending, or running with an expired lease (reclaim path). The
// returned items carry the new lease assigned by this claim.
func (s *Service) claimBatch(ctx context.Context, txm database.TxManager, bulkID uuid.UUID, n int) ([]Item, error) {
	if n <= 0 {
		n = s.batchSize
	}
	l := lease.New(s.leaseTTL)
	var items []Item
	err := txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		rows, err := db.Query(ctx, claimSQL, bulkID, l.Token, l.Generation, l.ExpiresAt, n)
		if err != nil {
			return kerr.Wrapf(err, "bulk.claimBatch", "claim items for bulk %s", bulkID)
		}
		defer rows.Close()
		for rows.Next() {
			var it Item
			if err := rows.Scan(&it.ID, &it.Seq, &it.Payload, &it.IdempotencyKey, &it.Attempts, &it.Lease.Token, &it.Lease.Generation, &it.Lease.ExpiresAt); err != nil {
				return kerr.Wrapf(err, "bulk.claimBatch", "scan claimed item")
			}
			items = append(items, it)
		}
		if err := rows.Err(); err != nil {
			return kerr.Wrapf(err, "bulk.claimBatch", "iterate claimed items")
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return items, nil
}

// effectiveBatchSize returns the smaller of the configured batch size and the
// remaining work when limit is positive.
func (s *Service) effectiveBatchSize(limit, processed int) int {
	if limit <= 0 {
		return s.batchSize
	}
	remain := limit - processed
	if remain < s.batchSize {
		return remain
	}
	return s.batchSize
}

// runItem executes one item. On success, fn's work and the 'done' mark commit in
// ONE transaction (atomic, resumable). On failure, fn's transaction is rolled
// back and a SECOND transaction records the failure or retry state — so a
// partial write never lingers but the outcome is durably ledgered.
func (s *Service) runItem(ctx context.Context, txm database.TxManager, item Item, fn ItemFunc) error {
	maxAttempts, err := s.maxAttemptsFor(ctx, txm, item.ID)
	if err != nil {
		return err
	}

	// If this item has already exhausted its retry budget, fail it without
	// invoking the worker again.
	if item.Attempts >= maxAttempts {
		return s.recordFailure(ctx, txm, item, errors.New("max attempts exceeded"), true)
	}

	fnErr := txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		if err := fn(ctx, db, item); err != nil {
			return err
		}
		// Mark done in the SAME tx as the work, fenced by the lease.
		res, err := db.Exec(ctx,
			`UPDATE bulk_items
			    SET status = 'done', processed_at = now()
			  WHERE id = $1
			    AND status = 'running'
			    AND lease_token = $2
			    AND lease_generation = $3
			    AND lease_expires_at > now()`,
			item.ID, item.Lease.Token, item.Lease.Generation)
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return ErrLeaseMismatch
		}
		return nil
	})
	if fnErr == nil {
		return nil
	}

	// The item failed: fn's writes rolled back with the tx above. Record the
	// outcome in its own transaction so it survives.
	dead := item.Attempts+1 >= maxAttempts
	return s.recordFailure(ctx, txm, item, fnErr, dead)
}

// recordFailure writes the failed or retry-pending state for item. If dead is
// true the item is marked failed; otherwise it is returned to pending so another
// worker can retry it.
func (s *Service) recordFailure(ctx context.Context, txm database.TxManager, item Item, cause error, dead bool) error {
	return txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		var res pgconn.CommandTag
		var err error
		if dead {
			res, err = db.Exec(ctx,
				`UPDATE bulk_items
				    SET status = 'failed',
				        last_error = left($2, 1000),
				        processed_at = now(),
				        attempts = attempts + 1,
				        lease_token = NULL,
				        lease_generation = 0,
				        lease_expires_at = NULL
				  WHERE id = $1
				    AND lease_token = $3
				    AND lease_generation = $4
				    AND lease_expires_at > now()`,
				item.ID, cause.Error(), item.Lease.Token, item.Lease.Generation)
		} else {
			res, err = db.Exec(ctx,
				`UPDATE bulk_items
				    SET status = 'pending',
				        last_error = left($2, 1000),
				        processed_at = now(),
				        attempts = attempts + 1,
				        lease_token = NULL,
				        lease_generation = 0,
				        lease_expires_at = NULL
				  WHERE id = $1
				    AND lease_token = $3
				    AND lease_generation = $4
				    AND lease_expires_at > now()`,
				item.ID, cause.Error(), item.Lease.Token, item.Lease.Generation)
		}
		if err != nil {
			return kerr.Wrapf(err, "bulk.recordFailure", "record failure for item %s", item.ID)
		}
		if res.RowsAffected() == 0 {
			return ErrLeaseMismatch
		}
		return nil
	})
}

// maxAttemptsFor reads the parent operation's max_attempts for itemID.
func (s *Service) maxAttemptsFor(ctx context.Context, txm database.TxManager, itemID uuid.UUID) (int, error) {
	var maxAttempts int
	err := txm.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx,
			`SELECT bo.max_attempts
			   FROM bulk_operations bo
			   JOIN bulk_items bi ON bi.bulk_id = bo.id
			  WHERE bi.id = $1`, itemID).Scan(&maxAttempts)
	})
	if err != nil {
		return 0, kerr.Wrapf(err, "bulk.maxAttemptsFor", "read max_attempts for item %s", itemID)
	}
	return maxAttempts, nil
}

// operationStatus returns the current status of the operation.
func (s *Service) operationStatus(ctx context.Context, txm database.TxManager, bulkID uuid.UUID) (string, error) {
	var status string
	err := txm.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx,
			`SELECT status FROM bulk_operations WHERE id = $1`, bulkID).Scan(&status)
	})
	if err != nil {
		return "", kerr.Wrapf(err, "bulk.operationStatus", "read status for bulk %s", bulkID)
	}
	return status, nil
}

// cancelPending marks all still-pending items for bulkID as cancelled.
func (s *Service) cancelPending(ctx context.Context, txm database.TxManager, bulkID uuid.UUID) error {
	return txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		_, err := db.Exec(ctx,
			`UPDATE bulk_items SET status = 'cancelled' WHERE bulk_id = $1 AND status = 'pending'`, bulkID)
		return err
	})
}

// mark sets the operation status (idempotent).
// transition is the single compare-and-swap for aggregate lifecycle states: it
// moves bulkID to `to` only from one of the legal `from` states, distinguishing
// a missing operation (KindNotFound) from an illegal source state
// (KindConflict). No unconditional status label exists anymore (F-04).
func (s *Service) transition(ctx context.Context, txm database.TxManager, bulkID uuid.UUID, to string, from ...string) error {
	return txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		tag, err := db.Exec(ctx,
			`UPDATE bulk_operations SET status = $2, updated_at = now()
			  WHERE id = $1 AND status = ANY($3)`, bulkID, to, from)
		if err != nil {
			return kerr.Wrapf(err, "bulk.transition", "set status %s", to)
		}
		if tag.RowsAffected() == 1 {
			return nil
		}
		var current string
		err = db.QueryRow(ctx, `SELECT status FROM bulk_operations WHERE id = $1`, bulkID).Scan(&current)
		if errors.Is(err, pgx.ErrNoRows) {
			return kerr.E(kerr.KindNotFound, "not_found", "bulk operation not found")
		}
		if err != nil {
			return kerr.Wrapf(err, "bulk.transition", "read status")
		}
		return kerr.E(kerr.KindConflict, "invalid_transition",
			fmt.Sprintf("bulk operation is %s; cannot transition to %s", current, to))
	})
}

// completeIfDrained marks the operation completed ONLY when no pending or
// running item remains — including a peer worker's live running item (F-04).
// It reports whether the completion happened.
func (s *Service) completeIfDrained(ctx context.Context, txm database.TxManager, bulkID uuid.UUID) (bool, error) {
	var done bool
	err := txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		tag, err := db.Exec(ctx, `
			UPDATE bulk_operations SET status = 'completed', updated_at = now()
			 WHERE id = $1 AND status = 'running'
			   AND NOT EXISTS (
			       SELECT 1 FROM bulk_items
			        WHERE bulk_id = $1 AND status IN ('pending', 'running'))`, bulkID)
		if err != nil {
			return kerr.Wrapf(err, "bulk.completeIfDrained", "conditional complete")
		}
		done = tag.RowsAffected() == 1
		return nil
	})
	return done, err
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
		    count(*) FILTER (WHERE status = 'pending'),
		    count(*) FILTER (WHERE status = 'cancelled'),
		    count(*) FILTER (WHERE status = 'running')
		  FROM bulk_items WHERE bulk_id = $1`, bulkID).
		Scan(&p.Done, &p.Failed, &p.Pending, &p.Cancelled, &p.Running); err != nil {
		return Progress{}, kerr.Wrapf(err, "bulk.Progress", "count items")
	}
	return p, nil
}

// ExplainClaimPlan returns the EXPLAIN output for the leased-claim SQL. It is
// exported so tests can assert the plan uses FOR UPDATE SKIP LOCKED without
// needing to duplicate the SQL string.
func (s *Service) ExplainClaimPlan(ctx context.Context, db database.TenantDB, bulkID uuid.UUID, n int) (string, error) {
	l := lease.New(s.leaseTTL)
	rows, err := db.Query(ctx, "EXPLAIN (FORMAT TEXT) "+claimSQL, bulkID, l.Token, l.Generation, l.ExpiresAt, n)
	if err != nil {
		return "", kerr.Wrapf(err, "bulk.ExplainClaimPlan", "explain claim plan")
	}
	defer rows.Close()
	var parts []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return "", err
		}
		parts = append(parts, line)
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	return strings.Join(parts, "\n"), nil
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
