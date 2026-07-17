package migration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/qatoolist/wowapi/v2/testkit"
)

// TestValidationArtifactSchema proves that a zero-mismatch validation report
// conforms to the defined artifact schema and that the schema validator rejects
// malformed inputs.
func TestValidationArtifactSchema(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test needs Postgres")
	}
	db := testkit.NewDB(t)
	ctx := context.Background()

	admin, err := db.Admin.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	defer admin.Release()

	if _, err := admin.Exec(ctx, "CREATE TABLE IF NOT EXISTS validate_test (id int primary key, legacy int, modern int)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	defer func() { _, _ = admin.Exec(ctx, "DROP TABLE IF EXISTS validate_test CASCADE") }()

	for i := range 10 {
		if _, err := admin.Exec(ctx, "INSERT INTO validate_test (id, legacy, modern) VALUES ($1, $2, $2)", i, i*10); err != nil {
			t.Fatalf("insert: %v", err)
		}
	}

	report := NewValidationReport("00031_test")
	check, err := Reconcile(ctx, admin.Conn(), "legacy_modern_parity",
		"SELECT count(*) FROM validate_test WHERE legacy <> modern")
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	report.AddCheck(check)

	if !report.Passed {
		t.Fatalf("expected report to pass, got checks %+v", report.Checks)
	}
	if check.Mismatch != 0 {
		t.Fatalf("expected zero mismatches, got %d", check.Mismatch)
	}

	data, err := report.ToJSON()
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := ValidateArtifactSchema(data); err != nil {
		t.Fatalf("artifact schema validation failed: %v", err)
	}

	// Non-zero mismatch must fail the report and still validate as schema.
	if _, err := admin.Exec(ctx, "INSERT INTO validate_test (id, legacy, modern) VALUES ($1, $2, $3)", 99, 1, 2); err != nil {
		t.Fatalf("insert mismatch: %v", err)
	}
	check2, err := Reconcile(ctx, admin.Conn(), "legacy_modern_parity_after_insert",
		"SELECT count(*) FROM validate_test WHERE legacy <> modern")
	if err != nil {
		t.Fatalf("reconcile after insert: %v", err)
	}
	if check2.Passed {
		t.Fatal("expected mismatch check to fail")
	}
	report2 := NewValidationReport("00031_test")
	report2.AddCheck(check2)
	if report2.Passed {
		t.Fatal("expected report2 to fail")
	}
	data2, err := report2.ToJSON()
	if err != nil {
		t.Fatalf("marshal report2: %v", err)
	}
	if err := ValidateArtifactSchema(data2); err != nil {
		t.Fatalf("artifact schema validation failed for mismatch report: %v", err)
	}

	// Malformed JSON must be rejected.
	if err := ValidateArtifactSchema([]byte(`{"not":"a report"}`)); err == nil {
		t.Fatal("expected schema validation to reject malformed report")
	}

	// Raw typed report must round-trip through JSON.
	var round ValidationReport
	if err := json.Unmarshal(data, &round); err != nil {
		t.Fatalf("unmarshal report: %v", err)
	}
	if round.SchemaVersion != validationReportSchemaVersion {
		t.Fatalf("schema version = %q, want %q", round.SchemaVersion, validationReportSchemaVersion)
	}
}
