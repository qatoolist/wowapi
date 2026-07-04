package migrations_test

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/migrations"
)

// expectedFiles lists the kernel SQL migrations in ascending goose order.
// Phase 2 shipped 00001–00002 (D-0025); Phase 3 pulls idempotency_keys forward
// into 00003 (D-0031).
var expectedFiles = []string{
	"00001_bootstrap.sql",
	"00002_core_identity.sql",
	"00003_idempotency.sql",
	"00004_org_party_capacity.sql",
	"00005_resource_relationship.sql",
	"00006_authz.sql",
	"00007_outbox_jobs.sql",
	"00008_rules.sql",
	"00009_workflow.sql",
	"00010_documents.sql",
	"00011_notify_webhook_integration.sql",
	"00012_idempotency_sweep.sql",
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
		name := name
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

// TestBootstrapContainsAppTenantIDNoForce checks that the bootstrap migration
// defines app_tenant_id() (required by RLS policies in later migrations) and
// does NOT contain FORCE ROW LEVEL SECURITY (no tenant tables exist in 00001).
func TestBootstrapContainsAppTenantIDNoForce(t *testing.T) {
	t.Parallel()

	fsys := migrations.Kernel()
	data := mustReadFile(t, fsys, "00001_bootstrap.sql")

	if !strings.Contains(data, "app_tenant_id") {
		t.Error("00001_bootstrap.sql: missing app_tenant_id function definition")
	}
	if strings.Contains(data, "FORCE") {
		t.Error("00001_bootstrap.sql: unexpected FORCE keyword — no tenant tables exist in bootstrap")
	}
}

// TestCoreIdentityNoRLSNoPassword verifies that the core-identity migration
// carries no ROW LEVEL SECURITY directives (these are global tables per
// docs/blueprint/03 §2 and D-0025) and no plaintext PASSWORD literals (roles
// are NOLOGIN; passwords must never appear in migration files).
func TestCoreIdentityNoRLSNoPassword(t *testing.T) {
	t.Parallel()

	fsys := migrations.Kernel()
	data := mustReadFile(t, fsys, "00002_core_identity.sql")

	if strings.Contains(data, "ROW LEVEL SECURITY") {
		t.Error("00002_core_identity.sql: global tables must not have ROW LEVEL SECURITY")
	}

	// Case-insensitive scan for PASSWORD to catch any plaintext credential slip.
	passwordRe := regexp.MustCompile(`(?i)\bPASSWORD\b`)
	if passwordRe.MatchString(data) {
		t.Error("00002_core_identity.sql: contains PASSWORD keyword — no credentials in migrations")
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
