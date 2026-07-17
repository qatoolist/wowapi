package jobs_test

import (
	"context"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/kernel/jobs"
	"github.com/qatoolist/wowapi/v2/testkit"
)

// depthMetrics is a fake observability.Metrics that records dlq_depth gauge sets
// keyed by their queue label, so the test asserts the exact series emitted.
type depthMetrics struct{ gauges map[string]float64 }

func (m *depthMetrics) ObserveRequest(_, _ string, _ int, _ time.Duration, _ int) {}
func (m *depthMetrics) IncCounter(_ string, _ float64, _ map[string]string)       {}
func (m *depthMetrics) ObserveHistogram(_ string, _ float64, _ map[string]string) {}
func (m *depthMetrics) SetGauge(name string, v float64, labels map[string]string) {
	key := name
	if q := labels["queue"]; q != "" {
		key = name + "/" + q
	}
	m.gauges[key] = v
}

// TestIntegrationDLQDepthGauge proves PublishDLQDepth counts the dead-lettered
// (discarded) jobs and emits them on the dlq_depth{queue="jobs"} gauge. NewDB
// gives an isolated, freshly-migrated database, so the seeded rows are the only
// discarded jobs present.
func TestIntegrationDLQDepthGauge(t *testing.T) {
	h := testkit.NewDB(t)
	ctx := context.Background()

	if n, err := jobs.CountDead(ctx, h.Platform); err != nil {
		t.Fatalf("CountDead baseline: %v", err)
	} else if n != 0 {
		t.Fatalf("baseline discarded jobs = %d, want 0 (fresh db)", n)
	}

	seedDiscardedJob(t, h, "depth-1")
	seedDiscardedJob(t, h, "depth-2")
	seedDiscardedJob(t, h, "depth-3")

	if n, err := jobs.CountDead(ctx, h.Platform); err != nil {
		t.Fatalf("CountDead: %v", err)
	} else if n != 3 {
		t.Fatalf("CountDead = %d, want 3", n)
	}

	fm := &depthMetrics{gauges: map[string]float64{}}
	if err := jobs.PublishDLQDepth(ctx, h.Platform, fm); err != nil {
		t.Fatalf("PublishDLQDepth: %v", err)
	}
	if got := fm.gauges["dlq_depth/jobs"]; got != 3 {
		t.Fatalf("dlq_depth{queue=jobs} = %v, want 3", got)
	}

	// nil sink must not panic and must not error (guards the worker's NoOp/nil path).
	if err := jobs.PublishDLQDepth(ctx, h.Platform, nil); err != nil {
		t.Fatalf("PublishDLQDepth(nil metrics): %v", err)
	}
}
