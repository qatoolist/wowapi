package testkit_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestIntegrationScratchConsumer is the headline Phase 5 proof (Goal 2 §Definition
// of Done 3–4): an EXTERNAL product repo can import wowapi through its public
// packages, define a module, and pass the module contract suite — without any
// edits to the framework. It scaffolds a throwaway module in a temp dir,
// `replace`s wowapi with this working tree, and runs `go test` there.
//
// It needs the Go toolchain and a database DSN; it skips when either is absent.
func TestIntegrationScratchConsumer(t *testing.T) {
	dsn := os.Getenv("WOWAPI_TEST_DSN")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		t.Skip("scratch-consumer test needs a database DSN (run `make up` and export DATABASE_URL)")
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not found")
	}

	repoRoot := findRepoRoot(t)
	dir := t.TempDir()
	writeFile(t, dir, "go.mod", `module scratch.example/acme

go 1.26

require github.com/qatoolist/wowapi v0.0.0

replace github.com/qatoolist/wowapi => `+repoRoot+"\n")

	// A minimal external module using ONLY public packages.
	writeFile(t, dir, "mod.go", scratchModuleSrc)
	writeFile(t, dir, "migrations/00001_widgets.sql", scratchMigration)
	writeFile(t, dir, "seeds/perms.yaml", scratchSeeds)
	writeFile(t, dir, "mod_test.go", scratchTest)

	// Resolve deps from the ambient module cache (warm from this repo's build),
	// not the network: the framework's own go.mod already pins every transitive
	// dep. If resolution genuinely can't be satisfied offline, skip rather than
	// fail — this test proves import-surface sufficiency, not proxy access
	// (review finding ARCH-50).
	env := append(os.Environ(), "DATABASE_URL="+dsn, "GOFLAGS=-mod=mod")
	tidy := exec.Command("go", "mod", "tidy")
	tidy.Dir = dir
	tidy.Env = env
	if out, err := tidy.CombinedOutput(); err != nil {
		if strings.Contains(string(out), "dial tcp") || strings.Contains(string(out), "lookup") ||
			strings.Contains(string(out), "proxy") || strings.Contains(string(out), "cannot find module") {
			t.Skipf("scratch consumer: module resolution needs network (cold cache); skipping:\n%s", out)
		}
		t.Fatalf("scratch consumer: go mod tidy: %v\n%s", err, out)
	}
	cmd := exec.Command("go", "test", "-run", "TestWidgetsContract", "-count=1", "./...")
	cmd.Dir = dir
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("scratch consumer test failed: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "ok") {
		t.Fatalf("scratch consumer test did not pass:\n%s", out)
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd() // .../wowapi/testkit
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Dir(wd)
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("could not locate repo root from %s", wd)
	}
	return root
}

func writeFile(t *testing.T, dir, rel, body string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

// A tiny external module: a "widgets" resource with one route + seeds + a
// migration, using only wowapi/module + wowapi/kernel/* public packages.
const scratchModuleSrc = `package acme

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/module"
)

//go:embed migrations/*.sql
var migEmbed embed.FS

//go:embed seeds/*.yaml
var seedEmbed embed.FS

type Module struct{}

func (Module) Name() string          { return "widgets" }
func (Module) DependsOn() []string   { return nil }

type Config struct {
	MaxWidgets int ` + "`json:\"max_widgets\"`" + `
}

func (Module) Register(mc module.Context) error {
	cfg := Config{MaxWidgets: 100}
	if err := mc.Config().Decode(&cfg); err != nil {
		return err
	}
	mig, _ := fs.Sub(migEmbed, "migrations")
	seed, _ := fs.Sub(seedEmbed, "seeds")
	mc.Migrations(mig)
	mc.Seeds(seed)
	mc.Routes().Handle(http.MethodGet, "/widgets", httpx.RouteMeta{Permission: "widgets.widget.list"},
		func(w http.ResponseWriter, r *http.Request) { httpx.WriteJSON(w, 200, httpx.OK([]string{})) })
	return nil
}
`

const scratchMigration = `-- +goose Up
CREATE TABLE widgets_widget (
  id uuid PRIMARY KEY,
  tenant_id uuid NOT NULL,
  name text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  version int NOT NULL DEFAULT 1,
  created_at timestamptz NOT NULL DEFAULT now(), created_by uuid NOT NULL,
  updated_at timestamptz, updated_by uuid
);
ALTER TABLE widgets_widget ENABLE ROW LEVEL SECURITY;
ALTER TABLE widgets_widget FORCE ROW LEVEL SECURITY;
CREATE POLICY widgets_widget_tenant_isolation ON widgets_widget
  USING (tenant_id = app_tenant_id()) WITH CHECK (tenant_id = app_tenant_id());
GRANT SELECT, INSERT, UPDATE ON widgets_widget TO app_rt;
-- +goose Down
DROP TABLE IF EXISTS widgets_widget;
`

const scratchSeeds = `permissions:
  - key: widgets.widget.list
    description: list widgets
resource_types:
  - key: widgets.widget
    description: a widget
roles:
  - key: widgets.org.viewer
    name: Viewer
    permissions: [widgets.widget.list]
`

const scratchTest = `package acme

import (
	"testing"

	"github.com/qatoolist/wowapi/testkit"
)

func TestWidgetsContract(t *testing.T) {
	testkit.RunModuleContract(t, Module{})
}
`
