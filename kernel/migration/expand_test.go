package migration

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/testkit"
)

// TestExpandPhaseOldReaderCompatibility proves that an expand-phase migration
// (nullable/default-safe column, NOT VALID constraint, and CREATE INDEX
// CONCURRENTLY) leaves the schema readable for both old-version and new-version
// application code.
func TestExpandPhaseOldReaderCompatibility(t *testing.T) {
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

	if _, err = admin.Exec(ctx, "CREATE TABLE IF NOT EXISTS expand_compat (id serial primary key, name text not null)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	defer func() { _, _ = admin.Exec(ctx, "DROP TABLE IF EXISTS expand_compat CASCADE") }()

	var ex ExpandPhase

	// Expand phase: additive, non-blocking changes.
	if err := ExecExpandDDL(ctx, admin.Conn(), ex.AddColumnNullableDefault("expand_compat", "status", "text", "'pending'")); err != nil {
		t.Fatalf("add column: %v", err)
	}
	if err := ExecExpandDDL(ctx, admin.Conn(), ex.AddNotValidCheck("expand_compat", "chk_status", "status IN ('pending', 'done')")); err != nil {
		t.Fatalf("add not valid constraint: %v", err)
	}
	if _, err = admin.Exec(ctx, ex.CreateIndexConcurrently("expand_compat", "idx_expand_compat_status", "status")); err != nil {
		t.Fatalf("create index concurrently: %v", err)
	}

	// Old reader: only inserts into the original columns.
	if _, err = admin.Exec(ctx, "INSERT INTO expand_compat (name) VALUES ($1)", "old-reader"); err != nil {
		t.Fatalf("old reader insert: %v", err)
	}

	// New reader: inserts into the expanded columns.
	if _, err = admin.Exec(ctx, "INSERT INTO expand_compat (name, status) VALUES ($1, $2)", "new-reader", "done"); err != nil {
		t.Fatalf("new reader insert: %v", err)
	}

	// Verify both rows are present and the default was applied to the old reader.
	var oldStatus, newStatus string
	if err := admin.QueryRow(ctx, "SELECT status FROM expand_compat WHERE name='old-reader'").Scan(&oldStatus); err != nil {
		t.Fatalf("select old reader: %v", err)
	}
	if oldStatus != "pending" {
		t.Fatalf("old reader default status = %q, want pending", oldStatus)
	}
	if err := admin.QueryRow(ctx, "SELECT status FROM expand_compat WHERE name='new-reader'").Scan(&newStatus); err != nil {
		t.Fatalf("select new reader: %v", err)
	}
	if newStatus != "done" {
		t.Fatalf("new reader status = %q, want done", newStatus)
	}

	// Verify the NOT VALID constraint rejects new violating rows.
	if _, err = admin.Exec(ctx, "INSERT INTO expand_compat (name, status) VALUES ($1, $2)", "bad", "invalid"); err == nil {
		t.Fatal("expected NOT VALID constraint to reject invalid status")
	}
}
