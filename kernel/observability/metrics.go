// Package observability is wowapi's observability port: the Metrics interface
// (RED signals + generic counters and gauges) and a no-op safe default.
// Third-party client libraries live in adapters/metrics/*; this package
// imports only the standard library and kernel-siblings.
//
// Wiring (composition root):
//
//	var m observability.Metrics = observability.NoOp // default — no adapter wired
//	m = promadapter.New()                            // swap in the real adapter
//	httpx.Chain(handler, httpx.RequestID(), httpx.Recover(log),
//	    observability.Requests(m), observability.AccessLog(log))
package observability

import "time"

// Metrics is the framework's metric sink. Implementations must be safe for
// concurrent use. All methods are hot-path cheap: no reflection, no map
// allocation on the call site. NoOp is the safe default when no adapter is
// wired so call sites never need a nil check.
type Metrics interface {
	// ObserveRequest records one HTTP request (RED per route). dur is the
	// wall-clock duration from first byte received to response flushed.
	// respBytes is the number of bytes written to the response body.
	ObserveRequest(route, method string, status int, dur time.Duration, respBytes int)

	// IncCounter increments a named counter by value. labels provides
	// low-cardinality dimensions. Intended for: authz denials, rate-limit
	// drops, outbox dead letters, webhook breaker opens, notification
	// delivery failures (blueprint 07 §9).
	IncCounter(name string, value float64, labels map[string]string)

	// SetGauge sets a named gauge to value. Intended for: outbox_pending,
	// job queue depth, workflow open tasks, pool stats,
	// outbox_dispatch_lag_seconds (blueprint 07 §9).
	SetGauge(name string, value float64, labels map[string]string)
}

// NoOp is the safe-default Metrics implementation whose methods are all
// no-ops. Wire it when no adapter is configured so callers never check nil.
var NoOp Metrics = noOp{}

type noOp struct{}

func (noOp) ObserveRequest(_, _ string, _ int, _ time.Duration, _ int) {}
func (noOp) IncCounter(_ string, _ float64, _ map[string]string)       {}
func (noOp) SetGauge(_ string, _ float64, _ map[string]string)         {}
