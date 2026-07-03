package pagination_test

// Hot-path benchmarks for cursor encode/decode (criterion #17).
//
// DecodeCursor runs on every non-first page request (attacker-reachable input
// from the client). EncodeCursor runs once per response when minting the next
// cursor. Both should be fast and allocation-bounded.

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/pagination"
)

var benchCursorValues = map[string]any{
	"created_at": time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
	"id":         uuid.MustParse("01234567-89ab-cdef-0123-456789abcdef"),
}

// BenchmarkCursorEncode measures EncodeCursor: JSON marshal + base64url
// encoding of a two-column keyset tuple.
func BenchmarkCursorEncode(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pagination.EncodeCursor(benchCursorValues)
	}
}

// BenchmarkCursorDecode measures DecodeCursor: the attacker-reachable parse
// path. Includes base64url decode + JSON decode with UseNumber + number
// conversion. Must not allocate unboundedly on valid input.
func BenchmarkCursorDecode(b *testing.B) {
	encoded, err := pagination.EncodeCursor(benchCursorValues)
	if err != nil {
		b.Fatalf("setup: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pagination.DecodeCursor(encoded)
	}
}

// BenchmarkCursorDecodeEmpty measures the zero-cursor fast path (first page
// of every list request, empty cursor string).
func BenchmarkCursorDecodeEmpty(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = pagination.DecodeCursor("")
	}
}

// BenchmarkCursorRoundTrip measures encode+decode together: the full path for
// a list endpoint that receives a cursor from the client.
func BenchmarkCursorRoundTrip(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		enc, err := pagination.EncodeCursor(benchCursorValues)
		if err != nil {
			b.Fatal(err)
		}
		_, _ = pagination.DecodeCursor(enc)
	}
}
