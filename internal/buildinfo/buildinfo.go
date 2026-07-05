// Package buildinfo reports the CLI/framework version and inspects a
// consuming repo's go.mod for the wowapi requirement (version-mismatch
// warning, D-0008). Private: not a consumer-facing contract.
package buildinfo

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

// ModulePath is wowapi's canonical module path.
const ModulePath = "github.com/qatoolist/wowapi"

// version is overridable at build time: -ldflags "-X ...buildinfo.version=v1.2.3"
var version = ""

// Version returns the CLI build version: the ldflags override, else the main
// module version stamped by `go install …@vX.Y.Z`, else "devel".
func Version() string {
	if version != "" {
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "devel"
}

// GoMod describes the nearest enclosing go.mod, if any.
type GoMod struct {
	Dir           string // directory containing go.mod
	ModulePath    string // the module line
	WowapiVersion string // required wowapi version; "" if not required
}

// IsFramework reports whether the go.mod belongs to wowapi itself.
func (g GoMod) IsFramework() bool { return g.ModulePath == ModulePath }

// FindGoMod walks upward from dir looking for a go.mod and parses the module
// path plus any wowapi requirement. Returns ok=false when none is found.
func FindGoMod(dir string) (GoMod, bool) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return GoMod{}, false
	}
	for {
		p := filepath.Join(dir, "go.mod")
		if info, err := os.Stat(p); err == nil && !info.IsDir() {
			g, err := parseGoMod(p)
			if err != nil {
				return GoMod{}, false
			}
			g.Dir = dir
			return g, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return GoMod{}, false
		}
		dir = parent
	}
}

func parseGoMod(path string) (GoMod, error) {
	f, err := os.Open(path)
	if err != nil {
		return GoMod{}, err
	}
	defer func() { _ = f.Close() }() // read-only file; a close error loses no data

	var g GoMod
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if rest, ok := strings.CutPrefix(line, "module "); ok && g.ModulePath == "" {
			g.ModulePath = strings.TrimSpace(rest)
			continue
		}
		// Matches both block-form ("\tgithub.com/... v1.2.3") and inline
		// ("require github.com/... v1.2.3") requirements.
		line = strings.TrimPrefix(line, "require ")
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] == ModulePath {
			g.WowapiVersion = fields[1]
		}
	}
	return g, sc.Err()
}
