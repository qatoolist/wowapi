package logging_test

// correlation_test.go — W01-E02-S001 (FBL-06 T2): trace/log correlation through
// the process logger construction path. A record emitted with a context
// carrying a real recording span must carry trace_id/span_id attrs equal to the
// span's own IDs (AC-W01-E02-S001-01); a record emitted with no active span —
// context.Background() or the NoOpTracer path — must not contain the keys AT
// ALL, not carry them with empty values (AC-W01-E02-S001-02, RISK-W01-E02-002).

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	oteladapter "github.com/qatoolist/wowapi/adapters/tracing/otel"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/logging"
	"github.com/qatoolist/wowapi/kernel/observability"
)

// decodeRecord parses the single JSON log line in buf into a key→value map so
// tests can assert on key PRESENCE, not just value equality.
func decodeRecord(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	line := strings.TrimSpace(buf.String())
	if line == "" {
		t.Fatal("no log output captured")
	}
	var rec map[string]any
	if err := json.Unmarshal([]byte(line), &rec); err != nil {
		t.Fatalf("log line is not valid JSON: %v\n%s", err, line)
	}
	return rec
}

func newJSONLogger(t *testing.T, buf *bytes.Buffer) *slog.Logger {
	t.Helper()
	l, err := logging.New(buf, config.Log{Level: "info", Format: "json"})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}
	return l
}

// AC-W01-E02-S001-01 — positive case: a record emitted inside a context
// carrying a real (non-no-op) recording span carries trace_id/span_id attrs
// whose values equal the span's own TraceID()/SpanID().
func TestLogRecordInsideActiveSpanCarriesTraceAndSpanIDs(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tr := oteladapter.New(exp, 1.0) // sample everything: a real recording span
	ctx, span := tr.StartSpan(context.Background(), "op")
	defer span.End()
	if span.TraceID() == "" || span.SpanID() == "" {
		t.Fatalf("fixture span has empty IDs (trace=%q span=%q)", span.TraceID(), span.SpanID())
	}

	var buf bytes.Buffer
	newJSONLogger(t, &buf).InfoContext(ctx, "inside span")

	rec := decodeRecord(t, &buf)
	if got, ok := rec["trace_id"]; !ok || got != span.TraceID() {
		t.Errorf("trace_id = %v (present=%v), want %q", got, ok, span.TraceID())
	}
	if got, ok := rec["span_id"]; !ok || got != span.SpanID() {
		t.Errorf("span_id = %v (present=%v), want %q", got, ok, span.SpanID())
	}
}

// AC-W01-E02-S001-02 — negative case, both flavors: with context.Background()
// (no span at all) and with the NoOpTracer's span context, the record's
// attribute set must not contain the trace_id/span_id KEYS. This asserts key
// absence from the decoded record, not empty-string values.
func TestLogRecordWithoutSpanOmitsCorrelationKeys(t *testing.T) {
	cases := map[string]func(t *testing.T) context.Context{
		"background context, no span": func(*testing.T) context.Context {
			return context.Background()
		},
		"NoOpTracer span": func(t *testing.T) context.Context {
			ctx, span := observability.NoOpTracer.StartSpan(context.Background(), "noop")
			t.Cleanup(span.End)
			return ctx
		},
	}
	for name, mkCtx := range cases {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			newJSONLogger(t, &buf).InfoContext(mkCtx(t), "outside span")

			rec := decodeRecord(t, &buf)
			if v, ok := rec["trace_id"]; ok {
				t.Errorf("trace_id key present (value %v); must be absent, not empty", v)
			}
			if v, ok := rec["span_id"]; ok {
				t.Errorf("span_id key present (value %v); must be absent, not empty", v)
			}
		})
	}
}

// Correlation contract: the
// correlation wrapper must not interfere with the existing redactAttr
// defense-in-depth — sensitive attrs stay redacted while correlation attrs are
// injected on the same record.
func TestCorrelationCoexistsWithSecretRedaction(t *testing.T) {
	exp := tracetest.NewInMemoryExporter()
	tr := oteladapter.New(exp, 1.0)
	ctx, span := tr.StartSpan(context.Background(), "op")
	defer span.End()

	var buf bytes.Buffer
	newJSONLogger(t, &buf).InfoContext(ctx, "boot", "db_password", "hunter2")

	rec := decodeRecord(t, &buf)
	if rec["db_password"] != "[redacted]" {
		t.Errorf("db_password = %v, want [redacted] (redaction must survive the wrapper)", rec["db_password"])
	}
	if strings.Contains(buf.String(), "hunter2") {
		t.Error("raw secret leaked into log output")
	}
	if rec["trace_id"] != span.TraceID() {
		t.Errorf("trace_id = %v, want %q alongside redaction", rec["trace_id"], span.TraceID())
	}
}
