package artifact

import (
	"sort"
	"time"
)

// TemplateVersion identifies a template revision and the instant from which it
// takes effect. Artifacts record which version they were produced under, so a
// document can always be regenerated or explained against the rules that applied
// on its effective date (roadmap E4: "templates versioned by effective date").
type TemplateVersion struct {
	Version       string
	EffectiveFrom time.Time
}

// Templates resolves which template version applies at a given date. It is an
// in-memory registry populated at boot (the versions themselves — the actual
// rendering assets — live wherever the product keeps them).
type Templates struct {
	byKind map[string][]TemplateVersion
}

// NewTemplates builds an empty registry.
func NewTemplates() *Templates {
	return &Templates{byKind: map[string][]TemplateVersion{}}
}

// Register adds a template version for a kind, effective from the given instant.
// Registering the same version again updates its effective date.
func (t *Templates) Register(kind, version string, effectiveFrom time.Time) {
	vers := t.byKind[kind]
	for i := range vers {
		if vers[i].Version == version {
			vers[i].EffectiveFrom = effectiveFrom
			t.byKind[kind] = vers
			return
		}
	}
	t.byKind[kind] = append(vers, TemplateVersion{Version: version, EffectiveFrom: effectiveFrom})
}

// Resolve returns the template version in effect for kind at instant `at` — the
// one with the latest EffectiveFrom not after `at`. ok is false when the kind is
// unknown or every version begins after `at`.
func (t *Templates) Resolve(kind string, at time.Time) (TemplateVersion, bool) {
	vers := t.byKind[kind]
	if len(vers) == 0 {
		return TemplateVersion{}, false
	}
	// Sort by EffectiveFrom ascending, then pick the last one <= at.
	sorted := make([]TemplateVersion, len(vers))
	copy(sorted, vers)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].EffectiveFrom.Before(sorted[j].EffectiveFrom) })
	var chosen TemplateVersion
	found := false
	for _, v := range sorted {
		if !v.EffectiveFrom.After(at) {
			chosen, found = v, true
		}
	}
	return chosen, found
}
