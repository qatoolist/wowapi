package outbox

import "testing"

// Relay comparability is part of the frozen v1 public API surface (the Go API
// compatibility gate flagged its loss when test seams were added as func
// fields — kept behind a comparable pointer instead). This is a compile-time
// regression: it fails to build if Relay ever becomes non-comparable again.
func TestRelayComparable(t *testing.T) {
	_ = Relay{} == Relay{}
}
