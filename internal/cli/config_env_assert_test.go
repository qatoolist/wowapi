package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SEC-6: `config validate --env prod` must fail when the loaded config does
// not actually declare environment=prod — the CI gate may not silently
// validate under laxer rules.
func TestConfigValidateEnvMismatchFails(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	// prod.yaml mis-declares stage (copy-paste error).
	if err := os.WriteFile(filepath.Join(cfgDir, "prod.yaml"), []byte("environment: stage\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"config", "validate", "--dir", cfgDir, "--env", "prod"}, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit = %d, want 1; stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "prod") || !strings.Contains(stderr.String(), "stage") {
		t.Errorf("error should name both environments: %s", stderr.String())
	}
}

func TestConfigValidateEnvMatchPasses(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "dev.yaml"), []byte("environment: dev\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	var stdout, stderr bytes.Buffer
	code := Run([]string{"config", "validate", "--dir", cfgDir, "--env", "dev"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit = %d, want 0; stderr: %s", code, stderr.String())
	}
}
