// Trace/log correlation (FBL-06 T1/T2): the vendor-neutral bridge between the
// tracing port and slog. ContextWithSpan/SpanFromContext carry the active port
// Span through a context.Context, and NewCorrelatingHandler wraps any
// slog.Handler so records emitted inside an active span carry trace_id/span_id
// attrs — and carry nothing at all (no empty-string noise) without one.

package observability

import (
	"context"
	"log/slog"

	"github.com/qatoolist/wowapi/v2/kernel/tracing"
)

// ContextWithSpan returns a context carrying s, retrievable via SpanFromContext.
// Defined in kernel/tracing (the leaf port package); forwarded here so
// adapters and composition roots can keep binding through observability.
// Real Tracer adapters call this from StartSpan so the active span is
// reachable vendor-neutrally at any downstream call site; NoOpTracer
// deliberately does not (its StartSpan returns ctx unchanged, keeping
// disabled tracing allocation-free).
func ContextWithSpan(ctx context.Context, s Span) context.Context {
	return tracing.ContextWithSpan(ctx, s)
}

// SpanFromContext returns the Span stored by ContextWithSpan and whether one
// was present. It never returns a non-nil Span with ok == false.
func SpanFromContext(ctx context.Context) (Span, bool) {
	return tracing.SpanFromContext(ctx)
}

// correlatingHandler decorates an slog.Handler with trace/log correlation.
type correlatingHandler struct{ inner slog.Handler }

// NewCorrelatingHandler wraps h so that every record whose context carries an
// active span (per SpanFromContext) with a non-empty TraceID gains trace_id and
// span_id attrs before delegation. Records with no active span — including the
// NoOpTracer path, whose StartSpan stores nothing in ctx — are delegated
// verbatim: the keys are genuinely absent, never present with empty values,
// and the pass-through adds no allocation.
//
// The wrapper is stateless and safe for concurrent use whenever h is.
// kernel/logging.New applies it to every logger it constructs, so correlation
// is a structural property of "was a span active", not a config flag.
func NewCorrelatingHandler(h slog.Handler) slog.Handler {
	return correlatingHandler{inner: h}
}

func (c correlatingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return c.inner.Enabled(ctx, level)
}

func (c correlatingHandler) Handle(ctx context.Context, r slog.Record) error {
	if s, ok := SpanFromContext(ctx); ok {
		if traceID := s.TraceID(); traceID != "" {
			r = r.Clone()
			r.AddAttrs(
				slog.String("trace_id", traceID),
				slog.String("span_id", s.SpanID()),
			)
		}
	}
	return c.inner.Handle(ctx, r)
}

func (c correlatingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return correlatingHandler{inner: c.inner.WithAttrs(attrs)}
}

func (c correlatingHandler) WithGroup(name string) slog.Handler {
	return correlatingHandler{inner: c.inner.WithGroup(name)}
}
