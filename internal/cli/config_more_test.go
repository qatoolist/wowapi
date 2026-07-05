package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------- resolve: explicit --base ----------

func TestConfigValidateExplicitBase(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "custom-base.yaml")
	if err := os.WriteFile(base, []byte("environment: dev\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	code, out, errOut := run(t, "config", "validate", "--base", base)
	if code != 0 {
		t.Fatalf("explicit --base exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, "config OK") {
		t.Fatalf("expected config OK, got %q", out)
	}
}

// ---------- assertEnv mismatch on print & doctor ----------

func TestConfigPrintEnvMismatch(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")
	// Overlay named prod but declaring dev — assertEnv must reject.
	writeYAML(t, dir, "prod.yaml", "environment: dev\n")
	code, _, errOut := run(t, "config", "print", "--redacted", "--dir", dir, "--env", "prod")
	if code != 1 {
		t.Fatalf("env mismatch should exit 1, got %d", code)
	}
	if !strings.Contains(errOut, "declares environment") {
		t.Fatalf("expected assertEnv message, got %q", errOut)
	}
}

func TestConfigDoctorEnvMismatch(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")
	writeYAML(t, dir, "prod.yaml", "environment: dev\n")
	code, _, errOut := run(t, "config", "doctor", "--dir", dir, "--env", "prod")
	if code != 1 {
		t.Fatalf("env mismatch should exit 1, got %d", code)
	}
	if !strings.Contains(errOut, "declares environment") {
		t.Fatalf("expected assertEnv message, got %q", errOut)
	}
}

// A doctor overlay that changes a value exercises the overlay-provenance path.
func TestConfigDoctorWithOverlay(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\nlog:\n  level: info\n")
	writeYAML(t, dir, "dev.yaml", "log:\n  level: warn\n")
	code, out, errOut := run(t, "config", "doctor", "--dir", dir, "--env", "dev")
	if code != 0 {
		t.Fatalf("doctor with overlay exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, "fingerprint=") {
		t.Fatalf("expected fingerprint line, got %q", out)
	}
}

// ---------- config diff: identical + load error ----------

func TestConfigDiffIdentical(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")
	writeYAML(t, dir, "dev.yaml", "log:\n  level: info\n")
	code, out, errOut := run(t, "config", "diff", "--dir", dir, "--from", "dev", "--to", "dev")
	if code != 0 {
		t.Fatalf("identical diff exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, "identical") {
		t.Fatalf("expected 'identical', got %q", out)
	}
}

func TestConfigDiffLoadError(t *testing.T) {
	dir := t.TempDir()
	// Invalid log level in base; the "from" (dev) overlay does not override it, so
	// loading dev fails first.
	writeYAML(t, dir, "base.yaml", "environment: dev\nlog:\n  level: boguslevel\n")
	writeYAML(t, dir, "dev.yaml", "environment: dev\n")
	writeYAML(t, dir, "prod.yaml", "environment: prod\n")
	code, _, errOut := run(t, "config", "diff", "--dir", dir, "--from", "dev", "--to", "prod")
	if code != 1 {
		t.Fatalf("load error should exit 1, got %d", code)
	}
	if !strings.Contains(errOut, "load dev") {
		t.Fatalf("expected 'load dev' error, got %q", errOut)
	}
}

// ---------- product-local configcheck delegation ----------

// writeConfigcheckModule creates a minimal Go module in cwd whose
// tools/configcheck/main.go exits non-zero iff it is passed FAILME, letting us
// drive both delegation branches.
func writeConfigcheckModule(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/cc\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	ccDir := filepath.Join(dir, "tools", "configcheck")
	if err := os.MkdirAll(ccDir, 0o755); err != nil {
		t.Fatal(err)
	}
	main := `package main

import (
	"fmt"
	"os"
)

func main() {
	for _, a := range os.Args[1:] {
		if a == "FAILME" {
			fmt.Fprintln(os.Stderr, "configcheck: forced failure")
			os.Exit(3)
		}
	}
	fmt.Println("product configcheck ran")
}
`
	if err := os.WriteFile(filepath.Join(ccDir, "main.go"), []byte(main), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestConfigDelegatesToProductChecker(t *testing.T) {
	dir := writeConfigcheckModule(t)
	t.Chdir(dir)
	// go run may compile; give it real work but no external deps.
	code, out, errOut := run(t, "config", "validate")
	if code != 0 {
		t.Fatalf("delegated validate exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, "product configcheck ran") {
		t.Fatalf("expected delegated output, got stdout=%q stderr=%q", out, errOut)
	}
}

func TestConfigDelegationPropagatesExitCode(t *testing.T) {
	dir := writeConfigcheckModule(t)
	t.Chdir(dir)
	// `go run` collapses a child os.Exit(3) into its own exit status 1 while
	// forwarding the child's stderr — so we assert non-zero + the child's message,
	// proving the product checker ran and its failure propagated.
	code, _, errOut := run(t, "config", "validate", "FAILME")
	if code == 0 {
		t.Fatalf("delegated failure should be non-zero, got %d", code)
	}
	if !strings.Contains(errOut, "forced failure") {
		t.Fatalf("expected the product checker's stderr to be forwarded, got %q", errOut)
	}
}
