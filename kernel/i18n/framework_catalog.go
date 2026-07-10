package i18n

// installFramework merges the framework's embedded per-locale YAML defaults into
// cat under the reserved kernel.* namespace. Called once by NewRegistry so every
// catalog the framework hands out already localizes problem titles, validation
// messages, and the framework's own well-known problem details.
//
// The strings themselves live in kernel/i18n/locales/<locale>/kernel.yaml and
// are compiled in via go:embed (embed.go) — they are NOT hardcoded Go maps
// anymore (B1). The framework_catalog_golden_test.go golden test asserts the
// loaded values are byte-identical to the historical hardcoded maps, so zero-
// config products localize in English exactly as before.
//
// A malformed embedded file is a programming error in the framework itself, not
// a product misconfiguration, so any load error panics: it can only happen if a
// developer breaks the shipped YAML, and it must fail every build's tests
// loudly rather than silently ship an empty framework catalog.
func installFramework(cat *Catalog) {
	bundles, err := FrameworkDefaultsSource().Load()
	if err != nil {
		panic("i18n: embedded framework catalog is malformed: " + err.Error())
	}
	for _, b := range bundles {
		for key, msg := range b.Messages {
			cat.Add(b.Locale, key, msg)
		}
	}
}
