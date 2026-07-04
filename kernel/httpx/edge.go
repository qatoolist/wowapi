package httpx

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// This file implements the edge middleware the blueprint's fixed chain calls for
// (07 §1): SecureHeaders → CORS → BodyLimit → Timeout. They are kernel-owned so
// every product built on wowapi ships the same baseline posture without wiring
// it by hand; the reference deployment (deployments/reference/) documents the
// reverse-proxy layer that terminates TLS in front of them.

// --- SecureHeaders ---------------------------------------------------------

// defaultHSTS is one year with subdomains — the standard production posture.
// HSTS is only meaningful over HTTPS; TLS is LB-terminated, so the header is
// emitted unconditionally and WithoutHSTS() opts out for plain-HTTP contexts.
const defaultHSTS = "max-age=31536000; includeSubDomains"

type secureHeadersConfig struct {
	hsts string // "" disables the header
	csp  string
}

// SecureHeadersOption tunes the SecureHeaders middleware.
type SecureHeadersOption func(*secureHeadersConfig)

// WithoutHSTS disables the Strict-Transport-Security header — use only where TLS
// is not terminated in front of the service (e.g. local plain-HTTP).
func WithoutHSTS() SecureHeadersOption {
	return func(c *secureHeadersConfig) { c.hsts = "" }
}

// WithHSTS overrides the Strict-Transport-Security value.
func WithHSTS(value string) SecureHeadersOption {
	return func(c *secureHeadersConfig) { c.hsts = value }
}

// WithCSP overrides the Content-Security-Policy value. The default only asserts
// frame-ancestors 'none' (clickjacking defense for a JSON API); products that
// serve HTML should set a fuller policy.
func WithCSP(value string) SecureHeadersOption {
	return func(c *secureHeadersConfig) { c.csp = value }
}

// SecureHeaders sets the baseline response security headers on every response
// (blueprint 07 §1): nosniff, frame-ancestors 'none', HSTS, and a strict
// Referrer-Policy. Headers are set before the handler runs so they are present
// even on handler-written error responses.
func SecureHeaders(opts ...SecureHeadersOption) Middleware {
	cfg := secureHeadersConfig{hsts: defaultHSTS, csp: "frame-ancestors 'none'"}
	for _, o := range opts {
		o(&cfg)
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "no-referrer")
			if cfg.csp != "" {
				h.Set("Content-Security-Policy", cfg.csp)
			}
			if cfg.hsts != "" {
				h.Set("Strict-Transport-Security", cfg.hsts)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// --- CORS ------------------------------------------------------------------

// CORSPolicy is the per-environment CORS allowlist (blueprint 07 §1). The zero
// policy allows no origins — CORS is deny-by-default, consistent with the rest
// of the framework. Products load the allowlist from their env config.
type CORSPolicy struct {
	AllowedOrigins   []string // exact-match origins; no wildcards (deny-by-default)
	AllowedMethods   []string // default: GET, POST, PUT, PATCH, DELETE, OPTIONS
	AllowedHeaders   []string // default: Authorization, Content-Type, X-Request-Id, Idempotency-Key
	ExposedHeaders   []string // response headers the browser may read
	AllowCredentials bool     // send Access-Control-Allow-Credentials
	MaxAge           time.Duration
}

var (
	defaultCORSMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	defaultCORSHeaders = []string{"Authorization", "Content-Type", "X-Request-Id", "Idempotency-Key"}
)

// CORS enforces an origin allowlist. A request whose Origin is on the list gets
// that exact origin echoed back (never "*", so it composes with credentials);
// a preflight (OPTIONS + Access-Control-Request-Method) is answered with 204 and
// never reaches the handler. Requests from disallowed origins are served without
// CORS headers — the browser, not the server, enforces the block.
func CORS(p CORSPolicy) Middleware {
	allowed := make(map[string]bool, len(p.AllowedOrigins))
	for _, o := range p.AllowedOrigins {
		allowed[o] = true
	}
	methods := p.AllowedMethods
	if len(methods) == 0 {
		methods = defaultCORSMethods
	}
	headers := p.AllowedHeaders
	if len(headers) == 0 {
		headers = defaultCORSHeaders
	}
	methodsHdr := strings.Join(methods, ", ")
	headersHdr := strings.Join(headers, ", ")
	exposeHdr := strings.Join(p.ExposedHeaders, ", ")
	maxAge := strconv.Itoa(int(p.MaxAge.Seconds()))

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			// Vary on Origin regardless: the response body/headers depend on it,
			// so shared caches must key on it.
			w.Header().Add("Vary", "Origin")

			if origin != "" && allowed[origin] {
				h := w.Header()
				h.Set("Access-Control-Allow-Origin", origin)
				if p.AllowCredentials {
					h.Set("Access-Control-Allow-Credentials", "true")
				}
				if exposeHdr != "" {
					h.Set("Access-Control-Expose-Headers", exposeHdr)
				}
				// Preflight: answer and stop before the handler.
				if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
					h.Add("Vary", "Access-Control-Request-Method")
					h.Add("Vary", "Access-Control-Request-Headers")
					h.Set("Access-Control-Allow-Methods", methodsHdr)
					h.Set("Access-Control-Allow-Headers", headersHdr)
					if p.MaxAge > 0 {
						h.Set("Access-Control-Max-Age", maxAge)
					}
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// --- BodyLimit -------------------------------------------------------------

// BodyLimit caps the request body at maxBytes by wrapping r.Body in an
// http.MaxBytesReader. Any read past the limit fails with *http.MaxBytesError,
// which DecodeJSON maps to 413; handlers that read the body directly see the
// same error. maxBytes <= 0 disables the cap.
func BodyLimit(maxBytes int64) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if maxBytes > 0 && r.Body != nil {
				r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// --- Timeout ---------------------------------------------------------------

// Timeout enforces a per-request deadline. On expiry the request context is
// cancelled (so in-flight DB work aborts) and a 503 is written. d <= 0 disables
// the timeout. Built on http.TimeoutHandler, which buffers the handler's
// response so a late write cannot corrupt the timeout reply.
func Timeout(d time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		if d <= 0 {
			return next
		}
		return http.TimeoutHandler(next, d, "request timeout")
	}
}
