package migration

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestBackfillCheckpointUsesSharedLeasePrimitive proves the checkpoint table
// stores lease_token, lease_generation, and lease_expires_at populated from
// kernel/lease, and that an existing checkpoint's last_key is preserved while
// the lease epoch is bumped.
func TestBackfillCheckpointUsesSharedLeasePrimitive(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test needs Postgres")
	}
	db := testkit.NewDB(t)
	ctx := context.Background()

	admin, err := db.Admin.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire admin conn: %v", err)
	}
	defer admin.Release()

	if err := EnsureCheckpointTable(ctx, admin.Conn()); err != nil {
		t.Fatalf("ensure checkpoint table: %v", err)
	}

	if _, err := admin.Exec(ctx, "CREATE TABLE IF NOT EXISTS backfill_lease_source (id serial primary key, data text)"); err != nil {
		t.Fatalf("create source table: %v", err)
	}
	defer func() { _, _ = admin.Exec(ctx, "DROP TABLE IF EXISTS backfill_lease_source CASCADE") }()

	for i := range 3 {
		if _, err := admin.Exec(ctx, "INSERT INTO backfill_lease_source (data) VALUES ($1)", fmt.Sprintf("row-%d", i)); err != nil {
			t.Fatalf("insert row %d: %v", i, err)
		}
	}

	const jobID = "test-backfill-lease-primitive"
	// Pre-seed a checkpoint as if it were written by the pre-lease interim code.
	if _, err := admin.Exec(ctx,
		`INSERT INTO migration.backfill_checkpoint (job_id, last_key) VALUES ($1, 1)`,
		jobID); err != nil {
		t.Fatalf("seed old checkpoint: %v", err)
	}

	cfg := BackfillConfig{
		JobID:     jobID,
		Table:     "backfill_lease_source",
		KeyColumn: "id",
		BatchSize: 10,
	}

	bf := NewBackfill(cfg)
	res, err := bf.Run(ctx, admin.Conn(), func(tx pgx.Tx, lastKey int64, count int) error {
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !res.Resumed {
		t.Fatal("expected resumed=true for existing checkpoint")
	}

	var token *string
	var generation int64
	var lastKey int64
	if err := admin.QueryRow(ctx,
		`SELECT lease_token, lease_generation, last_key FROM migration.backfill_checkpoint WHERE job_id = $1`, jobID).
		Scan(&token, &generation, &lastKey); err != nil {
		t.Fatalf("read checkpoint: %v", err)
	}
	// F-03 fencing contract: a COMPLETED run releases its lease (token cleared)
	// so a successor claims immediately; the stored generation is retained so
	// the next epoch is strictly greater. Only a crashed runner leaves a live
	// token behind (its successor waits out the expiry).
	if token != nil {
		t.Fatalf("lease_token = %q after a completed run, want released (NULL)", *token)
	}
	if generation == 0 {
		t.Fatal("lease_generation not bumped after run")
	}
	if lastKey != 3 {
		t.Fatalf("last_key = %d, want 3", lastKey)
	}
}

func mustParseUUID(s string) uuid.UUID {
	v, err := uuid.Parse(s)
	if err != nil {
		panic(err)
	}
	return v
}

// TestBackfillInterruptedAndResumed proves the harness commits a checkpoint
// after each batch and resumes without reprocessing or skipping any row.
func TestBackfillInterruptedAndResumed(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test needs Postgres")
	}
	db := testkit.NewDB(t)
	ctx := context.Background()

	admin, err := db.Admin.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire admin conn: %v", err)
	}
	defer admin.Release()

	if err := EnsureCheckpointTable(ctx, admin.Conn()); err != nil {
		t.Fatalf("ensure checkpoint table: %v", err)
	}

	if _, err := admin.Exec(ctx, "CREATE TABLE IF NOT EXISTS backfill_source (id serial primary key, data text, processed bool not null default false)"); err != nil {
		t.Fatalf("create source table: %v", err)
	}
	defer func() { _, _ = admin.Exec(ctx, "DROP TABLE IF EXISTS backfill_source CASCADE") }()

	// Seed 25 rows.
	for i := range 25 {
		if _, err := admin.Exec(ctx, "INSERT INTO backfill_source (data) VALUES ($1)", "row"); err != nil {
			t.Fatalf("insert row %d: %v", i, err)
		}
	}

	const jobID = "test-backfill-interrupt-resume"
	cfg := BackfillConfig{
		JobID:     jobID,
		Table:     "backfill_source",
		KeyColumn: "id",
		BatchSize: 5,
	}

	var processed int
	stopAfter := 7
	process := func(tx pgx.Tx, lastKey int64, count int) error {
		processed += count
		if _, err := tx.Exec(ctx, "UPDATE backfill_source SET processed = true WHERE id <= $1", lastKey); err != nil {
			return err
		}
		if processed >= stopAfter {
			return ErrBackfillStopped
		}
		return nil
	}

	// First run: stop once we have crossed the threshold.
	bf := NewBackfill(cfg)
	res1, err := bf.Run(ctx, admin.Conn(), process)
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	if res1.Processed < int64(stopAfter) {
		t.Fatalf("first run processed %d rows, want >= %d", res1.Processed, stopAfter)
	}
	if res1.Batches != 2 {
		t.Fatalf("first run batches = %d, want 2", res1.Batches)
	}

	// Resume: reset the in-memory counter so we can reason about total work.
	processed = 0
	process = func(tx pgx.Tx, lastKey int64, count int) error {
		processed += count
		if _, err := tx.Exec(ctx, "UPDATE backfill_source SET processed = true WHERE id <= $1", lastKey); err != nil {
			return err
		}
		return nil
	}

	res2, err := NewBackfill(cfg).Run(ctx, admin.Conn(), process)
	if err != nil {
		t.Fatalf("resume run: %v", err)
	}
	if !res2.Resumed {
		t.Fatal("resume run did not detect prior checkpoint")
	}

	// Total rows touched across both runs must be exactly the table size (no
	// duplicates, no skips). Because the first run committed a full batch, the
	// second run resumes after that batch.
	total := res1.Processed + res2.Processed
	if total != 25 {
		t.Fatalf("total processed rows = %d, want 25 (res1=%d res2=%d)", total, res1.Processed, res2.Processed)
	}

	var unprocessed int64
	if err := admin.QueryRow(ctx, "SELECT count(*) FROM backfill_source WHERE NOT processed").Scan(&unprocessed); err != nil {
		t.Fatalf("count unprocessed: %v", err)
	}
	if unprocessed != 0 {
		t.Fatalf("unprocessed rows = %d, want 0", unprocessed)
	}

	var processedAgain int64
	if err := admin.QueryRow(ctx, "SELECT count(*) FROM backfill_source WHERE processed").Scan(&processedAgain); err != nil {
		t.Fatalf("count processed: %v", err)
	}
	if processedAgain != 25 {
		t.Fatalf("processed rows = %d, want 25", processedAgain)
	}
}

// TestBackfillTenantScoped proves tenant-scoped iteration isolates rows.
func TestBackfillTenantScoped(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test needs Postgres")
	}
	db := testkit.NewDB(t)
	ctx := context.Background()

	admin, err := db.Admin.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	defer admin.Release()

	if err := EnsureCheckpointTable(ctx, admin.Conn()); err != nil {
		t.Fatalf("ensure checkpoint table: %v", err)
	}

	if _, err := admin.Exec(ctx, "CREATE TABLE IF NOT EXISTS backfill_tenant (id serial, tenant_id uuid not null, data text, primary key (tenant_id, id))"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	defer func() { _, _ = admin.Exec(ctx, "DROP TABLE IF EXISTS backfill_tenant CASCADE") }()

	t1 := mustParseUUID("11111111-1111-1111-1111-111111111111")
	t2 := mustParseUUID("22222222-2222-2222-2222-222222222222")
	for i := 0; i < 5; i++ {
		if _, err := admin.Exec(ctx, "INSERT INTO backfill_tenant (tenant_id, data) VALUES ($1, $2)", t1, "t1"); err != nil {
			t.Fatalf("insert t1: %v", err)
		}
		if _, err := admin.Exec(ctx, "INSERT INTO backfill_tenant (tenant_id, data) VALUES ($1, $2)", t2, "t2"); err != nil {
			t.Fatalf("insert t2: %v", err)
		}
	}

	var seen int
	bf := NewBackfill(BackfillConfig{
		JobID:     "tenant-backfill",
		Table:     "backfill_tenant",
		KeyColumn: "id",
		TenantID:  &t1,
		BatchSize: 10,
	})
	_, err = bf.Run(ctx, admin.Conn(), func(_ pgx.Tx, _ int64, count int) error {
		seen += count
		return nil
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if seen != 5 {
		t.Fatalf("tenant-scoped backfill saw %d rows, want 5", seen)
	}
}

var _ = errors.New // keep imports tidy when tests are skipped
