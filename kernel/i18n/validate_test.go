package i18n_test

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/i18n"
)

func fwLayer() i18n.Layer {
	return i18n.Layer{Name: "framework defaults", Policy: i18n.Policy{OwnsFramework: true}, Sources: []i18n.Source{i18n.FrameworkDefaultsSource()}}
}

func fsOverride(fsys fstest.MapFS) i18n.Layer {
	return i18n.Layer{Name: "product", Policy: i18n.Policy{OwnsFramework: true, FrameworkOverrideOnly: true}, Sources: []i18n.Source{i18n.NewFSSource(fsys, "locales", "yaml", "json")}}
}

func TestValidateCleanFrameworkDefaults(t *testing.T) {
	rep, err := i18n.Validate(i18n.ValidateOptions{DefaultLocale: "en", SupportedLocales: []string{"en"}}, fwLayer())
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if !rep.OK() {
		t.Fatalf("framework defaults should validate clean, got: %v", rep.Problems)
	}
	if rep.Keys == 0 {
		t.Fatal("expected framework keys counted")
	}
}

func TestValidateFailsOnDuplicateInLayer(t *testing.T) {
	fsys := fstest.MapFS{
		"locales/en/a.yaml": &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  app.hi: one\n")},
		"locales/en/b.yaml": &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  app.hi: two\n")},
	}
	rep, _ := i18n.Validate(i18n.ValidateOptions{DefaultLocale: "en"},
		i18n.Layer{Name: "product", Sources: []i18n.Source{i18n.NewFSSource(fsys, "locales", "yaml")}})
	if rep.OK() || !containsSub(rep.Problems, "duplicate key") {
		t.Fatalf("expected duplicate problem, got %v", rep.Problems)
	}
}

func TestValidateFailsOnOwnershipViolation(t *testing.T) {
	fsys := fstest.MapFS{
		"locales/en/x.yaml": &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  " + i18n.KeyProblemTitle(errors.KindNotFound) + ": hijack\n")},
	}
	rep, _ := i18n.Validate(i18n.ValidateOptions{DefaultLocale: "en"},
		fwLayer(),
		i18n.Layer{Name: "product", Sources: []i18n.Source{i18n.NewFSSource(fsys, "locales", "yaml")}})
	if rep.OK() || !containsSub(rep.Problems, "reserved framework key") {
		t.Fatalf("expected ownership problem, got %v", rep.Problems)
	}
}

func TestValidateFailsOnMissingDefaultLocaleCoverage(t *testing.T) {
	// A key defined only for mr (not en) has no fallback.
	fsys := fstest.MapFS{
		"locales/mr/product.yaml": &fstest.MapFile{Data: []byte("locale: mr\nmessages:\n  app.only_mr: फक्त\n")},
	}
	rep, _ := i18n.Validate(i18n.ValidateOptions{DefaultLocale: "en"}, fwLayer(), fsOverride(fsys))
	if rep.OK() || !containsSub(rep.Problems, "missing from the default locale") {
		t.Fatalf("expected default-coverage problem, got %v", rep.Problems)
	}
}

func TestValidateStrictCoverageGap(t *testing.T) {
	// app.hi exists in en but not mr; non-strict OK (fallback), strict fails.
	fsys := fstest.MapFS{
		"locales/en/product.yaml": &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  app.hi: Hi\n")},
		"locales/mr/product.yaml": &fstest.MapFile{Data: []byte("locale: mr\nmessages:\n  app.bye: निरोप\n")},
	}
	// app.bye only in mr -> hard miss regardless. Add app.bye to en to isolate the gap test.
	fsys["locales/en/product.yaml"] = &fstest.MapFile{Data: []byte("locale: en\nmessages:\n  app.hi: Hi\n  app.bye: Bye\n")}

	nonStrict, _ := i18n.Validate(i18n.ValidateOptions{DefaultLocale: "en", SupportedLocales: []string{"en", "mr"}}, fwLayer(), fsOverride(fsys))
	if !nonStrict.OK() {
		t.Fatalf("non-strict should tolerate fallback gap, got %v", nonStrict.Problems)
	}
	strict, _ := i18n.Validate(i18n.ValidateOptions{DefaultLocale: "en", SupportedLocales: []string{"en", "mr"}, StrictCoverage: true}, fwLayer(), fsOverride(fsys))
	if strict.OK() || !containsSub(strict.Problems, "is missing key") {
		t.Fatalf("strict should flag mr missing app.hi, got %v", strict.Problems)
	}
}

func TestValidatePlaceholderMismatch(t *testing.T) {
	// Override the framework "min" message for mr but drop its %s.
	fsys := fstest.MapFS{
		"locales/mr/kernel.yaml": &fstest.MapFile{Data: []byte("locale: mr\nmessages:\n  " + i18n.KeyValidationMessage("min") + ": \"किमान असणे आवश्यक\"\n")},
	}
	rep, _ := i18n.Validate(i18n.ValidateOptions{DefaultLocale: "en"}, fwLayer(), fsOverride(fsys))
	if rep.OK() || !containsSub(rep.Problems, "placeholder mismatch") {
		t.Fatalf("expected placeholder mismatch, got %v", rep.Problems)
	}
}

func TestValidatePlaceholderMatchIsClean(t *testing.T) {
	fsys := fstest.MapFS{
		"locales/mr/kernel.yaml": &fstest.MapFile{Data: []byte("locale: mr\nmessages:\n  " + i18n.KeyValidationMessage("min") + ": \"किमान %s असणे आवश्यक\"\n")},
	}
	rep, _ := i18n.Validate(i18n.ValidateOptions{DefaultLocale: "en", SupportedLocales: []string{"en", "mr"}}, fwLayer(), fsOverride(fsys))
	if !rep.OK() {
		t.Fatalf("matching placeholder should be clean, got %v", rep.Problems)
	}
}

func containsSub(problems []string, sub string) bool {
	for _, p := range problems {
		if strings.Contains(p, sub) {
			return true
		}
	}
	return false
}
