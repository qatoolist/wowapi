package jobs

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// TestTruncateRuneSafe proves truncate never splits a multibyte rune: a '世'
// (3 bytes) straddling the cut is dropped rather than producing invalid UTF-8,
// which a Postgres text column would reject.
func TestTruncateRuneSafe(t *testing.T) {
	// Short strings pass through unchanged.
	if got := truncate("short", 500); got != "short" {
		t.Fatalf("short string changed: %q", got)
	}
	// Pure ASCII cuts exactly at n.
	if got := truncate(strings.Repeat("x", 600), 500); got != strings.Repeat("x", 500) {
		t.Fatalf("ascii truncation length = %d, want 500", len(got))
	}
	// A multibyte rune straddling the cut backs up to the rune boundary.
	in := strings.Repeat("a", 499) + "世" // '世' is 3 bytes at offsets 499..501
	got := truncate(in, 500)
	if !utf8.ValidString(got) {
		t.Fatalf("truncate produced invalid UTF-8: %q", got)
	}
	if got != strings.Repeat("a", 499) {
		t.Fatalf("truncate should drop the straddling rune, got len %d", len(got))
	}
}
