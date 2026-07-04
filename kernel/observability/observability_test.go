package observability_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/observability"
)

// ---------- recording fake (satisfies observability.Metrics) ----------

type requestObs struct {
	route     string
	method    string
	status    int
	dur       time.Duration
	respBytes int
}

type recorder struct {
	reqs []requestObs
}

func (r *recorder) ObserveRequest(route, method string, status int, dur time.Duration, respBytes int) {
	r.reqs = append(r.reqs, requestObs{route, method, status, dur, respBytes})
}
func (r *recorder) IncCounter(_ string, _ float64, _ map[string]string) {}
func (r *recorder) SetGauge(_ string, _ float64, _ map[string]string)   {}

// ---------- NoOp ----------

func TestNoOpNeverPanics(t *testing.T) {
	m := observability.NoOp
	m.ObserveRequest("/health", "GET", 200, time.Millisecond, 42)
	m.IncCounter("authz_denied_total", 1.0, map[string]string{"kind": "role"})
	m.SetGauge("outbox_pending", 7.0, nil)
}

// ---------- Requests middleware ----------

// TestRequestsRecordsRouteMethodStatus verifies the middleware calls
// ObserveRequest with the matched route pattern (not the raw URL), the HTTP
// method, and the handler's response status.
func TestRequestsRecordsRouteMethodStatus(t *testing.T) {
	rec := &recorder{}

	// Use a ServeMux so r.Pattern is populated ("GET /widgets/{id}").
	mux := http.NewServeMux()
	mux.Handle("GET /widgets/{id}", observability.Requests(rec)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	))

	r := httptest.NewRequest(http.MethodGet, "/widgets/42", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if len(rec.reqs) != 1 {
		t.Fatalf("expected 1 observation, got %d", len(rec.reqs))
	}
	got := rec.reqs[0]
	// routeLabel strips the "GET " prefix from r.Pattern.
	if got.route != "/widgets/{id}" {
		t.Errorf("route = %q, want %q", got.route, "/widgets/{id}")
	}
	if got.method != http.MethodGet {
		t.Errorf("method = %q, want GET", got.method)
	}
	if got.status != http.StatusOK {
		t.Errorf("status = %d, want 200", got.status)
	}
	if got.dur <= 0 {
		t.Errorf("dur not captured: %v", got.dur)
	}
}

// TestRequestsCapturesNon200 ensures a 404 handler status is recorded.
func TestRequestsCapturesNon200(t *testing.T) {
	rec := &recorder{}
	h := observability.Requests(rec)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))

	r := httptest.NewRequest(http.MethodGet, "/missing", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if len(rec.reqs) != 1 {
		t.Fatalf("expected 1 observation, got %d", len(rec.reqs))
	}
	if rec.reqs[0].status != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.reqs[0].status)
	}
}

// TestRequestsCapturesResponseBytes checks that body bytes written by the
// handler are counted in the respBytes label.
func TestRequestsCapturesResponseBytes(t *testing.T) {
	rec := &recorder{}
	body := []byte(`{"data":"hello"}`)

	h := observability.Requests(rec)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))

	r := httptest.NewRequest(http.MethodGet, "/things", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if len(rec.reqs) != 1 {
		t.Fatalf("expected 1 observation, got %d", len(rec.reqs))
	}
	if rec.reqs[0].respBytes != len(body) {
		t.Errorf("respBytes = %d, want %d", rec.reqs[0].respBytes, len(body))
	}
}

// TestRequestsFallbackRouteWhenNoPattern verifies that a request not dispatched
// by a pattern-aware mux still produces a bounded route label ("unknown").
func TestRequestsFallbackRouteWhenNoPattern(t *testing.T) {
	rec := &recorder{}
	h := observability.Requests(rec)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/any/path", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if len(rec.reqs) != 1 {
		t.Fatalf("expected 1 observation, got %d", len(rec.reqs))
	}
	if rec.reqs[0].route != "unknown" {
		t.Errorf("route = %q, want %q", rec.reqs[0].route, "unknown")
	}
}

// ---------- AccessLog middleware ----------

// logLine is the shape of one slog JSON output line.
type logLine struct {
	Level     string `json:"level"`
	Msg       string `json:"msg"`
	RequestID string `json:"request_id"`
	Method    string `json:"method"`
	Status    int    `json:"status"`
	DurMs     *int64 `json:"dur_ms"`
	Bytes     *int   `json:"bytes"`
}

// TestAccessLogEmitsStructuredLine verifies AccessLog writes one JSON log line
// per request containing request_id (from context) and response status.
func TestAccessLogEmitsStructuredLine(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	h := httpx.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}),
		httpx.RequestID(),
		observability.AccessLog(logger),
	)

	req := httptest.NewRequest(http.MethodPost, "/things", nil)
	req.Header.Set("X-Request-Id", "test-rid-1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if buf.Len() == 0 {
		t.Fatal("AccessLog wrote no output")
	}
	var line logLine
	if err := json.Unmarshal(bytes.TrimRight(buf.Bytes(), "\n"), &line); err != nil {
		t.Fatalf("log line not valid JSON: %v\nraw: %s", err, buf.String())
	}
	if line.Msg != "request" {
		t.Errorf("msg = %q, want %q", line.Msg, "request")
	}
	if line.RequestID != "test-rid-1" {
		t.Errorf("request_id = %q, want %q", line.RequestID, "test-rid-1")
	}
	if line.Status != http.StatusCreated {
		t.Errorf("status = %d, want 201", line.Status)
	}
	if line.Method != http.MethodPost {
		t.Errorf("method = %q, want POST", line.Method)
	}
	if line.DurMs == nil {
		t.Error("dur_ms missing from log line")
	}
	if line.Bytes == nil {
		t.Error("bytes missing from log line")
	}
}
