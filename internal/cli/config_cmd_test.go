package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeYAML writes content to <dir>/<name>, creating dir if needed.
func writeYAML(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// ---------- validate ----------

func TestConfigValidateHappy(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")

	code, out, errOut := run(t, "config", "validate", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	if !strings.Contains(out, "config OK") {
		t.Errorf("stdout missing 'config OK': %q", out)
	}
	if !strings.Contains(out, "fingerprint=") {
		t.Errorf("stdout missing fingerprint: %q", out)
	}
}

func TestConfigValidateMissingEnvironment(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "log:\n  level: info\n")

	code, _, errOut := run(t, "config", "validate", "--dir", dir)
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	if !strings.Contains(errOut, "environment") {
		t.Errorf("stderr should mention 'environment': %q", errOut)
	}
}

func TestConfigValidateMultipleErrors(t *testing.T) {
	dir := t.TempDir()
	// No environment AND invalid log level: both errors must be reported.
	writeYAML(t, dir, "base.yaml", "log:\n  level: loud\n")

	code, _, errOut := run(t, "config", "validate", "--dir", dir)
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	if !strings.Contains(errOut, "environment") {
		t.Errorf("stderr missing 'environment': %q", errOut)
	}
	if !strings.Contains(errOut, "loud") {
		t.Errorf("stderr missing 'loud' (invalid log level): %q", errOut)
	}
}

func TestConfigValidateNoBaseFileSkipsLayer(t *testing.T) {
	// When base.yaml is absent and --base is not set, the loader uses only
	// env vars. Supply environment via env var to satisfy the fail-closed check.
	dir := t.TempDir()
	t.Setenv("WOWAPI__ENVIRONMENT", "dev")

	code, out, errOut := run(t, "config", "validate", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	if !strings.Contains(out, "config OK") {
		t.Errorf("stdout missing 'config OK': %q", out)
	}
}

// ---------- print ----------

func TestConfigPrintRequiresRedacted(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")

	code, _, errOut := run(t, "config", "print", "--dir", dir)
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	if !strings.Contains(errOut, "--redacted") {
		t.Errorf("stderr should mention --redacted: %q", errOut)
	}
}

func TestConfigPrintWithRedacted(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")

	code, out, errOut := run(t, "config", "print", "--redacted", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", err, out)
	}
	if m["environment"] != "dev" {
		t.Errorf("environment = %v, want \"dev\"", m["environment"])
	}
}

func TestConfigPrintLoadError(t *testing.T) {
	dir := t.TempDir()
	// Missing environment → load fails.
	writeYAML(t, dir, "base.yaml", "log:\n  level: info\n")

	code, _, errOut := run(t, "config", "print", "--redacted", "--dir", dir)
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	if !strings.Contains(errOut, "environment") {
		t.Errorf("stderr missing 'environment': %q", errOut)
	}
}

// ---------- schema ----------

func TestConfigSchema(t *testing.T) {
	code, out, errOut := run(t, "config", "schema")
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(out), &m); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %q", err, out)
	}
	if !strings.Contains(out, `"environment"`) {
		t.Errorf("schema missing 'environment': %q", out)
	}
	if !strings.Contains(out, `"additionalProperties"`) {
		t.Errorf("schema missing 'additionalProperties': %q", out)
	}
}

// ---------- doctor ----------

func TestConfigDoctorHappy(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")

	code, out, errOut := run(t, "config", "doctor", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	// environment was set in base.yaml → provenance = base-file.
	if !strings.Contains(out, "environment") {
		t.Errorf("doctor table missing 'environment': %q", out)
	}
	if !strings.Contains(out, "base-file") {
		t.Errorf("doctor table missing 'base-file': %q", out)
	}
	// http.addr was not in base.yaml → provenance = default.
	if !strings.Contains(out, "http.addr") {
		t.Errorf("doctor table missing 'http.addr': %q", out)
	}
	if !strings.Contains(out, "default") {
		t.Errorf("doctor table missing 'default' layer: %q", out)
	}
	if !strings.Contains(out, "fingerprint=") {
		t.Errorf("doctor output missing fingerprint line: %q", out)
	}
}

func TestConfigDoctorSortedOutput(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")

	code, out, _ := run(t, "config", "doctor", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	// Header must appear before data rows.
	headerIdx := strings.Index(out, "KEY")
	envIdx := strings.Index(out, "environment")
	if headerIdx < 0 || envIdx < 0 || headerIdx >= envIdx {
		t.Errorf("KEY header should precede environment row: %q", out)
	}
}

// ---------- env overlay ----------

func TestConfigEnvOverlayHappy(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\nlog:\n  level: info\n")
	writeYAML(t, dir, "dev.yaml", "log:\n  level: warn\n")

	code, out, errOut := run(t, "config", "validate", "--dir", dir, "--env", "dev")
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	if !strings.Contains(out, "config OK") {
		t.Errorf("stdout missing 'config OK': %q", out)
	}
}

func TestConfigEnvOverlayMissing(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")

	code, _, errOut := run(t, "config", "validate", "--dir", dir, "--env", "nonexistent")
	if code == 0 {
		t.Fatalf("expected non-zero exit, got 0")
	}
	if !strings.Contains(errOut, "nonexistent") {
		t.Errorf("stderr should mention the missing overlay name: %q", errOut)
	}
}

// ---------- subcommand routing ----------

func TestConfigUnknownSubcommand(t *testing.T) {
	code, _, errOut := run(t, "config", "flibbertigibbet")
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	if !strings.Contains(errOut, "unknown subcommand") {
		t.Errorf("stderr missing 'unknown subcommand': %q", errOut)
	}
}

func TestConfigNoSubcommand(t *testing.T) {
	code, _, errOut := run(t, "config")
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	// Usage should appear on stderr.
	if !strings.Contains(errOut, "subcommand") {
		t.Errorf("stderr missing usage text: %q", errOut)
	}
}

// ---------- help integration ----------

func TestHelpShowsConfigAsAvailable(t *testing.T) {
	code, out, _ := run(t, "help")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	// config must appear in the Available section, not only in Planned.
	if !strings.Contains(out, "config") {
		t.Errorf("help missing 'config': %q", out)
	}
	// validate subcommand should be mentioned to show it is available.
	if !strings.Contains(out, "validate") {
		t.Errorf("help missing 'validate' subcommand mention: %q", out)
	}
}
