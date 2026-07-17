// lint_cmd.go — `wowapi lint boundaries` (Phase 10) and `wowapi lint
// lifecycle` (backlog B9). boundaries enforces the import law in a product
// repo the same way scripts/lint_boundaries.sh does for the framework:
// modules are isolated (a module never imports another module's internals —
// they collaborate through ports), and in the framework repo the
// kernel/module/app/adapters layering holds. lifecycle checks the static
// provider/lifecycle manifest (kernel/lifecycle) for wiring mistakes (scope
// leaks, raw pools, tenant-scope escapes, migrate-only leaks into runtime,
// missing providers/cycles). Both rule checkers are pure functions so they
// are unit tested without a live `go list` / real kernel boot.
package cli

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/qatoolist/wowapi/internal/buildinfo"
	"github.com/qatoolist/wowapi/kernel/lifecycle"
)

func lintUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi lint <subcommand>

Subcommands:
  boundaries   check module isolation + layering (exit 1 on any violation)
  lifecycle    check the static provider/lifecycle manifest (exit 1 on any violation)
`)
}

func runLint(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		lintUsage(stderr)
		return 2
	}
	switch args[0] {
	case "boundaries":
		return runLintBoundaries(args[1:], stdout, stderr)
	case "lifecycle":
		return runLintLifecycle(args[1:], stdout, stderr)
	case "-h", "--help", "help":
		lintUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "wowapi lint: unknown subcommand %q\n", args[0])
		lintUsage(stderr)
		return 2
	}
}

// runLintLifecycle implements `wowapi lint lifecycle` (backlog B9): prints the
// static provider/lifecycle manifest (kernel/lifecycle.CurrentManifest, built
// by hand from kernel.New/app.Boot/module.Context) and lints it for the
// wiring-mistake classes the framework-competitive-architecture-benchmark's
// "DI / IoC: Static Lifecycle Graph For Go" section calls out: scope leaks,
// raw pools reaching modules, tenant-scoped values escaping their
// transaction, migrate-only services reaching API/worker runtime, and
// missing providers/cycles. Exit 0: manifest printed, no violations. Exit 1:
// manifest printed, violations printed to stderr — CI gates on this.
func runLintLifecycle(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("wowapi lint lifecycle", flag.ContinueOnError)
	fs.SetOutput(stderr)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	return lintLifecycleManifest(lifecycle.CurrentManifest(), stdout, stderr)
}

// lintLifecycleManifest prints m and lints it, returning the process exit code
// (0 clean, 1 on any violation). Split out from runLintLifecycle so the
// exit-code contract can be tested against a deliberately-broken manifest
// without needing CurrentManifest (which is always clean) to be dirty.
func lintLifecycleManifest(m lifecycle.Manifest, stdout, stderr io.Writer) int {
	fmt.Fprint(stdout, m.Print())

	violations := lifecycle.Lint(m)
	if len(violations) > 0 {
		fmt.Fprintln(stderr, "LIFECYCLE VIOLATIONS:")
		for _, v := range violations {
			fmt.Fprintf(stderr, "  %s\n", v.String())
		}
		return 1
	}
	fmt.Fprintln(stdout, "lifecycle lint: OK")
	return 0
}

func runLintBoundaries(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("wowapi lint boundaries", flag.ContinueOnError)
	fs.SetOutput(stderr)
	pkgs := fs.String("pkgs", "./...", "package pattern to lint")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(stderr, "wowapi lint boundaries: %v\n", err)
		return 1
	}
	gm, ok := buildinfo.FindGoMod(wd)
	if !ok {
		fmt.Fprintln(stderr, "wowapi lint boundaries: no go.mod found — run inside a Go module")
		return 1
	}
	imports, err := listImports(*pkgs)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi lint boundaries: %v\n", err)
		return 1
	}
	violations := checkBoundaries(imports, gm.ModulePath, gm.IsFramework())
	if len(violations) > 0 {
		fmt.Fprintln(stderr, "BOUNDARY VIOLATIONS:")
		for _, v := range violations {
			fmt.Fprintf(stderr, "  %s\n", v)
		}
		return 1
	}
	fmt.Fprintln(stdout, "boundary lint: OK")
	return 0
}

// listImports runs `go list` and returns package → production import paths.
func listImports(pattern string) (map[string][]string, error) {
	// Deliberate subprocess: boundary lint shells out to the local go toolchain
	// by design (W01-E01 gosec triage). Context-bound so it is cancellable.
	cmd := exec.CommandContext(context.Background(), "go", "list", "-f", "{{.ImportPath}}: {{join .Imports \" \"}}", pattern) // #nosec G204 -- fixed `go list` argv; only the package pattern varies, supplied by the CLI caller
	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("go list: %w\n%s", err, strings.TrimSpace(errBuf.String()))
	}
	res := map[string][]string{}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		pkg, imps, ok := strings.Cut(line, ": ")
		if !ok {
			continue
		}
		res[pkg] = strings.Fields(imps)
	}
	return res, nil
}

// checkBoundaries is the pure rule engine: given each package's imports, the
// module path, and whether this is the framework repo, it returns the list of
// human-readable violations (empty = clean).
func checkBoundaries(imports map[string][]string, modulePath string, isFramework bool) []string {
	var out []string
	modulesPrefix := modulePath + "/internal/modules/"

	// Layer rules apply to the framework repo (prod imports).
	type rule struct{ pkg, forbidden, reason string }
	var layerRules []rule
	if isFramework {
		// Mirrors scripts/lint_boundaries.sh production rules (the shell script
		// remains the authoritative framework gate for its vocabulary/Reveal/
		// test-import checks; this covers the import-layering law).
		add := func(pkg string, forbidden ...string) {
			for _, f := range forbidden {
				layerRules = append(layerRules, rule{pkg, f, pkg + " must not import " + f})
			}
		}
		// testkit is handled by the HARD rule below (avoids double-reporting).
		add("kernel", "module", "app", "adapters", "examples", "internal/testmodules")
		add("module", "app", "adapters", "examples", "internal/testmodules")
		add("adapters", "module", "app", "examples", "internal/testmodules")
		add("app", "examples", "internal/testmodules")
		add("cmd", "examples", "internal/testmodules")
		add("internal/cli", "module", "examples", "internal/testmodules")
		add("internal/tools", "module", "app", "adapters", "examples", "internal/testmodules")
	}

	for pkg, imps := range imports {
		selfMod := moduleName(pkg, modulesPrefix)
		for _, imp := range imps {
			// Module isolation: a module package importing a DIFFERENT module.
			if selfMod != "" {
				if other := moduleName(imp, modulesPrefix); other != "" && other != selfMod {
					out = append(out, fmt.Sprintf("module %q imports module %q (%s imports %s) — modules must collaborate via ports", selfMod, other, pkg, imp))
				}
			}
			// Framework layer rules.
			for _, r := range layerRules {
				if hasLayer(pkg, modulePath, r.pkg) && hasLayer(imp, modulePath, r.forbidden) {
					out = append(out, fmt.Sprintf("%s (%s imports %s)", r.reason, pkg, imp))
				}
			}
			// HARD rule: no production package (except testkit itself) imports
			// testkit — test files may, but this checks production imports.
			if isFramework && hasLayer(imp, modulePath, "testkit") && !hasLayer(pkg, modulePath, "testkit") {
				out = append(out, fmt.Sprintf("production code imports testkit (%s imports %s)", pkg, imp))
			}
		}
	}
	sort.Strings(out)
	return out
}

// moduleName returns the module directory name if pkg lives under
// <module>/internal/modules/<name>[/...], else "".
func moduleName(pkg, modulesPrefix string) string {
	if !strings.HasPrefix(pkg, modulesPrefix) {
		return ""
	}
	rest := strings.TrimPrefix(pkg, modulesPrefix)
	name, _, _ := strings.Cut(rest, "/")
	return name
}

// hasLayer reports whether pkg is the framework layer `layer` (path-segment
// aware: "kernel" matches kernel and kernel/config, never a "kernelx" sibling).
func hasLayer(pkg, modulePath, layer string) bool {
	base := modulePath + "/" + layer
	return pkg == base || strings.HasPrefix(pkg, base+"/")
}
