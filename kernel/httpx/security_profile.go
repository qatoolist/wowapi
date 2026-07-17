package httpx

import "github.com/qatoolist/wowapi/v2/kernel/config"

// security_profile.go — wires config.Security into the concrete middleware
// chain (backlog B7). This is the single seam a scaffold test point at to
// prove the SELECTED profile is actually wired, rather than trusting that a
// config field merely exists: SecurityChain(apiProfile) must produce a chain
// with no CSRF middleware in it (proven by behavior: a state-changing
// request with no token passes straight through), and
// SecurityChain(browserProfile) must produce one that rejects it.
//
// browserCSPDefault is deliberately more permissive than SecureHeaders'
// JSON-API default ("frame-ancestors 'none'" only): HTML responses need to
// load their own scripts/styles/images. It still denies framing (clickjack
// defense) and restricts everything else to same-origin, which a product
// serving HTML can loosen per-directive via Security.CSP if it needs a CDN
// or third-party embed.
const browserCSPDefault = "default-src 'self'; frame-ancestors 'none'; base-uri 'self'"

// SecurityChain returns the middlewares SecurityProfile-specific behavior
// requires, in outer-to-first-listed-runs-outermost order matching Chain's
// convention. Under SecurityProfileAPI (the default) it returns CSRF-free,
// cookie-free middleware only — i.e. behavior identical to a product that
// never wired this function at all, satisfying the B7 risk that the API
// profile must stay byte-for-byte behavior-preserving. Under
// SecurityProfileBrowser it additionally installs CSRFProtect and a CSP
// header profile suited to HTML.
//
// SecureHeaders/CORS/BodyLimit/Timeout are NOT part of this chain — they are
// the existing fixed edge chain (edge.go) shared by both profiles; this
// function only adds profile-SPECIFIC behavior on top of it. A composition
// root appends SecurityChain's output to (or splices it into) its existing
// Chain(...) call.
func SecurityChain(sec config.Security) []Middleware {
	profile := sec.Profile
	if profile == "" {
		profile = config.SecurityProfileAPI
	}
	if profile != config.SecurityProfileBrowser {
		return nil // API profile: no additional middleware, current behavior untouched
	}

	csp := sec.CSP
	if csp == "" {
		csp = browserCSPDefault
	}

	return []Middleware{
		SecureHeaders(WithCSP(csp)),
		CSRFProtect(CSRFPolicy{
			CookieName: sec.CSRF.CookieName,
			HeaderName: sec.CSRF.HeaderName,
			FieldName:  sec.CSRF.FieldName,
			SameSite:   sec.Cookie.SameSite,
			Secure:     sec.Cookie.Secure,
		}),
	}
}
