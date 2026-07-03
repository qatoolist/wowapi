package fakes

import (
	"encoding/binary"
	"sync"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/model"
)

// IDGen is a deterministic model.IDGen: every New() encodes a seed in byte 0 and
// a big-endian counter in the last 8 bytes, with valid RFC 4122 v7 version and
// variant bits set. The result is stable across runs and readable in failure
// output — e.g. seed 0x01, first call → 01000000-0000-7000-8000-000000000001.
// Safe for concurrent use.
type IDGen struct {
	mu      sync.Mutex
	seed    byte
	counter uint64
}

// NewIDGen returns a deterministic generator whose UUIDs carry seed in byte 0.
func NewIDGen(seed byte) *IDGen { return &IDGen{seed: seed} }

var _ model.IDGen = (*IDGen)(nil)

// New returns the next UUID in the deterministic sequence.
func (g *IDGen) New() uuid.UUID {
	g.mu.Lock()
	g.counter++
	n := g.counter
	g.mu.Unlock()

	var u uuid.UUID
	u[0] = g.seed
	binary.BigEndian.PutUint64(u[8:], n)
	// RFC 4122: version nibble in byte 6, variant bits in the top 2 of byte 8.
	// For readable counters (< 2^56) byte 8 is 0x00 → 0x80 after masking, so the
	// counter stays legible in the trailing hex.
	u[6] = (u[6] & 0x0f) | 0x70
	u[8] = (u[8] & 0x3f) | 0x80
	return u
}
