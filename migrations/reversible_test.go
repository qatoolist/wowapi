package migrations_test

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/migrations"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestIntegrationMigrationsReversible is the REL-03 oldest-supported upgrade
// and reversibility drill. It rebuilds the isolated database at the v1.0.0
// migration head, seeds disposable v1 data, upgrades it to the current head,
// proves the data survived, then rolls all migrations down and forward again.
// Rollback of 00001 keeps cluster-scoped roles and only revokes schema-public
// usage, so the drill remains isolated to the per-test database.
func TestIntegrationMigrationsReversible(t *testing.T) {
	h := testkit.NewDB(t) // isolated per-test DB, already migrated to head
	ctx := context.Background()

	const oldestSupportedVersion = 28 // v1.0.0 migration head; v1 is N/N-1 minor compatible.

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

	// Rebuild at the oldest supported release and seed disposable data there.
	v, err := database.MigrateReset(ctx, h.Admin, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("migrate down before oldest-supported seed: %v", err)
	}
	if v != 0 {
		t.Fatalf("after full rollback version = %d, want 0", v)
	}
	oldest, err := database.MigrateTo(ctx, h.Admin, migrations.Kernel(), migrations.SourceName, oldestSupportedVersion)
	if err != nil {
		t.Fatalf("migrate to oldest supported version: %v", err)
	}
	if oldest.Version != oldestSupportedVersion {
		t.Fatalf("oldest-supported version = %d, want %d", oldest.Version, oldestSupportedVersion)
	}
	const tenantID = "10000000-0000-0000-0000-000000000001"
	if _, err := h.Admin.Exec(ctx, `INSERT INTO tenants (id, slug, display_name, created_by) VALUES ($1, 'compat-drill', 'Compatibility Drill', $1)`, tenantID); err != nil {
		t.Fatalf("seed oldest-supported data: %v", err)
	}

	// Upgrade the oldest supported schema and prove its data survives.
	reup, err := database.Migrate(ctx, h.Admin, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("re-up: %v", err)
	}
	if reup.Version != head.Version {
		t.Fatalf("re-up version = %d, want head %d", reup.Version, head.Version)
	}
	var displayName string
	if err := h.Admin.QueryRow(ctx, `SELECT display_name FROM tenants WHERE id = $1`, tenantID).Scan(&displayName); err != nil {
		t.Fatalf("read upgraded oldest-supported data: %v", err)
	}
	if displayName != "Compatibility Drill" {
		t.Fatalf("upgraded tenant display_name = %q", displayName)
	}

	// Reverse on disposable data, then prove a clean forward reconstruction.
	v, err = database.MigrateReset(ctx, h.Admin, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("migrate down after upgrade: %v", err)
	}
	if v != 0 {
		t.Fatalf("after upgraded rollback version = %d, want 0", v)
	}
	reup, err = database.Migrate(ctx, h.Admin, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("re-up after upgraded rollback: %v", err)
	}
	if reup.Version != head.Version {
		t.Fatalf("re-up after upgraded rollback version = %d, want head %d", reup.Version, head.Version)
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
