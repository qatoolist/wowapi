package i18n

import "embed"

// embeddedLocales holds the framework's own per-locale YAML defaults
// (locales/<locale>/kernel.yaml). They are compiled into every wowapi binary so
// the framework's English strings ship with zero product configuration. The
// FrameworkDefaultsSource reads this FS; a golden test asserts the parsed
// messages are byte-identical to the historical hardcoded Go maps.
//
//go:embed locales
var embeddedLocales embed.FS

// FrameworkDefaultsSource returns the always-present, lowest-precedence source:
// the framework's embedded per-locale YAML defaults under the reserved kernel.*
// namespace. It is the canonical implementation of KindFrameworkDefaults and the
// first Layer every product loads.
func FrameworkDefaultsSource() Source {
	return &fsSource{
		fsys:   embeddedLocales,
		root:   "locales",
		kind:   KindFrameworkDefaults,
		label:  "<embedded framework defaults>",
		yaml:   true,
		json:   false,
		labelP: true,
	} //nolint:exhaustruct // remaining fields default to zero intentionally
}
