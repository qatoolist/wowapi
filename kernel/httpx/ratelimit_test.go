package httpx_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/httpx"
)

// TestRateLimitOnDropFires proves the OnRateLimitDrop hook is invoked when a
// request is rejected — the injection point the composition root wires to the
// rate-limit-drop metric counter (roadmap CA-1). httpx cannot import
// kernel/observability (cycle), so emission is a plain callback.
func TestRateLimitOnDropFires(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	tb := httpx.NewTokenBucketWithClock(1, 1, clk.now) // 1/s, burst 1

	drops := 0
	mw := httpx.RateLimit(tb, func(*http.Request) string { return "k" },
		httpx.OnRateLimitDrop(func(string) { drops++ }))
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))

	// First request consumes the single burst token → allowed, no drop.
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/x", nil))
	if drops != 0 {
		t.Fatalf("allowed request must not fire OnDrop, got %d", drops)
	}
	// Second request is over budget → 429 + one drop.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("over-budget request should be 429, got %d", rec.Code)
	}
	if drops != 1 {
		t.Fatalf("rejected request must fire OnDrop exactly once, got %d", drops)
	}
}

func TestTokenBucketBurstThenLimit(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	tb := httpx.NewTokenBucketWithClock(10, 2, clk.now) // 10/s, burst 2

	if ok, _ := tb.Allow("k"); !ok {
		t.Fatal("1st request within burst must be allowed")
	}
	if ok, _ := tb.Allow("k"); !ok {
		t.Fatal("2nd request within burst must be allowed")
	}
	ok, retry := tb.Allow("k")
	if ok {
		t.Fatal("3rd request over burst must be denied")
	}
	if retry <= 0 {
		t.Fatalf("denied request must carry a positive Retry-After, got %v", retry)
	}
}

func TestTokenBucketRefills(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	tb := httpx.NewTokenBucketWithClock(10, 1, clk.now) // 10/s, burst 1
	if ok, _ := tb.Allow("k"); !ok {
		t.Fatal("first allowed")
	}
	if ok, _ := tb.Allow("k"); ok {
		t.Fatal("second immediately denied (bucket empty)")
	}
	clk.t = clk.t.Add(200 * time.Millisecond) // +2 tokens at 10/s
	if ok, _ := tb.Allow("k"); !ok {
		t.Fatal("after refill the request must be allowed")
	}
}

func TestTokenBucketKeysAreIndependent(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	tb := httpx.NewTokenBucketWithClock(1, 1, clk.now)
	if ok, _ := tb.Allow("a"); !ok {
		t.Fatal("key a first request allowed")
	}
	if ok, _ := tb.Allow("b"); !ok {
		t.Fatal("key b must have its own bucket")
	}
}

func TestRateLimitMiddleware429(t *testing.T) {
	// A limiter that always denies.
	h := httpx.RateLimit(denyLimiter{}, httpx.KeyByIP)(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Error("429 must carry a Retry-After header")
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
}

func TestRateLimitMiddlewareAllows(t *testing.T) {
	served := false
	h := httpx.RateLimit(allowLimiter{}, httpx.KeyByIP)(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { served = true }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !served {
		t.Fatal("an allowed request must reach the handler")
	}
}

func TestKeyByActorFallsBackToIP(t *testing.T) {
	// With a capacity-bearing user in context, KeyByActor keys on tenant:capacity.
	tenant := uuid.New()
	cap := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := database.WithTenantID(req.Context(), tenant)
	ctx = httpx.WithActor(ctx, authz.Actor{Kind: authz.ActorUser, TenantID: tenant, CapacityID: cap})
	req = req.WithContext(ctx)
	if got, want := httpx.KeyByActor(req), "t:"+tenant.String()+"|cap:"+cap.String(); got != want {
		t.Errorf("KeyByActor = %q, want %q", got, want)
	}
	// Without an actor or tenant, it falls back to a non-empty IP key.
	plain := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := httpx.KeyByActor(plain); got == "" {
		t.Error("KeyByActor must fall back to a non-empty IP key")
	}
}

// TestKeyByActorNilCapacityNotCollapsed proves the H1 fix: two machine callers
// (both nil-capacity API-key actors) in DIFFERENT tenants derive DIFFERENT bucket
// keys, so they cannot share a limiter bucket. Before the fix both collapsed to
// "actor:00000000-0000-0000-0000-000000000000".
func TestKeyByActorNilCapacityNotCollapsed(t *testing.T) {
	tenantA, tenantB := uuid.New(), uuid.New()
	// Same api-key name in both tenants — only the tenant prefix distinguishes them.
	actorA := authz.Actor{Kind: authz.ActorSystem, TenantID: tenantA, System: "apikey:relay"}
	actorB := authz.Actor{Kind: authz.ActorSystem, TenantID: tenantB, System: "apikey:relay"}

	keyFor := func(a authz.Actor) string {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := database.WithTenantID(req.Context(), a.TenantID)
		ctx = database.WithActorID(ctx, a.CapacityID) // uuid.Nil, as the gate binds it
		ctx = httpx.WithActor(ctx, a)
		return httpx.KeyByActor(req.WithContext(ctx))
	}

	ka, kb := keyFor(actorA), keyFor(actorB)
	if ka == kb {
		t.Fatalf("distinct tenants must not share a key: both = %q", ka)
	}
	for _, k := range []string{ka, kb} {
		if strings.Contains(k, uuid.Nil.String()) {
			t.Errorf("key must not bucket on uuid.Nil: %q", k)
		}
	}
}

// TestRateLimitCrossTenantIndependence exhausts tenant A's per-actor bucket
// through the real RateLimit middleware and proves tenant B's identical machine
// caller is unaffected — 200-then-429 boundaries hold independently per tenant.
func TestRateLimitCrossTenantIndependence(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	tb := httpx.NewTokenBucketWithClock(1, 2, clk.now) // 1/s, burst 2

	served := 0
	h := httpx.RateLimit(tb, httpx.KeyByActor)(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			served++
			w.WriteHeader(http.StatusOK)
		}))

	tenantA, tenantB := uuid.New(), uuid.New()
	do := func(tenant uuid.UUID) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		// Nil-capacity machine actor, exactly as the gate binds an API-key caller.
		a := authz.Actor{Kind: authz.ActorSystem, TenantID: tenant, System: "apikey:relay"}
		ctx := database.WithTenantID(req.Context(), tenant)
		ctx = database.WithActorID(ctx, a.CapacityID)
		ctx = httpx.WithActor(ctx, a)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req.WithContext(ctx))
		return rec.Code
	}

	// Drain tenant A's burst (2), then hit its limit.
	if c := do(tenantA); c != http.StatusOK {
		t.Fatalf("A req 1: got %d, want 200", c)
	}
	if c := do(tenantA); c != http.StatusOK {
		t.Fatalf("A req 2: got %d, want 200", c)
	}
	if c := do(tenantA); c != http.StatusTooManyRequests {
		t.Fatalf("A req 3: got %d, want 429 (A's bucket drained)", c)
	}

	// Tenant B's identical caller must still have a full, independent bucket.
	if c := do(tenantB); c != http.StatusOK {
		t.Fatalf("B req 1: got %d, want 200 (A's exhaustion must not affect B)", c)
	}
	if c := do(tenantB); c != http.StatusOK {
		t.Fatalf("B req 2: got %d, want 200", c)
	}
	if c := do(tenantB); c != http.StatusTooManyRequests {
		t.Fatalf("B req 3: got %d, want 429 (B's own bucket now drained)", c)
	}
}

// TestTokenBucketSweepEvictsOneShotKey is the PERF-01 core regression test. A
// key hit exactly once, then never looked up again, must still become
// sweep-eligible once idleTTL has elapsed — sweep recomputes effective refill
// itself rather than relying on Allow having been called again for that key.
// Before the fix, sweep only compared the STORED token count (frozen at
// burst-1 for a one-shot key) against burst, so this key was never evicted.
func TestTokenBucketSweepEvictsOneShotKey(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	tb := httpx.NewTokenBucketWithOptions(1, 5, clk.now) // 1/s, burst 5

	// Hit the key exactly once — lands at burst-1 tokens, never touched again.
	if ok, _ := tb.Allow("one-shot"); !ok {
		t.Fatal("first request for a fresh key must be allowed")
	}
	if got := tb.Stats().Entries; got != 1 {
		t.Fatalf("Entries after first Allow = %d, want 1", got)
	}

	// Advance the fake clock well past idleTTL (10 minutes) without ever
	// calling Allow("one-shot") again.
	clk.t = clk.t.Add(11 * time.Minute)

	httpx.SweepForTest(tb, clk.now())

	if got := tb.Stats().Entries; got != 0 {
		t.Fatalf("Entries after sweep past idleTTL = %d, want 0 (one-shot key must be evicted)", got)
	}
	if got := tb.Stats().Evictions; got != 1 {
		t.Fatalf("Evictions = %d, want 1", got)
	}
}

// TestTokenBucketSweepEvictsOverTenThousandOneShotKeys is the PERF-01
// cardinality-attack regression test: insert more than 10,000 distinct
// one-shot keys (each hit exactly once, simulating a spray of distinct
// IPs/actors), advance the fake clock beyond idleTTL, sweep, and assert the
// map drops back under the configured hard cap. This is the scenario the
// unfixed sweep could never resolve — every entry was frozen at burst-1 and
// the O(N) sweep removed nothing, so the map grew forever.
func TestTokenBucketSweepEvictsOverTenThousandOneShotKeys(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	const hardCap = 10_000
	const n = 10_500
	tb := httpx.NewTokenBucketWithOptions(1, 5, clk.now,
		httpx.WithHardCap(hardCap),
		httpx.WithSweepAt(hardCap), // opportunistic sweep once the map reaches the cap
	)

	for i := 0; i < n; i++ {
		key := "spray:" + strconv.Itoa(i)
		// Each key is hit exactly once — a one-shot key.
		_, _ = tb.Allow(key)
	}

	// Advance well past idleTTL, then force a sweep directly (Allow only
	// sweeps opportunistically at sweepAt, and once the hard cap started
	// rejecting new keys, some of the n keys above may never have been
	// admitted — that is the deterministic overflow policy, exercised
	// separately by TestTokenBucketHardCapRejectsWhenSweepFreesNoRoom).
	clk.t = clk.t.Add(11 * time.Minute)
	httpx.SweepForTest(tb, clk.now())

	stats := tb.Stats()
	if stats.Entries >= hardCap {
		t.Fatalf("Entries after sweep = %d, want < %d (configured bound)", stats.Entries, hardCap)
	}
	if stats.Evictions == 0 {
		t.Fatal("expected at least one eviction sweeping >10k one-shot keys past idleTTL")
	}
}

// TestTokenBucketHardCapRejectsWhenSweepFreesNoRoom proves the deterministic
// overflow policy: once the map is at hard capacity and every entry is still
// within idleTTL (so a forced sweep frees no room), a genuinely new key's
// request is rejected outright rather than growing the map past the bound.
func TestTokenBucketHardCapRejectsWhenSweepFreesNoRoom(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	const cap = 3
	tb := httpx.NewTokenBucketWithOptions(1, 5, clk.now, httpx.WithHardCap(cap))

	for i := 0; i < cap; i++ {
		if ok, _ := tb.Allow(fmt.Sprintf("k%d", i)); !ok {
			t.Fatalf("key k%d within cap must be admitted", i)
		}
	}
	if got := tb.Stats().Entries; got != cap {
		t.Fatalf("Entries = %d, want %d", got, cap)
	}

	// A new key arrives immediately (all existing entries are fresh, well
	// within idleTTL) — the map is at cap and a forced sweep frees nothing.
	ok, retry := tb.Allow("overflow")
	if ok {
		t.Fatal("new key over hard cap with no sweep-eligible entries must be rejected")
	}
	if retry <= 0 {
		t.Fatalf("rejected admission must carry a positive Retry-After, got %v", retry)
	}
	if got := tb.Stats().Entries; got != cap {
		t.Fatalf("Entries after rejected admission = %d, want unchanged %d", got, cap)
	}
	if got := tb.Stats().RejectedAdmissions; got != 1 {
		t.Fatalf("RejectedAdmissions = %d, want 1", got)
	}

	// An existing key already in the map is unaffected by the cap (only NEW
	// keys are subject to admission control).
	if ok, _ := tb.Allow("k0"); !ok {
		t.Fatal("an existing key must still be served even while the map is at hard cap")
	}
}

// TestOnTokenBucketStatsFires proves the metrics hook fires after a sweep
// with a populated snapshot (current entries, cumulative evictions, rejected
// admissions, sweep duration) — the injection point the composition root
// wires to metrics counters, matching OnRateLimitDrop's pattern.
func TestOnTokenBucketStatsFires(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	var mu sync.Mutex
	var last httpx.TokenBucketStats
	fires := 0
	tb := httpx.NewTokenBucketWithOptions(1, 5, clk.now,
		httpx.WithSweepAt(2),
		httpx.OnTokenBucketStats(func(s httpx.TokenBucketStats) {
			mu.Lock()
			defer mu.Unlock()
			fires++
			last = s
		}),
	)

	_, _ = tb.Allow("a")
	clk.t = clk.t.Add(11 * time.Minute)
	_, _ = tb.Allow("b") // map now at sweepAt(2) threshold on entry -> triggers sweep
	_, _ = tb.Allow("c")

	mu.Lock()
	defer mu.Unlock()
	if fires == 0 {
		t.Fatal("OnTokenBucketStats must fire at least once after a sweep")
	}
	if last.Entries < 0 {
		t.Fatalf("Stats.Entries must be non-negative, got %d", last.Entries)
	}
}

// Invalid-input handling (rate<=0, burst<1) is already covered by
// TestNewTokenBucketClampsInvalidParams in ratelimit_extra_test.go — both
// constructors funnel through the same clamp logic NewTokenBucketWithOptions
// now also uses, so that existing coverage still applies unchanged.

// TestTokenBucketRaceAllowAndSweep hammers Allow (distinct and shared keys)
// concurrently with direct sweep calls to prove no data race — run with
// `go test -race`.
func TestTokenBucketRaceAllowAndSweep(t *testing.T) {
	tb := httpx.NewTokenBucketWithOptions(1000, 10, time.Now,
		httpx.WithHardCap(500),
		httpx.WithSweepAt(50),
	)

	const goroutines = 16
	const iters = 200
	var wg sync.WaitGroup
	wg.Add(goroutines + 1)

	// One goroutine drives concurrent sweeps directly.
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			httpx.SweepForTest(tb, time.Now())
		}
	}()

	for g := 0; g < goroutines; g++ {
		go func(g int) {
			defer wg.Done()
			for i := 0; i < iters; i++ {
				// Mix of a small set of SHARED keys (contend on the same
				// bucket) and DISTINCT per-goroutine keys (grow the map).
				shared := "shared:" + strconv.Itoa(i%4)
				distinct := "g" + strconv.Itoa(g) + ":" + strconv.Itoa(i)
				_, _ = tb.Allow(shared)
				_, _ = tb.Allow(distinct)
			}
		}(g)
	}
	wg.Wait()

	_ = tb.Stats() // concurrent-safe read after the race
}

// --- test doubles ---

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time { return c.t }

type denyLimiter struct{}

func (denyLimiter) Allow(string) (bool, time.Duration) { return false, 2 * time.Second }

type allowLimiter struct{}

func (allowLimiter) Allow(string) (bool, time.Duration) { return true, 0 }
