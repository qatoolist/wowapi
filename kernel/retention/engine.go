package retention

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// The disposition/DSR engine turns the retention primitives (holds, DSR ledger)
// into a working per-record-class lifecycle (roadmap E2). Because the framework
// must not run dynamic SQL over arbitrary product tables, each record class
// registers callbacks the engine orchestrates: Dispose (delete/anonymize records
// past retention), Export and Erase (fulfil DSRs). The framework owns the
// scheduling, the DSR ledger transitions, and the ordering; the product owns the
// data access — the same registry+callback shape used elsewhere in the kernel.

// DisposeFunc disposes the class's records whose retention lapsed on/before
// `before` — deleting or anonymizing them — and returns how many. It must itself
// skip records under legal hold (consult Holds.IsHeld). Runs in the caller's
// tenant transaction.
type DisposeFunc func(ctx context.Context, db database.TenantDB, before time.Time) (int, error)

// ExportFunc returns the class's data for a DSR subject (for a data-portability
// export). Runs in the caller's tenant transaction.
type ExportFunc func(ctx context.Context, db database.TenantDB, subjectRef string) (map[string]any, error)

// EraseFunc erases (or anonymizes) the class's data for a DSR subject and returns
// how many records were affected. Runs in the caller's tenant transaction.
type EraseFunc func(ctx context.Context, db database.TenantDB, subjectRef string) (int, error)

// RecordClass declares one class of product data and how to dispose/export/erase
// it. Any callback may be nil (a class with no Export contributes nothing to an
// export, etc.).
type RecordClass struct {
	Key       string
	Retention time.Duration // documentary; the Dispose callback enforces it
	Dispose   DisposeFunc
	Export    ExportFunc
	Erase     EraseFunc
}

// Registry is the boot-time catalog of record classes.
type Registry struct {
	classes map[string]RecordClass
	err     error
}

// NewRegistry builds an empty registry.
func NewRegistry() *Registry { return &Registry{classes: map[string]RecordClass{}} }

// Register adds a record class. Keys must be non-empty and unique; the first
// error is retained and surfaced by Err (checked at boot).
func (r *Registry) Register(c RecordClass) {
	if r.err != nil {
		return
	}
	if c.Key == "" {
		r.err = kerr.E(kerr.KindInternal, "invalid_record_class", "record class key is required")
		return
	}
	if _, dup := r.classes[c.Key]; dup {
		r.err = kerr.E(kerr.KindInternal, "duplicate_record_class", "record class already registered: "+c.Key)
		return
	}
	r.classes[c.Key] = c
}

// Err returns the first registration error, if any.
func (r *Registry) Err() error { return r.err }

func (r *Registry) ordered() []RecordClass {
	keys := make([]string, 0, len(r.classes))
	for k := range r.classes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]RecordClass, 0, len(keys))
	for _, k := range keys {
		out = append(out, r.classes[k])
	}
	return out
}

// Engine orchestrates disposition and DSR fulfilment over the registered classes.
type Engine struct {
	reg *Registry
	dsr *DSR
}

// NewEngine wires the engine over a record-class registry and the DSR ledger.
func NewEngine(reg *Registry, dsr *DSR) *Engine {
	return &Engine{reg: reg, dsr: dsr}
}

// SweepDisposition runs each class's Dispose for records whose retention lapsed by
// `at`, in the caller's tenant transaction, returning the total disposed. Classes
// without a Dispose callback are skipped. Intended to be driven periodically by
// the scheduler.
func (e *Engine) SweepDisposition(ctx context.Context, db database.TenantDB, at time.Time) (int, error) {
	total := 0
	for _, c := range e.reg.ordered() {
		if c.Dispose == nil {
			continue
		}
		n, err := c.Dispose(ctx, db, at)
		if err != nil {
			return total, kerr.Wrapf(err, "retention.SweepDisposition", "dispose class %s", c.Key)
		}
		total += n
	}
	return total, nil
}

// RunExport fulfils a pending export DSR: it invokes each class's Export for the
// subject, aggregates the results by class key, marks the request completed, and
// returns the payload. All work is in the caller's tenant tx, so a failure leaves
// the request pending and rolls back partial exports.
func (e *Engine) RunExport(ctx context.Context, db database.TenantDB, requestID uuid.UUID) (map[string]any, error) {
	req, err := e.dsr.Get(ctx, db, requestID)
	if err != nil {
		return nil, err
	}
	if req.Kind != KindExport {
		return nil, kerr.E(kerr.KindConflict, "wrong_kind", "DSR is not an export request")
	}
	if req.Status != "pending" {
		return nil, kerr.E(kerr.KindConflict, "not_pending", "DSR is not pending")
	}
	out := map[string]any{}
	for _, c := range e.reg.ordered() {
		if c.Export == nil {
			continue
		}
		data, err := c.Export(ctx, db, req.SubjectRef)
		if err != nil {
			return nil, kerr.Wrapf(err, "retention.RunExport", "export class %s", c.Key)
		}
		out[c.Key] = data
	}
	if err := e.dsr.Complete(ctx, db, requestID); err != nil {
		return nil, err
	}
	return out, nil
}

// RunErasure fulfils a pending erasure DSR: it invokes each class's Erase for the
// subject and marks the request completed, returning the total records affected.
// A legal hold that forbids erasure must be enforced by the product's Erase
// callback (skip held records) or by rejecting the DSR first with a reason.
func (e *Engine) RunErasure(ctx context.Context, db database.TenantDB, requestID uuid.UUID) (int, error) {
	req, err := e.dsr.Get(ctx, db, requestID)
	if err != nil {
		return 0, err
	}
	if req.Kind != KindErasure {
		return 0, kerr.E(kerr.KindConflict, "wrong_kind", "DSR is not an erasure request")
	}
	if req.Status != "pending" {
		return 0, kerr.E(kerr.KindConflict, "not_pending", "DSR is not pending")
	}
	total := 0
	for _, c := range e.reg.ordered() {
		if c.Erase == nil {
			continue
		}
		n, err := c.Erase(ctx, db, req.SubjectRef)
		if err != nil {
			return total, kerr.Wrapf(err, "retention.RunErasure", "erase class %s", c.Key)
		}
		total += n
	}
	if err := e.dsr.Complete(ctx, db, requestID); err != nil {
		return total, err
	}
	return total, nil
}
