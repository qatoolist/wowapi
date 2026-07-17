package migration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestExecDDLLockTimeoutAbortAndRetry proves that a DDL statement exceeding
// the lock budget against a concurrently-locked table aborts cleanly (no
// partial DDL) and retries within a bounded ceiling once the lock is released.
func TestExecDDLLockTimeoutAbortAndRetry(t *testing.T) {
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

	if _, err := admin.Exec(ctx, "CREATE TABLE IF NOT EXISTS migration_locktest (id int primary key)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	defer func() {
		_, _ = admin.Exec(ctx, "DROP TABLE IF EXISTS migration_locktest")
	}()

	// Holder connection keeps an ACCESS EXCLUSIVE lock for a controlled window.
	holder, err := db.Admin.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire holder conn: %v", err)
	}
	defer holder.Release()

	if _, err := holder.Exec(ctx, "BEGIN"); err != nil {
		t.Fatalf("holder begin: %v", err)
	}
	if _, err := holder.Exec(ctx, "LOCK TABLE migration_locktest IN ACCESS EXCLUSIVE MODE"); err != nil {
		t.Fatalf("holder lock: %v", err)
	}

	done := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		defer close(done)
		worker, err := db.Admin.Acquire(ctx)
		if err != nil {
			errCh <- err
			return
		}
		defer worker.Release()
		// Tight budget so we can observe an abort even on a fast machine.
		err = ExecDDL(ctx, worker.Conn(), "ALTER TABLE migration_locktest ADD COLUMN IF NOT EXISTS payload text", 100*time.Millisecond, 3)
		errCh <- err
	}()

	// Release the lock after the worker has had time to hit the budget once.
	time.Sleep(250 * time.Millisecond)
	if _, err := holder.Exec(ctx, "COMMIT"); err != nil {
		t.Fatalf("holder commit: %v", err)
	}

	<-done
	if err := <-errCh; err != nil {
		t.Fatalf("worker DDL: %v", err)
	}

	// Verify the column exists (retry eventually succeeded, no partial state).
	var col string
	if err := admin.QueryRow(ctx, "SELECT column_name FROM information_schema.columns WHERE table_name='migration_locktest' AND column_name='payload'").Scan(&col); err != nil {
		t.Fatalf("column not found after retry: %v", err)
	}
	if col != "payload" {
		t.Fatalf("unexpected column %q", col)
	}
}

// TestExecDDLLockTimeoutExhausted proves that if the lock is never released,
// the bounded retry ceiling is exhausted and ErrLockTimeout is returned.
func TestExecDDLLockTimeoutExhausted(t *testing.T) {
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

	if _, err := admin.Exec(ctx, "CREATE TABLE IF NOT EXISTS migration_locktest_exhausted (id int primary key)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	defer func() {
		_, _ = admin.Exec(ctx, "DROP TABLE IF EXISTS migration_locktest_exhausted")
	}()

	holder, err := db.Admin.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire holder: %v", err)
	}
	defer holder.Release()

	if _, err := holder.Exec(ctx, "BEGIN"); err != nil {
		t.Fatalf("holder begin: %v", err)
	}
	if _, err := holder.Exec(ctx, "LOCK TABLE migration_locktest_exhausted IN ACCESS EXCLUSIVE MODE"); err != nil {
		t.Fatalf("holder lock: %v", err)
	}
	defer func() { _, _ = holder.Exec(ctx, "ROLLBACK") }()

	worker, err := db.Admin.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire worker: %v", err)
	}
	defer worker.Release()

	start := time.Now()
	err = ExecDDL(ctx, worker.Conn(), "ALTER TABLE migration_locktest_exhausted ADD COLUMN IF NOT EXISTS x int", 50*time.Millisecond, 2)
	elapsed := time.Since(start)
	if !errors.Is(err, ErrLockTimeout) {
		t.Fatalf("expected ErrLockTimeout, got %v", err)
	}
	// 3 attempts: 50ms budget each + small backoff (50+100ms) => at least 150ms.
	if elapsed < 150*time.Millisecond {
		t.Fatalf("retries finished too quickly (%v)", elapsed)
	}
}

var _ = pgx.ErrNoRows // silence unused import if tests are skipped
