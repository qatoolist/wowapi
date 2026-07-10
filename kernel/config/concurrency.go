package config

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Concurrency is the framework's capacity-budget model (backlog B6, benchmark
// §Concurrency: "Capacity Budget Instead Of Independent Knobs"). Today's knobs
// — HTTP body/timeout limits, the per-client rate limiter, DB pool size, job
// runner pool size — are each independently bounded but nothing reasons about
// them TOGETHER across a deployment shape. Concurrency adds:
//
//  1. An in-flight HTTP request cap (HTTPMaxInFlight) enforced by a
//     backpressure middleware (kernel/httpx) that rejects with a configured
//     overload status BEFORE the DB pool is exhausted.
//
//  2. Deployment-shape knobs (Replicas, RuntimePoolMax, PlatformPoolMax,
//     MigratePoolMax, ReservedAdmin) that feed the capacity-budget formula:
//
//     replicas*(runtime_pool_max+platform_pool_max) + migrate_pool_max + reserved_admin <= db.max_conns
//
//     checked by CheckCapacity / enforced by Validate per CapacityMode.
//
//  3. A worker pool cap (WorkerMaxJobs) mirroring kernel/jobs' poolSize knob,
//     carried here so the same capacity budget can account for worker
//     concurrency alongside HTTP (kernel/jobs itself is unaffected — this is
//     advisory bookkeeping, not a new enforcement point in the runner).
//
// ROLLOUT (backlog B6 risk: "default budget too tight breaks existing
// deploys — ship advisory-then-enforced"):
//
//   - CapacityMode defaults to "advisory": an oversubscribed shape is
//     reported (CheckCapacity returns a non-nil warning; the boot path and
//     `wowapi config capacity` print it) but Validate() does NOT fail. Existing
//     deployments that haven't set Replicas/pool-max knobs are entirely
//     unaffected — CheckCapacity is a no-op while Replicas == 0 (unconfigured).
//   - CapacityMode: "enforced" flips Validate() to fail closed on the same
//     oversubscribed shape. A product opts in only after using
//     `wowapi config capacity` (or the advisory boot warning) to size its
//     deployment shape correctly.
//   - HTTPMaxInFlight defaults to 0, which the backpressure middleware treats
//     as "disabled" (pass-through, no limiter installed) — a bounded semaphore
//     is only sized and wired once a product sets a cap explicitly. No current
//     deployment starts returning the overload status unexpectedly.
type Concurrency struct {
	// HTTPMaxInFlight bounds concurrent in-flight HTTP requests via a bounded
	// semaphore in kernel/httpx's backpressure middleware. 0 disables the
	// limiter (pass-through) — the safe default for existing deployments.
	HTTPMaxInFlight int `conf:"http_max_in_flight" default:"0" json:"http_max_in_flight" doc:"max concurrent in-flight HTTP requests (0 disables the backpressure limiter)"`

	// WorkerMaxJobs mirrors kernel/jobs' runner pool size for capacity
	// bookkeeping; it does not itself change runner behavior (products still
	// configure the runner via jobs.WithPoolSize). 0 means "not tracked" here.
	WorkerMaxJobs int `conf:"worker_max_jobs" default:"0" json:"worker_max_jobs" doc:"worker pool size counted toward capacity bookkeeping (0 = not tracked)"`

	// PlatformMaxInFlight bounds concurrent in-flight requests specifically
	// against the platform (cross-tenant) pool, e.g. API-key verification.
	// 0 means "not tracked" — no separate platform limiter is installed.
	PlatformMaxInFlight int `conf:"platform_max_in_flight" default:"0" json:"platform_max_in_flight" doc:"max concurrent in-flight platform-pool work (0 = not tracked)"`

	// --- deployment-shape knobs feeding the capacity-budget formula ---

	// Replicas is the number of process replicas of THIS deployment (api or
	// worker) sharing the same database. 0 means "not configured" and the
	// entire capacity-budget check (CheckCapacity) is skipped — the formula is
	// undefined without a declared shape, so leaving this unset must never
	// produce a spurious pass or fail.
	Replicas int `conf:"replicas" default:"0" json:"replicas" doc:"process replica count for the capacity-budget formula (0 = not configured, check skipped)"`

	// RuntimePoolMax is the runtime (app_rt) pool's max_conns as budgeted per
	// replica. Distinct from db.max_conns (a single pool's own cap): this is
	// the value the deployment-shape formula multiplies by Replicas. Typically
	// equal to db.max_conns for a single-pool-per-process deployment.
	RuntimePoolMax int `conf:"runtime_pool_max" default:"0" json:"runtime_pool_max" doc:"runtime pool max_conns per replica, for the capacity-budget formula"`

	// PlatformPoolMax is the platform (app_platform) pool's max_conns per
	// replica.
	PlatformPoolMax int `conf:"platform_pool_max" default:"0" json:"platform_pool_max" doc:"platform pool max_conns per replica, for the capacity-budget formula"`

	// MigratePoolMax is the one-shot migrate process's pool max_conns (not
	// multiplied by Replicas: migrations run as a single job, not per-replica).
	MigratePoolMax int `conf:"migrate_pool_max" default:"0" json:"migrate_pool_max" doc:"migrate process pool max_conns (counted once, not per replica)"`

	// ReservedAdmin is connections held back for admin/operator access
	// (psql, pgAdmin, an emergency migration) that must always fit under
	// db_max_connections alongside the application's own budget.
	ReservedAdmin int `conf:"reserved_admin" default:"0" json:"reserved_admin" doc:"connections reserved for admin/operator access, counted in the capacity-budget formula"`

	// CapacityMode gates whether an oversubscribed shape fails boot
	// ("enforced") or only warns ("advisory", the default — see rollout
	// guidance on the Concurrency type doc).
	CapacityMode CapacityMode `conf:"capacity_mode" default:"advisory" json:"capacity_mode" doc:"capacity-budget enforcement: advisory (warn only, default) | enforced (fail boot on oversubscription)"`

	Overload Overload `conf:"overload" json:"overload"`
}

// CapacityMode selects whether an oversubscribed deployment shape fails boot
// or only warns.
type CapacityMode string

const (
	// CapacityModeAdvisory warns (CheckCapacity) without failing Validate.
	// This is the default so shipping this feature cannot break an existing
	// deployment that hasn't sized its shape knobs yet.
	CapacityModeAdvisory CapacityMode = "advisory"
	// CapacityModeEnforced fails Validate on an oversubscribed shape. Opt-in.
	CapacityModeEnforced CapacityMode = "enforced"
)

// valid reports whether m is a known CapacityMode.
func (m CapacityMode) valid() bool {
	switch m {
	case CapacityModeAdvisory, CapacityModeEnforced:
		return true
	}
	return false
}

// Overload configures the response the backpressure middleware sends when
// HTTPMaxInFlight is exceeded.
type Overload struct {
	// Status is the HTTP status code written on overload. Defaults to 503
	// (Service Unavailable) per benchmark §Concurrency; some deployments may
	// prefer 429 (Too Many Requests) to align with rate-limit semantics —
	// either is accepted, anything else is rejected by Validate.
	Status int `conf:"api_status" default:"503" json:"api_status" doc:"HTTP status returned on overload (503 or 429)"`
	// RetryAfter is the Retry-After hint (seconds, rounded up) sent with the
	// overload response.
	RetryAfter time.Duration `conf:"retry_after" default:"2s" json:"retry_after" doc:"Retry-After hint sent with the overload response"`
}

// ConcurrencyDefaults returns the safe, zero-impact defaults: no in-flight
// limiter installed (HTTPMaxInFlight=0), no deployment shape declared
// (Replicas=0, so CheckCapacity is a no-op), advisory capacity mode, and a
// 503+2s overload response for when a product DOES opt in.
func ConcurrencyDefaults() Concurrency {
	return Concurrency{
		HTTPMaxInFlight:     0,
		WorkerMaxJobs:       0,
		PlatformMaxInFlight: 0,
		Replicas:            0,
		RuntimePoolMax:      0,
		PlatformPoolMax:     0,
		MigratePoolMax:      0,
		ReservedAdmin:       0,
		CapacityMode:        CapacityModeAdvisory,
		Overload:            Overload{Status: http.StatusServiceUnavailable, RetryAfter: 2 * time.Second},
	}
}

// validate checks per-field ranges — NOT the cross-field capacity budget,
// which is CheckCapacity/checkCapacityEnforced (it needs db.max_conns from the
// enclosing Framework, so it cannot live as a Concurrency-only method).
func (c Concurrency) validate() error {
	var errs []error
	add := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }

	if c.HTTPMaxInFlight < 0 {
		add("concurrency.http_max_in_flight: must be >= 0 (0 disables the limiter), got %d", c.HTTPMaxInFlight)
	}
	if c.WorkerMaxJobs < 0 {
		add("concurrency.worker_max_jobs: must be >= 0, got %d", c.WorkerMaxJobs)
	}
	if c.PlatformMaxInFlight < 0 {
		add("concurrency.platform_max_in_flight: must be >= 0, got %d", c.PlatformMaxInFlight)
	}
	if c.Replicas < 0 {
		add("concurrency.replicas: must be >= 0, got %d", c.Replicas)
	}
	if c.RuntimePoolMax < 0 {
		add("concurrency.runtime_pool_max: must be >= 0, got %d", c.RuntimePoolMax)
	}
	if c.PlatformPoolMax < 0 {
		add("concurrency.platform_pool_max: must be >= 0, got %d", c.PlatformPoolMax)
	}
	if c.MigratePoolMax < 0 {
		add("concurrency.migrate_pool_max: must be >= 0, got %d", c.MigratePoolMax)
	}
	if c.ReservedAdmin < 0 {
		add("concurrency.reserved_admin: must be >= 0, got %d", c.ReservedAdmin)
	}
	if !c.CapacityMode.valid() {
		add("concurrency.capacity_mode: %q is not one of advisory|enforced", string(c.CapacityMode))
	}
	if c.Overload.Status != http.StatusServiceUnavailable && c.Overload.Status != http.StatusTooManyRequests {
		add("concurrency.overload.status: %d is not one of 503|429", c.Overload.Status)
	}
	if c.Overload.RetryAfter <= 0 {
		add("concurrency.overload.retry_after: must be > 0, got %v", c.Overload.RetryAfter)
	}

	return errors.Join(errs...)
}

// CapacityProblem describes a failed capacity-budget check: the computed
// demand exceeded the database's max_conns.
type CapacityProblem struct {
	Replicas        int
	RuntimePoolMax  int
	PlatformPoolMax int
	MigratePoolMax  int
	ReservedAdmin   int
	Demand          int
	DBMaxConns      int
}

func (p *CapacityProblem) Error() string {
	return fmt.Sprintf(
		"concurrency: capacity budget exceeded: replicas(%d)*(runtime_pool_max(%d)+platform_pool_max(%d)) + migrate_pool_max(%d) + reserved_admin(%d) = %d, which exceeds db.max_conns(%d)",
		p.Replicas, p.RuntimePoolMax, p.PlatformPoolMax, p.MigratePoolMax, p.ReservedAdmin, p.Demand, p.DBMaxConns,
	)
}

// CheckCapacity evaluates the deployment-shape formula from benchmark
// §Concurrency:
//
//	replicas*(runtime_pool_max+platform_pool_max) + migrate_pool_max + reserved_admin <= db_max_connections
//
// It returns nil when Replicas is 0 (shape not configured — the check is a
// deliberate no-op so an unconfigured product never gets a spurious result)
// or when the shape fits; otherwise it returns a *CapacityProblem describing
// the oversubscription. Callers decide what to do with a non-nil result:
// Validate() (via checkCapacityEnforced) turns it into a hard failure only in
// CapacityModeEnforced; the CLI (`wowapi config capacity`) and boot-time
// logging always print it regardless of mode.
func CheckCapacity(f Framework) *CapacityProblem {
	c := f.Concurrency
	if c.Replicas == 0 {
		return nil // shape not declared; nothing to check
	}
	demand := c.Replicas*(c.RuntimePoolMax+c.PlatformPoolMax) + c.MigratePoolMax + c.ReservedAdmin
	if demand <= f.DB.MaxConns {
		return nil
	}
	return &CapacityProblem{
		Replicas:        c.Replicas,
		RuntimePoolMax:  c.RuntimePoolMax,
		PlatformPoolMax: c.PlatformPoolMax,
		MigratePoolMax:  c.MigratePoolMax,
		ReservedAdmin:   c.ReservedAdmin,
		Demand:          demand,
		DBMaxConns:      f.DB.MaxConns,
	}
}

// checkCapacityEnforced returns a non-nil error only in CapacityModeEnforced,
// so Validate() fails closed on an oversubscribed shape when (and only when)
// a product has opted in. Advisory mode never contributes to Validate's
// error set — callers surface CheckCapacity's warning through their own
// (non-fatal) path instead (boot warnings, `wowapi config capacity`).
func checkCapacityEnforced(f Framework) error {
	if f.Concurrency.CapacityMode != CapacityModeEnforced {
		return nil
	}
	if p := CheckCapacity(f); p != nil {
		return p
	}
	return nil
}
