// Package i18n is wowapi's cross-cutting message-catalog and locale-negotiation
// kernel. It gives every product built on the framework one consistent way to
// localize synchronous API response strings — problem-detail titles/details and
// validation field messages — without each product re-implementing risky
// translation plumbing (GAP-001).
//
// The design has four pieces:
//
//   - Catalog: an in-process (locale, key) -> message map with a deterministic
//     fallback to a single default locale, and a final fallback to the key
//     itself so a missing translation can never break a response. Lookups are
//     read-only and allocation-free, safe to call on the request path.
//   - Registry: the module/product registration surface. It ships the framework's
//     own English catalog as the first bundle (problem titles + validation tag
//     messages, under the reserved "kernel." namespace) and lets modules add
//     their own bundles under their "<module>." prefix — mirroring how
//     kernel/notify and kernel/seeds accumulate module contributions and surface
//     ownership errors at boot via Err().
//   - Negotiate: RFC 9110 §12.5.4 Accept-Language q-value content negotiation,
//     ported from the battle-tested wowsociety implementation.
//   - Well-known keys (KeyProblemTitle / KeyValidationMessage): the stable keys
//     the framework's own English messages are stored under, so kernel/httpx and
//     kernel/validation can localize their output while keeping machine Codes
//     byte-stable.
//
// English is the default locale and the ultimate fallback. Internal logs are
// unaffected — they stay technical English by never routing through this package.
//
// Import boundary: stdlib + kernel/errors only. Never module, app, adapters, or
// testkit.
package i18n

import "sort"

// Catalog is an in-process message catalog keyed by locale then message key.
// The zero value is not usable; construct with NewCatalog. A nil *Catalog is a
// valid empty catalog (the zero-config path): every Lookup echoes the key.
type Catalog struct {
	def      string
	messages map[string]map[string]string // locale -> key -> message
	frozen   bool                         // sealed after boot; blocks further Add
}

// NewCatalog returns an empty Catalog whose fallback locale is def. def should
// be one of the locales later populated via Add so Lookup's fallback resolves to
// a real translation rather than echoing the key.
func NewCatalog(def string) *Catalog {
	return &Catalog{def: def, messages: make(map[string]map[string]string)}
}

// Add registers message for (locale, key), overwriting any prior value. Must be
// called on a Catalog built with NewCatalog (the zero value is not usable).
//
// Add is a no-op after Freeze: catalogs are sealed at boot (Decision 3) and are
// read-only on the request path, so a post-freeze mutation attempt is silently
// ignored rather than racing concurrent Lookups. Boot-time construction (the
// Loader, Registry) writes before Freeze; if you need a post-boot overlay, that
// is the separate opt-in B13 concern, not raw Add.
func (c *Catalog) Add(locale, key, message string) {
	if c.frozen {
		return
	}
	m, ok := c.messages[locale]
	if !ok {
		m = make(map[string]string)
		c.messages[locale] = m
	}
	m[key] = message
}

// Freeze seals the catalog for request-time reads. After Freeze, Add is a no-op,
// so the messages map is never mutated concurrently with Lookup. Boot calls this
// once, after every source and module bundle has been merged. Freeze is
// idempotent. A nil *Catalog Freeze is a no-op.
func (c *Catalog) Freeze() {
	if c == nil {
		return
	}
	c.frozen = true
}

// Frozen reports whether the catalog has been sealed.
func (c *Catalog) Frozen() bool { return c != nil && c.frozen }

// Lookup resolves key for locale. It returns the resolved message and the locale
// it was actually served from. Resolution is deterministic:
//
//  1. exact (locale, key) if present;
//  2. otherwise (default-locale, key) if present;
//  3. otherwise the key itself, with the default locale — never an error, so a
//     missing translation cannot break a response.
//
// A nil *Catalog echoes the key with an empty resolved locale.
func (c *Catalog) Lookup(locale, key string) (message, resolvedLocale string) {
	if c == nil {
		return key, ""
	}
	if locale != "" {
		if m, ok := c.messages[locale]; ok {
			if msg, ok := m[key]; ok {
				return msg, locale
			}
		}
	}
	if m, ok := c.messages[c.def]; ok {
		if msg, ok := m[key]; ok {
			return msg, c.def
		}
	}
	return key, c.def
}

// Supports reports whether locale has at least one registered message. A nil
// *Catalog supports nothing.
func (c *Catalog) Supports(locale string) bool {
	if c == nil {
		return false
	}
	m, ok := c.messages[locale]
	return ok && len(m) > 0
}

// Default returns the catalog's fallback locale.
func (c *Catalog) Default() string {
	if c == nil {
		return ""
	}
	return c.def
}

// Locales returns the sorted set of locales with at least one registered message.
func (c *Catalog) Locales() []string {
	if c == nil {
		return nil
	}
	out := make([]string, 0, len(c.messages))
	for loc := range c.messages {
		out = append(out, loc)
	}
	sort.Strings(out)
	return out
}
