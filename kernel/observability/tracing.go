package observability

import (
	"context"
	"net/http"
	"strconv"

	"github.com/qatoolist/wowapi/kernel/httpx"
)

// Tracer is wowapi's distributed-tracing port (roadmap O1), a sibling of Metrics:
// the kernel depends only on this interface, and a vendor binding (OpenTelemetry)
// lives in adapters/tracing/*. NoOpTracer is the safe default so tracing is
// literally zero-cost when no adapter is wired. A span started here nests under
// any span already in ctx, so API → relay → worker traces connect once the
// adapter propagates context across process boundaries.
type Tracer interface {
	// StartSpan begins a span named `name` as a child of any span in ctx and
	// returns a context carrying it plus the Span to End.
	StartSpan(ctx context.Context, name string) (context.Context, Span)
	// Inject returns the opaque cross-process carrier (a W3C traceparent) for the
	// span active in ctx, to embed in an outgoing event or job so a downstream
	// process continues the same trace. It returns "" when no span is active (or
	// for the NoOp tracer).
	Inject(ctx context.Context) string
	// Extract returns a context continuing the trace named by carrier (a
	// traceparent taken from an inbound request, event, or job). ctx is returned
	// unchanged when carrier is "".
	Extract(ctx context.Context, carrier string) context.Context
}

// Span is one unit of a trace. Implementations must be safe to End exactly once.
type Span interface {
	End()
	// SetAttr attaches a low-cardinality key/value to the span.
	SetAttr(key, value string)
	// RecordError marks the span as errored and records the error.
	RecordError(err error)
}

// NoOpTracer is the safe-default Tracer: every method is a no-op and StartSpan
// returns ctx unchanged, so call sites never need a nil check and disabled
// tracing adds no allocation.
var NoOpTracer Tracer = noopTracer{}

type noopTracer struct{}

func (noopTracer) StartSpan(ctx context.Context, _ string) (context.Context, Span) {
	return ctx, noopSpan{}
}
func (noopTracer) Inject(context.Context) string                         { return "" }
func (noopTracer) Extract(ctx context.Context, _ string) context.Context { return ctx }

type noopSpan struct{}

func (noopSpan) End()                {}
func (noopSpan) SetAttr(_, _ string) {}
func (noopSpan) RecordError(error)   {}

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
