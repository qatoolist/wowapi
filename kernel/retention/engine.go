package retention

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/qatoolist/wowapi/internal/sealer"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/audit"
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
	sealed  bool
}

// NewRegistry builds an empty registry.
func NewRegistry() *Registry { return &Registry{classes: map[string]RecordClass{}} }

// Seal freezes the registry once boot validation completes: any later Register
// panics rather than silently adding a record class the boot gates never saw
// (closure review 2026-07-17, F-10).
// The sealer.Authority parameter restricts sealing to the framework's boot
// path: internal/sealer is unimportable outside the wowapi module, so a
// product module cannot prematurely seal a shared registry during Register.
func (r *Registry) Seal(sealer.Authority) { r.sealed = true }

// Register adds a record class. Keys must be non-empty and unique; the first
// error is retained and surfaced by Err (checked at boot).
func (r *Registry) Register(c RecordClass) {
	if r.sealed {
		panic("retention: record-class registration after boot: the extension model is sealed")
	}
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
	reg       *Registry
	dsr       *DSR
	holds     *Holds
	artifacts ArtifactWriter
	audit     *audit.Writer
}

// NewEngine preserves the v1 constructor. Export fulfilment uses the legacy
// in-transaction payload path until compliance dependencies are supplied with
// NewEngineWithCompliance.
func NewEngine(reg *Registry, dsr *DSR) *Engine {
	return &Engine{reg: reg, dsr: dsr}
}

// NewEngineWithCompliance wires the engine over a record-class registry, the DSR ledger, and
// the compliance wrappers. holds, artifacts, and audit may be nil in unit tests
// that do not exercise hold-blocking or artifact-writing paths; passing nil for
// artifacts makes RunExportDetailed fail closed.
func NewEngineWithCompliance(reg *Registry, dsr *DSR, holds *Holds, artifacts ArtifactWriter, audit *audit.Writer) *Engine {
	return &Engine{reg: reg, dsr: dsr, holds: holds, artifacts: artifacts, audit: audit}
}

// SweepDisposition runs each class's Dispose for records whose retention lapsed by
// `at`, in the caller's tenant transaction, returning the total disposed. Classes
// without a Dispose callback are skipped. A class under a record_class legal hold
// is blocked by the central wrapper.
func (e *Engine) SweepDisposition(ctx context.Context, db database.TenantDB, at time.Time) (int, error) {
	total := 0
	for _, c := range e.reg.ordered() {
		if c.Dispose == nil {
			continue
		}
		if e.holds != nil {
			held, err := e.holds.IsHeld(ctx, db, "record_class", holdID(c.Key))
			if err != nil {
				return total, kerr.Wrapf(err, "retention.SweepDisposition", "check hold for class %s", c.Key)
			}
			if held {
				return total, fmt.Errorf("%w: record class %s", ErrHeld, c.Key)
			}
		}
		n, err := c.Dispose(ctx, db, at)
		if err != nil {
			return total, kerr.Wrapf(err, "retention.SweepDisposition", "dispose class %s", c.Key)
		}
		total += n
	}
	return total, nil
}

// RunExportDetailed fulfils a pending export DSR: it invokes each class's Export for the
// subject, builds an encrypted artifact manifest with explicit per-class status,
// and completes the request only after the artifact is written. All work is in
// the caller's tenant tx, so a failure leaves the request pending and rolls back
// partial exports.
func (e *Engine) RunExportDetailed(ctx context.Context, db database.TenantDB, requestID uuid.UUID) (*ArtifactManifest, error) {
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

	manifest := &ArtifactManifest{
		RequestID:       requestID,
		CreatedAt:       time.Now().UTC(),
		ExpiresAt:       time.Now().UTC().Add(30 * 24 * time.Hour),
		AccessPolicy:    "tenant_admin_only",
		PerClassResults: map[string]ClassResult{},
	}
	for _, c := range e.reg.ordered() {
		if c.Export == nil {
			manifest.PerClassResults[c.Key] = ClassResult{Status: ClassStatusNotApplicable}
			continue
		}
		data, err := c.Export(ctx, db, req.SubjectRef)
		if err != nil {
			return nil, kerr.Wrapf(err, "retention.RunExport", "export class %s", c.Key)
		}
		if len(data) == 0 {
			manifest.PerClassResults[c.Key] = ClassResult{Status: ClassStatusEmpty}
		} else {
			manifest.PerClassResults[c.Key] = ClassResult{Status: ClassStatusExported, Data: data}
		}
	}

	if e.artifacts == nil {
		return nil, kerr.E(kerr.KindInternal, "no_artifact_writer", "artifact writer not configured")
	}
	checksum, _, err := e.artifacts.Write(ctx, db, requestID, manifest)
	if err != nil {
		return nil, kerr.Wrapf(err, "retention.RunExport", "write artifact")
	}
	manifest.Checksum = checksum

	if err := e.dsr.Complete(ctx, db, requestID); err != nil {
		return nil, err
	}
	return manifest, nil
}

// RunExport preserves the v1 API while using the fail-closed artifact path. The
// returned map contains the exported per-class payloads; callers that need the
// artifact checksum or explicit empty/not-applicable statuses should use
// RunExportDetailed.
func (e *Engine) RunExport(ctx context.Context, db database.TenantDB, requestID uuid.UUID) (map[string]any, error) {
	if e.artifacts == nil {
		return e.runExportLegacy(ctx, db, requestID)
	}
	manifest, err := e.RunExportDetailed(ctx, db, requestID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]any, len(manifest.PerClassResults))
	for key, result := range manifest.PerClassResults {
		if result.Status == ClassStatusExported || result.Status == ClassStatusEmpty {
			out[key] = result.Data
		}
	}
	return out, nil
}

func (e *Engine) runExportLegacy(ctx context.Context, db database.TenantDB, requestID uuid.UUID) (map[string]any, error) {
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

// RunErasureDetailed fulfils a pending erasure DSR: it invokes each class's Erase for
// the subject through the central legal-hold wrapper, marks the request
// completed, and returns a per-class status map plus the total records affected.
func (e *Engine) RunErasureDetailed(ctx context.Context, db database.TenantDB, requestID uuid.UUID) (*ErasureResult, error) {
	req, err := e.dsr.Get(ctx, db, requestID)
	if err != nil {
		return nil, err
	}
	if req.Kind != KindErasure {
		return nil, kerr.E(kerr.KindConflict, "wrong_kind", "DSR is not an erasure request")
	}
	if req.Status != "pending" {
		return nil, kerr.E(kerr.KindConflict, "not_pending", "DSR is not pending")
	}

	if e.holds != nil {
		held, err := e.holds.IsHeld(ctx, db, "dsr_subject", holdID(req.SubjectRef))
		if err != nil {
			return nil, kerr.Wrapf(err, "retention.RunErasure", "check subject hold")
		}
		if held {
			return nil, fmt.Errorf("%w: dsr subject %s", ErrHeld, req.SubjectRef)
		}
	}

	result := &ErasureResult{Statuses: map[string]string{}}
	for _, c := range e.reg.ordered() {
		if c.Erase == nil {
			result.Statuses[c.Key] = ClassStatusNotApplicable
			continue
		}
		n, err := c.Erase(ctx, db, req.SubjectRef)
		if err != nil {
			return result, kerr.Wrapf(err, "retention.RunErasure", "erase class %s", c.Key)
		}
		if n == 0 {
			result.Statuses[c.Key] = ClassStatusEmpty
		} else {
			result.Statuses[c.Key] = ClassStatusErased
		}
		result.Total += n
	}
	if err := e.dsr.Complete(ctx, db, requestID); err != nil {
		return result, err
	}
	return result, nil
}

// RunErasure preserves the v1 API. Call RunErasureDetailed when per-class
// statuses are required.
func (e *Engine) RunErasure(ctx context.Context, db database.TenantDB, requestID uuid.UUID) (int, error) {
	result, err := e.RunErasureDetailed(ctx, db, requestID)
	if result == nil {
		return 0, err
	}
	return result.Total, err
}
