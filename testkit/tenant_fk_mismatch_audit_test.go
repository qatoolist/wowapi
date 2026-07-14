package testkit

import (
	"context"
	"fmt"
	"testing"
)

// TestIntegrationTenantFKMismatchAuditZero is DATA-01 T3 (W02-E02-S002): prove
// child.tenant_id = parent.tenant_id for every existing row across all 8
// tenant-scoped FK edges. The test uses the platform-role Admin pool to bypass
// RLS, exactly as the audit is required to do, and asserts a zero-mismatch
// report against the migrated schema.
func TestIntegrationTenantFKMismatchAuditZero(t *testing.T) {
	h := NewDB(t)
	ctx := context.Background()

	// Seed at least one legitimate parent/child pair per edge so the audit is
	// exercising real joins, not just empty tables.
	for _, e := range tenantFKEdges() {
		tenant := CreateTenant(t, h).ID
		parentID := e.seedParent(t, h, tenant)
		row := withTenant(e.childRow(t, h, tenant, parentID), tenant)
		if err := adminInsert(ctx, h, e.child, row); err != nil {
			t.Fatalf("seed legitimate %s row: %v", e.child, err)
		}
	}

	var total int64
	for _, e := range tenantFKEdges() {
		var n int64
		q := fmt.Sprintf(
			"SELECT COUNT(*) FROM %s c JOIN %s p ON c.%s = p.id WHERE c.tenant_id <> p.tenant_id",
			e.child, e.parent, e.fkCol,
		)
		if err := h.Admin.QueryRow(ctx, q).Scan(&n); err != nil {
			t.Fatalf("audit query for %s: %v", e.constraint, err)
		}
		if n != 0 {
			t.Errorf("edge %s: found %d cross-tenant mismatches", e.constraint, n)
		}
		total += n
	}
	if total != 0 {
		t.Fatalf("DATA-01 mismatch audit found %d total cross-tenant rows", total)
	}
	t.Logf("DATA-01 mismatch audit: %d edges checked, 0 mismatches", len(tenantFKEdges()))
}
