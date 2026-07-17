package database_test

// query_tracer_test.go — W01-E02-S002 (FBL-06 T3, D-08): the thin in-kernel
// pgx.QueryTracer over the observability.Tracer port. Integration tests run
// against a real Postgres (guardTestDSN skips when absent) with the otel
// adapter's in-memory exporter as the trace fixture — the trace TREE is
// asserted (parent/child span IDs), not a mock interaction.

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	oteladapter "github.com/qatoolist/wowapi/v2/adapters/tracing/otel"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/database"
)

// attrValue returns the string value of key on the span stub, or "" when absent.
func attrValue(s tracetest.SpanStub, key string) string {
	for _, kv := range s.Attributes {
		if string(kv.Key) == key {
			return kv.Value.AsString()
		}
	}
	return ""
}

// findChildSpan returns the first exported span whose parent span ID is
// parentSpanID, excluding the parent itself.
func findChildSpan(spans tracetest.SpanStubs, parentSpanID string) (tracetest.SpanStub, bool) {
	for _, s := range spans {
		if s.Parent.HasSpanID() && s.Parent.SpanID().String() == parentSpanID {
			return s, true
		}
	}
	return tracetest.SpanStub{}, false
}

// tracedFixture builds a real pool against the test DSN plus an in-memory
// exported tracer, returning both. The pool is closed on cleanup.
func tracedFixture(t *testing.T) (*tracetest.InMemoryExporter, *oteladapter.Tracer, *pgxpool.Pool) {
	t.Helper()
	dsn := guardTestDSN(t)
	exp := tracetest.NewInMemoryExporter()
	tr := oteladapter.New(exp, 1.0) // sample everything
	pool, err := database.NewPool(context.Background(), dsn, config.Defaults().DB,
		database.WithQueryTracer(tr))
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	t.Cleanup(pool.Close)
	return exp, tr, pool
}

// AC-W01-E02-S002-01 — a query executed inside a context carrying an active
// parent span produces a child span in the exported trace tree whose parent
// span ID equals the parent span's own ID.
func TestIntegrationQueryTracerChildSpanInTraceTree(t *testing.T) {
	exp, tr, pool := tracedFixture(t)

	ctx, parent := tr.StartSpan(context.Background(), "parent-op")
	if _, err := pool.Exec(ctx, "SELECT 1"); err != nil {
		t.Fatalf("query: %v", err)
	}
	parent.End()
	if err := tr.ForceFlush(context.Background()); err != nil {
		t.Fatalf("force flush: %v", err)
	}

	spans := exp.GetSpans()
	child, ok := findChildSpan(spans, parent.SpanID())
	if !ok {
		t.Fatalf("no query span parented under %q in exported trace tree (%d spans exported)",
			parent.SpanID(), len(spans))
	}
	if child.Name != "db.SELECT" {
		t.Errorf("query span name = %q, want db.SELECT", child.Name)
	}
	if got := child.SpanContext.TraceID().String(); got != parent.TraceID() {
		t.Errorf("query span trace ID = %q, want parent's %q", got, parent.TraceID())
	}
}

// AC-W01-E02-S002-02 (attrs) — the query span carries a statement-summary attr
// and a rows-affected attr.
func TestIntegrationQueryTracerStatementAndRowsAffectedAttrs(t *testing.T) {
	exp, tr, pool := tracedFixture(t)

	ctx, parent := tr.StartSpan(context.Background(), "parent-op")
	if _, err := pool.Exec(ctx, "SELECT 1"); err != nil {
		t.Fatalf("query: %v", err)
	}
	parent.End()
	if err := tr.ForceFlush(context.Background()); err != nil {
		t.Fatalf("force flush: %v", err)
	}

	child, ok := findChildSpan(exp.GetSpans(), parent.SpanID())
	if !ok {
		t.Fatal("no query span exported")
	}
	if got := attrValue(child, "db.statement"); got != "SELECT 1" {
		t.Errorf("db.statement = %q, want SELECT 1", got)
	}
	if got := attrValue(child, "db.rows_affected"); got != "1" {
		t.Errorf("db.rows_affected = %q, want 1", got)
	}
}

// AC-W01-E02-S002-02 (error marking) — a failed query marks its span errored
// via RecordError (status code Error plus a recorded exception event).
func TestIntegrationQueryTracerMarksFailedQueryErrored(t *testing.T) {
	exp, tr, pool := tracedFixture(t)

	ctx, parent := tr.StartSpan(context.Background(), "parent-op")
	if _, err := pool.Exec(ctx, "SELECT wow FROM table_that_does_not_exist"); err == nil {
		t.Fatal("query against a missing table must fail")
	}
	parent.End()
	if err := tr.ForceFlush(context.Background()); err != nil {
		t.Fatalf("force flush: %v", err)
	}

	child, ok := findChildSpan(exp.GetSpans(), parent.SpanID())
	if !ok {
		t.Fatal("failed query produced no span")
	}
	if child.Status.Code != codes.Error {
		t.Errorf("span status = %v, want Error", child.Status.Code)
	}
	var hasException bool
	for _, ev := range child.Events {
		if ev.Name == "exception" {
			hasException = true
		}
	}
	if !hasException {
		t.Error("span has no recorded exception event")
	}
}

// AC-W01-E02-S002-02 (literal leakage) — a sensitive-looking value bound as a
// query PARAMETER must never appear in any span attribute: the statement
// summary is the parameterized SQL text ($1 placeholders), not rendered SQL.
func TestIntegrationQueryTracerDoesNotLeakBoundParameterLiterals(t *testing.T) {
	const secret = "hunter2-super-secret-literal"
	exp, tr, pool := tracedFixture(t)

	ctx, parent := tr.StartSpan(context.Background(), "parent-op")
	if _, err := pool.Exec(ctx, "SELECT $1::text", secret); err != nil {
		t.Fatalf("query: %v", err)
	}
	parent.End()
	if err := tr.ForceFlush(context.Background()); err != nil {
		t.Fatalf("force flush: %v", err)
	}

	for _, s := range exp.GetSpans() {
		for _, kv := range s.Attributes {
			if kv.Value.Type() == attribute.STRING && strings.Contains(kv.Value.AsString(), secret) {
				t.Errorf("span %q attr %q leaked a bound parameter literal", s.Name, kv.Key)
			}
		}
	}
	child, ok := findChildSpan(exp.GetSpans(), parent.SpanID())
	if !ok {
		t.Fatal("no query span exported")
	}
	if got := attrValue(child, "db.statement"); !strings.Contains(got, "$1") {
		t.Errorf("db.statement = %q, want the parameterized text with $1 placeholder", got)
	}
}

// A query with NO parent span in context still works and produces a root span
// (parent-driven sampling is the otel adapter's concern, not the tracer's) —
// and, more importantly for the residual-risk note, never panics.
func TestIntegrationQueryTracerRootSpanWithoutParent(t *testing.T) {
	exp, tr, pool := tracedFixture(t)

	if _, err := pool.Exec(context.Background(), "SELECT 1"); err != nil {
		t.Fatalf("query: %v", err)
	}
	if err := tr.ForceFlush(context.Background()); err != nil {
		t.Fatalf("force flush: %v", err)
	}
	var found bool
	for _, s := range exp.GetSpans() {
		if s.Name == "db.SELECT" && !s.Parent.HasSpanID() {
			found = true
		}
	}
	if !found {
		t.Error("no root db.SELECT span exported for a parentless query")
	}
}

// Compile-time proof the sampler inheritance claim rests on: the otel adapter
// uses a ParentBased sampler, so an UNSAMPLED parent yields an unsampled (not
// exported) query span — no independent sampling decision at the DB layer.
func TestIntegrationQueryTracerInheritsParentSamplingDecision(t *testing.T) {
	dsn := guardTestDSN(t)
	exp := tracetest.NewInMemoryExporter()
	tr := oteladapter.New(exp, 0) // never sample new roots
	pool, err := database.NewPool(context.Background(), dsn, config.Defaults().DB,
		database.WithQueryTracer(tr))
	if err != nil {
		t.Fatalf("NewPool: %v", err)
	}
	defer pool.Close()

	ctx, parent := tr.StartSpan(context.Background(), "unsampled-parent")
	if _, err := pool.Exec(ctx, "SELECT 1"); err != nil {
		t.Fatalf("query: %v", err)
	}
	parent.End()
	if err := tr.ForceFlush(context.Background()); err != nil {
		t.Fatalf("force flush: %v", err)
	}
	if n := len(exp.GetSpans()); n != 0 {
		t.Errorf("unsampled parent still exported %d spans; query spans must inherit the parent decision", n)
	}
}
