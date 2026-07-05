package jobs_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/jobs"
	"github.com/qatoolist/wowapi/kernel/observability"
	"github.com/qatoolist/wowapi/testkit"
)

// fakeTracer records the carriers passed to Extract and the span names started,
// and injects a fixed traceparent, so a test can prove the job queue carries
// trace context across the async boundary (O1/CA-9). Mirrors the outbox test's
// fake tracer.
type fakeTracer struct {
	inject    string
	mu        sync.Mutex
	extracted []string
	spans     []string
}

func (f *fakeTracer) StartSpan(ctx context.Context, name string) (context.Context, observability.Span) {
	f.mu.Lock()
	f.spans = append(f.spans, name)
	f.mu.Unlock()
	return ctx, fakeSpan{}
}
func (f *fakeTracer) Inject(context.Context) string { return f.inject }
func (f *fakeTracer) Extract(ctx context.Context, carrier string) context.Context {
	f.mu.Lock()
	f.extracted = append(f.extracted, carrier)
	f.mu.Unlock()
	return ctx
}

type fakeSpan struct{}

func (fakeSpan) End()                {}
func (fakeSpan) SetAttr(_, _ string) {}
func (fakeSpan) RecordError(error)   {}

// TestIntegrationJobsTracePropagation is the O1/CA-9 regression for the job
// runner: an enqueued job captures the enqueuer's trace context, and the runner
// extracts it and executes the worker under a child span — so an async job
// continues the originating request's trace. Mirrors the outbox test.
func TestIntegrationJobsTracePropagation(t *testing.T) {
	h := testkit.NewDB(t)
	const carrier = "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01"
	tr := &fakeTracer{inject: carrier}
	tenant := testkit.CreateTenant(t, h)

	var ran int64
	reg := jobs.NewRegistry()
	reg.RegisterKind(jobKind, func(context.Context, database.TenantDB, []byte) error {
		atomic.AddInt64(&ran, 1)
		return nil
	}, jobs.DefaultRetry())
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}

	// Enqueue WITH the tracer so the job captures the injected traceparent.
	if err := h.TxM.WithTenant(testkit.TenantCtx(tenant.ID), func(ctx context.Context, db database.TenantDB) error {
		return jobs.Enqueue(ctx, db, testJob{N: "traced"}, jobs.WithTracer(tr))
	}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	// The job row stored the injected trace context.
	var stored string
	if err := h.Platform.QueryRow(context.Background(),
		`SELECT coalesce(trace_context,'') FROM jobs_queue WHERE kind = $1`, jobKind).Scan(&stored); err != nil {
		t.Fatalf("read trace_context: %v", err)
	}
	if stored != carrier {
		t.Fatalf("stored trace_context = %q, want the injected carrier %q", stored, carrier)
	}

	// The runner extracts that carrier and starts a child span when executing.
	r := jobs.NewRunner(h.Platform, h.TxM, reg, jobs.WithRunnerTracer(tr))
	n, err := r.ClaimOnce(context.Background())
	if err != nil {
		t.Fatalf("ClaimOnce: %v", err)
	}
	if n != 1 {
		t.Fatalf("ClaimOnce claimed %d, want 1", n)
	}
	if atomic.LoadInt64(&ran) != 1 {
		t.Fatalf("worker ran %d times, want 1", ran)
	}

	tr.mu.Lock()
	defer tr.mu.Unlock()
	found := false
	for _, c := range tr.extracted {
		if c == carrier {
			found = true
		}
	}
	if !found {
		t.Fatalf("runner must Extract the stored trace context to continue the trace; extracted=%v", tr.extracted)
	}
	// The runner opened a child span for the job execution.
	spanFound := false
	for _, name := range tr.spans {
		if name == "jobs.run "+jobKind {
			spanFound = true
		}
	}
	if !spanFound {
		t.Fatalf("runner must start a job-runner span %q; spans=%v", "jobs.run "+jobKind, tr.spans)
	}
}

// TestIntegrationJobsNoTracerNoContext proves backward compatibility: without a
// tracer the enqueue stores NULL trace_context (no behavior change) and the
// runner still executes the job.
func TestIntegrationJobsNoTracerNoContext(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)

	var ran int64
	reg := jobs.NewRegistry()
	reg.RegisterKind(jobKind, func(context.Context, database.TenantDB, []byte) error {
		atomic.AddInt64(&ran, 1)
		return nil
	}, jobs.DefaultRetry())
	if err := reg.Err(); err != nil {
		t.Fatalf("registry: %v", err)
	}

	if err := h.TxM.WithTenant(testkit.TenantCtx(tenant.ID), func(ctx context.Context, db database.TenantDB) error {
		return jobs.Enqueue(ctx, db, testJob{N: "plain"})
	}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	var stored *string
	if err := h.Platform.QueryRow(context.Background(),
		`SELECT trace_context FROM jobs_queue WHERE kind = $1`, jobKind).Scan(&stored); err != nil {
		t.Fatalf("read trace_context: %v", err)
	}
	if stored != nil {
		t.Fatalf("trace_context = %q, want NULL with no tracer", *stored)
	}

	// Runner with the default (NoOp) tracer still runs the job.
	r := jobs.NewRunner(h.Platform, h.TxM, reg)
	if n, err := r.ClaimOnce(context.Background()); err != nil || n != 1 {
		t.Fatalf("ClaimOnce = (%d, %v), want (1, nil)", n, err)
	}
	if atomic.LoadInt64(&ran) != 1 {
		t.Fatalf("worker ran %d times, want 1", ran)
	}
}
