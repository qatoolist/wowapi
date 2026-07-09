package httpx

import (
	"context"
	"net/http"

	"github.com/qatoolist/wowapi/kernel/errors"
	"github.com/qatoolist/wowapi/kernel/i18n"
)

// The request-scoped locale carrier lives in kernel/i18n (a leaf package) so
// both httpx and kernel/validation can read it without an import cycle. These
// thin wrappers keep the httpx-facing name (WithLocale) while delegating to the
// shared carrier the Locale middleware, WriteError, and validation all share.

// WithLocale returns a context carrying the negotiated locale tag and the
// catalog to resolve messages against. A nil cat is allowed (Lookup on a nil
// catalog safely echoes keys).
func WithLocale(ctx context.Context, locale string, cat *i18n.Catalog) context.Context {
	return i18n.WithContext(ctx, locale, cat)
}

// LocaleFrom returns the negotiated locale tag, or "" if none was bound.
func LocaleFrom(ctx context.Context) string { return i18n.LocaleFrom(ctx) }

// CatalogFrom returns the active message catalog, or nil if none was bound.
func CatalogFrom(ctx context.Context) *i18n.Catalog { return i18n.CatalogFrom(ctx) }

// Locale negotiates the request locale from the Accept-Language header against
// cat's supported locales (RFC 9110 q-values), binds the result into the request
// context, and sets Content-Language on the response. It is a no-op when cat is
// nil, keeping zero-config products English-only with unchanged behavior.
//
// Placement: it must run after RequestID/Recover but before the route handler,
// so the handler (and any WriteError it calls) sees the bound locale. It reads
// only the request header and writes only Content-Language, so it composes
// cleanly with the edge and authz-gate middleware.
func Locale(cat *i18n.Catalog) Middleware {
	return func(next http.Handler) http.Handler {
		if cat == nil {
			return next
		}
		supported := cat.Locales()
		def := cat.Default()
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loc := i18n.Negotiate(r.Header.Get("Accept-Language"), supported, def)
			// Vary so shared caches key on the negotiated language.
			w.Header().Add("Vary", "Accept-Language")
			w.Header().Set("Content-Language", loc)
			ctx := WithLocale(r.Context(), loc, cat)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// localizeTitle resolves the problem title for kind in the request's negotiated
// locale, falling back to the framework's English title (and, if no catalog is
// bound at all, to the caller-supplied fallback English title) so a missing
// catalog or translation never changes the machine-stable output shape.
func localizeTitle(ctx context.Context, kind errors.Kind, fallback string) string {
	cat := CatalogFrom(ctx)
	if cat == nil {
		return fallback
	}
	key := i18n.KeyProblemTitle(kind)
	msg, _ := cat.Lookup(LocaleFrom(ctx), key)
	// A total miss (no entry even in the default locale) echoes the key back; in
	// that case use the caller's English fallback so the raw key never leaks onto
	// the wire and the localized path matches the zero-config path exactly. This
	// covers kinds deliberately absent from the framework catalog (e.g.
	// KindIdempotencyExpired, which resolves to the KindInternal title via the
	// fallback, identical to httpx's p.Title=="" behavior).
	if msg == key {
		return fallback
	}
	return msg
}
