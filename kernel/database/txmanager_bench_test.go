package database_test

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/testkit"
)

// BenchmarkTenantTransactionOpenCommit measures the real PostgreSQL
// BEGIN -> SET LOCAL tenant binding -> COMMIT path used by every tenant request.
func BenchmarkTenantTransactionOpenCommit(b *testing.B) {
	h := testkit.NewDB(b)
	tenant := testkit.CreateTenantTB(b, h)
	ctx := testkit.TenantCtx(tenant.ID)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if err := h.TxM.WithTenant(ctx, func(context.Context, database.TenantDB) error { return nil }); err != nil {
			b.Fatalf("tenant transaction: %v", err)
		}
	}
}
