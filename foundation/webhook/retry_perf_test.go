package webhook_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/foundation/webhook"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/observability"
	"github.com/qatoolist/wowapi/testkit"
	"github.com/qatoolist/wowapi/testkit/fakes"
)

type retryQueryRecorder struct {
	mu         sync.Mutex
	statements []string
}

func (q *retryQueryRecorder) StartSpan(ctx context.Context, _ string) (context.Context, observability.Span) {
	return ctx, retryQuerySpan{record: q}
}
func (*retryQueryRecorder) Inject(context.Context) string                         { return "" }
func (*retryQueryRecorder) Extract(ctx context.Context, _ string) context.Context { return ctx }
func (q *retryQueryRecorder) reset() {
	q.mu.Lock()
	q.statements = nil
	q.mu.Unlock()
}

func (q *retryQueryRecorder) count(fragment string) int {
	q.mu.Lock()
	defer q.mu.Unlock()
	n := 0
	for _, statement := range q.statements {
		if strings.Contains(statement, fragment) {
			n++
		}
	}
	return n
}

type retryQuerySpan struct{ record *retryQueryRecorder }

func (retryQuerySpan) End()              {}
func (retryQuerySpan) RecordError(error) {}
func (retryQuerySpan) TraceID() string   { return "" }
func (retryQuerySpan) SpanID() string    { return "" }
func (s retryQuerySpan) SetAttr(key, value string) {
	if key != "db.statement" {
		return
	}
	s.record.mu.Lock()
	s.record.statements = append(s.record.statements, value)
	s.record.mu.Unlock()
}

type retryMetrics struct {
	mu     sync.Mutex
	gauges map[string]float64
}

func (*retryMetrics) ObserveRequest(string, string, int, time.Duration, int) {}
func (*retryMetrics) IncCounter(string, float64, map[string]string)          {}
func (m *retryMetrics) ObserveHistogram(name string, value float64, labels map[string]string) {
	m.SetGauge(name, value, labels)
}

func (m *retryMetrics) SetGauge(name string, value float64, labels map[string]string) {
	m.mu.Lock()
	m.gauges[name+"/"+labels["worker"]] = value
	m.mu.Unlock()
}

func (m *retryMetrics) has(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.gauges[name]
	return ok
}

func seedRetryBatch(t *testing.T, h *testkit.DBHandle, tenant uuid.UUID, endpoints, deliveries int) {
	t.Helper()
	endpointIDs := make([]uuid.UUID, endpoints)
	for i := range endpointIDs {
		endpointIDs[i] = seedOutboundEndpoint(t, h, tenant, "https://example.test/retry")
	}
	for i := range deliveries {
		if _, err := h.Admin.Exec(context.Background(), `INSERT INTO webhook_events
			(id,tenant_id,endpoint_id,direction,external_event_id,event_type,payload,delivery_status,received_at,next_attempt_at)
			VALUES ($1,$2,$3,'outbound',$4,'order.created','{}','failed',now()-interval '1 hour',now()-interval '1 minute')`,
			uuid.New(), tenant, endpointIDs[i%len(endpointIDs)], uuid.NewString()); err != nil {
			t.Fatalf("seed delivery %d: %v", i, err)
		}
	}
}

func TestIntegrationRetryOutboundBatchLoadsEndpointsOnce(t *testing.T) {
	recorder := &retryQueryRecorder{}
	h := testkit.NewDBWithOptions(t, testkit.DBOptions{PlatformPool: []database.Option{database.WithQueryTracer(recorder)}})
	tenant := testkit.CreateTenant(t, h)
	seedRetryBatch(t, h, tenant.ID, 3, 10)
	svc := webhook.New(&fakes.WebhookSender{StatusCode: 200}, &fakes.WebhookSecretResolver{Secret: testSecret}, model.UUIDv7())
	recorder.reset()

	if err := svc.RetryOutbound(context.Background(), h.PlatformTxM, tenant.ID, time.Now()); err != nil {
		t.Fatalf("RetryOutbound: %v", err)
	}
	if got := recorder.count("FROM webhook_endpoints WHERE id = ANY"); got != 1 {
		t.Fatalf("batch endpoint queries = %d, want exactly 1 for 10 deliveries / 3 endpoints", got)
	}
	if got := recorder.count("FROM webhook_endpoints WHERE id = $1"); got != 0 {
		t.Fatalf("per-delivery endpoint queries = %d, want 0", got)
	}
}

func TestIntegrationRetryOutboundMetrics(t *testing.T) {
	h := testkit.NewDB(t)
	tenant := testkit.CreateTenant(t, h)
	seedRetryBatch(t, h, tenant.ID, 1, 1)
	metrics := &retryMetrics{gauges: map[string]float64{}}
	svc := webhook.New(&fakes.WebhookSender{StatusCode: 200}, &fakes.WebhookSecretResolver{Secret: testSecret}, model.UUIDv7(), webhook.WithMetrics(metrics))

	if err := svc.RetryOutbound(context.Background(), h.PlatformTxM, tenant.ID, time.Now()); err != nil {
		t.Fatalf("RetryOutbound: %v", err)
	}
	for _, name := range []string{"worker_queue_lag_seconds/webhook_retry", "worker_batch_duration_seconds/webhook_retry"} {
		if !metrics.has(name) {
			t.Errorf("missing metric %s", name)
		}
	}
}
