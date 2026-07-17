package observability_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/observability"
)

type fakeSpan struct {
	name  string
	attrs map[string]string
	ended bool
}

func (s *fakeSpan) End()                  { s.ended = true }
func (s *fakeSpan) SetAttr(k, v string)   { s.attrs[k] = v }
func (s *fakeSpan) RecordError(err error) { s.attrs["error"] = err.Error() }
func (s *fakeSpan) TraceID() string       { return "" }
func (s *fakeSpan) SpanID() string        { return "" }

type fakeTracer struct {
	spans     []*fakeSpan
	extracted string
}

func (f *fakeTracer) StartSpan(ctx context.Context, name string) (context.Context, observability.Span) {
	s := &fakeSpan{name: name, attrs: map[string]string{}}
	f.spans = append(f.spans, s)
	return ctx, s
}

func (f *fakeTracer) Inject(context.Context) string { return "" }
func (f *fakeTracer) Extract(ctx context.Context, carrier string) context.Context {
	f.extracted = carrier
	return ctx
}

func TestTraceMiddlewareSpansRequest(t *testing.T) {
	tr := &fakeTracer{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /things/{id}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	h := observability.Trace(tr)(mux)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/things/42", nil)
	req.Header.Set("traceparent", "00-trace-span-01")
	h.ServeHTTP(rec, req)

	if tr.extracted != "00-trace-span-01" {
		t.Errorf("inbound traceparent not extracted: %q", tr.extracted)
	}
	if len(tr.spans) != 1 {
		t.Fatalf("started %d spans, want 1", len(tr.spans))
	}
	s := tr.spans[0]
	if !s.ended {
		t.Error("span must be ended")
	}
	if s.attrs["http.route"] != "/things/{id}" {
		t.Errorf("http.route = %q, want /things/{id}", s.attrs["http.route"])
	}
	if s.attrs["http.status"] != "418" {
		t.Errorf("http.status = %q, want 418", s.attrs["http.status"])
	}
	if s.attrs["http.method"] != "GET" {
		t.Errorf("http.method = %q, want GET", s.attrs["http.method"])
	}
}

func TestNoOpTracerIsZeroCost(t *testing.T) {
	// The NoOp tracer returns the same context and a span that swallows everything.
	ctx := context.Background()
	got, span := observability.NoOpTracer.StartSpan(ctx, "x")
	if got != ctx {
		t.Error("NoOpTracer must return the context unchanged")
	}
	span.SetAttr("k", "v")
	span.RecordError(http.ErrBodyNotAllowed)
	span.End() // must not panic
}
