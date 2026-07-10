package i18n

import (
	"fmt"
	"sort"
	"strings"
)

// ValidationReport is the outcome of Validate: the set of authoring defects
// found across a product's configured catalog sources, plus the coverage stats
// a CI check reports on success. A zero-length Problems slice means the catalog
// is valid.
type ValidationReport struct {
	// Locales is the sorted set of locales any source contributed to.
	Locales []string
	// Keys is the total number of distinct message keys across all locales.
	Keys int
	// Problems is the sorted list of human-readable defects. Empty == valid.
	Problems []string
}

// OK reports whether the catalog passed (no problems).
func (r ValidationReport) OK() bool { return len(r.Problems) == 0 }

// ValidateOptions configures Validate.
type ValidateOptions struct {
	// DefaultLocale is the fallback locale; every key must exist here (a key that
	// exists only in a non-default locale has no fallback and is a coverage hole).
	DefaultLocale string
	// SupportedLocales is the set of locales the product declares it supports.
	// Every key present in any locale must also be present in each supported
	// locale, OR resolvable via the default-locale fallback. Because Lookup always
	// falls back to DefaultLocale, the coverage rule is: a key missing from a
	// supported locale is a WARNING-level gap reported as a problem only when it is
	// also missing from the default locale (a total miss). We report BOTH: a hard
	// error for a key absent from the default locale, and a per-locale coverage
	// gap for a key present in the default locale but missing from a supported one.
	SupportedLocales []string
	// StrictCoverage, when true, promotes per-locale coverage gaps (key present in
	// default but missing from a supported locale) to hard problems. Default false:
	// the fallback makes them non-fatal, but `wowapi i18n validate --strict` (or
	// product CI) can require full coverage.
	StrictCoverage bool
}

// Validate loads the given layers WITHOUT building a servable catalog and checks
// the four defect classes the benchmark requires:
//
//   - namespace ownership + intra-layer duplicates (delegated to the loader, so
//     validate and boot agree exactly);
//   - locale coverage: every key present in the default locale; optionally every
//     key present in every supported locale (StrictCoverage);
//   - placeholder compatibility: a translation's %-verb count must match the
//     default-locale template's, so a localized min/max message can't drop or add
//     a parameter and render wrong.
//
// It never mutates global state and returns a ValidationReport; the CLI turns a
// non-OK report into a non-zero exit. layers must include the framework defaults
// layer first (the CLI supplies it) so kernel.* coverage is checked too.
func Validate(opts ValidateOptions, layers ...Layer) (ValidationReport, error) {
	def := opts.DefaultLocale
	if def == "" {
		def = DefaultLocale
	}

	// Reuse the loader's ownership + duplicate enforcement: a load error is a
	// real defect, surfaced as problems (split on "; " so each is its own line).
	cat, loadErr := LoadCatalog(def, layers...)
	var problems []string
	if loadErr != nil {
		msg := loadErr.Error()
		if i := strings.Index(msg, "load failed: "); i >= 0 {
			msg = msg[i+len("load failed: "):]
		}
		problems = append(problems, strings.Split(msg, "; ")...)
	}

	// If the load failed hard we may not have a catalog to inspect further.
	if cat == nil {
		sort.Strings(problems)
		return ValidationReport{Problems: problems}, nil
	}

	// Collect the universe of keys and per-locale presence.
	allKeys := map[string]bool{}
	for _, loc := range cat.Locales() {
		for k := range cat.messages[loc] {
			allKeys[k] = true
		}
	}

	// Coverage: every key must exist in the default locale (else no fallback).
	for key := range allKeys {
		if _, ok := cat.messages[def][key]; !ok {
			problems = append(problems, fmt.Sprintf("key %q is missing from the default locale %q (no fallback — a non-default locale defines it but %q does not)", key, def, def))
		}
	}

	// Per-supported-locale coverage (fallback makes these non-fatal unless strict).
	for _, loc := range opts.SupportedLocales {
		if loc == def {
			continue
		}
		for key := range allKeys {
			if _, ok := cat.messages[loc][key]; ok {
				continue
			}
			// Missing from this supported locale. If it's also missing from default
			// it's already reported above as a hard miss; skip the dup.
			if _, inDef := cat.messages[def][key]; !inDef {
				continue
			}
			gap := fmt.Sprintf("locale %q is missing key %q (falls back to %q)", loc, key, def)
			if opts.StrictCoverage {
				problems = append(problems, gap)
			}
		}
	}

	// Placeholder compatibility: compare each localized value's %-verb signature
	// against the default-locale template for the same key.
	for key := range cat.messages[def] {
		want := placeholderCount(cat.messages[def][key])
		for _, loc := range cat.Locales() {
			if loc == def {
				continue
			}
			val, ok := cat.messages[loc][key]
			if !ok {
				continue
			}
			if got := placeholderCount(val); got != want {
				problems = append(problems, fmt.Sprintf(
					"placeholder mismatch for key %q in locale %q: default %q has %d %%-verb(s), translation %q has %d",
					key, loc, cat.messages[def][key], want, val, got))
			}
		}
	}

	sort.Strings(problems)
	return ValidationReport{
		Locales:  cat.Locales(),
		Keys:     len(allKeys),
		Problems: problems,
	}, nil
}

// placeholderCount counts Go %-verbs in a template, treating "%%" as a literal
// percent (zero verbs). It is intentionally simple: the framework's own
// parameterised messages use only "%s", and the check only needs the arity to
// match between a template and its translation, not the exact verb.
func placeholderCount(s string) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			continue
		}
		if i+1 < len(s) && s[i+1] == '%' {
			i++ // escaped literal percent
			continue
		}
		n++
	}
	return n
}
