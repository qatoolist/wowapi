package outbox_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/outbox"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// F-07 starvation regression (adversarial-framework-review-2026-07-17): a
// failed event must be requeued on schedule even while UNRELATED pending
// traffic keeps the drain loop busy. Before the fix, RequeueFailed ran only in
// the idle ticker branch, which a sustained producer bypasses forever.
func TestIntegrationRelayRequeueNotStarvedByBusyTraffic(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)

	reg := outbox.NewHandlerRegistry()
	reg.Subscribe("busy.tick", "noop", func(context.Context, database.TenantDB, outbox.DispatchedEvent) error {
		return nil
	})
	reg.Subscribe("starved.event", "noop", func(context.Context, database.TenantDB, outbox.DispatchedEvent) error {
		return nil
	})

	// Seed one FAILED event that becomes due only AFTER sustained draining has
	// begun (closure review 2026-07-17: a row already due at startup can be
	// recovered by a single requeue before the first dispatch, which would let
	// this test pass even if recovery ran only once). failed_at = now() with a
	// 2-second effective cooldown (wrapping the REAL RequeueFailed) means the
	// row is NOT due until ~2s of busy traffic have already elapsed.
	writer := outbox.NewWriter(model.UUIDv7())
	if err := h.TxM.WithTenant(testkit.TenantCtx(tenant.ID), func(ctx context.Context, db database.TenantDB) error {
		return writer.Write(ctx, db, outbox.Event{Type: "starved.event"})
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.Admin.Exec(context.Background(),
		`UPDATE events_outbox SET dispatch_status = 'failed', failed_at = now()
		  WHERE event_type = 'starved.event'`); err != nil {
		t.Fatal(err)
	}

	// Continuous unrelated pending producer: keeps DispatchOnce returning n>0
	// so the relay never reaches its idle branch.
	prodCtx, stopProducer := context.WithCancel(context.Background())
	defer stopProducer()
	var produced atomic.Int64
	go func() {
		for prodCtx.Err() == nil {
			_ = h.TxM.WithTenant(testkit.TenantCtx(tenant.ID), func(ctx context.Context, db database.TenantDB) error {
				return writer.Write(ctx, db, outbox.Event{Type: "busy.tick"})
			})
			produced.Add(1)
			time.Sleep(2 * time.Millisecond)
		}
	}()

	relayCtx, stopRelay := context.WithCancel(context.Background())
	defer stopRelay()
	relay := outbox.NewRelay(h.Platform, h.TxM, reg, 1) // batch=1: guaranteed continuous drain
	// Wrap the REAL requeue with a 2s cooldown so the seeded row becomes due
	// mid-drain, never at startup.
	outbox.SetRequeueHook(relay, func(ctx context.Context, _ time.Duration) error {
		return relay.RequeueFailed(ctx, 2*time.Second)
	})
	relayDone := make(chan error, 1)
	go func() { relayDone <- relay.Run(relayCtx, 50*time.Millisecond) }()

	// The failed event must leave 'failed' (pending or dispatched) within a
	// bounded window despite the busy loop.
	deadline := time.After(8 * time.Second)
	for {
		var status string
		if err := h.Admin.QueryRow(context.Background(),
			`SELECT dispatch_status FROM events_outbox WHERE event_type = 'starved.event'`).Scan(&status); err != nil {
			t.Fatal(err)
		}
		if status != "failed" {
			break // requeued (pending) or already dispatched — recovery ran while busy
		}
		select {
		case <-deadline:
			stopRelay()
			<-relayDone
			t.Fatalf("failed event still 'failed' after 8s of busy traffic (%d unrelated events produced) — requeue starved by the drain loop", produced.Load())
		case <-time.After(50 * time.Millisecond):
		}
	}
	stopProducer()
	stopRelay()
	if err := <-relayDone; err != nil {
		t.Fatalf("relay.Run: %v", err)
	}
}
