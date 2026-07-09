package i18n

import (
	"fmt"
	"strings"

	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// Bundle is a set of messages a module (or the product) registers for one
// locale. Every key must be prefixed with the registering module's name
// ("<module>."); the framework owns the reserved "kernel." namespace.
type Bundle struct {
	// Locale is the BCP 47 locale tag these messages are written in (e.g. "en",
	// "mr"). Required.
	Locale string
	// Messages maps stable message keys to their translated text for Locale.
	Messages map[string]string
}

// Registry collects i18n bundles from the framework and from modules, mirroring
// kernel/notify's Registry and kernel/seeds' merge: contributions accumulate,
// ownership is enforced per module, and errors are surfaced at boot via Err()
// rather than panicking mid-registration. The framework's own English catalog is
// installed at construction, so a Registry is never empty.
type Registry struct {
	cat  *Catalog
	errs []error
}

// NewRegistry returns a Registry pre-loaded with the framework's English catalog
// (problem titles + validation messages under the reserved kernel.* namespace).
// English is the default locale and ultimate fallback.
func NewRegistry() *Registry {
	cat := NewCatalog(DefaultLocale)
	installFramework(cat)
	return &Registry{cat: cat}
}

// Register merges a module's bundle into the catalog. Every key must be prefixed
// with module + "." and must not fall in the reserved kernel.* namespace. A bad
// locale, ownership violation, or reserved-namespace write records an error
// retrievable via Err() (the whole bundle is validated, not short-circuited).
func (r *Registry) Register(module string, b Bundle) {
	if b.Locale == "" {
		r.errf("i18n: module %q registered a bundle with no locale", module)
		return
	}
	prefix := module + "."
	for key := range b.Messages {
		if strings.HasPrefix(key, reservedPrefix) {
			r.errf("i18n: module %q may not register reserved framework key %q", module, key)
			continue
		}
		if !strings.HasPrefix(key, prefix) {
			r.errf("i18n: module %q may not register key %q (must be prefixed %q)", module, key, prefix)
			continue
		}
	}
	// Apply only well-formed keys; malformed ones are reported above and skipped
	// so a single typo does not silently drop the rest of the bundle.
	for key, msg := range b.Messages {
		if strings.HasPrefix(key, reservedPrefix) || !strings.HasPrefix(key, prefix) {
			continue
		}
		r.cat.Add(b.Locale, key, msg)
	}
}

// Catalog returns the merged catalog. Safe to call at any point; the returned
// pointer reflects later Register calls (it is the live catalog).
func (r *Registry) Catalog() *Catalog { return r.cat }

func (r *Registry) errf(format string, args ...any) {
	r.errs = append(r.errs, kerr.E(kerr.KindInternal, "invalid_i18n_bundle", fmt.Sprintf(format, args...)))
}

// Err returns the accumulated registration errors joined, or nil. Callers (app
// boot) must check this before serving, consistent with the other registries.
func (r *Registry) Err() error {
	if len(r.errs) == 0 {
		return nil
	}
	msgs := make([]string, len(r.errs))
	for i, e := range r.errs {
		msgs[i] = e.Error()
	}
	return kerr.E(kerr.KindInternal, "i18n_registration_failed",
		"i18n bundle registration failed: "+strings.Join(msgs, "; "))
}
