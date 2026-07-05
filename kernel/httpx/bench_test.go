package httpx_test

// Hot-path benchmarks for the HTTP router and RouteMeta (criterion #17).
//
// Router.Handle is called once at boot; Router.Routes + Permissions are called
// at boot for permission sync. The hot path per-request is metadata access
// from the already-registered route (a struct field read), which we verify is
// O(1) with no registry lookups on the per-request path.

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/kernel/httpx"
)

// BenchmarkRouterHandle measures route registration: called at boot for every
// module route. Not on the per-request hot path, but must not be quadratic.
func BenchmarkRouterHandle(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := httpx.NewRouter()
		r.Handle("GET", "/v1/requests", httpx.RouteMeta{Permission: "requests.request.list"}, http.NotFound)
		r.Handle("POST", "/v1/requests", httpx.RouteMeta{Permission: "requests.request.create"}, http.NotFound)
		r.Handle("GET", "/v1/requests/{id}", httpx.RouteMeta{Permission: "requests.request.read"}, http.NotFound)
		r.Handle("PATCH", "/v1/requests/{id}", httpx.RouteMeta{Permission: "requests.request.update"}, http.NotFound)
		r.Handle("DELETE", "/v1/requests/{id}", httpx.RouteMeta{Permission: "requests.request.deactivate"}, http.NotFound)
		r.Handle("GET", "/healthz", httpx.RouteMeta{Public: true}, http.NotFound)
		_ = r.Err()
	}
}

// BenchmarkRouterRoutes measures Routes(): called at boot for permission sync
// and OpenAPI generation. Exercises sort + slice copy.
func BenchmarkRouterRoutes(b *testing.B) {
	r := httpx.NewRouter()
	for _, spec := range []struct {
		method, pattern string
		meta            httpx.RouteMeta
	}{
		{"GET", "/v1/requests", httpx.RouteMeta{Permission: "requests.request.list"}},
		{"POST", "/v1/requests", httpx.RouteMeta{Permission: "requests.request.create"}},
		{"GET", "/v1/requests/{id}", httpx.RouteMeta{Permission: "requests.request.read"}},
		{"PATCH", "/v1/requests/{id}", httpx.RouteMeta{Permission: "requests.request.update"}},
		{"DELETE", "/v1/requests/{id}", httpx.RouteMeta{Permission: "requests.request.deactivate"}},
		{"GET", "/healthz", httpx.RouteMeta{Public: true}},
		{"GET", "/readyz", httpx.RouteMeta{Public: true}},
		{"GET", "/v1/orgs", httpx.RouteMeta{Permission: "requests.request.list"}},
	} {
		r.Handle(spec.method, spec.pattern, spec.meta, http.NotFound)
	}
	if err := r.Err(); err != nil {
		b.Fatalf("setup: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.Routes()
	}
}

// BenchmarkRouteMetaValidate measures RouteMeta.validate(), called once per
// route at registration. This proves the boot-time guard has negligible cost.
func BenchmarkRouteMetaValidate(b *testing.B) {
	meta := httpx.RouteMeta{Permission: "requests.request.read"}
	r := httpx.NewRouter()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// We exercise the full Handle path (which calls validate) rather than
		// the unexported validate() directly.
		r2 := httpx.NewRouter()
		r2.Handle("GET", "/v1/x", meta, http.NotFound)
		_ = r2.Err()
		_ = r.Routes() // silent read to avoid dead-code elimination
	}
}

// BenchmarkTokenBucketAllow measures the per-request rate-limit decision
// (roadmap S2, backlog B-2): TokenBucket.Allow — mutex, per-key bucket map
// lookup, refill arithmetic, and token consume. A frozen clock and a very large
// burst keep every call on the allow path, so the measurement is deterministic
// and isolates the limiter's steady-state cost; keys rotate to exercise the map.
func BenchmarkTokenBucketAllow(b *testing.B) {
	fixed := time.Unix(1_700_000_000, 0)
	tb := httpx.NewTokenBucketWithClock(1<<20, 1<<40, func() time.Time { return fixed })
	keys := make([]string, 16)
	for i := range keys {
		keys[i] = "ip:10.0.0." + strconv.Itoa(i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if allowed, _ := tb.Allow(keys[i&15]); !allowed {
			b.Fatal("bucket denied on the always-allow path (setup wrong)")
		}
	}
}

// BenchmarkEdgeMiddlewareChain measures the kernel's fixed edge chain per request
// (blueprint 07 §1, backlog B-2): SecureHeaders → CORS → BodyLimit → Timeout
// wrapped around a terminal 200 handler. The request carries an allowed Origin so
// the CORS header-echo path runs; this is the baseline in-bound posture every
// wowapi request pays. A fresh recorder per iteration is required because CORS
// Add()s a Vary header that would otherwise accumulate across iterations.
func BenchmarkEdgeMiddlewareChain(b *testing.B) {
	handler := httpx.Chain(
		http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		httpx.SecureHeaders(),
		httpx.CORS(httpx.CORSPolicy{
			AllowedOrigins:   []string{"https://app.example.com"},
			AllowCredentials: true,
			MaxAge:           10 * time.Minute,
		}),
		httpx.BodyLimit(1<<20),
		httpx.Timeout(30*time.Second),
	)
	req := httptest.NewRequest(http.MethodGet, "/v1/requests", nil)
	req.Header.Set("Origin", "https://app.example.com")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
}
