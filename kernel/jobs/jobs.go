// Package jobs is wowapi's Postgres-backed job runner (D-0047 — a focused queue
// behind the framework interfaces, NOT River). Modules enqueue a job in the SAME
// transaction as their business write (so the job commits atomically with the
// write, or not at all); a worker process claims jobs with FOR UPDATE SKIP
// LOCKED, executes each in a transaction bound to the job's tenant, and retries
// with exponential backoff + jitter until success or exhaustion (DLQ). Contract:
// docs/blueprint/07-platform-services.md §3.
//
// The queue lives in jobs_queue (global; the tenant travels in the row's
// tenant_id, NULL for global jobs). job_runs is an append-only reporting mirror.
// jobs_queue is kernel-only: app_rt may only INSERT (enqueue), while the runner
// connects as app_platform and holds SELECT/INSERT/UPDATE on both tables.
//
// Import boundary (depguard): stdlib + kernel/database + kernel/errors +
// kernel/model + kernel/config + pgx + google/uuid. Never module/app/adapters/
// testkit in production code.
package jobs

import (
	"context"
	"time"

	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/lease"
)

// Job is a payload that knows its own kind. Payload structs implement it; the
// kind selects the registered Worker at execution time. A Job is JSON-marshaled
// into jobs_queue.payload on enqueue and handed back to the Worker as raw bytes.
type Job interface {
	// Kind is the stable identifier a Worker registers under. Naming mirrors
	// event kinds: "module.resource.verb" (e.g. "notify.email.send").
	Kind() string
}

// Worker executes one job. It receives the tenant-bound database facade (RLS is
// already scoped to the job's tenant via SET LOCAL) and the raw JSON payload it
// unmarshals into its own typed struct. A returned error triggers retry/backoff;
// a nil error marks the job completed.
//
// DELIVERY IS AT-LEAST-ONCE. The worker's DB effect commits in the tenant tx,
// but the queue 'completed' mark commits in a SEPARATE tx (a different role and
// pool), so a crash in between — or a reclaim of an over-running job — reruns
// the worker. Unlike event handlers (which get the processed_events inbox for
// exactly-once DB effects), jobs have NO framework-provided dedup. Therefore a
// worker MUST be idempotent by construction: DB-only work should be naturally
// idempotent (upserts, version checks); a worker with an EXTERNAL side effect
// (email, webhook, payment) MUST carry its own idempotency key against the
// provider, or it can double-fire (review finding ARCH-59).
//
// At registration time every worker must declare exactly one duplicate-safety
// mechanism via Idempotency (see RegisterKind). The runner passes the stable
// job idempotency key and lease context to the worker through ctx using
// WithJobContext; workers read them back with JobIDFromContext,
// IdempotencyKeyFromContext, and LeaseFromContext.
type Worker func(ctx context.Context, db database.TenantDB, payload []byte) error

// IdempotencyKind enumerates the duplicate-safety mechanisms a worker may
// declare. A worker must declare exactly one.
type IdempotencyKind int

const (
	// IdempotencyEffectLedger means the worker writes its effect to an
	// inbox/effect ledger unique on (job_id, effect_name).
	IdempotencyEffectLedger IdempotencyKind = iota
	// IdempotencyDomainCAS means the worker uses a domain compare-and-swap
	// (e.g. a version column or conditional upsert) to make itself safe.
	IdempotencyDomainCAS
	// IdempotencyProviderKey means the worker carries its own idempotency key
	// against an external provider.
	IdempotencyProviderKey
)

// Idempotency declares how a worker protects against duplicate effects. Exactly
// one kind must be set; EffectName is required only for IdempotencyEffectLedger.
type Idempotency struct {
	Kind       IdempotencyKind
	EffectName string // required when Kind == IdempotencyEffectLedger
}

// Validate checks that exactly one mechanism is declared.
func (i Idempotency) Validate() error {
	switch i.Kind {
	case IdempotencyEffectLedger:
		if i.EffectName == "" {
			return kerr.E(kerr.KindInternal, "invalid_idempotency", "effect-ledger idempotency requires EffectName")
		}
		return nil
	case IdempotencyDomainCAS, IdempotencyProviderKey:
		return nil
	default:
		return kerr.E(kerr.KindInternal, "invalid_idempotency", "worker must declare exactly one duplicate-safety mechanism: effect ledger, domain CAS, or provider idempotency key")
	}
}

// jobContextKey is the private type for context values passed to workers.
type jobContextKey int

const (
	jobIDKey jobContextKey = iota
	idempotencyKey
	leaseKey
)

// WithJobContext injects the job id, stable idempotency key, and lease context
// into ctx for a worker invocation. Callers should not need this directly; it
// is used by the runner.
func WithJobContext(ctx context.Context, jobID int64, idempKey string, l lease.Lease) context.Context {
	ctx = context.WithValue(ctx, jobIDKey, jobID)
	ctx = context.WithValue(ctx, idempotencyKey, idempKey)
	return context.WithValue(ctx, leaseKey, l)
}

// JobIDFromContext returns the job ID passed to the worker, or 0 if absent.
func JobIDFromContext(ctx context.Context) int64 {
	v, _ := ctx.Value(jobIDKey).(int64)
	return v
}

// IdempotencyKeyFromContext returns the stable idempotency key passed to the
// worker, or "" if absent.
func IdempotencyKeyFromContext(ctx context.Context) string {
	v, _ := ctx.Value(idempotencyKey).(string)
	return v
}

// LeaseFromContext returns the lease context passed to the worker, or a zero
// Lease if absent.
func LeaseFromContext(ctx context.Context) lease.Lease {
	v, _ := ctx.Value(leaseKey).(lease.Lease)
	return v
}

// BackoffPolicy maps a (1-based) attempt number to the delay before the next
// retry. It must be a pure function of attempt — no time.Now or rand at package
// init (jitter is derived deterministically from the attempt, see
// ExpJitterBackoff).
type BackoffPolicy func(attempt int) time.Duration

// RetryPolicy governs how a kind is retried. MaxAttempts is the total number of
// executions before a job is discarded to the DLQ; Backoff spaces the retries.
// The authoritative attempt ceiling for a specific job is its jobs_queue
// max_attempts column (set at enqueue, defaulting to 5) — MaxAttempts here is
// the policy default that DefaultRetry aligns with.
type RetryPolicy struct {
	MaxAttempts int
	Backoff     BackoffPolicy
}

const (
	// defaultMaxAttempts matches jobs_queue.max_attempts DEFAULT 5 (migration
	// 00007) so the registry policy and the table default agree.
	defaultMaxAttempts = 5
	// backoffBase is the first-retry delay; it doubles each attempt.
	backoffBase = time.Second
	// backoffCap is the ceiling on any single retry delay (blueprint: 1s→5m).
	backoffCap = 5 * time.Minute
)

// DefaultRetry is the blueprint default: 5 attempts, exponential backoff with
// jitter from 1s up to a 5m cap.
func DefaultRetry() RetryPolicy {
	return RetryPolicy{MaxAttempts: defaultMaxAttempts, Backoff: ExpJitterBackoff}
}

// ExpJitterBackoff returns an exponential backoff delay for the given attempt:
// base 1s doubling each attempt, capped at 5m, plus a deterministic jitter of up
// to 25% of the (capped) delay. The jitter is a pure function of attempt — no
// time.Now or rand — so it is safe to call anywhere and never touches package
// init. The result is non-decreasing in attempt and never exceeds 5m.
func ExpJitterBackoff(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	// Exponential growth without shift overflow: stop doubling once capped.
	d := backoffBase
	for i := 1; i < attempt; i++ {
		d <<= 1
		if d >= backoffCap {
			d = backoffCap
			break
		}
	}
	// Deterministic jitter in [0, d/4): spreads a thundering herd of same-attempt
	// retries without any global randomness. Adding to a capped d and re-capping
	// keeps the value bounded and monotonic across the plateau.
	if span := d / 4; span > 0 {
		// #nosec G115 -- attempt≥1 (guarded above) reinterpreted as a splitmix64
		// seed (distribution-only); the modulo result is < span, a positive
		// time.Duration, so the uint64→int64 conversion cannot overflow.
		d += time.Duration(jitter(uint64(attempt)) % uint64(span))
	}
	if d > backoffCap {
		d = backoffCap
	}
	return d
}

// jitter is a splitmix64 finalizer over seed — a fast, deterministic scramble so
// jitter(attempt) is well-distributed without seeding a global RNG.
func jitter(seed uint64) uint64 {
	seed += 0x9E3779B97F4A7C15
	seed = (seed ^ (seed >> 30)) * 0xBF58476D1CE4E5B9
	seed = (seed ^ (seed >> 27)) * 0x94D049BB133111EB
	return seed ^ (seed >> 31)
}
