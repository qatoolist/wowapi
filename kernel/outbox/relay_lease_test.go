package outbox_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/testkit"
)

type relayMetrics struct {
	mu     sync.Mutex
	gauges map[string]float64
}

func (*relayMetrics) ObserveRequest(string, string, int, time.Duration, int) {}
func (*relayMetrics) IncCounter(string, float64, map[string]string)          {}
func (m *relayMetrics) ObserveHistogram(name string, value float64, labels map[string]string) {
	m.SetGauge(name, value, labels)
}

func (m *relayMetrics) SetGauge(name string, value float64, labels map[string]string) {
	m.mu.Lock()
	m.gauges[name+"/"+labels["worker"]] = value
	m.mu.Unlock()
}

func (m *relayMetrics) has(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.gauges[name]
	return ok
}

func writeRelayEvent(t *testing.T, h *testkit.DBHandle, eventType string) {
	t.Helper()
	tenant := testkit.CreateTenant(t, h)
	writer := outbox.NewWriter(model.UUIDv7())
	if err := h.TxM.WithTenant(testkit.TenantCtx(tenant.ID), func(ctx context.Context, db database.TenantDB) error {
		return writer.Write(ctx, db, outbox.Event{Type: eventType})
	}); err != nil {
		t.Fatal(err)
	}
}

func TestIntegrationOutboxClaimCommitsBeforeTenantHandler(t *testing.T) {
	h := testkit.NewDB(t)
	writeRelayEvent(t, h, "perf.claim.committed")
	started := make(chan struct{})
	release := make(chan struct{})
	defer func() {
		select {
		case <-release:
		default:
			close(release)
		}
	}()
	reg := outbox.NewHandlerRegistry()
	reg.Subscribe("perf.claim.committed", "blocking.handler", func(context.Context, database.TenantDB, outbox.DispatchedEvent) error {
		close(started)
		<-release
		return nil
	})
	relay := outbox.NewRelay(h.Platform, h.TxM, reg, 1)
	result := make(chan error, 1)
	go func() {
		_, err := relay.DispatchOnce(context.Background())
		result <- err
	}()
	<-started

	probeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	tx, err := h.Admin.Begin(probeCtx)
	if err != nil {
		t.Fatal(err)
	}
	var id string
	if err := tx.QueryRow(probeCtx, `SELECT id::text FROM events_outbox FOR UPDATE`).Scan(&id); err != nil {
		t.Fatalf("claim transaction still holds the outbox row while tenant handler runs: %v", err)
	}
	if err := tx.Rollback(context.Background()); err != nil {
		t.Fatal(err)
	}
	close(release)
	if err := <-result; err != nil {
		t.Fatalf("DispatchOnce: %v", err)
	}
}

func TestIntegrationOutboxLeaseExpiryFencesDuplicateWorker(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)
	resource := testkit.CreateResourceTypeAndResource(t, h, tenant.ID, "requests.request")
	writer := outbox.NewWriter(model.UUIDv7())
	if err := h.TxM.WithTenant(testkit.TenantCtx(tenant.ID), func(ctx context.Context, db database.TenantDB) error {
		return writer.Write(ctx, db, outbox.Event{Type: "perf.lease.expiry", Resource: resource})
	}); err != nil {
		t.Fatal(err)
	}

	started := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	reg := outbox.NewHandlerRegistry()
	reg.Subscribe("perf.lease.expiry", "effect.once", func(context.Context, database.TenantDB, outbox.DispatchedEvent) error {
		if calls.Add(1) == 1 {
			close(started)
			<-release
		}
		return nil
	})
	const ttl = 75 * time.Millisecond
	relayA := outbox.NewRelay(h.Platform, h.TxM, reg, 1, outbox.WithRelayLeaseTTL(ttl))
	relayB := outbox.NewRelay(h.Platform, h.TxM, reg, 1, outbox.WithRelayLeaseTTL(5*time.Second))
	type dispatchResult struct {
		n   int
		err error
	}
	results := make(chan dispatchResult, 2)
	go func() {
		n, err := relayA.DispatchOnce(context.Background())
		results <- dispatchResult{n: n, err: err}
	}()
	<-started
	time.Sleep(ttl + 25*time.Millisecond)
	go func() {
		n, err := relayB.DispatchOnce(context.Background())
		results <- dispatchResult{n: n, err: err}
	}()

	deadline := time.Now().Add(2 * time.Second)
	for {
		var generation int64
		if err := h.Admin.QueryRow(context.Background(), `SELECT lease_generation FROM events_outbox`).Scan(&generation); err != nil {
			t.Fatal(err)
		}
		if generation >= 2 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("second relay did not reclaim the expired lease")
		}
		time.Sleep(5 * time.Millisecond)
	}
	close(release)

	first, second := <-results, <-results
	var staleSeen, successSeen bool
	for _, result := range []dispatchResult{first, second} {
		if errors.Is(result.err, outbox.ErrLeaseMismatch) {
			staleSeen = true
			continue
		}
		if result.err == nil && result.n == 1 {
			successSeen = true
			continue
		}
		t.Fatalf("unexpected dispatch result: n=%d err=%v", result.n, result.err)
	}
	if !staleSeen || !successSeen {
		t.Fatalf("stale/success results = (%v,%v), want both true", staleSeen, successSeen)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("handler effects = %d, want 1 after duplicate-worker reclaim", got)
	}
	var status string
	var token *string
	if err := h.Admin.QueryRow(context.Background(), `SELECT dispatch_status,lease_token FROM events_outbox`).Scan(&status, &token); err != nil {
		t.Fatal(err)
	}
	if status != "dispatched" || token != nil {
		t.Fatalf("final state = (%s,%v), want dispatched with cleared lease", status, token)
	}
}

func TestIntegrationOutboxRelayMetrics(t *testing.T) {
	h := testkit.NewDB(t)
	writeRelayEvent(t, h, "perf.relay.metrics")
	if _, err := h.Admin.Exec(context.Background(), `UPDATE events_outbox SET occurred_at=now()-interval '1 hour'`); err != nil {
		t.Fatal(err)
	}
	metrics := &relayMetrics{gauges: map[string]float64{}}
	relay := outbox.NewRelay(h.Platform, h.TxM, outbox.NewHandlerRegistry(), 1, outbox.WithRelayMetrics(metrics))
	if _, err := relay.DispatchOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"worker_queue_lag_seconds/outbox_relay", "worker_batch_duration_seconds/outbox_relay"} {
		if !metrics.has(name) {
			t.Errorf("missing metric %s", name)
		}
	}
}
