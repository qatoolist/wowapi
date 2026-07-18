// Package sealer holds the authority token app.Boot passes to every extension
// registry's Seal method (second closure audit 2026-07-17, F-10).
//
// Scope of the guarantee (explicitly narrower than "unforgeable"): the
// boundary is OUT-OF-MODULE. Product code — including modules running inside
// Register — cannot import an internal package, so it cannot obtain the
// Authority type at all and cannot prematurely seal a shared registry so that
// later legitimate module or seed registration panics instead of surfacing a
// collected boot-validation error. IN-repository code can construct
// Authority{} (Go permits zero-value composite literals of types with only
// unexported fields), so within this module the restriction is a reviewed
// convention: app.Boot is the only caller.
package sealer

// Authority authorizes sealing the extension model. Constructible only inside
// the wowapi module via Grant.
type Authority struct{ _ byte }

// Grant mints the sealing authority for the boot path.
func Grant() Authority { return Authority{} }
