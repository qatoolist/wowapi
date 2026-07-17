package migrations_test

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/migrations"
)

// expectedFiles lists the kernel SQL migrations in ascending goose order.
var expectedFiles = []string{
	"00001_baseline.sql",
}

// TestKernelListsExpectedFiles verifies that Kernel() exposes exactly the
// expected .sql files and in ascending order.
func TestKernelListsExpectedFiles(t *testing.T) {
	t.Parallel()

	fsys := migrations.Kernel()

	var got []string
	err := fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".sql") {
			got = append(got, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}

	if len(got) != len(expectedFiles) {
		t.Fatalf("got %d SQL files %v; want %d %v", len(got), got, len(expectedFiles), expectedFiles)
	}
	for i, name := range expectedFiles {
		if got[i] != name {
			t.Errorf("file[%d]: got %q; want %q", i, got[i], name)
		}
	}
}

// TestGooseMarkersPresent verifies every migration file contains both
// "-- +goose Up" and "-- +goose Down" markers required by the goose runner.
func TestGooseMarkersPresent(t *testing.T) {
	t.Parallel()

	fsys := migrations.Kernel()
	for _, name := range expectedFiles {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			data := mustReadFile(t, fsys, name)
			if !strings.Contains(data, "-- +goose Up") {
				t.Errorf("%s: missing '-- +goose Up' marker", name)
			}
			if !strings.Contains(data, "-- +goose Down") {
				t.Errorf("%s: missing '-- +goose Down' marker", name)
			}
		})
	}
}

// TestBaselineCarriesDirectFinalStateContract guards the clean baseline against
// accidentally reintroducing the abandoned upgrade choreography.
func TestBaselineCarriesDirectFinalStateContract(t *testing.T) {
	t.Parallel()

	fsys := migrations.Kernel()
	data := mustReadFile(t, fsys, "00001_baseline.sql")

	if !strings.Contains(data, "app_tenant_id") {
		t.Error("baseline: missing app_tenant_id function definition")
	}
	if !strings.Contains(data, "CREATE TABLE public.workflow_definitions") ||
		!strings.Contains(data, "definition_digest text NOT NULL") {
		t.Error("baseline: missing final workflow-definition identity")
	}
	if strings.Contains(data, "NOT VALID") || strings.Contains(data, "VALIDATE CONSTRAINT") ||
		strings.Contains(data, "nn1_compatible") {
		t.Error("baseline: contains abandoned online-upgrade choreography")
	}
	if strings.Contains(data, "DROP COLUMN") || strings.Contains(data, "_catalog_slot_") {
		t.Error("baseline: reconstructs abandoned physical column history")
	}
}

func TestBaselineContainsNoPassword(t *testing.T) {
	t.Parallel()

	fsys := migrations.Kernel()
	data := mustReadFile(t, fsys, "00001_baseline.sql")

	passwordRe := regexp.MustCompile(`(?i)\bPASSWORD\b`)
	if passwordRe.MatchString(data) {
		t.Error("baseline contains PASSWORD keyword — no credentials belong in migrations")
	}
}

// TestIdentityGrantMigrationHasRLSAndPartialUniqueIndex verifies that the
// identity_grant migration creates a FORCE-RLS tenant-scoped table with the
// expected partial unique index and app_platform-only write grants.
func TestIdentityGrantMigrationHasRLSAndPartialUniqueIndex(t *testing.T) {
	t.Parallel()

	fsys := migrations.Kernel()
	data := mustReadFile(t, fsys, "00001_baseline.sql")

	if !strings.Contains(data, "CREATE TABLE public.identity_grant") {
		t.Error("baseline: missing CREATE TABLE identity_grant")
	}
	if !strings.Contains(data, "ALTER TABLE ONLY public.identity_grant FORCE ROW LEVEL SECURITY") {
		t.Error("baseline: identity_grant missing FORCE ROW LEVEL SECURITY")
	}
	if !strings.Contains(data, "CREATE UNIQUE INDEX identity_grant_one_active_per_actor") {
		t.Error("baseline: missing one-active-grant-per-actor partial unique index")
	}
	if !strings.Contains(data, "WHERE (status = 'active'::text)") {
		t.Error("baseline: partial unique index predicate must be status = 'active'")
	}
	if !strings.Contains(data, "GRANT SELECT,INSERT,UPDATE ON TABLE public.identity_grant TO app_platform") {
		t.Error("baseline: missing app_platform grants")
	}

	// No DELETE grant: lifecycle-only.
	if strings.Contains(data, "DELETE ON identity_grant") {
		t.Error("baseline: must not grant DELETE on identity_grant")
	}
}

// mustReadFile reads a file from fsys and returns its content as a string,
// failing the test on any error.
func mustReadFile(t *testing.T, fsys fs.FS, name string) string {
	t.Helper()
	b, err := fs.ReadFile(fsys, name)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", name, err)
	}
	return string(b)
}
