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
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS migration.backfill_checkpoint (
			job_id text PRIMARY KEY,
			tenant_id uuid,
			last_key bigint NOT NULL DEFAULT 0,
			updated_at timestamptz NOT NULL DEFAULT now(),
			lease_token text,
			lease_generation bigint NOT NULL DEFAULT 0,
			lease_expires_at timestamptz
		)
	`)
	return err
}

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

	lastKey, resumed, err := b.readCheckpoint(ctx, conn)
	if err != nil {
		return nil, fmt.Errorf("read checkpoint: %w", err)
	}
	// Claim the checkpoint under a fresh lease epoch for this run.
	if err := b.claimCheckpoint(ctx, conn); err != nil {
		return nil, fmt.Errorf("claim checkpoint: %w", err)
	}

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

func (b *Backfill) tenantArg() any {
	if b.cfg.TenantID == nil {
		return nil
	}
	return *b.cfg.TenantID
}

func (b *Backfill) readCheckpoint(ctx context.Context, conn *pgx.Conn) (lastKey int64, resumed bool, err error) {
	var args []any
	query := "SELECT last_key FROM migration.backfill_checkpoint WHERE job_id = $1"
	args = append(args, b.cfg.JobID)
	if b.cfg.TenantID != nil {
		query += " AND tenant_id = $2"
		args = append(args, b.tenantArg())
	}
	err = conn.QueryRow(ctx, query, args...).Scan(&lastKey)
	if errors.Is(err, pgx.ErrNoRows) {
		if err := b.writeCheckpoint(ctx, conn, 0); err != nil {
			return 0, false, err
		}
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return lastKey, true, nil
}

// claimCheckpoint bumps the checkpoint into a new lease epoch for this run.
func (b *Backfill) claimCheckpoint(ctx context.Context, conn *pgx.Conn) error {
	b.lease = lease.New(5 * time.Minute)
	_, err := conn.Exec(ctx, `
		INSERT INTO migration.backfill_checkpoint (job_id, tenant_id, last_key, updated_at, lease_token, lease_generation, lease_expires_at)
		VALUES ($1, $2, 0, now(), $3, $4, $5)
		ON CONFLICT (job_id) DO UPDATE
		SET lease_token = EXCLUDED.lease_token,
		    lease_generation = EXCLUDED.lease_generation,
		    lease_expires_at = EXCLUDED.lease_expires_at
	`, b.cfg.JobID, b.tenantArg(), b.lease.Token, b.lease.Generation, b.lease.ExpiresAt)
	return err
}

func (b *Backfill) writeCheckpoint(ctx context.Context, conn *pgx.Conn, lastKey int64) error {
	b.lease = b.lease.Renew(5 * time.Minute)
	_, err := conn.Exec(ctx, `
		INSERT INTO migration.backfill_checkpoint (job_id, tenant_id, last_key, updated_at, lease_token, lease_generation, lease_expires_at)
		VALUES ($1, $2, $3, now(), $4, $5, $6)
		ON CONFLICT (job_id) DO UPDATE
		SET last_key = EXCLUDED.last_key,
		    updated_at = EXCLUDED.updated_at,
		    lease_token = EXCLUDED.lease_token,
		    lease_generation = EXCLUDED.lease_generation,
		    lease_expires_at = EXCLUDED.lease_expires_at
	`, b.cfg.JobID, b.tenantArg(), lastKey, b.lease.Token, b.lease.Generation, b.lease.ExpiresAt)
	return err
}

func (b *Backfill) writeCheckpointTx(ctx context.Context, tx pgx.Tx, lastKey int64) error {
	b.lease = b.lease.Renew(5 * time.Minute)
	_, err := tx.Exec(ctx, `
		INSERT INTO migration.backfill_checkpoint (job_id, tenant_id, last_key, updated_at, lease_token, lease_generation, lease_expires_at)
		VALUES ($1, $2, $3, now(), $4, $5, $6)
		ON CONFLICT (job_id) DO UPDATE
		SET last_key = EXCLUDED.last_key,
		    updated_at = EXCLUDED.updated_at,
		    lease_token = EXCLUDED.lease_token,
		    lease_generation = EXCLUDED.lease_generation,
		    lease_expires_at = EXCLUDED.lease_expires_at
	`, b.cfg.JobID, b.tenantArg(), lastKey, b.lease.Token, b.lease.Generation, b.lease.ExpiresAt)
	return err
}
