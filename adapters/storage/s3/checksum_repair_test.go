package s3_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	s3adapter "github.com/qatoolist/wowapi/v2/adapters/storage/s3"
	"github.com/qatoolist/wowapi/v2/kernel/observability"
	"github.com/qatoolist/wowapi/v2/kernel/storage"
)

type requestCounter struct {
	base http.RoundTripper
	mu   sync.Mutex
	gets int
}

func (r *requestCounter) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method == http.MethodGet {
		r.mu.Lock()
		r.gets++
		r.mu.Unlock()
	}
	return r.base.RoundTrip(req)
}

func (r *requestCounter) reset() {
	r.mu.Lock()
	r.gets = 0
	r.mu.Unlock()
}

func (r *requestCounter) getCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.gets
}

func putLegacyObject(t *testing.T, key string, body []byte) {
	t.Helper()
	cfg := testConfig()
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		t.Fatalf("legacy client: %v", err)
	}
	url, err := client.PresignedPutObject(context.Background(), cfg.Bucket, key, time.Minute)
	if err != nil {
		t.Fatalf("legacy presign: %v", err)
	}
	req, err := http.NewRequest(http.MethodPut, url.String(), bytes.NewReader(body))
	if err != nil {
		t.Fatalf("legacy request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("legacy upload: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		payload, _ := io.ReadAll(resp.Body)
		t.Fatalf("legacy upload = %d: %s", resp.StatusCode, payload)
	}
}

func TestS3CanonicalUploadStatNeverDownloadsBody(t *testing.T) {
	counter := &requestCounter{base: http.DefaultTransport}
	cfg := testConfig()
	cfg.Transport = counter
	a := requireMinio(t, cfg)
	ctx := context.Background()
	key := testKey(t)
	body := []byte("canonical checksum upload")
	defer func() { _ = a.Delete(ctx, key) }()

	want := sha256.Sum256(body)
	put, err := a.PresignPutChecksum(ctx, key, hex.EncodeToString(want[:]), time.Minute)
	if err != nil {
		t.Fatalf("PresignPutChecksum: %v", err)
	}
	if put.Headers["X-Amz-Checksum-Algorithm"] != "SHA256" || put.Headers["X-Amz-Checksum-Sha256"] == "" {
		t.Fatalf("checksum headers = %#v, want signed SHA256 metadata", put.Headers)
	}
	httpPut(t, put, body)
	counter.reset()

	info, err := a.Stat(ctx, key)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	want = sha256.Sum256(body)
	if info.Checksum != hex.EncodeToString(want[:]) {
		t.Fatalf("checksum = %q, want canonical SHA-256", info.Checksum)
	}
	if got := counter.getCount(); got != 0 {
		t.Fatalf("normal Stat issued %d GetObject requests, want 0", got)
	}
}

func TestS3LegacyObjectHashesOnlyThroughBoundedRepair(t *testing.T) {
	counter := &requestCounter{base: http.DefaultTransport}
	cfg := testConfig()
	cfg.Transport = counter
	a := requireMinio(t, cfg)
	ctx := context.Background()
	key := testKey(t)
	body := []byte("legacy object")
	putLegacyObject(t, key, body)
	defer func() { _ = a.Delete(ctx, key) }()

	counter.reset()
	if _, err := a.Stat(ctx, key); err == nil {
		t.Fatal("Stat legacy object succeeded, want checksum-missing error")
	}
	if got := counter.getCount(); got != 0 {
		t.Fatalf("ambient Stat issued %d GetObject requests, want 0", got)
	}

	info, err := a.RepairChecksum(ctx, key, storage.RepairOptions{
		Label:    "legacy-backfill",
		MaxBytes: int64(len(body)),
		Timeout:  time.Second,
	})
	if err != nil {
		t.Fatalf("RepairChecksum: %v", err)
	}
	want := sha256.Sum256(body)
	if info.Checksum != hex.EncodeToString(want[:]) {
		t.Fatalf("repaired checksum = %q, want SHA-256", info.Checksum)
	}
	if got := counter.getCount(); got != 1 {
		t.Fatalf("labeled repair issued %d GetObject requests, want 1", got)
	}

	counter.reset()
	if _, err := a.Stat(ctx, key); err != nil {
		t.Fatalf("Stat after repair: %v", err)
	}
	if got := counter.getCount(); got != 0 {
		t.Fatalf("Stat after repair issued %d GetObject requests, want 0", got)
	}
}

type metricRecorder struct {
	counters   map[string]float64
	histograms map[string][]float64
}

func (m *metricRecorder) ObserveRequest(string, string, int, time.Duration, int) {}
func (m *metricRecorder) SetGauge(string, float64, map[string]string)            {}
func (m *metricRecorder) IncCounter(name string, value float64, _ map[string]string) {
	m.counters[name] += value
}

func (m *metricRecorder) ObserveHistogram(name string, value float64, _ map[string]string) {
	m.histograms[name] = append(m.histograms[name], value)
}

var _ observability.Metrics = (*metricRecorder)(nil)

func TestS3RepairEmitsHitByteAndDurationMetrics(t *testing.T) {
	metrics := &metricRecorder{
		counters:   make(map[string]float64),
		histograms: make(map[string][]float64),
	}
	cfg := testConfig()
	cfg.Metrics = metrics
	a := requireMinio(t, cfg)
	ctx := context.Background()
	key := testKey(t)
	body := []byte("legacy metric payload")
	putLegacyObject(t, key, body)
	defer func() { _ = a.Delete(ctx, key) }()

	_, err := a.RepairChecksum(ctx, key, storage.RepairOptions{
		Label: "metric-test", MaxBytes: int64(len(body)), Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("RepairChecksum: %v", err)
	}
	if metrics.counters["storage_checksum_repair_hits_total"] != 1 {
		t.Fatalf("repair hits = %v, want 1", metrics.counters)
	}
	bytesObserved := metrics.histograms["storage_checksum_repair_bytes"]
	if len(bytesObserved) != 1 || bytesObserved[0] != float64(len(body)) {
		t.Fatalf("repair bytes = %v, want [%d]", bytesObserved, len(body))
	}
	durations := metrics.histograms["storage_checksum_repair_duration_seconds"]
	if len(durations) != 1 || durations[0] <= 0 {
		t.Fatalf("repair durations = %v, want one positive observation", durations)
	}
}

func TestS3ChecksumBackfillInterruptResumeNoDuplicates(t *testing.T) {
	metrics := &metricRecorder{
		counters:   make(map[string]float64),
		histograms: make(map[string][]float64),
	}
	cfg := testConfig()
	cfg.Metrics = metrics
	a := requireMinio(t, cfg)
	ctx := context.Background()
	prefix := testKey(t) + "/"
	for i := range 3 {
		key := fmt.Sprintf("%s%02d", prefix, i)
		putLegacyObject(t, key, []byte(fmt.Sprintf("legacy-%d", i)))
		defer func() { _ = a.Delete(ctx, key) }()
	}

	first, err := a.BackfillChecksums(ctx, s3adapter.BackfillOptions{
		Prefix:     prefix,
		MaxObjects: 1,
		Repair:     storage.RepairOptions{Label: "legacy-backfill", MaxBytes: 64, Timeout: time.Second},
	})
	if err != nil {
		t.Fatalf("first backfill batch: %v", err)
	}
	if first.Repaired != 1 || first.Complete || first.Cursor == "" {
		t.Fatalf("first batch = %+v, want one repair and resumable cursor", first)
	}

	second, err := a.BackfillChecksums(ctx, s3adapter.BackfillOptions{
		Prefix:     prefix,
		Cursor:     first.Cursor,
		MaxObjects: 10,
		Repair:     storage.RepairOptions{Label: "legacy-backfill", MaxBytes: 64, Timeout: time.Second},
	})
	if err != nil {
		t.Fatalf("resumed backfill: %v", err)
	}
	if second.Repaired != 2 || !second.Complete {
		t.Fatalf("resumed batch = %+v, want two repairs and completion", second)
	}

	restarted, err := a.BackfillChecksums(ctx, s3adapter.BackfillOptions{
		Prefix:     prefix,
		MaxObjects: 10,
		Repair:     storage.RepairOptions{Label: "legacy-backfill", MaxBytes: 64, Timeout: time.Second},
	})
	if err != nil {
		t.Fatalf("restart after lost cursor: %v", err)
	}
	if restarted.Repaired != 0 || !restarted.Complete {
		t.Fatalf("idempotent restart = %+v, want no duplicate repair", restarted)
	}
	if got := metrics.counters["storage_checksum_repair_hits_total"]; got != 3 {
		t.Fatalf("full-hash repairs = %v, want exactly 3 unique objects", got)
	}
}

func TestS3RepairRejectsUnlabeledAndOversizedWork(t *testing.T) {
	a := requireMinio(t, testConfig())
	ctx := context.Background()
	key := testKey(t)
	body := []byte("legacy object exceeding repair bound")
	putLegacyObject(t, key, body)
	defer func() { _ = a.Delete(ctx, key) }()

	if _, err := a.RepairChecksum(ctx, key, storage.RepairOptions{MaxBytes: int64(len(body)), Timeout: time.Second}); err == nil {
		t.Fatal("unlabeled repair succeeded")
	}
	if _, err := a.RepairChecksum(ctx, key, storage.RepairOptions{Label: "manual-repair", MaxBytes: 4, Timeout: time.Second}); err == nil {
		t.Fatal("oversized repair succeeded")
	}
}
