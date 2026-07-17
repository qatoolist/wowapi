package tracing_test

import (
	"context"
	"errors"
	"testing"

	"github.com/qatoolist/wowapi/kernel/tracing"
)

// TestNoOpTracer_StartSpan asserts the documented zero-cost contract: the
// returned context is unchanged (identity-equal) and the returned Span is
// usable without panicking.
func TestNoOpTracer_StartSpan(t *testing.T) {
	ctx := context.WithValue(context.Background(), struct{ k string }{"probe"}, "v")

	gotCtx, span := tracing.NoOpTracer.StartSpan(ctx, "op")

	if gotCtx != ctx {
		t.Fatalf("NoOpTracer.StartSpan must return ctx unchanged")
	}
	if span == nil {
		t.Fatal("NoOpTracer.StartSpan returned a nil Span")
	}
}

// TestNoOpTracer_InjectExtract covers the cross-process carrier contract for
// the no-op tracer: Inject always yields "" and Extract always returns ctx
// unchanged, regardless of the carrier value.
func TestNoOpTracer_InjectExtract(t *testing.T) {
	ctx := context.Background()

	if got := tracing.NoOpTracer.Inject(ctx); got != "" {
		t.Fatalf("NoOpTracer.Inject() = %q, want \"\"", got)
	}

	for _, carrier := range []string{"", "00-trace-span-01"} {
		got := tracing.NoOpTracer.Extract(ctx, carrier)
		if got != ctx {
			t.Fatalf("NoOpTracer.Extract(ctx, %q) did not return ctx unchanged", carrier)
		}
	}
}

// TestNoOpSpan_Behavior asserts the no-op span never panics and always
// reports empty trace/span IDs, so downstream correlation code can treat ""
// as "no correlation available" without a nil check.
func TestNoOpSpan_Behavior(t *testing.T) {
	_, span := tracing.NoOpTracer.StartSpan(context.Background(), "op")

	span.SetAttr("key", "value")
	span.RecordError(errors.New("boom"))
	span.End()
	span.End() // must be safe to End more than once for the no-op span

	if got := span.TraceID(); got != "" {
		t.Errorf("noopSpan.TraceID() = %q, want \"\"", got)
	}
	if got := span.SpanID(); got != "" {
		t.Errorf("noopSpan.SpanID() = %q, want \"\"", got)
	}
}

// fakeSpan is a minimal Span used to test the context-carriage helpers
// independently of any Tracer implementation.
type fakeSpan struct {
	tracing.Span
	traceID string
}

func (f fakeSpan) TraceID() string { return f.traceID }

// TestContextWithSpan_RoundTrip covers ContextWithSpan/SpanFromContext: a
// span stored via ContextWithSpan must be retrievable via SpanFromContext,
// distinct from any other value carried on the same context.
func TestContextWithSpan_RoundTrip(t *testing.T) {
	want := fakeSpan{traceID: "abc123"}
	ctx := tracing.ContextWithSpan(context.Background(), want)

	got, ok := tracing.SpanFromContext(ctx)
	if !ok {
		t.Fatal("SpanFromContext returned ok = false after ContextWithSpan")
	}
	if got.TraceID() != want.traceID {
		t.Fatalf("SpanFromContext round-trip mismatch: got TraceID %q, want %q", got.TraceID(), want.traceID)
	}
}

// TestSpanFromContext_Absent asserts SpanFromContext never returns a
// non-nil Span with ok == false when no span was ever stored.
func TestSpanFromContext_Absent(t *testing.T) {
	got, ok := tracing.SpanFromContext(context.Background())
	if ok {
		t.Fatal("SpanFromContext returned ok = true on a bare context")
	}
	if got != nil {
		t.Fatalf("SpanFromContext returned non-nil Span (%v) with ok = false", got)
	}
}

// TestNoOpTracer_DoesNotCarrySpan documents that NoOpTracer.StartSpan
// deliberately does NOT call ContextWithSpan, keeping disabled tracing
// allocation-free: SpanFromContext must find nothing on its returned ctx.
func TestNoOpTracer_DoesNotCarrySpan(t *testing.T) {
	ctx, _ := tracing.NoOpTracer.StartSpan(context.Background(), "op")

	if _, ok := tracing.SpanFromContext(ctx); ok {
		t.Fatal("NoOpTracer.StartSpan must not store a span retrievable via SpanFromContext")
	}
}
