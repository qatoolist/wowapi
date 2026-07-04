// Package otel adapts OpenTelemetry to the wowapi observability.Tracer port
// (roadmap O1). The kernel depends only on the port and its zero-cost NoOp; this
// adapter is wired by a product that wants real distributed traces, exactly as
// adapters/metrics/prometheus binds the Metrics port. It carries W3C trace
// context (traceparent) for cross-process propagation (API → relay → worker) and
// samples by a configurable ratio.
package otel

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/qatoolist/wowapi/kernel/observability"
)

// Tracer implements observability.Tracer over an OTel SDK TracerProvider.
type Tracer struct {
	provider *sdktrace.TracerProvider
	tracer   oteltrace.Tracer
	prop     propagation.TextMapPropagator
}

// New builds a Tracer over an OTel TracerProvider with a parent-based ratio
// sampler (a sampled parent trace is always followed; otherwise sampleRatio of
// new traces are kept). sampleRatio is clamped to [0,1] — this is the
// configurable sampling the platform exposes. exporter is the product's span
// exporter (OTLP, stdout, in-memory in tests). Call Shutdown to flush on exit.
func New(exporter sdktrace.SpanExporter, sampleRatio float64) *Tracer {
	var sampler sdktrace.Sampler
	switch {
	case sampleRatio <= 0:
		sampler = sdktrace.NeverSample()
	case sampleRatio >= 1:
		sampler = sdktrace.AlwaysSample()
	default:
		sampler = sdktrace.TraceIDRatioBased(sampleRatio)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter), // non-blocking; flush via ForceFlush/Shutdown
		sdktrace.WithSampler(sdktrace.ParentBased(sampler)),
	)
	return &Tracer{
		provider: tp,
		tracer:   tp.Tracer("github.com/qatoolist/wowapi"),
		prop:     propagation.TraceContext{},
	}
}

// StartSpan begins an OTel span as a child of any span in ctx.
func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, observability.Span) {
	ctx, span := t.tracer.Start(ctx, name)
	return ctx, otelSpan{span: span}
}

// Inject returns the W3C traceparent for the span in ctx (embed it in an outgoing
// event/job so a downstream process continues the trace).
func (t *Tracer) Inject(ctx context.Context) string {
	c := propagation.MapCarrier{}
	t.prop.Inject(ctx, c)
	return c["traceparent"]
}

// Extract returns a context continuing the trace named by carrier (a traceparent
// from an inbound request/event/job). ctx is returned unchanged when carrier is "".
func (t *Tracer) Extract(ctx context.Context, carrier string) context.Context {
	if carrier == "" {
		return ctx
	}
	return t.prop.Extract(ctx, propagation.MapCarrier{"traceparent": carrier})
}

// NewOTLP is the batteries-included constructor: it builds an OTLP-over-HTTP span
// exporter — auto-configured from the standard OTEL_EXPORTER_OTLP_ENDPOINT env var
// (e.g. http://jaeger:4318 in the compose stack) — and returns a Tracer sampling
// at ratio. The exporter connects lazily, so a collector that is not yet up does
// not fail construction.
func NewOTLP(ctx context.Context, sampleRatio float64) (*Tracer, error) {
	exp, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, err
	}
	return New(exp, sampleRatio), nil
}

// ForceFlush exports any buffered spans immediately (useful before a short-lived
// process exits, or in tests).
func (t *Tracer) ForceFlush(ctx context.Context) error { return t.provider.ForceFlush(ctx) }

// Shutdown flushes buffered spans and stops the provider; call it during graceful
// process shutdown.
func (t *Tracer) Shutdown(ctx context.Context) error { return t.provider.Shutdown(ctx) }

type otelSpan struct{ span oteltrace.Span }

func (s otelSpan) End() { s.span.End() }

func (s otelSpan) SetAttr(k, v string) { s.span.SetAttributes(attribute.String(k, v)) }

func (s otelSpan) RecordError(err error) {
	if err == nil {
		return
	}
	s.span.RecordError(err)
	s.span.SetStatus(codes.Error, err.Error())
}

// Compile-time assurance the adapter satisfies the port.
var _ observability.Tracer = (*Tracer)(nil)
