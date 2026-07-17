// Package sealer holds the unforgeable authority token app.Boot passes to
// every extension registry's Seal method (second closure audit 2026-07-17,
// F-10). Because the package is internal to the wowapi module, product code —
// including modules running inside Register — cannot construct an Authority
// and therefore cannot prematurely seal a shared registry so that later
// legitimate module or seed registration panics instead of surfacing a
// collected boot-validation error. Sealing authority belongs to the boot path
// alone.
package sealer

// Authority authorizes sealing the extension model. Constructible only inside
// the wowapi module via Grant.
type Authority struct{ _ byte }

// Grant mints the sealing authority for the boot path.
func Grant() Authority { return Authority{} }
