package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/httpx"
)

// ---------- CSRFProtect (double-submit-cookie; backlog B7) ----------

func csrfPolicy() httpx.CSRFPolicy {
	return httpx.CSRFPolicy{
		CookieName: "csrf_token",
		HeaderName: "X-CSRF-Token",
		FieldName:  "csrf_token",
		SameSite:   "lax",
		Secure:     true,
	}
}

// TestCSRFSafeMethodIssuesCookie proves a safe (GET) request is served AND
// receives a CSRF cookie so the client has a token to echo back on the next
// state-changing request.
func TestCSRFSafeMethodIssuesCookie(t *testing.T) {
	served := false
	h := httpx.CSRFProtect(csrfPolicy())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !served {
		t.Fatal("a safe GET request must reach the handler")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	cookies := rec.Result().Cookies()
	var found *http.Cookie
	for _, c := range cookies {
		if c.Name == "csrf_token" {
			found = c
		}
	}
	if found == nil {
		t.Fatal("a safe request must set the csrf_token cookie when absent")
	}
	if found.Value == "" {
		t.Error("issued CSRF cookie must carry a non-empty token")
	}
	if !found.Secure {
		t.Error("issued CSRF cookie must be Secure per policy")
	}
	if found.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want Lax", found.SameSite)
	}
}

// TestCSRFSafeMethodReusesExistingCookie proves the middleware does not churn
// a fresh token on every safe request once one is already set.
func TestCSRFSafeMethodReusesExistingCookie(t *testing.T) {
	h := httpx.CSRFProtect(csrfPolicy())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "existing-token"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	for _, c := range rec.Result().Cookies() {
		if c.Name == "csrf_token" {
			t.Fatalf("must not re-issue a cookie when one is already present, got new value %q", c.Value)
		}
	}
}

// TestCSRFUnsafeMethodRejectsMissingToken is the core acceptance test: a
// state-changing request without a valid CSRF token is rejected.
func TestCSRFUnsafeMethodRejectsMissingToken(t *testing.T) {
	served := false
	h := httpx.CSRFProtect(csrfPolicy())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served = true
	}))
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if served {
		t.Fatal("a POST with no CSRF token at all must not reach the handler")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "csrf") {
		t.Errorf("body should mention csrf: %s", rec.Body.String())
	}
}

// TestCSRFUnsafeMethodRejectsMismatchedToken proves a header token that
// doesn't match the cookie is rejected — the whole point of double-submit.
func TestCSRFUnsafeMethodRejectsMismatchedToken(t *testing.T) {
	served := false
	h := httpx.CSRFProtect(csrfPolicy())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served = true
	}))
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "cookie-value"})
	req.Header.Set("X-CSRF-Token", "different-value")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if served {
		t.Fatal("mismatched CSRF token must not reach the handler")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

// TestCSRFUnsafeMethodPassesWithValidToken: a valid token (cookie == header)
// passes through to the handler.
func TestCSRFUnsafeMethodPassesWithValidToken(t *testing.T) {
	served := false
	h := httpx.CSRFProtect(csrfPolicy())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "matching-token"})
	req.Header.Set("X-CSRF-Token", "matching-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !served {
		t.Fatal("matching cookie+header CSRF token must reach the handler")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

// TestCSRFUnsafeMethodRejectsEmptyCookieValue proves an explicitly-empty
// cookie (present but blank) is treated the same as a missing cookie, not as
// "matches an empty supplied token".
func TestCSRFUnsafeMethodRejectsEmptyCookieValue(t *testing.T) {
	served := false
	h := httpx.CSRFProtect(csrfPolicy())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served = true
	}))
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: ""})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if served {
		t.Fatal("an empty CSRF cookie value must not reach the handler")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

// TestCSRFUnsafeMethodNoFormFallbackConfigured proves that when FieldName is
// left empty (no form fallback wanted), a header-less request is rejected
// without falling back to parsing the body.
func TestCSRFUnsafeMethodNoFormFallbackConfigured(t *testing.T) {
	policy := csrfPolicy()
	policy.FieldName = ""
	served := false
	h := httpx.CSRFProtect(policy)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served = true
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("csrf_token=matching-token"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "matching-token"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if served {
		t.Fatal("with FieldName unset, a header-less request must be rejected, not parsed from the body")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

// TestCSRFUnsafeMethodAcceptsFormField proves the form-field fallback works
// for classic HTML form posts that cannot set a custom header.
func TestCSRFUnsafeMethodAcceptsFormField(t *testing.T) {
	served := false
	h := httpx.CSRFProtect(csrfPolicy())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served = true
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("csrf_token=form-token"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "form-token"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !served {
		t.Fatal("matching cookie+form-field CSRF token must reach the handler")
	}
}

// TestCSRFSafeMethodsExemptFromTokenCheck proves GET/HEAD/OPTIONS never
// require a token even without a pre-existing cookie.
func TestCSRFSafeMethodsExemptFromTokenCheck(t *testing.T) {
	for _, method := range []string{http.MethodGet, http.MethodHead, http.MethodOptions} {
		served := false
		h := httpx.CSRFProtect(csrfPolicy())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			served = true
			w.WriteHeader(http.StatusOK)
		}))
		req := httptest.NewRequest(method, "/", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if !served {
			t.Errorf("%s must be exempt from CSRF token check", method)
		}
	}
}

// TestCSRFCookieSameSiteVariants covers the strict/none/lax branches of
// sameSiteOf and their case-insensitive spelling, which the default-policy
// tests above (all "lax") don't exercise.
func TestCSRFCookieSameSiteVariants(t *testing.T) {
	cases := []struct {
		configured string
		want       http.SameSite
	}{
		{"strict", http.SameSiteStrictMode},
		{"Strict", http.SameSiteStrictMode},
		{"none", http.SameSiteNoneMode},
		{"None", http.SameSiteNoneMode},
		{"lax", http.SameSiteLaxMode},
		{"", http.SameSiteLaxMode}, // unrecognized/empty defaults to Lax
	}
	for _, c := range cases {
		p := httpx.CSRFPolicy{CookieName: "csrf_token", HeaderName: "X-CSRF-Token", SameSite: c.configured, Secure: true}
		h := httpx.CSRFProtect(p)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		var found *http.Cookie
		for _, ck := range rec.Result().Cookies() {
			if ck.Name == "csrf_token" {
				found = ck
			}
		}
		if found == nil {
			t.Fatalf("SameSite=%q: cookie not issued", c.configured)
		}
		if found.SameSite != c.want {
			t.Errorf("SameSite=%q: got %v, want %v", c.configured, found.SameSite, c.want)
		}
	}
}

// ---------- SecurityChain (config.Security → middleware wiring) ----------

// TestSecurityChainAPIProfileHasNoCSRF is the scaffold-proof acceptance test:
// the API profile (default) must NOT install CSRF middleware — a
// state-changing request with no token must pass straight through, exactly
// today's behavior.
func TestSecurityChainAPIProfileHasNoCSRF(t *testing.T) {
	sec := config.DefaultSecurity() // profile: api
	chain := httpx.SecurityChain(sec)

	served := false
	h := httpx.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served = true
		w.WriteHeader(http.StatusOK)
	}), chain...)

	req := httptest.NewRequest(http.MethodPost, "/", nil) // no CSRF token anywhere
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !served {
		t.Fatal("API profile must never block a request for a missing CSRF token")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 (API profile is unaffected by CSRF)", rec.Code)
	}
	if got := rec.Header().Get("Set-Cookie"); got != "" {
		t.Errorf("API profile must not set any cookie, got %q", got)
	}
}

// TestSecurityChainBrowserProfileHasCSRF proves the browser profile DOES
// install CSRF middleware: a state-changing request without a token is
// rejected.
func TestSecurityChainBrowserProfileHasCSRF(t *testing.T) {
	sec := config.DefaultSecurity()
	sec.Profile = config.SecurityProfileBrowser
	chain := httpx.SecurityChain(sec)

	served := false
	h := httpx.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served = true
	}), chain...)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if served {
		t.Fatal("browser profile must reject a state-changing request with no CSRF token")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want 403", rec.Code)
	}
}

// TestSecurityChainBrowserProfileSetsCSP proves the browser profile applies a
// CSP header (HTML-safe profile), unlike the bare API SecureHeaders default.
func TestSecurityChainBrowserProfileSetsCSP(t *testing.T) {
	sec := config.DefaultSecurity()
	sec.Profile = config.SecurityProfileBrowser
	chain := httpx.SecurityChain(sec)

	h := httpx.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), chain...)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Security-Policy"); got == "" {
		t.Error("browser profile must set a Content-Security-Policy header")
	}
}

// TestSecurityChainBrowserProfileValidTokenPasses closes the loop: browser
// profile + a valid CSRF token succeeds end-to-end through SecurityChain.
func TestSecurityChainBrowserProfileValidTokenPasses(t *testing.T) {
	sec := config.DefaultSecurity()
	sec.Profile = config.SecurityProfileBrowser
	chain := httpx.SecurityChain(sec)

	served := false
	h := httpx.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		served = true
		w.WriteHeader(http.StatusOK)
	}), chain...)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: "tok"})
	req.Header.Set("X-CSRF-Token", "tok")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !served {
		t.Fatal("a valid CSRF token must pass through the browser profile's SecurityChain")
	}
}

// TestSecurityChainEmptyProfileDefaultsToAPI proves a zero-valued
// config.Security (Profile == "") resolves to the API profile, exactly like
// an explicit SecurityProfileAPI — matching Security.Validate()'s own
// empty-resolves-to-api rule.
func TestSecurityChainEmptyProfileDefaultsToAPI(t *testing.T) {
	chain := httpx.SecurityChain(config.Security{})
	if chain != nil {
		t.Fatalf("empty Profile must resolve to the API profile (nil chain), got %d middlewares", len(chain))
	}
}

// TestSecurityChainCustomCSPOverridesDefault proves Security.CSP, when set,
// is what the browser profile's SecureHeaders emits instead of the built-in
// browserCSPDefault.
func TestSecurityChainCustomCSPOverridesDefault(t *testing.T) {
	sec := config.DefaultSecurity()
	sec.Profile = config.SecurityProfileBrowser
	sec.CSP = "default-src 'none'"
	chain := httpx.SecurityChain(sec)

	h := httpx.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), chain...)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Content-Security-Policy"); got != "default-src 'none'" {
		t.Errorf("CSP = %q, want the configured override", got)
	}
}
