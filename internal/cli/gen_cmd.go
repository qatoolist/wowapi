// gen_cmd.go — wowapi gen: code generation subcommands (crud, ...).
package cli

import (
	"bytes"
	"flag"
	"fmt"
	goformat "go/format"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/qatoolist/wowapi/internal/buildinfo"
)

func genUsage(w io.Writer) {
	fmt.Fprint(w, `usage: wowapi gen <subcommand> [flags]

Subcommands:
  crud           Generate CRUD scaffolding for a named resource
  rule           Generate a rule-point declaration
  workflow       Generate a workflow definition
  event-handler  Generate an outbox event handler
  recurring-job  Generate a leader-safe recurring job
  document-flow  Generate a document-class flow
  notification   Generate a notification template declaration
  webhook        Generate an inbound webhook handler
All subcommands accept --module. Non-CRUD subcommands also accept --name and
--force.

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
	FrameworkModule string
	Package         string // Go package name (base of modDir)
	ModuleName      string // same as Package
	Resource        string // resource name, e.g. "widget"
	ResourceTitle   string // title-cased resource, e.g. "Widget"
	Table           string // SQL table name, e.g. "widgets_widget"
	Fields          []fieldDef
	PermPrefix      string // permission key prefix, e.g. "widgets.widget"
}

type subsystemData struct {
	FrameworkModule string
	Package         string
	ModuleName      string
	Kind            string
	Name            string
	Title           string
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
	case "rule", "workflow", "event-handler", "recurring-job", "document-flow", "notification", "webhook":
		return runGenSubsystem(args[0], args[1:], stdout, stderr)
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
	if err := ensureGeneratedRegistrationSupport(*modDir); err != nil {
		fmt.Fprintf(stderr, "wowapi gen crud: upgrade module registration: %v\n", err)
		return 1
	}
	if err := wireGeneratedModule(*modDir, moduleName); err != nil {
		fmt.Fprintf(stderr, "wowapi gen crud: wire module: %v\n", err)
		return 1
	}
	data := crudData{
		FrameworkModule: buildinfo.ModulePath,
		Package:         moduleName,
		ModuleName:      moduleName,
		Resource:        *resource,
		ResourceTitle:   toCamel(*resource),
		Table:           moduleName + "_" + *resource,
		Fields:          fieldDefs,
		PermPrefix:      moduleName + "." + *resource,
	}

	migrationDir := filepath.Join(*modDir, "migrations")
	var migName string
	if *force {
		matches, err := filepath.Glob(filepath.Join(migrationDir, "*_"+*resource+".sql"))
		if err != nil {
			fmt.Fprintf(stderr, "wowapi gen crud: find existing migration: %v\n", err)
			return 1
		}
		if len(matches) > 0 {
			migName = filepath.Base(matches[len(matches)-1])
		}
	}
	if migName == "" {
		migNum, err := nextMigrationNumber(migrationDir)
		if err != nil {
			fmt.Fprintf(stderr, "wowapi gen crud: %v\n", err)
			return 1
		}
		migName = fmt.Sprintf("%05d_%s.sql", migNum, *resource)
	}

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

	if err := addCRUDPermissions(filepath.Join(*modDir, "seeds", "permissions.yaml"), data); err != nil {
		fmt.Fprintf(stderr, "wowapi gen crud: seed permissions: %v\n", err)
		return 1
	}

	return 0
}

func ensureGeneratedRegistrationSupport(modDir string) error {
	modulePath := filepath.Join(modDir, "module.go")
	src, err := os.ReadFile(modulePath) // #nosec G304 -- generator intentionally reads the selected module file
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if bytes.Contains(src, []byte("generatedRegistrations")) {
		return nil
	}
	const moduleMarker = "// Module implements"
	const openAPILine = "\tmc.OpenAPI(openapiFragment)\n"
	if !bytes.Contains(src, []byte(moduleMarker)) || !bytes.Contains(src, []byte(openAPILine)) {
		return fmt.Errorf("%s is missing supported wowapi module markers", modulePath)
	}
	declarations := "type generatedRegistration func(module.Context) error\n\n" +
		"var generatedRegistrations []generatedRegistration\n\n" +
		"func registerGenerated(fn generatedRegistration) {\n" +
		"\tgeneratedRegistrations = append(generatedRegistrations, fn)\n" +
		"}\n\n"
	updated := bytes.Replace(src, []byte(moduleMarker), []byte(declarations+moduleMarker), 1)
	loop := openAPILine +
		"\tfor _, register := range generatedRegistrations {\n" +
		"\t\tif err := register(mc); err != nil {\n" +
		"\t\t\treturn err\n" +
		"\t\t}\n" +
		"\t}\n"
	updated = bytes.Replace(updated, []byte(openAPILine), []byte(loop), 1)
	formatted, err := goformat.Source(updated)
	if err != nil {
		return fmt.Errorf("format upgraded module: %w", err)
	}
	return os.WriteFile(modulePath, formatted, 0o600) // #nosec G703 -- generator intentionally updates the selected module file
}

func addCRUDPermissions(seedPath string, data crudData) error {
	src, err := os.ReadFile(seedPath) // #nosec G304 -- generator intentionally reads the selected seed file
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var entries strings.Builder
	for _, verb := range []string{"create", "read", "list", "update", "deactivate"} {
		key := data.PermPrefix + "." + verb
		if strings.Contains(string(src), "key: "+key) {
			continue
		}
		fmt.Fprintf(&entries, "  - key: %s\n    description: Generated %s %s permission\n",
			key, data.Resource, verb)
	}
	if entries.Len() == 0 {
		return nil
	}
	updated := string(src)
	if strings.Contains(updated, "permissions: []") {
		updated = strings.Replace(updated, "permissions: []", "permissions:\n"+entries.String(), 1)
	} else {
		const nextSection = "\nresource_types:"
		if !strings.Contains(updated, nextSection) {
			return fmt.Errorf("%s has no resource_types section", seedPath)
		}
		updated = strings.Replace(updated, nextSection, entries.String()+nextSection, 1)
	}
	return os.WriteFile(seedPath, []byte(updated), 0o600) // #nosec G703 -- generator intentionally updates the selected seed file
}

func runGenSubsystem(kind string, args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("wowapi gen "+kind, flag.ContinueOnError)
	fs.SetOutput(stderr)
	modDir := fs.String("module", "", "module directory (required)")
	name := fs.String("name", "", "generated declaration name (required)")
	force := fs.Bool("force", false, "overwrite existing file")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *modDir == "" {
		fmt.Fprintf(stderr, "wowapi gen %s: --module is required\n", kind)
		return 2
	}
	if *name == "" {
		fmt.Fprintf(stderr, "wowapi gen %s: --name is required\n", kind)
		return 2
	}
	if !identRE.MatchString(*name) {
		fmt.Fprintf(stderr, "wowapi gen %s: --name %q must match ^[a-z][a-z0-9_]*$\n", kind, *name)
		return 1
	}
	moduleName := filepath.Base(*modDir)
	if !identRE.MatchString(moduleName) {
		fmt.Fprintf(stderr, "wowapi gen %s: --module last path segment %q is not a valid Go package name ([a-z][a-z0-9_]*)\n", kind, moduleName)
		return 1
	}

	if err := ensureGeneratedRegistrationSupport(*modDir); err != nil {
		fmt.Fprintf(stderr, "wowapi gen %s: upgrade module registration: %v\n", kind, err)
		return 1
	}
	if err := wireGeneratedModule(*modDir, moduleName); err != nil {
		fmt.Fprintf(stderr, "wowapi gen %s: wire module: %v\n", kind, err)
		return 1
	}
	data := subsystemData{
		FrameworkModule: buildinfo.ModulePath,
		Package:         moduleName,
		ModuleName:      moduleName,
		Kind:            kind,
		Name:            *name,
		Title:           toCamel(*name),
	}
	dest := filepath.Join(*modDir, *name+"_"+strings.ReplaceAll(kind, "-", "_")+".go")
	if err := renderToFile(dest, "templates/subsystem/subsystem.go.tmpl", data, *force); err != nil {
		fmt.Fprintf(stderr, "wowapi gen %s: %v\n", kind, err)
		return 1
	}
	fmt.Fprintln(stdout, dest)
	return 0
}

// mapFieldType maps a user-supplied type name to (goType, sqlType).
// UUID and timestamp fields use their semantic Go types; the CRUD template
// already imports uuid and time for its framework-owned fields.
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
		return "uuid.UUID", "uuid", true
	case "time", "timestamp", "timestamptz":
		return "time.Time", "timestamptz", true
	default:
		return "", "", false
	}
}
