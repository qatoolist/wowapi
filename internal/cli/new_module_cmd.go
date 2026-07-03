// new_module_cmd.go — wowapi new-module: scaffold a module package.
package cli

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
)

func newModuleUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi new-module [flags]

Scaffold a module package under <dir>/<name>/ that implements module.Module.
The generated module embeds migrations/*.sql, seeds/*.yaml, and openapi.json.

Flags:
  --name   Module name (required; must match ^[a-z][a-z0-9_]*$)
  --dir    Parent directory (default "internal/modules")
  --force  Overwrite existing files
`)
}

// newModuleData is the template data for the module scaffold templates.
type newModuleData struct {
	Name    string // module identifier, e.g. "widgets"
	Package string // Go package name — same as Name
}

// runNewModule implements `wowapi new-module`.
func runNewModule(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("wowapi new-module", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		name  = fs.String("name", "", "module name (required)")
		dir   = fs.String("dir", "internal/modules", "parent directory")
		force = fs.Bool("force", false, "overwrite existing files")
	)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *name == "" {
		fmt.Fprintln(stderr, "wowapi new-module: --name is required")
		newModuleUsage(stderr)
		return 2
	}
	if !identRE.MatchString(*name) {
		fmt.Fprintf(stderr, "wowapi new-module: --name %q must match ^[a-z][a-z0-9_]*$\n", *name)
		return 1
	}

	modDir := filepath.Join(*dir, *name)
	data := newModuleData{Name: *name, Package: *name}

	type fileSpec struct {
		dest string
		tmpl string
	}
	files := []fileSpec{
		{filepath.Join(modDir, "module.go"), "templates/module/module.go.tmpl"},
		{filepath.Join(modDir, "openapi.json"), "templates/module/openapi.json.tmpl"},
		{filepath.Join(modDir, "migrations", "00001_init.sql"), "templates/module/migration.sql.tmpl"},
		{filepath.Join(modDir, "seeds", "permissions.yaml"), "templates/module/seeds.yaml.tmpl"},
	}

	for _, spec := range files {
		if err := renderToFile(spec.dest, spec.tmpl, data, *force); err != nil {
			fmt.Fprintf(stderr, "wowapi new-module: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, spec.dest)
	}

	return 0
}
