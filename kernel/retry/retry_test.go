package retry

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v5"
)

func TestScheduleSequenceParity(t *testing.T) {
	sched := NewSchedule(NewSequenceBackOff(
		time.Second,
		5*time.Second,
		30*time.Second,
	))
	cases := []struct {
		attempt int
		want    time.Duration
	}{
		{0, time.Second}, // clamp below 1
		{1, time.Second},
		{2, 5 * time.Second},
		{3, 30 * time.Second},
		{4, 30 * time.Second}, // clamp to last
	}
	for _, tc := range cases {
		if got := sched.Next(tc.attempt); got != tc.want {
			t.Errorf("Next(%d) = %v, want %v", tc.attempt, got, tc.want)
		}
	}
}

func TestScheduleExponentialBackOff(t *testing.T) {
	bo := backoff.NewExponentialBackOff()
	bo.InitialInterval = time.Second
	bo.Multiplier = 2
	bo.MaxInterval = 10 * time.Second
	bo.RandomizationFactor = 0
	sched := NewSchedule(bo)
	if got := sched.Next(1); got != time.Second {
		t.Errorf("Next(1) = %v, want 1s", got)
	}
	if got := sched.Next(2); got != 2*time.Second {
		t.Errorf("Next(2) = %v, want 2s", got)
	}
	if got := sched.Next(5); got != 10*time.Second {
		t.Errorf("Next(5) = %v, want 10s (max interval cap)", got)
	}
	if got := sched.Next(10); got != 10*time.Second {
		t.Errorf("Next(10) = %v, want 10s (max interval cap)", got)
	}
}

func TestRetryFaultInjection(t *testing.T) {
	// Prove the shared library's Retry honours a max-tries bound under a
	// permanently failing operation.
	wantAttempts := 3
	gotAttempts := 0
	errFail := errors.New("injected transient failure")
	_, err := backoff.Retry(context.Background(), func() (int, error) {
		gotAttempts++
		return 0, errFail
	}, backoff.WithBackOff(backoff.NewConstantBackOff(time.Millisecond)),
		backoff.WithMaxTries(uint(wantAttempts)))
	if err == nil {
		t.Fatal("expected error from exhausted retries")
	}
	if gotAttempts != wantAttempts {
		t.Fatalf("attempts = %d, want %d", gotAttempts, wantAttempts)
	}
}

func TestRetryPermanentError(t *testing.T) {
	// A permanent error must not be retried.
	attempts := 0
	_, err := backoff.Retry(context.Background(), func() (int, error) {
		attempts++
		return 0, backoff.Permanent(errors.New("permanent"))
	}, backoff.WithBackOff(backoff.NewConstantBackOff(time.Millisecond)),
		backoff.WithMaxTries(10))
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("attempts = %d, want 1 for permanent error", attempts)
	}
}
