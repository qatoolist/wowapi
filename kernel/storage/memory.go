package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
)

// Memory is an in-process Adapter for tests and local dev. Presigned URLs are
// synthetic ("mem://<key>?..."); the test/client uploads by calling Put with the
// bytes (standing in for the client's PUT to the presigned URL). Concurrency-safe.
type Memory struct {
	mu   sync.RWMutex
	objs map[string][]byte
	now  func() time.Time
}

// NewMemory returns an empty in-memory object store.
func NewMemory() *Memory {
	return &Memory{objs: map[string][]byte{}, now: time.Now}
}

// Put simulates the client uploading bytes to a presigned PUT URL. Real adapters
// have no such method — the client PUTs directly to the object store.
func (m *Memory) Put(key string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := make([]byte, len(data))
	copy(cp, data)
	m.objs[key] = cp
}

// Has reports whether a key currently holds an object.
func (m *Memory) Has(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.objs[key]
	return ok
}

func (m *Memory) PresignPut(_ context.Context, key string, ttl time.Duration) (PresignedURL, error) {
	return PresignedURL{URL: "mem://" + key + "?m=put", Method: "PUT", ExpiresAt: m.now().Add(ttl)}, nil
}

func (m *Memory) PresignPutChecksum(_ context.Context, key, checksumSHA256 string, ttl time.Duration) (PresignedURL, error) {
	if raw, err := hex.DecodeString(checksumSHA256); err != nil || len(raw) != sha256.Size {
		return PresignedURL{}, kerr.E(kerr.KindValidation, "invalid_upload_checksum", "upload checksum must be lowercase-hex SHA-256")
	}
	return PresignedURL{URL: "mem://" + key + "?m=put", Method: "PUT", ExpiresAt: m.now().Add(ttl)}, nil
}

func (m *Memory) PresignGet(_ context.Context, key string, ttl time.Duration) (PresignedURL, error) {
	m.mu.RLock()
	_, ok := m.objs[key]
	m.mu.RUnlock()
	if !ok {
		return PresignedURL{}, objectNotFound(key)
	}
	return PresignedURL{URL: "mem://" + key + "?m=get", Method: "GET", ExpiresAt: m.now().Add(ttl)}, nil
}

func (m *Memory) Stat(_ context.Context, key string) (ObjectInfo, error) {
	m.mu.RLock()
	data, ok := m.objs[key]
	m.mu.RUnlock()
	if !ok {
		return ObjectInfo{}, objectNotFound(key)
	}
	sum := sha256.Sum256(data)
	return ObjectInfo{Size: int64(len(data)), Checksum: hex.EncodeToString(sum[:])}, nil
}

func (m *Memory) Peek(_ context.Context, key string, n int) ([]byte, error) {
	m.mu.RLock()
	data, ok := m.objs[key]
	m.mu.RUnlock()
	if !ok {
		return nil, objectNotFound(key)
	}
	if n > len(data) {
		n = len(data)
	}
	out := make([]byte, n)
	copy(out, data[:n])
	return out, nil
}

func (m *Memory) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.objs, key)
	return nil
}
