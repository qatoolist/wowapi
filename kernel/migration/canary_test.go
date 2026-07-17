package migration

import (
	"context"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/testkit"
)

// TestCanaryNAndNMinusOne proves both explicitly-required canary legs:
//   - N-1 code runs correctly against the N-expanded schema before and after backfill.
//   - N code runs correctly before and after backfill.
//
// Soak duration and error threshold are configurable (not hardcoded).
func TestCanaryNAndNMinusOne(t *testing.T) {
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

	if _, err := admin.Exec(ctx, "CREATE TABLE IF NOT EXISTS canary_items (id serial primary key, name text not null)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	defer func() { _, _ = admin.Exec(ctx, "DROP TABLE IF EXISTS canary_items CASCADE") }()

	// Seed pre-existing rows (N-1 data shape).
	for range 5 {
		if _, err := admin.Exec(ctx, "INSERT INTO canary_items (name) VALUES ($1)", "legacy"); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	// Expand phase: additive schema change.
	var ex ExpandPhase
	if err := ExecExpandDDL(ctx, admin.Conn(), ex.AddColumnNullableDefault("canary_items", "status", "text", "'pending'")); err != nil {
		t.Fatalf("expand column: %v", err)
	}

	oldReader := func() CanaryLeg {
		return CanaryLeg{
			Name:      "old-reader-select",
			SchemaAge: "",
			Version:   "N-1",
			Run: func(ctx context.Context) error {
				var name string
				return admin.QueryRow(ctx, "SELECT name FROM canary_items ORDER BY id LIMIT 1").Scan(&name)
			},
		}
	}

	newReader := func(schemaAge string) CanaryLeg {
		return CanaryLeg{
			Name:      "new-reader-select",
			SchemaAge: schemaAge,
			Version:   "N",
			Run: func(ctx context.Context) error {
				var status string
				return admin.QueryRow(ctx, "SELECT status FROM canary_items ORDER BY id LIMIT 1").Scan(&status)
			},
		}
	}

	// Before backfill: both old and new readers accept the expanded schema.
	cfg := SoakConfig{
		SoakDuration:   100 * time.Millisecond,
		ErrorThreshold: 0,
		MinSampleCount: 1,
		RateLimit:      10 * time.Millisecond,
	}

	res1, err := RunCanary(ctx, cfg, []CanaryLeg{
		oldReader(),
		newReader("before_backfill"),
	})
	if err != nil {
		t.Fatalf("canary before backfill: %v", err)
	}
	if !res1.Passed {
		t.Fatalf("canary before backfill failed: %s", res1.Summary())
	}

	// Backfill: update existing rows to the new representation.
	if _, err := admin.Exec(ctx, "UPDATE canary_items SET status = 'done'"); err != nil {
		t.Fatalf("backfill: %v", err)
	}

	// After backfill: both old and new readers still work; new reader now sees 'done'.
	res2, err := RunCanary(ctx, cfg, []CanaryLeg{
		oldReader(),
		newReader("after_backfill"),
	})
	if err != nil {
		t.Fatalf("canary after backfill: %v", err)
	}
	if !res2.Passed {
		t.Fatalf("canary after backfill failed: %s", res2.Summary())
	}

	var status string
	if err := admin.QueryRow(ctx, "SELECT status FROM canary_items ORDER BY id LIMIT 1").Scan(&status); err != nil {
		t.Fatalf("final status query: %v", err)
	}
	if status != "done" {
		t.Fatalf("post-backfill status = %q, want done", status)
	}
}

// TestPartialFleetRollout proves a canary where only part of the fleet is on
// N while N-1 remains active. This maps to the directive's "partial fleet
// rollout" drill.
func TestPartialFleetRollout(t *testing.T) {
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

	if _, err := admin.Exec(ctx, "CREATE TABLE IF NOT EXISTS partial_fleet_items (id serial primary key, name text)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	defer func() { _, _ = admin.Exec(ctx, "DROP TABLE IF EXISTS partial_fleet_items CASCADE") }()

	if _, err := admin.Exec(ctx, "INSERT INTO partial_fleet_items (name) VALUES ($1)", "item"); err != nil {
		t.Fatalf("seed: %v", err)
	}

	var ex ExpandPhase
	if err := ExecExpandDDL(ctx, admin.Conn(), ex.AddColumnNullableDefault("partial_fleet_items", "extra", "int", "0")); err != nil {
		t.Fatalf("expand: %v", err)
	}

	legs := []CanaryLeg{
		{
			Name:    "n-minus-one-fleet",
			Version: "N-1",
			Run: func(ctx context.Context) error {
				var name string
				return admin.QueryRow(ctx, "SELECT name FROM partial_fleet_items LIMIT 1").Scan(&name)
			},
		},
		{
			Name:    "n-fleet-partial",
			Version: "N",
			Run: func(ctx context.Context) error {
				var extra int
				return admin.QueryRow(ctx, "SELECT extra FROM partial_fleet_items LIMIT 1").Scan(&extra)
			},
		},
	}

	cfg := SoakConfig{
		SoakDuration:   100 * time.Millisecond,
		ErrorThreshold: 0,
		MinSampleCount: 1,
		RateLimit:      10 * time.Millisecond,
	}
	res, err := RunCanary(ctx, cfg, legs)
	if err != nil {
		t.Fatalf("partial fleet canary: %v", err)
	}
	if !res.Passed {
		t.Fatalf("partial fleet canary failed: %s", res.Summary())
	}
}
