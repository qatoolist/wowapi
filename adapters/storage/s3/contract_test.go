// Contract test: the same behavioral assertions run against BOTH
// storage.NewMemory() and the s3 adapter, so the two Adapter implementations
// are provably interchangeable from a caller's point of view (the promise the
// storage.Adapter port makes). The s3 leg skips/fails per requireMinio's gate
// exactly like the rest of this package's integration tests.
//
// Every case gets a fresh, unique key (uuid-suffixed): Memory starts empty
// each subtest, but the s3 leg runs against a persistent MinIO bucket shared
// across test runs, so a fixed literal key would collide with objects left
// over from a previous run.
package s3_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/storage"
)

// storagePutter is satisfied by storage.Memory (a synthetic Put standing in
// for the client's real PUT to a presigned URL). The s3 leg instead performs
// a genuine HTTP PUT to the presigned URL returned by PresignPut.
type storagePutter interface {
	Put(key string, data []byte)
}

// contractCase runs one behavioral assertion against an arbitrary
// storage.Adapter. key is unique to this invocation; put uploads body to key
// (which knows how to actually get bytes into that specific backend).
type contractCase struct {
	name string
	run  func(t *testing.T, ctx context.Context, a storage.Adapter, key string, put func(key string, body []byte))
}

var contractCases = []contractCase{
	{
		name: "Stat reports size and sha256 of the stored bytes",
		run: func(t *testing.T, ctx context.Context, a storage.Adapter, key string, put func(key string, body []byte)) {
			body := []byte("contract test payload")
			put(key, body)

			info, err := a.Stat(ctx, key)
			if err != nil {
				t.Fatalf("Stat: %v", err)
			}
			sum := sha256.Sum256(body)
			if info.Size != int64(len(body)) {
				t.Errorf("Size = %d, want %d", info.Size, len(body))
			}
			if want := hex.EncodeToString(sum[:]); info.Checksum != want {
				t.Errorf("Checksum = %s, want %s", info.Checksum, want)
			}
		},
	},
	{
		name: "Peek returns a leading prefix and the whole object when n exceeds size",
		run: func(t *testing.T, ctx context.Context, a storage.Adapter, key string, put func(key string, body []byte)) {
			body := []byte("0123456789abcdef")
			put(key, body)

			prefix, err := a.Peek(ctx, key, 4)
			if err != nil {
				t.Fatalf("Peek: %v", err)
			}
			if string(prefix) != "0123" {
				t.Errorf("Peek(4) = %q, want %q", prefix, "0123")
			}
			whole, err := a.Peek(ctx, key, len(body)+100)
			if err != nil {
				t.Fatalf("Peek(oversized): %v", err)
			}
			if string(whole) != string(body) {
				t.Errorf("Peek(oversized) = %q, want %q", whole, body)
			}
		},
	},
	{
		name: "missing key is KindNotFound on Stat, Peek, and PresignGet",
		run: func(t *testing.T, ctx context.Context, a storage.Adapter, key string, _ func(key string, body []byte)) {
			if _, err := a.Stat(ctx, key); kerr.KindOf(err) != kerr.KindNotFound {
				t.Errorf("Stat(missing) = %v, want KindNotFound", err)
			}
			if _, err := a.Peek(ctx, key, 10); kerr.KindOf(err) != kerr.KindNotFound {
				t.Errorf("Peek(missing) = %v, want KindNotFound", err)
			}
			if _, err := a.PresignGet(ctx, key, time.Minute); kerr.KindOf(err) != kerr.KindNotFound {
				t.Errorf("PresignGet(missing) = %v, want KindNotFound", err)
			}
		},
	},
	{
		name: "Delete is idempotent",
		run: func(t *testing.T, ctx context.Context, a storage.Adapter, key string, put func(key string, body []byte)) {
			put(key, []byte("gone soon"))

			if err := a.Delete(ctx, key); err != nil {
				t.Fatalf("first Delete: %v", err)
			}
			if _, err := a.Stat(ctx, key); kerr.KindOf(err) != kerr.KindNotFound {
				t.Fatalf("Stat after Delete = %v, want KindNotFound", err)
			}
			if err := a.Delete(ctx, key); err != nil {
				t.Fatalf("second Delete (of an absent key) must be a no-op, got %v", err)
			}
		},
	},
	{
		name: "PresignGet succeeds only once the key exists",
		run: func(t *testing.T, ctx context.Context, a storage.Adapter, key string, put func(key string, body []byte)) {
			if _, err := a.PresignGet(ctx, key, time.Minute); kerr.KindOf(err) != kerr.KindNotFound {
				t.Fatalf("PresignGet before upload = %v, want KindNotFound", err)
			}
			put(key, []byte("now it exists"))
			if _, err := a.PresignGet(ctx, key, time.Minute); err != nil {
				t.Fatalf("PresignGet after upload: %v", err)
			}
		},
	},
}

func TestContract_Memory(t *testing.T) {
	for _, tc := range contractCases {
		t.Run(tc.name, func(t *testing.T) {
			m := storage.NewMemory()
			key := "contract/" + uuid.NewString()
			tc.run(t, context.Background(), m, key, m.Put)
		})
	}
}

func TestContract_S3(t *testing.T) {
	a := requireMinio(t, testConfig())
	ctx := context.Background()
	for _, tc := range contractCases {
		t.Run(tc.name, func(t *testing.T) {
			key := "contract/" + uuid.NewString()
			put := func(key string, body []byte) {
				sum := sha256.Sum256(body)
				presigned, err := a.PresignPutChecksum(ctx, key, hex.EncodeToString(sum[:]), time.Minute)
				if err != nil {
					t.Fatalf("PresignPutChecksum: %v", err)
				}
				httpPut(t, presigned, body)
			}
			tc.run(t, ctx, a, key, put)
		})
	}
}

var _ storagePutter = (*storage.Memory)(nil)
