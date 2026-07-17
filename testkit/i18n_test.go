package testkit_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/errors"
	"github.com/qatoolist/wowapi/v2/kernel/httpx"
	"github.com/qatoolist/wowapi/v2/kernel/i18n"
	"github.com/qatoolist/wowapi/v2/testkit"
)

func testCat(t *testing.T) *i18n.Catalog {
	t.Helper()
	cat := i18n.NewRegistry().Catalog()
	cat.Add("mr", i18n.KeyProblemTitle(errors.KindNotFound), "सापडले नाही")
	return cat
}

func TestAssertNegotiatedLocale(t *testing.T) {
	cat := testCat(t)
	// Passing case: mr negotiates to mr.
	testkit.AssertNegotiatedLocale(t, cat, "mr-IN,mr;q=0.9,en;q=0.8", "mr")
	// Fallback case.
	testkit.AssertNegotiatedLocale(t, cat, "fr-FR", "en")
}

func TestNewLocaleRequestBindsLocale(t *testing.T) {
	cat := testCat(t)
	r := testkit.NewLocaleRequest(http.MethodGet, "/x", "mr", cat)
	if got := httpx.LocaleFrom(r.Context()); got != "mr" {
		t.Fatalf("bound locale = %q, want mr", got)
	}
}

func TestAssertLocalizedProblem(t *testing.T) {
	cat := testCat(t)
	ctx := i18n.WithContext(context.Background(), "mr", cat)
	rec := httptest.NewRecorder()
	httpx.WriteError(ctx, rec, errors.E(errors.KindNotFound, "not_found", "gone"))

	// Asserts the localized title AND the stable machine code at once.
	testkit.AssertLocalizedProblem(t, rec.Result(), "not_found", "सापडले नाही")
}
