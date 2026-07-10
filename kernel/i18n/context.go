package i18n

import "context"

// Request-scoped locale state lives here (not in kernel/httpx) so BOTH the HTTP
// layer and kernel/validation can read it without an import cycle: httpx imports
// validation, so the locale carrier must sit in a leaf package both can import.
// The httpx.Locale middleware binds these values; WriteError and
// validation.Validator.StructCtx read them.
type (
	localeKey  struct{}
	catalogKey struct{}
)

// WithContext binds the negotiated locale tag and the catalog to resolve
// messages against. A nil cat is allowed (Catalog.Lookup on nil echoes keys).
func WithContext(ctx context.Context, locale string, cat *Catalog) context.Context {
	ctx = context.WithValue(ctx, localeKey{}, locale)
	return context.WithValue(ctx, catalogKey{}, cat)
}

// LocaleFrom returns the bound locale tag, or "" if none.
func LocaleFrom(ctx context.Context) string {
	loc, _ := ctx.Value(localeKey{}).(string)
	return loc
}

// CatalogFrom returns the bound catalog, or nil if none.
func CatalogFrom(ctx context.Context) *Catalog {
	cat, _ := ctx.Value(catalogKey{}).(*Catalog)
	return cat
}
