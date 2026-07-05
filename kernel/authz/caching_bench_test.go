package authz

// Hot-path benchmarks for the ActiveAssignments cache (roadmap R1, backlog B-2).
//
// CachingStore fronts the single hottest authorization read — ActiveAssignments —
// per (tenant, actor). On a burst of requests from one principal the hit path
// (fresh entry) must serve without touching inner; the miss path reloads and
// re-caches. These are internal benchmarks (package authz) so they can inject a
// deterministic clock via newCachingStore, exactly like the caching unit tests.
// They exercise the real decorator against the in-memory countingStore fake
// (defined in caching_internal_test.go) — no DB, matching the evaluator benches.

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

// stepClk advances by step on every read, so successive now() calls straddle the
// TTL — used to force a deterministic cache miss on every iteration.
type stepClk struct {
	t    time.Time
	step time.Duration
}

func (c *stepClk) now() time.Time {
	c.t = c.t.Add(c.step)
	return c.t
}

// BenchmarkCachingStoreHit measures the fresh-cache fast path: actor key build,
// map lookup, and the defensive slices.Clone of the cached assignments. A frozen
// clock keeps every post-priming read inside the TTL, so inner is never touched.
func BenchmarkCachingStoreHit(b *testing.B) {
	clk := &fakeClk{t: time.Unix(1000, 0)}
	inner := &countingStore{asgs: []Assignment{
		{RoleKey: "core.tenant.admin", Perms: []string{"requests.request.read", "requests.request.list"}},
	}}
	c := newCachingStore(inner, time.Hour, clk.now)
	a := Actor{TenantID: uuid.New(), CapacityID: uuid.New()}
	ctx := context.Background()

	// Prime the cache so the timed loop is all hits.
	if _, err := c.ActiveAssignments(ctx, nil, a, clk.now()); err != nil {
		b.Fatalf("prime: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.ActiveAssignments(ctx, nil, a, clk.now()); err != nil {
			b.Fatalf("hit: %v", err)
		}
	}
	b.StopTimer()
	if inner.calls != 1 {
		b.Fatalf("inner called %d times, want 1 (loop must be all cache hits)", inner.calls)
	}
}

// BenchmarkCachingStoreMiss measures the reload path: the entry is always stale
// (the clock advances past the TTL between reads), so each call falls through to
// inner and re-caches — cloning on store and on return.
func BenchmarkCachingStoreMiss(b *testing.B) {
	clk := &stepClk{t: time.Unix(1000, 0), step: 2 * time.Second}
	inner := &countingStore{asgs: []Assignment{
		{RoleKey: "core.tenant.admin", Perms: []string{"requests.request.read", "requests.request.list"}},
	}}
	c := newCachingStore(inner, time.Second, clk.now)
	a := Actor{TenantID: uuid.New(), CapacityID: uuid.New()}
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := c.ActiveAssignments(ctx, nil, a, clk.t); err != nil {
			b.Fatalf("miss: %v", err)
		}
	}
	b.StopTimer()
	if inner.calls != b.N {
		b.Fatalf("inner called %d times, want %d (every read must miss)", inner.calls, b.N)
	}
}
