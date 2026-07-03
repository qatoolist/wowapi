package cli

import (
	"bytes"
	"strings"
	"testing"
)

func run(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	code = Run(args, &out, &errBuf)
	return code, out.String(), errBuf.String()
}

func TestVersionCommand(t *testing.T) {
	code, out, _ := run(t, "version")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	if !strings.HasPrefix(out, "wowapi ") {
		t.Errorf("unexpected output: %q", out)
	}
	// Running inside the framework repo must be detected (go.mod module path).
	if !strings.Contains(out, "framework repository") {
		t.Errorf("expected framework-repo context line, got: %q", out)
	}
}

func TestHelpListsPlannedCommands(t *testing.T) {
	code, out, _ := run(t, "help")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	for _, cmd := range []string{"version", "init", "new-module", "gen", "seed", "openapi", "lint", "config", "deploy"} {
		if !strings.Contains(out, cmd) {
			t.Errorf("help missing %q", cmd)
		}
	}
}

func TestNewModuleRequiresName(t *testing.T) {
	// All Phase-10 commands are now implemented; new-module without a --name is a
	// usage error, not a "planned" stub.
	code, _, errOut := run(t, "new-module")
	if code == 0 {
		t.Fatalf("new-module with no --name should fail")
	}
	if !strings.Contains(errOut, "name") {
		t.Errorf("expected a --name error, got: %q", errOut)
	}
}

func TestUnknownCommand(t *testing.T) {
	code, _, errOut := run(t, "frobnicate")
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	if !strings.Contains(errOut, "unknown command") {
		t.Errorf("message: %q", errOut)
	}
}
