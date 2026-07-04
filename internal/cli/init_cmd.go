// init_cmd.go — wowapi init: scaffold a product repository.
package cli

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/qatoolist/wowapi/internal/buildinfo"
)

func initUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi init [flags]

Scaffold a minimal product repository that depends on the wowapi framework.
Refuses if the target directory is non-empty unless --force is set.

Flags:
  --module   Go module path for the product (required, e.g. "github.com/acme/app")
  --name     Product name (default: last path segment of --module)
  --dir      Target directory (default ".")
  --force    Overwrite existing files and scaffold into non-empty directories
`)
}

// initData is the template data passed to every init template.
type initData struct {
	Module           string
	Name             string
	DBName           string // snake_case version of Name for SQL identifiers
	FrameworkModule  string
	FrameworkVersion string
}

// runInit implements `wowapi init`.
func runInit(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("wowapi init", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		dir    = fs.String("dir", ".", "target directory")
		module = fs.String("module", "", "Go module path (required)")
		name   = fs.String("name", "", "product name (default: last segment of --module)")
		force  = fs.Bool("force", false, "overwrite existing files")
	)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *module == "" {
		fmt.Fprintln(stderr, "wowapi init: --module is required")
		initUsage(stderr)
		return 2
	}

	productName := *name
	if productName == "" {
		parts := strings.Split(*module, "/")
		productName = parts[len(parts)-1]
	}

	target, err := filepath.Abs(*dir)
	if err != nil {
		fmt.Fprintf(stderr, "wowapi init: %v\n", err)
		return 1
	}

	// Refuse if non-empty and --force is not set.
	if !*force {
		entries, err := os.ReadDir(target)
		if err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(stderr, "wowapi init: %v\n", err)
			return 1
		}
		if len(entries) > 0 {
			fmt.Fprintf(stderr, "wowapi init: %s is not empty — use --force to scaffold into a non-empty directory\n", target)
			return 1
		}
	}

	fwVer := buildinfo.Version()
	if fwVer == "devel" {
		fwVer = "v0.0.0"
	}

	data := initData{
		Module:           *module,
		Name:             productName,
		DBName:           strings.ReplaceAll(productName, "-", "_"),
		FrameworkModule:  buildinfo.ModulePath,
		FrameworkVersion: fwVer,
	}

	type fileSpec struct {
		dest string
		tmpl string
	}
	files := []fileSpec{
		{"go.mod", "templates/init/go.mod.tmpl"},
		{".gitignore", "templates/init/gitignore.tmpl"},
		{"Makefile", "templates/init/Makefile.tmpl"},
		{"README.md", "templates/init/README.md.tmpl"},
		{"cmd/api/main.go", "templates/init/cmd_api_main.go.tmpl"},
		{"cmd/worker/main.go", "templates/init/cmd_worker_main.go.tmpl"},
		{"cmd/migrate/main.go", "templates/init/cmd_migrate_main.go.tmpl"},
		{"configs/base.yaml", "templates/init/configs_base.yaml.tmpl"},
		{"configs/local.yaml", "templates/init/configs_local.yaml.tmpl"},
		{"internal/wire/modules.go", "templates/init/internal_wire_modules.go.tmpl"},
		{"internal/appcfg/config.go", "templates/init/internal_appcfg_config.go.tmpl"},
		{"tools/configcheck/main.go", "templates/init/tools_configcheck_main.go.tmpl"},
	}

	for _, spec := range files {
		dest := filepath.Join(target, filepath.FromSlash(spec.dest))
		if err := renderToFile(dest, spec.tmpl, data, *force); err != nil {
			fmt.Fprintf(stderr, "wowapi init: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, spec.dest)
	}

	// internal/modules/ — empty dir with .gitkeep.
	keepPath := filepath.Join(target, "internal", "modules", ".gitkeep")
	if err := writeEmpty(keepPath, *force); err != nil {
		fmt.Fprintf(stderr, "wowapi init: %v\n", err)
		return 1
	}
	fmt.Fprintln(stdout, "internal/modules/.gitkeep")

	return 0
}
