package document

import "context"

// UploadEvent is passed to OnFileUpload hooks after a version's bytes are
// verified but before the version row is committed. A hook returning an error
// aborts the confirm (the version is not written). The canonical hook enqueues
// an async malware scan; the version lands scan_status=pending and downloads of
// confidential+ documents block until the scan clears it.
type UploadEvent struct {
	DocumentID  string
	Class       string
	VersionNo   int
	StorageKey  string
	MIME        string
	SizeBytes   int64
	Sensitivity Sensitivity
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
	sealed   bool
}

// NewHooks returns an empty hook set.
func NewHooks() *Hooks { return &Hooks{} }

// Seal freezes the hook set once boot validation completes: any later
// registration panics rather than silently attaching a hook the boot gates
// never saw (closure review 2026-07-17, F-10).
func (h *Hooks) Seal() { h.sealed = true }

func (h *Hooks) mustBeUnsealed() {
	if h.sealed {
		panic("document: hook registration after boot: the extension model is sealed")
	}
}

// OnFileUpload registers a confirm-time hook.
func (h *Hooks) OnFileUpload(fn UploadHook) {
	h.mustBeUnsealed()
	h.onUpload = append(h.onUpload, fn)
}

// OnDocumentAccess registers a download-time hook.
func (h *Hooks) OnDocumentAccess(fn AccessHook) {
	h.mustBeUnsealed()
	h.onAccess = append(h.onAccess, fn)
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
