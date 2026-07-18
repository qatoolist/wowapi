package requestbench

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	oteladapter "github.com/qatoolist/wowapi/adapters/tracing/otel"
	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/config"
	"github.com/qatoolist/wowapi/kernel/database"
	"github.com/qatoolist/wowapi/kernel/httpx"
	"github.com/qatoolist/wowapi/kernel/model"
	"github.com/qatoolist/wowapi/kernel/outbox"
	"github.com/qatoolist/wowapi/kernel/policy"
	"github.com/qatoolist/wowapi/kernel/resource"
	"github.com/qatoolist/wowapi/testkit"
)

var workloadProfiles = []string{
	"public", "authenticated-read", "authenticated-write",
	"resource-authz", "idempotent-write", "async-enqueue",
}

var (
	cacheStates            = []string{"cold", "warm"}
	concurrentTenantCounts = []int{1, 10, 100}
	costComponents         = []string{
		"pool_wait", "tx_setup", "authz_query", "handler_query", "serialization", "middleware",
	}
)

const (
	readPermission   = "perf.request.read"
	createPermission = "perf.request.create"
	seedValue        = 70701001
)

type tenantFixture struct {
	tenant, user, capacity, org uuid.UUID
	resource                    resource.Ref
}

type batchResult struct {
	Profile, Cache                          string
	ConcurrentTenants, Requests, Errors     int
	P50, P95, P99                           time.Duration
	Allocations                             float64
	SQLStatements                           int
	Bytes                                   int64
	Elapsed, PoolWait, TxDuration, LockWait time.Duration
	PlanHash                                string
	Cost                                    map[string]time.Duration
}

type publishedResult struct {
	Name                    string             `json:"name"`
	Profile                 string             `json:"profile"`
	Cache                   string             `json:"cache"`
	ConcurrentTenants       int                `json:"concurrent_tenants"`
	P50NS                   int64              `json:"p50_ns"`
	P95NS                   int64              `json:"p95_ns"`
	P99NS                   int64              `json:"p99_ns"`
	AllocationsPerRequest   float64            `json:"allocations_per_request"`
	SQLStatementsPerRequest float64            `json:"sql_statements_per_request"`
	BytesPerRequest         float64            `json:"bytes_per_request"`
	PoolWaitNSPerRequest    int64              `json:"pool_wait_ns_per_request"`
	TransactionNSPerRequest int64              `json:"transaction_ns_per_request"`
	LockWaitNSPerRequest    int64              `json:"lock_wait_ns_per_request"`
	PlanHash                string             `json:"plan_hash"`
	CostNSPerRequest        map[string]int64   `json:"cost_ns_per_request"`
	RelativeToReference     map[string]float64 `json:"relative_to_reference"`
}

type publication struct {
	SchemaVersion     int               `json:"schema_version"`
	Reference         string            `json:"reference"`
	ComparisonKind    string            `json:"comparison_kind"`
	AbsoluteSLOStatus string            `json:"absolute_slo_status"`
	SourceRevision    string            `json:"source_revision"`
	MeasuredAt        time.Time         `json:"measured_at"`
	Environment       map[string]string `json:"environment"`
	AttributionMethod string            `json:"attribution_method"`
	LockWaitMethod    string            `json:"lock_wait_method"`
	Results           []publishedResult `json:"results"`
}

type requestSuite struct {
	tb         testing.TB
	h          *testkit.DBHandle
	pool       *pgxpool.Pool
	txm        *recordingTxManager
	exporter   *tracetest.InMemoryExporter
	tracer     *oteladapter.Tracer
	tenants    []tenantFixture
	coldEval   authz.Evaluator
	warmEval   authz.Evaluator
	planHashes map[string]string
	seq        atomic.Uint64
}

type metricTotals struct {
	mu            sync.Mutex
	txSetup       time.Duration
	txDuration    time.Duration
	authzQuery    time.Duration
	handlerQuery  time.Duration
	serialization time.Duration
}

func (m *metricTotals) add(dst *time.Duration, d time.Duration) {
	m.mu.Lock()
	*dst += d
	m.mu.Unlock()
}

func (m *metricTotals) snapshot() metricTotals {
	m.mu.Lock()
	defer m.mu.Unlock()
	return metricTotals{txSetup: m.txSetup, txDuration: m.txDuration, authzQuery: m.authzQuery, handlerQuery: m.handlerQuery, serialization: m.serialization}
}

func (m *metricTotals) reset() {
	m.mu.Lock()
	m.txSetup, m.txDuration, m.authzQuery, m.handlerQuery, m.serialization = 0, 0, 0, 0, 0
	m.mu.Unlock()
}

type recordingTxManager struct {
	inner   database.TxManager
	metrics metricTotals
}

func (m *recordingTxManager) WithTenant(ctx context.Context, fn func(context.Context, database.TenantDB) error) error {
	return m.measure(ctx, false, fn)
}

func (m *recordingTxManager) WithTenantRO(ctx context.Context, fn func(context.Context, database.TenantDB) error) error {
	return m.measure(ctx, true, fn)
}

func (m *recordingTxManager) Platform(ctx context.Context, fn func(context.Context, database.DB) error) error {
	return m.inner.Platform(ctx, fn)
}

func (m *recordingTxManager) measure(ctx context.Context, readOnly bool, fn func(context.Context, database.TenantDB) error) error {
	var callback time.Duration
	wrapped := func(ctx context.Context, db database.TenantDB) error {
		start := time.Now()
		err := fn(ctx, db)
		callback = time.Since(start)
		return err
	}
	start := time.Now()
	var err error
	if readOnly {
		err = m.inner.WithTenantRO(ctx, wrapped)
	} else {
		err = m.inner.WithTenant(ctx, wrapped)
	}
	total := time.Since(start)
	setup := total - callback
	if setup < 0 {
		setup = 0
	}
	m.metrics.add(&m.metrics.txSetup, setup)
	m.metrics.add(&m.metrics.txDuration, total)
	return err
}

type timedEvaluator struct {
	inner   authz.Evaluator
	metrics *metricTotals
}

func (e timedEvaluator) Evaluate(ctx context.Context, db database.TenantDB, actor authz.Actor, permission string, target authz.Target) (authz.Decision, error) {
	start := time.Now()
	decision, err := e.inner.Evaluate(ctx, db, actor, permission, target)
	e.metrics.add(&e.metrics.authzQuery, time.Since(start))
	return decision, err
}

func (e timedEvaluator) Filter(ctx context.Context, db database.TenantDB, actor authz.Actor, permission, resourceType string) (authz.ListFilter, error) {
	start := time.Now()
	filter, err := e.inner.Filter(ctx, db, actor, permission, resourceType)
	e.metrics.add(&e.metrics.authzQuery, time.Since(start))
	return filter, err
}

type fixtureAuthenticator struct{ tenants []tenantFixture }

func (a fixtureAuthenticator) Authenticate(r *http.Request) (authz.Actor, error) {
	i, err := strconv.Atoi(r.Header.Get("X-Perf-Tenant"))
	if err != nil || i < 0 || i >= len(a.tenants) {
		return authz.Actor{}, errors.New("invalid benchmark tenant")
	}
	t := a.tenants[i]
	return authz.Actor{Kind: authz.ActorUser, UserID: t.user, CapacityID: t.capacity, TenantID: t.tenant}, nil
}

func newRequestSuite(tb testing.TB) *requestSuite {
	tb.Helper()
	h := testkit.NewDB(tb)
	dsn := os.Getenv("WOWAPI_TEST_DSN")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	fixtureDSN, err := dsnForDatabase(dsn, h.Name)
	if err != nil {
		tb.Fatalf("build benchmark fixture DSN: %v", err)
	}
	exp := tracetest.NewInMemoryExporter()
	tr := oteladapter.New(exp, 1)
	dbCfg := config.Defaults().DB
	dbCfg.MaxConns = 32
	pool, err := database.NewPool(context.Background(), fixtureDSN, dbCfg,
		database.WithSetRole("app_rt"), database.WithConnRLSGuard(), database.WithQueryTracer(tr))
	if err != nil {
		tb.Fatalf("build traced runtime pool: %v", err)
	}
	tb.Cleanup(func() { _ = tr.Shutdown(context.Background()); pool.Close() })
	baseTxM := database.NewManager(pool, dbCfg, database.WithRole("app_rt"), database.WithRLSGuard())
	s := &requestSuite{tb: tb, h: h, pool: pool, exporter: exp, tracer: tr}
	s.txm = &recordingTxManager{inner: baseTxM}
	s.seed()
	reg := authz.NewRegistry()
	reg.Register(authz.Permission{Key: readPermission})
	reg.Register(authz.Permission{Key: createPermission})
	if err := reg.Err(); err != nil {
		tb.Fatalf("permission registry: %v", err)
	}
	s.coldEval = authz.New(authz.Options{Store: authz.NewStore(), Registry: reg, Policies: policy.New()})
	s.warmEval = authz.New(authz.Options{Store: authz.NewCachingStore(authz.NewStore(), time.Hour), Registry: reg, Policies: policy.New()})
	s.planHashes = make(map[string]string, len(workloadProfiles))
	for _, profile := range workloadProfiles {
		s.planHashes[profile] = s.queryPlanHash(profile)
	}
	if err := tr.ForceFlush(context.Background()); err != nil {
		tb.Fatalf("flush setup traces: %v", err)
	}
	exp.Reset()
	return s
}

func (s *requestSuite) seed() {
	ctx := context.Background()
	mustExecTB(s.tb, s.h.Admin, ctx, `INSERT INTO permissions (key,module,description) VALUES ($1,'perf',$1),($2,'perf',$2) ON CONFLICT (key) DO NOTHING`, readPermission, createPermission)
	mustExecTB(s.tb, s.h.Admin, ctx, `INSERT INTO resource_types (key,module,description) VALUES ('perf.request','perf','benchmark request') ON CONFLICT (key) DO NOTHING`)
	s.tenants = make([]tenantFixture, 100)
	for i := range s.tenants {
		tf := tenantFixture{
			tenant: fixedUUID("tenant", i), user: fixedUUID("user", i), capacity: fixedUUID("capacity", i), org: fixedUUID("org", i),
			resource: resource.Ref{Type: "perf.request", ID: fixedUUID("resource", i)},
		}
		role := fixedUUID("role", i)
		mustExecTB(s.tb, s.h.Admin, ctx, `INSERT INTO tenants (id,slug,display_name,created_by) VALUES ($1,$2,$3,$4)`, tf.tenant, fmt.Sprintf("perf-%03d", i), fmt.Sprintf("Performance Tenant %03d", i), uuid.Nil)
		mustExecTB(s.tb, s.h.Admin, ctx, `INSERT INTO users (id,idp_subject,email,created_by) VALUES ($1,$2,$3,$4)`, tf.user, fmt.Sprintf("perf-sub-%03d", i), fmt.Sprintf("perf-%03d@example.test", i), uuid.Nil)
		mustExecTB(s.tb, s.h.Admin, ctx, `INSERT INTO acting_capacities (id,tenant_id,user_id,label,created_by) VALUES ($1,$2,$3,'benchmark',$4)`, tf.capacity, tf.tenant, tf.user, uuid.Nil)
		mustExecTB(s.tb, s.h.Admin, ctx, `INSERT INTO organizations (id,tenant_id,name,created_by) VALUES ($1,$2,$3,$4)`, tf.org, tf.tenant, fmt.Sprintf("Benchmark Org %03d", i), uuid.Nil)
		for row := range 10 {
			resourceID := tf.resource.ID
			if row > 0 {
				resourceID = fixedUUID("resource-"+itoa(row), i)
			}
			mustExecTB(s.tb, s.h.Admin, ctx, `INSERT INTO resources (id,tenant_id,resource_type,org_id,label,status,created_by) VALUES ($1,$2,'perf.request',$3,$4,'active',$5)`, resourceID, tf.tenant, tf.org, fmt.Sprintf("benchmark-%02d", row), uuid.Nil)
		}
		mustExecTB(s.tb, s.h.Admin, ctx, `INSERT INTO roles (id,tenant_id,key,name,created_by) VALUES ($1,$2,'benchmark','Benchmark',$3)`, role, tf.tenant, uuid.Nil)
		mustExecTB(s.tb, s.h.Admin, ctx, `INSERT INTO role_permissions (role_id,permission_key) VALUES ($1,$2),($1,$3)`, role, readPermission, createPermission)
		mustExecTB(s.tb, s.h.Admin, ctx, `INSERT INTO actor_assignments (id,tenant_id,capacity_id,role_id,scope_kind,valid_from,granted_by,created_by) VALUES ($1,$2,$3,$4,'tenant',now()-interval '1 hour',$5,$5)`, fixedUUID("assignment", i), tf.tenant, tf.capacity, role, uuid.Nil)
		for row := 0; row < 25; row++ {
			mustExecTB(s.tb, s.h.Admin, ctx, `INSERT INTO idempotency_keys (tenant_id,actor_scope,idem_key,request_hash,status,response_status,response_body,expires_at) VALUES ($1,'seed',$2,'seed-hash','completed',200,'{}',now()+interval '24 hours')`, tf.tenant, fmt.Sprintf("history-%02d", row))
		}
		s.tenants[i] = tf
	}
}

func (s *requestSuite) assertRLS(ctx context.Context) error {
	if err := database.AssertRLSEnforced(ctx, s.pool); err != nil {
		return err
	}
	var current, databaseName string
	if err := s.pool.QueryRow(ctx, `SELECT current_user, current_database()`).Scan(&current, &databaseName); err != nil {
		return err
	}
	if current != "app_rt" {
		return fmt.Errorf("current_user=%q, want app_rt", current)
	}
	if databaseName != s.h.Name {
		return fmt.Errorf("current_database=%q, want fixture database %q", databaseName, s.h.Name)
	}
	return nil
}

func (s *requestSuite) queryPlanHash(profile string) string {
	tf := s.tenants[0]
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tf.tenant), tf.capacity)
	var query string
	var args []any
	switch profile {
	case "public":
		query = `SELECT 1`
	case "authenticated-read":
		query = `SELECT status FROM resources WHERE id=$1 AND resource_type=$2`
		args = []any{tf.resource.ID, tf.resource.Type}
	case "authenticated-write":
		query = `UPDATE resources SET status=status WHERE id=$1 AND resource_type=$2`
		args = []any{tf.resource.ID, tf.resource.Type}
	case "resource-authz":
		query = `SELECT role_id,scope_kind,scope_id FROM actor_assignments WHERE capacity_id=$1 AND valid_from<=now() AND (valid_to IS NULL OR valid_to>now())`
		args = []any{tf.capacity}
	case "idempotent-write":
		query = `SELECT status,response_status,response_body FROM idempotency_keys WHERE actor_scope=$1 AND idem_key=$2`
		args = []any{"seed", "history-00"}
	case "async-enqueue":
		query = `SELECT id FROM events_outbox WHERE event_type=$1 ORDER BY id LIMIT 1`
		args = []any{"perf.request.created"}
	default:
		s.tb.Fatalf("query plan requested for unknown profile %q", profile)
	}
	var raw []byte
	err := s.txm.WithTenant(ctx, func(ctx context.Context, db database.TenantDB) error {
		return db.QueryRow(ctx, `EXPLAIN (FORMAT JSON) `+query, args...).Scan(&raw)
	})
	if err != nil {
		s.tb.Fatalf("explain %s representative query: %v", profile, err)
	}
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func (s *requestSuite) handler(cache string) http.Handler {
	eval := s.coldEval
	if cache == "warm" {
		eval = s.warmEval
	}
	timed := timedEvaluator{inner: eval, metrics: &s.txm.metrics}
	r := httpx.NewRouter()
	r.Handle(http.MethodGet, "/perf/public", httpx.RouteMeta{Public: true}, s.publicHandler)
	r.Handle(http.MethodGet, "/perf/read", httpx.RouteMeta{Permission: readPermission}, s.readHandler)
	r.Handle(http.MethodPost, "/perf/write", httpx.RouteMeta{Permission: createPermission, NoRequestBody: true}, s.writeHandler)
	r.Handle(http.MethodGet, "/perf/resources/{id}", httpx.RouteMeta{Permission: readPermission}, s.resourceHandler(timed))
	r.Handle(http.MethodPost, "/perf/idempotent", httpx.RouteMeta{Permission: createPermission, NoRequestBody: true}, s.idempotentHandler)
	r.Handle(http.MethodPost, "/perf/enqueue", httpx.RouteMeta{Permission: createPermission, NoRequestBody: true}, s.enqueueHandler)
	if err := r.Err(); err != nil {
		s.tb.Fatalf("request benchmark routes: %v", err)
	}
	mux := r.SecureHandler(fixtureAuthenticator{tenants: s.tenants}, timed, s.txm)
	return httpx.Chain(mux, httpx.SecureHeaders(), httpx.BodyLimit(1<<20), httpx.Timeout(30*time.Second))
}

func (s *requestSuite) publicHandler(w http.ResponseWriter, r *http.Request) {
	s.measureHandler(func() error { var one int; return s.pool.QueryRow(r.Context(), `SELECT 1`).Scan(&one) }, w, r)
}

func (s *requestSuite) readHandler(w http.ResponseWriter, r *http.Request) {
	s.tenantHandler(w, r, true, func(ctx context.Context, db database.TenantDB, _ tenantFixture) error {
		var status string
		return db.QueryRow(ctx, `SELECT status FROM idempotency_keys WHERE actor_scope='seed' AND idem_key='history-00'`).Scan(&status)
	})
}

func (s *requestSuite) writeHandler(w http.ResponseWriter, r *http.Request) {
	key := fmt.Sprintf("write-%d", s.seq.Add(1))
	s.tenantHandler(w, r, false, func(ctx context.Context, db database.TenantDB, _ tenantFixture) error {
		_, err := db.Exec(ctx, `INSERT INTO idempotency_keys (tenant_id,actor_scope,idem_key,request_hash,status,response_status,response_body,expires_at) VALUES (app_tenant_id(),'write',$1,'hash','completed',201,'{}',now()+interval '1 hour')`, key)
		return err
	})
}

func (s *requestSuite) resourceHandler(eval authz.Evaluator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.tenantHandler(w, r, true, func(ctx context.Context, db database.TenantDB, tf tenantFixture) error {
			decision, err := eval.Evaluate(ctx, db, authz.Actor{Kind: authz.ActorUser, UserID: tf.user, CapacityID: tf.capacity, TenantID: tf.tenant}, readPermission, authz.Target{Scope: authz.ScopeResource, Resource: tf.resource})
			if err != nil {
				return err
			}
			if !decision.Allowed {
				return errors.New("resource authorization denied")
			}
			var status string
			return db.QueryRow(ctx, `SELECT status FROM resources WHERE id=$1 AND resource_type='perf.request'`, tf.resource.ID).Scan(&status)
		})
	}
}

func (s *requestSuite) idempotentHandler(w http.ResponseWriter, r *http.Request) {
	key := fmt.Sprintf("idem-%d", s.seq.Add(1))
	store := database.NewIdemStore()
	s.tenantHandler(w, r, false, func(ctx context.Context, db database.TenantDB, _ tenantFixture) error {
		replay, err := store.Begin(ctx, db, "benchmark", key, "request-hash", time.Hour)
		if err != nil {
			return err
		}
		if !replay.Fresh {
			return errors.New("idempotency claim was not fresh")
		}
		return store.Complete(ctx, db, "benchmark", key, http.StatusCreated, []byte(`{"ok":true}`))
	})
}

func (s *requestSuite) enqueueHandler(w http.ResponseWriter, r *http.Request) {
	writer := outbox.NewWriter(model.UUIDv7())
	s.tenantHandler(w, r, false, func(ctx context.Context, db database.TenantDB, tf tenantFixture) error {
		return writer.Write(ctx, db, outbox.Event{Type: "perf.request.created", Resource: tf.resource, Payload: map[string]any{"sequence": s.seq.Add(1)}})
	})
}

func (s *requestSuite) tenantHandler(w http.ResponseWriter, r *http.Request, readOnly bool, fn func(context.Context, database.TenantDB, tenantFixture) error) {
	i, _ := strconv.Atoi(r.Header.Get("X-Perf-Tenant"))
	tf := s.tenants[i]
	call := func(ctx context.Context, db database.TenantDB) error {
		start := time.Now()
		err := fn(ctx, db, tf)
		s.txm.metrics.add(&s.txm.metrics.handlerQuery, time.Since(start))
		return err
	}
	var err error
	if readOnly {
		err = s.txm.WithTenantRO(r.Context(), call)
	} else {
		err = s.txm.WithTenant(r.Context(), call)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeJSON(w)
}

func (s *requestSuite) measureHandler(fn func() error, w http.ResponseWriter, _ *http.Request) {
	start := time.Now()
	err := fn()
	s.txm.metrics.add(&s.txm.metrics.handlerQuery, time.Since(start))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	s.writeJSON(w)
}

func (s *requestSuite) writeJSON(w http.ResponseWriter) {
	start := time.Now()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "profile": "request-benchmark", "seed": seedValue})
	s.txm.metrics.add(&s.txm.metrics.serialization, time.Since(start))
}

func (s *requestSuite) runBatch(ctx context.Context, profile, cache string, tenants int) (batchResult, error) {
	if !contains(workloadProfiles, profile) || !contains(cacheStates, cache) || tenants < 1 || tenants > len(s.tenants) {
		return batchResult{}, fmt.Errorf("invalid matrix selection")
	}
	h := s.handler(cache)
	if cache == "warm" {
		for i := range tenants {
			if _, err := s.oneRequest(ctx, h, profile, i); err != nil {
				return batchResult{}, fmt.Errorf("warm tenant %d: %w", i, err)
			}
		}
	}
	if err := s.tracer.ForceFlush(context.Background()); err != nil {
		return batchResult{}, err
	}
	s.txm.metrics.reset()
	s.exporter.Reset()
	beforeAcquire := s.pool.Stat().AcquireDuration()
	lockCtx, cancelLock := context.WithCancel(ctx)
	lockDone := make(chan time.Duration, 1)
	go s.sampleLockWait(lockCtx, lockDone)
	latencies := make([]time.Duration, tenants)
	var totalBytes atomic.Int64
	var failures atomic.Int64
	startBatch := time.Now()
	var wg sync.WaitGroup
	wg.Add(tenants)
	for i := range tenants {
		go func() {
			defer wg.Done()
			start := time.Now()
			n, err := s.oneRequest(ctx, h, profile, i)
			latencies[i] = time.Since(start)
			totalBytes.Add(n)
			if err != nil {
				failures.Add(1)
			}
		}()
	}
	wg.Wait()
	batchDuration := time.Since(startBatch)
	cancelLock()
	lockWait := <-lockDone
	if err := s.tracer.ForceFlush(context.Background()); err != nil {
		return batchResult{}, err
	}
	spans := s.exporter.GetSpans()
	sqlCount := 0
	for _, span := range spans {
		if strings.HasPrefix(span.Name, "db.") {
			sqlCount++
		}
	}
	poolWait := s.pool.Stat().AcquireDuration() - beforeAcquire
	m := s.txm.metrics.snapshot()
	authz := m.authzQuery
	handler := m.handlerQuery
	serialization := m.serialization
	var requestDuration time.Duration
	for _, latency := range latencies {
		requestDuration += latency
	}
	middleware := requestDuration - poolWait - m.txSetup - authz - handler - serialization
	if middleware < 0 {
		middleware = 0
	}
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	return batchResult{Profile: profile, Cache: cache, ConcurrentTenants: tenants, Requests: tenants, Errors: int(failures.Load()), P50: percentile(latencies, .50), P95: percentile(latencies, .95), P99: percentile(latencies, .99), SQLStatements: sqlCount, Bytes: totalBytes.Load(), Elapsed: batchDuration, PoolWait: poolWait, TxDuration: m.txDuration, LockWait: lockWait, PlanHash: s.planHashes[profile], Cost: map[string]time.Duration{"pool_wait": poolWait, "tx_setup": m.txSetup, "authz_query": authz, "handler_query": handler, "serialization": serialization, "middleware": middleware}}, nil
}

func (s *requestSuite) oneRequest(ctx context.Context, h http.Handler, profile string, tenant int) (int64, error) {
	method, path := http.MethodGet, "/perf/"+strings.TrimPrefix(profile, "authenticated-")
	if profile == "public" {
		path = "/perf/public"
	}
	if profile == "authenticated-read" {
		path = "/perf/read"
	}
	if profile == "authenticated-write" {
		method, path = http.MethodPost, "/perf/write"
	}
	if profile == "resource-authz" {
		path = "/perf/resources/" + s.tenants[tenant].resource.ID.String()
	}
	if profile == "idempotent-write" {
		method, path = http.MethodPost, "/perf/idempotent"
	}
	if profile == "async-enqueue" {
		method, path = http.MethodPost, "/perf/enqueue"
	}
	body := bytes.NewReader(nil)
	requestBytes := int64(0)
	req := httptest.NewRequestWithContext(ctx, method, path, body)
	req.Header.Set("X-Perf-Tenant", strconv.Itoa(tenant))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code < 200 || rec.Code >= 300 {
		return requestBytes + int64(rec.Body.Len()), fmt.Errorf("%s %s returned %d: %s", method, path, rec.Code, strings.TrimSpace(rec.Body.String()))
	}
	return requestBytes + int64(rec.Body.Len()), nil
}

func (s *requestSuite) sampleLockWait(ctx context.Context, done chan<- time.Duration) {
	ticker := time.NewTicker(200 * time.Microsecond)
	defer ticker.Stop()
	var observed time.Duration
	for {
		select {
		case <-ctx.Done():
			done <- observed
			return
		case <-ticker.C:
			start := time.Now()
			var waiting int
			err := s.h.Admin.QueryRow(context.Background(), `SELECT count(*) FROM pg_stat_activity WHERE datname=$1 AND wait_event_type='Lock'`, s.h.Name).Scan(&waiting)
			if err == nil && waiting > 0 {
				observed += time.Since(start)
			}
		}
	}
}

func configuredWarmupDuration() (time.Duration, error) {
	raw := os.Getenv("PERF_WARMUP_DURATION")
	if raw == "" {
		return 0, nil
	}
	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("PERF_WARMUP_DURATION: %w", err)
	}
	if duration < 0 {
		return 0, fmt.Errorf("PERF_WARMUP_DURATION must not be negative")
	}
	return duration, nil
}

func (s *requestSuite) warmForDuration(ctx context.Context, h http.Handler, profile string, tenants int, duration time.Duration) error {
	deadline := time.Now().Add(duration)
	for request := 0; time.Now().Before(deadline); request++ {
		if _, err := s.oneRequest(ctx, h, profile, request%tenants); err != nil {
			return fmt.Errorf("warmup request %d: %w", request, err)
		}
	}
	return nil
}

func BenchmarkRealPostgresRequests(b *testing.B) {
	s := newRequestSuite(b)
	if err := s.assertRLS(context.Background()); err != nil {
		b.Fatalf("RLS precondition: %v", err)
	}
	warmupDuration, err := configuredWarmupDuration()
	if err != nil {
		b.Fatal(err)
	}
	published := make([]publishedResult, 0, len(workloadProfiles)*len(cacheStates)*len(concurrentTenantCounts))
	for _, profile := range workloadProfiles {
		if err := s.warmForDuration(context.Background(), s.handler("warm"), profile, len(s.tenants), warmupDuration); err != nil {
			b.Fatalf("timed profile warmup: %v", err)
		}
		for _, cache := range cacheStates {
			for _, tenants := range concurrentTenantCounts {
				name := profile + "/" + cache + "/tenants-" + itoa(tenants)
				b.Run(name, func(b *testing.B) {
					var allocErr error
					allocs := testing.AllocsPerRun(1, func() {
						_, allocErr = s.runBatch(context.Background(), profile, cache, tenants)
					}) / float64(tenants)
					if allocErr != nil {
						b.Fatalf("allocation sample: %v", allocErr)
					}
					b.ReportAllocs()
					var aggregate batchResult
					b.ResetTimer()
					for range b.N {
						result, err := s.runBatch(context.Background(), profile, cache, tenants)
						if err != nil {
							b.Fatal(err)
						}
						aggregate = mergeResult(aggregate, result)
					}
					b.StopTimer()
					requests := float64(b.N * tenants)
					if requests == 0 {
						return
					}
					p50 := aggregate.P50.Nanoseconds() / int64(b.N)
					p95 := aggregate.P95.Nanoseconds() / int64(b.N)
					p99 := aggregate.P99.Nanoseconds() / int64(b.N)
					sqlPerRequest := float64(aggregate.SQLStatements) / requests
					bytesPerRequest := float64(aggregate.Bytes) / requests
					poolWait := aggregate.PoolWait.Nanoseconds() / int64(requests)
					txDuration := aggregate.TxDuration.Nanoseconds() / int64(requests)
					lockWait := aggregate.LockWait.Nanoseconds() / int64(requests)
					cost := make(map[string]int64, len(costComponents))
					for _, component := range costComponents {
						cost[component] = aggregate.Cost[component].Nanoseconds() / int64(requests)
					}
					b.ReportMetric(float64(p50), "p50-ns/request")
					b.ReportMetric(float64(p95), "p95-ns/request")
					b.ReportMetric(float64(p99), "p99-ns/request")
					b.ReportMetric(sqlPerRequest, "sql/request")
					b.ReportMetric(bytesPerRequest, "bytes/request")
					b.ReportMetric(float64(poolWait), "poolwait-ns/request")
					b.ReportMetric(float64(txDuration), "tx-ns/request")
					b.ReportMetric(float64(lockWait), "lockwait-ns/request")
					for _, component := range costComponents {
						b.ReportMetric(float64(cost[component]), component+"-ns/request")
					}
					published = append(published, publishedResult{
						Name: name, Profile: profile, Cache: cache, ConcurrentTenants: tenants,
						P50NS: p50, P95NS: p95, P99NS: p99, AllocationsPerRequest: allocs,
						SQLStatementsPerRequest: sqlPerRequest, BytesPerRequest: bytesPerRequest,
						PoolWaitNSPerRequest: poolWait, TransactionNSPerRequest: txDuration,
						LockWaitNSPerRequest: lockWait, PlanHash: "sha256:" + s.planHashes[profile],
						CostNSPerRequest:    cost,
						RelativeToReference: map[string]float64{"p50": 1, "p95": 1, "p99": 1, "allocations": 1, "sql_statements": 1},
					})
				})
			}
		}
	}
	if reportPath := os.Getenv("PERF_REPORT"); reportPath != "" {
		pub := publication{
			SchemaVersion: 1, Reference: "perf/reference-schema1.json",
			ComparisonKind:    "initial-reference-capture",
			AbsoluteSLOStatus: "conditional-on-DEC-Q9",
			SourceRevision:    os.Getenv("PERF_SOURCE_SHA"), MeasuredAt: time.Now().UTC(),
			Environment: map[string]string{
				"goos": runtime.GOOS, "goarch": runtime.GOARCH, "go_version": runtime.Version(),
				"container_image":  os.Getenv("PERF_CONTAINER_IMAGE"),
				"postgres_image":   os.Getenv("PERF_POSTGRES_IMAGE"),
				"postgres_version": os.Getenv("PERF_POSTGRES_VERSION"),
				"postgres_config":  os.Getenv("PERF_POSTGRES_CONFIG"),
				"network":          os.Getenv("PERF_NETWORK"),
			},
			AttributionMethod: "pgx query spans plus non-overlapping wall-clock phase timers and pgxpool acquire-duration deltas",
			LockWaitMethod:    "200us pg_stat_activity wait_event_type=Lock sampler during each batch",
			Results:           published,
		}
		if err := writePublication(reportPath, pub); err != nil {
			b.Fatalf("write publication: %v", err)
		}
	}
}

func validatePublication(pub publication) error {
	if pub.SchemaVersion != 1 || pub.Reference != "perf/reference-schema1.json" {
		return fmt.Errorf("unexpected publication reference contract")
	}
	if pub.AbsoluteSLOStatus != "conditional-on-DEC-Q9" {
		return fmt.Errorf("absolute SLO status must remain conditional on DEC-Q9")
	}
	if pub.Environment["goos"] != "linux" || pub.Environment["goarch"] != "amd64" {
		return fmt.Errorf("publication environment must be linux/amd64")
	}
	for _, field := range []string{"go_version", "container_image", "postgres_image", "postgres_version", "postgres_config", "network"} {
		if pub.Environment[field] == "" {
			return fmt.Errorf("publication environment lacks %q", field)
		}
	}
	expected := len(workloadProfiles) * len(cacheStates) * len(concurrentTenantCounts)
	if len(pub.Results) != expected {
		return fmt.Errorf("publication has %d results, want %d", len(pub.Results), expected)
	}
	seen := make(map[string]struct{}, expected)
	for _, result := range pub.Results {
		if !contains(workloadProfiles, result.Profile) || !contains(cacheStates, result.Cache) {
			return fmt.Errorf("invalid result dimensions for %q", result.Name)
		}
		if result.ConcurrentTenants != 1 && result.ConcurrentTenants != 10 && result.ConcurrentTenants != 100 {
			return fmt.Errorf("invalid tenant concurrency for %q", result.Name)
		}
		if _, ok := seen[result.Name]; ok {
			return fmt.Errorf("duplicate result %q", result.Name)
		}
		seen[result.Name] = struct{}{}
		if result.P50NS <= 0 || result.P95NS <= 0 || result.P99NS <= 0 ||
			result.AllocationsPerRequest <= 0 || result.SQLStatementsPerRequest <= 0 ||
			result.BytesPerRequest <= 0 || result.PlanHash == "" {
			return fmt.Errorf("result %q has incomplete measurements", result.Name)
		}
		for _, component := range costComponents {
			if _, ok := result.CostNSPerRequest[component]; !ok {
				return fmt.Errorf("result %q lacks cost component %q", result.Name, component)
			}
		}
		for _, metric := range []string{"p50", "p95", "p99", "allocations", "sql_statements"} {
			if result.RelativeToReference[metric] != 1 {
				return fmt.Errorf("initial reference result %q has %s ratio %v, want 1", result.Name, metric, result.RelativeToReference[metric])
			}
		}
	}
	return nil
}

func publicationPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, path), nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("repository root containing go.mod not found")
		}
		dir = parent
	}
}

func writePublication(path string, pub publication) error {
	if err := validatePublication(pub); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(pub, "", "  ")
	if err != nil {
		return err
	}
	path, err = publicationPath(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, append(raw, '\n'), 0o644)
}

func mergeResult(a, b batchResult) batchResult {
	if a.Cost == nil {
		a.Cost = map[string]time.Duration{}
	}
	a.P50 += b.P50
	a.P95 += b.P95
	a.P99 += b.P99
	a.SQLStatements += b.SQLStatements
	a.Bytes += b.Bytes
	a.PoolWait += b.PoolWait
	a.TxDuration += b.TxDuration
	a.LockWait += b.LockWait
	for k, v := range b.Cost {
		a.Cost[k] += v
	}
	return a
}

func percentile(sorted []time.Duration, q float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	i := int(float64(len(sorted)-1)*q + .5)
	if i >= len(sorted) {
		i = len(sorted) - 1
	}
	return sorted[i]
}

func fixedUUID(kind string, i int) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(fmt.Sprintf("wowapi-perf-%d-%s-%d", seedValue, kind, i)))
}

func dsnForDatabase(dsn, databaseName string) (string, error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return "", err
	}
	if u.Scheme != "postgres" && u.Scheme != "postgresql" {
		return "", fmt.Errorf("benchmark DSN must be a postgres URL")
	}
	u.Path = "/" + databaseName
	return u.String(), nil
}

func mustExecTB(tb testing.TB, pool *pgxpool.Pool, ctx context.Context, sql string, args ...any) {
	tb.Helper()
	if _, err := pool.Exec(ctx, sql, args...); err != nil {
		tb.Fatalf("seed benchmark fixture: %v\n%s", err, sql)
	}
}

func contains(items []string, v string) bool {
	for _, item := range items {
		if item == v {
			return true
		}
	}
	return false
}
func itoa(v int) string { return strconv.Itoa(v) }
