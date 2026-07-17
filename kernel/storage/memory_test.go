package storage_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"testing"
	"time"

	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/storage"
)

func TestMemoryStatAndPeek(t *testing.T) {
	m := storage.NewMemory()
	ctx := context.Background()
	data := []byte("hello, framework")
	m.Put("t/doc/1", data)

	info, err := m.Stat(ctx, "t/doc/1")
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(data)
	if info.Size != int64(len(data)) || info.Checksum != hex.EncodeToString(sum[:]) {
		t.Fatalf("stat mismatch: %+v", info)
	}
	prefix, err := m.Peek(ctx, "t/doc/1", 5)
	if err != nil {
		t.Fatal(err)
	}
	if string(prefix) != "hello" {
		t.Fatalf("peek = %q, want hello", prefix)
	}
}

func TestMemoryMissingObject(t *testing.T) {
	m := storage.NewMemory()
	ctx := context.Background()
	if _, err := m.Stat(ctx, "nope"); kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("stat of a missing key must be NotFound, got %v", err)
	}
	if _, err := m.PresignGet(ctx, "nope", time.Minute); kerr.KindOf(err) != kerr.KindNotFound {
		t.Fatalf("presign-get of a missing key must be NotFound, got %v", err)
	}
}

func TestMemoryPresignAndDelete(t *testing.T) {
	m := storage.NewMemory()
	ctx := context.Background()
	put, err := m.PresignPut(ctx, "k", time.Minute)
	if err != nil || put.Method != http.MethodPut || put.URL == "" {
		t.Fatalf("presign put: %+v %v", put, err)
	}
	m.Put("k", []byte("x"))
	if !m.Has("k") {
		t.Fatal("Has should be true after Put")
	}
	if err := m.Delete(ctx, "k"); err != nil {
		t.Fatal(err)
	}
	if m.Has("k") {
		t.Fatal("Has should be false after Delete")
	}
	// Delete is idempotent.
	if err := m.Delete(ctx, "k"); err != nil {
		t.Fatalf("delete of a missing key must be a no-op, got %v", err)
	}
}
