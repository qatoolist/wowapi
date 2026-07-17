package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/qatoolist/wowapi/kernel/database"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// Anchor is an externally-published audit-chain head.
type Anchor struct {
	Seq        int64     `json:"seq"`
	HeadHash   string    `json:"head_hash"`
	AnchoredAt time.Time `json:"anchored_at"`
}

// ExternalStore persists anchors outside the database and returns the latest
// anchor for verification. Implementations may target a public timestamping
// service, an append-only log, object storage, or (for tests) a local file.
type ExternalStore interface {
	// PublishAnchor writes a new anchor. It must be durable enough that a later
	// LatestAnchor for the same tenant returns it.
	PublishAnchor(ctx context.Context, db database.TenantDB, a Anchor) error

	// LatestAnchor returns the most recently published anchor for the tenant.
	// If no anchor exists, it returns a not-found error.
	LatestAnchor(ctx context.Context, db database.TenantDB) (Anchor, error)
}

// FileStore is a file-backed ExternalStore for tests and local development.
// Each tenant gets its own anchor file under Dir.
type FileStore struct {
	Dir string
}

// NewFileStore creates a file-backed anchor store rooted at dir.
func NewFileStore(dir string) *FileStore {
	return &FileStore{Dir: dir}
}

func (s *FileStore) tenantFile(ctx context.Context) (string, error) {
	tid, ok := database.TenantIDFrom(ctx)
	if !ok {
		return "", kerr.E(kerr.KindValidation, "no_tenant", "tenant id required for external anchor")
	}
	return filepath.Join(s.Dir, tid.String()+".anchor"), nil
}

// PublishAnchor writes the anchor as JSON, overwriting the tenant's previous
// file. Overwrite is acceptable because LatestAnchor only needs the most
// recent anchor; historical anchors are preserved by the caller if required.
func (s *FileStore) PublishAnchor(ctx context.Context, _ database.TenantDB, a Anchor) error {
	path, err := s.tenantFile(ctx)
	if err != nil {
		return err
	}
	b, err := json.Marshal(a)
	if err != nil {
		return kerr.Wrapf(err, "audit.FileStore.PublishAnchor", "marshal anchor")
	}
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return kerr.Wrapf(err, "audit.FileStore.PublishAnchor", "create dir")
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return kerr.Wrapf(err, "audit.FileStore.PublishAnchor", "write anchor file")
	}
	return nil
}

// LatestAnchor reads the tenant's anchor file.
func (s *FileStore) LatestAnchor(ctx context.Context, _ database.TenantDB) (Anchor, error) {
	path, err := s.tenantFile(ctx)
	if err != nil {
		return Anchor{}, err
	}
	b, err := os.ReadFile(path) // #nosec G304 -- path is derived from the configured anchor directory and context tenant
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Anchor{}, kerr.E(kerr.KindNotFound, "no_anchor", "no external anchor found")
		}
		return Anchor{}, kerr.Wrapf(err, "audit.FileStore.LatestAnchor", "read anchor file")
	}
	var a Anchor
	if err := json.Unmarshal(b, &a); err != nil {
		return Anchor{}, kerr.Wrapf(err, "audit.FileStore.LatestAnchor", "unmarshal anchor")
	}
	return a, nil
}

// ExternalAnchor periodically publishes the audit-chain head to an ExternalStore
// and verifies the live chain against the published anchor.
type ExternalAnchor struct {
	store  ExternalStore
	writer *Writer
}

// NewExternalAnchor builds an external anchoring service.
func NewExternalAnchor(store ExternalStore, writer *Writer) *ExternalAnchor {
	return &ExternalAnchor{store: store, writer: writer}
}

// AnchorNow records the tenant's current audit-chain head and publishes it to
// the external store. It also writes an audit row so the anchor action itself
// is part of the chain.
func (ea *ExternalAnchor) AnchorNow(ctx context.Context, db database.TenantDB) error {
	seq, hash, err := ea.writer.Anchor(ctx, db)
	if err != nil {
		return err
	}
	if seq <= 0 {
		// Nothing to anchor; treat as a no-op rather than an error.
		return nil
	}

	a := Anchor{
		Seq:        seq,
		HeadHash:   hash,
		AnchoredAt: time.Now().UTC(),
	}
	if err := ea.store.PublishAnchor(ctx, db, a); err != nil {
		return kerr.Wrapf(err, "audit.ExternalAnchor.AnchorNow", "publish anchor")
	}

	return ea.writer.Record(ctx, db, Entry{
		Action:     "audit.external_anchor",
		EntityType: "audit_chain",
		Metadata: map[string]any{
			"anchor_seq":  seq,
			"anchor_hash": hash,
		},
	})
}

// ErrAnchorTampered is returned when Verify detects that the live audit chain
// no longer contains the externally-anchored head.
var ErrAnchorTampered = kerr.E(kerr.KindConflict, "anchor_tampered", "audit chain tampered: anchored head missing")

// Verify fetches the latest external anchor and confirms the local chain still
// contains it using Writer.CheckAnchor. A detectable tamper error is returned
// when the anchor is missing or the chain head has been rewound.
func (ea *ExternalAnchor) Verify(ctx context.Context, db database.TenantDB) error {
	a, err := ea.store.LatestAnchor(ctx, db)
	if err != nil {
		return err
	}
	present, err := ea.writer.CheckAnchor(ctx, db, a.Seq, a.HeadHash)
	if err != nil {
		return kerr.Wrapf(err, "audit.ExternalAnchor.Verify", "check anchor")
	}
	if !present {
		return fmt.Errorf("%w: anchor seq=%d not present in live chain", ErrAnchorTampered, a.Seq)
	}
	return nil
}
