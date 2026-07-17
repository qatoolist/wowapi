// Package prometheus implements the observability.Metrics port using the
// Prometheus Go client. Wire New() into the kernel and mount Handler() at a
// /metrics endpoint on an internal (non-public) mux.
//
// Wiring example:
//
//	m := promadapter.New()
//	internalMux.Handle("/metrics", m.Handler())
//	// then pass m wherever observability.Metrics is expected
package prometheus

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/qatoolist/wowapi/v2/kernel/observability"
)

// Compile-time assertion: *Prometheus implements observability.Metrics.
var _ observability.Metrics = (*Prometheus)(nil)

// Prometheus is an observability.Metrics backed by a private prometheus
// Registry. Construct with New().
type Prometheus struct {
	reg      *prometheus.Registry
	requests *prometheus.HistogramVec

	counterMu sync.Mutex
	counters  map[string]*prometheus.CounterVec

	gaugeMu sync.Mutex
	gauges  map[string]*prometheus.GaugeVec

	histogramMu sync.Mutex
	histograms  map[string]*prometheus.HistogramVec
}

// New returns a Prometheus with a fresh, isolated registry pre-registered with
// the HTTP request duration histogram.
func New() *Prometheus {
	reg := prometheus.NewRegistry()
	requests := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds, by route, method, and status.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"route", "method", "status"},
	)
	reg.MustRegister(requests)
	return &Prometheus{
		reg:        reg,
		requests:   requests,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
	}
}

// ObserveRequest records one HTTP request in the http_request_duration_seconds
// histogram.
func (p *Prometheus) ObserveRequest(route, method string, status int, dur time.Duration, _ int) {
	p.requests.WithLabelValues(route, method, strconv.Itoa(status)).Observe(dur.Seconds())
}

// IncCounter increments a named counter by value. The first call for a given
// name registers a CounterVec whose label names are the sorted keys of labels;
// all subsequent calls must use the same set of label keys.
func (p *Prometheus) IncCounter(name string, value float64, labels map[string]string) {
	cv := p.getOrRegisterCounter(name, labels)
	cv.With(safeLabels(labels)).Add(value)
}

// ObserveHistogram records a value in a lazily registered histogram.
func (p *Prometheus) ObserveHistogram(name string, value float64, labels map[string]string) {
	hv := p.getOrRegisterHistogram(name, labels)
	hv.With(safeLabels(labels)).Observe(value)
}

// SetGauge sets a named gauge to value. The first call for a given name
// registers a GaugeVec whose label names are the sorted keys of labels; all
// subsequent calls must use the same set of label keys.
func (p *Prometheus) SetGauge(name string, value float64, labels map[string]string) {
	gv := p.getOrRegisterGauge(name, labels)
	gv.With(safeLabels(labels)).Set(value)
}

// Handler returns an http.Handler that serves the Prometheus text format.
// Mount it on an internal mux, not the public API mux.
func (p *Prometheus) Handler() http.Handler {
	return promhttp.HandlerFor(p.reg, promhttp.HandlerOpts{})
}

// Gatherer exposes the internal prometheus.Gatherer for integration testing
// and health-check scraping without depending on the Prometheus type directly.
func (p *Prometheus) Gatherer() prometheus.Gatherer { return p.reg }

func (p *Prometheus) getOrRegisterCounter(name string, labels map[string]string) *prometheus.CounterVec {
	p.counterMu.Lock()
	defer p.counterMu.Unlock()
	if cv, ok := p.counters[name]; ok {
		return cv
	}
	cv := prometheus.NewCounterVec(
		prometheus.CounterOpts{Name: name, Help: name},
		sortedKeys(labels),
	)
	p.reg.MustRegister(cv)
	p.counters[name] = cv
	return cv
}

func (p *Prometheus) getOrRegisterGauge(name string, labels map[string]string) *prometheus.GaugeVec {
	p.gaugeMu.Lock()
	defer p.gaugeMu.Unlock()
	if gv, ok := p.gauges[name]; ok {
		return gv
	}
	gv := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{Name: name, Help: name},
		sortedKeys(labels),
	)
	p.reg.MustRegister(gv)
	p.gauges[name] = gv
	return gv
}

func (p *Prometheus) getOrRegisterHistogram(name string, labels map[string]string) *prometheus.HistogramVec {
	p.histogramMu.Lock()
	defer p.histogramMu.Unlock()
	if hv, ok := p.histograms[name]; ok {
		return hv
	}
	hv := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{Name: name, Help: name, Buckets: prometheus.DefBuckets},
		sortedKeys(labels),
	)
	p.reg.MustRegister(hv)
	p.histograms[name] = hv
	return hv
}

// sortedKeys returns the keys of m in ascending order. A stable order is
// required because prometheus uses label names positionally when creating Vecs.
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// safeLabels converts a possibly-nil map to prometheus.Labels, avoiding
// nil-map panics in prometheus.CounterVec.With and GaugeVec.With.
func safeLabels(m map[string]string) prometheus.Labels {
	if m == nil {
		return prometheus.Labels{}
	}
	return prometheus.Labels(m)
}
