package chaos

import (
	"context"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/testkit"
)

// TestDuplicateWorkerLeaseExpiry is the named chaos test from DATA-02 T7. It
// proves that when worker A stalls past its lease, worker B can reclaim and
// complete the job, and A's late writes are rejected at every named boundary
// (domain, external, finalize) so exactly one logical effect is recorded.
func TestDuplicateWorkerLeaseExpiry(t *testing.T) {
	h := testkit.NewDB(t)
	domainStore := NewInMemoryStore()
	externalStore := NewInMemoryStore()

	reg := jobs.NewRegistry()
	harness := NewHarness(Config{
		T:              t,
		H:              h,
		Registry:       reg,
		DomainStore:    domainStore,
		ExternalStore:  externalStore,
		LeaseExpiry:    time.Minute,
		ReclaimTimeout: time.Minute,
	})

	attempts := harness.Run(context.Background())

	AssertExactlyOnce(t, attempts)

	if got := domainStore.Count(); got != 1 {
		t.Fatalf("domain effects = %d, want 1", got)
	}
	if got := externalStore.Count(); got != 1 {
		t.Fatalf("external effects = %d, want 1", got)
	}

	var status string
	if err := harness.Pool().QueryRow(context.Background(),
		`SELECT status FROM jobs_queue WHERE id = $1`, harness.JobID()).Scan(&status); err != nil {
		t.Fatalf("read job status: %v", err)
	}
	if status != "completed" {
		t.Fatalf("job status = %q, want completed", status)
	}
}
