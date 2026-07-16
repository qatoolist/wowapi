package retry

import (
	"sync"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// Schedule maps attempt numbers (1-based) to backoff durations using a
// cenkalti/backoff/v5 BackOff. It lets call sites that schedule future work
// (rather than retry synchronously) still express their schedule through the
// shared library. A Schedule is safe for concurrent use: delivery workers share
// the package-level schedules in notify/webhook, so the reset-and-iterate over
// the mutable BackOff must be atomic (adversarial review 2026-07-17, F-01 — an
// unguarded interleaving returned another caller's position and could drive a
// SequenceBackOff index past its slice, panicking the worker).
type Schedule struct {
	mu sync.Mutex
	bo backoff.BackOff
}

// NewSchedule builds a Schedule from a cenkalti/backoff/v5 BackOff. The BackOff
// is Reset before each Next call, so attempt numbers are stateless. A nil
// BackOff is caller misuse and is rejected immediately rather than deferred to
// a nil dereference on the first delivery retry.
func NewSchedule(bo backoff.BackOff) *Schedule {
	if bo == nil {
		panic("retry.NewSchedule: nil BackOff")
	}
	return &Schedule{bo: bo}
}

// Next returns the backoff duration for the given attempt number. Attempts
// below 1 are clamped to 1. Attempts beyond the configured schedule return the
// final value, matching the prior hand-rolled "clamp to last" behavior.
func (s *Schedule) Next(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bo.Reset()
	var d time.Duration
	for i := 1; i <= attempt; i++ {
		d = s.bo.NextBackOff()
		if d == backoff.Stop {
			return 0
		}
	}
	return d
}

// SequenceBackOff is a cenkalti/backoff/v5 BackOff that returns a fixed
// sequence of durations, clamping to the last value. It preserves exact
// schedule parity when migrating hand-rolled backoff tables.
type SequenceBackOff struct {
	values []time.Duration
	idx    int
}

// NewSequenceBackOff builds a BackOff over the given durations.
func NewSequenceBackOff(values ...time.Duration) *SequenceBackOff {
	return &SequenceBackOff{values: values}
}

// Reset implements backoff.BackOff.
func (s *SequenceBackOff) Reset() { s.idx = 0 }

// NextBackOff implements backoff.BackOff.
func (s *SequenceBackOff) NextBackOff() time.Duration {
	if len(s.values) == 0 {
		return backoff.Stop
	}
	if s.idx >= len(s.values) {
		s.idx = len(s.values) - 1
	}
	d := s.values[s.idx]
	s.idx++
	return d
}

// ExponentialBackOff is a convenience wrapper around
// backoff.NewExponentialBackOff that returns the configured BackOff.
func ExponentialBackOff() backoff.BackOff {
	return backoff.NewExponentialBackOff()
}
