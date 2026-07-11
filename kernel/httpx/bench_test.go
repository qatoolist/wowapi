package httpx_test

// Hot-path benchmarks for the HTTP router and RouteMeta (criterion #17).
//
// Router.Handle is called once at boot; Router.Routes + Permissions are called
// at boot for permission sync. The hot path per-request is metadata access
// from the already-registered route (a struct field read), which we verify is
// O(1) with no registry lookups on the per-request path.

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/policy"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/kernel/validation"
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

// benchmarkTokenBucketSweepAt builds a TokenBucket pre-populated with n
// one-shot keys (each hit exactly once — no key is ever looked up twice), all
// idle past idleTTL, then times sweep cost via SweepForTest. This isolates
// sweep's O(N) scan cost (PERF-01: sweep runs synchronously on the request
// path once the map reaches sweepAt) independent of Allow's own per-call cost,
// which BenchmarkTokenBucketAllow already covers.
func benchmarkTokenBucketSweepAt(b *testing.B, n int) {
	fixed := time.Unix(1_700_000_000, 0)
	clockAt := fixed
	tb := httpx.NewTokenBucketWithOptions(1, 5, func() time.Time { return clockAt })
	for i := 0; i < n; i++ {
		_, _ = tb.Allow("k" + strconv.Itoa(i)) // each key hit exactly once
	}
	swept := fixed.Add(11 * time.Minute) // past idleTTL for every key above

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpx.SweepForTest(tb, swept)
	}
}

// BenchmarkTokenBucketSweepAt10k measures sweep cost over 10,000 idle one-shot
// keys — the production sweepAt threshold (roadmap PERF-01).
func BenchmarkTokenBucketSweepAt10k(b *testing.B) {
	benchmarkTokenBucketSweepAt(b, 10_000)
}

// BenchmarkTokenBucketSweepAt100k measures sweep cost over 100,000 idle
// one-shot keys — 10x the production sweepAt threshold, proving the O(N) scan
// stays linear rather than degrading further (roadmap PERF-01).
func BenchmarkTokenBucketSweepAt100k(b *testing.B) {
	benchmarkTokenBucketSweepAt(b, 100_000)
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

// ── BenchmarkDispatch (backlog B5 / competitive-benchmark §Routing) ─────────
//
// The gap the competitive benchmark named: no benchmark exercised steady-state
// request dispatch through the REAL chain — ServeMux route lookup ->
// SecureHandler's gateRoute (authN -> tenant/actor bind -> authZ) -> handler —
// at product-scale route cardinality. This proves whether net/http.ServeMux's
// dispatch cost matters relative to the authz/DB budget before any router
// replacement (B11, P2) is considered.
//
// Route mix per buildDispatchRoutes: static, single path-param, and two-level
// path-param patterns (Go 1.22 ServeMux "{id}" syntax), matching how real
// modules register list/create/read/update/delete routes. Auth/authz are
// exercised the same cheap way the gate unit tests do (fakeAuth/fakeEval/
// fakeTxM from fakes_test.go) — no real DB, isolating ServeMux + gate overhead
// from database latency (which is measured separately in kernel/audit and
// kernel/sequence's DB-backed benches).

// buildDispatchRoutes registers n routes on a fresh Router: a repeating
// static/list/create/read/update/delete quintet per resource, so at n=2000 the
// mux holds a realistic mix of static and "{id}"-param patterns rather than n
// copies of one shape.
func buildDispatchRoutes(n int) *httpx.Router {
	r := httpx.NewRouter()
	verbs := []struct {
		method string
		suffix string
		perm   string
		public bool
	}{
		{http.MethodGet, "", "list", false},
		{http.MethodPost, "", "create", false},
		{http.MethodGet, "/{id}", "read", false},
		{http.MethodPatch, "/{id}", "update", false},
		{http.MethodDelete, "/{id}", "deactivate", false},
	}
	registered := 0
	for res := 0; registered < n; res++ {
		base := "/v1/resource" + strconv.Itoa(res)
		for _, v := range verbs {
			if registered >= n {
				break
			}
			meta := httpx.RouteMeta{Permission: "bench.resource" + strconv.Itoa(res) + "." + v.perm}
			if v.public {
				meta = httpx.RouteMeta{Public: true}
			}
			r.Handle(v.method, base+v.suffix, meta, func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			registered++
		}
	}
	return r
}

// dispatchTargets returns a handful of representative request paths spread
// across the registered route set (first, middle, last resource; static and
// {id} patterns), so the benchmark measures typical dispatch rather than only
// the first-registered route.
func dispatchTargets(n int) []*http.Request {
	resources := n / 5
	if resources < 1 {
		resources = 1
	}
	pick := func(i int) int { return i % resources }
	idxs := []int{pick(0), pick(resources / 2), pick(resources - 1)}
	var reqs []*http.Request
	for _, i := range idxs {
		base := "/v1/resource" + strconv.Itoa(i)
		reqs = append(
			reqs,
			httptest.NewRequest(http.MethodGet, base, nil),
			httptest.NewRequest(http.MethodGet, base+"/00000000-0000-0000-0000-000000000001", nil),
			httptest.NewRequest(http.MethodPatch, base+"/00000000-0000-0000-0000-000000000001", nil),
		)
	}
	return reqs
}

func benchmarkDispatchAt(b *testing.B, n int) {
	r := buildDispatchRoutes(n)
	if err := r.Err(); err != nil {
		b.Fatalf("setup: %v", err)
	}
	act := authz.Actor{Kind: authz.ActorUser, UserID: uuid.New(), CapacityID: uuid.New(), TenantID: uuid.New()}
	mux := r.SecureHandler(fakeAuth{actor: act}, fakeEval{dec: authz.Decision{Allowed: true}}, fakeTxM{})
	reqs := dispatchTargets(n)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, reqs[i%len(reqs)])
	}
}

// BenchmarkDispatch measures steady-state dispatch at increasing route
// cardinality (50/500/2000), serial (single goroutine).
func BenchmarkDispatch(b *testing.B) {
	for _, n := range []int{50, 500, 2000} {
		b.Run(strconv.Itoa(n)+"routes", func(b *testing.B) {
			benchmarkDispatchAt(b, n)
		})
	}
}

// BenchmarkDispatchParallel is the b.RunParallel companion: multiple
// goroutines dispatch concurrently through the same mux, exercising any
// lock contention in ServeMux's route match or the gate's per-request path
// (there is none expected — no shared mutable state on the hot path — but this
// proves it rather than assuming it).
func BenchmarkDispatchParallel(b *testing.B) {
	for _, n := range []int{50, 500, 2000} {
		b.Run(strconv.Itoa(n)+"routes", func(b *testing.B) {
			r := buildDispatchRoutes(n)
			if err := r.Err(); err != nil {
				b.Fatalf("setup: %v", err)
			}
			act := authz.Actor{Kind: authz.ActorUser, UserID: uuid.New(), CapacityID: uuid.New(), TenantID: uuid.New()}
			mux := r.SecureHandler(fakeAuth{actor: act}, fakeEval{dec: authz.Decision{Allowed: true}}, fakeTxM{})
			reqs := dispatchTargets(n)

			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					rec := httptest.NewRecorder()
					mux.ServeHTTP(rec, reqs[i%len(reqs)])
					i++
				}
			})
		})
	}
}

// ── Authz-gate: cached vs uncached store (backlog B5) ───────────────────────
//
// gateRoute takes an authz.Evaluator, which is built over an authz.Store. This
// benches the same gate path with two Store wirings: a CachingStore front (the
// framework's real opt-in decorator, kernel/authz/caching.go) primed for an
// all-hits steady state, and a raw in-memory fake store with no caching (every
// call pays the store's full cost). Both are in-memory — no DB — matching the
// existing kernel/authz cache benchmarks (caching_bench_test.go), which never
// exercise a real database either. This scopes out the uncached-against-a-real-
// database path per the backlog's stated limitation: the existing bench suite
// only touches Postgres in kernel/audit/kernel/sequence, and this package's
// benches follow that precedent rather than introducing a new DB dependency.

// benchStore is a minimal in-memory authz.Store fake: fixed assignments, no
// org hierarchy, no policies. Used directly (uncached path) and wrapped in
// authz.NewCachingStore (cached path). calls counts ActiveAssignments reads so
// the cached benchmark can PROVE its loop was all hits (calls stays 1).
type benchStore struct {
	asgs  []authz.Assignment
	calls int
}

func (s *benchStore) ActiveAssignments(context.Context, database.TenantDB, authz.Actor, time.Time) ([]authz.Assignment, error) {
	s.calls++
	return s.asgs, nil
}

func (s *benchStore) OrgAncestors(context.Context, database.TenantDB, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (s *benchStore) OrgSubtree(context.Context, database.TenantDB, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (s *benchStore) Policies(context.Context, database.TenantDB, authz.Actor, string, string) ([]authz.Policy, error) {
	return nil, nil
}

func (s *benchStore) ResourceOrg(context.Context, database.TenantDB, resource.Ref) (uuid.UUID, error) {
	return uuid.Nil, nil
}

func newBenchEvaluator(store authz.Store) authz.Evaluator {
	const perm = "bench.thing.read"
	reg := authz.NewRegistry()
	reg.Register(authz.Permission{Key: perm})
	if err := reg.Err(); err != nil {
		panic(err)
	}
	return authz.New(authz.Options{Store: store, Registry: reg, Policies: policy.New()})
}

func benchmarkAuthzGate(b *testing.B, store authz.Store, act authz.Actor) {
	router := httpx.NewRouter()
	router.Handle(http.MethodGet, "/thing", httpx.RouteMeta{Permission: "bench.thing.read"},
		func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	if err := router.Err(); err != nil {
		b.Fatalf("setup: %v", err)
	}
	mux := router.SecureHandler(fakeAuth{actor: act}, newBenchEvaluator(store), fakeTxM{})
	req := httptest.NewRequest(http.MethodGet, "/thing", nil)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
	}
}

// BenchmarkAuthzGateCachedHit exercises the gate over a CachingStore primed
// with the SAME actor the gate authenticates — CachingStore keys the cache on
// (tenant, capacity), so the primed entry is the one every timed request hits.
// This is the framework's recommended production wiring for the hot
// ActiveAssignments read.
func BenchmarkAuthzGateCachedHit(b *testing.B) {
	inner := &benchStore{asgs: []authz.Assignment{
		{RoleKey: "bench.reader", Perms: []string{"bench.thing.read"}},
	}}
	cached := authz.NewCachingStore(inner, time.Hour)
	act := authz.Actor{Kind: authz.ActorUser, UserID: uuid.New(), CapacityID: uuid.New(), TenantID: uuid.New()}
	// Prime the cache for this exact actor before timing so the loop is all hits.
	if _, err := cached.ActiveAssignments(context.Background(), nil, act, time.Now()); err != nil {
		b.Fatalf("prime: %v", err)
	}
	benchmarkAuthzGate(b, cached, act)
	if inner.calls != 1 {
		b.Fatalf("inner store called %d times, want 1 (timed loop must be all cache hits)", inner.calls)
	}
}

// BenchmarkAuthzGateUncachedMiss exercises the gate directly over the fake
// store with no caching decorator — every request pays the store's full
// per-call cost. LIMITATION: this fake store has O(1) in-memory lookups with
// no query latency, so it measures the gate's own per-request overhead
// (evaluator + engine bookkeeping) without caching, not a real database's
// round-trip cost. A real DB-backed uncached benchmark would need a Postgres
// fixture like kernel/audit/kernel/sequence's DB-backed benches use; this
// suite does not add one because the existing bench_test.go convention in
// this package never dials a database (only kernel/audit and kernel/sequence
// do, and only for their own hot paths).
func BenchmarkAuthzGateUncachedMiss(b *testing.B) {
	store := &benchStore{asgs: []authz.Assignment{
		{RoleKey: "bench.reader", Perms: []string{"bench.thing.read"}},
	}}
	act := authz.Actor{Kind: authz.ActorUser, UserID: uuid.New(), CapacityID: uuid.New(), TenantID: uuid.New()}
	benchmarkAuthzGate(b, store, act)
}

// ── JSON decode / body-limit (backlog B5) ───────────────────────────────────
//
// kernel/httpx/decode.go's DecodeJSON is the hot per-request body path for
// every create/update handler (strict decode, unknown-field rejection, size
// cap). This benches the typical-payload path and the oversized-rejected path
// so a regression in either (e.g. accidentally buffering the whole body before
// capping it) is caught.

type benchPayload struct {
	Name   string   `json:"name"`
	Email  string   `json:"email"`
	Age    int      `json:"age"`
	Tags   []string `json:"tags"`
	Active bool     `json:"active"`
}

var benchPayloadJSON = []byte(`{"name":"Ada Lovelace","email":"ada@example.com","age":36,"tags":["math","engine","analyst"],"active":true}`)

// BenchmarkDecodeJSONTypical decodes a realistic small JSON body under the
// framework's standard body-limit budget.
func BenchmarkDecodeJSONTypical(b *testing.B) {
	req := httptest.NewRequest(http.MethodPost, "/v1/things", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(bytes.NewReader(benchPayloadJSON))
		if _, err := httpx.DecodeJSON[benchPayload](req, 1<<20); err != nil {
			b.Fatalf("decode: %v", err)
		}
	}
}

// BenchmarkDecodeJSONOversizedRejected feeds a body over the limit so every
// iteration exercises the MaxBytesReader rejection path (the size guard must
// stay cheap — reject before parsing a large body).
func BenchmarkDecodeJSONOversizedRejected(b *testing.B) {
	oversized := append(bytes.Repeat([]byte(" "), 64), benchPayloadJSON...) // pad past a tiny limit
	const limit = 32                                                        // far below the payload; every call must reject
	req := httptest.NewRequest(http.MethodPost, "/v1/things", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(bytes.NewReader(oversized))
		if _, err := httpx.DecodeJSON[benchPayload](req, limit); err == nil {
			b.Fatal("expected oversized body to be rejected")
		}
	}
}

// BenchmarkBindAndValidateTypical exercises the full BindAndValidate path
// (decode + struct-tag validation) used by handlers, not just DecodeJSON.
func BenchmarkBindAndValidateTypical(b *testing.B) {
	v := validation.New()
	req := httptest.NewRequest(http.MethodPost, "/v1/things", nil)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req.Body = io.NopCloser(bytes.NewReader(benchPayloadJSON))
		if _, err := httpx.BindAndValidate[benchPayload](req, v, 1<<20); err != nil {
			b.Fatalf("bind: %v", err)
		}
	}
}
