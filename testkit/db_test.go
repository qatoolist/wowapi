package testkit

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/migrations"
)

// TestIntegrationMigrateFreshAndIdempotent drives Migrate against a genuinely
// EMPTY database (not the pre-migrated template) so the fresh path is under
// assertion, then reruns it: the second run must apply ZERO migrations and
// leave the version unchanged (acceptance #2 — fresh DB migrates idempotently;
// review finding ARCH-18: assert Applied==0, not just version equality).
func TestIntegrationMigrateFreshAndIdempotent(t *testing.T) {
	dsn := adminDSN(t) // also performs the skip-when-absent check
	ctx := context.Background()

	// Make a brand-new empty database to exercise the real fresh path (the
	// per-test clone from NewDB would already be migrated from the template).
	fresh := testDBName(t)
	createEmptyDB(ctx, t, dsn, fresh)
	t.Cleanup(func() { dropTestDB(context.Background(), dsn, fresh) })

	pool, err := newPoolDB(ctx, dsn, fresh, 2)
	if err != nil {
		t.Fatalf("fresh pool: %v", err)
	}
	defer pool.Close()

	r1, err := database.Migrate(ctx, pool, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("fresh Migrate: %v", err)
	}
	if r1.Applied == 0 {
		t.Fatal("fresh migrate applied 0 migrations — expected the full kernel set")
	}
	r2, err := database.Migrate(ctx, pool, migrations.Kernel(), migrations.SourceName)
	if err != nil {
		t.Fatalf("rerun Migrate: %v", err)
	}
	if r2.Applied != 0 {
		t.Fatalf("rerun applied %d migrations — must be a no-op", r2.Applied)
	}
	if r1.Version != r2.Version {
		t.Fatalf("version changed across idempotent runs: %d then %d", r1.Version, r2.Version)
	}
}

// TestIntegrationRLSProbe proves the full isolation contract on a probe table.
func TestIntegrationRLSProbe(t *testing.T) {
	h := NewDB(t)
	table := CreateProbeTable(t, h)
	AssertRLSIsolation(t, h, table, func(tenant uuid.UUID) map[string]any {
		return map[string]any{"id": uuid.New(), "note": "x"}
	})
}

// TestIntegrationNoTenantContextFails asserts WithTenant fails closed when no
// tenant was bound to the context (acceptance #5).
func TestIntegrationNoTenantContextFails(t *testing.T) {
	h := NewDB(t)
	err := h.TxM.WithTenant(context.Background(), func(ctx context.Context, db database.TenantDB) error {
		t.Error("fn ran despite missing tenant context")
		return nil
	})
	if !errors.Is(err, database.ErrNoTenantContext) {
		t.Fatalf("WithTenant without tenant = %v, want ErrNoTenantContext", err)
	}
}

// TestIntegrationKernelTablesExist checks the global spine tables shipped by
// migration 00002 exist and carry NO row security (they are global, D-0025).
func TestIntegrationKernelTablesExist(t *testing.T) {
	h := NewDB(t)
	ctx := context.Background()
	for _, tbl := range []string{"tenants", "users", "user_tenant_access"} {
		var exists bool
		if err := h.Admin.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM information_schema.tables
			 WHERE table_schema = 'public' AND table_name = $1)`, tbl).Scan(&exists); err != nil {
			t.Fatalf("query information_schema for %s: %v", tbl, err)
		}
		if !exists {
			t.Errorf("kernel table %q does not exist", tbl)
			continue
		}
		var rowsecurity bool
		if err := h.Admin.QueryRow(ctx,
			`SELECT rowsecurity FROM pg_tables WHERE schemaname = 'public' AND tablename = $1`,
			tbl).Scan(&rowsecurity); err != nil {
			t.Fatalf("query pg_tables for %s: %v", tbl, err)
		}
		if rowsecurity {
			t.Errorf("global table %q has row security enabled; want none (D-0025)", tbl)
		}
	}
}

// TestIntegrationReadOnlyTx asserts WithTenantRO rejects writes (BEGIN READ ONLY).
func TestIntegrationReadOnlyTx(t *testing.T) {
	h := NewDB(t)
	table := CreateProbeTable(t, h)
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return insertRow(ctx, db, table, map[string]any{"id": uuid.New(), "tenant_id": tenant, "note": "x"})
	})
	if err == nil {
		t.Fatal("INSERT succeeded inside a read-only transaction")
	}
}

// TestIntegrationVersionConflictHelper drives ExpectOneRow: an UPDATE matching
// zero rows surfaces as ErrVersionConflict.
func TestIntegrationVersionConflictHelper(t *testing.T) {
	h := NewDB(t)
	table := CreateProbeTable(t, h)
	tenant := uuid.New()
	ctx := database.WithTenantID(context.Background(), tenant)

	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		tag, err := db.Exec(ctx,
			"UPDATE "+quoteIdent(table)+" SET note = $1 WHERE id = $2", "y", uuid.New())
		if err != nil {
			return err
		}
		return database.ExpectOneRow(tag, "probe")
	})
	if !errors.Is(err, database.ErrVersionConflict) {
		t.Fatalf("ExpectOneRow on 0-row UPDATE = %v, want ErrVersionConflict", err)
	}
}

// TestIntegrationRoleReassertedPerTx proves SEC-11: a transaction that tries to
// revert its role in-band (RESET ROLE / SET ROLE NONE) must not escalate its
// own visibility, and must not leak state into a later tenant transaction. The
// TxManager re-asserts SET LOCAL ROLE app_rt (and the RLS guard) at the start
// of every tenant tx, so isolation holds deterministically regardless of which
// pooled connection each tx lands on. Both tenants are checked.
func TestIntegrationRoleReassertedPerTx(t *testing.T) {
	h := NewDB(t)
	table := CreateProbeTable(t, h)
	tenantA, tenantB := uuid.New(), uuid.New()
	ctx := context.Background()

	// Seed one row per tenant via the admin (owner) pool.
	for _, tn := range []uuid.UUID{tenantA, tenantB} {
		if _, err := h.Admin.Exec(ctx,
			"INSERT INTO "+quoteIdent(table)+" (id, tenant_id, note) VALUES ($1,$2,'x')",
			uuid.New(), tn); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	// Tx 1 (tenant A) tries to escalate by reverting its role mid-transaction,
	// then reads: it must still see only its own row despite the RESET.
	var aVisible int
	if err := h.TxM.WithTenant(database.WithTenantID(ctx, tenantA), func(ctx context.Context, db database.TenantDB) error {
		_, _ = db.Exec(ctx, "RESET ROLE")    // attempt to become the login role
		_, _ = db.Exec(ctx, "SET ROLE NONE") // and again, non-locally
		return db.QueryRow(ctx, "SELECT count(*) FROM "+quoteIdent(table)).Scan(&aVisible)
	}); err != nil {
		t.Fatalf("tenant A tx: %v", err)
	}
	if aVisible != 1 {
		t.Fatalf("tenant A saw %d rows after RESET ROLE — in-tx role escalation (SEC-11)", aVisible)
	}

	// Tx 2 (tenant B) is an innocent request on the (now potentially poisoned)
	// connection. It must see EXACTLY its own one row — never tenant A's.
	var visible int
	if err := h.TxM.WithTenant(database.WithTenantID(ctx, tenantB), func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx, "SELECT count(*) FROM "+quoteIdent(table)).Scan(&visible)
	}); err != nil {
		t.Fatalf("tenant B tx: %v", err)
	}
	if visible != 1 {
		t.Fatalf("tenant B saw %d rows — role leaked across the pooled connection (SEC-11)", visible)
	}
}
