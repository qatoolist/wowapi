package secrets

import (
	"strings"
	"testing"
)

func TestParseRef(t *testing.T) {
	tests := []struct {
		in       string
		provider string
		path     string
		wantErr  bool
	}{
		{"secretref://env/DB_DSN", "env", "DB_DSN", false},
		{"secretref://aws/prod/db/dsn", "aws", "prod/db/dsn", false},
		{"secretref://k8s/ns/name#key", "k8s", "ns/name#key", false},
		{"postgres://user:pass@host/db", "", "", true}, // raw value, not a ref
		{"secretref://", "", "", true},
		{"secretref://env", "", "", true},
		{"secretref://env/", "", "", true},
		{"", "", "", true},
	}
	for _, tc := range tests {
		ref, err := ParseRef(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseRef(%q): want error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseRef(%q): %v", tc.in, err)
			continue
		}
		if ref.Provider != tc.provider || ref.Path != tc.path {
			t.Errorf("ParseRef(%q) = %+v", tc.in, ref)
		}
		if ref.String() != tc.in {
			t.Errorf("round-trip: %q != %q", ref.String(), tc.in)
		}
	}
}

// TestParseRefErrorNeverEchoesValue: a raw secret passed where a ref was
// expected must not be echoed into the error (which lands in logs).
func TestParseRefErrorNeverEchoesValue(t *testing.T) {
	raw := "postgres://user:sup3rsecret@host/db"
	_, err := ParseRef(raw)
	if err == nil {
		t.Fatal("want error")
	}
	if strings.Contains(err.Error(), "sup3rsecret") {
		t.Errorf("error echoed raw candidate: %v", err)
	}
}
