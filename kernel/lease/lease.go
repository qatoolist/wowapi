// Package lease provides a shared lease/fencing primitive used by the job
// queue, notify/webhook delivery, and bulk processing. A lease is an opaque
// token paired with a monotonic generation and an expiry time. Fencing
// compares both token and generation: a reclaimed row gets a new token and a
// bumped generation, so a stale worker's finalize is rejected even if its
// token were somehow replayed.
//
// The primitive is intentionally persistence-agnostic. Callers store the
// lease columns in their own tables and use this package for generation,
// token creation, and comparison semantics.
package lease

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// Lease is a fencing token. Token is an unguessable opaque string.
// Generation is monotonically increasing within the logical row lifetime.
// ExpiresAt is the wall-clock time after which the lease is no longer valid.
type Lease struct {
	Token      string
	Generation int64
	ExpiresAt  time.Time
}

// IsExpired reports whether the lease has passed its expiry.
func (l Lease) IsExpired(now time.Time) bool {
	return !l.ExpiresAt.IsZero() && l.ExpiresAt.Before(now)
}

// Equals reports whether other has the same token and generation.
// It does not check expiry; use IsCurrent for that.
func (l Lease) Equals(other Lease) bool {
	return l.Token == other.Token && l.Generation == other.Generation
}

// IsCurrent reports whether other represents exactly this lease epoch and it
// has not expired.
func (l Lease) IsCurrent(other Lease, now time.Time) bool {
	return l.Equals(other) && !l.IsExpired(now)
}

// IsNewer reports whether other is a strictly newer generation of the same
// token. This is useful for reclaim checks where the generation must advance.
func (l Lease) IsNewer(other Lease) bool {
	return l.Token == other.Token && l.Generation > other.Generation
}

// New creates a fresh lease with the given TTL. Generation starts at 1.
func New(ttl time.Duration) Lease {
	return Lease{
		Token:      newToken(),
		Generation: 1,
		ExpiresAt:  time.Now().Add(ttl),
	}
}

// Renew returns a lease with the same token and generation but a refreshed
// expiry. It is used by heartbeats.
func (l Lease) Renew(ttl time.Duration) Lease {
	return Lease{
		Token:      l.Token,
		Generation: l.Generation,
		ExpiresAt:  time.Now().Add(ttl),
	}
}

// NextEpoch returns a lease for the same logical row but with a new token and
// a generation one greater than l. This is what ReclaimStalled uses to fence
// off a crashed worker.
func (l Lease) NextEpoch(ttl time.Duration) Lease {
	return Lease{
		Token:      newToken(),
		Generation: l.Generation + 1,
		ExpiresAt:  time.Now().Add(ttl),
	}
}

// BumpGeneration returns a lease with the same token but generation+1. Some
// consumers prefer to keep the token across a reclaim but advance generation.
func (l Lease) BumpGeneration() Lease {
	return Lease{
		Token:      l.Token,
		Generation: l.Generation + 1,
		ExpiresAt:  l.ExpiresAt,
	}
}

// Zero reports whether l is the zero lease (no token, generation 0).
func (l Lease) Zero() bool {
	return l.Token == "" && l.Generation == 0 && l.ExpiresAt.IsZero()
}

func newToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand only fails on systems without an entropy source. Fall
		// back to a time-based token so the primitive does not panic, while
		// still being opaque.
		return fmt.Sprintf("lease-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
