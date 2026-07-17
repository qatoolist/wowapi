package testkit

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/httpx"
	"github.com/qatoolist/wowapi/v2/kernel/i18n"
)

// AssertNegotiatedLocale asserts that Accept-Language header value acceptLang,
// negotiated against cat's supported locales (RFC 9110 q-values), resolves to
// wantLocale. It mirrors the real httpx.Locale middleware path, so a test proves
// the negotiation contract without standing up a server.
func AssertNegotiatedLocale(t *testing.T, cat *i18n.Catalog, acceptLang, wantLocale string) {
	t.Helper()
	got := i18n.Negotiate(acceptLang, cat.Locales(), cat.Default())
	if got != wantLocale {
		t.Errorf("Negotiate(%q) = %q, want %q", acceptLang, got, wantLocale)
	}
}

// NewLocaleRequest builds an *http.Request whose context already carries the
// negotiated locale and catalog, as the httpx.Locale middleware would bind them.
// Use it to drive a handler under a specific locale without wiring the full
// middleware chain — WriteError and validation inside the handler will localize
// against cat.
func NewLocaleRequest(method, target, locale string, cat *i18n.Catalog) *http.Request {
	r := httptest.NewRequestWithContext(context.Background(), method, target, nil)
	return r.WithContext(httpx.WithLocale(r.Context(), locale, cat))
}

// AssertLocalizedProblem decodes an application/problem+json response and asserts
// its title is the localized wantTitle while its machine Code equals wantCode —
// the core i18n invariant: user-facing text localizes, machine codes stay stable.
// It consumes resp.Body.
func AssertLocalizedProblem(t *testing.T, resp *http.Response, wantCode, wantTitle string) {
	t.Helper()
	defer func() { _ = resp.Body.Close() }()
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/problem+json") {
		t.Errorf("Content-Type = %q, want application/problem+json", ct)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read problem body: %v", err)
	}
	var p httpx.ProblemError
	if err := json.Unmarshal(body, &p); err != nil {
		t.Fatalf("problem body not JSON: %v (%s)", err, body)
	}
	if p.Code != wantCode {
		t.Errorf("problem code = %q, want %q (machine code must stay stable across locales)", p.Code, wantCode)
	}
	if p.Title != wantTitle {
		t.Errorf("problem title = %q, want localized %q", p.Title, wantTitle)
	}
}
