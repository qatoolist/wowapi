package s3

import (
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"

	"github.com/qatoolist/wowapi/kernel/storage"
)

// BackfillOptions describes one bounded, resumable legacy-checksum batch.
// Cursor is the last successfully classified or repaired key from a prior run.
type BackfillOptions struct {
	Prefix     string
	Cursor     string
	MaxObjects int
	Repair     storage.RepairOptions
}

// BackfillResult is the durable progress a caller persists between batches.
type BackfillResult struct {
	Cursor   string
	Scanned  int
	Repaired int
	Complete bool
}

// BackfillChecksums inventories real bucket objects and repairs at most
// MaxObjects legacy entries. Progress advances only after an object is either
// already canonical or successfully repaired. Replaying an older cursor is
// safe because RepairChecksum detects canonical metadata before hashing.
func (a *Adapter) BackfillChecksums(ctx context.Context, opts BackfillOptions) (BackfillResult, error) {
	result := BackfillResult{Cursor: opts.Cursor}
	if opts.MaxObjects <= 0 {
		return result, fmt.Errorf("checksum backfill max objects must be positive")
	}

	objects := a.client.ListObjects(ctx, a.bucket, minio.ListObjectsOptions{
		Prefix: opts.Prefix, Recursive: true, StartAfter: opts.Cursor,
	})
	for object := range objects {
		if object.Err != nil {
			return result, fmt.Errorf("inventory legacy checksum objects after %q: %w", result.Cursor, object.Err)
		}
		result.Scanned++

		info, err := a.client.StatObject(ctx, a.bucket, object.Key, minio.StatObjectOptions{Checksum: true})
		if err != nil {
			return result, fmt.Errorf("inventory checksum metadata for %q: %w", object.Key, err)
		}
		if _, ok := checksumFromInfo(info); ok {
			result.Cursor = object.Key
			continue
		}
		if result.Repaired == opts.MaxObjects {
			return result, nil
		}
		if _, err := a.RepairChecksum(ctx, object.Key, opts.Repair); err != nil {
			return result, fmt.Errorf("repair checksum for %q after %q: %w", object.Key, result.Cursor, err)
		}
		result.Cursor = object.Key
		result.Repaired++
		if result.Repaired == opts.MaxObjects {
			return result, nil
		}
	}
	result.Complete = true
	return result, nil
}
