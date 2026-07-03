package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Blueprint 12 §5's "CLI snapshot test": with a real Secret field in the
// framework config (db.dsn, Phase 2), every config command's output —
// stdout AND stderr, success AND failure — must carry the redaction marker
// and never the resolved value. Carried from the Phase 1 review as a Phase 2
// exit item.
func TestConfigCommandsNeverPrintSecretValues(t *testing.T) {
	const rawDSN = "postgres://app:sup3rsecret-pw@db:5432/prod"
	t.Setenv("TEST_APP_DSN", rawDSN)

	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	base := "environment: dev\ndb:\n  dsn: secretref://env/TEST_APP_DSN\n"
	if err := os.WriteFile(filepath.Join(cfgDir, "base.yaml"), []byte(base), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, args := range [][]string{
		{"config", "validate", "--dir", cfgDir},
		{"config", "print", "--redacted", "--dir", cfgDir},
		{"config", "doctor", "--dir", cfgDir},
	} {
		var stdout, stderr bytes.Buffer
		code := Run(args, &stdout, &stderr)
		out := stdout.String() + stderr.String()
		if code != 0 {
			t.Fatalf("%v: exit %d\n%s", args, code, out)
		}
		if strings.Contains(out, "sup3rsecret") || strings.Contains(out, rawDSN) {
			t.Errorf("%v leaked the resolved secret:\n%s", args, out)
		}
	}

	// print --redacted must carry the structural marker for the ref.
	var stdout, stderr bytes.Buffer
	if code := Run([]string{"config", "print", "--redacted", "--dir", cfgDir}, &stdout, &stderr); code != 0 {
		t.Fatalf("print: exit %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "[redacted:secretref://env/TEST_APP_DSN]") {
		t.Errorf("print --redacted should render the redaction marker:\n%s", stdout.String())
	}

	// Failure path: unresolvable ref — the error names the ref, never a value.
	badDir := filepath.Join(t.TempDir(), "configs")
	if err := os.MkdirAll(badDir, 0o755); err != nil {
		t.Fatal(err)
	}
	bad := "environment: dev\ndb:\n  dsn: secretref://env/DOES_NOT_EXIST_VAR\n"
	if err := os.WriteFile(filepath.Join(badDir, "base.yaml"), []byte(bad), 0o600); err != nil {
		t.Fatal(err)
	}
	stdout.Reset()
	stderr.Reset()
	if code := Run([]string{"config", "validate", "--dir", badDir}, &stdout, &stderr); code != 1 {
		t.Fatalf("validate with unresolvable ref: exit %d, want 1", code)
	}
	if !strings.Contains(stderr.String(), "DOES_NOT_EXIST_VAR") {
		t.Errorf("error should name the unresolvable ref:\n%s", stderr.String())
	}
}
