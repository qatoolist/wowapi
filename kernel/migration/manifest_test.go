package migration

import (
	"strings"
	"testing"
)

func manifestBody(extra string) string {
	b := `-- +wowapi:manifest
-- classification: online
-- rows_estimate: 1000
-- bytes_estimate: 65536
-- lock_timeout_ms: 2000
-- statement_timeout_ms: 5000
-- nn1_compatible: true
-- backfill_owner: none
-- validation_query: none
-- rollback_plan: goose Down reverses this additive-only migration
`
	return b + extra + `-- +wowapi:end
`
}

func manifestBodyOutside(extra string) string {
	return manifestBody("") + extra
}

func TestParseManifestComplete(t *testing.T) {
	m, err := ParseManifest("00031_test.sql", manifestBody(""))
	if err != nil {
		t.Fatalf("parse complete manifest: %v", err)
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("validate complete manifest: %v", err)
	}
	if m.Classification != Online {
		t.Fatalf("classification: got %q, want online", m.Classification)
	}
	if m.LockTimeoutMs != 2000 {
		t.Fatalf("lock_timeout_ms: got %d, want 2000", m.LockTimeoutMs)
	}
	if !m.NN1Compatible {
		t.Fatal("nn1_compatible: want true")
	}
}

func TestValidateMissingFields(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{
			name: "missing classification",
			body: strings.ReplaceAll(manifestBody(""), "-- classification: online\n", ""),
			want: "classification",
		},
		{
			name: "missing lock_timeout_ms",
			body: strings.ReplaceAll(manifestBody(""), "-- lock_timeout_ms: 2000\n", ""),
			want: "lock_timeout_ms",
		},
		{
			name: "missing backfill_owner",
			body: strings.ReplaceAll(manifestBody(""), "-- backfill_owner: none\n", ""),
			want: "backfill_owner",
		},
		{
			name: "missing validation_query",
			body: strings.ReplaceAll(manifestBody(""), "-- validation_query: none\n", ""),
			want: "validation_query",
		},
		{
			name: "missing rollback_plan",
			body: strings.ReplaceAll(manifestBody(""), "-- rollback_plan: goose Down reverses this additive-only migration\n", ""),
			want: "rollback_plan",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m, err := ParseManifest("00031_test.sql", tc.body)
			if err != nil {
				t.Fatalf("unexpected parse error: %v", err)
			}
			err = m.Validate()
			if err == nil {
				t.Fatalf("expected validation error containing %q, got nil", tc.want)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.want)
			}
		})
	}
}

func TestValidateOnlineLockBudget(t *testing.T) {
	body := strings.ReplaceAll(manifestBody(""), "-- lock_timeout_ms: 2000\n", "-- lock_timeout_ms: 3000\n")
	m, err := ParseManifest("00031_test.sql", body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	err = m.Validate()
	if err == nil || !strings.Contains(err.Error(), "<= 2000") {
		t.Fatalf("expected online lock budget error, got %v", err)
	}
}

func TestValidateStatementTimeoutOrdering(t *testing.T) {
	body := strings.ReplaceAll(manifestBody(""), "-- statement_timeout_ms: 5000\n", "-- statement_timeout_ms: 1000\n")
	m, err := ParseManifest("00031_test.sql", body)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	err = m.Validate()
	if err == nil || !strings.Contains(err.Error(), ">= lock_timeout_ms") {
		t.Fatalf("expected statement_timeout ordering error, got %v", err)
	}
}

func TestParseManifestUnknownKey(t *testing.T) {
	body := manifestBody("-- typo_field: 1\n")
	_, err := ParseManifest("00031_test.sql", body)
	if err == nil || !strings.Contains(err.Error(), "unknown manifest key") {
		t.Fatalf("expected unknown key error, got %v", err)
	}
}

func TestParseManifestDuplicateKey(t *testing.T) {
	body := manifestBody("-- classification: online\n")
	_, err := ParseManifest("00031_test.sql", body)
	if err == nil || !strings.Contains(err.Error(), "duplicate manifest key") {
		t.Fatalf("expected duplicate key error, got %v", err)
	}
}

func TestParseManifestIgnoresLinesOutsideBlock(t *testing.T) {
	body := manifestBodyOutside("-- typo_field: 1\n")
	m, err := ParseManifest("00031_test.sql", body)
	if err != nil {
		t.Fatalf("unexpected error for line outside block: %v", err)
	}
	if err := m.Validate(); err != nil {
		t.Fatalf("valid manifest rejected: %v", err)
	}
}

func TestMigrationVersion(t *testing.T) {
	v, err := MigrationVersion("00031_example.sql")
	if err != nil || v != 31 {
		t.Fatalf("MigrationVersion = %d, %v; want 31", v, err)
	}
	if _, err := MigrationVersion("example.sql"); err == nil {
		t.Fatal("expected error for non-migration filename")
	}
}
