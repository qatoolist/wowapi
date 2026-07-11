package httpx

import (
	"strconv"
	"time"
)

// SweepForTest exposes TokenBucket's unexported sweep for the external
// httpx_test package (PERF-01 regression tests need to trigger a sweep
// directly, independent of Allow's opportunistic sweepAt threshold). Not
// part of the public API — this file only builds under `go test`.
func SweepForTest(tb *TokenBucket, now time.Time) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	tb.sweep(now)
}

// SeedBucketsForTest inserts n synthetic idle buckets (last touched at `at`)
// directly into the map, bypassing Allow. Populating via Allow is O(n²) once
// the map passes sweepAt (every insert triggers a full synchronous scan),
// which made the sweep benchmarks' setup dominate their own measurement.
func SeedBucketsForTest(tb *TokenBucket, n int, at time.Time) {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	for i := 0; i < n; i++ {
		tb.buckets["k"+strconv.Itoa(i)] = &tokenBucket{tokens: 0, last: at}
	}
}
