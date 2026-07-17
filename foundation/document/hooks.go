package document

import (
	"context"
	"errors"

	"github.com/qatoolist/wowapi/internal/sealer"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// UploadEvent is passed to OnFileUpload hooks after a version's bytes are
// verified but before the version row is committed. A hook returning an error
// aborts the confirm (the version is not written). The canonical hook enqueues
// an async malware scan; the version lands scan_status=pending and downloads of
// confidential+ documents block until the scan clears it.
//
// UploadEvent's field set is FROZEN at its v1 shape (the seven fields below):
// consumers write unkeyed composite literals, and adding any field is a
// source-incompatible change for them (third closure audit 2026-07-17 — the
// same compatibility class as app.Hook). The transactional effect contract is
// delivered through the context instead: see UploadDeliveryFromContext.
//
// Effect contract (second closure audit 2026-07-17, F-05): the hook runs
// INSIDE the confirming transaction, which can still roll back after the hook
// returns (a later insert/update/outbox write or the commit itself can fail),
// and the same reserved upload is then retryable — so a hook effect is exactly
// once ONLY if it is written through the delivery's Tx (it commits and rolls
// back atomically with the confirmation; the canonical scan enqueue is an
// outbox write through it). An effect delivered OUTSIDE the transaction may be
// re-delivered on retry and MUST be idempotent keyed on the delivery's
// DeliveryID, which is stable across retries of the same reserved upload.
type UploadEvent struct {
	DocumentID  string
	Class       string
	VersionNo   int
	StorageKey  string
	MIME        string
	SizeBytes   int64
	Sensitivity Sensitivity
}

// UploadDelivery is the transactional execution context of one OnFileUpload
// invocation, carried on the hook's context (NOT on UploadEvent, whose v1
// field set is frozen for unkeyed-literal compatibility). A domain event
// describes what happened; the delivery carries the execution capabilities.
type UploadDelivery struct {
	// DeliveryID is the durable idempotency identifier for this upload's hook
	// effects: the upload session's id, identical on every retry of the same
	// reserved (document, version, key) confirmation. External (non-Tx) effects
	// must deduplicate on it.
	DeliveryID string
	// Tx is the confirming transaction's tenant handle. Effects written through
	// it are atomic with the confirmation: they are never visible if the
	// confirmation rolls back, and land exactly once when it commits.
	Tx database.TenantDB
}

type uploadDeliveryKey struct{}

// withUploadDelivery binds the confirming transaction's delivery context for
// the duration of the hook invocations.
func withUploadDelivery(ctx context.Context, d UploadDelivery) context.Context {
	return context.WithValue(ctx, uploadDeliveryKey{}, d)
}

// UploadDeliveryFromContext returns the transactional delivery context of the
// current OnFileUpload invocation: the retry-stable DeliveryID and the
// confirming transaction's tenant handle. It reports false outside a hook
// invocation. Delivered via the context so UploadEvent's frozen v1 shape stays
// source-compatible for unkeyed composite literals.
func UploadDeliveryFromContext(ctx context.Context) (UploadDelivery, bool) {
	d, ok := ctx.Value(uploadDeliveryKey{}).(UploadDelivery)
	return d, ok
}

// AccessEvent is passed to OnDocumentAccess hooks after authorization succeeds
// and before the presigned GET is minted. A hook returning an error denies the
// download. The watermark slot lives here.
type AccessEvent struct {
	DocumentID  string
	VersionNo   int
	Sensitivity Sensitivity
	ActorID     string
}

// UploadHook runs on confirm; AccessHook runs on download.
type (
	UploadHook func(context.Context, UploadEvent) error
	AccessHook func(context.Context, AccessEvent) error
)

// Hooks is the registry of upload/access hooks a module wires at boot.
type Hooks struct {
	onUpload []UploadHook
	onAccess []AccessHook
	errs     []error
	sealed   bool
}

// NewHooks returns an empty hook set.
func NewHooks() *Hooks { return &Hooks{} }

// Seal freezes the hook set once boot validation completes: any later
// registration panics rather than silently attaching a hook the boot gates
// never saw (closure review 2026-07-17, F-10).
// The sealer.Authority parameter restricts sealing to the framework's boot
// path: internal/sealer is unimportable outside the wowapi module, so a
// product module cannot prematurely seal a shared registry during Register.
func (h *Hooks) Seal(sealer.Authority) { h.sealed = true }

func (h *Hooks) mustBeUnsealed() {
	if h.sealed {
		panic("document: hook registration after boot: the extension model is sealed")
	}
}

// OnFileUpload registers a confirm-time hook. A nil hook is a collected boot
// error (second closure audit 2026-07-17, F-10): it would otherwise panic only
// when the first confirmation invokes it.
func (h *Hooks) OnFileUpload(fn UploadHook) {
	h.mustBeUnsealed()
	if fn == nil {
		h.errs = append(h.errs, kerr.E(kerr.KindInternal, "invalid_hook",
			"OnFileUpload registered a nil hook"))
		return
	}
	h.onUpload = append(h.onUpload, fn)
}

// OnDocumentAccess registers a download-time hook. A nil hook is a collected
// boot error, like OnFileUpload.
func (h *Hooks) OnDocumentAccess(fn AccessHook) {
	h.mustBeUnsealed()
	if fn == nil {
		h.errs = append(h.errs, kerr.E(kerr.KindInternal, "invalid_hook",
			"OnDocumentAccess registered a nil hook"))
		return
	}
	h.onAccess = append(h.onAccess, fn)
}

// Err returns accumulated registration errors joined, or nil; app.Boot gates
// on it like every other registry Err.
func (h *Hooks) Err() error {
	return errors.Join(h.errs...)
}

func (h *Hooks) runUpload(ctx context.Context, e UploadEvent) error {
	if h == nil {
		return nil
	}
	for _, fn := range h.onUpload {
		if err := fn(ctx, e); err != nil {
			return err
		}
	}
	return nil
}

func (h *Hooks) runAccess(ctx context.Context, e AccessEvent) error {
	if h == nil {
		return nil
	}
	for _, fn := range h.onAccess {
		if err := fn(ctx, e); err != nil {
			return err
		}
	}
	return nil
}
