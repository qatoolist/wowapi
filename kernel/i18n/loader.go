package i18n

import (
	"fmt"
	"sort"
	"strings"

	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
)

// SourceKind names a first-class catalog source. It is part of the stable
// loader contract: config files, `wowapi i18n validate`, and the scaffold all
// refer to these by string, so values must not change once shipped.
type SourceKind string

const (
	// KindFrameworkDefaults is the framework's own embedded per-locale YAML
	// (kernel/i18n/locales/<locale>/kernel.yaml). Always the first, lowest
	// precedence layer; owns the reserved kernel.* namespace.
	KindFrameworkDefaults SourceKind = "framework_defaults"
	// KindFS is product-local catalog files on an fs.FS (YAML and/or JSON).
	// Used for both product framework-override files and product/module catalogs.
	KindFS SourceKind = "fs"
	// KindGo is a compiled Go catalog bundle: an in-process []RawBundle a
	// product assembles in Go (generated or hand-written) for compile-time
	// ownership. Highest static precedence layer.
	KindGo SourceKind = "go"
	// KindDBOverlay is reserved for a future opt-in database overlay (B13). No
	// built-in implementation ships today; the contract reserves the kind and
	// the final precedence slot so an overlay can be added without a breaking
	// change. Decision 3: catalogs freeze at boot by default; the overlay is a
	// separate opt-in concern.
	KindDBOverlay SourceKind = "db_overlay"
)

// RawBundle is one locale's worth of messages a Source yields, tagged with the
// origin (file path, "<embedded>", "<go>") so validation and merge errors name
// exactly where a bad key came from. It is deliberately close to Bundle but
// carries provenance and is never namespace-checked by the producer — the
// Loader applies the layer's ownership policy centrally.
type RawBundle struct {
	// Locale is the BCP 47 tag these messages are written in (e.g. "en", "mr").
	Locale string
	// Messages maps fully-qualified message keys to translated text.
	Messages map[string]string
	// Origin is a human-readable provenance label used only in error messages
	// (e.g. "locales/mr/kernel.yaml", "<embedded framework defaults>").
	Origin string
}

// Source loads one or more RawBundles for a single precedence layer. It is the
// extension point of the subsystem: framework defaults, product fs files, and
// compiled Go bundles all implement it, and a future DB overlay (B13) can add
// another implementation without changing the Loader or the Catalog. A Source
// performs I/O and parsing only; it does not enforce namespace ownership — the
// Loader does that centrally against the layer's Policy, so every source kind
// gets identical, tested ownership rules.
type Source interface {
	// Kind reports which first-class source kind this is (for diagnostics and
	// config round-tripping).
	Kind() SourceKind
	// Load reads and parses the source into per-locale bundles. It returns an
	// error only for I/O or parse failures (unreadable file, malformed YAML/JSON,
	// duplicate key within a single file); ownership and cross-source duplicate
	// checks are the Loader's job.
	Load() ([]RawBundle, error)
}

// Policy declares what a precedence layer's sources are permitted to write. It
// makes the precedence rules explicit and testable rather than implied by load
// order alone.
type Policy struct {
	// OwnsFramework lets this layer write keys in the reserved kernel.* namespace.
	// Only the framework-defaults layer and a sanctioned product framework-override
	// layer set this true.
	OwnsFramework bool
	// FrameworkOverrideOnly, when set together with OwnsFramework, means the layer
	// may only OVERRIDE kernel.* keys that a lower layer already defined — it may
	// not introduce brand-new kernel.* keys. This is the product override contract:
	// a product may retranslate a framework string but may not invent framework
	// strings. Ignored unless OwnsFramework is true.
	FrameworkOverrideOnly bool
}

// Layer is one precedence tier: an ordered set of sources plus the ownership
// policy applied to every key they contribute. Layers are merged in the order
// passed to LoadCatalog; a later layer overriding an earlier layer's key is
// allowed (that is the whole point of precedence), but two sources WITHIN one
// layer defining the same (locale, key) is a conflict and fails validation.
type Layer struct {
	// Name is a short label used in error messages ("framework defaults",
	// "product overrides", "catalogs", "go bundles").
	Name string
	// Policy is the ownership rule applied to every key in this layer.
	Policy Policy
	// Sources are the sources that make up this layer, loaded in order. Their
	// contributions are pooled and intra-layer duplicates are rejected.
	Sources []Source
}

// LoadCatalog builds a frozen Catalog by merging layers in precedence order:
// earlier layers first, later layers overriding. def is the default/fallback
// locale (always "en" for the framework). It enforces, per the documented
// precedence rules:
//
//   - intra-layer duplicate keys (same locale+key from two sources in ONE layer)
//     fail — a hard authoring error;
//   - namespace ownership: only a layer whose Policy.OwnsFramework is set may
//     write kernel.* keys, and a FrameworkOverrideOnly layer may only override
//     kernel.* keys a lower layer already defined (never introduce new ones);
//   - a later layer overriding an earlier layer's key is allowed (precedence).
//
// All violations across all layers are accumulated and returned as one error so
// a single run surfaces every problem (mirrors the seeds/registry boot pattern).
// On success the returned Catalog is ready for request-time reads; callers seal
// it with Freeze once boot completes.
func LoadCatalog(def string, layers ...Layer) (*Catalog, error) {
	cat := NewCatalog(def)
	if err := mergeLayers(cat, layers); err != nil {
		return nil, err
	}
	return cat, nil
}

// mergeLayers folds layers into an existing catalog in precedence order. It is
// shared by LoadCatalog (fresh catalog) and Registry.ApplyLayers (a catalog
// already carrying framework defaults + module bundles), so both get identical
// ownership and duplicate semantics. Keys already present in cat from a PRIOR
// call count as "lower layers" for override-only ownership checks.
func mergeLayers(cat *Catalog, layers []Layer) error {
	var errs []string

	for _, layer := range layers {
		// seen tracks (locale,key) already written *in this layer* to detect
		// intra-layer duplicates; it resets per layer so a later layer may
		// legitimately override.
		seen := make(map[string]string) // locale\x00key -> origin
		for _, src := range layer.Sources {
			bundles, err := src.Load()
			if err != nil {
				errs = append(errs, fmt.Sprintf("%s: %v", layer.Name, err))
				continue
			}
			for _, b := range bundles {
				if b.Locale == "" {
					errs = append(errs, fmt.Sprintf("%s: bundle from %s has no locale", layer.Name, originOf(b)))
					continue
				}
				for _, key := range sortedKeys(b.Messages) {
					if verr := checkOwnership(layer, cat, key, originOf(b)); verr != "" {
						errs = append(errs, verr)
						continue
					}
					sk := b.Locale + "\x00" + key
					if prev, dup := seen[sk]; dup {
						errs = append(errs, fmt.Sprintf(
							"%s: duplicate key %q for locale %q in the same layer (from %s and %s)",
							layer.Name, key, b.Locale, prev, originOf(b),
						))
						continue
					}
					seen[sk] = originOf(b)
					cat.Add(b.Locale, key, b.Messages[key])
				}
			}
		}
	}

	if len(errs) > 0 {
		sort.Strings(errs)
		return kerr.E(kerr.KindInternal, "invalid_i18n_catalog",
			"i18n catalog load failed: "+strings.Join(errs, "; "))
	}
	return nil
}

// checkOwnership returns a non-empty error string if key violates layer's
// namespace policy. Non-kernel.* keys are always allowed by the loader (module
// vs product namespace separation is enforced where those keys are produced —
// module Registry.Register — and by `wowapi i18n validate`); the loader guards
// only the reserved framework namespace, which is the security-sensitive one.
func checkOwnership(layer Layer, cat *Catalog, key, origin string) string {
	if !strings.HasPrefix(key, reservedPrefix) {
		return ""
	}
	if !layer.Policy.OwnsFramework {
		return fmt.Sprintf("%s: %s may not write reserved framework key %q (kernel.* is framework-owned; use a sanctioned framework-override source)",
			layer.Name, origin, key)
	}
	if layer.Policy.FrameworkOverrideOnly {
		// May only override a key a lower layer already defined.
		if _, exists := cat.messages[cat.def][key]; !exists {
			// Check any locale, not just default: the key must exist somewhere
			// lower. Framework defaults define every kernel.* key in en (the
			// default locale), so an override for a new locale still requires the
			// key to exist in the default locale — which it does for real keys.
			if !cat.hasKeyAnyLocale(key) {
				return fmt.Sprintf("%s: %s overrides framework key %q that no lower layer defines (a product may retranslate framework strings, not invent new kernel.* keys)",
					layer.Name, origin, key)
			}
		}
	}
	return ""
}

// hasKeyAnyLocale reports whether key exists under any locale already loaded.
func (c *Catalog) hasKeyAnyLocale(key string) bool {
	for _, m := range c.messages {
		if _, ok := m[key]; ok {
			return true
		}
	}
	return false
}

func originOf(b RawBundle) string {
	if b.Origin == "" {
		return "<unknown source>"
	}
	return b.Origin
}

func sortedKeys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
