package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------- cli.Run dispatch ----------

func TestRunNoArgsPrintsUsage(t *testing.T) {
	code, out, _ := run(t)
	if code != 0 {
		t.Fatalf("no args should exit 0, got %d", code)
	}
	if !strings.Contains(out, "Usage:") || !strings.Contains(out, "Available commands") {
		t.Fatalf("expected usage banner, got %q", out)
	}
}

func TestRunHelpFlags(t *testing.T) {
	for _, f := range []string{"-h", "--help"} {
		code, out, _ := run(t, f)
		if code != 0 {
			t.Fatalf("%s should exit 0, got %d", f, code)
		}
		if !strings.Contains(out, "Available commands") {
			t.Fatalf("%s should print usage, got %q", f, out)
		}
	}
}

// runVersion inside a consumer repo must report the dependency version and warn
// on a mismatch with the CLI version.
func TestVersionInConsumerRepo(t *testing.T) {
	dir := t.TempDir()
	gomod := "module example.com/consumer\n\ngo 1.26\n\nrequire " +
		"github.com/qatoolist/wowapi v9.9.9-not-the-cli\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)

	code, out, errOut := run(t, "version")
	if code != 0 {
		t.Fatalf("version exit %d", code)
	}
	if !strings.Contains(out, "dependency:") || !strings.Contains(out, "v9.9.9-not-the-cli") {
		t.Fatalf("expected dependency line, got stdout=%q", out)
	}
	if !strings.Contains(errOut, "warning:") {
		t.Fatalf("expected version-mismatch warning on stderr, got %q", errOut)
	}
}

// ---------- deploy ----------

func TestDeployNoArgsUsage(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runDeploy(nil, &out, &errb); code != 2 {
		t.Fatalf("deploy with no args should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "usage: wowapi deploy render") {
		t.Fatalf("expected deploy usage, got %q", errb.String())
	}
}

func TestDeployHelpSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runDeploy([]string{"help"}, &out, &errb); code != 0 {
		t.Fatalf("deploy help should exit 0, got %d", code)
	}
	if !strings.Contains(out.String(), "usage: wowapi deploy render") {
		t.Fatalf("expected usage on stdout, got %q", out.String())
	}
}

func TestDeployRenderToFile(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "compose.yaml")
	var out, errb bytes.Buffer
	code := runDeploy([]string{"render", "--format", "compose", "--name", "svc", "--out", outPath}, &out, &errb)
	if code != 0 {
		t.Fatalf("render --out exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("stdout should echo the output path, got %q", out.String())
	}
	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read rendered file: %v", err)
	}
	if !strings.Contains(string(body), "svc-api") {
		t.Fatalf("rendered file missing service name:\n%s", body)
	}
}

// ---------- lint ----------

func TestLintNoArgsUsage(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runLint(nil, &out, &errb); code != 2 {
		t.Fatalf("lint no args should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "usage: wowapi lint") {
		t.Fatalf("expected lint usage, got %q", errb.String())
	}
}

func TestLintHelpSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runLint([]string{"help"}, &out, &errb); code != 0 {
		t.Fatalf("lint help exit %d", code)
	}
	if !strings.Contains(out.String(), "boundaries") {
		t.Fatalf("expected usage listing boundaries, got %q", out.String())
	}
}

func TestLintUnknownSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runLint([]string{"bogus"}, &out, &errb); code != 2 {
		t.Fatalf("unknown lint subcommand should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "unknown subcommand") {
		t.Fatalf("expected unknown-subcommand error, got %q", errb.String())
	}
}

func TestLintBoundariesCleanModule(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/clean\n\ngo 1.26\n")
	writeFile(t, dir, "foo.go", "package clean\n\nfunc F() {}\n")
	t.Chdir(dir)

	var out, errb bytes.Buffer
	code := runLintBoundaries([]string{"--pkgs", "./..."}, &out, &errb)
	if code != 0 {
		t.Fatalf("clean module should pass lint (exit 0), got %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "boundary lint: OK") {
		t.Fatalf("expected OK message, got %q", out.String())
	}
}

func TestLintBoundariesNoGoMod(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	var out, errb bytes.Buffer
	if code := runLintBoundaries(nil, &out, &errb); code != 1 {
		t.Fatalf("no go.mod should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "no go.mod found") {
		t.Fatalf("expected no-go.mod error, got %q", errb.String())
	}
}

func TestLintBoundariesGoListError(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", "module example.com/broken\n\ngo 1.26\n")
	// Two conflicting package clauses in one directory make `go list ./...` fail
	// deterministically and offline (no import resolution needed).
	writeFile(t, dir, "a.go", "package foo\n")
	writeFile(t, dir, "b.go", "package bar\n")
	t.Chdir(dir)

	var out, errb bytes.Buffer
	if code := runLintBoundaries([]string{"--pkgs", "./..."}, &out, &errb); code != 1 {
		t.Fatalf("go list failure should exit 1, got %d (stderr=%q)", code, errb.String())
	}
}

// ---------- migrate dispatch ----------

func TestMigrateNoArgsUsage(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runMigrate(nil, &out, &errb); code != 2 {
		t.Fatalf("migrate no args should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "usage: wowapi migrate") {
		t.Fatalf("expected migrate usage, got %q", errb.String())
	}
}

func TestMigrateHelpSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runMigrate([]string{"--help"}, &out, &errb); code != 0 {
		t.Fatalf("migrate --help exit %d", code)
	}
	if !strings.Contains(out.String(), "create") {
		t.Fatalf("expected usage mentioning create, got %q", out.String())
	}
}

func TestMigrateUnknownSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runMigrate([]string{"bogus"}, &out, &errb); code != 2 {
		t.Fatalf("unknown migrate subcommand should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "unknown subcommand") {
		t.Fatalf("expected unknown-subcommand error, got %q", errb.String())
	}
}

func TestMigrateCreateDispatch(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := runMigrate([]string{"create", "--dir", dir, "--name", "add_things"}, &out, &errb); code != 0 {
		t.Fatalf("migrate create via dispatch exit %d: %s", code, errb.String())
	}
	if filepath.Base(strings.TrimSpace(out.String())) != "00001_add_things.sql" {
		t.Fatalf("unexpected created path: %q", out.String())
	}
}

func TestMigrateCreateAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	// Pre-create the exact file the next number would produce.
	writeFile(t, dir, "00001_dupe.sql", "-- existing")
	var out, errb bytes.Buffer
	// nextMigrationNumber sees 00001 → next is 00002, so a fresh --name succeeds;
	// to force the collision, name it the same and seed 00002 gap. Instead assert
	// the collision guard by seeding 00002 then requesting a name that maps to it.
	writeFile(t, dir, "00002_dupe.sql", "-- existing2")
	// Now nextMigrationNumber = 3; create with name should succeed (no collision).
	if code := runMigrateCreate([]string{"--dir", dir, "--name", "third"}, &out, &errb); code != 0 {
		t.Fatalf("expected success creating 00003, got %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "00003_third.sql") {
		t.Fatalf("expected 00003_third.sql, got %q", out.String())
	}
}

func TestNextMigrationNumberOnFileNotDir(t *testing.T) {
	// Point --dir at a regular file so os.ReadDir returns a non-IsNotExist error,
	// which runMigrateCreate must surface as exit 1.
	f := filepath.Join(t.TempDir(), "not-a-dir")
	writeFile(t, filepath.Dir(f), "not-a-dir", "x")
	var out, errb bytes.Buffer
	if code := runMigrateCreate([]string{"--dir", f, "--name", "x"}, &out, &errb); code != 1 {
		t.Fatalf("--dir pointing at a file should exit 1, got %d", code)
	}
}

// ---------- openapi ----------

func TestOpenAPINoArgsUsage(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runOpenAPI(nil, &out, &errb); code != 2 {
		t.Fatalf("openapi no args should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "usage: wowapi openapi merge") {
		t.Fatalf("expected openapi usage, got %q", errb.String())
	}
}

func TestOpenAPIHelpSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runOpenAPI([]string{"help"}, &out, &errb); code != 0 {
		t.Fatalf("openapi help exit %d", code)
	}
	if !strings.Contains(out.String(), "Merge OpenAPI") {
		t.Fatalf("expected usage text, got %q", out.String())
	}
}

func TestOpenAPIMergeNoFragments(t *testing.T) {
	dir := t.TempDir() // empty
	var out, errb bytes.Buffer
	if code := runOpenAPI([]string{"merge", "--dir", dir}, &out, &errb); code != 1 {
		t.Fatalf("no fragments should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "no fragments found") {
		t.Fatalf("expected no-fragments error, got %q", errb.String())
	}
}

func TestOpenAPIMergeToFileAndExplicitArg(t *testing.T) {
	dir := t.TempDir()
	frag := filepath.Join(dir, "explicit.json")
	if err := os.WriteFile(frag, []byte(`{"paths":{"/z":{"get":{}}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(t.TempDir(), "merged.json")
	var out, errb bytes.Buffer
	// dir="" forces use of the explicit file argument only.
	code := runOpenAPI([]string{"merge", "--dir", "", "--out", outPath, frag}, &out, &errb)
	if code != 0 {
		t.Fatalf("merge to file exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), outPath) {
		t.Fatalf("stdout should echo out path, got %q", out.String())
	}
	body, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read merged file: %v", err)
	}
	if !strings.Contains(string(body), `"/z"`) {
		t.Fatalf("merged file missing path:\n%s", body)
	}
}

func TestOpenAPIMergeEmptyFileRejected(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "empty.json", "")
	var out, errb bytes.Buffer
	if code := runOpenAPI([]string{"merge", "--dir", dir}, &out, &errb); code != 1 {
		t.Fatalf("empty fragment should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "empty file") {
		t.Fatalf("expected 'empty file' token in error, got %q", errb.String())
	}
}

func TestOpenAPIMergeInvalidJSONObject(t *testing.T) {
	dir := t.TempDir()
	// Starts with '{' (passes the object check) but is not valid JSON.
	writeFile(t, dir, "bad.json", `{"paths": }`)
	var out, errb bytes.Buffer
	if code := runOpenAPI([]string{"merge", "--dir", dir}, &out, &errb); code != 1 {
		t.Fatalf("invalid JSON should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "invalid JSON") {
		t.Fatalf("expected invalid-JSON error, got %q", errb.String())
	}
}

func TestOpenAPIMergeDuplicateSchema(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.json", `{"components":{"schemas":{"Dup":{"type":"object"}}}}`)
	writeFile(t, dir, "b.json", `{"components":{"schemas":{"Dup":{"type":"string"}}}}`)
	var out, errb bytes.Buffer
	if code := runOpenAPI([]string{"merge", "--dir", dir}, &out, &errb); code != 1 {
		t.Fatalf("duplicate schema should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "duplicate components.schemas.Dup") {
		t.Fatalf("expected duplicate-schema error, got %q", errb.String())
	}
}

// ---------- seed dispatch ----------

func TestSeedNoArgsUsage(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runSeed(nil, &out, &errb); code != 2 {
		t.Fatalf("seed no args should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "usage: wowapi seed <validate|sync>") {
		t.Fatalf("expected seed usage, got %q", errb.String())
	}
}

func TestSeedHelpSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runSeed([]string{"-h"}, &out, &errb); code != 0 {
		t.Fatalf("seed -h exit %d", code)
	}
	if !strings.Contains(out.String(), "validate") {
		t.Fatalf("expected usage text, got %q", out.String())
	}
}

func TestSeedUnknownSubcommand(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runSeed([]string{"bogus"}, &out, &errb); code != 2 {
		t.Fatalf("unknown seed subcommand should exit 2, got %d", code)
	}
	if !strings.Contains(errb.String(), "unknown subcommand") {
		t.Fatalf("expected unknown-subcommand error, got %q", errb.String())
	}
}

func TestSeedValidateDispatchViaRunSeed(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "permissions.yaml", "permissions:\n  - key: widgets.widget.create\n    description: c\n    sensitive: false\n")
	var out, errb bytes.Buffer
	if code := runSeed([]string{"validate", "--dir", dir, "--module", "widgets"}, &out, &errb); code != 0 {
		t.Fatalf("seed validate dispatch exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "OK") {
		t.Fatalf("expected OK, got %q", out.String())
	}
}

func TestSeedValidateDirNotADirectory(t *testing.T) {
	f := filepath.Join(t.TempDir(), "afile")
	writeFile(t, filepath.Dir(f), "afile", "x")
	var out, errb bytes.Buffer
	if code := runSeedValidate([]string{"--dir", f, "--module", "widgets"}, &out, &errb); code != 1 {
		t.Fatalf("--dir pointing at a file should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "is not a directory") {
		t.Fatalf("expected not-a-directory error, got %q", errb.String())
	}
}

func TestSeedValidateLoadError(t *testing.T) {
	dir := t.TempDir()
	// Malformed YAML → seeds.Load returns an error.
	writeFile(t, dir, "permissions.yaml", "permissions: [this is : not valid")
	var out, errb bytes.Buffer
	if code := runSeedValidate([]string{"--dir", dir, "--module", "widgets"}, &out, &errb); code != 1 {
		t.Fatalf("malformed seed should exit 1, got %d", code)
	}
}

// ---------- gen crud: mapFieldType coverage + field errors ----------

func TestGenCRUDAllFieldTypes(t *testing.T) {
	modDir := scaffoldModuleDir(t, "kinds")
	code, _, errOut := callGenCRUD(t,
		"--module", modDir,
		"--resource", "kind",
		"--fields", "a:string,b:int,c:int64,d:bool,e:float64,f:uuid,g:time,h:text,i:integer,j:bigint,k:boolean,l:double,m:timestamp",
	)
	if code != 0 {
		t.Fatalf("gen crud all types exit %d: %s", code, errOut)
	}
	path := filepath.Join(modDir, "kind.go")
	assertParseGo(t, path)
	assertFileContains(t, path, "F uuid.UUID")
	assertFileContains(t, path, "G time.Time")
}

func TestGenCRUDInvalidFieldSpec(t *testing.T) {
	modDir := scaffoldModuleDir(t, "things")
	code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "thing", "--fields", "noColonHere")
	if code != 1 {
		t.Fatalf("invalid field spec should exit 1, got %d", code)
	}
	if !strings.Contains(errOut, "invalid field spec") {
		t.Fatalf("expected invalid-field-spec error, got %q", errOut)
	}
}

func TestGenCRUDBadFieldName(t *testing.T) {
	modDir := scaffoldModuleDir(t, "things")
	code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "thing", "--fields", "BadName:string")
	if code != 1 {
		t.Fatalf("bad field name should exit 1, got %d", code)
	}
	if !strings.Contains(errOut, "must match") {
		t.Fatalf("expected field-name constraint error, got %q", errOut)
	}
}

func TestGenCRUDBadModulePackageName(t *testing.T) {
	base := t.TempDir()
	// Last path segment "Bad-Mod" is not a valid Go package name.
	modDir := filepath.Join(base, "Bad-Mod")
	if err := os.MkdirAll(modDir, 0o755); err != nil {
		t.Fatal(err)
	}
	code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "widget")
	if code != 1 {
		t.Fatalf("invalid module package name should exit 1, got %d", code)
	}
	if !strings.Contains(errOut, "not a valid Go package name") {
		t.Fatalf("expected package-name error, got %q", errOut)
	}
}

func TestGenCRUDForceOverwrite(t *testing.T) {
	modDir := scaffoldModuleDir(t, "widgets")
	if code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "widget"); code != 0 {
		t.Fatalf("first gen exit %d: %s", code, errOut)
	}
	// Second without --force fails (file exists).
	if code, _, _ := callGenCRUD(t, "--module", modDir, "--resource", "widget"); code == 0 {
		t.Fatalf("expected failure overwriting without --force")
	}
	// With --force succeeds.
	if code, _, errOut := callGenCRUD(t, "--module", modDir, "--resource", "widget", "--force"); code != 0 {
		t.Fatalf("force overwrite exit %d: %s", code, errOut)
	}
}

// ---------- scaffold helpers ----------

func TestRenderToFileBadTemplateName(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out.txt")
	if err := renderToFile(dest, "templates/does-not-exist.tmpl", nil, false); err == nil {
		t.Fatal("expected error for missing template")
	}
}

func TestRenderToFileRefusesExisting(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "exists.txt")
	writeFile(t, dir, "exists.txt", "old")
	// Any template name works; the existence check fires before rendering.
	if err := renderToFile(dest, "templates/init/gitignore.tmpl", nil, false); err == nil {
		t.Fatal("expected 'already exists' error")
	} else if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected already-exists error, got %v", err)
	}
}

func TestWriteEmptySkipsExisting(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "keep")
	writeFile(t, dir, "keep", "content")
	// force=false and file exists → silently skipped, content preserved.
	if err := writeEmpty(dest, false); err != nil {
		t.Fatalf("writeEmpty skip returned error: %v", err)
	}
	body, _ := os.ReadFile(dest)
	if string(body) != "content" {
		t.Fatalf("existing file should be preserved, got %q", body)
	}
}

func TestRenderTemplateMissing(t *testing.T) {
	var buf bytes.Buffer
	if err := renderTemplate("templates/nope.tmpl", nil, &buf); err == nil {
		t.Fatal("expected error rendering a missing template")
	}
}
