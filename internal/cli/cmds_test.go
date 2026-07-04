package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ---------- migrate create ----------

func TestMigrateCreateNextNumber(t *testing.T) {
	dir := t.TempDir()
	// Seed an existing 00007 migration so the next is 00008.
	if err := os.WriteFile(filepath.Join(dir, "00007_outbox.sql"), []byte("-- +goose Up"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out, errb bytes.Buffer
	code := runMigrateCreate([]string{"--dir", dir, "--name", "add_widgets"}, &out, &errb)
	if code != 0 {
		t.Fatalf("exit %d, stderr=%s", code, errb.String())
	}
	created := strings.TrimSpace(out.String())
	if filepath.Base(created) != "00008_add_widgets.sql" {
		t.Fatalf("created %q, want 00008_add_widgets.sql", created)
	}
	body, _ := os.ReadFile(created)
	if !strings.Contains(string(body), "+goose Up") || !strings.Contains(string(body), "+goose Down") {
		t.Fatalf("migration skeleton missing goose markers:\n%s", body)
	}
}

func TestMigrateCreateEmptyDirStartsAtOne(t *testing.T) {
	dir := t.TempDir()
	var out, errb bytes.Buffer
	if code := runMigrateCreate([]string{"--dir", dir, "--name", "first"}, &out, &errb); code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	if filepath.Base(strings.TrimSpace(out.String())) != "00001_first.sql" {
		t.Fatalf("empty dir should start at 00001, got %s", out.String())
	}
}

func TestMigrateCreateRejectsBadName(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runMigrateCreate([]string{"--dir", t.TempDir(), "--name", "Bad-Name"}, &out, &errb); code != 1 {
		t.Fatalf("bad name should exit 1, got %d", code)
	}
}

// ---------- openapi merge ----------

func TestOpenAPIMerge(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.json", `{"paths":{"/a":{"get":{}}},"components":{"schemas":{"A":{"type":"object"}}}}`)
	writeFile(t, dir, "b.json", `{"paths":{"/b":{"get":{}}},"components":{"schemas":{"B":{"type":"object"}}}}`)
	var out, errb bytes.Buffer
	code := runOpenAPI([]string{"merge", "--dir", dir, "--title", "T", "--version", "1.2.3"}, &out, &errb)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	got := out.String()
	for _, want := range []string{`"/a"`, `"/b"`, `"A"`, `"B"`, `"1.2.3"`, `"3.1.0"`} {
		if !strings.Contains(got, want) {
			t.Fatalf("merged doc missing %s:\n%s", want, got)
		}
	}
}

func TestOpenAPIMergeDuplicatePathFails(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.json", `{"paths":{"/x":{"get":{}}}}`)
	writeFile(t, dir, "b.json", `{"paths":{"/x":{"post":{}}}}`)
	var out, errb bytes.Buffer
	if code := runOpenAPI([]string{"merge", "--dir", dir}, &out, &errb); code != 1 {
		t.Fatalf("duplicate path should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "duplicate path") {
		t.Fatalf("expected duplicate-path error, got %s", errb.String())
	}
}

// ---------- seed validate ----------

func TestSeedValidateOK(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "permissions.yaml", "permissions:\n  - key: widgets.widget.create\n    description: c\n    sensitive: false\n")
	var out, errb bytes.Buffer
	code := runSeedValidate([]string{"--dir", dir, "--module", "widgets"}, &out, &errb)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "OK") || !strings.Contains(out.String(), "1 permissions") {
		t.Fatalf("unexpected output: %s", out.String())
	}
}

func TestSeedValidateForeignKeyFails(t *testing.T) {
	dir := t.TempDir()
	// A key not prefixed by the module must fail validation.
	writeFile(t, dir, "permissions.yaml", "permissions:\n  - key: other.thing.create\n    description: c\n    sensitive: false\n")
	var out, errb bytes.Buffer
	if code := runSeedValidate([]string{"--dir", dir, "--module", "widgets"}, &out, &errb); code != 1 {
		t.Fatalf("foreign-prefixed key should exit 1, got %d (%s)", code, errb.String())
	}
}

func TestSeedValidateRequiresModule(t *testing.T) {
	var out, errb bytes.Buffer
	// A missing required flag is a usage error → exit 2 (CLI-03).
	if code := runSeedValidate([]string{"--dir", t.TempDir()}, &out, &errb); code != 2 {
		t.Fatalf("missing --module should exit 2, got %d", code)
	}
}

func TestMigrateCreateMissingNameIsUsageError(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runMigrateCreate([]string{"--dir", t.TempDir()}, &out, &errb); code != 2 {
		t.Fatalf("missing --name should exit 2 (usage), got %d", code)
	}
}

// ---------- openapi merge: reject non-object fragments (CLI-02) ----------

func TestOpenAPIMergeRejectsNullFragment(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "a.json", `{"paths":{"/a":{"get":{}}}}`)
	writeFile(t, dir, "null.json", `null`)
	var out, errb bytes.Buffer
	if code := runOpenAPI([]string{"merge", "--dir", dir}, &out, &errb); code != 1 {
		t.Fatalf("a null fragment must fail, got exit %d", code)
	}
	if !strings.Contains(errb.String(), "expected a JSON object") {
		t.Fatalf("expected object-required error, got %s", errb.String())
	}
}

// ---------- lint boundaries (pure checker) ----------

func TestCheckBoundariesModuleIsolation(t *testing.T) {
	const mod = "github.com/acme/app"
	imports := map[string][]string{
		mod + "/internal/modules/billing/service": {mod + "/internal/modules/catalog/store", "fmt"},
		mod + "/internal/modules/catalog/store":   {"fmt"},
	}
	v := checkBoundaries(imports, mod, false)
	if len(v) != 1 || !strings.Contains(v[0], `module "billing" imports module "catalog"`) {
		t.Fatalf("expected one cross-module violation, got %v", v)
	}
}

func TestCheckBoundariesClean(t *testing.T) {
	const mod = "github.com/acme/app"
	imports := map[string][]string{
		mod + "/internal/modules/billing/service": {mod + "/kernel/database", "fmt"},
	}
	if v := checkBoundaries(imports, mod, false); len(v) != 0 {
		t.Fatalf("expected no violations, got %v", v)
	}
}

func TestCheckBoundariesFrameworkLayering(t *testing.T) {
	const mod = "github.com/qatoolist/wowapi"
	imports := map[string][]string{
		mod + "/kernel/config": {mod + "/app", "fmt"}, // kernel importing app — illegal
	}
	v := checkBoundaries(imports, mod, true)
	if len(v) != 1 || !strings.Contains(v[0], "kernel must not import app") {
		t.Fatalf("expected kernel-layer violation, got %v", v)
	}
}

// ---------- deploy render ----------

func TestDeployRenderCompose(t *testing.T) {
	var out, errb bytes.Buffer
	// --env must be a config-valid environment ("stage", not "staging").
	code := runDeploy([]string{"render", "--format", "compose", "--name", "acme", "--image", "acme:1.0", "--env", "stage"}, &out, &errb)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errb.String())
	}
	got := out.String()
	// The DSN must be a secretref reference (config.DB.DSN is a Secret), never ${VAR}.
	for _, want := range []string{"acme-api", "acme-worker", "acme-migrate", "acme:1.0", "stage", "secretref://env/WOWAPI_DB_DSN"} {
		if !strings.Contains(got, want) {
			t.Fatalf("compose manifest missing %s:\n%s", want, got)
		}
	}
	if strings.Contains(got, "${WOWAPI_DB_DSN}") {
		t.Fatalf("manifest must not emit a raw ${VAR} DSN (config.Secret needs a secretref):\n%s", got)
	}
}

// A --env value the config loader would reject must fail render, not emit a
// manifest that cannot boot (finding: default was "production", valid is "prod").
func TestDeployRenderRejectsInvalidEnv(t *testing.T) {
	for _, bad := range []string{"production", "staging", "PROD", "qa"} {
		var out, errb bytes.Buffer
		if code := runDeploy([]string{"render", "--env", bad}, &out, &errb); code != 2 {
			t.Fatalf("--env %q should be rejected (exit 2), got %d", bad, code)
		}
	}
	// The documented default renders cleanly (it is a valid env).
	var out, errb bytes.Buffer
	if code := runDeploy([]string{"render"}, &out, &errb); code != 0 {
		t.Fatalf("default --env must render: exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "WOWAPI__ENVIRONMENT: prod") {
		t.Fatalf("default env should be prod:\n%s", out.String())
	}
}

func TestDeployRenderEnvAndBadFormat(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runDeploy([]string{"render", "--format", "env", "--env", "prod"}, &out, &errb); code != 0 {
		t.Fatalf("env render exit %d: %s", code, errb.String())
	}
	if !strings.Contains(out.String(), "WOWAPI__ENVIRONMENT=prod") {
		t.Fatalf("env output missing environment: %s", out.String())
	}
	out.Reset()
	errb.Reset()
	if code := runDeploy([]string{"render", "--format", "bogus"}, &out, &errb); code != 2 {
		t.Fatalf("bad format should exit 2, got %d", code)
	}
}

// TestGenCRUDRejectsUnknownFieldType is the CLI-01 regression: an unsupported
// field type must fail (exit 1) rather than emit undefined, unbuildable Go.
func TestGenCRUDRejectsUnknownFieldType(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "widgets")
	var out, errb bytes.Buffer
	code := runGen([]string{"crud", "--module", dir, "--resource", "widget", "--fields", "price:decimal"}, &out, &errb)
	if code != 1 {
		t.Fatalf("unknown field type should exit 1, got %d", code)
	}
	if !strings.Contains(errb.String(), "unknown field type") {
		t.Fatalf("expected unknown-type error, got %s", errb.String())
	}
	// Nothing should have been written.
	if _, err := os.Stat(filepath.Join(dir, "widget.go")); err == nil {
		t.Fatal("no file should be generated when a field type is invalid")
	}
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
