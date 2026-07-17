package workflow_test

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/audit"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/observability"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/workflow"
	"github.com/qatoolist/wowapi/testkit"
)

const sweepBatchLimit = 100

type queryRecorder struct {
	mu         sync.Mutex
	statements []string
}

func (q *queryRecorder) StartSpan(ctx context.Context, _ string) (context.Context, observability.Span) {
	return ctx, querySpan{record: q}
}
func (*queryRecorder) Inject(context.Context) string                         { return "" }
func (*queryRecorder) Extract(ctx context.Context, _ string) context.Context { return ctx }
func (q *queryRecorder) reset() {
	q.mu.Lock()
	q.statements = nil
	q.mu.Unlock()
}

func (q *queryRecorder) count(fragment string) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	count := 0
	for _, statement := range q.statements {
		if strings.Contains(statement, fragment) {
			count++
		}
	}
	return count
}

type querySpan struct{ record *queryRecorder }

func (querySpan) End()              {}
func (querySpan) RecordError(error) {}
func (querySpan) TraceID() string   { return "" }
func (querySpan) SpanID() string    { return "" }
func (s querySpan) SetAttr(k, v string) {
	if k != "db.statement" {
		return
	}
	s.record.mu.Lock()
	s.record.statements = append(s.record.statements, v)
	s.record.mu.Unlock()
}

type gaugeRecorder struct {
	mu     sync.Mutex
	values map[string]float64
}

func (*gaugeRecorder) ObserveRequest(string, string, int, time.Duration, int) {}
func (*gaugeRecorder) IncCounter(string, float64, map[string]string)          {}
func (m *gaugeRecorder) ObserveHistogram(name string, value float64, labels map[string]string) {
	m.SetGauge(name, value, labels)
}

func (m *gaugeRecorder) SetGauge(name string, value float64, labels map[string]string) {
	key := name + "/" + labels["worker"]
	m.mu.Lock()
	m.values[key] = value
	m.mu.Unlock()
}

func (m *gaugeRecorder) has(key string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.values[key]
	return ok
}

func newSweepRuntime(tb testing.TB, h *testkit.DBHandle, metrics observability.Metrics) *workflow.Runtime {
	tb.Helper()
	reg := workflow.NewRegistry()
	def, err := workflow.ParseDefinition([]byte(linearDef))
	if err != nil {
		tb.Fatalf("parse definition: %v", err)
	}
	if err := reg.RegisterDefinition(def); err != nil {
		tb.Fatalf("register definition: %v", err)
	}
	reg.RegisterAssigneeResolver("test.approver", func(context.Context, workflow.ResolveInput) ([]workflow.Assignee, error) {
		return []workflow.Assignee{{Kind: workflow.KindCapacity, Ref: uuid.NewString()}}, nil
	})
	reg.RegisterAutoAction("requests.provision", func(context.Context, workflow.AutoInput) (map[string]any, error) {
		return map[string]any{"provisioned": true}, nil
	})
	if err := reg.Err(); err != nil {
		tb.Fatalf("registry: %v", err)
	}
	if err := workflow.SyncDefinitions(context.Background(), h.Platform, reg); err != nil {
		tb.Fatalf("sync definitions: %v", err)
	}
	rt := workflow.NewRuntime(h.TxM, reg, fakeEvaluator{allow: map[string]bool{"workflow.instance.override": true}},
		outbox.NewWriter(model.UUIDv7()), model.UUIDv7(), audit.New(model.UUIDv7(), nil), workflow.WithRuntimeMetrics(metrics))
	return rt
}

func seedSweepFixture(tb testing.TB, h *testkit.DBHandle, cardinality int) (uuid.UUID, uuid.UUID) {
	tb.Helper()
	tenant := testkit.CreateTenantTB(tb, h)
	var definitionID uuid.UUID
	instanceID := uuid.New()
	resourceID := uuid.New()
	ctx := context.Background()
	if err := h.Admin.QueryRow(ctx, `SELECT id FROM workflow_definitions
		WHERE key = 'requests.approval' AND version = 1`).Scan(&definitionID); err != nil {
		tb.Fatalf("load synchronized definition: %v", err)
	}
	if _, err := h.Admin.Exec(ctx, `INSERT INTO workflow_instances
		(id,tenant_id,definition_id,resource_type,resource_id,current_step,status,context,started_by,created_by)
		VALUES ($1,$2,$3,'requests.request',$4,'manager_review','running','{}',$5,$5)`, instanceID, tenant.ID, definitionID, resourceID, uuid.Nil); err != nil {
		tb.Fatalf("insert instance: %v", err)
	}
	if _, err := h.Admin.Exec(ctx, `INSERT INTO workflow_tasks
		(id,tenant_id,instance_id,step_key,task_type,status,remind_after,created_by)
		SELECT gen_random_uuid(),$1,$2,'manager_review','approval','open',now()-interval '1 hour',$3
		FROM generate_series(1,$4)`, tenant.ID, instanceID, uuid.Nil, cardinality); err != nil {
		tb.Fatalf("insert %d tasks: %v", cardinality, err)
	}
	return tenant.ID, instanceID
}

func runSweep(tb testing.TB, h *testkit.DBHandle, rt *workflow.Runtime, tenant uuid.UUID) (int, int) {
	tb.Helper()
	var reminders, escalations int
	if err := h.TxM.WithTenant(testkit.TenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		var err error
		reminders, escalations, err = rt.SweepSLA(ctx, db, time.Now())
		return err
	}); err != nil {
		tb.Fatalf("SweepSLA: %v", err)
	}
	return reminders, escalations
}

func TestIntegrationSweepSLABoundedQueriesAtDueCardinalities(t *testing.T) {
	for _, cardinality := range []int{10, 1_000, 100_000} {
		t.Run(fmt.Sprintf("due_%d", cardinality), func(t *testing.T) {
			recorder := &queryRecorder{}
			h := testkit.NewDBWithOptions(t, testkit.DBOptions{RuntimePool: []database.Option{database.WithQueryTracer(recorder)}})
			rt := newSweepRuntime(t, h, observability.NoOp)
			tenant, _ := seedSweepFixture(t, h, cardinality)
			recorder.reset()

			reminders, escalations := runSweep(t, h, rt, tenant)
			want := cardinality
			if want > sweepBatchLimit {
				want = sweepBatchLimit
			}
			if reminders != want || escalations != 0 {
				t.Fatalf("SweepSLA = (%d,%d), want (%d,0)", reminders, escalations, want)
			}
			if got := recorder.count("UPDATE workflow_tasks"); got != 2 {
				t.Fatalf("guard-flip statements = %d, want exactly 2 fixed batch statements", got)
			}
			if got := recorder.count("FROM workflow_instances WHERE id = ANY"); got != 1 {
				t.Fatalf("batch instance loads = %d, want 1", got)
			}
			if got := recorder.count("FROM workflow_definitions WHERE id = ANY"); got != 1 {
				t.Fatalf("batch definition loads = %d, want 1", got)
			}
		})
	}
}

func TestIntegrationSweepSLASafeReinvocationAndNoDoubleReminder(t *testing.T) {
	h := testkit.NewDB(t)
	rt := newSweepRuntime(t, h, observability.NoOp)
	tenant, _ := seedSweepFixture(t, h, 2*sweepBatchLimit+1)
	for i, want := range []int{sweepBatchLimit, sweepBatchLimit, 1, 0} {
		got, _ := runSweep(t, h, rt, tenant)
		if got != want {
			t.Fatalf("invocation %d reminders = %d, want %d", i+1, got, want)
		}
	}
	var events int
	if err := h.Admin.QueryRow(context.Background(), `SELECT count(*) FROM events_outbox
		WHERE tenant_id=$1 AND event_type='workflow.requests.approval.reminded'`, tenant).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if events != 2*sweepBatchLimit+1 {
		t.Fatalf("reminder events = %d, want %d", events, 2*sweepBatchLimit+1)
	}
}

func TestIntegrationSweepSLAConcurrentWorkersDoNotDoubleRemind(t *testing.T) {
	h := testkit.NewDB(t)
	rt := newSweepRuntime(t, h, observability.NoOp)
	tenant, _ := seedSweepFixture(t, h, sweepBatchLimit)

	start := make(chan struct{})
	results := make(chan int, 2)
	for range 2 {
		go func() {
			<-start
			reminders, _ := runSweep(t, h, rt, tenant)
			results <- reminders
		}()
	}
	close(start)
	if got := <-results + <-results; got != sweepBatchLimit {
		t.Fatalf("concurrent reminder total = %d, want %d", got, sweepBatchLimit)
	}
}

func TestIntegrationSweepSLARemindAfterIndexPlan(t *testing.T) {
	h := testkit.NewDB(t)
	_ = newSweepRuntime(t, h, observability.NoOp)
	tenant, _ := seedSweepFixture(t, h, 5_000)
	if _, err := h.Admin.Exec(context.Background(), "ANALYZE workflow_tasks"); err != nil {
		t.Fatal(err)
	}
	var planLines []string
	if err := h.TxM.WithTenant(testkit.TenantCtx(tenant), func(ctx context.Context, db database.TenantDB) error {
		rows, err := db.Query(ctx, `EXPLAIN (COSTS OFF)
			SELECT id,instance_id,step_key FROM workflow_tasks
			WHERE status='open' AND remind_after IS NOT NULL AND remind_after <= $1
			  AND (last_reminded_at IS NULL OR last_reminded_at < remind_after)
			ORDER BY remind_after,id LIMIT $2`, time.Now(), sweepBatchLimit)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var line string
			if err := rows.Scan(&line); err != nil {
				return err
			}
			planLines = append(planLines, line)
		}
		return rows.Err()
	}); err != nil {
		t.Fatal(err)
	}
	plan := strings.Join(planLines, "\n")
	t.Logf("reminder EXPLAIN plan:\n%s", plan)
	if !strings.Contains(plan, "wft_remind_after") || !strings.Contains(plan, "Index Scan") {
		t.Fatalf("reminder plan does not use wft_remind_after index:\n%s", plan)
	}
}

func TestIntegrationSweepSLAMetrics(t *testing.T) {
	h := testkit.NewDB(t)
	metrics := &gaugeRecorder{values: map[string]float64{}}
	rt := newSweepRuntime(t, h, metrics)
	tenant, _ := seedSweepFixture(t, h, 1)
	runSweep(t, h, rt, tenant)
	for _, key := range []string{"worker_queue_lag_seconds/workflow_sla", "worker_batch_duration_seconds/workflow_sla"} {
		if !metrics.has(key) {
			t.Errorf("missing metric %s", key)
		}
	}
}

func BenchmarkSweepSLABatch(b *testing.B) {
	for _, cardinality := range []int{10, 1_000, 100_000} {
		b.Run(fmt.Sprintf("due_%d", cardinality), func(b *testing.B) {
			h := testkit.NewDB(b)
			rt := newSweepRuntime(b, h, observability.NoOp)
			tenant, _ := seedSweepFixture(b, h, cardinality)
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if _, err := h.Admin.Exec(context.Background(), `UPDATE workflow_tasks
					SET last_reminded_at=NULL WHERE tenant_id=$1`, tenant); err != nil {
					b.Fatal(err)
				}
				if _, err := h.Admin.Exec(context.Background(), `DELETE FROM events_outbox WHERE tenant_id=$1`, tenant); err != nil {
					b.Fatal(err)
				}
				b.StartTimer()
				runSweep(b, h, rt, tenant)
				b.StopTimer()
			}
		})
	}
}
