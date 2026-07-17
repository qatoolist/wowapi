package retention

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/v2/kernel/audit"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
)

// Per-class DSR status values.
const (
	ClassStatusExported      = "exported"
	ClassStatusErased        = "erased"
	ClassStatusNotApplicable = "not_applicable"
	ClassStatusPartial       = "partial"
	ClassStatusEmpty         = "empty"
)

// ClassResult records the outcome for one registered class in a DSR export.
type ClassResult struct {
	Status string         `json:"status"`
	Count  int            `json:"count,omitempty"`
	Data   map[string]any `json:"data,omitempty"`
}

// ArtifactManifest is the durable, auditable description of a DSR export.
type ArtifactManifest struct {
	RequestID       uuid.UUID              `json:"request_id"`
	CreatedAt       time.Time              `json:"created_at"`
	ExpiresAt       time.Time              `json:"expires_at"`
	Checksum        string                 `json:"checksum"`
	PerClassResults map[string]ClassResult `json:"per_class_results"`
	AccessPolicy    string                 `json:"access_policy"`
}

// ErasureResult reports per-class erasure statuses plus the total affected.
type ErasureResult struct {
	Total    int               `json:"total"`
	Statuses map[string]string `json:"statuses"`
}

// ArtifactWriter persists an encrypted DSR export artifact and decrypts it on
// download, recording an audit row for each download.
type ArtifactWriter interface {
	// Write encrypts the manifest and persists it. It returns the checksum of the
	// encrypted payload and the path/identifier needed to read it back.
	Write(ctx context.Context, db database.TenantDB, requestID uuid.UUID, manifest *ArtifactManifest) (checksum string, path string, err error)

	// Read decrypts the artifact at path and records a download-audit row.
	Read(ctx context.Context, db database.TenantDB, path string) ([]byte, error)
}

// FileArtifactWriter writes encrypted artifacts to the local filesystem.
type FileArtifactWriter struct {
	dir   string
	key   []byte
	audit *audit.Writer
}

// NewFileArtifactWriter creates a filesystem-backed artifact writer. key must
// be exactly 32 bytes for AES-256-GCM.
func NewFileArtifactWriter(dir string, key []byte, audit *audit.Writer) *FileArtifactWriter {
	return &FileArtifactWriter{dir: dir, key: key, audit: audit}
}

// TestKey returns a deterministic 32-byte AES key for tests. It may be
// overridden by the WOWAPI_DSR_ARTIFACT_TEST_KEY environment variable (hex).
func TestKey() []byte {
	if k := os.Getenv("WOWAPI_DSR_ARTIFACT_TEST_KEY"); k != "" {
		b, err := hex.DecodeString(k)
		if err == nil && len(b) == 32 {
			return b
		}
	}
	return []byte{
		0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11,
		0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11,
		0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11,
		0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11,
	}
}

type artifactEnvelope struct {
	RequestID uuid.UUID `json:"request_id"`
	Algorithm string    `json:"algorithm"`
	Checksum  string    `json:"checksum"`
	Data      string    `json:"data"`
}

// Write serializes the manifest, encrypts it with AES-256-GCM, writes a JSON
// envelope containing the SHA-256 checksum of the ciphertext, and records a
// creation-audit row. The manifest's Checksum field is set to the same
// checksum so the stored envelope is self-describing.
func (w *FileArtifactWriter) Write(ctx context.Context, db database.TenantDB, requestID uuid.UUID, manifest *ArtifactManifest) (string, string, error) {
	if len(w.key) != 32 {
		return "", "", kerr.E(kerr.KindInternal, "bad_key", "artifact key must be 32 bytes")
	}

	manifest.RequestID = requestID
	if manifest.CreatedAt.IsZero() {
		manifest.CreatedAt = time.Now().UTC()
	}
	if manifest.ExpiresAt.IsZero() {
		manifest.ExpiresAt = manifest.CreatedAt.Add(30 * 24 * time.Hour)
	}
	if manifest.AccessPolicy == "" {
		manifest.AccessPolicy = "tenant_admin_only"
	}

	// Serialize and encrypt the manifest. The checksum is over the encrypted
	// payload (the bytes actually written to storage) so it verifies the artifact
	// as stored.
	plain, err := json.Marshal(manifest)
	if err != nil {
		return "", "", kerr.Wrapf(err, "retention.FileArtifactWriter.Write", "marshal manifest")
	}

	ciphertext, err := encryptAESGCM(w.key, plain)
	if err != nil {
		return "", "", kerr.Wrapf(err, "retention.FileArtifactWriter.Write", "encrypt")
	}

	path := filepath.Join(w.dir, requestID.String()+".artifact")
	if err := os.MkdirAll(w.dir, 0o700); err != nil {
		return "", "", kerr.Wrapf(err, "retention.FileArtifactWriter.Write", "create dir")
	}
	checksum := sha256.Sum256(ciphertext)
	checksumHex := hex.EncodeToString(checksum[:])
	manifest.Checksum = checksumHex

	env := artifactEnvelope{
		RequestID: requestID,
		Algorithm: "sha256",
		Checksum:  checksumHex,
		Data:      base64.StdEncoding.EncodeToString(ciphertext),
	}
	b, err := json.Marshal(env)
	if err != nil {
		return "", "", kerr.Wrapf(err, "retention.FileArtifactWriter.Write", "marshal envelope")
	}
	if err := os.WriteFile(path, b, 0o600); err != nil {
		return "", "", kerr.Wrapf(err, "retention.FileArtifactWriter.Write", "write file")
	}

	if w.audit != nil {
		_ = w.audit.Record(ctx, db, audit.Entry{
			Action:     "dsr.artifact.created",
			EntityType: "dsr_request",
			EntityID:   requestID,
			Metadata: map[string]any{
				"path":     path,
				"checksum": checksumHex,
			},
		})
	}
	return checksumHex, path, nil
}

// Read decrypts the artifact at path and records a download-audit row.
func (w *FileArtifactWriter) Read(ctx context.Context, db database.TenantDB, path string) ([]byte, error) {
	if len(w.key) != 32 {
		return nil, kerr.E(kerr.KindInternal, "bad_key", "artifact key must be 32 bytes")
	}
	b, err := os.ReadFile(path) // #nosec G304 -- reads the artifact path previously returned by this writer
	if err != nil {
		return nil, kerr.Wrapf(err, "retention.FileArtifactWriter.Read", "read file")
	}
	var env artifactEnvelope
	if err := json.Unmarshal(b, &env); err != nil {
		return nil, kerr.Wrapf(err, "retention.FileArtifactWriter.Read", "unmarshal envelope")
	}
	ciphertext, err := base64.StdEncoding.DecodeString(env.Data)
	if err != nil {
		return nil, kerr.Wrapf(err, "retention.FileArtifactWriter.Read", "decode ciphertext")
	}
	plain, err := decryptAESGCM(w.key, ciphertext)
	if err != nil {
		return nil, kerr.Wrapf(err, "retention.FileArtifactWriter.Read", "decrypt")
	}

	if w.audit != nil {
		_ = w.audit.Record(ctx, db, audit.Entry{
			Action:     "dsr.artifact.download",
			EntityType: "dsr_request",
			Metadata: map[string]any{
				"path":     path,
				"checksum": env.Checksum,
			},
		})
	}
	return plain, nil
}

func encryptAESGCM(key, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

func decryptAESGCM(key, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return nil, errors.New("ciphertext too short")
	}
	nonce, ct := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	return gcm.Open(nil, nonce, ct, nil)
}

// holdID derives a deterministic UUID from a string identifier so legal holds
// can be placed on concepts (record-class keys, DSR subject refs) that are not
// themselves UUIDs.
func holdID(s string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("wowapi:hold:"+s))
}

// ErrHeld is returned by the Engine when a Dispose or Erase callback is
// blocked by a legal hold.
var ErrHeld = kerr.E(kerr.KindConflict, "held", "blocked by legal hold")
