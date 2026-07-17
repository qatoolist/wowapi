package i18n

import (
	"io/fs"
	"sort"
	"strings"

	kerr "github.com/qatoolist/wowapi/v2/kernel/errors"
)

// SourceSpec is a config-neutral description of one configured catalog source,
// the bridge between a product's i18n config section (parsed in the product's
// appcfg package) and the framework loader. It mirrors the scaffolded config
// shape (kind/path/formats/overrides_framework/enabled) without the framework
// depending on any product config type, keeping kernel/i18n a leaf package.
type SourceSpec struct {
	// Kind selects the source: "framework_defaults", "fs", "go", or "db_overlay".
	Kind SourceKind
	// Path is the fs root for a KindFS source (e.g. "locales"), relative to Root.
	Path string
	// Formats limits which file extensions a KindFS source reads ("yaml","json").
	// Empty means both.
	Formats []string
	// OverridesFramework marks a KindFS (or KindGo) source as a sanctioned
	// framework-override layer: it may retranslate kernel.* keys (but not invent
	// them). Product/module catalog sources leave this false.
	OverridesFramework bool
	// Enabled gates the source; a disabled source contributes nothing (used for
	// the scaffolded-but-off go/db_overlay stubs).
	Enabled bool
	// Go supplies compiled bundles for a KindGo source.
	Go []RawBundle
}

// BuildLayers turns a product's ordered source specs into loader Layers in the
// canonical precedence: framework defaults (implicit, always first) → product
// framework-override fs/go sources → product/module catalog fs sources → Go
// bundle sources. It groups specs into the right layer by kind and
// OverridesFramework so the generated api/worker/migrate binaries can hand the
// result straight to app.Boot(app.WithI18nLayers(...)).
//
// root is the product filesystem the KindFS paths resolve against (os.DirFS at
// the product root). The framework-defaults source is NOT emitted here — the
// registry already installs it; these layers stack on top. A db_overlay spec is
// rejected today (no built-in overlay ships; B13), so a product that enables one
// fails loudly rather than silently getting nothing.
func BuildLayers(root fs.FS, specs []SourceSpec) ([]Layer, error) {
	var overrideSources, catalogSources, goSources []Source
	var errs []string

	for _, s := range specs {
		if !s.Enabled {
			continue
		}
		switch s.Kind {
		case KindFrameworkDefaults:
			// Always installed by the registry; a config entry for it is a no-op
			// (tolerated so the scaffolded config can list it for clarity).
		case KindFS:
			src := NewFSSource(root, s.Path, s.Formats...)
			if s.OverridesFramework {
				overrideSources = append(overrideSources, src)
			} else {
				catalogSources = append(catalogSources, src)
			}
		case KindGo:
			src := NewGoSource(s.Go...)
			goSources = append(goSources, src)
		case KindDBOverlay:
			errs = append(errs, "db_overlay source is not supported yet (opt-in overlay is a separate concern, B13); disable it")
		default:
			errs = append(errs, "unknown i18n source kind "+string(s.Kind))
		}
	}

	if len(errs) > 0 {
		sort.Strings(errs)
		return nil, kerr.E(kerr.KindInternal, "invalid_i18n_config", strings.Join(errs, "; "))
	}

	var layers []Layer
	if len(overrideSources) > 0 {
		layers = append(layers, Layer{
			Name:    "product framework overrides",
			Policy:  Policy{OwnsFramework: true, FrameworkOverrideOnly: true},
			Sources: overrideSources,
		})
	}
	if len(catalogSources) > 0 {
		layers = append(layers, Layer{Name: "product/module catalogs", Sources: catalogSources})
	}
	if len(goSources) > 0 {
		layers = append(layers, Layer{Name: "go bundles", Sources: goSources})
	}
	return layers, nil
}
