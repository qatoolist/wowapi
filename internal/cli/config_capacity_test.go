package cli

import (
	"strings"
	"testing"
)

// TestConfigCapacityHappy proves a within-budget shape exits 0 and prints a
// clear OK line.
func TestConfigCapacityHappy(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", `
environment: dev
db:
  max_conns: 100
concurrency:
  replicas: 2
  runtime_pool_max: 16
  platform_pool_max: 8
  migrate_pool_max: 2
  reserved_admin: 4
`)
	code, out, errOut := run(t, "config", "capacity", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	if !strings.Contains(out, "capacity OK") {
		t.Errorf("stdout missing 'capacity OK': %q", out)
	}
}

// TestConfigCapacityOversubscribedFails proves an oversubscribed shape exits
// non-zero and reports the computed demand vs db.max_conns — this is the
// `wowapi config capacity` lint the backlog calls for, independent of the
// deployment's own concurrency.capacity_mode (the CLI always treats an
// oversubscribed shape as a lint failure; capacity_mode only governs whether
// `wowapi config validate`/boot itself fails).
func TestConfigCapacityOversubscribedFails(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", `
environment: dev
db:
  max_conns: 20
concurrency:
  replicas: 3
  runtime_pool_max: 16
  platform_pool_max: 8
  migrate_pool_max: 2
  reserved_admin: 4
`)
	code, _, errOut := run(t, "config", "capacity", "--dir", dir)
	if code != 1 {
		t.Fatalf("exit %d, want 1; stderr: %s", code, errOut)
	}
	if !strings.Contains(errOut, "capacity budget exceeded") {
		t.Errorf("stderr missing capacity-budget message: %q", errOut)
	}
	if !strings.Contains(errOut, "78") || !strings.Contains(errOut, "20") {
		t.Errorf("stderr should cite computed demand (78) and db.max_conns (20): %q", errOut)
	}
}

// TestConfigCapacityUnconfiguredShapeIsOK proves that leaving Replicas at its
// zero value (deployment shape not declared) is reported as OK, not a
// failure — the check is a deliberate no-op until a product opts in.
func TestConfigCapacityUnconfiguredShapeIsOK(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")
	code, out, errOut := run(t, "config", "capacity", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	if !strings.Contains(out, "capacity OK") {
		t.Errorf("stdout missing 'capacity OK': %q", out)
	}
	if !strings.Contains(out, "not configured") {
		t.Errorf("stdout should note the shape is not configured: %q", out)
	}
}

// TestConfigCapacityLoadError proves a broken config surfaces the load error
// and exits 1, same as the other config subcommands.
func TestConfigCapacityLoadError(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "log:\n  level: loud\n") // missing environment + bad level
	code, _, errOut := run(t, "config", "capacity", "--dir", dir)
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	if !strings.Contains(errOut, "environment") {
		t.Errorf("stderr should mention 'environment': %q", errOut)
	}
}

// TestConfigCapacityBadFlag proves an unparseable flag exits 2, matching the
// other config subcommands' flag.ContinueOnError handling.
func TestConfigCapacityBadFlag(t *testing.T) {
	code, _, _ := run(t, "config", "capacity", "--not-a-real-flag")
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
}

// TestConfigCapacityMissingOverlay proves a requested --env overlay that
// does not exist on disk is a hard error, same as `config validate`.
func TestConfigCapacityMissingOverlay(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")
	code, _, errOut := run(t, "config", "capacity", "--dir", dir, "--env", "nonexistent")
	if code == 0 {
		t.Fatalf("expected non-zero exit, got 0")
	}
	if !strings.Contains(errOut, "nonexistent") {
		t.Errorf("stderr should mention the missing overlay name: %q", errOut)
	}
}

// TestConfigCapacityEnvMismatch proves --env <X> against an overlay that
// declares a different environment is rejected (SEC-6: a prod.yaml that
// (mis)declares another environment must fail the gate).
func TestConfigCapacityEnvMismatch(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")
	writeYAML(t, dir, "stage.yaml", "environment: dev\n") // declares dev, but requested as --env stage
	code, _, errOut := run(t, "config", "capacity", "--dir", dir, "--env", "stage")
	if code != 1 {
		t.Fatalf("exit %d, want 1; stderr: %s", code, errOut)
	}
	if !strings.Contains(errOut, "stage") {
		t.Errorf("stderr should mention the requested environment: %q", errOut)
	}
}

// TestConfigCapacityWarningsLoopDoesNotBreakSuccess exercises the same
// warnings-forwarding loop runConfigValidate/runConfigDoctor use (currently
// empty in this fixture, matching how those commands' happy-path tests also
// only cover the empty-Warnings case) and confirms a within-budget shape
// still exits 0 and prints "capacity OK".
func TestConfigCapacityWarningsLoopDoesNotBreakSuccess(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", `
environment: dev
db:
  max_conns: 100
concurrency:
  replicas: 1
  runtime_pool_max: 16
  platform_pool_max: 8
`)
	code, out, errOut := run(t, "config", "capacity", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	if !strings.Contains(out, "capacity OK") {
		t.Errorf("stdout missing 'capacity OK': %q", out)
	}
}
