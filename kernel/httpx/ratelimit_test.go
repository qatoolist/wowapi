package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpx"
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
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))

	// First request consumes the single burst token → allowed, no drop.
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/x", nil))
	if drops != 0 {
		t.Fatalf("allowed request must not fire OnDrop, got %d", drops)
	}
	// Second request is over budget → 429 + one drop.
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/x", nil))
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
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
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
	// With an actor in context, KeyByActor keys on the actor id.
	actor := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(database.WithActorID(req.Context(), actor))
	if got := httpx.KeyByActor(req); got != "actor:"+actor.String() {
		t.Errorf("KeyByActor = %q, want actor:%s", got, actor)
	}
	// Without an actor, it falls back to the IP key.
	plain := httptest.NewRequest(http.MethodGet, "/", nil)
	if got := httpx.KeyByActor(plain); got == "" {
		t.Error("KeyByActor must fall back to a non-empty IP key")
	}
}

// --- test doubles ---

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time { return c.t }

type denyLimiter struct{}

func (denyLimiter) Allow(string) (bool, time.Duration) { return false, 2 * time.Second }

type allowLimiter struct{}

func (allowLimiter) Allow(string) (bool, time.Duration) { return true, 0 }
