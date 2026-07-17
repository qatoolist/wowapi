// Package storage is the object-storage port for the document framework: the
// kernel never talks to S3/minio/GCS directly, it talks to an Adapter. Uploads
// and downloads go through short-lived presigned URLs so blob bytes never
// transit the API process. The memory adapter (NewMemory) backs tests and local
// dev; a production adapter (S3/minio) implements the same four methods.
package storage

import (
	"context"
	"time"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// PresignedURL is a time-boxed request the client sends directly to storage.
// Headers are part of the signature and must be copied verbatim by the client.
type PresignedURL struct {
	URL       string
	Method    string
	Headers   map[string]string
	ExpiresAt time.Time
}

// ObjectInfo is what the store reports about a stored blob at confirm time.
// Checksum is the lowercase hex SHA-256 of the bytes.
type ObjectInfo struct {
	Size     int64
	Checksum string
}

// RepairOptions bounds the exceptional full-body checksum repair path.
// Label is required so repair work is explicit and attributable.
type RepairOptions struct {
	Label    string
	MaxBytes int64
	Timeout  time.Duration
}

// ChecksumRepairer is an optional storage capability for legacy objects.
// It is deliberately separate from Adapter so normal Stat cannot fall back to
// body hashing and existing storage adapters remain source-compatible.
type ChecksumRepairer interface {
	RepairChecksum(ctx context.Context, key string, opts RepairOptions) (ObjectInfo, error)
}

// ChecksumUploader is the checksum-enforcing upload capability used by the
// document framework. It is optional on Adapter to preserve third-party
// adapter compatibility; framework upload initiation fails closed without it.
type ChecksumUploader interface {
	PresignPutChecksum(ctx context.Context, key, checksumSHA256 string, ttl time.Duration) (PresignedURL, error)
}

// Adapter is the object-storage port.
//
// A tenant-prefixed Key ("<tenant>/<document>/<version>") is minted by the
// document service; the adapter treats keys as opaque. PresignPut is retained
// for adapter compatibility; framework uploads require ChecksumUploader.
type Adapter interface {
	PresignPut(ctx context.Context, key string, ttl time.Duration) (PresignedURL, error)
	PresignGet(ctx context.Context, key string, ttl time.Duration) (PresignedURL, error)
	Stat(ctx context.Context, key string) (ObjectInfo, error)
	// Peek returns up to n leading bytes for MIME sniffing (http.DetectContentType
	// reads at most 512). Returns KindNotFound if the object is absent.
	Peek(ctx context.Context, key string, n int) ([]byte, error)
	Delete(ctx context.Context, key string) error
}

// ErrObjectNotFound is the sentinel kind adapters return when a key is absent;
// callers match with errors.KindOf(err) == errors.KindNotFound.
func objectNotFound(key string) error {
	return kerr.E(kerr.KindNotFound, "object_not_found", "storage object not found: "+key)
}
