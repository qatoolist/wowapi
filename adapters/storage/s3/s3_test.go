// Integration tests against a REAL S3-compatible endpoint (the wowapi devbox
// compose: localhost:9000, root user wowapi / wowapi-local-only). Gated like
// the framework's DB tests: an unreachable store skips — unless
// WOWAPI_REQUIRE_S3=1 (CI), where skipping is a failure; storage tests must
// never silently skip out of that gate.
//
// Proves the storage.Adapter port beyond storage.NewMemory: the full
// presigned round trip over real HTTP, the confirm-time verification
// semantics (sha256, KindNotFound, idempotent delete), the configured-expiry
// clamp on presigned URLs, and product-shaped concurrency across keys.
package s3_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/storage"

	s3adapter "github.com/qatoolist/wowapi/adapters/storage/s3"
)

const testBucket = "wowapi-storage-it"

func testEndpoint() string {
	if v := os.Getenv("S3_TEST_ENDPOINT"); v != "" {
		return v
	}
	return "localhost:9000"
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// testConfig is the adapter config every integration test starts from.
func testConfig() s3adapter.Config {
	return s3adapter.Config{
		Endpoint:     testEndpoint(),
		Bucket:       testBucket,
		AccessKey:    envOr("S3_ACCESS_KEY", "wowapi"),
		SecretKey:    envOr("S3_SECRET_KEY", "wowapi-local-only"),
		UseSSL:       false,
		PresignTTL:   5 * time.Minute,
		CreateBucket: true, // throwaway integration bucket
	}
}

// requireMinio gates on a reachable S3-compatible endpoint the same way
// testkit's requireDB gates on Postgres, then returns a constructed adapter
// (New itself is under test past the gate — its failure is a real failure).
func requireMinio(t *testing.T, cfg s3adapter.Config) *s3adapter.Adapter {
	t.Helper()
	conn, err := net.DialTimeout("tcp", testEndpoint(), 3*time.Second)
	if err != nil {
		if os.Getenv("WOWAPI_REQUIRE_S3") == "1" {
			t.Fatalf("WOWAPI_REQUIRE_S3=1 but S3/minio unreachable at %s: %v", testEndpoint(), err)
		}
		t.Skipf("S3/minio unreachable at %s (%v); set WOWAPI_REQUIRE_S3=1 to make this a failure", testEndpoint(), err)
	}
	_ = conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	a, err := s3adapter.New(ctx, cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return a
}

func testKey(t *testing.T) string {
	return "it/" + t.Name() + "/" + uuid.NewString()
}

// httpPut PUTs body to a presigned URL — the client leg of the presign flow.
func httpPut(t *testing.T, presigned storage.PresignedURL, body []byte) {
	t.Helper()
	if presigned.Method != http.MethodPut {
		t.Fatalf("presigned method = %q, want PUT", presigned.Method)
	}
	req, err := http.NewRequest(http.MethodPut, presigned.URL, bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	for name, value := range presigned.Headers {
		req.Header.Set(name, value)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("PUT to presigned URL: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("presigned PUT = %d\n%s", resp.StatusCode, b)
	}
}

// amzExpires extracts the X-Amz-Expires seconds a presigned URL carries.
func amzExpires(t *testing.T, rawURL string) int {
	t.Helper()
	u, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse presigned URL: %v", err)
	}
	v := u.Query().Get("X-Amz-Expires")
	if v == "" {
		t.Fatalf("presigned URL carries no X-Amz-Expires: %s", rawURL)
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		t.Fatalf("X-Amz-Expires %q: %v", v, err)
	}
	return n
}

// The core round trip: presigned PUT → Stat (size + sha256) → presigned GET
// download → Peek → Delete → gone (and Delete stays idempotent).
func TestS3_PresignedRoundTrip(t *testing.T) {
	a := requireMinio(t, testConfig())
	ctx := context.Background()
	key := testKey(t)
	body := []byte("wowapi storage adapter payload — round trip over real minio")
	sum := sha256.Sum256(body)

	put, err := a.PresignPutChecksum(ctx, key, hex.EncodeToString(sum[:]), 0) // 0 → configured default TTL
	if err != nil {
		t.Fatalf("PresignPut: %v", err)
	}
	if got, want := amzExpires(t, put.URL), int((5 * time.Minute).Seconds()); got != want {
		t.Errorf("presigned PUT X-Amz-Expires = %d, want the configured %d", got, want)
	}
	httpPut(t, put, body)

	info, err := a.Stat(ctx, key)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size != int64(len(body)) {
		t.Errorf("Stat size = %d, want %d", info.Size, len(body))
	}
	if want := hex.EncodeToString(sum[:]); info.Checksum != want {
		t.Errorf("Stat checksum = %s, want %s (lowercase hex sha256 of the bytes)", info.Checksum, want)
	}

	get, err := a.PresignGet(ctx, key, 30*time.Second)
	if err != nil {
		t.Fatalf("PresignGet: %v", err)
	}
	if get.Method != http.MethodGet {
		t.Errorf("presigned GET method = %q", get.Method)
	}
	if got := amzExpires(t, get.URL); got != 30 {
		t.Errorf("presigned GET X-Amz-Expires = %d, want the requested 30", got)
	}
	resp, err := http.Get(get.URL)
	if err != nil {
		t.Fatalf("GET presigned URL: %v", err)
	}
	downloaded, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil || resp.StatusCode != http.StatusOK {
		t.Fatalf("download = %d, err %v", resp.StatusCode, err)
	}
	if !bytes.Equal(downloaded, body) {
		t.Fatalf("downloaded bytes differ: got %d bytes", len(downloaded))
	}

	prefix, err := a.Peek(ctx, key, 10)
	if err != nil {
		t.Fatalf("Peek: %v", err)
	}
	if !bytes.Equal(prefix, body[:10]) {
		t.Errorf("Peek = %q, want %q", prefix, body[:10])
	}
	// Peek past the end returns the whole (short) object, like Memory.
	whole, err := a.Peek(ctx, key, len(body)+512)
	if err != nil {
		t.Fatalf("Peek(oversized): %v", err)
	}
	if !bytes.Equal(whole, body) {
		t.Errorf("oversized Peek = %d bytes, want the full %d", len(whole), len(body))
	}

	if err := a.Delete(ctx, key); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := a.Stat(ctx, key); kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("Stat after Delete = %v, want KindNotFound", err)
	}
	if err := a.Delete(ctx, key); err != nil {
		t.Fatalf("Delete of a missing key must be a no-op, got %v", err)
	}
}

// Wrong-checksum path: what the document service's confirm compares is Stat's
// checksum of the STORED bytes — a lying declared checksum can never match it.
// (The full confirm-time rejection runs in document_e2e_test.go.)
func TestS3_WrongChecksum_StatReportsTrueBytes(t *testing.T) {
	a := requireMinio(t, testConfig())
	ctx := context.Background()
	key := testKey(t)
	body := []byte("actual uploaded bytes")
	defer func() { _ = a.Delete(ctx, key) }()

	truth := sha256.Sum256(body)
	put, err := a.PresignPutChecksum(ctx, key, hex.EncodeToString(truth[:]), time.Minute)
	if err != nil {
		t.Fatalf("PresignPut: %v", err)
	}
	httpPut(t, put, body)

	info, err := a.Stat(ctx, key)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	lying := hex.EncodeToString(make([]byte, sha256.Size)) // all-zero "checksum"
	if info.Checksum == lying {
		t.Fatal("Stat checksum equals the lying declared checksum — confirm could not reject tampering")
	}
	if info.Checksum != hex.EncodeToString(truth[:]) {
		t.Fatalf("Stat checksum = %s, want the sha256 of the stored bytes", info.Checksum)
	}
}

// Absent keys are KindNotFound on every read path (the confirm flow maps
// Stat's KindNotFound to its "upload_missing" validation error).
func TestS3_MissingObject_IsKindNotFound(t *testing.T) {
	a := requireMinio(t, testConfig())
	ctx := context.Background()
	key := testKey(t) // never uploaded

	if _, err := a.Stat(ctx, key); kerr.KindOf(err) != kerr.KindNotFound {
		t.Errorf("Stat(missing) = %v, want KindNotFound", err)
	}
	if _, err := a.Peek(ctx, key, 512); kerr.KindOf(err) != kerr.KindNotFound {
		t.Errorf("Peek(missing) = %v, want KindNotFound", err)
	}
	if _, err := a.PresignGet(ctx, key, time.Minute); kerr.KindOf(err) != kerr.KindNotFound {
		t.Errorf("PresignGet(missing) = %v, want KindNotFound", err)
	}
}

// The configured TTL is an upper bound: a caller asking for longer is clamped
// down, a caller asking for shorter keeps its shorter expiry.
func TestS3_PresignTTL_ConfiguredExpiryClamp(t *testing.T) {
	a := requireMinio(t, testConfig()) // PresignTTL: 5m
	ctx := context.Background()
	key := testKey(t)

	empty := sha256.Sum256(nil)
	over, err := a.PresignPutChecksum(ctx, key, hex.EncodeToString(empty[:]), 24*time.Hour)
	if err != nil {
		t.Fatalf("PresignPut: %v", err)
	}
	if got, want := amzExpires(t, over.URL), int((5 * time.Minute).Seconds()); got != want {
		t.Errorf("over-long request: X-Amz-Expires = %d, want clamp to configured %d", got, want)
	}
	under, err := a.PresignPutChecksum(ctx, key, hex.EncodeToString(empty[:]), 45*time.Second)
	if err != nil {
		t.Fatalf("PresignPut: %v", err)
	}
	if got := amzExpires(t, under.URL); got != 45 {
		t.Errorf("shorter request: X-Amz-Expires = %d, want the requested 45", got)
	}
	if !over.ExpiresAt.After(time.Now()) {
		t.Error("ExpiresAt is not in the future")
	}
}

// A zero-byte upload must Stat/Peek cleanly (empty prefix, not an error): the
// ranged GET behind Peek gets an InvalidRange from S3 on empty objects.
func TestS3_EmptyObject(t *testing.T) {
	a := requireMinio(t, testConfig())
	ctx := context.Background()
	key := testKey(t)
	defer func() { _ = a.Delete(ctx, key) }()

	empty := sha256.Sum256(nil)
	put, err := a.PresignPutChecksum(ctx, key, hex.EncodeToString(empty[:]), time.Minute)
	if err != nil {
		t.Fatalf("PresignPut: %v", err)
	}
	httpPut(t, put, nil)

	info, err := a.Stat(ctx, key)
	if err != nil {
		t.Fatalf("Stat(empty): %v", err)
	}
	empty = sha256.Sum256(nil)
	if info.Size != 0 || info.Checksum != hex.EncodeToString(empty[:]) {
		t.Errorf("Stat(empty) = %+v, want size 0 + sha256 of no bytes", info)
	}
	prefix, err := a.Peek(ctx, key, 512)
	if err != nil {
		t.Fatalf("Peek(empty): %v", err)
	}
	if len(prefix) != 0 {
		t.Errorf("Peek(empty) = %d bytes, want 0", len(prefix))
	}
}

// New fails closed when the bucket does not exist and CreateBucket is unset —
// production buckets are provisioned out of band, so a boot-time typo or a
// not-yet-provisioned bucket must fail the boot, not surface as 500s later.
func TestS3_New_MissingBucketFailsClosedWithoutCreateBucket(t *testing.T) {
	cfg := testConfig()
	cfg.Bucket = "wowapi-storage-it-missing-" + uuid.NewString()
	cfg.CreateBucket = false

	conn, err := net.DialTimeout("tcp", testEndpoint(), 3*time.Second)
	if err != nil {
		if os.Getenv("WOWAPI_REQUIRE_S3") == "1" {
			t.Fatalf("WOWAPI_REQUIRE_S3=1 but S3/minio unreachable at %s: %v", testEndpoint(), err)
		}
		t.Skipf("S3/minio unreachable at %s (%v); set WOWAPI_REQUIRE_S3=1 to make this a failure", testEndpoint(), err)
	}
	_ = conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if _, err := s3adapter.New(ctx, cfg); err == nil {
		t.Fatal("New succeeded against a nonexistent bucket with CreateBucket=false, want a fail-closed error")
	}
}

// New with CreateBucket=true provisions a fresh bucket that does not yet
// exist (the local/dev overlay path), and the adapter is immediately usable.
func TestS3_New_CreateBucketProvisionsFreshBucket(t *testing.T) {
	cfg := testConfig()
	cfg.Bucket = "wowapi-storage-it-fresh-" + uuid.NewString()
	cfg.CreateBucket = true

	a := requireMinio(t, cfg)
	ctx := context.Background()
	key := testKey(t)
	body := []byte("hello fresh bucket")
	sum := sha256.Sum256(body)
	put, err := a.PresignPutChecksum(ctx, key, hex.EncodeToString(sum[:]), time.Minute)
	if err != nil {
		t.Fatalf("PresignPutChecksum against freshly created bucket: %v", err)
	}
	httpPut(t, put, body)
	if _, err := a.Stat(ctx, key); err != nil {
		t.Fatalf("Stat against freshly created bucket: %v", err)
	}
}

// Concurrent New calls with CreateBucket=true against the SAME not-yet-existing
// bucket race on MakeBucket; the loser must tolerate "already exists" rather
// than fail (New's ensureBucket race-tolerance branch).
func TestS3_New_ConcurrentCreateBucketRaceIsTolerated(t *testing.T) {
	conn, err := net.DialTimeout("tcp", testEndpoint(), 3*time.Second)
	if err != nil {
		if os.Getenv("WOWAPI_REQUIRE_S3") == "1" {
			t.Fatalf("WOWAPI_REQUIRE_S3=1 but S3/minio unreachable at %s: %v", testEndpoint(), err)
		}
		t.Skipf("S3/minio unreachable at %s (%v); set WOWAPI_REQUIRE_S3=1 to make this a failure", testEndpoint(), err)
	}
	_ = conn.Close()

	cfg := testConfig()
	cfg.Bucket = "wowapi-storage-it-race-" + uuid.NewString()
	cfg.CreateBucket = true

	const n = 5
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			_, errs[i] = s3adapter.New(ctx, cfg)
		}()
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("New (racer %d): %v, want every racer to tolerate the create race", i, err)
		}
	}
}

// Peek(n<=0) still enforces existence: KindNotFound for an absent key, and an
// empty (non-nil-erroring) prefix for one that exists.
func TestS3_Peek_NonPositiveN(t *testing.T) {
	a := requireMinio(t, testConfig())
	ctx := context.Background()

	missing := testKey(t)
	if _, err := a.Peek(ctx, missing, 0); kerr.KindOf(err) != kerr.KindNotFound {
		t.Errorf("Peek(missing, 0) = %v, want KindNotFound", err)
	}

	present := testKey(t)
	defer func() { _ = a.Delete(ctx, present) }()
	body := []byte("some bytes")
	sum := sha256.Sum256(body)
	put, err := a.PresignPutChecksum(ctx, present, hex.EncodeToString(sum[:]), time.Minute)
	if err != nil {
		t.Fatalf("PresignPutChecksum: %v", err)
	}
	httpPut(t, put, body)
	prefix, err := a.Peek(ctx, present, 0)
	if err != nil {
		t.Fatalf("Peek(present, 0): %v", err)
	}
	if len(prefix) != 0 {
		t.Errorf("Peek(present, 0) = %d bytes, want 0", len(prefix))
	}
	if _, err := a.Peek(ctx, present, -5); err != nil {
		t.Fatalf("Peek(present, -5): %v", err)
	}
}

// Product-shaped concurrency (first-use rule): parallel presign→PUT→Stat→
// Delete across distinct keys, the way concurrent document uploads land.
func TestS3_ConcurrentRoundTrips(t *testing.T) {
	a := requireMinio(t, testConfig())
	ctx := context.Background()

	const n = 8
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = func() error {
				key := fmt.Sprintf("it/%s/%d-%s", t.Name(), i, uuid.NewString())
				body := bytes.Repeat([]byte{byte('a' + i)}, 1024*(i+1))
				sum := sha256.Sum256(body)
				put, err := a.PresignPutChecksum(ctx, key, hex.EncodeToString(sum[:]), time.Minute)
				if err != nil {
					return fmt.Errorf("presign %d: %w", i, err)
				}
				req, err := http.NewRequest(http.MethodPut, put.URL, bytes.NewReader(body))
				if err != nil {
					return err
				}
				for name, value := range put.Headers {
					req.Header.Set(name, value)
				}
				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return fmt.Errorf("put %d: %w", i, err)
				}
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					return fmt.Errorf("put %d: status %d", i, resp.StatusCode)
				}
				info, err := a.Stat(ctx, key)
				if err != nil {
					return fmt.Errorf("stat %d: %w", i, err)
				}
				if info.Size != int64(len(body)) || info.Checksum != hex.EncodeToString(sum[:]) {
					return fmt.Errorf("stat %d mismatch: %+v", i, info)
				}
				if err := a.Delete(ctx, key); err != nil {
					return fmt.Errorf("delete %d: %w", i, err)
				}
				if _, err := a.Stat(ctx, key); kerr.KindOf(err) != kerr.KindNotFound {
					return fmt.Errorf("stat-after-delete %d: %v, want KindNotFound", i, err)
				}
				return nil
			}()
		}()
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("worker %d: %v", i, err)
		}
	}
}
