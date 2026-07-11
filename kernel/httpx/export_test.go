package httpx

import "time"

// SweepForTest exposes TokenBucket's unexported sweep for the external
// httpx_test package (PERF-01 regression tests need to trigger a sweep
// directly, independent of Allow's opportunistic sweepAt threshold). Not
// part of the public API — this file only builds under `go test`.
func SweepForTest(tb *TokenBucket, now time.Time) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.sweep(now)
}
