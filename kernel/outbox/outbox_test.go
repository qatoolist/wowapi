package outbox_test

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/observability"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/testkit"
)

// fakeTracer records the carrier passed to Extract and injects a fixed
// traceparent, so a test can prove the outbox carries trace context across the
// async boundary (CA-9).
type fakeTracer struct {
	inject    string
	mu        sync.Mutex
	extracted []string
}

func (f *fakeTracer) StartSpan(ctx context.Context, _ string) (context.Context, observability.Span) {
	return ctx, fakeSpan{}
}
func (f *fakeTracer) Inject(context.Context) string { return f.inject }
func (f *fakeTracer) Extract(ctx context.Context, carrier string) context.Context {
	f.mu.Lock()
	f.extracted = append(f.extracted, carrier)
	f.mu.Unlock()
	return ctx
}

type fakeSpan struct{}

func (fakeSpan) End()                {}
func (fakeSpan) SetAttr(_, _ string) {}
func (fakeSpan) RecordError(error)   {}

// TestIntegrationOutboxTracePropagation is the O1/CA-9 regression: an event
// captures the writer's trace context, and the relay extracts it when it
// dispatches — so an async handler continues the originating request's trace.
func TestIntegrationOutboxTracePropagation(t *testing.T) {
	h := testkit.NewDB(t)
	const carrier = "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"
	tr := &fakeTracer{inject: carrier}
	w := outbox.NewWriter(model.UUIDv7(), outbox.WithWriterTracer(tr))
	tn := testkit.CreateTenant(t, h)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	var dispatched int64
	reg := outbox.NewHandlerRegistry()
	reg.Subscribe("requests.request.changed", "h",
		func(context.Context, database.TenantDB, outbox.DispatchedEvent) error {
			atomic.AddInt64(&dispatched, 1)
			return nil
		})

	if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
		return w.Write(ctx, db, outbox.Event{Type: "requests.request.changed", Resource: res})
	}); err != nil {
		t.Fatal(err)
	}

	// The event row stored the injected trace context.
	var stored string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT coalesce(trace_context,'') FROM events_outbox WHERE tenant_id = $1`, tn.ID).Scan(&stored); err != nil {
		t.Fatalf("read trace_context: %v", err)
	}
	if stored != carrier {
		t.Fatalf("stored trace_context = %q, want the injected carrier", stored)
	}

	// The relay extracts that carrier when dispatching (continuing the trace).
	relay := outbox.NewRelay(h.Platform, h.TxM, reg, 10, outbox.WithRelayTracer(tr))
	if _, err := relay.DispatchOnce(context.Background()); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if atomic.LoadInt64(&dispatched) != 1 {
		t.Fatalf("expected 1 dispatch, got %d", dispatched)
	}
	tr.mu.Lock()
	defer tr.mu.Unlock()
	found := false
	for _, c := range tr.extracted {
		if c == carrier {
			found = true
		}
	}
	if !found {
		t.Fatalf("relay must Extract the stored trace context to continue the trace; extracted=%v", tr.extracted)
	}
}

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
// TestIntegrationOutboxHotAggregateThroughput is the R2/CA-4 load
// characterization: it emits a burst of events onto ONE hot aggregate — the
// worst case for the per-aggregate advisory lock (relay.go pg_advisory_xact_lock)
// that serializes dispatch to preserve ordering — drains them with several
// concurrent relay workers, and reports the observed throughput envelope. The
// measured number feeds docs/operations/load-characterization.md; the assertion
// is on correctness (every event dispatched exactly once) so it is not flaky.
func TestIntegrationOutboxHotAggregateThroughput(t *testing.T) {
	h := testkit.NewDB(t)
	w := outbox.NewWriter(model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	res := testkit.CreateResourceTypeAndResource(t, h, tn.ID, "requests.request")

	const total = 200
	var dispatched int64
	reg := outbox.NewHandlerRegistry()
	reg.Subscribe("requests.request.changed", "count",
		func(context.Context, database.TenantDB, outbox.DispatchedEvent) error {
			atomic.AddInt64(&dispatched, 1)
			return nil
		})

	for i := 0; i < total; i++ {
		i := i
		if err := h.TxM.WithTenant(testkit.TenantCtx(tn.ID), func(ctx context.Context, db database.TenantDB) error {
			return w.Write(ctx, db, outbox.Event{Type: "requests.request.changed", Resource: res, Payload: map[string]any{"i": i}})
		}); err != nil {
			t.Fatal(err)
		}
	}

	relay := outbox.NewRelay(h.Platform, h.TxM, reg, 20)
	const workers = 4
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(workers)
	for wkr := 0; wkr < workers; wkr++ {
		go func() {
			defer wg.Done()
			for atomic.LoadInt64(&dispatched) < total && ctx.Err() == nil {
				n, err := relay.DispatchOnce(ctx)
				if err != nil {
					t.Errorf("dispatch: %v", err)
					return
				}
				if n == 0 {
					time.Sleep(time.Millisecond) // contended or momentarily drained
				}
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(start)

	got := atomic.LoadInt64(&dispatched)
	if got != total {
		t.Fatalf("dispatched %d events, want %d exactly once (advisory-lock serialization stalled?)", got, total)
	}
	t.Logf("R2 hot-aggregate throughput: %d events drained in %v = %.0f events/sec (%d relay workers, single hot aggregate, advisory-lock serialized)",
		total, elapsed.Round(time.Millisecond), float64(total)/elapsed.Seconds(), workers)
}

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
