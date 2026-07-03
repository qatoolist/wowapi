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

// PresignedURL is a time-boxed URL the client uses directly against the object
// store. Method is "PUT" (upload) or "GET" (download).
type PresignedURL struct {
	URL       string
	Method    string
	ExpiresAt time.Time
}

// ObjectInfo is what the store reports about a stored blob at confirm time.
// Checksum is the lowercase hex SHA-256 of the bytes.
type ObjectInfo struct {
	Size     int64
	Checksum string
}

// Adapter is the object-storage port.
//
// A tenant-prefixed Key ("<tenant>/<document>/<version>") is minted by the
// document service; the adapter treats keys as opaque. PresignPut/PresignGet
// return URLs valid for ttl. Stat and Peek support confirm-time verification
// (size + SHA-256 + a MIME-sniff prefix); a real adapter implements Stat via a
// HEAD and Peek via a ranged GET. Delete backs retention voiding.
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
