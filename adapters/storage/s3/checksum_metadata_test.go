package s3_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sync"
	"testing"
	"time"
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
	if info.Checksum != hex.EncodeToString(want[:]) {
		t.Fatalf("checksum = %q, want canonical SHA-256", info.Checksum)
	}
	if got := counter.getCount(); got != 0 {
		t.Fatalf("normal Stat issued %d GetObject requests, want 0", got)
	}
}
