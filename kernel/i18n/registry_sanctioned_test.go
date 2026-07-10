package i18n_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/i18n"
)

func TestRegisterFrameworkLocaleHappyPath(t *testing.T) {
	r := i18n.NewRegistry()
	r.RegisterFrameworkLocale(i18n.Bundle{
		Locale:   "mr",
		Messages: map[string]string{i18n.KeyProblemTitle(errors.KindNotFound): "सापडले नाही"},
	})
	if err := r.Err(); err != nil {
		t.Fatalf("valid framework-locale bundle rejected: %v", err)
	}
	if msg, loc := r.Catalog().Lookup("mr", i18n.KeyProblemTitle(errors.KindNotFound)); msg != "सापडले नाही" || loc != "mr" {
		t.Fatalf("mr framework title = (%q,%q)", msg, loc)
	}
}

func TestRegisterFrameworkLocaleRejectsNonKernelKey(t *testing.T) {
	r := i18n.NewRegistry()
	r.RegisterFrameworkLocale(i18n.Bundle{Locale: "mr", Messages: map[string]string{"app.x": "y"}})
	if err := r.Err(); err == nil || !strings.Contains(err.Error(), "reserved") {
		t.Fatalf("non-kernel key must be rejected, got %v", err)
	}
}

func TestRegisterFrameworkLocaleRejectsInventedKey(t *testing.T) {
	r := i18n.NewRegistry()
	r.RegisterFrameworkLocale(i18n.Bundle{Locale: "mr", Messages: map[string]string{"kernel.problem.made_up": "z"}})
	if err := r.Err(); err == nil || !strings.Contains(err.Error(), "not invent") {
		t.Fatalf("invented kernel key must be rejected, got %v", err)
	}
}

func TestApplyLayersOnRegistryOverridesFrameworkDefault(t *testing.T) {
	// Registry already carries framework defaults; a product override fs layer
	// retranslates a kernel.* key for mr while en still falls back.
	r := i18n.NewRegistry()
	fsys := fstest.MapFS{
		"locales/mr/kernel.yaml": &fstest.MapFile{Data: []byte(
			"locale: mr\nmessages:\n  " + i18n.KeyProblemTitle(errors.KindNotFound) + ": \"सापडले नाही\"\n",
		)},
		"locales/en/product.yaml": &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  app.hi: \"Hi\"\n")},
	}
	r.ApplyLayers(
		i18n.Layer{
			Name: "product overrides", Policy: i18n.Policy{OwnsFramework: true, FrameworkOverrideOnly: true},
			Sources: []i18n.Source{i18n.NewFSSource(fsys, "locales", "yaml")},
		},
	)
	if err := r.Err(); err != nil {
		t.Fatalf("ApplyLayers: %v", err)
	}
	cat := r.Catalog()
	if msg, _ := cat.Lookup("mr", i18n.KeyProblemTitle(errors.KindNotFound)); msg != "सापडले नाही" {
		t.Fatalf("mr override failed: %q", msg)
	}
	if msg, _ := cat.Lookup("en", i18n.KeyProblemTitle(errors.KindNotFound)); msg != "Not found" {
		t.Fatalf("en should fall back to embedded default: %q", msg)
	}
	if msg, _ := cat.Lookup("en", "app.hi"); msg != "Hi" {
		t.Fatalf("product key not loaded: %q", msg)
	}
}

func TestApplyLayersRecordsOwnershipError(t *testing.T) {
	r := i18n.NewRegistry()
	fsys := fstest.MapFS{
		"locales/en/x.yaml": &fstest.MapFile{Data: []byte(
			"locale: en\nmessages:\n  " + i18n.KeyProblemTitle(errors.KindNotFound) + ": \"hijack\"\n",
		)},
	}
	// A plain catalog layer (no OwnsFramework) may not write kernel.*.
	r.ApplyLayers(i18n.Layer{Name: "catalogs", Sources: []i18n.Source{i18n.NewFSSource(fsys, "locales", "yaml")}})
	if err := r.Err(); err == nil || !strings.Contains(err.Error(), "reserved framework key") {
		t.Fatalf("expected ownership error, got %v", err)
	}
}
