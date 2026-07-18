package compat

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestReleaseIdentityLintDiscriminates(t *testing.T) {
	script, err := filepath.Abs(filepath.Join("..", "..", "scripts", "lint_release_identity.sh"))
	if err != nil {
		t.Fatal(err)
	}
	newFixture := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		for _, dir := range []string{"app", "adapters", "cmd", "foundation", "internal/buildinfo", "kernel", "module", "testkit", "ci", "docs/operations"} {
			if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
				t.Fatal(err)
			}
		}
		files := map[string]string{
			"go.mod": "module github.com/qatoolist/wowapi\n",
			"internal/buildinfo/buildinfo.go": `package buildinfo
const ModulePath = "github.com/qatoolist/wowapi"
`,
			"ci/release-line.json": `{"bootstrap_tag":"v1.2.0"}`,
			"README.md":            "clean v1.2.0\n",
			"CHANGELOG.md":         "clean v1.2.0\n",
			"docs/SRS.md":          "clean v1.2.0\n",
			"docs/operations/upgrade-and-deprecation-policy.md": "clean v1.2.0\n",
		}
		for name, body := range files {
			if err := os.WriteFile(filepath.Join(root, name), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return root
	}
	run := func(root string) error {
		return exec.Command("sh", script, root).Run()
	}

	t.Run("canonical fixture passes", func(t *testing.T) {
		if err := run(newFixture(t)); err != nil {
			t.Fatalf("canonical identity rejected: %v", err)
		}
	})
	t.Run("wrong module fails", func(t *testing.T) {
		root := newFixture(t)
		staleMajor := "github.com/qatoolist/wowapi/" + "v2"
		if err := os.WriteFile(filepath.Join(root, "go.mod"), []byte("module "+staleMajor+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := run(root); err == nil {
			t.Fatal("wrong module identity passed")
		}
	})
	t.Run("stale import fails", func(t *testing.T) {
		root := newFixture(t)
		staleImport := "github.com/qatoolist/wowapi/" + "v2/kernel"
		if err := os.WriteFile(filepath.Join(root, "app", "stale.go"), []byte("package app\nimport _ \""+staleImport+"\"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := run(root); err == nil {
			t.Fatal("stale /v2 import passed")
		}
	})
	t.Run("wrong bootstrap fails", func(t *testing.T) {
		root := newFixture(t)
		if err := os.WriteFile(filepath.Join(root, "ci", "release-line.json"), []byte(`{"bootstrap_tag":"v1.1.0"}`), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := run(root); err == nil {
			t.Fatal("wrong bootstrap identity passed")
		}
	})
}
