package cli

import (
	"bytes"
	"testing"
)

// TestDLQArgValidation covers the argument checks that fail BEFORE any DB
// connection, so they run without DATABASE_URL. Each must exit 2.
func TestDLQArgValidation(t *testing.T) {
	cases := [][]string{
		{"jobs"},                        // missing action
		{"bogus", "list"},               // unknown domain
		{"jobs", "frob"},                // unknown action
		{"jobs", "replay"},              // missing id
		{"jobs", "replay", "abc"},       // non-numeric job id
		{"events", "discard", "not-id"}, // non-uuid event id
		{"events", "inspect"},           // missing id
	}
	for _, args := range cases {
		var out, errb bytes.Buffer
		if code := runDLQ(args, &out, &errb); code != 2 {
			t.Errorf("runDLQ(%v) = %d, want 2 (arg validation)", args, code)
		}
	}
}
