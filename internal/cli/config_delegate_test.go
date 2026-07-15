package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestConfigDoctorDiscoversProductRootFromNestedSubdir proves DX-07 T3: when
// invoked from a nested subdirectory of a product repo, `config doctor` uses
// `go env GOMOD` to find the product root, engages the product-local
// tools/configcheck, and reports that product validation ran.
func TestConfigDoctorDiscoversProductRootFromNestedSubdir(t *testing.T) {
	root := makeProductRoot(t)
	subdir := filepath.Join(root, "internal", "appcfg")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatal(err)
	}

	code, out, errOut := runInDir(t, subdir, "config", "doctor", "--dir", filepath.Join(root, "configs"))
	if code != 0 {
		t.Fatalf("exit %d; stdout: %s; stderr: %s", code, out, errOut)
	}
	if !strings.Contains(errOut, "product validation: engaged") {
		t.Fatalf("expected product validation engaged, got stderr: %q", errOut)
	}
	if !strings.Contains(out, "fingerprint=") {
		t.Fatalf("expected fingerprint output, got stdout: %q", out)
	}
}

// TestConfigDoctorDiscoversProductRootFromOutsideRepo proves DX-07 T3: when
// invoked from outside the product repo with --project, `config doctor` uses
// the explicit path to find the product checker, engages it, and reports that
// product validation ran.
func TestConfigDoctorDiscoversProductRootFromOutsideRepo(t *testing.T) {
	root := makeProductRoot(t)
	outside := t.TempDir()

	code, out, errOut := runInDir(t, outside, "config", "doctor", "--dir", filepath.Join(root, "configs"), "--project", root)
	if code != 0 {
		t.Fatalf("exit %d; stdout: %s; stderr: %s", code, out, errOut)
	}
	if !strings.Contains(errOut, "product validation: engaged") {
		t.Fatalf("expected product validation engaged, got stderr: %q", errOut)
	}
	if !strings.Contains(out, "fingerprint=") {
		t.Fatalf("expected fingerprint output, got stdout: %q", out)
	}
}

// TestConfigDoctorReportsSkippedProductValidation proves the explicit reporting
// when no product checker exists: the framework-only path still runs and prints
// a clear "product validation: skipped" message.
func TestConfigDoctorReportsSkippedProductValidation(t *testing.T) {
	dir := t.TempDir()
	writeYAML(t, dir, "base.yaml", "environment: dev\n")

	code, out, errOut := run(t, "config", "doctor", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	if !strings.Contains(errOut, "product validation: skipped") {
		t.Fatalf("expected skipped product validation message, got stderr: %q", errOut)
	}
	if !strings.Contains(out, "fingerprint=") {
		t.Fatalf("expected fingerprint output, got stdout: %q", out)
	}
}

// makeProductRoot creates a minimal Go module with a tools/configcheck/main.go
// that the CLI can delegate to. The checker is intentionally tiny: it validates
// that it received the expected subcommand and prints a fingerprint line so the
// test can confirm delegation succeeded.
func makeProductRoot(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	writeFile(t, root, "go.mod", `module example.com/product

go 1.22
`)
	if err := os.MkdirAll(filepath.Join(root, "configs"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, root, "configs/base.yaml", "environment: dev\n")
	checkerDir := filepath.Join(root, "tools", "configcheck")
	if err := os.MkdirAll(checkerDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, checkerDir, "main.go", `package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 1 || os.Args[1] != "doctor" {
		fmt.Fprintf(os.Stderr, "unexpected mode: %v\n", os.Args)
		os.Exit(2)
	}
	fmt.Println("KEY\tLAYER")
	fmt.Println("environment\tbase-file")
	fmt.Println("http.addr\tdefault")
	fmt.Println("fingerprint=product-checker")
}
`)
	return root
}

// runInDir runs the CLI with the given args after changing into dir. The
// directory is restored after the call.
func runInDir(t *testing.T, dir string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Chdir(orig); err != nil {
			t.Errorf("restore wd: %v", err)
		}
	}()
	return run(t, args...)
}
