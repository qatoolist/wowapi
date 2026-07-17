package httpx_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/kernel/httpx"
)

// ---------- SecureHeaders (blueprint 07 §1: nosniff, frame-ancestors 'none', HSTS) ----------

func TestSecureHeadersDefaults(t *testing.T) {
	h := httpx.SecureHeaders()(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
	}
	if got := rec.Header().Get("Content-Security-Policy"); !strings.Contains(got, "frame-ancestors 'none'") {
		t.Errorf("CSP = %q, want it to contain frame-ancestors 'none'", got)
	}
	if got := rec.Header().Get("Strict-Transport-Security"); got == "" {
		t.Error("Strict-Transport-Security must be set by default")
	}
	if got := rec.Header().Get("Referrer-Policy"); got == "" {
		t.Error("Referrer-Policy must be set by default")
	}
}

func TestSecureHeadersWithoutHSTS(t *testing.T) {
	h := httpx.SecureHeaders(httpx.WithoutHSTS())(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if got := rec.Header().Get("Strict-Transport-Security"); got != "" {
		t.Errorf("HSTS must be absent when disabled, got %q", got)
	}
	// The other headers must still be present.
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("nosniff must still be set with HSTS disabled")
	}
}

// ---------- CORS (blueprint 07 §1: allowlist per env) ----------

func TestCORSAllowedOriginEchoed(t *testing.T) {
	served := false
	h := httpx.CORS(httpx.CORSPolicy{AllowedOrigins: []string{"https://app.example.com"}})(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { served = true }))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Errorf("ACAO = %q, want the echoed origin", got)
	}
	if !strings.Contains(rec.Header().Get("Vary"), "Origin") {
		t.Error("Vary must include Origin when CORS echoes the request origin")
	}
	if !served {
		t.Error("a non-preflight request must still reach the handler")
	}
}

func TestCORSDisallowedOriginNoHeaders(t *testing.T) {
	served := false
	h := httpx.CORS(httpx.CORSPolicy{AllowedOrigins: []string{"https://app.example.com"}})(
		http.HandlerFunc(func(http.ResponseWriter, *http.Request) { served = true }))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("ACAO must be empty for a disallowed origin, got %q", got)
	}
	if !served {
		t.Error("a disallowed simple request is still served; the browser enforces the block")
	}
}

func TestCORSPreflightShortCircuits(t *testing.T) {
	served := false
	h := httpx.CORS(httpx.CORSPolicy{
		AllowedOrigins: []string{"https://app.example.com"},
		AllowedMethods: []string{"GET", "POST"},
		MaxAge:         600 * time.Second,
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { served = true }))
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want 204", rec.Code)
	}
	if served {
		t.Error("a preflight OPTIONS must NOT reach the handler")
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(got, "POST") {
		t.Errorf("ACAM = %q, want it to contain POST", got)
	}
	if got := rec.Header().Get("Access-Control-Max-Age"); got == "" {
		t.Error("Access-Control-Max-Age must be set on a preflight")
	}
}

func TestCORSCredentialsNeverWildcard(t *testing.T) {
	h := httpx.CORS(httpx.CORSPolicy{
		AllowedOrigins:   []string{"https://app.example.com"},
		AllowCredentials: true,
	})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("ACAC = %q, want true", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got == "*" {
		t.Error("ACAO must never be * when credentials are allowed")
	}
}

// ---------- BodyLimit (blueprint 07 §1: BodyLimit(1MB default)) ----------

func TestBodyLimitCapsRequestBody(t *testing.T) {
	h := httpx.BodyLimit(10)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			http.Error(w, "too big", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("x", 100)))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("status = %d, want 413 when body exceeds the limit", rec.Code)
	}
}

func TestBodyLimitAllowsWithinLimit(t *testing.T) {
	h := httpx.BodyLimit(1024)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := io.Copy(io.Discard, r.Body); err != nil {
			http.Error(w, "too big", http.StatusRequestEntityTooLarge)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("small"))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 for a body within the limit", rec.Code)
	}
}

// ---------- Timeout (blueprint 07 §1: Timeout(30s default)) ----------

func TestTimeoutFiresOnSlowHandler(t *testing.T) {
	h := httpx.Timeout(30 * time.Millisecond)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-time.After(2 * time.Second):
			w.WriteHeader(http.StatusOK)
		case <-r.Context().Done():
			// timed out; the middleware owns the response
		}
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503 on handler timeout", rec.Code)
	}
}

func TestTimeoutPassesFastHandler(t *testing.T) {
	h := httpx.Timeout(time.Second)(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusTeapot {
		t.Errorf("status = %d, want the handler's 418 to pass through", rec.Code)
	}
}
