package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/kernel/httpx"
)

// TestNewTokenBucketDefaults exercises the production constructor (real clock).
func TestNewTokenBucketDefaults(t *testing.T) {
	tb := httpx.NewTokenBucket(100, 3)
	if ok, _ := tb.Allow("k"); !ok {
		t.Fatal("a fresh bucket must allow the first request")
	}
}

// TestNewTokenBucketClampsInvalidParams covers the guard branches: a non-positive
// rate and a sub-1 burst are clamped up to 1 so the limiter stays well-formed.
func TestNewTokenBucketClampsInvalidParams(t *testing.T) {
	clk := &fakeClock{t: time.Unix(0, 0)}
	tb := httpx.NewTokenBucketWithClock(0, 0, clk.now) // rate<=0 → 1/s, burst<1 → 1

	if ok, _ := tb.Allow("k"); !ok {
		t.Fatal("clamped burst of 1 must allow the first request")
	}
	if ok, retry := tb.Allow("k"); ok || retry <= 0 {
		t.Fatalf("second request must be denied with a positive retry (burst clamped to 1); ok=%v retry=%v", ok, retry)
	}
}

// TestKeyByIPWithoutPort covers the SplitHostPort-error branch: a RemoteAddr with
// no port falls back to using the whole value as the host.
func TestKeyByIPWithoutPort(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "1.2.3.4" // no ":port"
	if got := httpx.KeyByIP(r); got != "ip:1.2.3.4" {
		t.Fatalf("KeyByIP = %q, want ip:1.2.3.4", got)
	}
}
