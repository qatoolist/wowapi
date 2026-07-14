package httpx

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"io"
	"net/http"
)

// csrf.go — CSRF defense for SecurityProfileBrowser (backlog B7; benchmark
// "Security: Profiles, Not Handler-Level Advice"). ONLY wired into the
// request chain when a product selects the browser/session security profile
// (see SecurityChain); the default API profile never imports this behavior.
//
// Pattern chosen: double-submit cookie, not the synchronizer-token pattern.
// Rationale: the synchronizer-token pattern requires the server to persist
// the expected token per session (or per form) and look it up on every
// unsafe request — that is a session store, which backlog B7 explicitly
// scopes OUT ("do NOT build a session store — CSRF token issuance/validation
// only"). Double-submit cookie needs no server-side state: the token is
// generated once, handed to the client as a cookie, and the client must echo
// it back via a header (or form field, for classic HTML posts) that
// same-site JavaScript/forms can read but a cross-site attacker cannot set
// (the attacker's forged request auto-attaches the cookie but has no way to
// read its value to also set the header). This is the standard approach
// recommended by OWASP's CSRF Cheat Sheet for stateless/token-only APIs and
// is what Django's CsrfViewMiddleware and Rails' non-session CSRF mode both
// implement as their cookie-based option.
//
// The CSRF cookie is intentionally NOT HttpOnly: the client must be able to
// read it (to place it in the request header). This is safe because the
// cookie is not a secret credential on its own — knowledge of it only lets
// the holder prove they are same-origin, which is exactly what CSRF defends
// against needing.

// safeCSRFMethods are exempt from token verification (RFC 9110 "safe"
// methods never mutate state, so there is nothing for CSRF to protect).
var safeCSRFMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodHead:    true,
	http.MethodOptions: true,
	http.MethodTrace:   true,
}

// CSRFPolicy configures the double-submit-cookie token names and the cookie
// attributes used when issuing a fresh token. Mirrors config.CSRF +
// config.CookieDefaults so SecurityChain can build one directly from loaded
// config without kernel/httpx depending on cookie-attribute parsing logic
// living in kernel/config.
type CSRFPolicy struct {
	CookieName string // e.g. "csrf_token"
	HeaderName string // e.g. "X-CSRF-Token"
	FieldName  string // form field fallback, e.g. "csrf_token"
	SameSite   string // "strict"|"lax"|"none" (case-insensitive)
	Secure     bool
	// MaxFormBytes bounds the request-body read performed by the form-field
	// fallback (gosec G120 / FBL-09). CSRF's chain position is app-controlled:
	// ordered outside BodyLimit, r.FormValue would otherwise buffer an
	// unbounded body before the token check, so the middleware is defensively
	// self-bounding regardless of ordering. Zero means the default
	// csrfDefaultMaxFormBytes (1 MiB, matching http.max_body_bytes' default).
	MaxFormBytes int64
}

// csrfDefaultMaxFormBytes is the form-fallback body bound applied when
// CSRFPolicy.MaxFormBytes is unset — 1 MiB, the same value as the framework's
// default http.max_body_bytes guardrail, so with default config the inner
// bound is exactly as permissive as the outer BodyLimit middleware and
// changes no in-bound request's behavior.
const csrfDefaultMaxFormBytes = 1 << 20

// sameSiteOf maps a policy's SameSite string to the net/http enum, defaulting
// to Lax for any unrecognized value (Validate() rejects unrecognized values
// long before this runs in a properly-validated config, so this default only
// matters for a policy built by hand outside the config path).
func sameSiteOf(v string) http.SameSite {
	switch v {
	case "strict", "Strict", "STRICT":
		return http.SameSiteStrictMode
	case "none", "None", "NONE":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// csrfRandReader is the entropy source for newCSRFToken. Overridable only by
// this package's own tests (to exercise the error path deterministically);
// production code always runs with the default crypto/rand.Reader.
var csrfRandReader io.Reader = rand.Reader

// newCSRFToken returns a fresh random token, base64url-encoded (cookie- and
// header-safe, no padding to keep it clean in both contexts).
func newCSRFToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := io.ReadFull(csrfRandReader, buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// CSRFProtect implements the double-submit-cookie CSRF defense. Safe methods
// (GET/HEAD/OPTIONS/TRACE) pass through untouched, issuing a token cookie
// only if one is not already present. Unsafe methods require the cookie
// value to be present and to match a token supplied via the configured
// header (checked first) or form field (fallback for HTML form posts);
// otherwise the request is rejected with 403 before reaching the handler.
//
// This middleware is only ever installed by SecurityChain when
// config.SecurityProfileBrowser is selected — it must never run in the
// default API profile's chain.
func CSRFProtect(p CSRFPolicy) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if safeCSRFMethods[r.Method] {
				ensureCSRFCookie(w, r, p)
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(p.CookieName)
			if err != nil || cookie.Value == "" {
				writeCSRFRejected(w, "csrf: missing token cookie")
				return
			}

			supplied := r.Header.Get(p.HeaderName)
			if supplied == "" && p.FieldName != "" {
				// FormValue parses the body for form-encoded/multipart requests only;
				// it never consumes a JSON body, so JSON callers must use the header.
				// The read is defensively capped (MaxFormBytes): an oversized body
				// fails the form parse, leaving supplied empty → 403 below, instead
				// of being fully buffered (gosec G120).
				limit := p.MaxFormBytes
				if limit <= 0 {
					limit = csrfDefaultMaxFormBytes
				}
				r.Body = http.MaxBytesReader(w, r.Body, limit)
				supplied = r.FormValue(p.FieldName)
			}
			if supplied == "" {
				writeCSRFRejected(w, "csrf: missing token in header or form field")
				return
			}
			if subtle.ConstantTimeCompare([]byte(supplied), []byte(cookie.Value)) != 1 {
				writeCSRFRejected(w, "csrf: token mismatch")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ensureCSRFCookie sets a fresh CSRF token cookie only when the request does
// not already carry one, so a client's existing valid token is never churned
// on every safe navigation.
func ensureCSRFCookie(w http.ResponseWriter, r *http.Request, p CSRFPolicy) {
	if c, err := r.Cookie(p.CookieName); err == nil && c.Value != "" {
		return
	}
	token, err := newCSRFToken()
	if err != nil {
		// Token generation failure means we simply don't issue a cookie this
		// round; the next safe request tries again. Failing the response here
		// would turn a transient entropy hiccup into an outage for a defense
		// that is only needed on unsafe methods.
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     p.CookieName,
		Value:    token,
		Path:     "/",
		Secure:   p.Secure,
		SameSite: sameSiteOf(p.SameSite),
		// Intentionally NOT HttpOnly — see the package-level doc comment: the
		// client must be able to read this value to echo it back.
	})
}

// writeCSRFRejected writes a minimal 403 problem-details body. Kept
// dependency-free of the request's error/i18n context (kerr.E + WriteError)
// because CSRF rejection happens before authentication/tenant binding in the
// chain and must never depend on state that middleware hasn't set up yet.
func writeCSRFRejected(w http.ResponseWriter, detail string) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte(`{"status":403,"code":"csrf_rejected","title":"CSRF validation failed","detail":"` + detail + `"}`))
}
