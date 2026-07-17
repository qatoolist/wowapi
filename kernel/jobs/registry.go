package jobs

import (
	"github.com/qatoolist/wowapi/v2/internal/sealer"

	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
)

// entry is a registered kind: the worker that executes it, its idempotency
// declaration, and its retry policy.
type entry struct {
	worker      Worker
	idempotency Idempotency
	retry       RetryPolicy
}

// Registry collects the (kind → worker + retry policy) bindings during module
// boot. It accumulates errors (duplicate kind, empty kind, nil worker) rather
// than panicking, so RegisterKind reads cleanly at call sites and boot fails
// once via Err() with every problem reported together (mirrors outbox's
// HandlerRegistry).
type Registry struct {
	kinds  map[string]entry
	errs   []error
	sealed bool
}

// Seal freezes the registry once boot validation completes: any later
// RegisterKind panics rather than introducing a job kind the running worker
// pool would dispatch without boot validation (closure review 2026-07-17, F-10).
// The sealer.Authority parameter restricts sealing to the framework's boot
// path: internal/sealer is unimportable outside the wowapi module, so a
// product module cannot prematurely seal a shared registry during Register.
func (r *Registry) Seal(sealer.Authority) { r.sealed = true }

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{kinds: map[string]entry{}}
}

// RegisterKind preserves the v1 registration API. Legacy workers are treated
// as declaring domain-level compare-and-swap protection; new registrations
// should use RegisterKindWithIdempotency so the mechanism is explicit.
func (r *Registry) RegisterKind(kind string, w Worker, rp RetryPolicy) {
	r.RegisterKindWithIdempotency(kind, w, Idempotency{Kind: IdempotencyDomainCAS}, rp)
}

// RegisterKindWithIdempotency binds a worker, idempotency declaration, and retry policy to a
// job kind. Registering the same kind twice, an empty kind, a nil worker, or a
// worker without exactly one declared duplicate-safety mechanism records an
// error surfaced by Err(). A zero-value RetryPolicy is filled from DefaultRetry
// so a caller can register with just a worker and idempotency.
func (r *Registry) RegisterKindWithIdempotency(kind string, w Worker, idem Idempotency, rp RetryPolicy) {
	if r.sealed {
		panic("jobs: job-kind registration after boot: the extension model is sealed")
	}
	if kind == "" {
		r.errs = append(r.errs, kerr.E(kerr.KindInternal, "invalid_kind",
			"RegisterKind requires a non-empty kind"))
		return
	}
	if w == nil {
		r.errs = append(r.errs, kerr.E(kerr.KindInternal, "invalid_worker",
			"RegisterKind: worker for kind "+kind+" is nil"))
		return
	}
	if err := idem.Validate(); err != nil {
		r.errs = append(r.errs, kerr.E(kerr.KindInternal, "invalid_idempotency",
			"RegisterKind: kind "+kind+": "+err.Error()))
		return
	}
	if _, dup := r.kinds[kind]; dup {
		r.errs = append(r.errs, kerr.E(kerr.KindInternal, "duplicate_kind",
			"job kind "+kind+" registered more than once"))
		return
	}
	if rp.MaxAttempts <= 0 {
		rp.MaxAttempts = defaultMaxAttempts
	}
	if rp.Backoff == nil {
		rp.Backoff = ExpJitterBackoff
	}
	r.kinds[kind] = entry{worker: w, idempotency: idem, retry: rp}
}

// lookup returns the entry for a kind.
func (r *Registry) lookup(kind string) (entry, bool) {
	e, ok := r.kinds[kind]
	return e, ok
}

// Err returns the accumulated registration errors joined into one, or nil. Boot
// calls this after all modules have registered and refuses to start on error.
func (r *Registry) Err() error {
	if len(r.errs) == 0 {
		return nil
	}
	msg := r.errs[0].Error()
	for i := 1; i < len(r.errs); i++ {
		msg += "; " + r.errs[i].Error()
	}
	return kerr.E(kerr.KindInternal, "registration_failed", "job registration failed: "+msg)
}
