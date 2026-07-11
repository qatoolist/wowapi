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

// TokenBucketStats is a point-in-time snapshot of TokenBucket bookkeeping,
// exposed via OnTokenBucketStats (roadmap PERF-01). Counters are cumulative
// since construction; Entries is the current map size.
type TokenBucketStats struct {
	// Entries is the current number of live buckets in the map.
	Entries int
	// Evictions is the cumulative count of buckets removed by sweep.
	Evictions uint64
	// RejectedAdmissions is the cumulative count of new-key admissions refused
	// because the map was at HardCap and a synchronous sweep freed no room
	// (see TokenBucketOption WithHardCap).
	RejectedAdmissions uint64
	// SweepDuration is how long the most recently completed sweep took.
	SweepDuration time.Duration
}

// TokenBucketOption customizes a TokenBucket at construction time.
type TokenBucketOption func(*TokenBucket)

// WithHardCap bounds the maximum number of live keys TokenBucket will hold.
// When a genuinely new key arrives and the map is already at cap, TokenBucket
// forces a synchronous sweep first; if the map is still at cap afterward
// (every entry is either non-idle or within idleTTL — i.e. under active,
// possibly adversarial, load), the new key's request is rejected as if rate
// limited rather than growing the map further. This is the deterministic
// overflow policy: correctness under a cardinality attack must not depend on
// sweep ever running to completion. cap <= 0 disables the bound (unbounded,
// the pre-PERF-01 behavior) — not recommended for production.
func WithHardCap(cap int) TokenBucketOption {
	return func(tb *TokenBucket) { tb.hardCap = cap }
}

// WithSweepAt overrides the map size at which Allow triggers an opportunistic
// sweep (default 10000). Exposed for tests exercising sweep behavior at
// smaller cardinalities without waiting for the production threshold.
func WithSweepAt(n int) TokenBucketOption {
	return func(tb *TokenBucket) { tb.sweepAt = n }
}

// OnTokenBucketStats registers a callback fired after every sweep (opportunistic
// or forced) with the current stats snapshot. The composition root wires this
// to metrics counters — httpx must not import kernel/observability (cycle), so
// emission is a plain callback, matching OnRateLimitDrop's pattern.
func OnTokenBucketStats(fn func(TokenBucketStats)) TokenBucketOption {
	return func(tb *TokenBucket) { tb.onStats = fn }
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
	// idleTTL after which a bucket that would now be full (accounting for
	// refill since it was last touched) is eligible for sweeping.
	idleTTL time.Duration
	// sweepAt bounds the map: an opportunistic sweep runs when it grows past
	// this many keys.
	sweepAt int
	// hardCap is the absolute ceiling on live keys (0 = unbounded). See
	// WithHardCap.
	hardCap int
	// onStats, evictions, and rejectedAdmissions back TokenBucketStats/
	// OnTokenBucketStats.
	onStats            func(TokenBucketStats)
	evictions          uint64
	rejectedAdmissions uint64
}

// NewTokenBucket builds a limiter of rate tokens/sec with the given burst.
func NewTokenBucket(ratePerSec float64, burst int) *TokenBucket {
	return NewTokenBucketWithClock(ratePerSec, burst, time.Now)
}

// NewTokenBucketWithClock is NewTokenBucket with an injectable clock (tests).
func NewTokenBucketWithClock(ratePerSec float64, burst int, now func() time.Time) *TokenBucket {
	return NewTokenBucketWithOptions(ratePerSec, burst, now)
}

// NewTokenBucketWithOptions is NewTokenBucketWithClock plus TokenBucketOptions
// (hard capacity, sweep threshold, stats callback). Kept as a separate
// constructor rather than changing NewTokenBucket/NewTokenBucketWithClock's
// signatures, which existing callers (including wowsociety) depend on.
func NewTokenBucketWithOptions(ratePerSec float64, burst int, now func() time.Time, opts ...TokenBucketOption) *TokenBucket {
	if ratePerSec <= 0 {
		ratePerSec = 1
	}
	if burst < 1 {
		burst = 1
	}
	tb := &TokenBucket{
		rate:    ratePerSec,
		burst:   float64(burst),
		now:     now,
		buckets: make(map[string]*tokenBucket),
		idleTTL: 10 * time.Minute,
		sweepAt: 10000,
	}
	for _, o := range opts {
		o(tb)
	}
	return tb
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
		// A genuinely new key. Enforce the hard cap deterministically: if the
		// map is still at/over capacity, force a synchronous sweep, then — if
		// still at capacity — reject this admission rather than growing the
		// map further. Sweep-eligibility does not depend on this key (or any
		// other) ever being looked up again (PERF-01).
		if tb.hardCap > 0 && len(tb.buckets) >= tb.hardCap {
			tb.sweep(now)
			if len(tb.buckets) >= tb.hardCap {
				tb.rejectedAdmissions++
				tb.emitStats()
				return false, time.Duration(1 / tb.rate * float64(time.Second))
			}
		}
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

// sweep drops buckets that are idle beyond idleTTL AND would now be full —
// full is computed with the SAME refill formula Allow uses, projected forward
// from the bucket's own last-touch time, not merely the last value Allow
// happened to store. This is the PERF-01 fix: a one-shot key (hit exactly
// once, never looked up again) still accrues virtual refill over time and
// becomes sweep-eligible once idleTTL has passed, instead of being frozen at
// burst-1 forever. Caller must hold tb.mu.
func (tb *TokenBucket) sweep(now time.Time) {
	start := time.Now() // wall-clock sweep cost, independent of the injected clock
	for k, b := range tb.buckets {
		if now.Sub(b.lastSeen) <= tb.idleTTL {
			continue
		}
		effective := math.Min(tb.burst, b.tokens+now.Sub(b.last).Seconds()*tb.rate)
		if effective >= tb.burst {
			delete(tb.buckets, k)
			tb.evictions++
		}
	}
	tb.emitStatsWithDuration(time.Since(start))
}

// emitStats fires onStats with a zero sweep duration (used from the hard-cap
// rejection path, which does not itself measure a fresh sweep).
func (tb *TokenBucket) emitStats() {
	tb.emitStatsWithDuration(0)
}

// emitStatsWithDuration fires onStats, if registered, with the current
// snapshot. Caller must hold tb.mu.
func (tb *TokenBucket) emitStatsWithDuration(d time.Duration) {
	if tb.onStats == nil {
		return
	}
	tb.onStats(TokenBucketStats{
		Entries:            len(tb.buckets),
		Evictions:          tb.evictions,
		RejectedAdmissions: tb.rejectedAdmissions,
		SweepDuration:      d,
	})
}

// Stats returns a current snapshot of TokenBucket bookkeeping. Safe for
// concurrent use.
func (tb *TokenBucket) Stats() TokenBucketStats {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return TokenBucketStats{
		Entries:            len(tb.buckets),
		Evictions:          tb.evictions,
		RejectedAdmissions: tb.rejectedAdmissions,
	}
}
