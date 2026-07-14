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
	"00013_dlq_admin.sql",
	"00014_schedules.sql",
	"00015_sequences.sql",
	"00016_bulk_operations.sql",
	"00017_audit_logs.sql",
	"00018_audit_chain.sql",
	"00019_api_keys.sql",
	"00020_retention_dsr.sql",
	"00021_artifacts.sql",
	"00022_notification_channel_prefs.sql",
	"00023_audit_tx_id.sql",
	"00024_outbox_trace_context.sql",
	"00025_jobs_trace_context.sql",
	"00026_notify_trace_context.sql",
	"00027_audit_anchors.sql",
	"00028_jobs_rls.sql",
	"00029_permissions_step_up.sql",
	"00030_privileged_services.sql",
	"00031_seed_sync_runs.sql",
	"00032_version_counters_and_upload_sessions.sql",
	"00033_document_upload_session_document_id_index.sql",
	"00034_tenant_fk_parent_indexes.sql",
	"00035_tenant_fk_composite_not_valid.sql",
	"00036_tenant_fk_validate_and_cleanup.sql",
	"00037_audit_hash_version.sql",
	"00038_jobs_lease_columns.sql",
	"00039_identity_grant.sql",
	"00040_notify_webhook_lease_columns.sql",
	"00041_bulk_operation_processor_lock.sql",
	"00042_backfill_checkpoint_lease_columns.sql",
	"00043_webhook_failed_signature_audit.sql",
	"00044_bulk_items_lease_and_lifecycle.sql",
	"00045_goose_version_platform_select.sql",
	"00046_authz_epoch.sql",
	"00047_perf04_sweeper_outbox_leases.sql",
	"00048_rule_versions_resolution_indexes.sql",
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

// TestIdentityGrantMigrationHasRLSAndPartialUniqueIndex verifies that the
// identity_grant migration creates a FORCE-RLS tenant-scoped table with the
// expected partial unique index and app_platform-only write grants.
func TestIdentityGrantMigrationHasRLSAndPartialUniqueIndex(t *testing.T) {
	t.Parallel()

	fsys := migrations.Kernel()
	data := mustReadFile(t, fsys, "00039_identity_grant.sql")

	if !strings.Contains(data, "CREATE TABLE identity_grant") {
		t.Error("00039_identity_grant.sql: missing CREATE TABLE identity_grant")
	}
	if !strings.Contains(data, "FORCE ROW LEVEL SECURITY") {
		t.Error("00039_identity_grant.sql: missing FORCE ROW LEVEL SECURITY")
	}
	if !strings.Contains(data, "CREATE UNIQUE INDEX identity_grant_one_active_per_actor") {
		t.Error("00039_identity_grant.sql: missing one-active-grant-per-actor partial unique index")
	}
	if !strings.Contains(data, "WHERE status = 'active'") {
		t.Error("00039_identity_grant.sql: partial unique index predicate must be status = 'active'")
	}
	if !strings.Contains(data, "GRANT SELECT, INSERT, UPDATE ON identity_grant TO app_platform") {
		t.Error("00039_identity_grant.sql: missing app_platform grants")
	}

	// No DELETE grant: lifecycle-only.
	if strings.Contains(data, "DELETE ON identity_grant") {
		t.Error("00039_identity_grant.sql: must not grant DELETE on identity_grant")
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
