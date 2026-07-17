package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The B1 (GAP-001B) i18n scaffold acceptance tests. Fast tests assert the
// generated files/wiring by text; the slow acceptance test renders, compiles,
// and RUNS a product proving Accept-Language returns product/kernel-override/
// YAML/JSON/Go strings with NO product-authored loader code.

// TestInitScaffoldsLocalesTree: init generates the locales/ tree (en + a sample
// second locale mr), product YAML + JSON catalogs, and the compiled Go bundle
// sample.
func TestInitScaffoldsLocalesTree(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	for _, rel := range []string{
		"locales/en/kernel.yaml",
		"locales/en/product.yaml",
		"locales/en.json",
		"locales/mr/kernel.yaml",
		"locales/mr/product.yaml",
		"internal/i18n/catalogs/en.go",
	} {
		assertFileExists(t, filepath.Join(dir, filepath.FromSlash(rel)))
	}
	// The en kernel override retranslates a framework key; the Go bundle is valid Go.
	assertFileContains(t, filepath.Join(dir, "locales", "en", "kernel.yaml"), "kernel.problem.not_found")
	assertParseGo(t, filepath.Join(dir, "internal", "i18n", "catalogs", "en.go"))
}

// TestInitI18nConfigSection: the appcfg config carries an I18nConfig section and
// configs/base.yaml documents the i18n: block, so a product discovers the knob.
func TestInitI18nConfigSection(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	cfgPath := filepath.Join(dir, "internal", "appcfg", "config.go")
	assertParseGo(t, cfgPath)
	assertFileContains(t, cfgPath, "I18n I18nConfig")
	assertFileContains(t, cfgPath, "func (c I18nConfig) Layers()")
	assertFileContains(t, filepath.Join(dir, "configs", "base.yaml"), "i18n:")
}

// TestInitThreeBinariesLoadSameI18nSources: api, worker, AND migrate must build
// the SAME i18n layers from cfg.I18n and pass them into app.Boot, so all three
// share one catalog lifecycle (benchmark acceptance: "init renders API, worker,
// and migrate binaries that load the same configured i18n catalog sources before
// boot completes").
func TestInitThreeBinariesLoadSameI18nSources(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	for _, main := range []string{"api", "worker", "migrate"} {
		p := filepath.Join(dir, "cmd", main, "main.go")
		assertParseGo(t, p)
		assertFileContains(t, p, "cfg.I18n.Layers()")
		assertFileContains(t, p, "app.WithI18nLayers(i18nLayers...)")
	}
}

// TestInitI18nAcceptanceEndToEnd is the load-bearing proof: render a product,
// point it at this framework checkout, drop in a product-side _test.go that boots
// the catalog exactly as the generated binaries do (cfg.I18n.Layers() ->
// app.Boot(app.WithI18nLayers)), and run `go test` inside it. It proves — with no
// product-authored loader code — that a product-local kernel.* override beats the
// embedded framework default AND that a product YAML key, a product JSON key, and
// a compiled Go bundle key are all served through the ONE lifecycle, negotiated
// by Accept-Language.
func TestInitI18nAcceptanceEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("compiles and runs the rendered product against the real framework; skipped in -short")
	}
	dir := buildRenderedProduct(t)

	// A product-side test: the ONLY i18n code a product writes is this test's
	// assertions — never a loader. It uses the generated appcfg + the framework
	// httpx/i18n to negotiate Accept-Language and read the merged catalog.
	acceptanceTest := `package acceptance

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/i18n"

	"github.com/acme/compiletest/internal/appcfg"
)

// chdirToProductRoot moves CWD to the directory holding go.mod, so the relative
// locales/ path in appcfg's i18n config resolves — exactly as the deployed
// binaries run from the product root. Go test starts in the test package dir.
func chdirToProductRoot(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for dir := wd; ; {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if err := os.Chdir(dir); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = os.Chdir(wd) })
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from " + wd)
		}
		dir = parent
	}
}

// buildCatalog reproduces what cmd/api|worker|migrate do at boot: build layers
// from config, apply them over the framework defaults, freeze. No loader code.
func buildCatalog(t *testing.T) *i18n.Catalog {
	t.Helper()
	chdirToProductRoot(t)
	cfg := appcfg.I18nConfig{DefaultLocale: "en", SupportedLocales: []string{"en", "mr"}, LocalesDir: "locales", GoBundles: true}
	layers, err := cfg.Layers()
	if err != nil {
		t.Fatalf("Layers: %v", err)
	}
	r := i18n.NewRegistry()
	r.ApplyLayers(layers...)
	if err := r.Err(); err != nil {
		t.Fatalf("ApplyLayers: %v", err)
	}
	r.Freeze()
	return r.Catalog()
}

func TestAcceptLanguageServesAllSources(t *testing.T) {
	cat := buildCatalog(t)
	nf := i18n.KeyProblemTitle(errors.KindNotFound)

	cases := []struct{ accept, key, want string }{
		// product locales/en/kernel.yaml override BEATS the embedded framework default "Not found".
		{"en", nf, "Resource not found"},
		// mr override via Accept-Language negotiation.
		{"mr", nf, "संसाधन सापडले नाही"},
		// product YAML / JSON / Go bundle keys, all one lifecycle.
		{"en", "compiletest.i18n.sample_yaml", "Loaded from a product YAML catalog"},
		{"en", "compiletest.i18n.sample_json", "Loaded from a product JSON catalog"},
		{"en", "compiletest.i18n.sample_go", "Loaded from a compiled Go catalog bundle"},
	}
	for _, c := range cases {
		// Negotiate the locale from an Accept-Language header exactly as the edge
		// middleware does, then look the key up in that locale.
		loc := i18n.Negotiate(c.accept, cat.Locales(), cat.Default())
		got, _ := cat.Lookup(loc, c.key)
		if got != c.want {
			t.Errorf("Accept-Language %q key %q: got %q, want %q", c.accept, c.key, got, c.want)
		}
	}
}
`
	acceptDir := filepath.Join(dir, "internal", "acceptance")
	if err := os.MkdirAll(acceptDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(acceptDir, "i18n_test.go"), []byte(acceptanceTest), 0o644); err != nil {
		t.Fatal(err)
	}

	cmd := exec.Command("go", "test", "./internal/acceptance/...")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("product i18n acceptance test failed:\n%s", out)
	}
}

// TestInitI18nValidatePassesOnScaffold: the shipped locales/ tree must pass
// `wowapi i18n validate` out of the box (a product's CI runs this), proving the
// scaffolded catalogs are coverage-clean, ownership-clean, and placeholder-clean.
func TestInitI18nValidatePassesOnScaffold(t *testing.T) {
	dir := t.TempDir()
	code, _, errOut := callInit(t, "--module", "github.com/acme/myapp", "--dir", dir)
	if code != 0 {
		t.Fatalf("init exit %d: %s", code, errOut)
	}
	var out, errb strings.Builder
	vc := runI18n([]string{"validate", "--dir", filepath.Join(dir, "locales"), "--supported", "en,mr"}, &out, &errb)
	if vc != 0 {
		t.Fatalf("wowapi i18n validate failed on scaffold: %s", errb.String())
	}
	if !strings.Contains(out.String(), "OK:") {
		t.Fatalf("expected OK, got %q", out.String())
	}
}
