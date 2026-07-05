package buildinfo

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersion_LdflagsOverride(t *testing.T) {
	orig := version
	t.Cleanup(func() { version = orig })

	version = "v9.9.9"
	if got := Version(); got != "v9.9.9" {
		t.Fatalf("Version() = %q, want the ldflags override v9.9.9", got)
	}
}

func TestVersion_FallsBackToBuildInfoOrDevel(t *testing.T) {
	orig := version
	t.Cleanup(func() { version = orig })

	version = ""
	// With no ldflags override, Version() returns either the main-module
	// version stamped by the go toolchain or "devel". In `go test` the main
	// module version is empty, so this exercises the "devel" fallback.
	got := Version()
	if got == "" {
		t.Fatal("Version() returned empty string; want a version or \"devel\"")
	}
	if got != "devel" {
		t.Logf("Version() = %q (build-info path, not the devel fallback)", got)
	}
}

func TestIsFramework(t *testing.T) {
	if !(GoMod{ModulePath: ModulePath}).IsFramework() {
		t.Errorf("GoMod with ModulePath=%q should be the framework", ModulePath)
	}
	if (GoMod{ModulePath: "example.com/acme-ops"}).IsFramework() {
		t.Error("a product module path must not be reported as the framework")
	}
}

func writeGoMod(t *testing.T, dir, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindGoMod_BlockRequireForm(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir, "module example.com/acme-ops\n\ngo 1.26\n\nrequire (\n\t"+ModulePath+" v0.4.2\n)\n")

	g, ok := FindGoMod(dir)
	if !ok {
		t.Fatal("FindGoMod did not find the go.mod we just wrote")
	}
	if g.ModulePath != "example.com/acme-ops" {
		t.Errorf("ModulePath = %q, want example.com/acme-ops", g.ModulePath)
	}
	if g.WowapiVersion != "v0.4.2" {
		t.Errorf("WowapiVersion = %q, want v0.4.2", g.WowapiVersion)
	}
	if g.Dir != dir {
		t.Errorf("Dir = %q, want %q", g.Dir, dir)
	}
	if g.IsFramework() {
		t.Error("a consuming module must not be reported as the framework")
	}
}

func TestFindGoMod_InlineRequireForm(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir, "module example.com/acme-ops\n\ngo 1.26\n\nrequire "+ModulePath+" v1.0.0\n")

	g, ok := FindGoMod(dir)
	if !ok {
		t.Fatal("FindGoMod did not find the go.mod")
	}
	if g.WowapiVersion != "v1.0.0" {
		t.Errorf("WowapiVersion = %q, want v1.0.0 (inline require form)", g.WowapiVersion)
	}
}

func TestFindGoMod_FrameworkOwnGoMod(t *testing.T) {
	dir := t.TempDir()
	writeGoMod(t, dir, "module "+ModulePath+"\n\ngo 1.26\n")

	g, ok := FindGoMod(dir)
	if !ok {
		t.Fatal("FindGoMod did not find the go.mod")
	}
	if !g.IsFramework() {
		t.Error("wowapi's own go.mod should be reported as the framework")
	}
	if g.WowapiVersion != "" {
		t.Errorf("WowapiVersion = %q, want empty (framework does not require itself)", g.WowapiVersion)
	}
}

func TestFindGoMod_WalksUpFromSubdir(t *testing.T) {
	root := t.TempDir()
	writeGoMod(t, root, "module example.com/acme-ops\n\ngo 1.26\n")
	sub := filepath.Join(root, "cmd", "api")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	g, ok := FindGoMod(sub)
	if !ok {
		t.Fatal("FindGoMod should walk up to the enclosing go.mod")
	}
	if g.Dir != root {
		t.Errorf("Dir = %q, want the ancestor %q", g.Dir, root)
	}
}

func TestFindGoMod_NotFound(t *testing.T) {
	// A temp dir with no go.mod anywhere up to the filesystem root.
	dir := t.TempDir()
	if _, ok := FindGoMod(dir); ok {
		t.Error("FindGoMod should return ok=false when no go.mod exists above dir")
	}
}

func TestFindGoMod_UnreadableGoMod(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root ignores file permissions; cannot exercise the open-error path")
	}
	dir := t.TempDir()
	p := filepath.Join(dir, "go.mod")
	writeGoMod(t, dir, "module example.com/acme-ops\n")
	if err := os.Chmod(p, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(p, 0o644) }) // let TempDir cleanup remove it

	// go.mod exists (Stat succeeds) but Open fails → parseGoMod errors → FindGoMod
	// reports not-found rather than a partial result.
	if _, ok := FindGoMod(dir); ok {
		t.Error("FindGoMod should return ok=false when the go.mod cannot be opened")
	}
}
