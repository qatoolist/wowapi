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

// RegisterFrameworkLocale is the SANCTIONED path for a product to supply
// translations of the framework's own kernel.* strings (e.g. a Marathi
// translation of the framework's problem titles). It is the guarded replacement
// for raw Catalog.Add: unlike Add it validates that every key is a kernel.* key
// AND already exists in the framework defaults (so a product may retranslate a
// framework string but not invent a new kernel.* key), recording violations via
// Err() like every other registration path. The product composition root calls
// this; modules must not (they have no reason to touch kernel.*, and Register
// rejects kernel.* from them).
//
// In practice the scaffold wires product kernel.* overrides through the config-
// driven fs override layer (ApplyLayers), which is the file-based equivalent;
// RegisterFrameworkLocale is the in-code equivalent for products that prefer a
// Go bundle for their framework overrides.
func (r *Registry) RegisterFrameworkLocale(b Bundle) {
	if b.Locale == "" {
		r.errf("i18n: framework-locale bundle registered with no locale")
		return
	}
	for key, msg := range b.Messages {
		if !strings.HasPrefix(key, reservedPrefix) {
			r.errf("i18n: RegisterFrameworkLocale key %q must be in the reserved %q namespace", key, reservedPrefix)
			continue
		}
		if !r.cat.hasKeyAnyLocale(key) {
			r.errf("i18n: RegisterFrameworkLocale may retranslate framework keys, not invent %q", key)
			continue
		}
		r.cat.Add(b.Locale, key, msg)
	}
}

// ApplyLayers merges configured source layers (framework overrides, product/
// module catalog files, Go bundles) into the registry's live catalog in
// precedence order, ON TOP of the framework defaults and any module bundles
// already registered. It applies the same ownership and intra-layer duplicate
// rules as LoadCatalog. Boot calls this once after modules have registered and
// before Freeze; a violation is recorded via Err() so boot fails closed with
// every other registration error. Pass the framework-defaults layer FIRST only
// when constructing a standalone catalog — here the registry already carries the
// framework defaults, so callers pass just the product layers.
func (r *Registry) ApplyLayers(layers ...Layer) {
	if err := mergeLayers(r.cat, layers); err != nil {
		r.errs = append(r.errs, err)
	}
}

// Freeze seals the catalog after boot so request-time reads never race a write.
func (r *Registry) Freeze() { r.cat.Freeze() }

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
