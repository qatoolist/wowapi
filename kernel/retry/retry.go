package retry

import (
	"time"

	"github.com/cenkalti/backoff/v5"
)

// Schedule maps attempt numbers (1-based) to backoff durations using a
// cenkalti/backoff/v5 BackOff. It lets call sites that schedule future work
// (rather than retry synchronously) still express their schedule through the
// shared library.
type Schedule struct {
	bo backoff.BackOff
}

// NewSchedule builds a Schedule from a cenkalti/backoff/v5 BackOff. The BackOff
// is Reset before each Next call, so attempt numbers are stateless.
func NewSchedule(bo backoff.BackOff) *Schedule {
	return &Schedule{bo: bo}
}

// Next returns the backoff duration for the given attempt number. Attempts
// below 1 are clamped to 1. Attempts beyond the configured schedule return the
// final value, matching the prior hand-rolled "clamp to last" behavior.
func (s *Schedule) Next(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
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
