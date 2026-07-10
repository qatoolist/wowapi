package cli

import (
	"bytes"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
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

// TestInitHintPointsToReadme: the next-steps hint must NOT imply `make migrate-up`
// works bare — it needs APP_ENV + the DB DSNs + a running Postgres, all documented
// in the generated README. The hint points there instead of over-promising.
func TestInitHintPointsToReadme(t *testing.T) {
	base := t.TempDir()
	code, out, errOut := callInit(t, "myapp", "--module", "github.com/acme/myapp", "--dir", base)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, "README") {
		t.Errorf("init hint should point to the README for the env-dependent steps; got:\n%s", out)
	}
	if strings.Contains(out, "migrate-up") {
		t.Errorf("init hint must not imply `make migrate-up` runs with no setup; got:\n%s", out)
	}
}

// TestInitMigrateMainSyncsSeeds is the GAP-003 regression: the generated
// cmd/migrate must run seeds.Sync after module migrations so a fresh
// production database gets its authorization/resource catalogs populated —
// without this, the framework docs' own PF-9 finding (deploy → empty
// catalogs → deny-everything) reproduces in every scaffolded product.
func TestInitMigrateMainSyncsSeeds(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	migratePath := filepath.Join(dir, "cmd", "migrate", "main.go")
	assertFileContains(t, migratePath, "kernel/seeds")
	assertFileContains(t, migratePath, "seeds.Sync(ctx, pool, booted.Seeds)")
	assertParseGo(t, migratePath)
}

// TestInitMigrateMainSyncsRuleDefinitions is the GAP-007 lifecycle regression,
// mirroring TestInitMigrateMainSyncsSeeds exactly: the generated cmd/migrate
// must run rules.SyncDefinitions AFTER seeds.Sync (same privileged pool,
// same deploy point) so a fresh database gets its rule_definitions mirror
// populated — without this, rule_versions.rule_key's FK fails for any
// registered point until a product hand-writes a SQL mirror (the gap
// wowsociety's rulemirror_test.go drift guard existed to catch).
func TestInitMigrateMainSyncsRuleDefinitions(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	migratePath := filepath.Join(dir, "cmd", "migrate", "main.go")
	assertFileContains(t, migratePath, "kernel/rules")
	assertFileContains(t, migratePath, "rules.SyncDefinitions(ctx, pool, k.Rules)")
	assertParseGo(t, migratePath)

	// Ordering: rule-definition sync must run after seed sync in the file text
	// (matches the runtime ordering requirement — rule_versions may reference
	// rule_definitions, and seeds establish the catalogs rule points may need).
	content, err := os.ReadFile(migratePath)
	if err != nil {
		t.Fatal(err)
	}
	seedIdx := strings.Index(string(content), "seeds.Sync(ctx, pool, booted.Seeds)")
	ruleIdx := strings.Index(string(content), "rules.SyncDefinitions(ctx, pool, k.Rules)")
	if seedIdx == -1 || ruleIdx == -1 || ruleIdx < seedIdx {
		t.Fatalf("rule-definition sync must appear AFTER seed sync in generated migrate main (seedIdx=%d ruleIdx=%d)", seedIdx, ruleIdx)
	}
}

// TestInitAPIMainWiresSeedCatalogsReadinessCheck is the GAP-003 "clear failure
// mode" acceptance criterion: the generated api main must wire
// app.CatalogsSeeded into /readyz so a pod whose migrate step skipped seed
// sync reports NOT ready with an actionable message, instead of only
// surfacing as scattered per-request 403s.
func TestInitAPIMainWiresSeedCatalogsReadinessCheck(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	apiPath := filepath.Join(dir, "cmd", "api", "main.go")
	assertFileContains(t, apiPath, "app.CatalogsSeeded")
	assertParseGo(t, apiPath)
}

// ---------- GAP-008: storage / OIDC / i18n scaffold wiring ----------

// TestInitAppcfgHasStorageConfig: the generated internal/appcfg.Config must
// declare a StorageConfig section (mirroring wowsociety's hand-written one)
// so a product can enable the S3/MinIO adapter via config alone, with no
// product-side config-struct boilerplate. Validate() must delegate to it
// (ARCH-10 composition contract), matching how Auth is already handled.
func TestInitAppcfgHasStorageConfig(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	cfgPath := filepath.Join(dir, "internal", "appcfg", "config.go")
	assertParseGo(t, cfgPath)
	for _, want := range []string{
		"StorageConfig",
		`conf:"storage"`,
		"func (s StorageConfig) Enabled() bool",
		"c.Storage.Validate()",
	} {
		assertFileContains(t, cfgPath, want)
	}
}

// TestInitAPIMainWiresOptionalStorage: the generated cmd/api main must wire the
// S3/MinIO adapter into kernel.Deps.Storage when cfg.Storage.Enabled(), with no
// product-side boilerplate beyond what wowapi generated (GAP-008 acceptance:
// "standard storage ... configuration can be wired without product-specific
// boilerplate").
func TestInitAPIMainWiresOptionalStorage(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	apiPath := filepath.Join(dir, "cmd", "api", "main.go")
	assertParseGo(t, apiPath)
	for _, want := range []string{
		`s3adapter "github.com/qatoolist/wowapi/adapters/storage/s3"`,
		"kernel/storage",
		"cfg.Storage.Enabled()",
		"s3adapter.New(ctx, s3adapter.Config{",
		"Storage: store",
	} {
		assertFileContains(t, apiPath, want)
	}
}

// TestInitWorkerMainWiresOptionalStorage mirrors TestInitAPIMainWiresOptionalStorage
// for cmd/worker: the worker touches the same store for document/retention jobs.
func TestInitWorkerMainWiresOptionalStorage(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	workerPath := filepath.Join(dir, "cmd", "worker", "main.go")
	assertParseGo(t, workerPath)
	for _, want := range []string{
		`s3adapter "github.com/qatoolist/wowapi/adapters/storage/s3"`,
		"cfg.Storage.Enabled()",
		"s3adapter.New(ctx, s3adapter.Config{",
		"Storage:  store",
	} {
		assertFileContains(t, workerPath, want)
	}
}

// TestInitAPIMainWiresLocaleMiddleware: the generated api main must install the
// framework's i18n locale-negotiation middleware (kernel/i18n landed this
// branch) unconditionally — booted.I18n is always a non-nil catalog
// (framework English is pre-loaded), so this is a pure framework-standard
// concern with no product config gate, unlike storage/OIDC.
func TestInitAPIMainWiresLocaleMiddleware(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	apiPath := filepath.Join(dir, "cmd", "api", "main.go")
	assertParseGo(t, apiPath)
	assertFileContains(t, apiPath, "httpx.Locale(booted.I18n)")
}

// TestInitConfigsBaseDocumentsStorage: the generated configs/base.yaml should
// document the optional storage section the same way it already documents
// auth.oidc, so a product discovers the knob without reading Go source.
func TestInitConfigsBaseDocumentsStorage(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	assertFileContains(t, filepath.Join(dir, "configs", "base.yaml"), "storage:")
}

// buildRenderedProduct scaffolds a product into a fresh temp dir, points its
// go.mod at THIS wowapi checkout via a replace directive (so the compile test
// runs entirely offline against the framework under development, never a
// published version), and runs `go mod tidy`. It returns the product dir.
// Callers then run `go build ./...` (and any other go tool) inside it.
func buildRenderedProduct(t *testing.T, extraInitArgs ...string) string {
	t.Helper()
	wowapiDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	// scaffold_test.go lives in internal/cli; the module root is two levels up.
	wowapiRoot, err := filepath.Abs(filepath.Join(wowapiDir, "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(wowapiRoot, "go.mod")); err != nil {
		t.Fatalf("could not locate wowapi module root at %s: %v", wowapiRoot, err)
	}

	dir := t.TempDir()
	args := append([]string{"--module", "github.com/acme/compiletest", "--dir", dir}, extraInitArgs...)
	code, _, errOut := callInit(t, args...)
	if code != 0 {
		t.Fatalf("init exit %d: %s", code, errOut)
	}

	gomodPath := filepath.Join(dir, "go.mod")
	gomod, err := os.ReadFile(gomodPath)
	if err != nil {
		t.Fatal(err)
	}
	replaced := string(gomod) + fmt.Sprintf("\nreplace github.com/qatoolist/wowapi => %s\n", wowapiRoot)
	if err := os.WriteFile(gomodPath, []byte(replaced), 0o644); err != nil {
		t.Fatal(err)
	}

	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = dir
	if out, err := tidy.CombinedOutput(); err != nil {
		t.Fatalf("go mod tidy failed: %v\n%s", err, out)
	}
	return dir
}

// TestInitRenderedProductCompiles is the GAP-008 regression net: the scaffold
// must not merely PARSE (assertParseGo) but actually COMPILE against the real
// framework, in the zero-config state (storage disabled, OIDC disabled — the
// var store storage.Adapter / var userAuth httpx.Authenticator branches both
// take their nil/DenyAll path). This is the same binary regardless of runtime
// config, so it also covers the "enabled" code paths at the type level; the
// dedicated config-composition test below additionally proves the enabled
// runtime values round-trip through the strict config loader.
func TestInitRenderedProductCompiles(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles the rendered product against the real framework; skipped in -short")
	}
	dir := buildRenderedProduct(t)
	build := exec.Command("go", "build", "./...")
	build.Dir = dir
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build ./... failed on rendered product:\n%s", out)
	}
}

// TestInitConfigcheckSchemaCoversStorageAndAuth proves the generated
// tools/configcheck links the COMPOSED product config (appcfg.Config) — not
// just config.Framework — including the new Storage/Auth sections, without
// any hand-written checker (GAP-008 acceptance criterion #1). `schema` needs
// no config files on disk, so it is a fast, deterministic way to prove the
// composed type is actually reachable through the generated tool.
func TestInitConfigcheckSchemaCoversStorageAndAuth(t *testing.T) {
	if testing.Short() {
		t.Skip("runs `go run ./tools/configcheck` against the real framework; skipped in -short")
	}
	dir := buildRenderedProduct(t)
	cc := exec.Command("go", "run", "./tools/configcheck", "schema")
	cc.Dir = dir
	out, err := cc.CombinedOutput()
	if err != nil {
		t.Fatalf("configcheck schema failed: %v\n%s", err, out)
	}
	for _, want := range []string{`"storage"`, `"auth"`, `"oidc"`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("configcheck schema missing %q in output:\n%s", want, out)
		}
	}
}

// TestInitConfigcheckValidatesStorageAndOIDCOverlay proves a product can
// enable storage + OIDC purely via config (a local.yaml-style overlay) and
// have the GENERATED configcheck validate it end to end — no hand-written
// checker, no product-side config struct edits (GAP-008 acceptance criterion
// #3: "standard storage and OIDC/JWT configuration can be wired without
// product-specific boilerplate"). The strict config loader rejects unknown
// keys, so this also proves Storage/Auth are wired into appcfg.Config for
// real, not just present as dead struct fields.
func TestInitConfigcheckValidatesStorageAndOIDCOverlay(t *testing.T) {
	if testing.Short() {
		t.Skip("runs `go run ./tools/configcheck` against the real framework; skipped in -short")
	}
	dir := buildRenderedProduct(t)

	// "dev" (not a bespoke name) — config.Framework.Environment is a closed
	// enum (local|dev|stage|prod); the overlay must declare a value the
	// strict loader accepts, and --env dev asserts the overlay actually
	// declares environment: dev (configcheck's CI-gate behavior).
	overlay := `environment: dev
log:
  level: debug
  format: text
db:
  dsn: "secretref://env/DATABASE_URL"
  migrate_dsn: "secretref://env/MIGRATE_URL"
  platform_dsn: "secretref://env/PLATFORM_URL"
auth:
  oidc:
    issuer: "https://idp.example.com/"
    audience: "compiletest"
storage:
  endpoint: "localhost:9000"
  bucket: "compiletest-docs"
  access_key: "secretref://env/S3_ACCESS_KEY"
  secret_key: "secretref://env/S3_SECRET_KEY"
  presign_ttl: 15m
`
	if err := os.WriteFile(filepath.Join(dir, "configs", "dev.yaml"), []byte(overlay), 0o644); err != nil {
		t.Fatal(err)
	}

	cc := exec.Command("go", "run", "./tools/configcheck", "validate", "--env", "dev")
	cc.Dir = dir
	cc.Env = append(os.Environ(),
		"DATABASE_URL=postgres://app_rt:x@localhost:5432/compiletest?sslmode=disable",
		"MIGRATE_URL=postgres://app_migrate:x@localhost:5432/compiletest?sslmode=disable",
		"PLATFORM_URL=postgres://app_platform:x@localhost:5432/compiletest?sslmode=disable",
		"S3_ACCESS_KEY=minioadmin",
		"S3_SECRET_KEY=minioadmin",
	)
	out, err := cc.CombinedOutput()
	if err != nil {
		t.Fatalf("configcheck validate --env ci failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "OK: configuration valid") {
		t.Errorf("expected validation success, got:\n%s", out)
	}
}

// TestInitMigrateMainLoadsComposedConfig: the generated cmd/migrate must load
// the COMPOSED product config (appcfg.Load) — not bare config.Framework — for
// the same reason wowsociety's hand-written migrate does: the strict loader
// rejects unknown keys, so product-owned sections in the deployed overlays
// (auth.*, storage.*) would otherwise abort the migrate while the generated
// api/worker accept the very same file. One overlay must serve all three
// processes (GAP-008 follow-up).
func TestInitMigrateMainLoadsComposedConfig(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	migratePath := filepath.Join(dir, "cmd", "migrate", "main.go")
	assertParseGo(t, migratePath)
	assertFileContains(t, migratePath, `"github.com/acme/myapp/internal/appcfg"`)
	assertFileContains(t, migratePath, "appcfg.Load()")
	// The kernel still receives only the framework subset.
	assertFileContains(t, migratePath, "kernel.New(cfg.Framework")
	// The old framework-only load path must be gone.
	content, err := os.ReadFile(migratePath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "config.Load[config.Framework]") {
		t.Errorf("generated migrate main still loads bare config.Framework — must load the composed appcfg.Config:\n%s", content)
	}
}

// TestInitMigrateMainAcceptsComposedOverlay is the behavioral proof for the
// above: a rendered product whose overlay carries storage:+auth: sections must
// be ACCEPTED by the rendered migrate main's config load. `migrate down`
// against an environment: stage overlay reaches the down-guard ("refusing to
// reset") strictly AFTER config load and BEFORE any DB connection, so the
// guard message proves the composed overlay parsed cleanly with no database
// required — while "unknown key" output would reproduce the bug.
func TestInitMigrateMainAcceptsComposedOverlay(t *testing.T) {
	if testing.Short() {
		t.Skip("runs `go run ./cmd/migrate` against the real framework; skipped in -short")
	}
	dir := buildRenderedProduct(t)

	overlay := `environment: stage
db:
  dsn: "secretref://env/DATABASE_URL"
  migrate_dsn: "secretref://env/MIGRATE_URL"
  platform_dsn: "secretref://env/PLATFORM_URL"
auth:
  oidc:
    issuer: "https://idp.example.com/"
    audience: "compiletest"
storage:
  endpoint: "localhost:9000"
  bucket: "compiletest-docs"
  access_key: "secretref://env/S3_ACCESS_KEY"
  secret_key: "secretref://env/S3_SECRET_KEY"
  presign_ttl: 15m
`
	if err := os.WriteFile(filepath.Join(dir, "configs", "stage.yaml"), []byte(overlay), 0o644); err != nil {
		t.Fatal(err)
	}

	mig := exec.Command("go", "run", "./cmd/migrate", "down")
	mig.Dir = dir
	mig.Env = append(os.Environ(),
		"APP_ENV=stage",
		"DATABASE_URL=postgres://app_rt:x@localhost:5432/compiletest?sslmode=disable",
		"MIGRATE_URL=postgres://app_migrate:x@localhost:5432/compiletest?sslmode=disable",
		"PLATFORM_URL=postgres://app_platform:x@localhost:5432/compiletest?sslmode=disable",
		"S3_ACCESS_KEY=minioadmin",
		"S3_SECRET_KEY=minioadmin",
	)
	out, err := mig.CombinedOutput()
	if err == nil {
		t.Fatalf("migrate down in stage must refuse (down-guard), got success:\n%s", out)
	}
	if strings.Contains(string(out), "unknown key") {
		t.Errorf("migrate rejected the composed overlay (strict loader saw product sections as unknown keys) — it must load appcfg.Config:\n%s", out)
	}
	if !strings.Contains(string(out), "refusing to reset") {
		t.Errorf("expected the down-guard message (proof config load succeeded), got:\n%s", out)
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
