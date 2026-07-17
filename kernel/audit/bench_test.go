package audit_test

// DB-backed hot-path benchmark for the audit writer (roadmap E1/S6, backlog B-2).
//
// Writer.Record is on the business-write hot path: every recorded change appends
// a row inside the caller's transaction AND extends the tenant's SHA-256 hash
// chain (chainHash over the length-prefixed fields). This benchmark exercises the
// real code path against Postgres via testkit — the three round trips (lock chain
// head, insert row, advance head) plus the chain hash — so it catches a
// regression in either the SQL or the hashing. It is not an in-memory stand-in.

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/audit"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// BenchmarkAuditRecord measures one field-level audit append inside a tenant
// transaction — the real per-change cost, including the hash-chain extension.
func BenchmarkAuditRecord(b *testing.B) {
	h := testkit.NewDB(b)
	w := audit.New(model.UUIDv7(), nil)
	tenant, actor := uuid.New(), uuid.New()
	ctx := auditCtx(tenant, actor, "bench-req")
	entity := uuid.New()
	e := audit.Entry{
		Action: "receipt.update", EntityType: "receipt", EntityID: entity,
		Field: "amount", OldValue: "100", NewValue: "150", ActorKind: "user",
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			return w.Record(ctx, db, e)
		}); err != nil {
			b.Fatalf("record: %v", err)
		}
	}
}
