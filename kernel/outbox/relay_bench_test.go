package outbox_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/outbox"
	"github.com/qatoolist/wowapi/v2/kernel/resource"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// BenchmarkRelayDispatchBatch measures a real ten-event PostgreSQL relay batch:
// leased claim, inbox dedup, tenant handler transaction, and fenced finalize.
// Fixture writes are excluded from the timed region.
func BenchmarkRelayDispatchBatch(b *testing.B) {
	const batchSize = 10
	h := testkit.NewDB(b)
	tenant := testkit.CreateTenantTB(b, h)
	ctx := testkit.TenantCtx(tenant.ID)
	if _, err := h.Admin.Exec(context.Background(),
		`INSERT INTO resource_types (key, module, description)
		 VALUES ('benchmark.relay', 'benchmark', 'relay benchmark')
		 ON CONFLICT (key) DO NOTHING`); err != nil {
		b.Fatalf("create resource type: %v", err)
	}
	resources := make([]resource.Ref, batchSize)
	for i := range resources {
		resources[i] = resource.Ref{Type: "benchmark.relay", ID: uuid.New()}
		if _, err := h.Admin.Exec(context.Background(),
			`INSERT INTO resources (id, tenant_id, resource_type, label, status, created_by)
			 VALUES ($1, $2, $3, 'relay benchmark', 'active', $4)`,
			resources[i].ID, tenant.ID, resources[i].Type, uuid.Nil); err != nil {
			b.Fatalf("create resource: %v", err)
		}
	}
	writer := outbox.NewWriter(model.UUIDv7())
	registry := outbox.NewHandlerRegistry()
	registry.Subscribe("benchmark.relay.ready", "benchmark-handler",
		func(context.Context, database.TenantDB, outbox.DispatchedEvent) error { return nil })
	relay := outbox.NewRelay(h.Platform, h.TxM, registry, batchSize)

	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		b.StopTimer()
		err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			for _, ref := range resources {
				if err := writer.Write(ctx, db, outbox.Event{Type: "benchmark.relay.ready", Resource: ref}); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			b.Fatalf("seed relay batch: %v", err)
		}
		b.StartTimer()
		processed, err := relay.DispatchOnce(ctx)
		if err != nil || processed != batchSize {
			b.Fatalf("dispatch batch: processed=%d err=%v", processed, err)
		}
	}
}
