package i18n_test

import (
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/i18n"
)

// ---------- Catalog: lookup + deterministic fallback ----------

func TestCatalogLookupExactLocale(t *testing.T) {
	c := i18n.NewCatalog("en")
	c.Add("en", "greeting", "Hello")
	c.Add("mr", "greeting", "नमस्कार")

	msg, loc := c.Lookup("mr", "greeting")
	if msg != "नमस्कार" || loc != "mr" {
		t.Fatalf("Lookup(mr) = (%q,%q), want (नमस्कार,mr)", msg, loc)
	}
}

func TestCatalogFallsBackToDefaultLocale(t *testing.T) {
	c := i18n.NewCatalog("en")
	c.Add("en", "greeting", "Hello")
	// mr has no "greeting" -> falls back to en deterministically.
	msg, loc := c.Lookup("mr", "greeting")
	if msg != "Hello" || loc != "en" {
		t.Fatalf("Lookup(mr) with only en = (%q,%q), want (Hello,en)", msg, loc)
	}
}

func TestCatalogUnknownKeyReturnsKey(t *testing.T) {
	c := i18n.NewCatalog("en")
	// Missing even in default: never error, echo the key so a response never breaks.
	msg, loc := c.Lookup("mr", "does.not.exist")
	if msg != "does.not.exist" || loc != "en" {
		t.Fatalf("Lookup(missing) = (%q,%q), want (does.not.exist,en)", msg, loc)
	}
}

func TestCatalogEmptyLocaleUsesDefault(t *testing.T) {
	c := i18n.NewCatalog("en")
	c.Add("en", "k", "v")
	msg, loc := c.Lookup("", "k")
	if msg != "v" || loc != "en" {
		t.Fatalf("Lookup(\"\") = (%q,%q), want (v,en)", msg, loc)
	}
}

func TestCatalogLocalesAndSupported(t *testing.T) {
	c := i18n.NewCatalog("en")
	c.Add("en", "k", "v")
	c.Add("mr", "k", "w")
	locs := c.Locales()
	if len(locs) != 2 {
		t.Fatalf("Locales() = %v, want 2 entries", locs)
	}
	// Supports reports whether a locale has any registered message.
	if !c.Supports("en") || !c.Supports("mr") || c.Supports("fr") {
		t.Fatalf("Supports wrong: en/mr should be true, fr false")
	}
	if c.Default() != "en" {
		t.Fatalf("Default() = %q, want en", c.Default())
	}
}

func TestNilCatalogLookupIsSafe(t *testing.T) {
	// A nil *Catalog behaves as an empty catalog (zero-config path): echo the key
	// with an empty resolved locale, never panic.
	var c *i18n.Catalog
	msg, loc := c.Lookup("mr", "k")
	if msg != "k" || loc != "" {
		t.Fatalf("nil Lookup = (%q,%q), want (k,\"\")", msg, loc)
	}
	if c.Supports("en") {
		t.Fatalf("nil Supports must be false")
	}
}

// ---------- Negotiate: RFC 9110 Accept-Language q-values ----------

func TestNegotiateAcceptanceProof(t *testing.T) {
	supported := []string{"en", "mr"}
	got := i18n.Negotiate("mr-IN,mr;q=0.9,en;q=0.8", supported, "en")
	if got != "mr" {
		t.Fatalf("Negotiate(mr-IN,...) = %q, want mr", got)
	}
}

func TestNegotiateFallsBackDeterministically(t *testing.T) {
	supported := []string{"en", "mr"}
	cases := map[string]string{
		"":                    "en", // empty header
		"fr-FR,fr;q=0.9":      "en", // nothing supported
		"*":                   "en", // wildcard is no preference
		"de;q=0,en;q=0.5":     "en", // de explicitly refused (q=0)
		"garbage;q=notafloat": "en", // malformed q skipped
	}
	for header, want := range cases {
		if got := i18n.Negotiate(header, supported, "en"); got != want {
			t.Errorf("Negotiate(%q) = %q, want %q", header, got, want)
		}
	}
}

func TestNegotiateQValueOrdering(t *testing.T) {
	supported := []string{"en", "mr"}
	// mr has higher q than en despite appearing second.
	if got := i18n.Negotiate("en;q=0.3,mr;q=0.9", supported, "en"); got != "mr" {
		t.Fatalf("higher-q mr should win, got %q", got)
	}
	// Equal q keeps header order (leftmost wins).
	if got := i18n.Negotiate("en,mr", supported, "en"); got != "en" {
		t.Fatalf("equal-q should keep header order (en), got %q", got)
	}
}

func TestNegotiatePrimarySubtagMatch(t *testing.T) {
	// A supported "mr" matches an offered "mr-IN" on the primary subtag.
	if got := i18n.Negotiate("mr-IN", []string{"en", "mr"}, "en"); got != "mr" {
		t.Fatalf("mr-IN should match supported mr, got %q", got)
	}
}

func TestNegotiateCaseInsensitiveAndUppercaseQ(t *testing.T) {
	supported := []string{"en", "mr"}
	// Offered tag case is normalized.
	if got := i18n.Negotiate("MR-IN", supported, "en"); got != "mr" {
		t.Errorf("MR-IN should match mr, got %q", got)
	}
	// Uppercase Q= is honored.
	if got := i18n.Negotiate("en;Q=0.1,mr;Q=0.9", supported, "en"); got != "mr" {
		t.Errorf("uppercase Q= should be parsed, got %q", got)
	}
}

// ---------- Registry: module-prefixed bundle registration ----------

func TestRegistryFrameworkBundleAlwaysPresent(t *testing.T) {
	r := i18n.NewRegistry()
	cat := r.Catalog()
	// The framework catalog ships problem titles + validation messages in English.
	title, loc := cat.Lookup("en", i18n.KeyProblemTitle(errors.KindNotFound))
	if title == "" || loc != "en" {
		t.Fatalf("framework problem title missing for not_found: (%q,%q)", title, loc)
	}
	if !cat.Supports("en") {
		t.Fatalf("framework catalog must support en")
	}
}

func TestRegistryModuleBundleMustBePrefixed(t *testing.T) {
	r := i18n.NewRegistry()
	// A module may only register keys under its own "<module>." prefix.
	r.Register("orders", i18n.Bundle{
		Locale:   "en",
		Messages: map[string]string{"orders.msg.hello": "Hi"},
	})
	if err := r.Err(); err != nil {
		t.Fatalf("valid module bundle rejected: %v", err)
	}
	r.Register("orders", i18n.Bundle{
		Locale:   "en",
		Messages: map[string]string{"billing.msg.x": "nope"},
	})
	if err := r.Err(); err == nil || !strings.Contains(err.Error(), "orders") {
		t.Fatalf("cross-module key should be rejected, got %v", err)
	}
}

func TestRegistryMergesLocales(t *testing.T) {
	r := i18n.NewRegistry()
	r.Register("orders", i18n.Bundle{Locale: "en", Messages: map[string]string{"orders.msg.hi": "Hi"}})
	r.Register("orders", i18n.Bundle{Locale: "mr", Messages: map[string]string{"orders.msg.hi": "नमस्कार"}})
	if err := r.Err(); err != nil {
		t.Fatalf("Err = %v", err)
	}
	cat := r.Catalog()
	if msg, _ := cat.Lookup("mr", "orders.msg.hi"); msg != "नमस्कार" {
		t.Fatalf("mr merge failed: %q", msg)
	}
	if msg, _ := cat.Lookup("en", "orders.msg.hi"); msg != "Hi" {
		t.Fatalf("en merge failed: %q", msg)
	}
}

func TestRegistryRejectsFrameworkPrefixFromModules(t *testing.T) {
	r := i18n.NewRegistry()
	// Modules must not shadow reserved framework keys (kernel.* namespace).
	r.Register("kernel", i18n.Bundle{Locale: "en", Messages: map[string]string{"kernel.problem.not_found": "x"}})
	if err := r.Err(); err == nil {
		t.Fatalf("module registering reserved kernel.* prefix must be rejected")
	}
}
