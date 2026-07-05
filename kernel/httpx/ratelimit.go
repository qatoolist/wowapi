package httpx

import (
	"math"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
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

// KeyByActor derives a per-principal, tenant-scoped bucket key so one tenant's
// callers can never share a limiter bucket with another's (cross-tenant DoS). It
// is stable per principal even for machine callers, whose audit CapacityID is
// uuid.Nil for every API-key / system / webhook actor — keying on that id alone
// would collapse ALL machine traffic across ALL tenants into one bucket.
//
// The key is always prefixed with the bound tenant id, then the strongest
// available principal identifier: the user's capacity, else an api-key/system
// name, else the JWT subject (user id). If no principal identifier is available
// it falls back to per-IP — still tenant-prefixed when a tenant is bound — rather
// than a single shared bucket. Place RateLimit AFTER the authz gate to use this.
func KeyByActor(r *http.Request) string {
	ctx := r.Context()
	actor, hasActor := ActorFrom(ctx)

	// Tenant prefix: prefer the tenant bound in context (set by the gate), else
	// the actor's own tenant. Every key is scoped to it so buckets never cross
	// tenants.
	tenant, ok := database.TenantIDFrom(ctx)
	if !ok && hasActor && actor.TenantID != uuid.Nil {
		tenant, ok = actor.TenantID, true
	}
	prefix := ""
	if ok {
		prefix = "t:" + tenant.String() + "|"
	}

	if hasActor {
		if id := principalID(actor); id != "" {
			return prefix + id
		}
	} else if cid, ok := database.ActorIDFrom(ctx); ok && cid != uuid.Nil {
		// No full actor bound but a non-nil audit capacity is — key on it.
		return prefix + "cap:" + cid.String()
	}

	// No principal identifier at all: per-IP, still tenant-prefixed.
	return prefix + KeyByIP(r)
}

// principalID returns the strongest stable identifier of a, or "" if a carries
// none. A machine principal (api key / system / webhook) has a nil CapacityID but
// a distinct System name; a human has a capacity; a bare token still has its
// subject (user id).
func principalID(a authz.Actor) string {
	switch {
	case a.CapacityID != uuid.Nil:
		return "cap:" + a.CapacityID.String()
	case a.System != "":
		return "sys:" + a.System
	case a.UserID != uuid.Nil:
		return "usr:" + a.UserID.String()
	default:
		return ""
	}
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
