package observability

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/qatoolist/wowapi/v2/kernel/httpx"
)

// Requests returns a httpx.Middleware that records RED (Rate, Errors, Duration)
// metrics for every request via m.ObserveRequest.
//
// The route label is taken from r.Pattern (populated by net/http.ServeMux in
// Go 1.22+); the method prefix ("GET ") is stripped because method is already
// a separate label. When r.Pattern is empty (handler not dispatched by a
// pattern-aware mux) the label falls back to "unknown", keeping cardinality
// bounded.
//
// Position in the chain — after RequestID and Recover (which must be outermost),
// wrapping the handler tightly:
//
//	httpx.Chain(handler, httpx.RequestID(), httpx.Recover(log), observability.Requests(m))
func Requests(m Metrics) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w}
			start := time.Now()
			next.ServeHTTP(sw, r)
			route := routeLabel(r.Pattern)
			m.ObserveRequest(route, r.Method, sw.statusCode(), time.Since(start), sw.written)
		})
	}
}

// AccessLog returns a httpx.Middleware that emits one structured INFO line per
// request carrying: request_id (from httpx.RequestIDFrom), method, route
// (r.Pattern), status, dur_ms, and bytes. Allocations are limited to the slog
// call itself.
//
// Position in the chain alongside Requests:
//
//	httpx.Chain(handler, httpx.RequestID(), httpx.Recover(log),
//	    observability.Requests(m), observability.AccessLog(log))
func AccessLog(logger *slog.Logger) httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := &statusWriter{ResponseWriter: w}
			start := time.Now()
			next.ServeHTTP(sw, r)
			logger.InfoContext(r.Context(), "request",
				"request_id", httpx.RequestIDFrom(r.Context()),
				"method", r.Method,
				"route", r.Pattern,
				"status", sw.statusCode(),
				"dur_ms", time.Since(start).Milliseconds(),
				"bytes", sw.written,
			)
		})
	}
}

// statusWriter wraps http.ResponseWriter to capture the HTTP status code and
// the total bytes written to the response body.
type statusWriter struct {
	http.ResponseWriter
	status  int // 0 until WriteHeader or first Write
	written int // total body bytes written
}

// WriteHeader records the status code and delegates to the underlying writer.
// Only the first call takes effect; subsequent calls are ignored by the
// underlying http.ResponseWriter too, so no double-reporting.
func (sw *statusWriter) WriteHeader(code int) {
	if sw.status == 0 {
		sw.status = code
	}
	sw.ResponseWriter.WriteHeader(code)
}

// Write records body bytes and, if WriteHeader was not yet called, records an
// implicit 200.
func (sw *statusWriter) Write(b []byte) (int, error) {
	if sw.status == 0 {
		sw.status = http.StatusOK
	}
	n, err := sw.ResponseWriter.Write(b)
	sw.written += n
	return n, err
}

// Unwrap returns the underlying ResponseWriter, enabling access to optional
// interfaces (http.Flusher, http.Hijacker) via http.ResponseController.
func (sw *statusWriter) Unwrap() http.ResponseWriter { return sw.ResponseWriter }

// statusCode returns the captured status code, defaulting to 200 when neither
// WriteHeader nor Write was called (empty 200 response).
func (sw *statusWriter) statusCode() int {
	if sw.status == 0 {
		return http.StatusOK
	}
	return sw.status
}

// routeLabel produces a bounded-cardinality route label from r.Pattern.
// Go 1.22+ ServeMux includes the method in the pattern ("GET /path/{id}");
// we strip it since method is already a separate label dimension.
func routeLabel(pattern string) string {
	if pattern == "" {
		return "unknown"
	}
	if sp := strings.IndexByte(pattern, ' '); sp >= 0 {
		return pattern[sp+1:]
	}
	return pattern
}
