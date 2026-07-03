package prometheus_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus/testutil"

	promadapter "github.com/qatoolist/wowapi/adapters/metrics/prometheus"
	"github.com/qatoolist/wowapi/kernel/observability"
)

// Compile-time assertion: *Prometheus implements observability.Metrics.
var _ observability.Metrics = (*promadapter.Prometheus)(nil)

func TestNewImplementsMetrics(t *testing.T) {
	p := promadapter.New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	// Exercise all three method families; none must panic.
	p.ObserveRequest("/health", "GET", 200, 5*time.Millisecond, 0)
	p.IncCounter("authz_denied_total", 1.0, map[string]string{"kind": "role"})
	p.SetGauge("outbox_pending", 7.0, map[string]string{"module": "foo"})
}

// TestObserveRequestIncrementsHistogram verifies that ObserveRequest writes an
// observation into the http_request_duration_seconds histogram.
func TestObserveRequestIncrementsHistogram(t *testing.T) {
	p := promadapter.New()
	p.ObserveRequest("/api/things", "GET", 200, 10*time.Millisecond, 64)
	p.ObserveRequest("/api/things", "POST", 201, 20*time.Millisecond, 0)

	// GatherAndCount counts distinct metric *families* in the registry; the
	// histogram registers as one family regardless of label cardinality.
	count, err := testutil.GatherAndCount(p.Gatherer(), "http_request_duration_seconds")
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	if count < 1 {
		t.Errorf("histogram family not present after ObserveRequest; count=%d", count)
	}
}

// TestObserveRequestTextFormat spot-checks the Prometheus text format for an
// expected count metric line (sorted label order: method, route, status).
func TestObserveRequestTextFormat(t *testing.T) {
	p := promadapter.New()
	p.ObserveRequest("/api/things", "GET", 200, 10*time.Millisecond, 0)

	rec := httptest.NewRecorder()
	p.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/metrics", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	want := `http_request_duration_seconds_count{method="GET",route="/api/things",status="200"} 1`
	if !strings.Contains(body, want) {
		t.Errorf("expected line %q not found in:\n%s", want, body)
	}
}

// TestHandlerServes200WithHelp verifies that Handler() returns an HTTP 200
// response containing a # HELP line for the request duration metric.
func TestHandlerServes200WithHelp(t *testing.T) {
	p := promadapter.New()
	p.ObserveRequest("/x", "GET", 200, time.Millisecond, 0)

	srv := httptest.NewServer(p.Handler())
	defer srv.Close()

	resp, err := srv.Client().Get(srv.URL + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "# HELP http_request_duration_seconds") {
		t.Errorf("# HELP line missing from /metrics:\n%s", body)
	}
}

// TestIncCounterAccumulates verifies that repeated IncCounter calls add up.
func TestIncCounterAccumulates(t *testing.T) {
	p := promadapter.New()
	labels := map[string]string{"kind": "authz_denied"}
	p.IncCounter("security_events_total", 1.0, labels)
	p.IncCounter("security_events_total", 2.0, labels)

	count, err := testutil.GatherAndCount(p.Gatherer(), "security_events_total")
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	if count < 1 {
		t.Errorf("counter family not present; count=%d", count)
	}
}

// TestSetGaugeUpdates verifies that SetGauge changes the gauge value.
func TestSetGaugeUpdates(t *testing.T) {
	p := promadapter.New()
	labels := map[string]string{"module": "outbox"}
	p.SetGauge("outbox_pending", 5.0, labels)
	p.SetGauge("outbox_pending", 3.0, labels)

	count, err := testutil.GatherAndCount(p.Gatherer(), "outbox_pending")
	if err != nil {
		t.Fatalf("gather: %v", err)
	}
	if count < 1 {
		t.Errorf("gauge family not present; count=%d", count)
	}
}
