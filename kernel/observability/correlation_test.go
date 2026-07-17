package observability_test

// correlation_test.go — W01-E02-S001 (FBL-06 T1/T2): the correlating handler's
// contract matrix (attrs present with an active span, keys genuinely absent
// without one), the Trace-middleware end-to-end path, and the
// allocation-neutrality benchmark for the no-op path (AC-W01-E02-S001-03).

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	oteladapter "github.com/qatoolist/wowapi/v2/adapters/tracing/otel"
	"github.com/qatoolist/wowapi/v2/kernel/observability"
)

// stubSpan is a minimal observability.Span double with fixed IDs, so the
// wrapper's behavior is testable independently of any tracing adapter.
type stubSpan struct{ trace, span string }

func (stubSpan) End()                {}
func (stubSpan) SetAttr(_, _ string) {}
func (stubSpan) RecordError(error)   {}
func (s stubSpan) TraceID() string   { return s.trace }
func (s stubSpan) SpanID() string    { return s.span }

func decodeLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var rec map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &rec); err != nil {
		t.Fatalf("log line is not valid JSON: %v\n%s", err, buf.String())
	}
	return rec
}

func TestSpanFromContextRoundtrip(t *testing.T) {
	s := stubSpan{trace: "t1", span: "s1"}
	ctx := observability.ContextWithSpan(context.Background(), s)
	got, ok := observability.SpanFromContext(ctx)
	if !ok || got != observability.Span(s) {
		t.Fatalf("SpanFromContext = %v/%v, want stored span", got, ok)
	}
	if _, ok := observability.SpanFromContext(context.Background()); ok {
		t.Fatal("empty context must not carry a span")
	}
}

// The correlation contract matrix: present both ways under an active span,
// absent both ways (key absence, not empty values) without one.
func TestCorrelatingHandlerMatrix(t *testing.T) {
	cases := map[string]struct {
		ctx       func() context.Context
		wantTrace string // "" means the keys must be ABSENT
		wantSpan  string
	}{
		"active span injects both attrs": {
			ctx: func() context.Context {
				return observability.ContextWithSpan(context.Background(),
					stubSpan{trace: "0123456789abcdef0123456789abcdef", span: "0123456789abcdef"})
			},
			wantTrace: "0123456789abcdef0123456789abcdef",
			wantSpan:  "0123456789abcdef",
		},
		"no span in context injects nothing": {
			ctx: context.Background,
		},
		"NoOpTracer StartSpan context injects nothing": {
			ctx: func() context.Context {
				ctx, _ := observability.NoOpTracer.StartSpan(context.Background(), "noop")
				return ctx
			},
		},
		"span with empty trace ID injects nothing (no empty-string noise)": {
			ctx: func() context.Context {
				return observability.ContextWithSpan(context.Background(), stubSpan{})
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			l := slog.New(observability.NewCorrelatingHandler(slog.NewJSONHandler(&buf, nil)))
			l.InfoContext(tc.ctx(), "msg")

			rec := decodeLine(t, &buf)
			gotTrace, traceOK := rec["trace_id"]
			gotSpan, spanOK := rec["span_id"]
			if tc.wantTrace == "" {
				if traceOK || spanOK {
					t.Fatalf("correlation keys must be absent; got trace_id=%v(%v) span_id=%v(%v)",
						gotTrace, traceOK, gotSpan, spanOK)
				}
				return
			}
			if gotTrace != tc.wantTrace || gotSpan != tc.wantSpan {
				t.Fatalf("trace_id=%v span_id=%v, want %q/%q", gotTrace, gotSpan, tc.wantTrace, tc.wantSpan)
			}
		})
	}
}

// WithAttrs/WithGroup must preserve both the delegate's behavior and the
// wrapper's injection (the derived handler stays correlating).
func TestCorrelatingHandlerSurvivesWithAttrsAndWithGroup(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(observability.NewCorrelatingHandler(slog.NewJSONHandler(&buf, nil)))
	l := base.With("component", "outbox").WithGroup("g")
	ctx := observability.ContextWithSpan(context.Background(), stubSpan{trace: "tt", span: "ss"})
	l.InfoContext(ctx, "msg", "k", "v")

	rec := decodeLine(t, &buf)
	if rec["component"] != "outbox" {
		t.Errorf("WithAttrs attr lost: %v", rec["component"])
	}
	g, _ := rec["g"].(map[string]any)
	if g == nil || g["k"] != "v" {
		t.Errorf("WithGroup grouping lost: %v", rec["g"])
	}
	// The injected attrs land in the open group — same slog semantics any
	// record-appended attr has after WithGroup; presence is the contract.
	if g == nil || g["trace_id"] != "tt" || g["span_id"] != "ss" {
		t.Errorf("correlation attrs lost after WithGroup: %v", rec)
	}
}

// End-to-end through the runtime path: Trace(tr) opens the request span and
// AccessLog's InfoContext(r.Context()) — through a correlating handler exactly
// as logging.New builds it — must emit an access-log line whose trace_id/span_id
// match the exported request span (AC-W01-E02-S001-01, Trace(tr) flavor).
func TestAccessLogInsideTraceMiddlewareCarriesExportedSpanIDs(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tr := oteladapter.New(exp, 1.0)

	var buf bytes.Buffer
	logger := slog.New(observability.NewCorrelatingHandler(slog.NewJSONHandler(&buf, nil)))

	h := observability.Trace(tr)(observability.AccessLog(logger)(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		})))
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)

	if err := tr.ForceFlush(context.Background()); err != nil {
		t.Fatalf("force flush: %v", err)
	}
	spans := exp.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("exported %d spans, want 1", len(spans))
	}
	sc := spans[0].SpanContext
	rec := decodeLine(t, &buf)
	if rec["trace_id"] != sc.TraceID().String() {
		t.Errorf("access log trace_id = %v, want %q", rec["trace_id"], sc.TraceID().String())
	}
	if rec["span_id"] != sc.SpanID().String() {
		t.Errorf("access log span_id = %v, want %q", rec["span_id"], sc.SpanID().String())
	}
}

// AC-W01-E02-S001-03 — allocation neutrality of the no-op path: emitting a log
// record through the correlating wrapper with no active span must allocate no
// more than the plain handler. Compare allocs/op of the two benchmarks.
func BenchmarkPlainHandlerNoSpan(b *testing.B) {
	l := slog.New(slog.NewJSONHandler(io.Discard, nil))
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		l.InfoContext(ctx, "msg", "k", "v")
	}
}

func BenchmarkCorrelatingHandlerNoSpan(b *testing.B) {
	l := slog.New(observability.NewCorrelatingHandler(slog.NewJSONHandler(io.Discard, nil)))
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		l.InfoContext(ctx, "msg", "k", "v")
	}
}
