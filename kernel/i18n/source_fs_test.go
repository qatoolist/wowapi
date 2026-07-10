package i18n_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/qatoolist/wowapi/kernel/i18n"
)

func TestFSSourceKinds(t *testing.T) {
	if k := i18n.NewFSSource(fstest.MapFS{}, "locales").Kind(); k != i18n.KindFS {
		t.Fatalf("fs kind = %q", k)
	}
	if k := i18n.NewGoSource().Kind(); k != i18n.KindGo {
		t.Fatalf("go kind = %q", k)
	}
	if k := i18n.FrameworkDefaultsSource().Kind(); k != i18n.KindFrameworkDefaults {
		t.Fatalf("framework kind = %q", k)
	}
}

func TestFSSourceInvalidYAMLErrors(t *testing.T) {
	fsys := fstest.MapFS{"locales/en/bad.yaml": &fstest.MapFile{Data: []byte("messages: [not-a-map")}}
	if _, err := i18n.NewFSSource(fsys, "locales", "yaml").Load(); err == nil || !strings.Contains(err.Error(), "invalid YAML") {
		t.Fatalf("expected invalid YAML error, got %v", err)
	}
}

func TestFSSourceEmptyJSONTolerated(t *testing.T) {
	// An empty/whitespace-only JSON catalog is a placeholder, not a boot failure —
	// symmetric with an empty YAML file (review B1-corr #1).
	for _, body := range []string{"", "   \n", "{}"} {
		fsys := fstest.MapFS{"locales/en.json": &fstest.MapFile{Data: []byte(body)}}
		if _, err := i18n.NewFSSource(fsys, "locales", "json").Load(); err != nil {
			t.Fatalf("empty JSON %q should be tolerated, got %v", body, err)
		}
	}
}

func TestFSSourceInvalidJSONErrors(t *testing.T) {
	fsys := fstest.MapFS{"locales/en.json": &fstest.MapFile{Data: []byte(`{"messages": {`)}}
	if _, err := i18n.NewFSSource(fsys, "locales", "json").Load(); err == nil || !strings.Contains(err.Error(), "invalid JSON") {
		t.Fatalf("expected invalid JSON error, got %v", err)
	}
}

func TestFSSourceUnknownFieldRejected(t *testing.T) {
	fsys := fstest.MapFS{"locales/en/x.yaml": &fstest.MapFile{Data: []byte("locale: en\nmesages: {}\n")}} // typo
	if _, err := i18n.NewFSSource(fsys, "locales", "yaml").Load(); err == nil {
		t.Fatalf("unknown field should be rejected")
	}
}

func TestFSSourceLocaleFieldMismatch(t *testing.T) {
	fsys := fstest.MapFS{"locales/en/x.yaml": &fstest.MapFile{Data: []byte("locale: fr\nmessages:\n  app.x: y\n")}}
	if _, err := i18n.NewFSSource(fsys, "locales", "yaml").Load(); err == nil || !strings.Contains(err.Error(), "disagrees") {
		t.Fatalf("expected locale mismatch error, got %v", err)
	}
}

func TestFSSourceDefaultFormatsBothYAMLAndJSON(t *testing.T) {
	fsys := fstest.MapFS{
		"locales/en/a.yaml": &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  app.y: Y\n")},
		"locales/en.json":   &fstest.MapFile{Data: []byte(`{"messages":{"app.j":"J"}}`)},
	}
	cat, err := i18n.LoadCatalog("en", i18n.Layer{Name: "c", Sources: []i18n.Source{i18n.NewFSSource(fsys, "locales")}})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m, _ := cat.Lookup("en", "app.y"); m != "Y" {
		t.Fatalf("yaml key: %q", m)
	}
	if m, _ := cat.Lookup("en", "app.j"); m != "J" {
		t.Fatalf("json key: %q", m)
	}
}

func TestFSSourceIgnoresNonCatalogExtensions(t *testing.T) {
	fsys := fstest.MapFS{
		"locales/README.md": &fstest.MapFile{Data: []byte("# not a catalog")},
		"locales/en/x.yaml": &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  app.x: X\n")},
	}
	cat, err := i18n.LoadCatalog("en", i18n.Layer{Name: "c", Sources: []i18n.Source{i18n.NewFSSource(fsys, "locales", "yaml")}})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if m, _ := cat.Lookup("en", "app.x"); m != "X" {
		t.Fatalf("yaml key: %q", m)
	}
}

func TestGoSourceDefaultsOrigin(t *testing.T) {
	src := i18n.NewGoSource(i18n.RawBundle{Locale: "en", Messages: map[string]string{"app.g": "G"}})
	bundles, err := src.Load()
	if err != nil || len(bundles) != 1 || bundles[0].Origin == "" {
		t.Fatalf("go source load = %v, %v", bundles, err)
	}
}

func TestLoadCatalogBundleNoLocaleFails(t *testing.T) {
	// A Go bundle with no locale is caught by the loader.
	src := i18n.NewGoSource(i18n.RawBundle{Origin: "go", Messages: map[string]string{"app.x": "y"}})
	if _, err := i18n.LoadCatalog("en", i18n.Layer{Name: "go", Sources: []i18n.Source{src}}); err == nil ||
		!strings.Contains(err.Error(), "no locale") {
		t.Fatalf("expected no-locale error, got %v", err)
	}
}

func TestNilCatalogFreezeSafe(t *testing.T) {
	var c *i18n.Catalog
	c.Freeze() // must not panic
	if c.Frozen() {
		t.Fatal("nil catalog is not frozen")
	}
}
