package i18n_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/i18n"
)

func TestBuildLayersCanonicalPrecedence(t *testing.T) {
	root := fstest.MapFS{
		"locales/mr/kernel.yaml":  &fstest.MapFile{Data: []byte("locale: mr\nmessages:\n  " + i18n.KeyProblemTitle(errors.KindNotFound) + ": \"सापडले नाही\"\n")},
		"locales/en/product.yaml": &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  app.hi: \"Hi\"\n")},
	}
	specs := []i18n.SourceSpec{
		{Kind: i18n.KindFrameworkDefaults, Enabled: true},
		// The canonical single fs source over locales/: overrides_framework lets its
		// kernel.* files retranslate framework strings, while its product/module
		// keys pass freely (non-kernel keys are never namespace-guarded).
		{Kind: i18n.KindFS, Path: "locales", Formats: []string{"yaml", "json"}, OverridesFramework: true, Enabled: true},
		{Kind: i18n.KindGo, Enabled: true, Go: []i18n.RawBundle{{Locale: "en", Origin: "go", Messages: map[string]string{"app.g": "G"}}}},
		{Kind: i18n.KindGo, Enabled: false, Go: []i18n.RawBundle{{Locale: "en", Messages: map[string]string{"app.off": "X"}}}},
	}
	layers, err := i18n.BuildLayers(root, specs)
	if err != nil {
		t.Fatalf("BuildLayers: %v", err)
	}
	// framework-defaults spec produces no layer (registry installs it); override
	// fs + go => 2 layers, in order.
	if len(layers) != 2 {
		t.Fatalf("got %d layers, want 2: %+v", len(layers), layers)
	}
	r := i18n.NewRegistry()
	r.ApplyLayers(layers...)
	if err := r.Err(); err != nil {
		t.Fatalf("ApplyLayers: %v", err)
	}
	cat := r.Catalog()
	if m, _ := cat.Lookup("mr", i18n.KeyProblemTitle(errors.KindNotFound)); m != "सापडले नाही" {
		t.Errorf("mr override: %q", m)
	}
	if m, _ := cat.Lookup("en", "app.hi"); m != "Hi" {
		t.Errorf("product key missing: %q", m)
	}
	if m, _ := cat.Lookup("en", "app.g"); m != "G" {
		t.Errorf("go bundle key: %q", m)
	}
	if m, _ := cat.Lookup("en", "app.off"); m != "app.off" {
		t.Errorf("disabled source must contribute nothing, got %q", m)
	}
}

func TestBuildLayersRejectsUnknownKind(t *testing.T) {
	_, err := i18n.BuildLayers(fstest.MapFS{}, []i18n.SourceSpec{{Kind: "bogus", Enabled: true}})
	if err == nil || !strings.Contains(err.Error(), "unknown i18n source kind") {
		t.Fatalf("expected unknown-kind rejection, got %v", err)
	}
}

func TestBuildLayersEmptyWhenAllDisabled(t *testing.T) {
	layers, err := i18n.BuildLayers(fstest.MapFS{}, []i18n.SourceSpec{
		{Kind: i18n.KindFS, Path: "locales", Enabled: false},
	})
	if err != nil || len(layers) != 0 {
		t.Fatalf("all-disabled = %v layers, %v", len(layers), err)
	}
}
