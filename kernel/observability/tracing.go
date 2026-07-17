package observability

import (
	"net/http"
	"strconv"

	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/tracing"
)

// Tracer is wowapi's distributed-tracing port (roadmap O1), a sibling of
// Metrics. The port itself is DEFINED in kernel/tracing — a stdlib-only leaf
// package, so deep kernel packages (e.g. kernel/database's pgx query tracer)
// can consume it without an import cycle through this package's httpx
// middleware. It is re-exported here by alias: composition roots and adapters
// keep binding observability.Tracer, and every value is interchangeable with
// tracing.Tracer. See kernel/tracing for the full contract docs.
type Tracer = tracing.Tracer

// Span is one unit of a trace — alias of tracing.Span; see kernel/tracing.
type Span = tracing.Span

// NoOpTracer is the safe-default Tracer: every method is a no-op and StartSpan
// returns ctx unchanged, so call sites never need a nil check and disabled
// tracing adds no allocation. Same value as tracing.NoOpTracer.
var NoOpTracer Tracer = tracing.NoOpTracer

// Trace returns a httpx.Middleware that opens a server span per request, tags it
// with the route/method/status/request-id, and ends it. The request context is
// replaced with one carrying the span so handler and downstream StartSpan calls
// nest under it. Zero-cost with NoOpTracer.
//
// Position: after RequestID (so the request id is available) and Recover.
func Trace(tr Tracer) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Continue an inbound distributed trace when the caller sent one.
			ctx := tr.Extract(r.Context(), r.Header.Get("traceparent"))
			ctx, span := tr.StartSpan(ctx, "HTTP "+r.Method)
			defer span.End()
			span.SetAttr("http.method", r.Method)
			if id := httpx.RequestIDFrom(ctx); id != "" {
				span.SetAttr("http.request_id", id)
			}
			sw := &statusWriter{ResponseWriter: w}
			rr := r.WithContext(ctx)
			next.ServeHTTP(sw, rr)
			// The mux populates Pattern on the request it dispatched (rr), so the
			// route is known now — attach it as a bounded-cardinality attr.
			span.SetAttr("http.route", routeLabel(rr.Pattern))
			span.SetAttr("http.status", strconv.Itoa(sw.statusCode()))
		})
	}
}
