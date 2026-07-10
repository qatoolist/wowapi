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
	fmt.Fprint(w, `usage: wowapi init [<name>] [flags]

Scaffold a minimal product repository that depends on the wowapi framework.

  wowapi init myapp --module github.com/acme/myapp
      creates ./myapp/ and scaffolds the product inside it.
  wowapi init --module github.com/acme/myapp
      scaffolds into the current directory (--dir ".").

Refuses if the target directory is non-empty unless --force is set.

Arguments:
  <name>     Optional. Creates a new subdirectory <dir>/<name> and scaffolds the
             product inside it; also the default product name.

Flags:
  --module   Go module path for the product (required, e.g. "github.com/acme/app")
  --name     Product name (default: <name>, else the last segment of --module)
  --dir      Base directory (default "."); with <name>, the product goes in <dir>/<name>
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
	// Extract a leading positional <name> BEFORE flag parsing: Go's flag package
	// stops at the first non-flag arg, so `wowapi init myapp --module x` would
	// otherwise leave the flags unparsed. A trailing positional
	// (`wowapi init --module x myapp`) is picked up from fs.Args() after Parse.
	var positional string
	if len(args) > 0 && args[0] != "" && !strings.HasPrefix(args[0], "-") {
		positional, args = args[0], args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if rest := fs.Args(); len(rest) > 0 {
		if positional == "" {
			positional, rest = rest[0], rest[1:]
		}
		if len(rest) > 0 {
			fmt.Fprintf(stderr, "wowapi init: unexpected extra arguments: %s\n", strings.Join(rest, " "))
			initUsage(stderr)
			return 2
		}
	}
	if *module == "" {
		fmt.Fprintln(stderr, "wowapi init: --module is required")
		initUsage(stderr)
		return 2
	}

	// Product name: --name wins; else the positional <name>; else the module's last segment.
	productName := *name
	if productName == "" {
		if positional != "" {
			productName = filepath.Base(positional)
		} else {
			parts := strings.Split(*module, "/")
			productName = parts[len(parts)-1]
		}
	}

	// Target: with a positional <name>, create a NEW subdirectory <dir>/<name> and
	// scaffold inside it; otherwise scaffold directly into --dir (default ".").
	targetRel := *dir
	if positional != "" {
		targetRel = filepath.Join(*dir, positional)
	}
	target, err := filepath.Abs(targetRel)
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

	fmt.Fprintf(stdout, "\nScaffolded product %q into %s\n", productName, target)
	// Only `go mod tidy` is setup-free. build/migrate/run need APP_ENV + the DB DSNs
	// + a running Postgres (fail-closed), so point at the README rather than imply a
	// bare `make migrate-up` works.
	if positional != "" {
		fmt.Fprintf(stdout, "Next: cd %s && go mod tidy — then see README.md \"Getting started\" (set APP_ENV + DB DSNs, start Postgres, migrate, run).\n", positional)
	} else {
		fmt.Fprintln(stdout, "Next: go mod tidy — then see README.md \"Getting started\" (set APP_ENV + DB DSNs, start Postgres, migrate, run).")
	}

	return 0
}
