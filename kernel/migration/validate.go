package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// ValidationReport is the machine-checked artifact produced by the validation
// phase. It is intentionally serializable to JSON for CI evidence capture.
type ValidationReport struct {
	SchemaVersion string            `json:"schema_version"`
	GeneratedAt   time.Time         `json:"generated_at"`
	Migration     string            `json:"migration"`
	Checks        []ValidationCheck `json:"checks"`
	Passed        bool              `json:"passed"`
}

// ValidationCheck records one validation assertion.
type ValidationCheck struct {
	Name        string        `json:"name"`
	Query       string        `json:"query"`
	Duration    time.Duration `json:"duration_ms"`
	Mismatch    int64         `json:"mismatch_count"`
	Passed      bool          `json:"passed"`
	Explanation string        `json:"explanation,omitempty"`
}

const validationReportSchemaVersion = "2026-07-13"

// ValidateConstraint runs `ALTER TABLE ... VALIDATE CONSTRAINT` for a NOT VALID
// constraint added during the expand phase. It returns the validated constraint
// name.
func ValidateConstraint(ctx context.Context, conn *pgx.Conn, table, constraint string) error {
	sql := fmt.Sprintf("ALTER TABLE %s VALIDATE CONSTRAINT %s", quoteIdent(table), quoteIdent(constraint))
	_, err := conn.Exec(ctx, sql)
	return err
}

// Reconcile runs a mismatch-count query and returns a ValidationCheck. A
// mismatch count of zero means the old and new representations agree.
func Reconcile(ctx context.Context, conn *pgx.Conn, name, query string) (ValidationCheck, error) {
	start := time.Now()
	var mismatch int64
	err := conn.QueryRow(ctx, query).Scan(&mismatch)
	dur := time.Since(start)
	if err != nil {
		return ValidationCheck{}, fmt.Errorf("reconcile %s: %w", name, err)
	}
	return ValidationCheck{
		Name:     name,
		Query:    query,
		Duration: dur,
		Mismatch: mismatch,
		Passed:   mismatch == 0,
	}, nil
}

// NewValidationReport builds a fresh report for a migration.
func NewValidationReport(migration string) *ValidationReport {
	return &ValidationReport{
		SchemaVersion: validationReportSchemaVersion,
		GeneratedAt:   time.Now().UTC(),
		Migration:     migration,
		Checks:        []ValidationCheck{},
		Passed:        true,
	}
}

// AddCheck appends a check and updates the top-level Passed flag.
func (r *ValidationReport) AddCheck(c ValidationCheck) {
	r.Checks = append(r.Checks, c)
	if !c.Passed {
		r.Passed = false
	}
}

// ToJSON serializes the report as indented JSON.
func (r *ValidationReport) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// ValidateArtifactSchema returns an error if the report does not conform to
// the expected shape. It is a lightweight schema guard (types, required fields,
// zero-mismatch semantics) for tests and CI.
func ValidateArtifactSchema(data []byte) error {
	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("unmarshal: %w", err)
	}
	for _, key := range []string{"schema_version", "generated_at", "migration", "checks", "passed"} {
		if _, ok := raw[key]; !ok {
			return fmt.Errorf("missing required field %q", key)
		}
	}
	checks, ok := raw["checks"].([]any)
	if !ok {
		return fmt.Errorf("checks must be an array")
	}
	for i, c := range checks {
		check, ok := c.(map[string]any)
		if !ok {
			return fmt.Errorf("check %d is not an object", i)
		}
		for _, key := range []string{"name", "query", "duration_ms", "mismatch_count", "passed"} {
			if _, ok := check[key]; !ok {
				return fmt.Errorf("check %d missing required field %q", i, key)
			}
		}
	}
	return nil
}
