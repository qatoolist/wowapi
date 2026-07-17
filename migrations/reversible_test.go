package migrations_test

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/migrations"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIntegrationMigrationsReversible proves the clean all-up → all-down →
// all-up reconstruction invariant: the migration set applies to a head, rolls
// fully back to an empty schema, and reconstructs the identical head. Rollback
// Down keeps cluster-scoped roles and database-scoped extensions (which may be
// shared by other schemas/modules), while removing every kernel-owned object.
//
// The abandoned-V1 replay (rebuild at the old v1.0.0 migration head, seed
// disposable v1 data, upgrade) was removed in the clean-V1 reset: the pre-reset
// v1.0.0/v1.1.0 releases are unsupported and no upgrade path from them exists.
// A genuine N-1 drill returns only once the clean line has a real predecessor
// release (>= v1.2.0).
func TestIntegrationMigrationsReversible(t *testing.T) {
	h := testkit.NewDB(t) // isolated per-test DB, already migrated to head
	ctx := context.Background()

	head, err := database.Migrate(ctx, h.Admin, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("read head: %v", err)
	}
	if head.Version != 1 {
		t.Fatalf("clean kernel head = %d, want the single baseline at version 1", head.Version)
	}
	if !tableExists(t, h, "idempotency_keys") {
		t.Fatal("head schema should contain idempotency_keys")
	}

	// Full rollback to an empty schema.
	v, err := database.MigrateReset(ctx, h.Admin, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("migrate down: %v", err)
	}
	if v != 0 {
		t.Fatalf("after full rollback version = %d, want 0", v)
	}
	if tableExists(t, h, "idempotency_keys") {
		t.Fatal("full rollback should have dropped idempotency_keys")
	}
	var retainedExtensions int
	if err := h.Admin.QueryRow(ctx, `SELECT count(*) FROM pg_extension
		WHERE extname IN ('btree_gist', 'citext')`).Scan(&retainedExtensions); err != nil {
		t.Fatalf("check retained extensions: %v", err)
	}
	if retainedExtensions != 2 {
		t.Fatalf("database-scoped extensions retained after reset = %d, want 2", retainedExtensions)
	}

	// Clean forward reconstruction to the identical head.
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
