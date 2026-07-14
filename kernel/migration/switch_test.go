package migration

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/testkit"
)

// TestSwitchRollbackAfterSwitch proves application rollback after switch with
// no destructive Down: the compatibility flag moves back to the previous
// version while the expanded schema remains in place.
func TestSwitchRollbackAfterSwitch(t *testing.T) {
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

	if err := EnsureCompatFlagTable(ctx, admin.Conn()); err != nil {
		t.Fatalf("ensure compat flag table: %v", err)
	}

	if _, err := admin.Exec(ctx, "CREATE TABLE IF NOT EXISTS switch_items (id serial primary key)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	defer func() { _, _ = admin.Exec(ctx, "DROP TABLE IF EXISTS switch_items CASCADE") }()

	// Expand phase: add a column. This is the additive schema that must survive
	// the application rollback.
	var ex ExpandPhase
	if err := ExecExpandDDL(ctx, admin.Conn(), ex.AddColumnNullableDefault("switch_items", "new_col", "int", "0")); err != nil {
		t.Fatalf("expand: %v", err)
	}

	const flagKey = "canary-switch-test"
	// Switch to N.
	if err := SetCompatibility(ctx, admin.Conn(), flagKey, "N"); err != nil {
		t.Fatalf("set N: %v", err)
	}
	flag, err := GetCompatibility(ctx, admin.Conn(), flagKey)
	if err != nil {
		t.Fatalf("get flag: %v", err)
	}
	if flag.Version != "N" {
		t.Fatalf("flag version = %q, want N", flag.Version)
	}

	// Application rollback: move the flag back to N-1. No Down is executed.
	if err := RollbackAfterSwitch(ctx, admin.Conn(), flagKey, "N-1"); err != nil {
		t.Fatalf("rollback after switch: %v", err)
	}
	flag, err = GetCompatibility(ctx, admin.Conn(), flagKey)
	if err != nil {
		t.Fatalf("get flag after rollback: %v", err)
	}
	if flag.Version != "N-1" {
		t.Fatalf("flag version after rollback = %q, want N-1", flag.Version)
	}

	// The expanded column still exists, proving rollback was by application
	// version, not destructive schema reversal.
	var col string
	if err := admin.QueryRow(ctx, "SELECT column_name FROM information_schema.columns WHERE table_name='switch_items' AND column_name='new_col'").Scan(&col); err != nil {
		t.Fatalf("expanded column missing after rollback: %v", err)
	}
	if col != "new_col" {
		t.Fatalf("unexpected column %q", col)
	}
}
