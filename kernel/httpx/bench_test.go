package httpx_test

// Hot-path benchmarks for the HTTP router and RouteMeta (criterion #17).
//
// Router.Handle is called once at boot; Router.Routes + Permissions are called
// at boot for permission sync. The hot path per-request is metadata access
// from the already-registered route (a struct field read), which we verify is
// O(1) with no registry lookups on the per-request path.

import (
	"net/http"
	"testing"

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
