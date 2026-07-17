package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/i18n"
)

func writeLocaleFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestI18nValidateCleanProduct(t *testing.T) {
	dir := t.TempDir()
	writeLocaleFile(t, dir, "en/product.yaml", "locale: en\nmessages:\n  app.hi: Hi\n")
	var out, errb bytes.Buffer
	code := runI18n([]string{"validate", "--dir", dir}, &out, &errb)
	if code != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s", code, errb.String())
	}
	if !strings.Contains(out.String(), "OK:") {
		t.Fatalf("expected OK, got %q", out.String())
	}
}

func TestI18nValidateFailsOnDuplicate(t *testing.T) {
	dir := t.TempDir()
	writeLocaleFile(t, dir, "en/a.yaml", "locale: en\nmessages:\n  app.hi: one\n")
	writeLocaleFile(t, dir, "en/b.yaml", "locale: en\nmessages:\n  app.hi: two\n")
	var out, errb bytes.Buffer
	code := runI18n([]string{"validate", "--dir", dir}, &out, &errb)
	if code != 1 || !strings.Contains(errb.String(), "duplicate key") {
		t.Fatalf("expected duplicate failure, code=%d stderr=%s", code, errb.String())
	}
}

func TestI18nValidateFailsOnOwnership(t *testing.T) {
	dir := t.TempDir()
	// A brand-new kernel.* key the framework does not define — override layer may
	// not invent it.
	writeLocaleFile(t, dir, "en/kernel.yaml", "locale: en\nmessages:\n  kernel.problem.made_up: nope\n")
	var out, errb bytes.Buffer
	code := runI18n([]string{"validate", "--dir", dir}, &out, &errb)
	if code != 1 || !strings.Contains(errb.String(), "not invent") && !strings.Contains(errb.String(), "no lower layer") {
		t.Fatalf("expected ownership failure, code=%d stderr=%s", code, errb.String())
	}
}

func TestI18nValidateFailsOnMissingCoverage(t *testing.T) {
	dir := t.TempDir()
	// Key only in mr, absent from default en -> no fallback.
	writeLocaleFile(t, dir, "mr/product.yaml", "locale: mr\nmessages:\n  app.only_mr: फक्त\n")
	var out, errb bytes.Buffer
	code := runI18n([]string{"validate", "--dir", dir, "--supported", "en,mr"}, &out, &errb)
	if code != 1 || !strings.Contains(errb.String(), "missing from the default locale") {
		t.Fatalf("expected coverage failure, code=%d stderr=%s", code, errb.String())
	}
}

func TestI18nValidateFailsOnPlaceholderDrift(t *testing.T) {
	dir := t.TempDir()
	// Retranslate framework min message for mr but drop its %s.
	writeLocaleFile(t, dir, "mr/kernel.yaml",
		"locale: mr\nmessages:\n  "+i18n.KeyValidationMessage("min")+": \"किमान आवश्यक\"\n")
	var out, errb bytes.Buffer
	code := runI18n([]string{"validate", "--dir", dir, "--supported", "en,mr"}, &out, &errb)
	if code != 1 || !strings.Contains(errb.String(), "placeholder mismatch") {
		t.Fatalf("expected placeholder failure, code=%d stderr=%s", code, errb.String())
	}
}

func TestI18nValidateStrictCoverage(t *testing.T) {
	dir := t.TempDir()
	writeLocaleFile(t, dir, "en/product.yaml", "locale: en\nmessages:\n  app.hi: Hi\n")
	writeLocaleFile(t, dir, "mr/product.yaml", "locale: mr\nmessages:\n  app.hi: नमस्कार\n  app.extra: extra\n")
	// app.extra only in mr -> hard miss; add to en to isolate strict gap on app.hi.
	writeLocaleFile(t, dir, "en/product.yaml", "locale: en\nmessages:\n  app.hi: Hi\n  app.extra: Extra\n")
	writeLocaleFile(t, dir, "mr/product.yaml", "locale: mr\nmessages:\n  app.extra: extra\n") // mr missing app.hi

	var out, errb bytes.Buffer
	// non-strict passes (fallback), strict fails.
	if code := runI18n([]string{"validate", "--dir", dir, "--supported", "en,mr"}, &out, &errb); code != 0 {
		t.Fatalf("non-strict should pass, code=%d stderr=%s", code, errb.String())
	}
	out.Reset()
	errb.Reset()
	if code := runI18n([]string{"validate", "--dir", dir, "--supported", "en,mr", "--strict"}, &out, &errb); code != 1 {
		t.Fatalf("strict should fail on mr missing app.hi, code=%d stderr=%s", code, errb.String())
	}
}

func TestI18nValidateBadDir(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runI18n([]string{"validate", "--dir", filepath.Join(t.TempDir(), "nope")}, &out, &errb); code != 1 {
		t.Fatalf("expected exit 1 for missing dir, got %d", code)
	}
}

func TestI18nUsageAndUnknown(t *testing.T) {
	var out, errb bytes.Buffer
	if code := runI18n(nil, &out, &errb); code != 2 {
		t.Fatalf("no args should be usage exit 2, got %d", code)
	}
	if code := runI18n([]string{"help"}, &out, &errb); code != 0 {
		t.Fatalf("help exit 0, got %d", code)
	}
	if code := runI18n([]string{"bogus"}, &out, &errb); code != 2 {
		t.Fatalf("unknown subcommand exit 2, got %d", code)
	}
	_ = errors.KindNotFound
}
