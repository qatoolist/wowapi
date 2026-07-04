package otel_test

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	oteladapter "github.com/qatoolist/wowapi/adapters/tracing/otel"
)

func TestOtelTracerRecordsSpanAndPropagates(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tr := oteladapter.New(exp, 1.0) // sample everything

	ctx, span := tr.StartSpan(context.Background(), "op")
	span.SetAttr("http.route", "/x")
	span.RecordError(errors.New("boom"))

	// A traceparent is available for the active span (cross-process carrier).
	tp := tr.Inject(ctx)
	if tp == "" {
		t.Fatal("Inject returned empty traceparent for an active span")
	}

	span.End()
	// Batched: flush before reading (Shutdown would reset the in-memory exporter).
	if err := tr.ForceFlush(context.Background()); err != nil {
		t.Fatalf("force flush: %v", err)
	}
	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("recorded %d spans, want 1", len(spans))
	}
	if spans[0].Name != "op" {
		t.Errorf("span name = %q, want op", spans[0].Name)
	}

	// Extract continues the trace: a span started from the extracted context
	// shares the same trace id as the original.
	ext := tr.Extract(context.Background(), tp)
	if got := tr.Inject(ext); got != tp {
		t.Errorf("extract→inject round-trip = %q, want %q", got, tp)
	}
}

func TestNewOTLPConstructs(t *testing.T) {
	// The OTLP exporter connects lazily, so construction succeeds even with no
	// collector running (endpoint auto-configured from OTEL_EXPORTER_OTLP_ENDPOINT).
	tr, err := oteladapter.NewOTLP(context.Background(), 0.1)
	if err != nil {
		t.Fatalf("NewOTLP: %v", err)
	}
	_ = tr.Shutdown(context.Background())
}

func TestOtelSampleRatioClamped(t *testing.T) {
	// Out-of-range ratios must not panic (clamped to [0,1]).
	exp := tracetest.NewInMemoryExporter()
	_ = oteladapter.New(exp, -5)
	_ = oteladapter.New(exp, 42)
}
