package migration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/qatoolist/wowapi/testkit"
)

// F-03 regressions (adversarial-framework-review-2026-07-17): checkpoint
// identity must be (job_id, tenant); only the current unexpired lease owner may
// advance a checkpoint; lease generation must increase across ownership epochs;
// a checkpoint never moves backward.

func backfillFencingSetup(t *testing.T) (*pgx.Conn, func()) {
	t.Helper()
	if testing.Short() {
		t.Skip("integration test needs Postgres")
	}
	db := testkit.NewDB(t)
	ctx := context.Background()
	admin, err := db.Admin.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire admin conn: %v", err)
	}
	if err := EnsureCheckpointTable(ctx, admin.Conn()); err != nil {
		t.Fatalf("ensure checkpoint table: %v", err)
	}
	if _, err := admin.Exec(ctx, `CREATE TABLE IF NOT EXISTS fencing_source
		(id bigserial primary key, tenant_id uuid, data text)`); err != nil {
		t.Fatalf("create source: %v", err)
	}
	return admin.Conn(), func() { admin.Release() }
}

func seedFencingRows(t *testing.T, conn *pgx.Conn, tenant uuid.UUID, n int) {
	t.Helper()
	for i := range n {
		if _, err := conn.Exec(context.Background(),
			`INSERT INTO fencing_source (tenant_id, data) VALUES ($1, $2)`,
			tenant, fmt.Sprintf("r%d", i)); err != nil {
			t.Fatal(err)
		}
	}
}

// (1) Two tenants sharing a stable JobID must have INDEPENDENT checkpoints:
// each resumes from its own last_key and neither corrupts the other's row.
func TestBackfillTwoTenantsIndependentCheckpoints(t *testing.T) {
	conn, release := backfillFencingSetup(t)
	defer release()
	ctx := context.Background()

	tenantA, tenantB := uuid.New(), uuid.New()
	seedFencingRows(t, conn, tenantA, 5)
	seedFencingRows(t, conn, tenantB, 3)

	run := func(tenant uuid.UUID) *BackfillResult {
		b := NewBackfill(BackfillConfig{
			JobID: "shared-job", Table: "fencing_source", KeyColumn: "id",
			TenantID: &tenant, BatchSize: 2,
		})
		res, err := b.Run(ctx, conn, func(tx pgx.Tx, lastKey int64, count int) error { return nil })
		if err != nil {
			t.Fatalf("backfill tenant %s: %v", tenant, err)
		}
		return res
	}

	resA := run(tenantA)
	if resA.Processed != 5 {
		t.Fatalf("tenant A processed %d rows, want 5", resA.Processed)
	}
	resB := run(tenantB)
	if resB.Processed != 3 {
		t.Fatalf("tenant B processed %d rows, want 3 — shared-job checkpoint identity collided across tenants", resB.Processed)
	}

	var rows int
	if err := conn.QueryRow(ctx,
		`SELECT count(*) FROM migration.backfill_checkpoint WHERE job_id = 'shared-job'`).Scan(&rows); err != nil {
		t.Fatal(err)
	}
	if rows != 2 {
		t.Fatalf("checkpoint rows for shared-job = %d, want 2 (one per tenant)", rows)
	}

	// Re-running tenant A must resume (no rows left), never restart at 0 because
	// tenant B's run touched a shared row.
	resA2 := run(tenantA)
	if resA2.Processed != 0 || !resA2.Resumed {
		t.Fatalf("tenant A re-run processed %d (resumed=%v), want 0 processed from its own checkpoint", resA2.Processed, resA2.Resumed)
	}
}

// (2) A live, unexpired lease excludes a second claimant.
func TestBackfillClaimExcludesLiveOwner(t *testing.T) {
	conn, release := backfillFencingSetup(t)
	defer release()
	ctx := context.Background()

	b1 := NewBackfill(BackfillConfig{JobID: "exclusive-job", Table: "fencing_source", KeyColumn: "id"})
	if _, _, err := b1.claimCheckpoint(ctx, conn); err != nil {
		t.Fatalf("first claim: %v", err)
	}
	b2 := NewBackfill(BackfillConfig{JobID: "exclusive-job", Table: "fencing_source", KeyColumn: "id"})
	if _, _, err := b2.claimCheckpoint(ctx, conn); err == nil {
		t.Fatal("second claim succeeded while the first lease is live — claim must only take absent or expired leases")
	}
}

// (3) An expired lease is reclaimable with an INCREASED generation, and the
// stale owner's checkpoint writes are rejected without moving last_key.
func TestBackfillStaleOwnerWriteRejected(t *testing.T) {
	conn, release := backfillFencingSetup(t)
	defer release()
	ctx := context.Background()

	stale := NewBackfill(BackfillConfig{JobID: "reclaim-job", Table: "fencing_source", KeyColumn: "id"})
	if _, _, err := stale.claimCheckpoint(ctx, conn); err != nil {
		t.Fatalf("initial claim: %v", err)
	}
	staleGen := stale.lease.Generation
	if err := stale.writeCheckpoint(ctx, conn, 10); err != nil {
		t.Fatalf("owner write: %v", err)
	}

	// Force expiry, then a new runner reclaims.
	if _, err := conn.Exec(ctx,
		`UPDATE migration.backfill_checkpoint SET lease_expires_at = now() - interval '1 minute'
		  WHERE job_id = 'reclaim-job'`); err != nil {
		t.Fatal(err)
	}
	fresh := NewBackfill(BackfillConfig{JobID: "reclaim-job", Table: "fencing_source", KeyColumn: "id"})
	lastKey, _, err := fresh.claimCheckpoint(ctx, conn)
	if err != nil {
		t.Fatalf("reclaim of expired lease: %v", err)
	}
	if lastKey != 10 {
		t.Fatalf("reclaim read last_key %d, want 10", lastKey)
	}
	if fresh.lease.Generation <= staleGen {
		t.Fatalf("reclaimed generation %d not greater than stale %d — stale writers cannot be fenced", fresh.lease.Generation, staleGen)
	}

	// The stale owner must now be rejected and must not move the checkpoint.
	if err := stale.writeCheckpoint(ctx, conn, 99); err == nil {
		t.Fatal("stale owner's checkpoint write succeeded after reclaim — lease is metadata, not a fence")
	}
	var got int64
	if err := conn.QueryRow(ctx,
		`SELECT last_key FROM migration.backfill_checkpoint WHERE job_id = 'reclaim-job'`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 10 {
		t.Fatalf("stale write moved last_key to %d, want 10 preserved", got)
	}
}

// (4) A checkpoint never moves backward, even for the current owner.
func TestBackfillCheckpointMonotonic(t *testing.T) {
	conn, release := backfillFencingSetup(t)
	defer release()
	ctx := context.Background()

	b := NewBackfill(BackfillConfig{JobID: "monotonic-job", Table: "fencing_source", KeyColumn: "id"})
	if _, _, err := b.claimCheckpoint(ctx, conn); err != nil {
		t.Fatalf("claim: %v", err)
	}
	if err := b.writeCheckpoint(ctx, conn, 20); err != nil {
		t.Fatalf("advance to 20: %v", err)
	}
	if err := b.writeCheckpoint(ctx, conn, 5); err == nil {
		t.Fatal("checkpoint regressed from 20 to 5 without error")
	}
	var got int64
	if err := conn.QueryRow(ctx,
		`SELECT last_key FROM migration.backfill_checkpoint WHERE job_id = 'monotonic-job'`).Scan(&got); err != nil {
		t.Fatal(err)
	}
	if got != 20 {
		t.Fatalf("last_key = %d after regress attempt, want 20", got)
	}
	_ = time.Now // keep time imported if assertions above change
}
