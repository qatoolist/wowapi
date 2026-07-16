package outbox

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type requeueMetrics struct {
	mu       sync.Mutex
	counters map[string]float64
}

func (*requeueMetrics) ObserveRequest(string, string, int, time.Duration, int) {}
func (*requeueMetrics) SetGauge(string, float64, map[string]string)            {}
func (m *requeueMetrics) IncCounter(name string, v float64, _ map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.counters == nil {
		m.counters = map[string]float64{}
	}
	m.counters[name] += v
}

func (m *requeueMetrics) counter(name string) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.counters[name]
}

// F-07 regression (adversarial-framework-review-2026-07-17): a persistently
// failing RequeueFailed must be OBSERVABLE — each failure increments
// outbox_requeue_errors_total, and after bounded consecutive failures Run
// returns the error instead of silently leaving failed events stuck forever.
func TestRelayRunSurfacesRequeueFailures(t *testing.T) {
	metrics := &requeueMetrics{}
	relay := NewRelay(nil, nil, NewHandlerRegistry(), 10, WithRelayMetrics(metrics))
	relay.dispatchFn = func(context.Context) (int, error) { return 0, nil }

	sentinel := errors.New("requeue permission denied")
	var calls atomic.Int64
	relay.requeue = func(ctx context.Context, cooldown time.Duration) error {
		calls.Add(1)
		return sentinel
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err := relay.Run(ctx, 5*time.Millisecond)
	if err == nil {
		t.Fatal("Run returned nil despite persistent requeue failures — recovery errors are still swallowed")
	}
	if !errors.Is(err, sentinel) {
		t.Fatalf("Run error %v does not wrap the requeue failure %v", err, sentinel)
	}
	if got := calls.Load(); got != relayRequeueMaxConsecutiveFailures {
		t.Fatalf("requeue attempted %d times, want the bounded %d before surfacing", got, relayRequeueMaxConsecutiveFailures)
	}
	if got := metrics.counter("outbox_requeue_errors_total"); got != float64(relayRequeueMaxConsecutiveFailures) {
		t.Fatalf("outbox_requeue_errors_total = %v, want %d (every failure must be counted)", got, relayRequeueMaxConsecutiveFailures)
	}
}

// F-07: a transient requeue failure must not kill the relay — the consecutive
// counter resets on success, keeping the loop resilient while staying visible.
func TestRelayRunRequeueRecoversAfterTransientFailure(t *testing.T) {
	metrics := &requeueMetrics{}
	relay := NewRelay(nil, nil, NewHandlerRegistry(), 10, WithRelayMetrics(metrics))
	relay.dispatchFn = func(context.Context) (int, error) { return 0, nil }

	var calls atomic.Int64
	relay.requeue = func(ctx context.Context, cooldown time.Duration) error {
		if calls.Add(1)%2 == 1 {
			return errors.New("transient")
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	if err := relay.Run(ctx, 5*time.Millisecond); err != nil {
		t.Fatalf("Run returned %v; alternating transient failures must not exceed the consecutive bound", err)
	}
	if metrics.counter("outbox_requeue_errors_total") == 0 {
		t.Fatal("transient requeue failures were not counted")
	}
}
