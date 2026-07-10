package i18n

// GoSource is a compiled Go catalog bundle: an in-process set of RawBundles a
// product assembles in Go (generated code or hand-written) and passes to the
// loader. It is the Go-native, compile-time-owned equivalent of the YAML/JSON
// catalog files — a product package exports these bundles and the composition
// root feeds them into the Go layer of LoadCatalog.
//
// Idiomatic shape: a product's internal/i18n/catalogs package exposes a function
// returning []i18n.RawBundle (or an i18n.Source built from them via NewGoSource),
// which the generated cmd mains pass through config-driven wiring. Because it is
// plain Go, the compiler enforces that the bundles exist and typecheck.
type GoSource struct {
	bundles []RawBundle
}

// NewGoSource wraps compiled RawBundles as a Source. Each bundle should set a
// stable Origin (e.g. "internal/i18n/catalogs/en.go") for clear validation
// errors. The loader applies the configured layer's ownership policy to these
// keys exactly as it does for fs sources.
func NewGoSource(bundles ...RawBundle) *GoSource {
	return &GoSource{bundles: bundles}
}

// Kind reports KindGo.
func (g *GoSource) Kind() SourceKind { return KindGo }

// Load returns the compiled bundles. It never errors — the bundles are already
// in memory and typechecked; ownership/duplicate validation is the loader's job.
func (g *GoSource) Load() ([]RawBundle, error) {
	out := make([]RawBundle, 0, len(g.bundles))
	for _, b := range g.bundles {
		if b.Origin == "" {
			b.Origin = "<go bundle>"
		}
		out = append(out, b)
	}
	return out, nil
}
