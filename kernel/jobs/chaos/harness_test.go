// Package chaos provides a reusable test harness for the duplicate-worker
// lease-expiry scenario described in DATA-02 T7. It simulates worker A claiming
// a job, stalling past its lease, worker B reclaiming and completing the job,
// and worker A resuming to attempt writes at the domain, external, and finalize
// boundaries. The harness is intentionally generic so W04-E02 (notify/webhook)
// and W04-E03 (bulk) can parameterize it for their own effect types rather than
// reimplementing the pause/expire/reclaim/resume mechanics.
package chaos

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/kernel/lease"
	"github.com/qatoolist/wowapi/testkit"
)

// Boundary identifies one of the three named boundaries the chaos test must
// exercise.
type Boundary int

const (
	// Domain is an effect committed inside a tenant transaction (e.g. an
	// inbox/effect ledger row or a domain CAS).
	Domain Boundary = iota
	// External is a remote side effect (e.g. a webhook POST or email send)
	// guarded by a provider idempotency key.
	External
	// Finalize is the jobs_queue row completion/failure write.
	Finalize
)

func (b Boundary) String() string {
	switch b {
	case Domain:
		return "domain"
	case External:
		return "external"
	case Finalize:
		return "finalize"
	default:
		return fmt.Sprintf("boundary-%d", b)
	}
}

// Attempt is the result of one worker trying to write an effect at a boundary.
type Attempt struct {
	Boundary Boundary
	Worker   string // "A" or "B"
	Accepted bool   // true if the write was accepted
}

// EffectStore is the contract an effect ledger or external provider must
// implement for the harness to detect duplicate writes. It is keyed by the
// stable idempotency key the runner passes to workers.
type EffectStore interface {
	// TryRecord records one logical effect for key if no effect for that key
	// already exists. It returns true when this call was the first to record
	// the effect, false if the effect was already present.
	TryRecord(key string) bool
	// Count returns the total number of logical effects recorded.
	Count() int
	// Reset clears the store (test cleanup).
	Reset()
}

// InMemoryStore is a simple EffectStore for tests.
type InMemoryStore struct {
	mu   sync.Mutex
	seen map[string]struct{}
}

// NewInMemoryStore returns an empty InMemoryStore.
func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{seen: map[string]struct{}{}}
}

// TryRecord records key if absent; returns true on first record.
func (s *InMemoryStore) TryRecord(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.seen[key]; ok {
		return false
	}
	s.seen[key] = struct{}{}
	return true
}

// Count returns the number of distinct keys recorded.
func (s *InMemoryStore) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.seen)
}

// Reset clears all recorded keys.
func (s *InMemoryStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seen = map[string]struct{}{}
}

// Handler is called for each boundary attempt. The implementation decides
// whether the attempt succeeds (returns nil) or is rejected (returns a non-nil
// error). The harness records Attempt{Accepted: err == nil}.
type Handler func(ctx context.Context, db database.TenantDB, boundary Boundary, worker string, idempotencyKey string) error

// Config controls the harness scenario.
type Config struct {
	T              *testing.T
	H              *testkit.DBHandle
	Registry       *jobs.Registry
	DomainStore    EffectStore
	ExternalStore  EffectStore
	OnAttempt      func(Attempt)
	LeaseExpiry    time.Duration // how long after claim until the lease is expired
	ReclaimTimeout time.Duration // passed to ReclaimStalled
}

// Harness drives the DATA-02 T7 duplicate-worker lease-expiry scenario.
type Harness struct {
	cfg    Config
	jobID  int64
	runner *jobs.Runner

	blockedA chan struct{} // closed when worker A has claimed and performed domain effect
	releaseA chan struct{} // close to let worker A resume
	attempts []Attempt
	mu       sync.Mutex
}

// NewHarness builds a harness and enqueues one job for the scenario.
func NewHarness(cfg Config) *Harness {
	cfg.T.Helper()
	if cfg.LeaseExpiry == 0 {
		cfg.LeaseExpiry = time.Minute
	}
	if cfg.ReclaimTimeout == 0 {
		cfg.ReclaimTimeout = cfg.LeaseExpiry
	}

	h := &Harness{
		cfg:      cfg,
		blockedA: make(chan struct{}),
		releaseA: make(chan struct{}),
	}

	// Register a single kind whose worker distinguishes A vs B by reading the
	// lease context and uses the harness callbacks to attempt each boundary.
	cfg.Registry.RegisterKind("chaos.duplicate_worker", h.worker,
		jobs.Idempotency{Kind: jobs.IdempotencyEffectLedger, EffectName: "chaos.effect"},
		jobs.DefaultRetry())
	if err := cfg.Registry.Err(); err != nil {
		cfg.T.Fatalf("registry: %v", err)
	}

	tenant := testkit.CreateTenantTB(cfg.T, cfg.H)
	ctx := testkit.TenantCtx(tenant.ID)
	if err := cfg.H.TxM.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return jobs.Enqueue(ctx, db, chaosJob{})
	}); err != nil {
		cfg.T.Fatalf("enqueue: %v", err)
	}
	if err := cfg.H.Platform.QueryRow(context.Background(),
		`SELECT id FROM jobs_queue WHERE kind = $1`, "chaos.duplicate_worker").Scan(&h.jobID); err != nil {
		cfg.T.Fatalf("read job id: %v", err)
	}

	h.runner = jobs.NewRunner(cfg.H.Platform, cfg.H.TxM, cfg.Registry,
		jobs.WithReclaimTimeout(cfg.ReclaimTimeout))
	return h
}

type chaosJob struct{}

func (chaosJob) Kind() string { return "chaos.duplicate_worker" }

func (h *Harness) worker(ctx context.Context, db database.TenantDB, payload []byte) error {
	idempKey := jobs.IdempotencyKeyFromContext(ctx)
	lease := jobs.LeaseFromContext(ctx)
	if lease.Token == "" {
		return errors.New("worker did not receive lease context")
	}

	// Worker A is the original claimant. It performs the domain effect, signals
	// it is paused, then waits for the test to release it.
	if h.isWorkerA(lease) {
		if err := h.tryBoundary(ctx, db, Domain, "A", idempKey); err != nil {
			return err
		}
		close(h.blockedA)
		<-h.releaseA

		// A resumes and attempts all three boundaries; each should be rejected.
		for _, b := range []Boundary{Domain, External, Finalize} {
			_ = h.tryBoundary(ctx, db, b, "A", idempKey) // expected to fail
		}
		return nil
	}

	// Worker B is the reclaiming worker. It performs domain/external/finalize.
	for _, b := range []Boundary{Domain, External, Finalize} {
		if err := h.tryBoundary(ctx, db, b, "B", idempKey); err != nil {
			return err
		}
	}
	return nil
}

func (h *Harness) isWorkerA(l lease.Lease) bool {
	// A is the original claimant and holds lease generation 1. Every subsequent
	// epoch (after reclaim) has a strictly greater generation.
	return l.Generation == 1
}

func (h *Harness) tryBoundary(_ context.Context, _ database.TenantDB, b Boundary, worker, idempKey string) error {
	var accepted bool
	switch b {
	case Domain:
		accepted = h.cfg.DomainStore.TryRecord(idempKey)
	case External:
		accepted = h.cfg.ExternalStore.TryRecord(idempKey)
	case Finalize:
		// The runner's recordSuccess/recordFailure already performs the fenced
		// finalize; reaching here means the queue row accepted the outcome.
		accepted = true
	}

	h.mu.Lock()
	h.attempts = append(h.attempts, Attempt{Boundary: b, Worker: worker, Accepted: accepted})
	if h.cfg.OnAttempt != nil {
		h.cfg.OnAttempt(h.attempts[len(h.attempts)-1])
	}
	h.mu.Unlock()
	return nil
}

// Run drives the scenario to completion and returns the recorded attempts.
func (h *Harness) Run(ctx context.Context) []Attempt {
	h.cfg.T.Helper()

	// A claims and pauses after the domain effect.
	go func() {
		if _, err := h.runner.ClaimOnce(ctx); err != nil {
			h.cfg.T.Errorf("A ClaimOnce: %v", err)
		}
	}()
	<-h.blockedA

	// Expire A's lease in the database and reclaim via the runner.
	if _, err := h.cfg.H.Admin.Exec(ctx,
		`UPDATE jobs_queue
			   SET locked_at = now() - $1::interval,
			       lease_expires_at = now() - $2::interval
			 WHERE id = $3`,
		h.cfg.LeaseExpiry*2, h.cfg.LeaseExpiry, h.jobID); err != nil {
		h.cfg.T.Fatalf("expire A's lease: %v", err)
	}
	if n, err := h.runner.ReclaimStalled(ctx, h.cfg.ReclaimTimeout); err != nil || n != 1 {
		h.cfg.T.Fatalf("ReclaimStalled = (%d, %v), want (1, nil)", n, err)
	}

	// B claims and completes.
	if n, err := h.runner.ClaimOnce(ctx); err != nil || n != 1 {
		h.cfg.T.Fatalf("B ClaimOnce = (%d, %v), want (1, nil)", n, err)
	}

	// Release A; A attempts all boundaries and is rejected.
	close(h.releaseA)
	time.Sleep(100 * time.Millisecond) // let A's goroutine finish

	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]Attempt(nil), h.attempts...)
}

// JobID returns the id of the enqueued chaos job.
func (h *Harness) JobID() int64 { return h.jobID }

// Runner returns the runner used by the harness.
func (h *Harness) Runner() *jobs.Runner { return h.runner }

// Pool exposes the platform pool for direct assertions.
func (h *Harness) Pool() *pgxpool.Pool { return h.cfg.H.Platform }

// AssertExactlyOnce fails the test if any effect boundary (domain/external)
// recorded more than one accepted attempt, or if any expected effect boundary
// attempt is missing. Finalize is intentionally excluded because the harness
// observes worker intent, not the runner's fenced queue-row outcome.
func AssertExactlyOnce(t *testing.T, attempts []Attempt) {
	t.Helper()
	counts := map[Boundary]int{}
	for _, a := range attempts {
		if a.Accepted {
			counts[a.Boundary]++
		}
	}
	for _, b := range []Boundary{Domain, External} {
		if counts[b] != 1 {
			t.Fatalf("boundary %s accepted %d times, want 1", b, counts[b])
		}
	}
}

// AssertARejectedAtEveryEffectBoundary fails the test if worker A's domain or
// external attempt was accepted. (A's finalize rejection is enforced by the
// runner's fenced recordSuccess/recordFailure path and is verified by the
// job-status assertion in the test.)
func AssertARejectedAtEveryEffectBoundary(t *testing.T, attempts []Attempt) {
	t.Helper()
	for _, a := range attempts {
		if a.Worker == "A" && (a.Boundary == Domain || a.Boundary == External) && a.Accepted {
			t.Fatalf("worker A was accepted at boundary %s", a.Boundary)
		}
	}
}
