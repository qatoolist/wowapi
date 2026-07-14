// Package tracing is wowapi's vendor-neutral distributed-tracing port: the
// Tracer/Span interfaces, the zero-cost NoOpTracer default, and the
// context-carriage helpers (ContextWithSpan/SpanFromContext). It imports only
// the standard library, so ANY kernel package — including deep ones like
// kernel/database, which sits below the httpx/authz middleware stack — can
// consume the port without an import cycle.
//
// kernel/observability re-exports these types by alias and hosts the
// HTTP-facing pieces (the Trace middleware, the correlating slog handler);
// composition roots and adapters keep binding through observability. Vendor
// bindings (OpenTelemetry) live in adapters/tracing/*.
package tracing

import "context"

// Tracer is the distributed-tracing port (roadmap O1): the kernel depends only
// on this interface, and a vendor binding (OpenTelemetry) lives in
// adapters/tracing/*. NoOpTracer is the safe default so tracing is literally
// zero-cost when no adapter is wired. A span started here nests under any span
// already in ctx, so API → relay → worker traces connect once the adapter
// propagates context across process boundaries.
type Tracer interface {
	// StartSpan begins a span named `name` as a child of any span in ctx and
	// returns a context carrying it plus the Span to End. Real adapters must
	// return a context from which the span is retrievable via SpanFromContext;
	// NoOpTracer returns ctx unchanged.
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
//
// TraceID/SpanID give consumers (the log-correlation handler, the pgx query
// tracer) a vendor-neutral view of the active trace identifiers: a real adapter
// returns the canonical lowercase-hex string forms; the no-op span — and any
// span without a valid trace context — returns "" for both, so callers never
// need a nil check and can treat "" as "no correlation available".
type Span interface {
	End()
	// SetAttr attaches a low-cardinality key/value to the span.
	SetAttr(key, value string)
	// RecordError marks the span as errored and records the error.
	RecordError(err error)
	// TraceID returns the span's trace ID in canonical string form, or "" for
	// a no-op span or a span with no valid trace context.
	TraceID() string
	// SpanID returns the span's span ID in canonical string form, or "" for a
	// no-op span or a span with no valid trace context.
	SpanID() string
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
func (noopSpan) TraceID() string     { return "" }
func (noopSpan) SpanID() string      { return "" }

// spanCtxKey is the private context key under which ContextWithSpan stores the
// active port-level Span.
type spanCtxKey struct{}

// ContextWithSpan returns a context carrying s, retrievable via SpanFromContext.
// Real Tracer adapters call this from StartSpan so the active span is reachable
// vendor-neutrally at any downstream call site; NoOpTracer deliberately does
// not (its StartSpan returns ctx unchanged, keeping disabled tracing
// allocation-free).
func ContextWithSpan(ctx context.Context, s Span) context.Context {
	return context.WithValue(ctx, spanCtxKey{}, s)
}

// SpanFromContext returns the Span stored by ContextWithSpan and whether one
// was present. It never returns a non-nil Span with ok == false.
func SpanFromContext(ctx context.Context) (Span, bool) {
	s, ok := ctx.Value(spanCtxKey{}).(Span)
	return s, ok
}
