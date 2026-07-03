package testkit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
)

// RowFactory produces the non-tenant columns for one probe row. It is called
// once per insert (once as tenant A, once for the WITH CHECK probe) so unique
// columns such as the primary key differ between calls — the assert supplies
// tenant_id itself, so the factory MUST NOT set it (any tenant_id it returns is
// overwritten). The uuid argument is the tenant the row is being written under.
type RowFactory func(tenant uuid.UUID) map[string]any

// AssertRLSIsolation proves the four tenant-isolation properties for a
// tenant-scoped table (03 §1): (1) rows written as tenant A are invisible to
// tenant B; (2) a query without tenant context fails (no default tenant);
// (3) WITH CHECK blocks writing a row whose tenant_id differs from the bound
// tenant; (4) the writing tenant sees its own row. The table must carry the
// standard tenant_id column; row describes one minimal row EXCLUDING tenant_id.
func AssertRLSIsolation(t *testing.T, h *DBHandle, table string, row RowFactory) {
	t.Helper()
	if !identRE.MatchString(table) {
		t.Fatalf("testkit: invalid table name %q", table)
	}

	tenantA := uuid.New()
	tenantB := uuid.New()
	ctxA := database.WithTenantID(context.Background(), tenantA)
	ctxB := database.WithTenantID(context.Background(), tenantB)

	// (4) tenant A writes its own row (committed in its own tx).
	if err := h.TxM.WithTenant(ctxA, func(ctx context.Context, db database.TenantDB) error {
		return insertRow(ctx, db, table, withTenant(row(tenantA), tenantA))
	}); err != nil {
		t.Fatalf("testkit: insert as tenant A: %v", err)
	}

	// (1) tenant B sees none of tenant A's rows.
	if n := countAs(t, h, ctxB, table); n != 0 {
		t.Errorf("testkit: RLS leak — tenant B sees %d row(s) of tenant A, want 0", n)
	}

	// (4) tenant A sees exactly its own row.
	if n := countAs(t, h, ctxA, table); n != 1 {
		t.Errorf("testkit: tenant A sees %d row(s) of its own data, want 1", n)
	}

	// (2) no tenant on the context fails closed before any query runs.
	err := h.TxM.WithTenant(context.Background(), func(ctx context.Context, db database.TenantDB) error {
		return nil
	})
	if !errors.Is(err, database.ErrNoTenantContext) {
		t.Errorf("testkit: WithTenant without tenant context = %v, want ErrNoTenantContext", err)
	}

	// (2, defense in depth) a raw runtime query with no SET LOCAL binding must
	// ERROR — app_tenant_id() raises when app.tenant_id is absent (fail closed),
	// so assert an error, NOT zero rows.
	var n int
	rawErr := h.Runtime.QueryRow(context.Background(),
		"SELECT count(*) FROM "+quoteIdent(table)).Scan(&n)
	if rawErr == nil {
		t.Errorf("testkit: raw runtime SELECT with no tenant binding returned %d rows, want an error", n)
	}

	// (3) WITH CHECK blocks writing a row for a different tenant.
	err = h.TxM.WithTenant(ctxA, func(ctx context.Context, db database.TenantDB) error {
		return insertRow(ctx, db, table, withTenant(row(tenantA), tenantB))
	})
	if err == nil {
		t.Errorf("testkit: WITH CHECK did not block tenant A writing a tenant B row")
	}
}

// countAs counts visible rows under the tenant bound in ctx.
func countAs(t *testing.T, h *DBHandle, ctx context.Context, table string) int {
	t.Helper()
	var n int
	if err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx, "SELECT count(*) FROM "+quoteIdent(table)).Scan(&n)
	}); err != nil {
		t.Fatalf("testkit: count rows: %v", err)
	}
	return n
}

// withTenant returns cols with tenant_id forced to id (any existing tenant_id
// is overwritten).
func withTenant(cols map[string]any, id uuid.UUID) map[string]any {
	if cols == nil {
		cols = map[string]any{}
	}
	cols["tenant_id"] = id
	return cols
}

// insertRow builds a parameterized INSERT from sorted, validated column names.
func insertRow(ctx context.Context, db database.TenantDB, table string, cols map[string]any) error {
	keys := sortedKeys(cols)
	names := make([]string, len(keys))
	placeholders := make([]string, len(keys))
	args := make([]any, len(keys))
	for i, k := range keys {
		if !identRE.MatchString(k) {
			return fmt.Errorf("testkit: invalid column name %q", k)
		}
		names[i] = quoteIdent(k)
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = cols[k]
	}
	sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		quoteIdent(table), strings.Join(names, ", "), strings.Join(placeholders, ", "))
	_, err := db.Exec(ctx, sql, args...)
	return err
}
