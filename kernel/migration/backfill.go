package migration

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/qatoolist/wowapi/kernel/lease"
)

// ErrBackfillStopped is returned when the backfill halts because the process
// callback returned nil after signalling it has processed enough rows. It is not
// an error condition for the resume test: the checkpoint is still committed.
var ErrBackfillStopped = errors.New("migration: backfill stopped by callback")

// BackfillConfig controls a resumable, tenant-scoped, keyset-paginated backfill.
type BackfillConfig struct {
	JobID        string
	Table        string
	KeyColumn    string // defaults to "id"
	TenantColumn string // defaults to "tenant_id"; ignored when TenantID is nil
	TenantID     *uuid.UUID
	BatchSize    int
	RateLimit    time.Duration // sleep between batches (zero = none)
	Window       time.Duration // max total runtime (zero = no limit)
}

func (c BackfillConfig) keyColumn() string {
	if c.KeyColumn == "" {
		return "id"
	}
	return c.KeyColumn
}

func (c BackfillConfig) tenantColumn() string {
	if c.TenantColumn == "" {
		return "tenant_id"
	}
	return c.TenantColumn
}

// BackfillResult reports what a Run completed.
type BackfillResult struct {
	Processed int64
	Batches   int64
	Resumed   bool
}

// ProcessBatch is called once per batch inside the batch transaction. It
// receives the highest key in the batch and the number of rows in the batch.
// Returning ErrBackfillStopped causes the harness to commit this batch and
// return without processing further batches.
type ProcessBatch func(tx pgx.Tx, lastKey int64, count int) error

// EnsureCheckpointTable creates the migration.backfill_checkpoint table that
// the interim checkpoint-lease primitive uses. Callers with schema-owner
// privileges (tests or the migrate owner) create it once per database.
func EnsureCheckpointTable(ctx context.Context, conn *pgx.Conn) error {
	if _, err := conn.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS migration"); err != nil {
		return err
	}
	// Identity is (job_id, tenant_id): a tenant-scoped job checkpoints per
	// tenant; global jobs use the all-zeros sentinel (adversarial review
	// 2026-07-17, F-03 — a job_id-only key made a second tenant's checkpoint
	// collide with the first's). Mirrors migration 00049.
	if _, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migration.backfill_checkpoint (
			job_id text NOT NULL,
			tenant_id uuid NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
			last_key bigint NOT NULL DEFAULT 0,
			updated_at timestamptz NOT NULL DEFAULT now(),
			lease_token text,
			lease_generation bigint NOT NULL DEFAULT 0,
			lease_expires_at timestamptz,
			PRIMARY KEY (job_id, tenant_id)
		)
	`); err != nil {
		return err
	}
	// Upgrade a pre-F-03 table in place (idempotent on the new shape).
	_, err := conn.Exec(ctx, `
		DO $$
		BEGIN
			IF EXISTS (
				SELECT 1 FROM pg_constraint c
				 JOIN pg_class r ON r.oid = c.conrelid
				 JOIN pg_namespace n ON n.oid = r.relnamespace
				WHERE n.nspname = 'migration' AND r.relname = 'backfill_checkpoint'
				  AND c.contype = 'p' AND array_length(c.conkey, 1) = 1
			) THEN
				UPDATE migration.backfill_checkpoint
				   SET tenant_id = '00000000-0000-0000-0000-000000000000'
				 WHERE tenant_id IS NULL;
				ALTER TABLE migration.backfill_checkpoint
					ALTER COLUMN tenant_id SET DEFAULT '00000000-0000-0000-0000-000000000000',
					ALTER COLUMN tenant_id SET NOT NULL;
				ALTER TABLE migration.backfill_checkpoint
					DROP CONSTRAINT backfill_checkpoint_pkey;
				ALTER TABLE migration.backfill_checkpoint
					ADD PRIMARY KEY (job_id, tenant_id);
			END IF;
		END $$;
	`)
	return err
}

// Checkpoint fencing errors (F-03): claims and writes are compare-and-swap
// operations, never unconditional labels.
var (
	// ErrCheckpointLeaseHeld means another runner holds a live, unexpired lease
	// on this (job, tenant) checkpoint.
	ErrCheckpointLeaseHeld = errors.New("migration: backfill checkpoint lease held by another runner")
	// ErrStaleCheckpointWrite means this runner's lease epoch is no longer
	// current (reclaimed or expired), or the write would move last_key backward.
	ErrStaleCheckpointWrite = errors.New("migration: stale or non-monotonic backfill checkpoint write rejected")
)

// globalTenantSentinel is the tenant_id for jobs without a tenant scope.
var globalTenantSentinel = uuid.Nil

// NewBackfill builds a backfill runner for the given configuration.
func NewBackfill(cfg BackfillConfig) *Backfill {
	return &Backfill{cfg: cfg}
}

// Backfill is the DATA-09 T4 harness. It uses the shared lease primitive from
// kernel/lease for checkpoint safety: each committed batch bumps the checkpoint
// lease generation so concurrent or resumed runs can detect a stale epoch.
type Backfill struct {
	cfg   BackfillConfig
	lease lease.Lease
}

// Run executes the backfill. It reads the checkpoint, processes batches, and
// updates the checkpoint after each committed batch.
func (b *Backfill) Run(ctx context.Context, conn *pgx.Conn, process ProcessBatch) (*BackfillResult, error) {
	if b.cfg.BatchSize <= 0 {
		b.cfg.BatchSize = 1000
	}

	// Claim atomically creates-or-takes-over the (job, tenant) checkpoint and
	// returns the authoritative resume key — read and claim are one CAS so a
	// concurrent runner can never read a key it does not own (F-03).
	lastKey, resumed, err := b.claimCheckpoint(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("claim checkpoint: %w", err)
	}
	// Release the lease when this run ends so a successor (resume, next window)
	// claims immediately; a CRASHED runner skips this and its successor waits
	// for expiry — that asymmetry is the fence.
	defer func() { b.releaseCheckpoint(context.WithoutCancel(ctx), conn) }()

	res := &BackfillResult{Resumed: resumed}
	start := time.Now()

	for {
		if b.cfg.Window > 0 && time.Since(start) > b.cfg.Window {
			return res, nil
		}

		batchLast, batchCount, stop, err := b.runBatch(ctx, conn, process, lastKey)
		if err != nil {
			return nil, err
		}
		if batchCount == 0 {
			return res, nil
		}

		res.Processed += int64(batchCount)
		res.Batches++
		lastKey = batchLast

		if stop {
			return res, nil
		}

		if b.cfg.RateLimit > 0 {
			select {
			case <-time.After(b.cfg.RateLimit):
			case <-ctx.Done():
				return res, ctx.Err()
			}
		}
	}
}

func (b *Backfill) runBatch(ctx context.Context, conn *pgx.Conn, process ProcessBatch, lastKey int64) (batchLast int64, batchCount int, stop bool, err error) {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return 0, 0, false, fmt.Errorf("begin batch: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	query, args := b.selectArgs(lastKey)
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return 0, 0, false, err
	}
	for rows.Next() {
		var k int64
		if err := rows.Scan(&k); err != nil {
			rows.Close()
			return 0, 0, false, err
		}
		batchLast = k
		batchCount++
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return 0, 0, false, err
	}
	if batchCount == 0 {
		return 0, 0, false, nil
	}

	if err := process(tx, batchLast, batchCount); err != nil {
		if errors.Is(err, ErrBackfillStopped) {
			stop = true
		} else {
			return 0, 0, false, err
		}
	}
	if err := b.writeCheckpointTx(ctx, tx, batchLast); err != nil {
		return 0, 0, false, fmt.Errorf("write checkpoint: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, 0, false, fmt.Errorf("commit batch: %w", err)
	}
	return batchLast, batchCount, stop, nil
}

func (b *Backfill) selectArgs(lastKey int64) (string, []any) {
	kc := b.cfg.keyColumn()
	if b.cfg.TenantID != nil {
		tc := b.cfg.tenantColumn()
		return fmt.Sprintf(
			"SELECT %s FROM %s WHERE %s = $1 AND %s > $2 ORDER BY %s LIMIT %d",
			quoteIdent(kc), quoteIdent(b.cfg.Table), quoteIdent(tc), quoteIdent(kc), quoteIdent(kc), b.cfg.BatchSize,
		), []any{*b.cfg.TenantID, lastKey}
	}
	return fmt.Sprintf(
		"SELECT %s FROM %s WHERE %s > $1 ORDER BY %s LIMIT %d",
		quoteIdent(kc), quoteIdent(b.cfg.Table), quoteIdent(kc), quoteIdent(kc), b.cfg.BatchSize,
	), []any{lastKey}
}

func (b *Backfill) tenantArg() uuid.UUID {
	if b.cfg.TenantID == nil {
		return globalTenantSentinel
	}
	return *b.cfg.TenantID
}

// claimCheckpoint atomically claims the (job, tenant) checkpoint for this run:
// it creates an absent row, or takes over an ABSENT-OR-EXPIRED lease with a
// strictly increased stored generation. A live lease held by another runner is
// never displaced (F-03: the lease is a fence, not metadata). It returns the
// authoritative last_key and whether the checkpoint pre-existed.
func (b *Backfill) claimCheckpoint(ctx context.Context, conn *pgx.Conn) (lastKey int64, resumed bool, err error) {
	fresh := lease.New(checkpointLeaseTTL)
	var generation int64
	var inserted bool
	err = conn.QueryRow(ctx, `
		INSERT INTO migration.backfill_checkpoint (job_id, tenant_id, last_key, updated_at, lease_token, lease_generation, lease_expires_at)
		VALUES ($1, $2, 0, now(), $3, 1, $4)
		ON CONFLICT (job_id, tenant_id) DO UPDATE
		SET lease_token = EXCLUDED.lease_token,
		    lease_generation = migration.backfill_checkpoint.lease_generation + 1,
		    lease_expires_at = EXCLUDED.lease_expires_at,
		    updated_at = now()
		WHERE migration.backfill_checkpoint.lease_token IS NULL
		   OR migration.backfill_checkpoint.lease_expires_at <= now()
		RETURNING last_key, lease_generation, (xmax = 0) AS inserted
	`, b.cfg.JobID, b.tenantArg(), fresh.Token, fresh.ExpiresAt).Scan(&lastKey, &generation, &inserted)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, fmt.Errorf("%w: job %q tenant %s", ErrCheckpointLeaseHeld, b.cfg.JobID, b.tenantArg())
	}
	if err != nil {
		return 0, false, err
	}
	b.lease = lease.Lease{Token: fresh.Token, Generation: generation, ExpiresAt: fresh.ExpiresAt}
	return lastKey, !inserted, nil
}

// checkpointLeaseTTL bounds how long a crashed runner blocks a successor.
const checkpointLeaseTTL = 5 * time.Minute

// releaseCheckpoint clears this runner's lease if it is still the current
// epoch, letting the next runner claim without waiting out the TTL. Losing the
// race (already reclaimed) is fine — the fence already excludes this runner.
func (b *Backfill) releaseCheckpoint(ctx context.Context, conn *pgx.Conn) {
	_, _ = conn.Exec(ctx, `
		UPDATE migration.backfill_checkpoint
		   SET lease_token = NULL, lease_expires_at = NULL, updated_at = now()
		 WHERE job_id = $1 AND tenant_id = $2
		   AND lease_token = $3 AND lease_generation = $4
	`, b.cfg.JobID, b.tenantArg(), b.lease.Token, b.lease.Generation)
}

// checkpointWriteSQL fences every checkpoint advance on the complete claimed
// identity — (job, tenant), lease token, generation, unexpired ownership — and
// on monotonic progress. Exactly one row must be affected (F-03).
const checkpointWriteSQL = `
	UPDATE migration.backfill_checkpoint
	   SET last_key = $3,
	       updated_at = now(),
	       lease_expires_at = $6
	 WHERE job_id = $1 AND tenant_id = $2
	   AND lease_token = $4 AND lease_generation = $5
	   AND lease_expires_at > now()
	   AND last_key <= $3`

func (b *Backfill) writeCheckpoint(ctx context.Context, conn *pgx.Conn, lastKey int64) error {
	renewed := b.lease.Renew(checkpointLeaseTTL)
	tag, err := conn.Exec(ctx, checkpointWriteSQL,
		b.cfg.JobID, b.tenantArg(), lastKey, b.lease.Token, b.lease.Generation, renewed.ExpiresAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("%w: job %q tenant %s last_key %d", ErrStaleCheckpointWrite, b.cfg.JobID, b.tenantArg(), lastKey)
	}
	b.lease = renewed
	return nil
}

func (b *Backfill) writeCheckpointTx(ctx context.Context, tx pgx.Tx, lastKey int64) error {
	renewed := b.lease.Renew(checkpointLeaseTTL)
	tag, err := tx.Exec(ctx, checkpointWriteSQL,
		b.cfg.JobID, b.tenantArg(), lastKey, b.lease.Token, b.lease.Generation, renewed.ExpiresAt)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("%w: job %q tenant %s last_key %d", ErrStaleCheckpointWrite, b.cfg.JobID, b.tenantArg(), lastKey)
	}
	b.lease = renewed
	return nil
}
