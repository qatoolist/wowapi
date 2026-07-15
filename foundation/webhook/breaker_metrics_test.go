package webhook

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestBreakerStateValueTransitions covers the metric-friendly state accessor
// added for the webhook_breaker_state gauge (roadmap CA-1): 0=closed, 1=open,
// 2=half-open.
func TestBreakerStateValueTransitions(t *testing.T) {
	now := time.Now()
	b := &breakerState{}

	if v := b.stateValue(now); v != 0 {
		t.Fatalf("fresh breaker should be closed (0), got %v", v)
	}
	for i := 0; i < BreakerFailureThreshold; i++ {
		b.recordFailure(now)
	}
	if v := b.stateValue(now); v != 1 {
		t.Fatalf("breaker should be open (1) after %d failures, got %v", BreakerFailureThreshold, v)
	}
	if v := b.stateValue(now.Add(BreakerCooldown)); v != 2 {
		t.Fatalf("breaker should be half-open (2) after cooldown, got %v", v)
	}
	b.recordSuccess()
	if v := b.stateValue(now.Add(BreakerCooldown)); v != 0 {
		t.Fatalf("breaker should be closed (0) after success, got %v", v)
	}
}

// fakeMetrics records the last SetGauge value per name.
type fakeMetrics struct{ gauges map[string]float64 }

func (f *fakeMetrics) ObserveRequest(_, _ string, _ int, _ time.Duration, _ int) {}
func (f *fakeMetrics) IncCounter(_ string, _ float64, _ map[string]string)       {}
func (f *fakeMetrics) ObserveHistogram(_ string, _ float64, _ map[string]string) {}
func (f *fakeMetrics) SetGauge(name string, v float64, _ map[string]string)      { f.gauges[name] = v }

// TestEmitBreakerState proves the service actually emits the gauge with the
// current breaker state (was: zero emission sites).
func TestEmitBreakerState(t *testing.T) {
	fm := &fakeMetrics{gauges: map[string]float64{}}
	s := &Service{metrics: fm, now: time.Now}

	b := &breakerState{}
	for i := 0; i < BreakerFailureThreshold; i++ {
		b.recordFailure(s.now())
	}
	s.emitBreakerState(uuid.New(), b)

	if got := fm.gauges["webhook_breaker_state"]; got != 1 {
		t.Fatalf("expected webhook_breaker_state=1 (open), got %v", got)
	}
}
