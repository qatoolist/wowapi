// gen_cmd.go — wowapi gen: code generation subcommands (crud, ...).
package cli

import (
	"flag"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

func genUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi gen <subcommand> [flags]

Subcommands:
  crud   Generate CRUD scaffolding for a named resource inside a module directory

wowapi gen crud flags:
  --module     Module directory (required, e.g. "internal/modules/widgets")
  --resource   Resource name (required; must match ^[a-z][a-z0-9_]*$, e.g. "widget")
  --fields     Comma-separated field definitions, e.g. "title:string,count:int"
  --force      Overwrite existing files
`)
}

// fieldDef holds a single user-declared field for a CRUD resource.
type fieldDef struct {
	Name    string // original snake_case name
	GoName  string // CamelCase Go field name
	GoType  string // Go type (built-in or stdlib)
	SQLType string // SQL column type
}

// crudData is the template data passed to crud templates.
type crudData struct {
	Package       string // Go package name (base of modDir)
	ModuleName    string // same as Package
	Resource      string // resource name, e.g. "widget"
	ResourceTitle string // title-cased resource, e.g. "Widget"
	Table         string // SQL table name, e.g. "widgets_widget"
	Fields        []fieldDef
	PermPrefix    string // permission key prefix, e.g. "widgets.widget"
}

// runGen dispatches `wowapi gen <subcommand>`.
func runGen(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		genUsage(stderr)
		return 2
	}
	switch args[0] {
	case "crud":
		return runGenCRUD(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "wowapi gen: unknown subcommand %q\n", args[0])
		genUsage(stderr)
		return 2
	}
}

// runGenCRUD implements `wowapi gen crud`.
func runGenCRUD(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("wowapi gen crud", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var (
		modDir   = fs.String("module", "", "module directory (required)")
		resource = fs.String("resource", "", "resource name (required)")
		fields   = fs.String("fields", "", `comma-separated field list (e.g. "title:string,count:int")`)
		force    = fs.Bool("force", false, "overwrite existing files")
	)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *modDir == "" {
		fmt.Fprintln(stderr, "wowapi gen crud: --module is required")
		genUsage(stderr)
		return 2
	}
	if *resource == "" {
		fmt.Fprintln(stderr, "wowapi gen crud: --resource is required")
		genUsage(stderr)
		return 2
	}
	if !identRE.MatchString(*resource) {
		fmt.Fprintf(stderr, "wowapi gen crud: --resource %q must match ^[a-z][a-z0-9_]*$\n", *resource)
		return 1
	}

	var fieldDefs []fieldDef
	if *fields != "" {
		for _, part := range strings.Split(*fields, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			kv := strings.SplitN(part, ":", 2)
			if len(kv) != 2 {
				fmt.Fprintf(stderr, "wowapi gen crud: invalid field spec %q (expected name:type)\n", part)
				return 1
			}
			fname, ftype := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
			if !identRE.MatchString(fname) {
				fmt.Fprintf(stderr, "wowapi gen crud: field name %q must match ^[a-z][a-z0-9_]*$\n", fname)
				return 1
			}
			goType, sqlType, ok := mapFieldType(ftype)
			if !ok {
				fmt.Fprintf(stderr, "wowapi gen crud: unknown field type %q for %q — supported: string, int, int64, bool, float64, uuid, time\n", ftype, fname)
				return 1
			}
			fieldDefs = append(fieldDefs, fieldDef{
				Name:    fname,
				GoName:  toCamel(fname),
				GoType:  goType,
				SQLType: sqlType,
			})
		}
	}

	moduleName := filepath.Base(*modDir)
	if !identRE.MatchString(moduleName) {
		fmt.Fprintf(stderr, "wowapi gen crud: --module last path segment %q is not a valid Go package name ([a-z][a-z0-9_]*)\n", moduleName)
		return 1
	}
	data := crudData{
		Package:       moduleName,
		ModuleName:    moduleName,
		Resource:      *resource,
		ResourceTitle: toCamel(*resource),
		Table:         moduleName + "_" + *resource,
		Fields:        fieldDefs,
		PermPrefix:    moduleName + "." + *resource,
	}

	migNum, err := nextMigrationNumber(filepath.Join(*modDir, "migrations"))
	if err != nil {
		fmt.Fprintf(stderr, "wowapi gen crud: %v\n", err)
		return 1
	}
	migName := fmt.Sprintf("%05d_%s.sql", migNum, *resource)

	type fileSpec struct {
		dest string
		tmpl string
	}
	files := []fileSpec{
		{filepath.Join(*modDir, *resource+".go"), "templates/crud/resource.go.tmpl"},
		{filepath.Join(*modDir, "migrations", migName), "templates/crud/migration.sql.tmpl"},
	}

	for _, spec := range files {
		if err := renderToFile(spec.dest, spec.tmpl, data, *force); err != nil {
			fmt.Fprintf(stderr, "wowapi gen crud: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, spec.dest)
	}

	return 0
}

// mapFieldType maps a user-supplied type name to (goType, sqlType).
// All returned goType values are built-in Go types to keep generated files
// import-free; consumers can refine to uuid.UUID / time.Time as needed.
// ok is false for an unrecognized type — the caller rejects it rather than
// emitting an undefined Go type that only fails at `go build` (CLI-01).
func mapFieldType(typ string) (goType, sqlType string, ok bool) {
	switch strings.ToLower(typ) {
	case "string", "text":
		return "string", "text", true
	case "int", "integer":
		return "int", "int", true
	case "int64", "bigint":
		return "int64", "bigint", true
	case "bool", "boolean":
		return "bool", "boolean", true
	case "float64", "float", "double":
		return "float64", "double precision", true
	case "uuid":
		return "string", "uuid", true // TODO: replace with uuid.UUID + import as needed
	case "time", "timestamp", "timestamptz":
		return "string", "timestamptz", true // TODO: replace with time.Time + import as needed
	default:
		return "", "", false
	}
}
