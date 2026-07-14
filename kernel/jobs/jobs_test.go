package jobs_test

import (
	"context"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/jobs"
)

// sampleJob is a payload struct that implements jobs.Job.
type sampleJob struct {
	To string `json:"to"`
}

func (sampleJob) Kind() string { return "notify.email.send" }

func TestJobKind(t *testing.T) {
	var j jobs.Job = sampleJob{To: "a@b.test"}
	if got := j.Kind(); got != "notify.email.send" {
		t.Fatalf("Kind() = %q, want %q", got, "notify.email.send")
	}
}

func TestDefaultRetry(t *testing.T) {
	rp := jobs.DefaultRetry()
	if rp.MaxAttempts != 5 {
		t.Errorf("DefaultRetry MaxAttempts = %d, want 5", rp.MaxAttempts)
	}
	if rp.Backoff == nil {
		t.Fatal("DefaultRetry Backoff is nil")
	}
	// The default backoff is ExpJitterBackoff: first attempt ~1s (>= base).
	if d := rp.Backoff(1); d < time.Second || d > time.Second+time.Second/4 {
		t.Errorf("DefaultRetry Backoff(1) = %v, want [1s, 1.25s]", d)
	}
}

func TestExpJitterBackoffCappedAndMonotonic(t *testing.T) {
	const cap = 5 * time.Minute
	prev := time.Duration(-1)
	for attempt := 1; attempt <= 25; attempt++ {
		d := jobs.ExpJitterBackoff(attempt)
		if d < 0 {
			t.Fatalf("attempt %d: negative backoff %v", attempt, d)
		}
		if d > cap {
			t.Errorf("attempt %d: backoff %v exceeds cap %v", attempt, d, cap)
		}
		if d < prev {
			t.Errorf("attempt %d: backoff %v decreased from %v (want non-decreasing)", attempt, d, prev)
		}
		prev = d
	}
	// Deep attempts saturate at the cap exactly.
	if d := jobs.ExpJitterBackoff(40); d != cap {
		t.Errorf("ExpJitterBackoff(40) = %v, want cap %v", d, cap)
	}
}

func TestExpJitterBackoffDeterministic(t *testing.T) {
	// Pure function of attempt: same input, same output (no rand/time).
	for _, a := range []int{1, 3, 7, 0, -5} {
		first := jobs.ExpJitterBackoff(a)
		second := jobs.ExpJitterBackoff(a)
		if first != second {
			t.Errorf("ExpJitterBackoff(%d) is not deterministic: %v vs %v", a, first, second)
		}
	}
	// attempt < 1 is clamped to attempt 1.
	if jobs.ExpJitterBackoff(0) != jobs.ExpJitterBackoff(1) {
		t.Error("ExpJitterBackoff(0) should equal ExpJitterBackoff(1)")
	}
}

// noopWorker satisfies jobs.Worker for registry tests.
func noopWorker(context.Context, database.TenantDB, []byte) error { return nil }

func TestRegistryDuplicateKind(t *testing.T) {
	r := jobs.NewRegistry()
	r.RegisterKindWithIdempotency("k1", noopWorker, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.DefaultRetry())
	if err := r.Err(); err != nil {
		t.Fatalf("single registration errored: %v", err)
	}
	r.RegisterKindWithIdempotency("k1", noopWorker, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.DefaultRetry()) // duplicate
	if err := r.Err(); err == nil {
		t.Fatal("duplicate kind did not produce an error")
	}
}

func TestRegistryInvalidRegistrations(t *testing.T) {
	r := jobs.NewRegistry()
	r.RegisterKindWithIdempotency("", noopWorker, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.DefaultRetry()) // empty kind
	r.RegisterKindWithIdempotency("k2", nil, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.DefaultRetry())      // nil worker
	if err := r.Err(); err == nil {
		t.Fatal("invalid registrations did not produce an error")
	}
}

func TestRegistryCleanErr(t *testing.T) {
	r := jobs.NewRegistry()
	r.RegisterKindWithIdempotency("a", noopWorker, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.RetryPolicy{}) // zero policy is filled
	r.RegisterKindWithIdempotency("b", noopWorker, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.DefaultRetry())
	if err := r.Err(); err != nil {
		t.Fatalf("valid registrations errored: %v", err)
	}
}

func TestRegistryRejectsMissingIdempotency(t *testing.T) {
	r := jobs.NewRegistry()
	r.RegisterKindWithIdempotency("no-idem", noopWorker, jobs.Idempotency{}, jobs.DefaultRetry()) // zero value = no valid mechanism
	if err := r.Err(); err == nil {
		t.Fatal("worker without idempotency declaration should be rejected")
	}
}

func TestRegistryRejectsEffectLedgerWithoutEffectName(t *testing.T) {
	r := jobs.NewRegistry()
	r.RegisterKindWithIdempotency("no-effect-name", noopWorker, jobs.Idempotency{Kind: jobs.IdempotencyEffectLedger}, jobs.DefaultRetry())
	if err := r.Err(); err == nil {
		t.Fatal("effect-ledger idempotency without EffectName should be rejected")
	}
}

func TestRegistryAcceptsEachIdempotencyKind(t *testing.T) {
	r := jobs.NewRegistry()
	r.RegisterKindWithIdempotency("ledger", noopWorker, jobs.Idempotency{Kind: jobs.IdempotencyEffectLedger, EffectName: "test.effect"}, jobs.DefaultRetry())
	r.RegisterKindWithIdempotency("cas", noopWorker, jobs.Idempotency{Kind: jobs.IdempotencyDomainCAS}, jobs.DefaultRetry())
	r.RegisterKindWithIdempotency("provider", noopWorker, jobs.Idempotency{Kind: jobs.IdempotencyProviderKey}, jobs.DefaultRetry())
	if err := r.Err(); err != nil {
		t.Fatalf("valid idempotency declarations errored: %v", err)
	}
}
