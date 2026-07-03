package jobs

import (
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// entry is a registered kind: the worker that executes it and its retry policy.
type entry struct {
	worker Worker
	retry  RetryPolicy
}

// Registry collects the (kind → worker + retry policy) bindings during module
// boot. It accumulates errors (duplicate kind, empty kind, nil worker) rather
// than panicking, so RegisterKind reads cleanly at call sites and boot fails
// once via Err() with every problem reported together (mirrors outbox's
// HandlerRegistry).
type Registry struct {
	kinds map[string]entry
	errs  []error
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{kinds: map[string]entry{}}
}

// RegisterKind binds a worker and retry policy to a job kind. Registering the
// same kind twice, an empty kind, or a nil worker records an error surfaced by
// Err(). A zero-value RetryPolicy is filled from DefaultRetry so a caller can
// register with just a worker.
func (r *Registry) RegisterKind(kind string, w Worker, rp RetryPolicy) {
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
	r.kinds[kind] = entry{worker: w, retry: rp}
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
