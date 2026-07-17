package config_test

// Hot-path benchmarks for config value access (criterion #17).
//
// The Framework struct is loaded once at boot and then read on every request
// (e.g., http.MaxBodyBytes, DB.QueryTimeout). The benchmark demonstrates that
// reading a bound config value is a plain struct field read — O(1), zero
// allocations, nanoseconds — with no map or reflection lookup on the hot path.
//
// This is the correctness property criterion #17 requires: hot paths are free
// of reflection/registry lookups.

import (
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/config"
)

// BenchmarkConfigFieldReadMaxBodyBytes shows that reading MaxBodyBytes (the
// guard applied on every inbound HTTP request) is a single field dereference.
// Expected: ~1 ns/op, 0 allocs/op.
func BenchmarkConfigFieldReadMaxBodyBytes(b *testing.B) {
	cfg := config.Defaults()
	b.ReportAllocs()
	b.ResetTimer()
	var sink int64
	for i := 0; i < b.N; i++ {
		sink = cfg.HTTP.MaxBodyBytes
	}
	_ = sink
}

// BenchmarkConfigFieldReadQueryTimeout shows that reading QueryTimeout (applied
// per-DB call) is a field dereference with no runtime type dispatch.
// Expected: ~1 ns/op, 0 allocs/op.
func BenchmarkConfigFieldReadQueryTimeout(b *testing.B) {
	cfg := config.Defaults()
	b.ReportAllocs()
	b.ResetTimer()
	var sink int64
	for i := 0; i < b.N; i++ {
		sink = int64(cfg.DB.QueryTimeout)
	}
	_ = sink
}

// BenchmarkConfigDefaults measures Defaults() construction: called once at
// boot (or in tests per call). Establishes the cost of the bottom-layer build.
func BenchmarkConfigDefaults(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.Defaults()
	}
}

// BenchmarkConfigValidate measures Validate() on a fully-populated config:
// called once at boot. Not on the hot path but must not be surprisingly slow
// since it blocks the process from serving traffic.
func BenchmarkConfigValidate(b *testing.B) {
	cfg := config.Defaults()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cfg.Validate()
	}
}
