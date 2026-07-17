package outbox_test

import (
	"context"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/outbox"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// closedPool builds a real pool against the test DSN and immediately closes it,
// so every subsequent call on it returns the pgxpool "closed pool" error. This
// drives the DB-failure branches (begin/query/exec error) of the relay and DLQ
// helpers deterministically, without racing a live connection.
func closedPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	dsn := os.Getenv("WOWAPI_TEST_DSN")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		if os.Getenv("WOWAPI_REQUIRE_DB") != "" {
			t.Fatal("WOWAPI_REQUIRE_DB is set but neither WOWAPI_TEST_DSN nor DATABASE_URL is available for closed-pool integration coverage")
		}
		t.Skip("no DSN for closed-pool error injection")
	}
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("build pool: %v", err)
	}
	pool.Close()
	return pool
}

// TestIntegrationOutboxWriteValidation covers Write's two pre-DB guards: an
// empty Type and a payload that is not JSON-encodable both fail as
// KindInternal without inserting a row.
func TestIntegrationOutboxWriteValidation(t *testing.T) {
	h := testkit.NewDB(t)
	w := outbox.NewWriter(model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	var emptyType, badPayload error
	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		emptyType = w.Write(ctx, db, outbox.Event{Type: ""})
		// A channel cannot be marshalled to JSON — the marshal guard fires.
		badPayload = w.Write(ctx, db, outbox.Event{Type: "x.y.z", Payload: make(chan int)})
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if kerr.KindOf(emptyType) != kerr.KindInternal {
		t.Fatalf("empty Type: got kind %v, want internal (err=%v)", kerr.KindOf(emptyType), emptyType)
	}
	if kerr.KindOf(badPayload) != kerr.KindInternal {
		t.Fatalf("bad payload: got kind %v, want internal (err=%v)", kerr.KindOf(badPayload), badPayload)
	}
	// Neither invalid write left a row behind.
	var n int
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT count(*) FROM events_outbox WHERE tenant_id=$1`, tn.ID).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("invalid Write inserted %d rows, want 0", n)
	}
}

// TestHandlerRegistryErrAggregation covers the Subscribe validation/dedup error
// branches and Err(): a clean registry returns nil; a registry that accumulated
// multiple bad subscriptions joins them into one subscription_failed error.
func TestHandlerRegistryErrAggregation(t *testing.T) {
	// Clean registry: no errors.
	clean := outbox.NewHandlerRegistry()
	clean.Subscribe("a.b.c", "h1", func(context.Context, database.TenantDB, outbox.DispatchedEvent) error { return nil })
	if err := clean.Err(); err != nil {
		t.Fatalf("clean registry Err() = %v, want nil", err)
	}

	reg := outbox.NewHandlerRegistry()
	ok := func(context.Context, database.TenantDB, outbox.DispatchedEvent) error { return nil }
	// Invalid: empty eventType, empty handlerName, nil fn — each records an error.
	reg.Subscribe("", "h", ok)
	reg.Subscribe("a.b.c", "", ok)
	reg.Subscribe("a.b.c", "h", nil)
	// Valid, then a duplicate of it — the duplicate records an error.
	reg.Subscribe("a.b.c", "dup", ok)
	reg.Subscribe("a.b.c", "dup", ok)

	err := reg.Err()
	if err == nil {
		t.Fatal("Err() = nil, want aggregated subscription failure")
	}
	if kerr.KindOf(err) != kerr.KindInternal {
		t.Fatalf("Err() kind = %v, want internal", kerr.KindOf(err))
	}
	msg := err.Error()
	for _, want := range []string{"Subscribe requires", "subscribed to a.b.c more than once", "; "} {
		if !strings.Contains(msg, want) {
			t.Fatalf("aggregated Err() = %q, missing %q", msg, want)
		}
	}
}

// TestIntegrationOutboxDispatchNoSubscribers covers the len(subs)==0 fast path:
// an event with no registered handler is a dispatch no-op that still transitions
// to 'dispatched'. Uses batchSize 0 to exercise NewRelay's default (→100).
func TestIntegrationOutboxDispatchNoSubscribers(t *testing.T) {
	h := testkit.NewDB(t)
	w := outbox.NewWriter(model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	ctx := testkit.TenantCtx(tn.ID)

	if err := h.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return w.Write(ctx, db, outbox.Event{Type: "unsubscribed.event.happened"})
	}); err != nil {
		t.Fatal(err)
	}

	reg := outbox.NewHandlerRegistry() // no subscribers
	relay := outbox.NewRelay(h.Platform, h.TxM, reg, 0)
	n, err := relay.DispatchOnce(context.Background())
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if n != 1 {
		t.Fatalf("processed %d events, want 1", n)
	}
	var status string
	if err := h.Admin.QueryRow(context.Background(),
		`SELECT dispatch_status FROM events_outbox WHERE tenant_id=$1`, tn.ID).Scan(&status); err != nil {
		t.Fatal(err)
	}
	if status != "dispatched" {
		t.Fatalf("no-subscriber event status = %q, want dispatched", status)
	}
}

// TestIntegrationOutboxRelayRun drives the Run loop end to end: it drains a
// batch (n>0 continue), then idles on the poll ticker (which calls
// RequeueFailed) until the context is cancelled, at which point Run returns nil.
func TestIntegrationOutboxRelayRun(t *testing.T) {
	h := testkit.NewDB(t)
	w := outbox.NewWriter(model.UUIDv7())
	tn := testkit.CreateTenant(t, h)
	octx := testkit.TenantCtx(tn.ID)

	var dispatched int64
	reg := outbox.NewHandlerRegistry()
	reg.Subscribe("run.loop.event", "h",
		func(context.Context, database.TenantDB, outbox.DispatchedEvent) error {
			atomic.AddInt64(&dispatched, 1)
			return nil
		})

	const total = 3
	for i := 0; i < total; i++ {
		if err := h.TxM.WithTenant(octx, func(ctx context.Context, db database.TenantDB) error {
			return w.Write(ctx, db, outbox.Event{Type: "run.loop.event", Payload: map[string]any{"i": i}})
		}); err != nil {
			t.Fatal(err)
		}
	}

	relay := outbox.NewRelay(h.Platform, h.TxM, reg, 10)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- relay.Run(ctx, 5*time.Millisecond) }()

	// Wait for all events to drain, giving the poll ticker time to fire
	// (RequeueFailed) at least once while idle.
	deadline := time.After(10 * time.Second)
	for atomic.LoadInt64(&dispatched) < total {
		select {
		case <-deadline:
			cancel()
			t.Fatalf("only %d/%d dispatched before timeout", atomic.LoadInt64(&dispatched), total)
		case <-time.After(5 * time.Millisecond):
		}
	}
	time.Sleep(30 * time.Millisecond) // let the idle ticker path (RequeueFailed) run
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v, want nil on cancellation", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
	if got := atomic.LoadInt64(&dispatched); got != total {
		t.Fatalf("dispatched %d, want %d", got, total)
	}
}

// TestIntegrationOutboxRelayRunDefaultPollAndCancel covers Run's poll<=0 default
// (→1s) and the ctx.Done() return arm with no work pending.
func TestIntegrationOutboxRelayRunDefaultPollAndCancel(t *testing.T) {
	h := testkit.NewDB(t)
	reg := outbox.NewHandlerRegistry()
	relay := outbox.NewRelay(h.Platform, h.TxM, reg, 10)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- relay.Run(ctx, 0) }() // poll<=0 → default 1s
	time.Sleep(50 * time.Millisecond)         // ensure it reached the idle select
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned %v, want nil", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
}

// TestOutboxRelayPoolFailures covers the DB-failure branches of DispatchOnce
// (begin), RequeueFailed, and Run (DispatchOnce error with a live ctx surfaces
// the error) using a closed pool.
func TestOutboxRelayPoolFailures(t *testing.T) {
	h := testkit.NewDB(t)
	reg := outbox.NewHandlerRegistry()
	relay := outbox.NewRelay(closedPool(t), h.TxM, reg, 10)

	if _, err := relay.DispatchOnce(context.Background()); err == nil {
		t.Fatal("DispatchOnce on closed pool = nil error, want begin failure")
	}
	if err := relay.RequeueFailed(context.Background(), time.Second); err == nil {
		t.Fatal("RequeueFailed on closed pool = nil error, want failure")
	}
	// Run with a non-cancelled context: DispatchOnce errors and ctx.Err() is nil,
	// so Run surfaces the error instead of returning nil.
	if err := relay.Run(context.Background(), time.Second); err == nil {
		t.Fatal("Run on closed pool = nil error, want surfaced DispatchOnce failure")
	}
}

// TestNilTracerOptionsAreNoOps covers the nil-guard false arm of the tracer
// options: passing nil must not override the default NoOpTracer.
func TestNilTracerOptionsAreNoOps(t *testing.T) {
	// Construction must not panic and returns usable values.
	if w := outbox.NewWriter(model.UUIDv7(), outbox.WithWriterTracer(nil)); w == nil {
		t.Fatal("NewWriter with nil tracer returned nil")
	}
	pool := closedPool(t)
	if r := outbox.NewRelay(pool, nil, outbox.NewHandlerRegistry(), 10, outbox.WithRelayTracer(nil)); r == nil {
		t.Fatal("NewRelay with nil tracer returned nil")
	}
}
