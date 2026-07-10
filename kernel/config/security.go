package config

import (
	"errors"
	"fmt"
)

// security.go — SecurityProfile config (backlog B7; benchmark "Security:
// Profiles, Not Handler-Level Advice"). Security posture is selected by
// PROFILE, not assembled ad hoc per product: the API profile (default)
// reproduces exactly what the framework does today (bearer/API-key auth,
// CSRF-free by contract, strict JSON, CORS allowlist, RLS guard); the
// browser/session profile is new, opt-in surface that additionally wires
// CSRF protection, SameSite cookie defaults, and a CSP header profile for
// HTML responses. Selecting the API profile (or leaving Security unset,
// which resolves to the same zero-cost default) must never change existing
// behavior — that is the B7 risk this file is written to respect.
//
// A safe outbound HTTP client (DNS/IP-blocking SSRF guard) is a separate
// concern already delivered by kernel/webhook.HTTPSender (backlog B2); this
// file does not duplicate it — see the doc comment on Security below.

// SecurityProfile selects the framework's security posture for a deployment.
type SecurityProfile string

const (
	// SecurityProfileAPI is the DEFAULT profile: bearer/API-key auth, no
	// cookies, CSRF disabled by contract (there is no cookie-based session to
	// forge), strict JSON decoding, CORS allowlist, RLS guard. This is
	// exactly what wowapi does today — selecting it (or leaving Security
	// unset) changes NOTHING.
	SecurityProfileAPI SecurityProfile = "api"

	// SecurityProfileBrowser is the OPT-IN profile for products that serve a
	// browser/cookie-session client: it additionally wires CSRF token
	// enforcement on state-changing requests, SameSite cookie defaults, and a
	// CSP header suitable for HTML. No product gains this behavior by
	// selecting anything other than this profile.
	SecurityProfileBrowser SecurityProfile = "browser"
)

// Valid reports whether p is a known security profile.
func (p SecurityProfile) Valid() bool {
	switch p {
	case SecurityProfileAPI, SecurityProfileBrowser:
		return true
	}
	return false
}

// Security is the framework-owned security-profile configuration (Framework
// field, loaded/validated once at boot like every other Framework section).
// CSRF and Cookie are only enforced/consulted when Profile is
// SecurityProfileBrowser; under SecurityProfileAPI they are inert (and may be
// left at their zero value).
type Security struct {
	// Profile selects the security posture. Empty resolves to
	// SecurityProfileAPI via DefaultSecurity()/Defaults() — see the fail-safe
	// note on Validate().
	Profile SecurityProfile `conf:"profile" default:"api" json:"profile" doc:"security profile: api (default, bearer/API-key, CSRF-free) or browser (cookie/session, CSRF-protected)"`
	// CSRF configures the double-submit-cookie token names. Only consulted
	// when Profile is browser.
	CSRF CSRF `conf:"csrf" json:"csrf"`
	// Cookie configures SameSite/Secure defaults for any cookie the browser
	// profile sets (the CSRF cookie today; a product's own session cookie
	// should follow the same defaults). Only consulted when Profile is
	// browser.
	Cookie CookieDefaults `conf:"cookie" json:"cookie"`
	// CSP overrides the Content-Security-Policy value applied by the browser
	// profile's header chain. Empty uses a conservative built-in default
	// (kernel/httpx.SecurityChain's browserCSPDefault). Only applied when
	// Profile is browser — the API profile keeps httpx.SecureHeaders'
	// existing "frame-ancestors 'none'" default untouched.
	CSP string `conf:"csp" json:"csp" doc:"Content-Security-Policy value for the browser profile; empty uses the built-in HTML-safe default"`
}

// CSRF names the cookie/header/form-field pair used by the double-submit-
// cookie CSRF defense (kernel/httpx.CSRFProtect). Both names must be
// non-empty when Profile is browser.
type CSRF struct {
	CookieName string `conf:"cookie_name" default:"csrf_token" json:"cookie_name" doc:"name of the CSRF token cookie set for browser clients"`
	HeaderName string `conf:"header_name" default:"X-CSRF-Token" json:"header_name" doc:"request header the client must echo the CSRF token in"`
	FieldName  string `conf:"field_name" default:"csrf_token" json:"field_name" doc:"form field accepted as an alternative to the header for classic HTML form posts"`
}

// CookieDefaults configures the SameSite/Secure attributes applied to cookies
// the browser profile sets. This is CSRF-token issuance policy only — the
// framework does not build or own a session store (backlog B7 scope: token
// issuance/validation, not session management).
type CookieDefaults struct {
	// SameSite is one of "strict", "lax", or "none" (case-insensitive on
	// input; canonicalized to lowercase). "lax" is the safe, commonly-usable
	// default recommended by OWASP for session-adjacent cookies.
	SameSite string `conf:"same_site" default:"lax" json:"same_site" doc:"cookie SameSite attribute: strict|lax|none"`
	// Secure marks cookies HTTPS-only. Required (validated) when SameSite is
	// "none", since browsers reject SameSite=None without Secure.
	Secure bool `conf:"secure" default:"true" json:"secure" doc:"set the Secure attribute on cookies (required when same_site=none)"`
}

// DefaultSecurity returns the compiled default: the API profile, exactly
// reproducing today's framework behavior. Framework-level Defaults() embeds
// this so a Framework zero value plus Defaults() always validates.
func DefaultSecurity() Security {
	return Security{
		Profile: SecurityProfileAPI,
		CSRF:    CSRF{CookieName: "csrf_token", HeaderName: "X-CSRF-Token", FieldName: "csrf_token"},
		Cookie:  CookieDefaults{SameSite: "lax", Secure: true},
	}
}

// Validate checks the security section. Like Framework.Validate it returns
// ALL problems joined. Under SecurityProfileAPI, CSRF/Cookie are not
// enforced — they may be zero-valued (the API profile ignores them by
// contract) — so a plain `Security{Profile: SecurityProfileAPI}` (as a
// hand-written config file might produce before defaults are applied)
// validates cleanly. Under SecurityProfileBrowser every field the CSRF
// middleware depends on must be coherent, since an incoherent browser
// profile would silently disable the CSRF defense (config validate must
// reject that, not boot into it).
func (s Security) Validate() error {
	var errs []error
	add := func(format string, args ...any) { errs = append(errs, fmt.Errorf(format, args...)) }

	profile := s.Profile
	if profile == "" {
		profile = SecurityProfileAPI // unset resolves to the safe default, not an error
	}
	if !profile.Valid() {
		add("security.profile: %q is not one of api|browser", string(s.Profile))
		return errors.Join(errs...)
	}

	if profile != SecurityProfileBrowser {
		return errors.Join(errs...) // API profile: CSRF/Cookie are inert, nothing else to check
	}

	if s.CSRF.CookieName == "" {
		add("security.csrf.cookie_name: required when security.profile is browser")
	}
	if s.CSRF.HeaderName == "" {
		add("security.csrf.header_name: required when security.profile is browser")
	}

	switch normalizeSameSite(s.Cookie.SameSite) {
	case "strict", "lax":
	case "none":
		if !s.Cookie.Secure {
			add("security.cookie.secure: must be true when security.cookie.same_site is none (browsers reject SameSite=None without Secure)")
		}
	default:
		add("security.cookie.same_site: %q is not one of strict|lax|none", s.Cookie.SameSite)
	}

	return errors.Join(errs...)
}

// normalizeSameSite lowercases the configured SameSite value for comparison;
// the stored/serialized value is left as the operator wrote it.
func normalizeSameSite(v string) string {
	out := make([]byte, len(v))
	for i := 0; i < len(v); i++ {
		c := v[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out[i] = c
	}
	return string(out)
}
