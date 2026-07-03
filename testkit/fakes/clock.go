// Package fakes holds the deterministic test doubles wowapi injects through the
// same constructors production uses (08 §2): a manual-advance clock and a
// deterministic IDGen. No test hooks live in production code — tests supply
// these where production supplies real implementations.
package fakes

import (
	"sync"
	"time"
)

// Clock is a manual-advance fake time source. It satisfies the ambient
// interface{ Now() time.Time } the kernel expects, but never moves on its own —
// tests call Advance to make time pass deterministically. Safe for concurrent use.
type Clock struct {
	mu  sync.Mutex
	now time.Time
}

// NewClock returns a Clock pinned at start.
func NewClock(start time.Time) *Clock { return &Clock{now: start} }

// Now returns the current fake time.
func (c *Clock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

// Advance moves the clock forward by d.
func (c *Clock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}
