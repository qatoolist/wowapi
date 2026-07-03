package outbox_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/testkit"
)

func countEvents(t *testing.T, h *testkit.DBHandle, tenant uuid.UUID, typ string) int {
	t.Helper()
	var n int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM events_outbox WHERE tenant_id = $1 AND event_type = $2`, tenant, typ).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// TestIntegrationOutboxAtomicWithBusinessTx proves the event is emitted iff the
// business transaction commits: a tx that returns an error rolls back the event
// too; a committing tx persists it.
func TestIntegrationOutboxAtomicWithBusinessTx(t *testing.T) {
	h := testkit.NewDB(t)
	w := outbox.NewWriter(model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	// Rolled-back business tx: no event.
	wantErr := errors.New("business failure")
	err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		if e := w.Write(ctx, db, outbox.Event{Type: "requests.request.created"}); e != nil {
			return e
		}
		return wantErr // roll back
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected business error, got %v", err)
	}
	if n := countEvents(t, h, tn.ID, "requests.request.created"); n != 0 {
		t.Fatalf("event survived a rolled-back tx: %d", n)
	}

	// Committed business tx: event present.
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return w.Write(ctx, db, outbox.Event{Type: "requests.request.created", Payload: map[string]any{"x": 1}})
	}); err != nil {
		t.Fatal(err)
	}
	if n := countEvents(t, h, tn.ID, "requests.request.created"); n != 1 {
		t.Fatalf("committed event missing: %d", n)
	}
}

// TestIntegrationOutboxRelayDispatchAndInbox proves the relay dispatches a
// pending event to its handler exactly once (inbox dedup on redelivery) within
// the event's tenant.
func TestIntegrationOutboxRelayDispatchAndInbox(t *testing.T) {
	h := testkit.NewDB(t)
	w := outbox.NewWriter(model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	var (
		mu         sync.Mutex
		seen       []uuid.UUID
		seenTenant uuid.UUID
	)
	reg := outbox.NewHandlerRegistry()
	reg.Subscribe("requests.request.approved", "requests.notify", func(ctx context.Context, db database.TenantDB, e outbox.DispatchedEvent) error {
		mu.Lock()
		seen = append(seen, e.ID)
		seenTenant = e.TenantID
		mu.Unlock()
		var p map[string]any
		_ = json.Unmarshal(e.Payload, &p)
		return nil
	})

	// Emit two events.
	for i := 0; i < 2; i++ {
		if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
			return w.Write(ctx, db, outbox.Event{Type: "requests.request.approved", Payload: map[string]any{"i": i}})
		}); err != nil {
			t.Fatal(err)
		}
	}

	relay := outbox.NewRelay(h.Platform, h.TxM, reg, 10)
	n, err := relay.DispatchOnce(context.Background())
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if n != 2 {
		t.Fatalf("dispatched %d events, want 2", n)
	}
	if len(seen) != 2 {
		t.Fatalf("handler saw %d events, want 2", len(seen))
	}
	if seenTenant != tn.ID {
		t.Fatalf("handler saw tenant %s, want %s", seenTenant, tn.ID)
	}

	// A second dispatch pass has nothing pending (already dispatched); and even
	// if the same events were re-presented, the inbox would dedup. Re-mark one
	// as pending and re-dispatch: the handler must NOT run again (inbox).
	if _, err := h.Admin.Exec(context.Background(),
		`UPDATE events_outbox SET dispatch_status='pending' WHERE tenant_id=$1`, tn.ID); err != nil {
		t.Fatal(err)
	}
	before := len(seen)
	if _, err := relay.DispatchOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(seen) != before {
		t.Fatalf("inbox dedup failed: handler ran again (%d → %d)", before, len(seen))
	}
}

// TestIntegrationOutboxPerAggregateOrderUnderRetry proves per-aggregate order
// survives a transient handler failure (review finding ARCH-53): two events for
// the same aggregate, the handler fails the FIRST once. The relay must not
// dispatch the second before the first succeeds — the handler observes them in
// occurred_at order.
func TestIntegrationOutboxPerAggregateOrderUnderRetry(t *testing.T) {
	h := testkit.NewDB(t)
	w := outbox.NewWriter(model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	var mu sync.Mutex
	var order []int
	failFirst := int64(1)
	reg := outbox.NewHandlerRegistry()
	reg.Subscribe("requests.request.changed", "order.check", func(ctx context.Context, db database.TenantDB, e outbox.DispatchedEvent) error {
		var p struct {
			Seq int `json:"seq"`
		}
		_ = json.Unmarshal(e.Payload, &p)
		if p.Seq == 1 && atomic.SwapInt64(&failFirst, 0) == 1 {
			return errors.New("transient failure on seq 1")
		}
		mu.Lock()
		order = append(order, p.Seq)
		mu.Unlock()
		return nil
	})

	// Two events for the same aggregate, in order.
	for seq := 1; seq <= 2; seq++ {
		s := seq
		if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
			return w.Write(ctx, db, outbox.Event{Type: "requests.request.changed", Resource: res, Payload: map[string]any{"seq": s}})
		}); err != nil {
			t.Fatal(err)
		}
	}

	relay := outbox.NewRelay(h.Platform, h.TxM, reg, 10)
	// First pass: seq 1 fails (marked failed), seq 2 is NOT claimed (blocked by
	// the earlier undispatched event for the aggregate).
	if _, err := relay.DispatchOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	// Requeue the failed event immediately (cooldown 0) and dispatch until drained.
	for i := 0; i < 5; i++ {
		if err := relay.RequeueFailed(context.Background(), 0); err != nil {
			t.Fatal(err)
		}
		if _, err := relay.DispatchOnce(context.Background()); err != nil {
			t.Fatal(err)
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(order) != 2 || order[0] != 1 || order[1] != 2 {
		t.Fatalf("per-aggregate order violated: got %v, want [1 2]", order)
	}
}

// TestIntegrationOutboxDLQ proves a poison event dead-letters after max_attempts
// instead of retrying forever (review finding ARCH-54).
func TestIntegrationOutboxDLQ(t *testing.T) {
	h := testkit.NewDB(t)
	w := outbox.NewWriter(model.UUIDv7())
	tn := testkit.CreateTenant(t, h)

	reg := outbox.NewHandlerRegistry()
	reg.Subscribe("x.poison.event", "always.fails", func(ctx context.Context, db database.TenantDB, e outbox.DispatchedEvent) error {
		return errors.New("permanent failure")
	})
	if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		return w.Write(ctx, db, outbox.Event{Type: "x.poison.event"})
	}); err != nil {
		t.Fatal(err)
	}
	// Lower the ceiling so the test is quick.
	if _, err := h.Admin.Exec(context.Background(), `UPDATE events_outbox SET max_attempts = 3 WHERE tenant_id = $1`, tn.ID); err != nil {
		t.Fatal(err)
	}
	relay := outbox.NewRelay(h.Platform, h.TxM, reg, 10)
	for i := 0; i < 5; i++ {
		_ = relay.RequeueFailed(context.Background(), 0)
		if _, err := relay.DispatchOnce(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	var status string
	if err := h.Admin.QueryRow(context.Background(), `SELECT dispatch_status FROM events_outbox WHERE tenant_id = $1`, tn.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "dead" {
		t.Fatalf("poison event status = %q, want dead (DLQ)", status)
	}
}

// TestIntegrationOutboxTenantIsolation proves a module writing an event under
// tenant A cannot be read by tenant B through the runtime (RLS).
func TestIntegrationOutboxTenantIsolation(t *testing.T) {
	h := testkit.NewDB(t)
	w := outbox.NewWriter(model.UUIDv7())
	a := testkit.CreateTenant(t, h)
	b := testkit.CreateTenant(t, h)

	if err := h.TxM.WithTenant(testkit.TenantCtx(a.ID), func(ctx context.Context, db database.TenantDB) error {
		return w.Write(ctx, db, outbox.Event{Type: "x.y.z"})
	}); err != nil {
		t.Fatal(err)
	}
	// Tenant B sees zero of A's events through the RLS-enforced runtime.
	var n int
	if err := h.TxM.WithTenantRO(testkit.TenantCtx(b.ID), func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx, `SELECT count(*) FROM events_outbox`).Scan(&n)
	}); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("tenant B saw %d of tenant A's events (RLS leak)", n)
	}
}
