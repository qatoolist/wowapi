package pagination_test

import (
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/pagination"
)

// DecodeCursor parses attacker-reachable input (the client round-trips the cursor
// verbatim). This fuzz test asserts it never panics and only ever fails with a
// KindValidation error — never a bare/internal error, never an allocation blowup
// (roadmap S8). The seed corpus runs in normal CI; `go test -fuzz` explores.
func FuzzDecodeCursor(f *testing.F) {
	valid, _ := pagination.EncodeCursorWithSig("created_at:asc,id:asc", map[string]any{"id": int64(1), "created_at": "2026-07-04T00:00:00Z"})
	signed, _ := pagination.EncodeCursorWithSig("id:asc", map[string]any{"id": int64(1)})
	for _, s := range []string{
		valid, signed, "", "!!!", "eyJ", "e30", "____",
		strings.Repeat("A", 5000), "{}", "[]", "bnVsbA",
	} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		cur, err := pagination.DecodeCursor(s)
		if err != nil {
			if errors.KindOf(err) != errors.KindValidation {
				t.Fatalf("decode of %q returned non-validation error: %v", s, err)
			}
			return
		}
		// A successful decode must yield a self-consistent, panic-free cursor.
		_ = cur.Values()
		_ = cur.Sig()
		_ = cur.IsZero()
	})
}
