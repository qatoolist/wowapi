package cli

// W01-E04-S001-T002 (DX-01 T5) — the isolated-temp-dir generate→build→boot→
// smoke harness, and its own proving runs (AC-W01-E04-S001-02).
//
// This is the shared E2E primitive the story's governing instruction calls
// out ("T5 harness = shared primitive for DX-02/DX-04"): a real `wowapi`
// BINARY (not an in-process call) scaffolds a product into a fresh temp dir,
// then the full developer pipeline runs against it — `go mod tidy` → `go mod
// download` → `go build ./...` → a boot-and-validate smoke test — with every
// step's full output captured, so a failure names the exact step and shows
// its output rather than a bare red/green signal.
//
// Two CLI invocation paths are proven:
//
//   - source-built CLI ("devel"): the sanctioned source workflow, `init
//     --local-framework <checkout>` (an explicit replace directive; fully
//     offline). A flag-less devel invocation is also asserted to fail closed
//     pre-write — the pipeline's old failure mode (an unresolvable version
//     discovered only at `go mod download`, see
//     evidence/DX-01/t1-t4-prefix-failfirst.log) is now impossible.
//
//   - released CLI: a binary stamped with a release version via -ldflags
//     (exactly how the release pipeline stamps it), whose flag-less `init`
//     pins that version in go.mod. `go mod download` then fetches the
//     framework AT that version from a local file:// module proxy packaged
//     from this checkout — hermetic (no network), yet exercising the same
//     resolution path a published release does.
//
// DX-02's generator-output-boots test (gen_crud_boots_test.go) reuses the
// same underlying scaffold primitive (buildRenderedProduct); a future DX-04
// golden-consumer story is expected to call scaffoldPipeline with different
// generator arguments/assertions, per the story's forward reference.

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// e2eReleaseVersion is the release version the "released CLI" stand-in is
// stamped with, and the version the local file proxy serves this checkout as.
// e2eReleaseVersion is a release version that does not collide with any
// published tag or module-cache entry, so the released-CLI path is forced to
// resolve the framework from the local file proxy packaged from this checkout.
const e2eReleaseVersion = "v0.2.0-w06shared.1"

// buildWowapiCLI compiles cmd/wowapi into a temp dir and returns the binary
// path. ldflagsVersion != "" stamps a release version exactly as the release
// pipeline does; "" yields a source build. -buildvcs=false keeps the binary's
// version deterministic ("devel"/the stamp) regardless of the checkout's
// current VCS state, so the harness result never depends on whether the tree
// happens to be clean or pushed.
func buildWowapiCLI(t *testing.T, ldflagsVersion string) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "wowapi")
	args := []string{"build", "-buildvcs=false", "-o", bin}
	if ldflagsVersion != "" {
		args = append(args, "-ldflags",
			"-X github.com/qatoolist/wowapi/internal/buildinfo.version="+ldflagsVersion)
	}
	args = append(args, "./cmd/wowapi")
	cmd := exec.Command("go", args...)
	cmd.Dir = wowapiCheckoutRoot(t)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("building wowapi CLI (%q): %v\n%s", ldflagsVersion, err, out)
	}
	return bin
}

// modCacheProxyURL primes the local module cache and returns its download
// directory as a file:// proxy URL. Compiled build-cache entries do not prove
// that the corresponding module proxy ZIPs are present (notably after a
// setup-go cache restore), so prime the proxy material explicitly before the
// generated consumer is restricted to offline resolution.
func modCacheProxyURL(t *testing.T) string {
	t.Helper()
	root := wowapiCheckoutRoot(t)
	gomod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatalf("read go.mod for module-cache proxy: %v", err)
	}
	gosum, err := os.ReadFile(filepath.Join(root, "go.sum"))
	if err != nil {
		t.Fatalf("read go.sum for module-cache proxy: %v", err)
	}
	primeModuleCacheProxy(t, gomod, gosum)

	out, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		t.Fatalf("go env GOMODCACHE: %v", err)
	}
	return "file://" + filepath.ToSlash(filepath.Join(strings.TrimSpace(string(out)), "cache", "download"))
}

// primeModuleCacheProxy downloads a complete module graph from a disposable
// module so priming cannot rewrite the repository's go.sum.
func primeModuleCacheProxy(t *testing.T, gomod, gosum []byte) {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), gomod, 0o644); err != nil {
		t.Fatal(err)
	}
	if len(gosum) != 0 {
		if err := os.WriteFile(filepath.Join(dir, "go.sum"), gosum, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	download := exec.Command("go", "mod", "download", "all")
	download.Dir = dir
	download.Env = append(os.Environ(), "GOWORK=off")
	if out, err := download.CombinedOutput(); err != nil {
		t.Fatalf("prime local module-cache proxy: %v\n%s", err, out)
	}
}

// primeReleasedModuleCacheProxy makes the full dependency graph of a tagged
// framework release available before golden-consumer resolution goes offline.
func primeReleasedModuleCacheProxy(t *testing.T, version string) {
	t.Helper()
	gomod := fmt.Appendf(nil, "module example.com/wowapi-proxy-prime\n\ngo 1.26.0\n\nrequire github.com/qatoolist/wowapi %s\n", version)
	primeModuleCacheProxy(t, gomod, nil)
}

// buildFrameworkProxy packages THIS checkout as module version `version`
// inside a file:// GOPROXY directory (list, .info, .mod, .zip), so a
// released-CLI scaffold's `go mod download github.com/qatoolist/wowapi@version`
// succeeds hermetically. Only the files a consumer build needs are zipped.
// purgeCachedFrameworkVersion removes any previously downloaded/extracted copy
// of the synthetic framework version from the shared GOMODCACHE. The proxy zips
// the CURRENT checkout under a CONSTANT version string, so a cached copy from
// an earlier run can silently serve stale framework code to consumer builds —
// masking working-tree changes with bogus compile errors (observed during the
// 2026-07-17 adversarial-review remediation, on both the host cache and the
// toolbox container's gomod volume).
func purgeCachedFrameworkVersion(t *testing.T, version string) {
	t.Helper()
	out, err := exec.Command("go", "env", "GOMODCACHE").Output()
	if err != nil {
		t.Fatalf("go env GOMODCACHE: %v", err)
	}
	gmc := strings.TrimSpace(string(out))
	if gmc == "" {
		return
	}
	targets := []string{
		filepath.Join(gmc, "github.com", "qatoolist", "wowapi@"+version),
	}
	dl := filepath.Join(gmc, "cache", "download", "github.com", "qatoolist", "wowapi", "@v")
	if entries, err := os.ReadDir(dl); err == nil {
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), version+".") {
				targets = append(targets, filepath.Join(dl, e.Name()))
			}
		}
	}
	for _, path := range targets {
		if _, err := os.Stat(path); err != nil {
			continue
		}
		// Module-cache contents are read-only; make writable before removal.
		_ = filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
			if err == nil {
				_ = os.Chmod(p, 0o755)
			}
			return nil
		})
		_ = os.Chmod(filepath.Dir(path), 0o755)
		if err := os.RemoveAll(path); err != nil {
			t.Logf("purge cached framework %s: %v", path, err)
		}
	}
}

func buildFrameworkProxy(t *testing.T, version string) string {
	t.Helper()
	purgeCachedFrameworkVersion(t, version)
	root := wowapiCheckoutRoot(t)
	proxy := t.TempDir()
	vdir := filepath.Join(proxy, filepath.FromSlash("github.com/qatoolist/wowapi/@v"))
	if err := os.MkdirAll(vdir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vdir, "list"), []byte(version+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	info := fmt.Sprintf("{\"Version\":%q,\"Time\":\"2026-07-13T00:00:00Z\"}\n", version)
	if err := os.WriteFile(filepath.Join(vdir, version+".info"), []byte(info), 0o644); err != nil {
		t.Fatal(err)
	}
	gomod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vdir, version+".mod"), gomod, 0o644); err != nil {
		t.Fatal(err)
	}

	zf, err := os.Create(filepath.Join(vdir, version+".zip"))
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(zf)
	prefix := "github.com/qatoolist/wowapi@" + version + "/"
	// Whitelist of top-level entries a consumer build needs — module metadata
	// plus every Go source tree. A future top-level Go package missing here
	// fails the harness build step loudly, with the missing import named.
	include := []string{
		"go.mod", "go.sum", "LICENSE", "NOTICE", "README.md",
		"adapters", "app", "cmd", "foundation", "internal", "kernel", "migrations", "module", "testkit",
	}
	for _, entry := range include {
		abs := filepath.Join(root, entry)
		st, err := os.Stat(abs)
		if err != nil {
			t.Fatalf("framework proxy zip: missing expected entry %s: %v", entry, err)
		}
		if !st.IsDir() {
			addZipFile(t, zw, abs, prefix+entry)
			continue
		}
		err = filepath.WalkDir(abs, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			name := d.Name()
			if d.IsDir() {
				if strings.HasPrefix(name, ".") {
					return filepath.SkipDir
				}
				if path != abs {
					if _, nestedErr := os.Stat(filepath.Join(path, "go.mod")); nestedErr == nil {
						return filepath.SkipDir
					} else if !os.IsNotExist(nestedErr) {
						return nestedErr
					}
				}
				return nil
			}
			if !d.Type().IsRegular() || strings.HasPrefix(name, ".") {
				return nil // module zips must not contain symlinks or oddities
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			addZipFile(t, zw, path, prefix+filepath.ToSlash(rel))
			return nil
		})
		if err != nil {
			t.Fatalf("framework proxy zip: walking %s: %v", entry, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zf.Close(); err != nil {
		t.Fatal(err)
	}
	return "file://" + filepath.ToSlash(proxy)
}

func addZipFile(t *testing.T, zw *zip.Writer, src, name string) {
	t.Helper()
	f, err := os.Open(src)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }() // read-only source file
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(w, f); err != nil {
		t.Fatal(err)
	}
}

// runPipelineStep executes one pipeline step in dir with the given extra env,
// failing the test with the step's name and FULL output on any error — the
// diagnostic contract task-002 requires (a failure names its step, never a
// bare red signal).
func runPipelineStep(t *testing.T, step, dir string, env []string, name string, args ...string) string {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pipeline step %q failed: %v\n--- %s %s output ---\n%s",
			step, err, name, strings.Join(args, " "), out)
	}
	t.Logf("pipeline step %q ok", step)
	return string(out)
}

// bootSmokeSource is the minimal boot-and-validate smoke test written into a
// scaffolded product (which ships no tests of its own): kernel.New →
// app.Register(wire.Modules()) → app.Boot, exercising the full registration-
// validation gate with no database (a no-op TxManager satisfies kernel.Deps).
const bootSmokeSource = `package boottest

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/qatoolist/wowapi/app"
	"github.com/qatoolist/wowapi/kernel"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/storage"

	"MODULE_PATH/internal/wire"
)

type noopTxM struct{}

var errNoDB = errors.New("boot smoke runs without a database")

func (noopTxM) WithTenant(context.Context, func(context.Context, database.TenantDB) error) error {
	return errNoDB
}

func (noopTxM) WithTenantRO(context.Context, func(context.Context, database.TenantDB) error) error {
	return errNoDB
}

func (noopTxM) Platform(context.Context, func(context.Context, database.DB) error) error {
	return errNoDB
}

func TestScaffoldBoots(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	k, err := kernel.New(config.Defaults(), log, kernel.Deps{Tx: noopTxM{}, Storage: storage.NewMemory()})
	if err != nil {
		t.Fatalf("kernel.New: %v", err)
	}
	a := app.New()
	a.Register(wire.Modules()...)
	if _, err := a.Boot(context.Background(), k, nil); err != nil {
		t.Fatalf("boot: %v", err)
	}
}
`

// scaffoldPipeline is the reusable generate→build→boot→smoke primitive: it
// runs `cli init` (a real subprocess) with the given args into a fresh temp
// dir, then drives the full developer pipeline inside it. goEnv configures
// module resolution for the go steps (proxy chain, sumdb, workspace mode).
// It returns the product directory for further assertions or generator steps
// (e.g. a future DX-04 caller running `gen crud` before the build step).
func scaffoldPipeline(t *testing.T, cli, modulePath string, initArgs, goEnv []string) string {
	t.Helper()
	dir := t.TempDir()

	args := append([]string{"init", "--module", modulePath, "--dir", dir}, initArgs...)
	initCmd := exec.Command(cli, args...)
	out, err := initCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("pipeline step \"init\" failed: %v\n--- wowapi %s output ---\n%s",
			err, strings.Join(args, " "), out)
	}
	t.Logf("pipeline step \"init\" ok")

	runPipelineStep(t, "go mod tidy", dir, goEnv, "go", "mod", "tidy")
	runPipelineStep(t, "go mod download", dir, goEnv, "go", "mod", "download")
	runPipelineStep(t, "go build ./...", dir, goEnv, "go", "build", "./...")

	bootDir := filepath.Join(dir, "internal", "boottest")
	if err := os.MkdirAll(bootDir, 0o755); err != nil {
		t.Fatal(err)
	}
	smoke := strings.ReplaceAll(bootSmokeSource, "MODULE_PATH", modulePath)
	if err := os.WriteFile(filepath.Join(bootDir, "boot_test.go"), []byte(smoke), 0o644); err != nil {
		t.Fatal(err)
	}
	runPipelineStep(t, "boot smoke test", dir, goEnv, "go", "test", "./internal/boottest/")
	return dir
}

// hermeticGoEnv pins module resolution for the pipeline's go steps to the
// given file:// proxy chain, neutralizing developer-machine overrides — on a
// contributor workstation GOPRIVATE/GONOPROXY typically route
// github.com/qatoolist/* straight to VCS, which would bypass the harness
// proxy and reintroduce a network dependency.
func hermeticGoEnv(proxy string) []string {
	return []string{
		"GOWORK=off",
		"GOFLAGS=-mod=mod",
		"GOSUMDB=off",
		"GOPRIVATE=",
		"GONOPROXY=",
		"GONOSUMDB=",
		"GOPROXY=" + proxy,
	}
}

// TestE2EScaffoldSourceBuiltCLI proves the source-built ("devel") invocation
// path end to end, and that its old silent-failure mode is gone: a flag-less
// devel init fails closed pre-write, and the sanctioned `--local-framework`
// workflow completes the whole pipeline offline.
func TestE2EScaffoldSourceBuiltCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the CLI and a full scaffolded product; skipped in -short")
	}
	cli := buildWowapiCLI(t, "") // devel build

	// Guard: the pre-T001 failure mode (init succeeds, pipeline dies at `go
	// mod download` on an unresolvable version) must be impossible — a
	// flag-less devel init fails closed BEFORE any file is written.
	empty := t.TempDir()
	failCmd := exec.Command(cli, "init", "--module", "github.com/acme/e2efailclosed", "--dir", empty)
	out, err := failCmd.CombinedOutput()
	if err == nil {
		t.Fatalf("flag-less devel init must fail closed, but succeeded:\n%s", out)
	}
	if entries, _ := os.ReadDir(empty); len(entries) != 0 {
		t.Fatalf("flag-less devel init failed but wrote files first: %v", entries)
	}
	if !strings.Contains(string(out), "remediation") {
		t.Errorf("fail-closed init output missing remediation:\n%s", out)
	}

	goEnv := hermeticGoEnv(modCacheProxyURL(t)) // offline: deps from the local module cache
	dir := scaffoldPipeline(t, cli, "github.com/acme/e2esource",
		[]string{"--local-framework", wowapiCheckoutRoot(t)}, goEnv)

	gomod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gomod), "replace github.com/qatoolist/wowapi => ") {
		t.Errorf("source-path go.mod missing the local replace directive:\n%s", gomod)
	}
}

// TestE2EScaffoldReleasedCLI proves the released-CLI invocation path end to
// end: a release-stamped binary's flag-less init pins its own version, and
// the pipeline resolves the framework AT that version from a module proxy
// (a local file:// proxy packaged from this checkout — hermetic).
func TestE2EScaffoldReleasedCLI(t *testing.T) {
	if testing.Short() {
		t.Skip("builds the CLI, a module proxy, and a full scaffolded product; skipped in -short")
	}
	cli := buildWowapiCLI(t, e2eReleaseVersion)
	frameworkProxy := buildFrameworkProxy(t, e2eReleaseVersion)

	goEnv := hermeticGoEnv(frameworkProxy + "," + modCacheProxyURL(t))
	dir := scaffoldPipeline(t, cli, "github.com/acme/e2erelease", nil, goEnv)

	gomod, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(gomod), "github.com/qatoolist/wowapi "+e2eReleaseVersion) {
		t.Errorf("released-path go.mod must pin the CLI's release version:\n%s", gomod)
	}
	if strings.Contains(string(gomod), "replace ") {
		t.Errorf("released-path go.mod must not carry a replace directive:\n%s", gomod)
	}
}
