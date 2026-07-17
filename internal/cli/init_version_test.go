package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// W01-E04-S001-T001 (DX-01 T1-T4) — fail-closed, pre-write framework-version
// resolution for `wowapi init`.
//
// Fail-first history: before this task, init_cmd.go:122-123 unconditionally
// substituted the literal `v0.0.0` for a devel build's version and SUCCEEDED,
// writing a go.mod whose framework requirement can never resolve ("unknown
// revision v0.0.0" at the very next `go mod download`) — captured in
// evidence/DX-01/t1-t4-prefix-failfirst.log. These tests pin the replacement
// behavior: every resolution path either produces a verified, resolvable
// go.mod shape or fails closed BEFORE any file is written.

// callInitRaw invokes runInit with the args exactly as given — no test-harness
// --local-framework injection (contrast callInit) — so the version-resolution
// paths themselves are what is under test.
func callInitRaw(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	code = runInit(args, &out, &errBuf)
	return code, out.String(), errBuf.String()
}

// withVCSInfo stubs the binary's VCS build stamp for the duration of the test.
func withVCSInfo(t *testing.T, revision string, modified, ok bool) {
	t.Helper()
	orig := initVCSInfo
	initVCSInfo = func() (string, bool, bool) { return revision, modified, ok }
	t.Cleanup(func() { initVCSInfo = orig })
}

// withResolver stubs the `go list -m` module-resolution subprocess.
func withResolver(t *testing.T, fn func(module, query string) (string, error)) {
	t.Helper()
	orig := resolveModuleVersion
	resolveModuleVersion = fn
	t.Cleanup(func() { resolveModuleVersion = orig })
}

// withBuildVersion stubs the version stamped into this binary.
func withBuildVersion(t *testing.T, v string) {
	t.Helper()
	orig := initBuildVersion
	initBuildVersion = func() string { return v }
	t.Cleanup(func() { initBuildVersion = orig })
}

// assertNoFilesWritten enforces the fail-CLOSED half of every failure path:
// init must exit before its first file write.
func assertNoFilesWritten(t *testing.T, dir string) {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("init failed but wrote files first (must fail closed pre-write): %v", names)
	}
}

// --- default path (no flags, devel build): DX-01 T3 + T4 ---

// TestInitDevelNoVCSInfoFailsClosed: a source build without a VCS stamp (the
// natural state of a test binary, stubbed here for determinism) has nothing
// to derive a version from — init must fail closed with remediation, never
// fall back to a placeholder. This is the direct replacement for the deleted
// `devel` → `v0.0.0` fallback.
func TestInitDevelNoVCSInfoFailsClosed(t *testing.T) {
	withVCSInfo(t, "", false, false)
	dir := t.TempDir()
	code, _, errOut := callInitRaw(t, "--module", "github.com/acme/app", "--dir", dir)
	if code == 0 {
		t.Fatal("devel-build init with no flags and no VCS metadata must fail closed")
	}
	assertNoFilesWritten(t, dir)
	for _, want := range []string{"--framework-version", "--local-framework", "remediation"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr missing remediation %q:\n%s", want, errOut)
		}
	}
}

// TestInitDevelDirtyTreeFailsClosed: a devel build stamped vcs.modified=true
// cannot name an exact commit — fail closed, no placeholder.
func TestInitDevelDirtyTreeFailsClosed(t *testing.T) {
	withVCSInfo(t, "abcdefabcdefabcdefabcdefabcdefabcdefabcd", true, true)
	dir := t.TempDir()
	code, _, errOut := callInitRaw(t, "--module", "github.com/acme/app", "--dir", dir)
	if code == 0 {
		t.Fatal("devel-build init from a dirty tree must fail closed")
	}
	assertNoFilesWritten(t, dir)
	if !strings.Contains(errOut, "dirty") || !strings.Contains(errOut, "remediation") {
		t.Errorf("stderr should name the dirty tree and a remediation:\n%s", errOut)
	}
}

// TestInitDevelUnreachableCommitFailsClosed: a clean devel build whose commit
// the go tool cannot resolve (e.g. never pushed) must fail closed.
func TestInitDevelUnreachableCommitFailsClosed(t *testing.T) {
	withVCSInfo(t, "abcdefabcdefabcdefabcdefabcdefabcdefabcd", false, true)
	withResolver(t, func(module, query string) (string, error) {
		return "", os.ErrNotExist
	})
	dir := t.TempDir()
	code, _, errOut := callInitRaw(t, "--module", "github.com/acme/app", "--dir", dir)
	if code == 0 {
		t.Fatal("devel-build init at an unresolvable commit must fail closed")
	}
	assertNoFilesWritten(t, dir)
	if !strings.Contains(errOut, "abcdefabcdefab"[:12]) || !strings.Contains(errOut, "remediation") {
		t.Errorf("stderr should name the commit and a remediation:\n%s", errOut)
	}
}

// TestInitDevelCleanReachableDerivesVersion: the success half of the default
// path — a clean, resolvable commit yields the go tool's own canonical
// pseudo-version in the generated go.mod, and never the bare `v0.0.0`.
func TestInitDevelCleanReachableDerivesVersion(t *testing.T) {
	const rev = "abcdefabcdefabcdefabcdefabcdefabcdefabcd"
	const canonical = "v0.3.1-0.20260713054412-abcdefabcdef"
	withVCSInfo(t, rev, false, true)
	withResolver(t, func(module, query string) (string, error) {
		if module != "github.com/qatoolist/wowapi" || query != rev {
			t.Errorf("resolver called with %s@%s, want the framework module at the stamped revision", module, query)
		}
		return canonical, nil
	})
	dir := t.TempDir()
	code, _, errOut := callInitRaw(t, "--module", "github.com/acme/app", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	gomod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gomod), "github.com/qatoolist/wowapi "+canonical) {
		t.Errorf("go.mod missing derived canonical version %q:\n%s", canonical, gomod)
	}
	if strings.Contains(string(gomod), "wowapi v0.0.0\n") {
		t.Errorf("go.mod carries the removed bare v0.0.0 placeholder:\n%s", gomod)
	}
}

// --- --framework-version: DX-01 T1 ---

// TestInitFrameworkVersionUnresolvableFailsClosed drives the REAL
// goResolveModuleVersion subprocess with GOPROXY=off, so resolution fails
// deterministically offline: init must fail closed pre-write and print the
// exact version-discovery command.
func TestInitFrameworkVersionUnresolvableFailsClosed(t *testing.T) {
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOFLAGS", "-mod=mod")
	dir := t.TempDir()
	code, _, errOut := callInitRaw(t,
		"--module", "github.com/acme/app", "--dir", dir,
		"--framework-version", "v9.9.9-does-not-exist")
	if code == 0 {
		t.Fatal("unresolvable --framework-version must fail closed")
	}
	assertNoFilesWritten(t, dir)
	if !strings.Contains(errOut, "go list -m -versions github.com/qatoolist/wowapi") {
		t.Errorf("stderr missing the exact version-discovery remediation command:\n%s", errOut)
	}
}

// TestInitFrameworkVersionVerifiedIsWritten: the success half — the verified
// (canonicalized) version is what lands in go.mod, with no replace directive.
func TestInitFrameworkVersionVerifiedIsWritten(t *testing.T) {
	withResolver(t, func(module, query string) (string, error) {
		if query != "v1.2.3" {
			t.Errorf("resolver query = %q, want the flag value", query)
		}
		return "v1.2.3", nil
	})
	dir := t.TempDir()
	code, _, errOut := callInitRaw(t,
		"--module", "github.com/acme/app", "--dir", dir,
		"--framework-version", "v1.2.3")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	gomod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gomod), "github.com/qatoolist/wowapi v1.2.3") {
		t.Errorf("go.mod missing verified version:\n%s", gomod)
	}
	if strings.Contains(string(gomod), "replace ") {
		t.Errorf("explicit-version path must not emit a replace directive:\n%s", gomod)
	}
}

// --- --local-framework: DX-01 T2 ---

func TestInitLocalFrameworkRelativePathFailsClosed(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInitRaw(t,
		"--module", "github.com/acme/app", "--dir", dir,
		"--local-framework", "relative/wowapi")
	if code == 0 {
		t.Fatal("relative --local-framework must fail closed")
	}
	assertNoFilesWritten(t, dir)
	if !strings.Contains(errOut, "absolute") {
		t.Errorf("stderr should demand an absolute path:\n%s", errOut)
	}
}

func TestInitLocalFrameworkNonexistentPathFailsClosed(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInitRaw(t,
		"--module", "github.com/acme/app", "--dir", dir,
		"--local-framework", filepath.Join(t.TempDir(), "no-such-checkout"))
	if code == 0 {
		t.Fatal("nonexistent --local-framework must fail closed")
	}
	assertNoFilesWritten(t, dir)
	if !strings.Contains(errOut, "existing directory") {
		t.Errorf("stderr should say the directory does not exist:\n%s", errOut)
	}
}

func TestInitLocalFrameworkNonFrameworkDirFailsClosed(t *testing.T) {
	notFramework := t.TempDir()
	if err := os.WriteFile(filepath.Join(notFramework, "go.mod"), []byte("module example.com/other\n\ngo 1.26\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	code, _, errOut := callInitRaw(t,
		"--module", "github.com/acme/app", "--dir", dir,
		"--local-framework", notFramework)
	if code == 0 {
		t.Fatal("--local-framework at a non-wowapi module must fail closed")
	}
	assertNoFilesWritten(t, dir)
	if !strings.Contains(errOut, "github.com/qatoolist/wowapi") {
		t.Errorf("stderr should name the expected framework module:\n%s", errOut)
	}
}

// TestInitLocalFrameworkWritesReplaceAndWarns: the success half — a real
// checkout path yields an explicit replace directive plus a visible dev-mode
// warning; the require line uses the inert canonical placeholder, never a
// bare unresolvable version.
func TestInitLocalFrameworkWritesReplaceAndWarns(t *testing.T) {
	root := wowapiCheckoutRoot(t)
	dir := t.TempDir()
	code, _, errOut := callInitRaw(t,
		"--module", "github.com/acme/app", "--dir", dir,
		"--local-framework", root)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if !strings.Contains(errOut, "dev mode") {
		t.Errorf("stderr missing the visible dev-mode warning:\n%s", errOut)
	}
	gomod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gomod), "replace github.com/qatoolist/wowapi => "+root) {
		t.Errorf("go.mod missing the replace directive:\n%s", gomod)
	}
	if !strings.Contains(string(gomod), "github.com/qatoolist/wowapi "+localReplaceVersion) {
		t.Errorf("go.mod require line should carry the inert canonical placeholder:\n%s", gomod)
	}
}

// --- flag interaction ---

func TestInitBothVersionFlagsRejected(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInitRaw(t,
		"--module", "github.com/acme/app", "--dir", dir,
		"--framework-version", "v1.2.3",
		"--local-framework", wowapiCheckoutRoot(t))
	if code == 0 {
		t.Fatal("passing both version flags must be rejected")
	}
	assertNoFilesWritten(t, dir)
	if !strings.Contains(errOut, "mutually exclusive") {
		t.Errorf("stderr should say the flags are mutually exclusive:\n%s", errOut)
	}
}

// --- the real resolver subprocess, hermetic ---

// TestGoResolveModuleVersionFromModuleCache proves the real `go list -m`
// resolution path end-to-end without network: an exact version already in the
// local module cache (a pinned dependency of this repo) resolves under
// GOPROXY=off, and a bogus version fails.
func TestGoResolveModuleVersionFromModuleCache(t *testing.T) {
	if testing.Short() {
		t.Skip("spawns go subprocesses; skipped in -short")
	}
	t.Setenv("GOPROXY", "off")
	t.Setenv("GOFLAGS", "-mod=mod")
	v, err := goResolveModuleVersion("github.com/google/uuid", "v1.6.0")
	if err != nil {
		t.Fatalf("cached exact version should resolve offline: %v", err)
	}
	if v != "v1.6.0" {
		t.Fatalf("resolved %q, want v1.6.0", v)
	}
	if _, err := goResolveModuleVersion("github.com/google/uuid", "v99.99.99"); err == nil {
		t.Fatal("bogus version must not resolve")
	}
}

// --- Go 1.24+ stamped main-module versions (the SF-7 defect shape) ---

// TestInitStampedDirtyVersionFailsClosed is THE SF-7 regression test: a
// locally built (`go build`) CLI from a dirty tree is stamped with a
// `…+dirty` pseudo-version — before this task, init wrote that unresolvable
// string straight into the scaffold's go.mod (captured in
// evidence/DX-01/t1-t4-prefix-failfirst.log). It must fail closed instead.
func TestInitStampedDirtyVersionFailsClosed(t *testing.T) {
	withBuildVersion(t, "v1.0.1-0.20260713072141-05dce5c8a548+dirty")
	dir := t.TempDir()
	code, _, errOut := callInitRaw(t, "--module", "github.com/acme/app", "--dir", dir)
	if code == 0 {
		t.Fatal("init from a +dirty-stamped CLI must fail closed")
	}
	assertNoFilesWritten(t, dir)
	if !strings.Contains(errOut, "dirty") || !strings.Contains(errOut, "remediation") {
		t.Errorf("stderr should name the dirty stamp and a remediation:\n%s", errOut)
	}
}

// TestInitStampedPseudoVersionVerifiedBeforeWrite: a clean `go build` stamp is
// an exact pseudo-version, but only usable if the go tool can resolve it —
// verified pre-write, written on success, failed closed otherwise.
func TestInitStampedPseudoVersionVerifiedBeforeWrite(t *testing.T) {
	const stamped = "v1.0.1-0.20260713072141-05dce5c8a548"
	t.Run("resolvable", func(t *testing.T) {
		withBuildVersion(t, stamped)
		withResolver(t, func(module, query string) (string, error) {
			if query != stamped {
				t.Errorf("resolver query = %q, want the stamped pseudo-version", query)
			}
			return stamped, nil
		})
		dir := t.TempDir()
		code, _, errOut := callInitRaw(t, "--module", "github.com/acme/app", "--dir", dir)
		if code != 0 {
			t.Fatalf("exit %d: %s", code, errOut)
		}
		gomod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(gomod), "github.com/qatoolist/wowapi "+stamped) {
			t.Errorf("go.mod missing verified stamped pseudo-version:\n%s", gomod)
		}
	})
	t.Run("unresolvable", func(t *testing.T) {
		withBuildVersion(t, stamped)
		withResolver(t, func(module, query string) (string, error) {
			return "", os.ErrNotExist
		})
		dir := t.TempDir()
		code, _, errOut := callInitRaw(t, "--module", "github.com/acme/app", "--dir", dir)
		if code == 0 {
			t.Fatal("unresolvable stamped pseudo-version must fail closed")
		}
		assertNoFilesWritten(t, dir)
		if !strings.Contains(errOut, "remediation") {
			t.Errorf("stderr missing remediation:\n%s", errOut)
		}
	})
}

// TestInitStampedReleaseVersionUsedAsIs: a real tagged release (`go install
// …@vX.Y.Z`) is resolvable by construction and used directly — no resolver
// subprocess, no replace directive.
func TestInitStampedReleaseVersionUsedAsIs(t *testing.T) {
	withBuildVersion(t, "v1.4.2")
	withResolver(t, func(module, query string) (string, error) {
		t.Errorf("resolver must not be called for a tagged release stamp (called with %s@%s)", module, query)
		return "", os.ErrInvalid
	})
	dir := t.TempDir()
	code, _, errOut := callInitRaw(t, "--module", "github.com/acme/app", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	gomod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gomod), "github.com/qatoolist/wowapi v1.4.2") {
		t.Errorf("go.mod missing release version:\n%s", gomod)
	}
	if strings.Contains(string(gomod), "replace ") {
		t.Errorf("release path must not emit a replace directive:\n%s", gomod)
	}
}
