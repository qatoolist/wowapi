package httpx

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// Rate limiting (roadmap S2, blueprint 07 §1). The kernel provides an in-process
// token-bucket limiter and a middleware that returns 429 + Retry-After + an
// RFC 7807 body. Limits are guardrails against resource exhaustion (OWASP API
// "unrestricted resource consumption"), not billing. Products place RateLimit in
// the chain and choose a key function: per-IP for unauthenticated edges,
// per-actor once the request is authenticated, or a custom per-permission key
// (e.g. a tighter bucket for PII-export routes).
//
// "In-process per pod" means each replica limits independently; a shared
// (Redis) limiter is a later adapter behind the same RateLimiter interface.

// RateLimiter decides whether a request keyed by `key` may proceed now. When it
// may not, retryAfter is a hint for the Retry-After header.
type RateLimiter interface {
	Allow(key string) (allowed bool, retryAfter time.Duration)
}

// RateLimitOption customizes the RateLimit middleware.
type RateLimitOption func(*rateLimitCfg)

type rateLimitCfg struct {
	onDrop func(route string)
}

// OnRateLimitDrop registers a callback fired whenever a request is rejected
// (429). The composition root wires this to a metrics counter — httpx must not
// import kernel/observability (observability imports httpx), so the emission is
// injected as a plain callback (roadmap CA-1). route is r.Pattern.
func OnRateLimitDrop(fn func(route string)) RateLimitOption {
	return func(c *rateLimitCfg) { c.onDrop = fn }
}

// RateLimit rejects requests that exceed the limiter with 429 + Retry-After. The
// keyFn derives the bucket key from the request (see KeyByIP / KeyByActor). A nil
// keyFn defaults to KeyByIP.
func RateLimit(limiter RateLimiter, keyFn func(*http.Request) string, opts ...RateLimitOption) Middleware {
	if keyFn == nil {
		keyFn = KeyByIP
	}
	var cfg rateLimitCfg
	for _, o := range opts {
		o(&cfg)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowed, retryAfter := limiter.Allow(keyFn(r))
			if !allowed {
				secs := max(int(math.Ceil(retryAfter.Seconds())), 1)
				w.Header().Set("Retry-After", strconv.Itoa(secs))
				if cfg.onDrop != nil {
					cfg.onDrop(r.Pattern)
				}
				WriteError(r.Context(), w, kerr.E(kerr.KindRateLimited, "rate_limited",
					"rate limit exceeded; retry later"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// KeyByIP keys on the client IP (RemoteAddr host). Behind the reference proxy,
// which sets X-Real-IP / X-Forwarded-For, a product that trusts its proxy should
// supply a keyFn reading the forwarded header instead.
func KeyByIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return "ip:" + host
}

// KeyByActor keys on the authenticated actor (capacity) id when present, else
// falls back to KeyByIP. Place RateLimit AFTER the authz gate to use this.
func KeyByActor(r *http.Request) string {
	if actor, ok := database.ActorIDFrom(r.Context()); ok {
		return "actor:" + actor.String()
	}
	return KeyByIP(r)
}

// --- token bucket ---

type tokenBucket struct {
	tokens   float64
	last     time.Time
	lastSeen time.Time
}

// TokenBucket is an in-memory per-key token-bucket RateLimiter. Each key refills
// at `rate` tokens/sec up to `burst`. Idle buckets are swept so the map cannot
// grow without bound under a spray of distinct keys.
type TokenBucket struct {
	mu      sync.Mutex
	rate    float64
	burst   float64
	now     func() time.Time
	buckets map[string]*tokenBucket
	// idleTTL after which a full, untouched bucket is eligible for sweeping.
	idleTTL time.Duration
	// sweepEvery bounds the map: a sweep runs when it grows past this many keys.
	sweepAt int
}

// NewTokenBucket builds a limiter of rate tokens/sec with the given burst.
func NewTokenBucket(ratePerSec float64, burst int) *TokenBucket {
	return NewTokenBucketWithClock(ratePerSec, burst, time.Now)
}

// NewTokenBucketWithClock is NewTokenBucket with an injectable clock (tests).
func NewTokenBucketWithClock(ratePerSec float64, burst int, now func() time.Time) *TokenBucket {
	if ratePerSec <= 0 {
		ratePerSec = 1
	}
	if burst < 1 {
		burst = 1
	}
	return &TokenBucket{
		rate:    ratePerSec,
		burst:   float64(burst),
		now:     now,
		buckets: make(map[string]*tokenBucket),
		idleTTL: 10 * time.Minute,
		sweepAt: 10000,
	}
}

// Allow consumes one token for key, refilling first. On denial it returns the
// time until one token is available.
func (tb *TokenBucket) Allow(key string) (bool, time.Duration) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	now := tb.now()

	if len(tb.buckets) >= tb.sweepAt {
		tb.sweep(now)
	}

	b := tb.buckets[key]
	if b == nil {
		b = &tokenBucket{tokens: tb.burst, last: now}
		tb.buckets[key] = b
	}
	// Refill proportional to elapsed time, capped at burst.
	b.tokens = math.Min(tb.burst, b.tokens+now.Sub(b.last).Seconds()*tb.rate)
	b.last = now
	b.lastSeen = now

	if b.tokens >= 1 {
		b.tokens -= 1
		return true, 0
	}
	deficit := 1 - b.tokens
	return false, time.Duration(deficit / tb.rate * float64(time.Second))
}

// sweep drops buckets that are full (not currently limited) and idle beyond
// idleTTL — removing them is lossless since a fresh bucket starts full too.
func (tb *TokenBucket) sweep(now time.Time) {
	for k, b := range tb.buckets {
		if b.tokens >= tb.burst && now.Sub(b.lastSeen) > tb.idleTTL {
			delete(tb.buckets, k)
		}
	}
}
