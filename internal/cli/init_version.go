// init_version.go — wowapi init: pre-write framework-version resolution (DX-01,
// W01-E04-S001-T001).
//
// `init` must never write a go.mod whose framework requirement cannot resolve.
// Exactly one of three mutually exclusive paths decides the framework line,
// and every path either yields a verified value or fails closed BEFORE any
// file is written:
//
//  1. --framework-version vX.Y.Z — verified against the module proxy/VCS via
//     `go list -m` before any write.
//  2. --local-framework /abs/path — an explicit replace directive to a local
//     wowapi checkout, path-validated before any write (dev mode, warned).
//  3. neither flag — a released CLI uses its own stamped release version; a
//     source (devel) build derives the canonical version for the exact VCS
//     revision stamped into the binary, and fails closed when the build has
//     no VCS stamp, was built from a dirty tree, or the revision does not
//     resolve (e.g. never pushed).
//
// The old unconditional `devel` → `v0.0.0` fallback is deleted; no code path
// can silently write an unresolvable placeholder again.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/qatoolist/wowapi/internal/buildinfo"
)

// frameworkResolution is the pre-write outcome: the go.mod require-line
// version, plus the optional local-checkout replace target (dev mode).
type frameworkResolution struct {
	Version        string // require-line version; resolvable, or inert under LocalFramework
	LocalFramework string // non-empty => emit `replace FrameworkModule => LocalFramework`
	Warning        string // non-empty => visible dev-mode warning for stderr
}

// localReplaceVersion is the require-line value used only together with a
// `replace` directive to a local checkout: with a directory replacement the
// go tool never resolves the require version, and this canonical zero
// pseudo-version is Go's own convention for locally replaced modules. It is
// not the deleted bare-`v0.0.0` fallback — it is never written without the
// replace directive that satisfies it.
const localReplaceVersion = "v0.0.0-00010101000000-000000000000"

// Test seams: a test binary carries no VCS stamp, and hermetic tests must not
// depend on the real module proxy, so both effects are injectable.
var (
	initBuildVersion     = buildinfo.Version
	initVCSInfo          = readBuildVCSInfo
	resolveModuleVersion = goResolveModuleVersion
)

// pseudoVersionRE matches the go tool's VCS-derived pseudo-version shape
// (…yyyymmddhhmmss-abcdefabcdef), which Go 1.24+ `go build` stamps into a
// locally built binary's main-module version — optionally with a "+dirty"
// build-metadata suffix when the working tree had uncommitted changes.
var pseudoVersionRE = regexp.MustCompile(`[-.]\d{14}-[0-9a-f]{12}(\+dirty)?$`)

// readBuildVCSInfo reports the VCS revision stamped into this binary at build
// time (`go build` with the default -buildvcs=auto) and whether the working
// tree was dirty when it was built.
func readBuildVCSInfo() (revision string, modified bool, ok bool) {
	bi, biOK := debug.ReadBuildInfo()
	if !biOK {
		return "", false, false
	}
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value == "true"
		}
	}
	return revision, modified, revision != ""
}

// goResolveModuleVersion resolves module@query (query: a semver version or a
// VCS revision) to its canonical module version via `go list -m -json`. It
// runs inside a throwaway one-off module so that the target directory (which
// has no go.mod yet) and any enclosing module's replace directives cannot
// influence the answer.
func goResolveModuleVersion(module, query string) (string, error) {
	tmp, err := os.MkdirTemp("", "wowapi-init-resolve-*")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmp) }() // best-effort temp cleanup
	stub := "module wowapi.invalid/resolvecheck\n\ngo 1.26\n"
	if err := os.WriteFile(filepath.Join(tmp, "go.mod"), []byte(stub), 0o600); err != nil {
		return "", err
	}
	// Deliberate subprocess: version resolution IS a `go list -m` call (W01-E01
	// gosec triage). Context-bound so it is cancellable.
	cmd := exec.CommandContext(context.Background(), "go", "list", "-m", "-json", module+"@"+query) // #nosec G204 -- fixed `go list -m -json` argv; module@query is the CLI caller's own requested version
	cmd.Dir = tmp
	cmd.Env = append(os.Environ(), "GOWORK=off")
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s", msg)
	}
	var m struct {
		Version string `json:"Version"`
	}
	if err := json.Unmarshal(out, &m); err != nil || m.Version == "" {
		return "", fmt.Errorf("could not parse `go list -m -json` output for %s@%s", module, query)
	}
	return m.Version, nil
}

// resolveFrameworkVersion picks exactly one of the three resolution paths and
// verifies it. Every returned error carries an exact, copy-pasteable
// remediation; runInit prints it and exits non-zero before any file write.
func resolveFrameworkVersion(explicitVersion, localFramework string) (frameworkResolution, error) {
	fw := buildinfo.ModulePath

	if explicitVersion != "" && localFramework != "" {
		return frameworkResolution{}, fmt.Errorf(
			"--framework-version and --local-framework are mutually exclusive\nremediation: pass exactly one of them")
	}

	if localFramework != "" {
		if !filepath.IsAbs(localFramework) {
			return frameworkResolution{}, fmt.Errorf(
				"--local-framework %q is not an absolute path\nremediation: pass the absolute path of your wowapi checkout, e.g.:\n  wowapi init ... --local-framework \"$(cd path/to/wowapi && pwd)\"",
				localFramework)
		}
		path := filepath.Clean(localFramework)
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			return frameworkResolution{}, fmt.Errorf(
				"--local-framework %q is not an existing directory\nremediation: pass the absolute path of your wowapi checkout, e.g.:\n  wowapi init ... --local-framework \"$(cd path/to/wowapi && pwd)\"",
				localFramework)
		}
		g, ok := buildinfo.FindGoMod(path)
		if !ok || filepath.Clean(g.Dir) != path || !g.IsFramework() {
			return frameworkResolution{}, fmt.Errorf(
				"--local-framework %q does not contain a go.mod declaring module %s\nremediation: point --local-framework at the root of a wowapi checkout (the directory holding its go.mod)",
				localFramework, fw)
		}
		return frameworkResolution{
			Version:        localReplaceVersion,
			LocalFramework: path,
			Warning: fmt.Sprintf(
				"wowapi init: dev mode — go.mod resolves %s via a local replace directive (=> %s); remove the replace and pin a released version before publishing this module",
				fw, path),
		}, nil
	}

	if explicitVersion != "" {
		v, err := resolveModuleVersion(fw, explicitVersion)
		if err != nil {
			return frameworkResolution{}, fmt.Errorf(
				"--framework-version %q does not resolve for %s:\n  %w\nremediation: list the available versions with:\n  go list -m -versions %s",
				explicitVersion, fw, err, fw)
		}
		return frameworkResolution{Version: v}, nil
	}

	// No flags: classify the version stamped into this binary by shape.
	//
	//   vX.Y.Z (tagged release, `go install …@vX.Y.Z`) — resolvable by
	//     construction; used as-is.
	//   …+dirty (Go 1.24+ `go build` from a dirty tree — the SF-7 defect
	//     shape) — names no exact commit state; fail closed.
	//   pseudo-version (clean `go build` from a checkout) — exact, but only
	//     resolvable if the commit is reachable; verified before any write.
	//   "devel" (unstamped build, e.g. `go run` or -buildvcs=off) — derive
	//     from the vcs.revision build setting, verified; else fail closed.
	const remediation = "remediation: pass an explicit resolvable version or a local checkout:\n" +
		"  wowapi init ... --framework-version vX.Y.Z   (discover versions: go list -m -versions %s)\n" +
		"  wowapi init ... --local-framework /absolute/path/to/wowapi"
	stamped := initBuildVersion()
	switch {
	case strings.HasSuffix(stamped, "+dirty"):
		return frameworkResolution{}, fmt.Errorf(
			"this CLI was built from a dirty working tree (stamped version %s), so no exact framework version can be derived\n"+remediation, stamped, fw)

	case stamped != "devel" && pseudoVersionRE.MatchString(stamped):
		v, err := resolveModuleVersion(fw, stamped)
		if err != nil {
			return frameworkResolution{}, fmt.Errorf(
				"this CLI is a source build stamped %s, which does not resolve for %s (commit not pushed?):\n  %v\n"+remediation, stamped, fw, err, fw)
		}
		return frameworkResolution{Version: v}, nil

	case stamped != "devel":
		// Tagged release version.
		return frameworkResolution{Version: stamped}, nil
	}

	// Unstamped source (devel) build: derive the canonical version of the
	// exact commit this binary was built from, and verify it resolves. `go
	// list -m <module>@<revision>` both derives the canonical pseudo-version
	// (or tag) and proves the revision is reachable by the go tool — an
	// unpushed or unknown commit fails here, closed, before any write.
	rev, modified, ok := initVCSInfo()
	if !ok {
		return frameworkResolution{}, fmt.Errorf(
			"this is a source (devel) build of the CLI without VCS build metadata, so no framework version can be derived\n"+remediation, fw)
	}
	if modified {
		return frameworkResolution{}, fmt.Errorf(
			"this is a source (devel) build of the CLI from a dirty working tree (commit %.12s + uncommitted changes), so no exact framework version can be derived\n"+remediation, rev, fw)
	}
	v, err := resolveModuleVersion(fw, rev)
	if err != nil {
		return frameworkResolution{}, fmt.Errorf(
			"this is a source (devel) build of the CLI at commit %.12s, which does not resolve for %s (not pushed?):\n  %v\n"+remediation, rev, fw, err, fw)
	}
	return frameworkResolution{Version: v}, nil
}
