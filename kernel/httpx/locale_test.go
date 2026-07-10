package httpx_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/i18n"
	"github.com/qatoolist/wowapi/kernel/validation"
)

// buildTestCatalog returns a catalog with a test-only Marathi bundle proving
// localization without baking product translations into the framework.
func buildTestCatalog(t *testing.T) *i18n.Catalog {
	t.Helper()
	reg := i18n.NewRegistry()
	// Add mr translations for the framework's own problem title + validation
	// message by writing directly to the merged catalog (the framework namespace
	// is reserved for the framework itself, which is exactly this call site's role
	// in a test).
	cat := reg.Catalog()
	cat.Add("mr", i18n.KeyProblemTitle(errors.KindNotFound), "सापडले नाही")
	cat.Add("mr", i18n.KeyValidationMessage("required"), "हे फील्ड आवश्यक आहे")
	return cat
}

func TestLocaleMiddlewareNegotiatesAndSetsContentLanguage(t *testing.T) {
	cat := buildTestCatalog(t)
	var gotLocale string
	h := httpx.Locale(cat)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLocale = httpx.LocaleFrom(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "mr-IN,mr;q=0.9,en;q=0.8")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if gotLocale != "mr" {
		t.Fatalf("negotiated locale = %q, want mr", gotLocale)
	}
	if cl := rec.Header().Get("Content-Language"); cl != "mr" {
		t.Fatalf("Content-Language = %q, want mr", cl)
	}
}

func TestLocaleMiddlewareUnsupportedFallsBackToDefault(t *testing.T) {
	cat := buildTestCatalog(t)
	var gotLocale string
	h := httpx.Locale(cat)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotLocale = httpx.LocaleFrom(r.Context())
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "fr-FR,fr;q=0.9")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if gotLocale != "en" {
		t.Fatalf("unsupported locale should fall back to en, got %q", gotLocale)
	}
	if cl := rec.Header().Get("Content-Language"); cl != "en" {
		t.Fatalf("Content-Language = %q, want en", cl)
	}
}

func TestLocaleMiddlewareNilCatalogIsNoOp(t *testing.T) {
	// Zero-config: no catalog wired => no locale bound, no Content-Language,
	// behavior identical to today.
	called := false
	h := httpx.Locale(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if loc := httpx.LocaleFrom(r.Context()); loc != "" {
			t.Errorf("nil catalog should bind no locale, got %q", loc)
		}
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Accept-Language", "mr")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if !called {
		t.Fatal("handler not called")
	}
	if cl := rec.Header().Get("Content-Language"); cl != "" {
		t.Errorf("nil catalog should set no Content-Language, got %q", cl)
	}
}

// ---------- localized WriteError ----------

func TestWriteErrorLocalizesTitleWithStableCode(t *testing.T) {
	cat := buildTestCatalog(t)
	ctx := httpx.WithLocale(context.Background(), "mr", cat)
	ctx = httpx.WithRequestID(ctx, "req-x")

	rec := httptest.NewRecorder()
	httpx.WriteError(ctx, rec, errors.E(errors.KindNotFound, "not_found", "gone"))

	p := decodeProblem(t, rec.Body.Bytes())
	if p.Title != "सापडले नाही" {
		t.Errorf("title not localized: %q", p.Title)
	}
	if p.Code != "not_found" { // machine code stays byte-stable
		t.Errorf("code changed: %q, want not_found", p.Code)
	}
	if p.Status != 404 {
		t.Errorf("status = %d", p.Status)
	}
}

func TestWriteErrorEnglishUnchangedWhenNoCatalog(t *testing.T) {
	// Zero-config path: no locale/catalog in ctx => historical English title.
	rec := httptest.NewRecorder()
	httpx.WriteError(context.Background(), rec, errors.E(errors.KindNotFound, "not_found", "gone"))
	p := decodeProblem(t, rec.Body.Bytes())
	if p.Title != "Not found" {
		t.Errorf("English title changed: %q, want 'Not found'", p.Title)
	}
	if p.Code != "not_found" {
		t.Errorf("code = %q", p.Code)
	}
}

// TestAcceptanceProofEndToEnd is the GAP-001 acceptance proof: a request with
// Accept-Language "mr-IN,mr;q=0.9,en;q=0.8" flows through the real Locale
// middleware and produces a validation problem whose title AND field message are
// Marathi, while the machine code and field path stay byte-stable — using a
// test-only Marathi bundle (no product translations in the framework).
func TestAcceptanceProofEndToEnd(t *testing.T) {
	cat := i18n.NewRegistry().Catalog()
	cat.Add("mr", i18n.KeyProblemTitle(errors.KindValidation), "प्रमाणीकरण अयशस्वी")
	cat.Add("mr", i18n.KeyValidationMessage("required"), "हे फील्ड आवश्यक आहे")

	v := validation.New()
	type body struct {
		Name string `json:"name" validate:"required"`
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := httpx.BindAndValidate[body](r, v, 1<<20)
		if err != nil {
			httpx.WriteError(r.Context(), w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	h := httpx.Chain(handler, httpx.RequestID(), httpx.Locale(cat))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":""}`))
	req.Header.Set("Accept-Language", "mr-IN,mr;q=0.9,en;q=0.8")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != 400 {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if cl := rec.Header().Get("Content-Language"); cl != "mr" {
		t.Errorf("Content-Language = %q, want mr", cl)
	}
	p := decodeProblem(t, rec.Body.Bytes())
	if p.Title != "प्रमाणीकरण अयशस्वी" {
		t.Errorf("title not Marathi: %q", p.Title)
	}
	if p.Code != "validation_failed" {
		t.Errorf("machine code changed: %q", p.Code)
	}
	if len(p.Errors) != 1 || p.Errors[0].Field != "name" || p.Errors[0].Code != "required" {
		t.Fatalf("field error field/code not stable: %+v", p.Errors)
	}
	if p.Errors[0].Message != "हे फील्ड आवश्यक आहे" {
		t.Errorf("field message not Marathi: %q", p.Errors[0].Message)
	}
}

// TestWriteErrorValidationDetailLocalizesViaShippedEntry proves the framework's
// own shipped `detail.validation_failed` English catalog entry localizes the
// validation error's top-level Detail (not just field messages), while Code and
// field Code/Field stay byte-stable — closing GAP-001's Detail gap. The
// `cat.Add(...)` calls here are the exact mechanism the user guide's
// "Translating the framework's own strings" section documents for adding a
// kernel.* translation (docs/user-guide/validation-errors.md, Localizing
// responses (i18n)): direct Catalog.Add on the booted catalog, not
// module.Context.I18n/Register (which rejects the reserved kernel.* prefix).
func TestWriteErrorValidationDetailLocalizesViaShippedEntry(t *testing.T) {
	cat := i18n.NewRegistry().Catalog()
	cat.Add("mr", i18n.KeyValidationMessage("required"), "हे फील्ड आवश्यक आहे")
	cat.Add("mr", i18n.KeyDetail("validation_failed"), "प्रमाणीकरण अयशस्वी झाले")

	v := validation.New()
	type body struct {
		Name string `json:"name" validate:"required"`
	}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := httpx.BindAndValidate[body](r, v, 1<<20)
		if err != nil {
			httpx.WriteError(r.Context(), w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	h := httpx.Chain(handler, httpx.RequestID(), httpx.Locale(cat))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":""}`))
	req.Header.Set("Accept-Language", "mr")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	p := decodeProblem(t, rec.Body.Bytes())
	if p.Detail != "प्रमाणीकरण अयशस्वी झाले" {
		t.Errorf("validation detail not localized: %q", p.Detail)
	}
	if p.Code != "validation_failed" {
		t.Errorf("machine code changed: %q", p.Code)
	}
	if len(p.Errors) != 1 || p.Errors[0].Field != "name" || p.Errors[0].Code != "required" {
		t.Fatalf("field error field/code not stable: %+v", p.Errors)
	}
}

// ---------- localized Detail ----------

func TestWriteErrorLocalizesDetailWithStableCode(t *testing.T) {
	cat := buildTestCatalog(t)
	// Test-bundle detail.<code> entry in the second locale (mirrors the title
	// test-bundle pattern above).
	cat.Add("mr", i18n.KeyDetail("not_found"), "आढळले नाही")
	ctx := httpx.WithLocale(context.Background(), "mr", cat)

	rec := httptest.NewRecorder()
	httpx.WriteError(ctx, rec, errors.E(errors.KindNotFound, "not_found", "gone"))

	p := decodeProblem(t, rec.Body.Bytes())
	if p.Detail != "आढळले नाही" {
		t.Errorf("detail not localized: %q", p.Detail)
	}
	if p.Code != "not_found" { // machine code stays byte-stable
		t.Errorf("code changed: %q, want not_found", p.Code)
	}
}

func TestWriteErrorDetailFallsBackToMsgWhenNoCatalogEntry(t *testing.T) {
	cat := buildTestCatalog(t) // has no detail.not_found entry
	ctx := httpx.WithLocale(context.Background(), "mr", cat)

	rec := httptest.NewRecorder()
	httpx.WriteError(ctx, rec, errors.E(errors.KindNotFound, "not_found", "gone"))

	p := decodeProblem(t, rec.Body.Bytes())
	if p.Detail != "gone" {
		t.Errorf("detail should fall back to producer Msg byte-identically, got %q", p.Detail)
	}
}

func TestWriteErrorInternalKindStillExposesNoDetail(t *testing.T) {
	cat := buildTestCatalog(t)
	cat.Add("mr", i18n.KeyDetail("internal"), "should never be used")
	ctx := httpx.WithLocale(context.Background(), "mr", cat)

	rec := httptest.NewRecorder()
	httpx.WriteError(ctx, rec, errors.E(errors.KindInternal, "internal", "sensitive cause"))

	p := decodeProblem(t, rec.Body.Bytes())
	if p.Detail != "" {
		t.Errorf("internal-kind Detail must stay empty, got %q", p.Detail)
	}
}

func TestWriteErrorUnknownLocaleFallsBackToEnglishTitle(t *testing.T) {
	cat := buildTestCatalog(t)
	// A locale with no translation for this key falls back to en deterministically.
	ctx := httpx.WithLocale(context.Background(), "mr", cat)
	rec := httptest.NewRecorder()
	// KindConflict has no mr translation in the test catalog.
	httpx.WriteError(ctx, rec, errors.E(errors.KindConflict, "conflict", "dup"))
	p := decodeProblem(t, rec.Body.Bytes())
	if p.Title != "Conflict" {
		t.Errorf("missing-mr title should fall back to English, got %q", p.Title)
	}
}

// TestWriteErrorKindAbsentFromCatalogNeverLeaksKey guards the divergence risk:
// with a catalog WIRED, a Kind absent from the framework catalog
// (KindIdempotencyExpired) must NOT leak the raw catalog key as the title — it
// must render exactly like the zero-config path ("Internal error" via the
// empty-title fallback), keeping the two paths in lockstep.
func TestWriteErrorKindAbsentFromCatalogNeverLeaksKey(t *testing.T) {
	cat := buildTestCatalog(t)
	ctx := httpx.WithLocale(context.Background(), "mr", cat)
	rec := httptest.NewRecorder()
	httpx.WriteError(ctx, rec, errors.E(errors.KindIdempotencyExpired, "idempotency_key_expired", "expired"))
	p := decodeProblem(t, rec.Body.Bytes())
	if strings.Contains(p.Title, "kernel.problem") {
		t.Fatalf("raw catalog key leaked as title: %q", p.Title)
	}
	if p.Title != "Internal error" {
		t.Errorf("absent-kind title = %q, want 'Internal error' (matches zero-config path)", p.Title)
	}
	// Machine code still stable/correct regardless.
	if p.Code != "idempotency_key_expired" {
		t.Errorf("code = %q", p.Code)
	}
}
