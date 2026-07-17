package retry

import (
	"sync"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v5"
)

// F-01 regression (adversarial-framework-review-2026-07-17): a Schedule shared
// by concurrent delivery workers must return the exact per-attempt duration and
// must not race or panic. Before the fix, Next reset and advanced the same
// mutable BackOff from every goroutine: interleavings could return another
// caller's position or drive SequenceBackOff.idx past the slice.
func TestScheduleNextConcurrentExactDurations(t *testing.T) {
	values := []time.Duration{
		1 * time.Second, 2 * time.Second, 4 * time.Second, 8 * time.Second,
	}
	s := NewSchedule(NewSequenceBackOff(values...))

	const goroutines = 8
	const rounds = 200
	var wg sync.WaitGroup
	errs := make(chan string, goroutines*rounds)
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for r := 0; r < rounds; r++ {
				for attempt := 1; attempt <= len(values)+2; attempt++ {
					want := values[min(attempt, len(values))-1]
					if got := s.Next(attempt); got != want {
						errs <- "attempt " + time.Duration(attempt).String() +
							": got " + got.String() + ", want " + want.String()
						return
					}
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Fatalf("concurrent Next returned wrong duration: %s", e)
	}
}

// blockingBackOff raises the LIKELIHOOD of reset/advance interleavings (the
// class the review's isolated reproduction used to drive SequenceBackOff past
// its slice) by adding non-blocking channel polls between operations. It does
// not deterministically force a specific interleaving — the concurrent
// exact-duration test above plus -race are the load-bearing assertions; this
// test adds panic coverage under heavier goroutine churn.
type blockingBackOff struct {
	inner   *SequenceBackOff
	release chan struct{}
}

func (b *blockingBackOff) Reset() {
	b.inner.Reset()
	select {
	case <-b.release:
	default:
	}
}

func (b *blockingBackOff) NextBackOff() time.Duration {
	// Yield aggressively so interleavings actually occur without -race.
	for i := 0; i < 3; i++ {
		select {
		case <-b.release:
		default:
		}
	}
	return b.inner.NextBackOff()
}

func TestScheduleNextInterleavedNoPanic(t *testing.T) {
	bo := &blockingBackOff{
		inner:   NewSequenceBackOff(1*time.Second, 2*time.Second, 3*time.Second),
		release: make(chan struct{}),
	}
	s := NewSchedule(bo)
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Schedule.Next panicked under interleaving: %v", r)
		}
	}()
	var wg sync.WaitGroup
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 500; i++ {
				_ = s.Next(3)
				_ = s.Next(5)
			}
		}()
	}
	wg.Wait()
}

// F-01: a nil BackOff is caller misuse and must be rejected at construction,
// not deferred to a nil-dereference panic on first delivery retry.
func TestNewScheduleNilBackOffPanicsWithClearMessage(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("NewSchedule(nil) did not reject the nil BackOff")
		}
	}()
	_ = NewSchedule(nil)
}

var _ backoff.BackOff = (*blockingBackOff)(nil)
