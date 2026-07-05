package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/kernel/httpx"
)

// serveMW runs a single request through a middleware-wrapped 200 handler.
func serveMW(mw httpx.Middleware, r *http.Request) *httptest.ResponseRecorder {
	h := mw(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

// TestSecureHeadersWithHSTS overrides the HSTS value.
func TestSecureHeadersWithHSTS(t *testing.T) {
	rec := serveMW(httpx.SecureHeaders(httpx.WithHSTS("max-age=42")), httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("Strict-Transport-Security"); got != "max-age=42" {
		t.Fatalf("HSTS = %q, want max-age=42", got)
	}
}

// TestSecureHeadersWithCSP overrides the Content-Security-Policy value.
func TestSecureHeadersWithCSP(t *testing.T) {
	rec := serveMW(httpx.SecureHeaders(httpx.WithCSP("default-src 'self'")), httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("Content-Security-Policy"); got != "default-src 'self'" {
		t.Fatalf("CSP = %q, want default-src 'self'", got)
	}
}

// TestSecureHeadersEmptyCSPOmitsHeader proves an explicitly-empty CSP disables
// the header rather than emitting a blank one.
func TestSecureHeadersEmptyCSPOmitsHeader(t *testing.T) {
	rec := serveMW(httpx.SecureHeaders(httpx.WithCSP("")), httptest.NewRequest(http.MethodGet, "/", nil))
	if _, set := rec.Header()["Content-Security-Policy"]; set {
		t.Fatal("empty CSP must not emit a Content-Security-Policy header")
	}
	// The other baseline headers are still present.
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("nosniff must still be set")
	}
}

// TestTimeoutDisabledPassesThrough proves d <= 0 returns the handler unwrapped
// (no TimeoutHandler), so the request is served normally.
func TestTimeoutDisabledPassesThrough(t *testing.T) {
	served := false
	h := httpx.Timeout(0)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		served = true
		w.WriteHeader(http.StatusTeapot)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if !served || rec.Code != http.StatusTeapot {
		t.Fatalf("disabled timeout must pass through unchanged; served=%v code=%d", served, rec.Code)
	}
}

// TestCORSExposesHeadersAndCredentials covers the credentials + exposed-headers
// branches on an allowed non-preflight (simple) request.
func TestCORSExposesHeadersAndCredentials(t *testing.T) {
	mw := httpx.CORS(httpx.CORSPolicy{
		AllowedOrigins:   []string{"https://app.example"},
		ExposedHeaders:   []string{"X-Total-Count"},
		AllowCredentials: true,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.example")
	rec := serveMW(mw, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example" {
		t.Fatalf("allow-origin = %q, want the exact origin echoed", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("allow-credentials = %q, want true", got)
	}
	if got := rec.Header().Get("Access-Control-Expose-Headers"); got != "X-Total-Count" {
		t.Fatalf("expose-headers = %q, want X-Total-Count", got)
	}
}

// TestCORSPreflightMaxAge covers the MaxAge branch of a preflight response.
func TestCORSPreflightMaxAge(t *testing.T) {
	mw := httpx.CORS(httpx.CORSPolicy{
		AllowedOrigins: []string{"https://app.example"},
		MaxAge:         600 * time.Second,
	})
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://app.example")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := serveMW(mw, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("preflight = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Max-Age"); got != "600" {
		t.Fatalf("max-age = %q, want 600", got)
	}
}
