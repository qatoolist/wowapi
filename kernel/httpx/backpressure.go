package httpx

import (
	"math"
	"net/http"
	"strconv"

	"github.com/qatoolist/wowapi/kernel/config"
	kerr "github.com/qatoolist/wowapi/kernel/errors"
)

// Backpressure (backlog B6, benchmark §Concurrency) is a bounded-semaphore
// in-flight limiter: it rejects requests with the configured overload status
// + Retry-After BEFORE they reach the handler (and therefore before they can
// touch the DB pool), once more than maxInFlight requests are concurrently
// in-flight through this middleware instance. It is the edge-of-process half
// of the capacity budget the sibling config.Concurrency / CheckCapacity model
// on the deployment-shape side: the limiter caps THIS process's concurrent
// work so it can never fan out more DB/downstream work than the declared
// pool budgets assume.
//
// maxInFlight <= 0 disables the limiter entirely (plain pass-through, no
// semaphore allocated) — this is config.Concurrency's HTTPMaxInFlight default
// of 0, so a product that has not opted in sees no behavior change and can
// never start returning the overload status unexpectedly (backlog B6 rollout
// guard).
//
// Position in the chain: near the edge, before auth/DB work — after
// RequestID/Recover/SecureHeaders/CORS (so a rejected request still carries
// correlation id, security headers, and CORS headers) but before BodyLimit,
// the authz gate, Locale, and any handler that touches the database. See
// internal/cli/templates/init/cmd_api_main.go.tmpl for the wired order.
func Backpressure(maxInFlight int, overload config.Overload, opts ...BackpressureOption) Middleware {
	if maxInFlight <= 0 {
		return func(next http.Handler) http.Handler { return next }
	}

	var cfg backpressureCfg
	for _, o := range opts {
		o(&cfg)
	}

	sem := make(chan struct{}, maxInFlight)
	status := overload.Status
	if status == 0 {
		status = http.StatusServiceUnavailable
	}
	retrySecs := max(int(math.Ceil(overload.RetryAfter.Seconds())), 1)
	retryHeader := strconv.Itoa(retrySecs)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			select {
			case sem <- struct{}{}:
				if cfg.onInFlightChange != nil {
					cfg.onInFlightChange(len(sem))
				}
				defer func() {
					<-sem
					if cfg.onInFlightChange != nil {
						cfg.onInFlightChange(len(sem))
					}
				}()
				next.ServeHTTP(w, r)
			default:
				w.Header().Set("Retry-After", retryHeader)
				if cfg.onOverload != nil {
					cfg.onOverload(r.Pattern)
				}
				writeOverload(r, w, status)
			}
		})
	}
}

// BackpressureOption customizes the Backpressure middleware.
type BackpressureOption func(*backpressureCfg)

type backpressureCfg struct {
	onOverload       func(route string)
	onInFlightChange func(n int)
}

// OnBackpressureOverload registers a callback fired whenever a request is
// rejected for being over the in-flight cap. The composition root wires this
// to a metrics counter (rejected-overload count) — httpx must not import
// kernel/observability (observability imports httpx), so emission is
// injected as a plain callback, mirroring OnRateLimitDrop. route is
// r.Pattern.
func OnBackpressureOverload(fn func(route string)) BackpressureOption {
	return func(c *backpressureCfg) { c.onOverload = fn }
}

// OnInFlightChange registers a callback fired with the current in-flight
// count every time a request enters or leaves the limiter — the wiring point
// for an in-flight gauge metric. It is called synchronously on the hot path,
// so the callback must be cheap (a gauge Set, not I/O).
func OnInFlightChange(fn func(n int)) BackpressureOption {
	return func(c *backpressureCfg) { c.onInFlightChange = fn }
}

// writeOverload writes the configured overload status as an RFC 9457
// problem-details body. It does not go through WriteError/kerr.E because the
// overload status is a deploy-time config choice (503 default or 429), not a
// fixed Kind→status mapping from the error taxonomy — but the body shape is
// identical (WriteError's ProblemError), so clients that already parse
// wowapi's problem+json bodies handle it exactly like any other response.
// The title/detail are plain English (not run through localizeTitle/
// localizeDetail): the error taxonomy has no dedicated overload Kind to key
// an i18n catalog lookup off of, and adding one is out of this backlog item's
// scope (kernel/i18n and kernel/errors are untouched here).
func writeOverload(r *http.Request, w http.ResponseWriter, status int) {
	code := "overloaded"
	title := "Service overloaded"
	if status == http.StatusTooManyRequests {
		code = kerr.KindRateLimited.DefaultCode()
		title = "Rate limited"
	}
	writeProblem(w, ProblemError{
		Type:      problemTypeBase + code,
		Title:     title,
		Status:    status,
		Code:      code,
		Detail:    "the service is at capacity; retry after the indicated delay",
		RequestID: RequestIDFrom(r.Context()),
	})
}
