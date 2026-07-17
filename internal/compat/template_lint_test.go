package compat

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Negative lint fixtures (fourth closure audit 2026-07-17): the template
// consumer-path lint must actually FAIL on forbidden constructs — a guard
// that has never been seen rejecting anything proves nothing. Three cases:
// an informational Booted field read, an alias of the booted value (which
// would put field reads out of the field check's reach), and a clean
// accessor-only template that must pass.
func TestTemplateLintRejectsForbiddenReads(t *testing.T) {
	script, err := filepath.Abs(filepath.Join("..", "..", "scripts", "lint_templates.sh"))
	if err != nil {
		t.Fatal(err)
	}
	run := func(t *testing.T, content string) (int, string) {
		t.Helper()
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "main.go.tmpl"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command("sh", script, dir)
		out, err := cmd.CombinedOutput()
		code := 0
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else if err != nil {
			t.Fatalf("running lint: %v\n%s", err, out)
		}
		return code, string(out)
	}

	t.Run("informational field read fails", func(t *testing.T) {
		code, out := run(t, "seedReport, err := seeds.Apply(ctx, pool, booted.Seeds, opts)\n")
		if code == 0 {
			t.Fatalf("lint passed a booted.Seeds read:\n%s", out)
		}
		if !strings.Contains(out, "informational Booted field") {
			t.Fatalf("lint output does not name the violation:\n%s", out)
		}
	})

	t.Run("alias of booted fails", func(t *testing.T) {
		code, out := run(t, "b := booted\nmux := b.Router\n")
		if code == 0 {
			t.Fatalf("lint passed an alias of booted:\n%s", out)
		}
		if !strings.Contains(out, "aliasing") {
			t.Fatalf("lint output does not name the alias violation:\n%s", out)
		}
	})

	t.Run("literal framework module path fails", func(t *testing.T) {
		code, out := run(t, "import \"github.com/qatoolist/wowapi/kernel\"\n")
		if code == 0 {
			t.Fatalf("lint passed a literal framework module import:\n%s", out)
		}
		if !strings.Contains(out, "literal framework module path") {
			t.Fatalf("lint output does not name the module-path violation:\n%s", out)
		}
	})

	t.Run("accessor-only template passes", func(t *testing.T) {
		code, out := run(t, "mux := booted.RuntimeRouter().SecureHandler(auth, booted.RuntimeAuthz(), booted.RuntimeTx())\nbooted, err := a.Boot(ctx, k, nil)\n")
		if code != 0 {
			t.Fatalf("lint rejected an accessor-only template:\n%s", out)
		}
	})
}
