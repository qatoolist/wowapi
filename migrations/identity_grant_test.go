package migrations_test

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/migrations"
	"github.com/qatoolist/wowapi/testkit"
)

// TestIdentityGrantMigrationUpDown verifies that the identity_grant migration
// applies cleanly and that goose Down rolls it back to a state where the table
// no longer exists. This is the migration reversibility drill for the new
// table.
func TestIdentityGrantMigrationUpDown(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	// The test database is already migrated up by testkit. Confirm the table
	// exists, then roll back to version 0 and back up again.
	var exists bool
	if err := h.Admin.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables
		 WHERE table_schema = 'public' AND table_name = 'identity_grant')`).Scan(&exists); err != nil {
		t.Fatalf("check table exists: %v", err)
	}
	if !exists {
		t.Fatal("identity_grant table missing before down test")
	}

	if _, err := database.MigrateReset(ctx, h.Admin, migrations.Kernel(), migrations.SourceName); err != nil {
		t.Fatalf("migration down to 0: %v", err)
	}

	if err := h.Admin.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables
		 WHERE table_schema = 'public' AND table_name = 'identity_grant')`).Scan(&exists); err != nil {
		t.Fatalf("check table after down: %v", err)
	}
	if exists {
		t.Fatal("identity_grant table still exists after goose Down")
	}

	if _, err := database.Migrate(ctx, h.Admin, migrations.Kernel(), migrations.SourceName); err != nil {
		t.Fatalf("migration up after down: %v", err)
	}

	if err := h.Admin.QueryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM information_schema.tables
		 WHERE table_schema = 'public' AND table_name = 'identity_grant')`).Scan(&exists); err != nil {
		t.Fatalf("check table after re-up: %v", err)
	}
	if !exists {
		t.Fatal("identity_grant table missing after re-up")
	}
}

// TestIdentityGrantRLSCatalog verifies that identity_grant is registered in
// the catalog as a FORCE-RLS tenant-scoped table with the expected policies,
// the correct role grants, and the one-active-grant-per-actor partial unique
// index.
func TestIdentityGrantRLSCatalog(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	var enabled, forced bool
	if err := h.Admin.QueryRow(ctx,
		`SELECT c.relrowsecurity, c.relforcerowsecurity
		   FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		  WHERE n.nspname = 'public' AND c.relname = 'identity_grant'`).Scan(&enabled, &forced); err != nil {
		t.Fatalf("query pg_class for identity_grant: %v", err)
	}
	if !enabled {
		t.Error("identity_grant: ROW LEVEL SECURITY not enabled")
	}
	if !forced {
		t.Error("identity_grant: FORCE ROW LEVEL SECURITY not enabled")
	}

	var tenantPolicy, platformPolicy bool
	if err := h.Admin.QueryRow(ctx,
		`SELECT
		  EXISTS (SELECT 1 FROM pg_policies p WHERE p.schemaname='public' AND p.tablename='identity_grant' AND p.policyname='identity_grant_tenant_isolation'),
		  EXISTS (SELECT 1 FROM pg_policies p WHERE p.schemaname='public' AND p.tablename='identity_grant' AND p.policyname='identity_grant_platform_all')`).
		Scan(&tenantPolicy, &platformPolicy); err != nil {
		t.Fatalf("query policies for identity_grant: %v", err)
	}
	if !tenantPolicy {
		t.Error("identity_grant: missing tenant isolation policy")
	}
	if !platformPolicy {
		t.Error("identity_grant: missing app_platform bypass policy")
	}

	var appRTSelect, appRTInsert, appRTUpdate bool
	if err := h.Admin.QueryRow(ctx,
		`SELECT
		  EXISTS (SELECT 1 FROM information_schema.role_table_grants g WHERE g.table_schema='public' AND g.table_name='identity_grant' AND g.grantee='app_rt' AND g.privilege_type='SELECT'),
		  EXISTS (SELECT 1 FROM information_schema.role_table_grants g WHERE g.table_schema='public' AND g.table_name='identity_grant' AND g.grantee='app_rt' AND g.privilege_type='INSERT'),
		  EXISTS (SELECT 1 FROM information_schema.role_table_grants g WHERE g.table_schema='public' AND g.table_name='identity_grant' AND g.grantee='app_rt' AND g.privilege_type='UPDATE')`).
		Scan(&appRTSelect, &appRTInsert, &appRTUpdate); err != nil {
		t.Fatalf("query app_rt grants for identity_grant: %v", err)
	}
	if appRTSelect || appRTInsert || appRTUpdate {
		t.Errorf("identity_grant: app_rt must have no grants, got SELECT=%v INSERT=%v UPDATE=%v", appRTSelect, appRTInsert, appRTUpdate)
	}

	var platformSelect, platformInsert, platformUpdate bool
	if err := h.Admin.QueryRow(ctx,
		`SELECT
		  EXISTS (SELECT 1 FROM information_schema.role_table_grants g WHERE g.table_schema='public' AND g.table_name='identity_grant' AND g.grantee='app_platform' AND g.privilege_type='SELECT'),
		  EXISTS (SELECT 1 FROM information_schema.role_table_grants g WHERE g.table_schema='public' AND g.table_name='identity_grant' AND g.grantee='app_platform' AND g.privilege_type='INSERT'),
		  EXISTS (SELECT 1 FROM information_schema.role_table_grants g WHERE g.table_schema='public' AND g.table_name='identity_grant' AND g.grantee='app_platform' AND g.privilege_type='UPDATE')`).
		Scan(&platformSelect, &platformInsert, &platformUpdate); err != nil {
		t.Fatalf("query app_platform grants for identity_grant: %v", err)
	}
	if !platformSelect {
		t.Error("identity_grant: app_platform missing SELECT grant")
	}
	if !platformInsert {
		t.Error("identity_grant: app_platform missing INSERT grant")
	}
	if !platformUpdate {
		t.Error("identity_grant: app_platform missing UPDATE grant")
	}

	var indexExists bool
	if err := h.Admin.QueryRow(ctx,
		`SELECT EXISTS (
		  SELECT 1 FROM pg_indexes
		   WHERE schemaname = 'public' AND tablename = 'identity_grant'
		     AND indexname = 'identity_grant_one_active_per_actor'
		)`).Scan(&indexExists); err != nil {
		t.Fatalf("query index for identity_grant: %v", err)
	}
	if !indexExists {
		t.Error("identity_grant: partial unique index identity_grant_one_active_per_actor missing")
	}
}
