// Constructor-level unit tests: config validation fails before any network
// I/O, and endpoint parsing accepts the shapes the devbox/compose emit.
// Networked behavior lives in s3_test.go (real minio) and document_e2e_test.go.
package s3

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// valid is a syntactically complete config; tests break one field at a time.
// (No minio needs to be listening: every case below must fail validation
// before the client touches the endpoint.)
func valid() Config {
	return Config{
		Endpoint:  "localhost:9000",
		Bucket:    "unit-test",
		AccessKey: "ak",
		SecretKey: "sk",
	}
}

func TestNew_ConstructorErrors(t *testing.T) {
	cases := map[string]struct {
		mutate func(*Config)
		want   string // substring the error must carry
	}{
		"missing endpoint":   {func(c *Config) { c.Endpoint = "" }, "endpoint required"},
		"missing bucket":     {func(c *Config) { c.Bucket = "" }, "bucket required"},
		"missing access key": {func(c *Config) { c.AccessKey = "" }, "access key required"},
		"missing secret key": {func(c *Config) { c.SecretKey = "" }, "secret key required"},
		"negative TTL":       {func(c *Config) { c.PresignTTL = -time.Minute }, "presign TTL"},
		"TTL above S3 cap":   {func(c *Config) { c.PresignTTL = 8 * 24 * time.Hour }, "presign TTL"},
		"bogus scheme":       {func(c *Config) { c.Endpoint = "ftp://host:9000" }, "unsupported scheme"},
	}
	for name, tc := range cases {
		cfg := valid()
		tc.mutate(&cfg)
		_, err := New(context.Background(), cfg)
		if err == nil {
			t.Errorf("%s: New() succeeded, want error", name)
			continue
		}
		if !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: error %q does not contain %q", name, err, tc.want)
		}
	}
}

// Every field wrong at once: the constructor must report ALL problems joined
// (the same boot-failure ergonomics Framework.Validate guarantees).
func TestNew_ReportsAllProblemsJoined(t *testing.T) {
	_, err := New(context.Background(), Config{PresignTTL: -1})
	if err == nil {
		t.Fatal("New(zero Config) succeeded")
	}
	for _, want := range []string{"endpoint required", "bucket required", "access key required", "secret key required", "presign TTL"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("joined error misses %q: %v", want, err)
		}
	}
}

func TestSplitEndpoint(t *testing.T) {
	cases := map[string]struct {
		endpoint string
		useSSL   bool
		wantHost string
		wantTLS  bool
		wantErr  bool
	}{
		"bare host honors use_ssl false":  {"localhost:9000", false, "localhost:9000", false, false},
		"bare host honors use_ssl true":   {"s3.example.com", true, "s3.example.com", true, false},
		"http scheme forces plaintext":    {"http://minio:9000", true, "minio:9000", false, false},
		"https scheme forces TLS":         {"https://s3.example.com", false, "s3.example.com", true, false},
		"schemeless colon-slash rejected": {"gopher://x", false, "", false, true},
		"scheme without host rejected":    {"http://", false, "", false, true},
		"unparseable URL rejected":        {"http://%zz", false, "", false, true},
	}
	for name, tc := range cases {
		host, tls, err := splitEndpoint(tc.endpoint, tc.useSSL)
		if tc.wantErr {
			if err == nil {
				t.Errorf("%s: splitEndpoint(%q) err = nil, want error", name, tc.endpoint)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: splitEndpoint(%q) err = %v", name, tc.endpoint, err)
			continue
		}
		if host != tc.wantHost || tls != tc.wantTLS {
			t.Errorf("%s: splitEndpoint(%q) = (%q, %v), want (%q, %v)",
				name, tc.endpoint, host, tls, tc.wantHost, tc.wantTLS)
		}
	}
}

// badBucketAdapter builds a live Adapter (bypassing New's validation) whose
// bucket name minio-go rejects client-side ("Bucket name contains invalid
// characters") on every operation. This exercises the kerr.Wrapf error path
// of each method against the REAL minio-go SDK's own validation logic — no
// mock, no fake transport — the one class of adapter error that a healthy,
// reachable MinIO instance can never otherwise produce for these tests.
func badBucketAdapter(t *testing.T) *Adapter {
	t.Helper()
	client, err := minio.New("localhost:9000", &minio.Options{
		Creds:        credentials.NewStaticV4("ak", "sk", ""),
		Secure:       false,
		BucketLookup: minio.BucketLookupPath,
	})
	if err != nil {
		t.Fatalf("minio.New: %v", err)
	}
	return &Adapter{client: client, bucket: "Invalid_Bucket_Name!", presignTTL: time.Minute}
}

func TestAdapterMethods_WrapNonNotFoundErrors(t *testing.T) {
	a := badBucketAdapter(t)
	ctx := context.Background()

	if _, err := a.Stat(ctx, "k"); err == nil || strings.Contains(err.Error(), "not found") {
		t.Errorf("Stat with an invalid bucket = %v, want a wrapped non-NotFound error", err)
	}
	if _, err := a.Peek(ctx, "k", 4); err == nil || strings.Contains(err.Error(), "not found") {
		t.Errorf("Peek with an invalid bucket = %v, want a wrapped non-NotFound error", err)
	}
	if err := a.Delete(ctx, "k"); err == nil {
		t.Error("Delete with an invalid bucket succeeded, want a wrapped error")
	}
	if _, err := a.PresignPut(ctx, "k", time.Minute); err == nil {
		t.Error("PresignPut with an invalid bucket succeeded, want a wrapped error")
	}
	if _, err := a.PresignGet(ctx, "k", time.Minute); err == nil {
		t.Error("PresignGet with an invalid bucket succeeded, want a wrapped error")
	}
}

// ensureBucket's initial BucketExists probe also surfaces the same
// client-side validation error class; New must propagate it as a boot
// failure rather than panicking or masking it.
func TestNew_EnsureBucketProbeErrorPropagates(t *testing.T) {
	cfg := valid()
	cfg.Bucket = "Invalid_Bucket_Name!"
	if _, err := New(context.Background(), cfg); err == nil {
		t.Fatal("New with an invalid bucket name succeeded, want the BucketExists probe error")
	}
}

// A host that minio.New itself rejects (post-splitEndpoint validation, e.g. an
// unparseable escape sequence) must surface as a wrapped client-construction
// error, not a panic — no network reachability is required since the SDK
// validates the endpoint URL before ever dialing.
func TestNew_ClientConstructionErrorPropagates(t *testing.T) {
	cfg := valid()
	cfg.Endpoint = "%zz"
	_, err := New(context.Background(), cfg)
	if err == nil {
		t.Fatal("New with an unparseable endpoint succeeded, want a client-construction error")
	}
	if !strings.Contains(err.Error(), "storage: client:") {
		t.Errorf("error %q does not carry the client-construction prefix", err)
	}
}
