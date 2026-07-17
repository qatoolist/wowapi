package i18n_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/qatoolist/wowapi/v2/kernel/i18n"
)

// frameworkLayer is the always-first embedded-defaults layer used across tests.
func frameworkLayer() i18n.Layer {
	return i18n.Layer{
		Name:    "framework defaults",
		Policy:  i18n.Policy{OwnsFramework: true},
		Sources: []i18n.Source{i18n.FrameworkDefaultsSource()},
	}
}

// overrideLayer wraps sources as a sanctioned product framework-override layer:
// may override kernel.* but not introduce new kernel.* keys.
func overrideLayer(srcs ...i18n.Source) i18n.Layer {
	return i18n.Layer{
		Name:    "product overrides",
		Policy:  i18n.Policy{OwnsFramework: true, FrameworkOverrideOnly: true},
		Sources: srcs,
	}
}

// catalogsLayer is the product/module catalog layer: no framework ownership.
func catalogsLayer(srcs ...i18n.Source) i18n.Layer {
	return i18n.Layer{Name: "catalogs", Sources: srcs}
}

func TestLoadCatalogFrameworkDefaultsOnly(t *testing.T) {
	cat, err := i18n.LoadCatalog("en", frameworkLayer())
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	msg, loc := cat.Lookup("en", i18n.KeyValidationMessage("required"))
	if msg != "this field is required" || loc != "en" {
		t.Fatalf("framework default lookup = (%q,%q)", msg, loc)
	}
}

func TestLoadCatalogPrecedenceProductOverridesKernel(t *testing.T) {
	// Product ships locales/mr/kernel.yaml overriding a framework key for mr,
	// while en still falls back to the embedded framework default. This is the
	// benchmark's acceptance test.
	fsys := fstest.MapFS{
		"locales/mr/kernel.yaml": &fstest.MapFile{Data: []byte(
			"locale: mr\nmessages:\n  " + i18n.KeyProblemTitle(kindNotFound()) + ": \"सापडले नाही\"\n",
		)},
	}
	cat, err := i18n.LoadCatalog(
		"en",
		frameworkLayer(),
		overrideLayer(i18n.NewFSSource(fsys, "locales", "yaml")),
	)
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	// mr override wins.
	if msg, loc := cat.Lookup("mr", i18n.KeyProblemTitle(kindNotFound())); msg != "सापडले नाही" || loc != "mr" {
		t.Fatalf("mr override = (%q,%q), want (सापडले नाही,mr)", msg, loc)
	}
	// en still falls back to the embedded framework default.
	if msg, loc := cat.Lookup("en", i18n.KeyProblemTitle(kindNotFound())); msg != "Not found" || loc != "en" {
		t.Fatalf("en fallback = (%q,%q), want (Not found,en)", msg, loc)
	}
}

func TestLoadCatalogIntraLayerDuplicateFails(t *testing.T) {
	fsys := fstest.MapFS{
		"locales/en/a.yaml": &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  app.hi: \"one\"\n")},
		"locales/en/b.yaml": &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  app.hi: \"two\"\n")},
	}
	_, err := i18n.LoadCatalog("en", catalogsLayer(i18n.NewFSSource(fsys, "locales", "yaml")))
	if err == nil || !strings.Contains(err.Error(), "duplicate key") {
		t.Fatalf("expected intra-layer duplicate error, got %v", err)
	}
}

func TestLoadCatalogLaterLayerOverrideAllowed(t *testing.T) {
	// Same key in two DIFFERENT layers is a legitimate precedence override.
	fsA := fstest.MapFS{"locales/en/a.yaml": &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  app.hi: \"low\"\n")}}
	fsB := fstest.MapFS{"go/en.yaml": &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  app.hi: \"high\"\n")}}
	cat, err := i18n.LoadCatalog(
		"en",
		catalogsLayer(i18n.NewFSSource(fsA, "locales", "yaml")),
		i18n.Layer{Name: "go bundles", Sources: []i18n.Source{i18n.NewFSSource(fsB, "go", "yaml")}},
	)
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	if msg, _ := cat.Lookup("en", "app.hi"); msg != "high" {
		t.Fatalf("later layer should win: got %q", msg)
	}
}

func TestLoadCatalogNonOverrideLayerCannotWriteKernel(t *testing.T) {
	fsys := fstest.MapFS{
		"locales/en/x.yaml": &fstest.MapFile{Data: []byte(
			"locale: en\nmessages:\n  " + i18n.KeyProblemTitle(kindNotFound()) + ": \"hijack\"\n",
		)},
	}
	_, err := i18n.LoadCatalog("en", frameworkLayer(), catalogsLayer(i18n.NewFSSource(fsys, "locales", "yaml")))
	if err == nil || !strings.Contains(err.Error(), "reserved framework key") {
		t.Fatalf("catalog layer writing kernel.* must fail, got %v", err)
	}
}

func TestLoadCatalogOverrideLayerCannotInventKernelKey(t *testing.T) {
	// FrameworkOverrideOnly may retranslate existing kernel.* keys, not invent new ones.
	fsys := fstest.MapFS{
		"locales/en/x.yaml": &fstest.MapFile{Data: []byte(
			"locale: en\nmessages:\n  kernel.problem.made_up: \"nope\"\n",
		)},
	}
	_, err := i18n.LoadCatalog("en", frameworkLayer(), overrideLayer(i18n.NewFSSource(fsys, "locales", "yaml")))
	if err == nil || !strings.Contains(err.Error(), "no lower layer defines") {
		t.Fatalf("override layer inventing kernel.* key must fail, got %v", err)
	}
}

func TestLoadCatalogJSONSource(t *testing.T) {
	fsys := fstest.MapFS{
		"locales/en.json": &fstest.MapFile{Data: []byte(`{"locale":"en","messages":{"app.json":"J"}}`)},
	}
	cat, err := i18n.LoadCatalog("en", catalogsLayer(i18n.NewFSSource(fsys, "locales", "json")))
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	if msg, _ := cat.Lookup("en", "app.json"); msg != "J" {
		t.Fatalf("json key = %q", msg)
	}
}

func TestLoadCatalogGoSource(t *testing.T) {
	src := i18n.NewGoSource(
		i18n.RawBundle{Locale: "en", Origin: "internal/i18n/catalogs/en.go", Messages: map[string]string{"app.go": "G"}},
	)
	cat, err := i18n.LoadCatalog("en", i18n.Layer{Name: "go bundles", Sources: []i18n.Source{src}})
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	if msg, _ := cat.Lookup("en", "app.go"); msg != "G" {
		t.Fatalf("go key = %q", msg)
	}
}

func TestLoadCatalogFullPrecedenceChain(t *testing.T) {
	// framework defaults -> product kernel override -> product catalog -> go bundle.
	overrideFS := fstest.MapFS{
		"locales/en/kernel.yaml": &fstest.MapFile{Data: []byte(
			"locale: en\nmessages:\n  " + i18n.KeyProblemTitle(kindNotFound()) + ": \"Missing\"\n",
		)},
	}
	catalogFS := fstest.MapFS{
		"locales/en/product.yaml": &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  app.greet: \"Hi\"\n")},
		"locales/en.json":         &fstest.MapFile{Data: []byte(`{"messages":{"app.bye":"Bye"}}`)},
	}
	goSrc := i18n.NewGoSource(i18n.RawBundle{Locale: "en", Origin: "go", Messages: map[string]string{"app.compiled": "C"}})

	cat, err := i18n.LoadCatalog(
		"en",
		frameworkLayer(),
		overrideLayer(i18n.NewFSSource(overrideFS, "locales", "yaml")),
		catalogsLayer(i18n.NewFSSource(catalogFS, "locales", "yaml", "json")),
		i18n.Layer{Name: "go bundles", Sources: []i18n.Source{goSrc}},
	)
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	checks := map[string]string{
		i18n.KeyProblemTitle(kindNotFound()): "Missing", // product override beat embedded default
		"app.greet":                          "Hi",      // product yaml
		"app.bye":                            "Bye",     // product json (flat, locale from filename)
		"app.compiled":                       "C",       // go bundle
	}
	for key, want := range checks {
		if msg, _ := cat.Lookup("en", key); msg != want {
			t.Errorf("Lookup(%q) = %q, want %q", key, msg, want)
		}
	}
}

func TestLoadCatalogMissingFSRootIsNotAnError(t *testing.T) {
	// A configured fs source whose directory does not exist yet must not fail the
	// whole load (a product may enable the source before authoring files).
	fsys := fstest.MapFS{}
	if _, err := i18n.LoadCatalog("en", frameworkLayer(), catalogsLayer(i18n.NewFSSource(fsys, "locales", "yaml"))); err != nil {
		t.Fatalf("missing fs root should be tolerated, got %v", err)
	}
}

func TestFreezeBlocksPostBootAdd(t *testing.T) {
	cat, err := i18n.LoadCatalog("en", frameworkLayer())
	if err != nil {
		t.Fatalf("LoadCatalog: %v", err)
	}
	cat.Freeze()
	if !cat.Frozen() {
		t.Fatal("catalog should report frozen")
	}
	before, _ := cat.Lookup("en", "app.late")
	cat.Add("en", "app.late", "should be ignored")
	after, _ := cat.Lookup("en", "app.late")
	if before != after {
		t.Fatalf("Add after Freeze mutated catalog: %q -> %q", before, after)
	}
}
