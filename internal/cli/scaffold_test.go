package cli

import (
	"bytes"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// ---------- test helpers ----------

func callInit(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	code = runInit(args, &out, &errBuf)
	return code, out.String(), errBuf.String()
}

func callNewModule(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	code = runNewModule(args, &out, &errBuf)
	return code, out.String(), errBuf.String()
}

func callGenCRUD(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errBuf bytes.Buffer
	code = runGen(append([]string{"crud"}, args...), &out, &errBuf)
	return code, out.String(), errBuf.String()
}

// assertParseGo checks that the file at path is syntactically valid Go.
func assertParseGo(t *testing.T, path string) {
	t.Helper()
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		content, _ := os.ReadFile(path)
		t.Fatalf("generated Go file %s is not syntactically valid:\n%v\nContent:\n%s", path, err, content)
	}
}

// assertFileContains reads path and fails if it does not contain substr.
func assertFileContains(t *testing.T, path, substr string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	if !strings.Contains(string(content), substr) {
		t.Errorf("%s: expected to contain %q\nActual content:\n%s", path, substr, content)
	}
}

// assertFileMatches fails if the file content does not match the regexp pattern.
func assertFileMatches(t *testing.T, path, pattern string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	if !regexp.MustCompile(pattern).Match(content) {
		t.Errorf("%s: expected to match %q\nActual content:\n%s", path, pattern, content)
	}
}

// assertFileExists fails if path does not exist.
func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s to exist: %v", path, err)
	}
}

// scaffoldModuleDir creates a minimal module directory structure for gen crud tests.
func scaffoldModuleDir(t *testing.T, moduleName string) string {
	t.Helper()
	base := t.TempDir()
	modDir := filepath.Join(base, moduleName)
	if err := os.MkdirAll(filepath.Join(modDir, "migrations"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(modDir, "seeds"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Placeholder migration so nextMigrationNumber returns 2.
	placeholder := filepath.Join(modDir, "migrations", "00001_init.sql")
	if err := os.WriteFile(placeholder, []byte("-- placeholder\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return modDir
}

// ---------- wowapi init ----------

func TestInitCreatesGoMod(t *testing.T) {
	dir := t.TempDir()
	code, out, errOut := callInit(t, "--module", "github.com/acme/testapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	if !strings.Contains(out, "go.mod") {
		t.Errorf("stdout missing go.mod: %q", out)
	}
	gomod := filepath.Join(dir, "go.mod")
	assertFileExists(t, gomod)
	assertFileContains(t, gomod, "module github.com/acme/testapp")
	assertFileContains(t, gomod, "github.com/qatoolist/wowapi")
}

func TestInitCreatesAllFiles(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	expected := []string{
		filepath.Join(dir, "go.mod"),
		filepath.Join(dir, ".gitignore"),
		filepath.Join(dir, "Makefile"),
		filepath.Join(dir, "README.md"),
		filepath.Join(dir, "cmd", "api", "main.go"),
		filepath.Join(dir, "cmd", "worker", "main.go"),
		filepath.Join(dir, "cmd", "migrate", "main.go"),
		filepath.Join(dir, "configs", "base.yaml"),
		filepath.Join(dir, "configs", "local.yaml"),
		filepath.Join(dir, "internal", "modules", ".gitkeep"),
		filepath.Join(dir, "internal", "wire", "modules.go"),
		filepath.Join(dir, "internal", "appcfg", "config.go"), // product config layer (D-0002)
		filepath.Join(dir, "tools", "configcheck", "main.go"), // CLI config-check binary (D-0003)
	}
	for _, f := range expected {
		assertFileExists(t, f)
	}
}

func TestInitGoFilesParseOK(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	for _, f := range []string{
		filepath.Join(dir, "cmd", "api", "main.go"),
		filepath.Join(dir, "cmd", "worker", "main.go"),
		filepath.Join(dir, "cmd", "migrate", "main.go"),
		filepath.Join(dir, "internal", "appcfg", "config.go"),
		filepath.Join(dir, "tools", "configcheck", "main.go"),
	} {
		assertParseGo(t, f)
	}
}

func TestInitNameFromModule(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/coolapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	// local.yaml should contain the product name derived from the module path.
	assertFileContains(t, filepath.Join(dir, "configs", "local.yaml"), "coolapp")
}

func TestInitExplicitName(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--name", "myproduct", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	assertFileContains(t, filepath.Join(dir, "configs", "local.yaml"), "myproduct")
}

func TestInitNonEmptyDirRefused(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errOut := callInit(t, "--module", "github.com/acme/app", "--dir", dir)
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	if !strings.Contains(errOut, "not empty") {
		t.Errorf("stderr should mention 'not empty': %q", errOut)
	}
}

func TestInitForceAllowsNonEmptyDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, errOut := callInit(t, "--module", "github.com/acme/app", "--dir", dir, "--force")
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
}

func TestInitMissingModule(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--dir", dir)
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	if !strings.Contains(errOut, "--module") {
		t.Errorf("stderr should mention --module: %q", errOut)
	}
}

func TestInitGoModHasFrameworkRequire(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/app", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	assertFileContains(t, filepath.Join(dir, "go.mod"), "github.com/qatoolist/wowapi")
}

func TestInitLocalYAMLHasEnvironment(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/app", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	assertFileContains(t, filepath.Join(dir, "configs", "local.yaml"), "environment: local")
}

// TestInitPositionalNameCreatesSubdir: `wowapi init <name> --module ...` creates a
// NEW subdirectory <name> under the base dir and scaffolds the product inside it,
// with the product name derived from the positional arg. Flags after the positional
// must still parse.
func TestInitPositionalNameCreatesSubdir(t *testing.T) {
	base := t.TempDir()
	code, _, errOut := callInit(t, "myapp", "--module", "github.com/acme/myapp", "--dir", base)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	// The product is created inside base/myapp, not directly in base.
	assertFileExists(t, filepath.Join(base, "myapp", "go.mod"))
	assertFileExists(t, filepath.Join(base, "myapp", "cmd", "api", "main.go"))
	assertFileContains(t, filepath.Join(base, "myapp", "configs", "local.yaml"), "myapp")
	if _, err := os.Stat(filepath.Join(base, "go.mod")); !os.IsNotExist(err) {
		t.Fatal("scaffold must go into base/myapp, not base directly")
	}
}

// TestInitPositionalRequiresModule: the positional name adds dir/name ergonomics but
// --module stays required (chosen behaviour).
func TestInitPositionalRequiresModule(t *testing.T) {
	base := t.TempDir()
	code, _, errOut := callInit(t, "myapp", "--dir", base)
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	if !strings.Contains(errOut, "--module") {
		t.Errorf("stderr should mention --module: %q", errOut)
	}
}

// TestInitNameFlagOverridesPositional: --name overrides the product name; the
// positional still controls the target subdirectory.
func TestInitNameFlagOverridesPositional(t *testing.T) {
	base := t.TempDir()
	code, _, errOut := callInit(t, "myapp", "--module", "github.com/acme/myapp", "--name", "custom", "--dir", base)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	assertFileExists(t, filepath.Join(base, "myapp", "go.mod"))
	assertFileContains(t, filepath.Join(base, "myapp", "configs", "local.yaml"), "custom")
}

// TestInitRejectsExtraArgs: more than one positional arg is a usage error.
func TestInitRejectsExtraArgs(t *testing.T) {
	code, _, _ := callInit(t, "a", "b", "--module", "github.com/acme/app")
	if code != 2 {
		t.Fatalf("exit %d, want 2 for extra positional args", code)
	}
}

// ---------- wowapi new-module ----------

func TestNewModuleCreatesFiles(t *testing.T) {
	base := t.TempDir()
	code, out, errOut := callNewModule(t, "--name", "widgets", "--dir", base)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	modDir := filepath.Join(base, "widgets")
	for _, f := range []string{
		filepath.Join(modDir, "module.go"),
		filepath.Join(modDir, "openapi.json"),
		filepath.Join(modDir, "migrations", "00001_init.sql"),
		filepath.Join(modDir, "seeds", "permissions.yaml"),
	} {
		assertFileExists(t, f)
	}
	if !strings.Contains(out, "module.go") {
		t.Errorf("stdout missing module.go: %q", out)
	}
}

func TestNewModuleModuleGoParsable(t *testing.T) {
	base := t.TempDir()
	code, _, errOut := callNewModule(t, "--name", "items", "--dir", base)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	assertParseGo(t, filepath.Join(base, "items", "module.go"))
}

func TestNewModuleNameReturnValue(t *testing.T) {
	base := t.TempDir()
	code, _, errOut := callNewModule(t, "--name", "gadgets", "--dir", base)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	assertFileContains(t, filepath.Join(base, "gadgets", "module.go"), `return "gadgets"`)
}

func TestNewModulePackageClause(t *testing.T) {
	base := t.TempDir()
	code, _, errOut := callNewModule(t, "--name", "orders", "--dir", base)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	assertFileContains(t, filepath.Join(base, "orders", "module.go"), "package orders")
}

func TestNewModuleBadName(t *testing.T) {
	base := t.TempDir()
	code, _, errOut := callNewModule(t, "--name", "BadName", "--dir", base)
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	if !strings.Contains(errOut, "must match") {
		t.Errorf("stderr should mention constraint: %q", errOut)
	}
}

func TestNewModuleMissingName(t *testing.T) {
	code, _, errOut := callNewModule(t, "--dir", t.TempDir())
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	if !strings.Contains(errOut, "--name") {
		t.Errorf("stderr should mention --name: %q", errOut)
	}
}

func TestNewModuleForceOverwrites(t *testing.T) {
	base := t.TempDir()
	// First scaffold.
	code, _, errOut := callNewModule(t, "--name", "parts", "--dir", base)
	if code != 0 {
		t.Fatalf("first scaffold: exit %d; stderr: %s", code, errOut)
	}
	// Second without --force should fail.
	code, _, _ = callNewModule(t, "--name", "parts", "--dir", base)
	if code == 0 {
		t.Fatalf("expected non-zero exit when overwriting without --force")
	}
	// Third with --force should succeed.
	code, _, errOut = callNewModule(t, "--name", "parts", "--dir", base, "--force")
	if code != 0 {
		t.Fatalf("force overwrite: exit %d; stderr: %s", code, errOut)
	}
}

func TestNewModuleOpenAPIHasName(t *testing.T) {
	base := t.TempDir()
	code, _, errOut := callNewModule(t, "--name", "invoices", "--dir", base)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	assertFileContains(t, filepath.Join(base, "invoices", "openapi.json"), "invoices module")
}

// ---------- wowapi gen crud ----------

func TestGenCRUDCreatesFiles(t *testing.T) {
	modDir := scaffoldModuleDir(t, "widgets")
	code, out, errOut := callGenCRUD(t, "--module", modDir, "--resource", "widget")
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	assertFileExists(t, filepath.Join(modDir, "widget.go"))
	// Migration should be 00002_ since 00001_ already exists.
	assertFileExists(t, filepath.Join(modDir, "migrations", "00002_widget.sql"))
	if !strings.Contains(out, "widget.go") {
		t.Errorf("stdout missing widget.go: %q", out)
	}
}

func TestGenCRUDResourceGoParsable(t *testing.T) {
	modDir := scaffoldModuleDir(t, "things")
	code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "thing")
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	assertParseGo(t, filepath.Join(modDir, "thing.go"))
}

func TestGenCRUDWithFields(t *testing.T) {
	modDir := scaffoldModuleDir(t, "products")
	code, _, errOut := callGenCRUD(t,
		"--module", modDir,
		"--resource", "product",
		"--fields", "title:string,count:int,active:bool",
	)
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	goFile := filepath.Join(modDir, "product.go")
	assertParseGo(t, goFile)
	// gofmt column-aligns struct fields, so match name + type across any run of
	// whitespace rather than a single literal space.
	assertFileMatches(t, goFile, `Title\s+string`)
	assertFileMatches(t, goFile, `Count\s+int`)
	assertFileMatches(t, goFile, `Active\s+bool`)
}

func TestGenCRUDPermissionKeys(t *testing.T) {
	modDir := scaffoldModuleDir(t, "widgets")
	code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "widget")
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	goFile := filepath.Join(modDir, "widget.go")
	for _, perm := range []string{
		"widgets.widget.create",
		"widgets.widget.read",
		"widgets.widget.list",
		"widgets.widget.update",
		"widgets.widget.delete",
	} {
		assertFileContains(t, goFile, perm)
	}
}

func TestGenCRUDMigrationRLS(t *testing.T) {
	modDir := scaffoldModuleDir(t, "orders")
	code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "order")
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	migDir := filepath.Join(modDir, "migrations")
	entries, err := os.ReadDir(migDir)
	if err != nil {
		t.Fatal(err)
	}
	var migPath string
	for _, e := range entries {
		if strings.Contains(e.Name(), "_order.sql") {
			migPath = filepath.Join(migDir, e.Name())
		}
	}
	if migPath == "" {
		t.Fatal("migration file not found")
	}
	assertFileContains(t, migPath, "ENABLE ROW LEVEL SECURITY")
	assertFileContains(t, migPath, "FORCE ROW LEVEL SECURITY")
	assertFileContains(t, migPath, "app_tenant_id()")
	assertFileContains(t, migPath, "GRANT SELECT, INSERT, UPDATE ON orders_order TO app_rt")
}

func TestGenCRUDMigrationGooseMarkers(t *testing.T) {
	modDir := scaffoldModuleDir(t, "items")
	code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "item")
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	migDir := filepath.Join(modDir, "migrations")
	entries, _ := os.ReadDir(migDir)
	var migPath string
	for _, e := range entries {
		if strings.Contains(e.Name(), "_item.sql") {
			migPath = filepath.Join(migDir, e.Name())
		}
	}
	if migPath == "" {
		t.Fatal("migration not found")
	}
	assertFileContains(t, migPath, "-- +goose Up")
	assertFileContains(t, migPath, "-- +goose Down")
}

func TestGenCRUDPackageMatchesModule(t *testing.T) {
	modDir := scaffoldModuleDir(t, "alerts")
	code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "alert")
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	assertFileContains(t, filepath.Join(modDir, "alert.go"), "package alerts")
}

func TestGenCRUDBadResource(t *testing.T) {
	modDir := scaffoldModuleDir(t, "things")
	code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "BadRes")
	if code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	if !strings.Contains(errOut, "must match") {
		t.Errorf("stderr missing constraint message: %q", errOut)
	}
}

func TestGenCRUDMissingModule(t *testing.T) {
	code, _, errOut := callGenCRUD(t, "--resource", "widget")
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	if !strings.Contains(errOut, "--module") {
		t.Errorf("stderr missing --module: %q", errOut)
	}
}

func TestGenCRUDMissingResource(t *testing.T) {
	modDir := scaffoldModuleDir(t, "things")
	code, _, errOut := callGenCRUD(t, "--module", modDir)
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	if !strings.Contains(errOut, "--resource") {
		t.Errorf("stderr missing --resource: %q", errOut)
	}
}

func TestGenUnknownSubcommand(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := runGen([]string{"bogus"}, &out, &errBuf)
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
	if !strings.Contains(errBuf.String(), "unknown subcommand") {
		t.Errorf("stderr missing 'unknown subcommand': %q", errBuf.String())
	}
}

func TestGenNoSubcommand(t *testing.T) {
	var out, errBuf bytes.Buffer
	code := runGen([]string{}, &out, &errBuf)
	if code != 2 {
		t.Fatalf("exit %d, want 2", code)
	}
}

func TestGenCRUDMigrationNumbering(t *testing.T) {
	// When no existing migrations, first crud should produce 00001_.
	base := t.TempDir()
	modDir := filepath.Join(base, "fresh")
	if err := os.MkdirAll(filepath.Join(modDir, "migrations"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(modDir, "seeds"), 0o755); err != nil {
		t.Fatal(err)
	}
	code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "thing")
	if code != 0 {
		t.Fatalf("exit %d; stderr: %s", code, errOut)
	}
	assertFileExists(t, filepath.Join(modDir, "migrations", "00001_thing.sql"))
}
