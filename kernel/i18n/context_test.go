package i18n_test

import (
	"context"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/i18n"
)

func TestContextRoundTrip(t *testing.T) {
	cat := i18n.NewCatalog("en")
	cat.Add("mr", "k", "v")
	ctx := i18n.WithContext(context.Background(), "mr", cat)

	if got := i18n.LocaleFrom(ctx); got != "mr" {
		t.Errorf("LocaleFrom = %q, want mr", got)
	}
	if got := i18n.CatalogFrom(ctx); got != cat {
		t.Errorf("CatalogFrom = %v, want the bound catalog", got)
	}
}

func TestContextEmpty(t *testing.T) {
	ctx := context.Background()
	if got := i18n.LocaleFrom(ctx); got != "" {
		t.Errorf("unbound LocaleFrom = %q, want empty", got)
	}
	if got := i18n.CatalogFrom(ctx); got != nil {
		t.Errorf("unbound CatalogFrom = %v, want nil", got)
	}
}

func TestNilCatalogDefaultAndLocales(t *testing.T) {
	var c *i18n.Catalog
	if c.Default() != "" {
		t.Errorf("nil Default = %q, want empty", c.Default())
	}
	if c.Locales() != nil {
		t.Errorf("nil Locales = %v, want nil", c.Locales())
	}
}

func TestRegistryRejectsBundleWithNoLocale(t *testing.T) {
	r := i18n.NewRegistry()
	r.Register("orders", i18n.Bundle{Messages: map[string]string{"orders.k": "v"}})
	if err := r.Err(); err == nil {
		t.Fatal("bundle with no locale must be rejected")
	}
}
