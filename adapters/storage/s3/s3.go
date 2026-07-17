// Package s3 is the framework's production object-storage adapter: the
// storage.Adapter port (kernel/storage) implemented against S3-compatible
// endpoints (AWS S3, MinIO) with the minio-go SDK.
//
// Every framework-issued upload signs S3's SHA-256 checksum algorithm header,
// so Stat can verify integrity from HEAD metadata without downloading the body.
// Legacy objects without canonical checksum metadata fail normal Stat closed;
// only the explicitly labeled, size/time-bounded RepairChecksum path may hash
// their bytes.
//   - Stat/Peek/PresignGet of an absent key return errors.KindNotFound.
//   - Delete is idempotent: deleting an absent key is a clean no-op.
//   - Presigned URLs are short-lived: a caller-requested ttl <= 0 or above the
//     configured Config.PresignTTL is clamped to Config.PresignTTL.
package s3

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/observability"
	"github.com/qatoolist/wowapi/v2/kernel/storage"
)

// defaultPresignTTL matches the kernel document service's put TTL.
const defaultPresignTTL = 15 * time.Minute

// Config is everything the adapter needs to talk to an S3-compatible store.
type Config struct {
	// Endpoint is host:port, optionally with an explicit http:// or https://
	// scheme (the devbox compose exports S3_ENDPOINT=http://minio:9000). When a
	// scheme is present it decides TLS and UseSSL is ignored.
	Endpoint string
	// Bucket holds every object; keys are minted by the kernel document
	// service and treated as opaque by this adapter.
	Bucket string
	// Region of the bucket. Empty is fine for MinIO.
	Region string
	// AccessKey / SecretKey are static credentials. Callers are expected to
	// resolve these from their own secret store before constructing Config.
	AccessKey string
	SecretKey string
	// UseSSL connects over TLS when Endpoint carries no explicit scheme.
	UseSSL bool
	// PresignTTL is the default AND upper bound for presigned URL validity.
	// Zero means defaultPresignTTL (15m, the kernel's own put TTL).
	PresignTTL time.Duration
	// Transport overrides MinIO's HTTP transport. It is primarily useful for
	// request accounting in integration tests; nil uses the SDK default.
	Transport http.RoundTripper
	// Metrics receives labeled legacy-repair hit, byte, and duration samples.
	// Nil uses observability.NoOp.
	Metrics observability.Metrics
	// CreateBucket makes New create the bucket when absent (local/dev overlays
	// only). Default false: production buckets are provisioned out of band,
	// and New fails closed at boot when the bucket is missing.
	CreateBucket bool
}

// Adapter implements storage.Adapter against an S3-compatible endpoint using
// path-style addressing (bucket in the path, as MinIO requires).
type Adapter struct {
	client     *minio.Client
	bucket     string
	presignTTL time.Duration
	metrics    observability.Metrics
}

// The port is the contract; fail the build, not the boot, if it drifts.
var (
	_ storage.Adapter          = (*Adapter)(nil)
	_ storage.ChecksumRepairer = (*Adapter)(nil)
	_ storage.ChecksumUploader = (*Adapter)(nil)
)

// New validates cfg, builds the client, and verifies the bucket — a missing or
// unreachable bucket fails boot closed rather than surfacing as 500s on the
// first upload. When cfg.CreateBucket is set the bucket is created if absent.
func New(ctx context.Context, cfg Config) (*Adapter, error) {
	var errs []error
	add := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }
	if cfg.Endpoint == "" {
		add("storage: endpoint required")
	}
	if cfg.Bucket == "" {
		add("storage: bucket required")
	}
	if cfg.AccessKey == "" {
		add("storage: access key required")
	}
	if cfg.SecretKey == "" {
		add("storage: secret key required")
	}
	if cfg.PresignTTL < 0 || cfg.PresignTTL > 7*24*time.Hour {
		add("storage: presign TTL %s invalid (0 = default %s; S3 caps presigned URLs at 7 days)",
			cfg.PresignTTL, defaultPresignTTL)
	}
	if err := errors.Join(errs...); err != nil {
		return nil, err
	}

	host, secure, err := splitEndpoint(cfg.Endpoint, cfg.UseSSL)
	if err != nil {
		return nil, err
	}
	ttl := cfg.PresignTTL
	if ttl == 0 {
		ttl = defaultPresignTTL
	}

	client, err := minio.New(host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: secure,
		Region: cfg.Region,
		// Path-style addressing: MinIO serves buckets under the path, and
		// virtual-host style needs wildcard DNS we cannot assume.
		BucketLookup: minio.BucketLookupPath,
		Transport:    cfg.Transport,
	})
	if err != nil {
		return nil, fmt.Errorf("storage: client: %w", err)
	}

	metrics := cfg.Metrics
	if metrics == nil {
		metrics = observability.NoOp
	}
	a := &Adapter{client: client, bucket: cfg.Bucket, presignTTL: ttl, metrics: metrics}
	if err := a.ensureBucket(ctx, cfg.CreateBucket, cfg.Region); err != nil {
		return nil, err
	}
	return a, nil
}

// splitEndpoint accepts "host:port", "http://host:port" or "https://host:port"
// and returns the bare host plus whether to use TLS.
func splitEndpoint(endpoint string, useSSL bool) (host string, secure bool, err error) {
	switch {
	case strings.HasPrefix(endpoint, "http://"), strings.HasPrefix(endpoint, "https://"):
		u, perr := url.Parse(endpoint)
		if perr != nil {
			return "", false, fmt.Errorf("storage: endpoint %q: %w", endpoint, perr)
		}
		if u.Host == "" {
			return "", false, fmt.Errorf("storage: endpoint %q has no host", endpoint)
		}
		return u.Host, u.Scheme == "https", nil
	case strings.Contains(endpoint, "://"):
		return "", false, fmt.Errorf("storage: endpoint %q: unsupported scheme (want http, https, or none)", endpoint)
	default:
		return endpoint, useSSL, nil
	}
}

func (a *Adapter) ensureBucket(ctx context.Context, create bool, region string) error {
	exists, err := a.client.BucketExists(ctx, a.bucket)
	if err != nil {
		return fmt.Errorf("storage: probing bucket %q: %w", a.bucket, err)
	}
	if exists {
		return nil
	}
	if !create {
		return fmt.Errorf("storage: bucket %q does not exist (set storage.create_bucket only in local/dev; provision production buckets out of band)", a.bucket)
	}
	if err := a.client.MakeBucket(ctx, a.bucket, minio.MakeBucketOptions{Region: region}); err != nil {
		// Tolerate the create/create race: someone else won.
		code := minio.ToErrorResponse(err).Code
		if code == "BucketAlreadyOwnedByYou" || code == "BucketAlreadyExists" {
			return nil
		}
		return fmt.Errorf("storage: creating bucket %q: %w", a.bucket, err)
	}
	return nil
}

// clampTTL applies the short-TTL policy: non-positive means "the configured
// default", anything longer than the configured TTL is clamped down to it.
func (a *Adapter) clampTTL(ttl time.Duration) time.Duration {
	if ttl <= 0 || ttl > a.presignTTL {
		return a.presignTTL
	}
	return ttl
}

// PresignPut rejects unsigned-checksum uploads. It remains only for Adapter
// source compatibility; framework callers use PresignPutChecksum.
func (a *Adapter) PresignPut(_ context.Context, _ string, _ time.Duration) (storage.PresignedURL, error) {
	return storage.PresignedURL{}, kerr.E(kerr.KindValidation, "upload_checksum_required", "S3 uploads require PresignPutChecksum")
}

// PresignPutChecksum returns a checksum-enforcing presigned PUT. The client
// must copy every returned header; S3 validates and persists the SHA-256.
func (a *Adapter) PresignPutChecksum(ctx context.Context, key, checksumSHA256 string, ttl time.Duration) (storage.PresignedURL, error) {
	rawChecksum, err := hex.DecodeString(checksumSHA256)
	if err != nil || len(rawChecksum) != sha256.Size || checksumSHA256 != strings.ToLower(checksumSHA256) {
		return storage.PresignedURL{}, kerr.E(kerr.KindValidation, "invalid_upload_checksum", "upload checksum must be lowercase-hex SHA-256")
	}
	checksumB64 := base64.StdEncoding.EncodeToString(rawChecksum)
	ttl = a.clampTTL(ttl)
	headers := http.Header{
		"X-Amz-Checksum-Algorithm": []string{"SHA256"},
		"X-Amz-Checksum-Sha256":    []string{checksumB64},
	}
	u, err := a.client.PresignHeader(ctx, http.MethodPut, a.bucket, key, ttl, nil, headers)
	if err != nil {
		return storage.PresignedURL{}, kerr.Wrapf(err, "storage.s3.PresignPut", "presign put %q", key)
	}
	return storage.PresignedURL{
		URL: u.String(), Method: http.MethodPut,
		Headers: map[string]string{
			"X-Amz-Checksum-Algorithm": "SHA256",
			"X-Amz-Checksum-Sha256":    checksumB64,
		},
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}

// PresignGet returns a presigned GET URL for an EXISTING object (mirroring
// Memory: presigning a download of an absent key is KindNotFound, not a URL
// that 404s later).
func (a *Adapter) PresignGet(ctx context.Context, key string, ttl time.Duration) (storage.PresignedURL, error) {
	if _, err := a.client.StatObject(ctx, a.bucket, key, minio.StatObjectOptions{}); err != nil {
		if isNotFound(err) {
			return storage.PresignedURL{}, objectNotFound(key)
		}
		return storage.PresignedURL{}, kerr.Wrapf(err, "storage.s3.PresignGet", "stat %q", key)
	}
	ttl = a.clampTTL(ttl)
	u, err := a.client.PresignedGetObject(ctx, a.bucket, key, ttl, nil)
	if err != nil {
		return storage.PresignedURL{}, kerr.Wrapf(err, "storage.s3.PresignGet", "presign get %q", key)
	}
	return storage.PresignedURL{URL: u.String(), Method: http.MethodGet, ExpiresAt: time.Now().Add(ttl)}, nil
}

// Stat reports size and canonical SHA-256 from HEAD metadata only. Missing
// metadata is an explicit legacy-object error; normal reads never hash bytes.
func (a *Adapter) Stat(ctx context.Context, key string) (storage.ObjectInfo, error) {
	info, err := a.client.StatObject(ctx, a.bucket, key, minio.StatObjectOptions{Checksum: true})
	if err != nil {
		if isNotFound(err) {
			return storage.ObjectInfo{}, objectNotFound(key)
		}
		return storage.ObjectInfo{}, kerr.Wrapf(err, "storage.s3.Stat", "stat %q", key)
	}
	if checksum, ok := checksumFromInfo(info); ok {
		return storage.ObjectInfo{Size: info.Size, Checksum: checksum}, nil
	}
	return storage.ObjectInfo{}, kerr.E(
		kerr.KindConflict,
		"storage_checksum_missing",
		"storage object lacks canonical checksum metadata; use labeled checksum repair: "+key,
	)
}

// RepairChecksum hashes one legacy object under explicit resource bounds and
// persists the result as storage-owned metadata via an in-place server copy.
func (a *Adapter) RepairChecksum(ctx context.Context, key string, opts storage.RepairOptions) (storage.ObjectInfo, error) {
	if strings.TrimSpace(opts.Label) == "" {
		return storage.ObjectInfo{}, kerr.E(kerr.KindValidation, "repair_label_required", "checksum repair label is required")
	}
	if opts.MaxBytes <= 0 {
		return storage.ObjectInfo{}, kerr.E(kerr.KindValidation, "repair_max_bytes_required", "checksum repair max bytes must be positive")
	}
	if opts.Timeout <= 0 {
		return storage.ObjectInfo{}, kerr.E(kerr.KindValidation, "repair_timeout_required", "checksum repair timeout must be positive")
	}

	info, err := a.client.StatObject(ctx, a.bucket, key, minio.StatObjectOptions{Checksum: true})
	if err != nil {
		if isNotFound(err) {
			return storage.ObjectInfo{}, objectNotFound(key)
		}
		return storage.ObjectInfo{}, kerr.Wrapf(err, "storage.s3.RepairChecksum", "stat %q", key)
	}
	if checksum, ok := checksumFromInfo(info); ok {
		return storage.ObjectInfo{Size: info.Size, Checksum: checksum}, nil
	}
	if info.Size > opts.MaxBytes {
		return storage.ObjectInfo{}, kerr.E(kerr.KindValidation, "repair_object_too_large", "storage object exceeds checksum repair byte bound")
	}

	labels := map[string]string{"label": opts.Label}
	a.metrics.IncCounter("storage_checksum_repair_hits_total", 1, labels)
	started := time.Now()
	defer func() {
		observability.ObserveHistogram(a.metrics, "storage_checksum_repair_duration_seconds", time.Since(started).Seconds(), labels)
	}()

	repairCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()
	obj, err := a.client.GetObject(repairCtx, a.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return storage.ObjectInfo{}, kerr.Wrapf(err, "storage.s3.RepairChecksum", "get %q", key)
	}
	defer func() { _ = obj.Close() }()
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(obj, opts.MaxBytes+1))
	if err != nil {
		if isNotFound(err) {
			return storage.ObjectInfo{}, objectNotFound(key)
		}
		return storage.ObjectInfo{}, kerr.Wrapf(err, "storage.s3.RepairChecksum", "hash %q", key)
	}
	if n > opts.MaxBytes {
		return storage.ObjectInfo{}, kerr.E(kerr.KindValidation, "repair_object_too_large", "storage object exceeds checksum repair byte bound")
	}
	observability.ObserveHistogram(a.metrics, "storage_checksum_repair_bytes", float64(n), labels)
	checksum := hex.EncodeToString(h.Sum(nil))
	metadata := make(map[string]string, len(info.UserMetadata)+1)
	for name, value := range info.UserMetadata {
		metadata[name] = value
	}
	metadata[repairChecksumMetadata] = checksum
	_, err = a.client.CopyObject(repairCtx,
		minio.CopyDestOptions{
			Bucket: a.bucket, Object: key,
			UserMetadata: metadata, ReplaceMetadata: true,
		},
		minio.CopySrcOptions{Bucket: a.bucket, Object: key, MatchETag: info.ETag},
	)
	if err != nil {
		return storage.ObjectInfo{}, kerr.Wrapf(err, "storage.s3.RepairChecksum", "persist checksum %q", key)
	}
	return storage.ObjectInfo{Size: n, Checksum: checksum}, nil
}

// Peek returns up to n leading bytes (ranged GET) for MIME sniffing.
func (a *Adapter) Peek(ctx context.Context, key string, n int) ([]byte, error) {
	if n <= 0 {
		// Existence still matters: Peek of an absent key must be KindNotFound.
		if _, err := a.Stat(ctx, key); err != nil {
			return nil, err
		}
		return []byte{}, nil
	}
	opts := minio.GetObjectOptions{}
	if err := opts.SetRange(0, int64(n)-1); err != nil {
		return nil, kerr.Wrapf(err, "storage.s3.Peek", "range %q", key)
	}
	obj, err := a.client.GetObject(ctx, a.bucket, key, opts)
	if err != nil {
		return nil, kerr.Wrapf(err, "storage.s3.Peek", "get %q", key)
	}
	defer func() { _ = obj.Close() }()
	data, err := io.ReadAll(io.LimitReader(obj, int64(n)))
	if err != nil {
		switch {
		case isNotFound(err):
			return nil, objectNotFound(key)
		case minio.ToErrorResponse(err).Code == "InvalidRange":
			// Zero-byte object: any range is unsatisfiable; the prefix is empty.
			return []byte{}, nil
		default:
			return nil, kerr.Wrapf(err, "storage.s3.Peek", "read %q", key)
		}
	}
	return data, nil
}

// Delete removes key; deleting an absent key is a clean no-op (S3 DeleteObject
// semantics, matching Memory — retention voiding must be idempotent).
func (a *Adapter) Delete(ctx context.Context, key string) error {
	if err := a.client.RemoveObject(ctx, a.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		if isNotFound(err) {
			return nil
		}
		return kerr.Wrapf(err, "storage.s3.Delete", "delete %q", key)
	}
	return nil
}

const repairChecksumMetadata = "wowapi-sha256"

func checksumFromInfo(info minio.ObjectInfo) (string, bool) {
	if checksum, ok := decodeSHA256Checksum(info.ChecksumSHA256); ok {
		return checksum, true
	}
	return decodeHexSHA256(info.Metadata.Get("X-Amz-Meta-" + repairChecksumMetadata))
}

func decodeHexSHA256(value string) (string, bool) {
	if len(value) != sha256.Size*2 {
		return "", false
	}
	raw, err := hex.DecodeString(value)
	if err != nil || len(raw) != sha256.Size {
		return "", false
	}
	return strings.ToLower(value), true
}

// decodeSHA256Checksum converts the base64 SHA-256 S3 reports in HeadObject
// metadata to the lowercase hex the storage port speaks. ok is false for
// missing or malformed metadata.
func decodeSHA256Checksum(b64 string) (hexSum string, ok bool) {
	if b64 == "" {
		return "", false
	}
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil || len(raw) != sha256.Size {
		return "", false
	}
	return hex.EncodeToString(raw), true
}

// isNotFound recognizes the S3 shapes of "no such object".
func isNotFound(err error) bool {
	resp := minio.ToErrorResponse(err)
	return resp.Code == "NoSuchKey" || resp.StatusCode == http.StatusNotFound
}

// objectNotFound mirrors the framework storage package's sentinel exactly
// (same Kind, code, and message shape); callers match on KindOf == KindNotFound.
func objectNotFound(key string) error {
	return kerr.E(kerr.KindNotFound, "object_not_found", "storage object not found: "+key)
}
