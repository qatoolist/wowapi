package webhook

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// breakerState is the in-memory per-endpoint circuit breaker.
//
// States:
//   - closed  (openedAt.IsZero()): normal delivery; failures accumulate.
//   - open    (!openedAt.IsZero() && elapsed < BreakerCooldown): deliveries skipped.
//   - half-open (elapsed >= BreakerCooldown): one probe attempt allowed per
//     cooldown window; success → closed, failure → back to open.
type breakerState struct {
	mu          sync.Mutex
	failures    int       // consecutive failures while closed
	openedAt    time.Time // when the breaker last opened (zero = closed)
	lastProbeAt time.Time // when we last allowed a half-open probe
}

// allow reports whether a delivery attempt is permitted at now.
func (b *breakerState) allow(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.openedAt.IsZero() {
		return true // closed
	}
	if now.Sub(b.openedAt) >= BreakerCooldown {
		// Half-open: allow one probe per cooldown window.
		if b.lastProbeAt.IsZero() || now.Sub(b.lastProbeAt) >= BreakerCooldown {
			b.lastProbeAt = now
			return true
		}
	}
	return false // open
}

// recordSuccess closes the breaker and resets the failure counter.
func (b *breakerState) recordSuccess() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures = 0
	b.openedAt = time.Time{}
	b.lastProbeAt = time.Time{}
}

// recordFailure increments the failure counter and opens the breaker when the
// threshold is reached. now is the caller's clock value (injectable for tests).
func (b *breakerState) recordFailure(now time.Time) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failures++
	if b.openedAt.IsZero() && b.failures >= BreakerFailureThreshold {
		b.openedAt = now
	}
}

// stateValue reports the breaker state as a metric-friendly number at now:
// 0 = closed, 1 = open (blocking), 2 = half-open (probing).
func (b *breakerState) stateValue(now time.Time) float64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.openedAt.IsZero() {
		return 0
	}
	if now.Sub(b.openedAt) >= BreakerCooldown {
		return 2 // half-open
	}
	return 1 // open
}

// isOpen reports whether the breaker is in the blocking (open) state at now.
func (b *breakerState) isOpen(now time.Time) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.openedAt.IsZero() {
		return false
	}
	return now.Sub(b.openedAt) < BreakerCooldown
}

// breakerRegistry holds per-endpoint breaker state (in-memory, per-process).
type breakerRegistry struct {
	mu    sync.Mutex
	state map[uuid.UUID]*breakerState
}

func newBreakerRegistry() *breakerRegistry {
	return &breakerRegistry{state: make(map[uuid.UUID]*breakerState)}
}

// get returns the breakerState for endpointID, creating it on first access.
func (r *breakerRegistry) get(endpointID uuid.UUID) *breakerState {
	r.mu.Lock()
	defer r.mu.Unlock()
	if b, ok := r.state[endpointID]; ok {
		return b
	}
	b := &breakerState{}
	r.state[endpointID] = b
	return b
}
