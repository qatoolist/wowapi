package migrations_test

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/migrations"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationMigrationsReversible is the O2 forward/down drill: on an
// isolated database already at head, roll every migration back to 0 and forward
// again, proving each migration's `-- +goose Down` block is present and correct
// and that a full re-apply reproduces the head version. Safe for concurrent
// tests: rollback of 00001 keeps the cluster-scoped roles and only revokes
// schema-public usage (a per-database object).
func TestIntegrationMigrationsReversible(t *testing.T) {
	h := testkit.NewDB(t) // isolated per-test DB, already migrated to head
	ctx := context.Background()

	// Read the current head version (idempotent Up is a no-op here).
	head, err := database.Migrate(ctx, h.Admin, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("read head: %v", err)
	}
	if head.Version == 0 {
		t.Fatal("expected a migrated head database")
	}
	if !tableExists(t, h, "idempotency_keys") {
		t.Fatal("head schema should contain idempotency_keys")
	}

	// Roll everything back.
	v, err := database.MigrateReset(ctx, h.Admin, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("migrate down: %v", err)
	}
	if v != 0 {
		t.Fatalf("after full rollback version = %d, want 0", v)
	}
	if tableExists(t, h, "idempotency_keys") {
		t.Fatal("down migrations should have dropped idempotency_keys")
	}

	// Roll forward again — must return to the same head version.
	reup, err := database.Migrate(ctx, h.Admin, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("re-up: %v", err)
	}
	if reup.Version != head.Version {
		t.Fatalf("re-up version = %d, want head %d", reup.Version, head.Version)
	}
	if !tableExists(t, h, "idempotency_keys") {
		t.Fatal("re-up should have recreated idempotency_keys")
	}
}

func tableExists(t *testing.T, h *testkit.DBHandle, name string) bool {
	t.Helper()
	var exists bool
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables
		                 WHERE table_schema = 'public' AND table_name = $1)`, name).Scan(&exists); err != nil {
		t.Fatalf("table check %q: %v", name, err)
	}
	return exists
}
