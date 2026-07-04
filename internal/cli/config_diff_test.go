package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigDiffBetweenEnvs(t *testing.T) {
	cfgDir := filepath.Join(t.TempDir(), "configs")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(cfgDir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("base.yaml", "log:\n  level: warn\n")
	write("local.yaml", "environment: local\n")
	write("dev.yaml", "environment: dev\nlog:\n  level: debug\n")

	var out, errb bytes.Buffer
	code := Run([]string{"config", "diff", "--dir", cfgDir, "--from", "local", "--to", "dev"}, &out, &errb)
	if code != 0 {
		t.Fatalf("exit = %d, want 0; stderr: %s", code, errb.String())
	}
	got := out.String()
	if !strings.Contains(got, "--- local") || !strings.Contains(got, "+++ dev") {
		t.Fatalf("diff missing env headers:\n%s", got)
	}
	// The environment and the overridden log level must show up as changes.
	if !strings.Contains(got, "dev") || !strings.Contains(got, "debug") {
		t.Fatalf("diff did not surface the changed values:\n%s", got)
	}
}

func TestConfigDiffRequiresBothEnvs(t *testing.T) {
	var out, errb bytes.Buffer
	if code := Run([]string{"config", "diff", "--from", "dev"}, &out, &errb); code != 2 {
		t.Fatalf("missing --to should exit 2, got %d", code)
	}
}
