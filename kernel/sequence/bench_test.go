package sequence_test

// DB-backed hot-path benchmark for the numbered-series allocator (roadmap E3,
// backlog B-2).
//
// Allocate is the statutory-numbering hot path: it runs INSIDE the caller's
// tenant transaction, advancing the counter row (INSERT … ON CONFLICT DO UPDATE)
// and writing a ledger row, so a number is consumed only on commit and concurrent
// callers serialize on the counter. This benchmark drives the real path against
// Postgres via testkit — not an in-memory stand-in — so it catches a regression
// in the allocation SQL.

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/sequence"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// BenchmarkSequenceAllocate measures one gap-free number allocation inside a
// tenant transaction — the counter increment plus the ledger-row insert.
func BenchmarkSequenceAllocate(b *testing.B) {
	h := testkit.NewDB(b)
	a := sequence.New(model.UUIDv7())
	ctx := database.WithTenantID(context.Background(), uuid.New())

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			_, err := a.Allocate(ctx, db, "receipt")
			return err
		}); err != nil {
			b.Fatalf("allocate: %v", err)
		}
	}
}
