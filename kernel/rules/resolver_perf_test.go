package rules_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/config"
	"github.com/qatoolist/wowapi/v2/kernel/database"
	"github.com/qatoolist/wowapi/v2/kernel/model"
	"github.com/qatoolist/wowapi/v2/kernel/rules"
	"github.com/qatoolist/wowapi/v2/testkit"
)

type queryCounter struct {
	count atomic.Int64
}

func (q *queryCounter) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	q.count.Add(1)
	return ctx
}

func (*queryCounter) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

type captureQueryTracer struct {
	mu   sync.Mutex
	sql  string
	args []any
}

func (q *captureQueryTracer) TraceQueryStart(ctx context.Context, _ *pgx.Conn, data pgx.TraceQueryStartData) context.Context {
	if strings.Contains(data.SQL, "FROM rule_versions") &&
		!strings.HasPrefix(strings.TrimSpace(data.SQL), "EXPLAIN") {
		q.mu.Lock()
		q.sql = data.SQL
		q.args = append(q.args[:0], data.Args...)
		q.mu.Unlock()
	}
	return ctx
}

func (*captureQueryTracer) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func (q *captureQueryTracer) snapshot() (string, []any) {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.sql, append([]any(nil), q.args...)
}

func tracedTxManager(t *testing.T, h *testkit.DBHandle, tracer pgx.QueryTracer) database.TxManager {
	t.Helper()
	cfg := h.Runtime.Config()
	cfg.ConnConfig.Tracer = tracer
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	t.Logf("pool created: %v, err: %v", pool != nil, err)
	if err != nil {
		t.Fatalf("create traced runtime pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return database.NewManager(pool, config.Defaults().DB, database.WithRLSGuard())
}

func countedTxManager(t *testing.T, h *testkit.DBHandle, counter *queryCounter) database.TxManager {
	t.Helper()
	return tracedTxManager(t, h, counter)
}

func TestIntegrationResolverQueryCountConstantWithDepth(t *testing.T) {
	const key = "core.retention.audit_days"
	counts := make(map[int]int64, 3)
	legacyCounts := make(map[int]int64, 3)
	ran := false

	for _, depth := range []int{3, 10, 50} {
		t.Run(fmt.Sprintf("depth_%d", depth), func(t *testing.T) {
			h := testkit.NewDB(t)
			ran = true
			seedRuleDef(t, h, key)
			registry := reg(t, false)
			store := rules.NewStore(registry, model.UUIDv7())
			resolver := rules.NewResolver(registry, authz.NewStore().OrgAncestors)
			tenant := testkit.CreateTenant(t, h)
			ctx := testkit.TenantCtx(tenant.ID)

			var parent *uuid.UUID
			var leaf uuid.UUID
			for i := range depth {
				org := testkit.CreateOrg(t, h, tenant.ID, parent, fmt.Sprintf("depth-%d", i))
				leaf = org
				parent = &org
			}
			proposeActivate(t, h, ctx, store, rules.Proposal{
				Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`17`),
			})

			at := time.Now()
			counter := &queryCounter{}
			txm := countedTxManager(t, h, counter)
			counter.count.Store(0)
			var resolved rules.Resolved
			if err := txm.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
				var err error
				resolved, err = resolver.Resolve(ctx, db, key, leaf, at)
				return err
			}); err != nil {
				t.Fatalf("resolve at depth %d: %v", depth, err)
			}
			var got int
			if err := resolved.Decode(&got); err != nil {
				t.Fatalf("decode at depth %d: %v", depth, err)
			}
			if got != 17 || resolved.Scope != rules.ScopeTenant {
				t.Fatalf("depth %d resolved %d at %s; want tenant value 17", depth, got, resolved.Scope)
			}
			counts[depth] = counter.count.Load()

			legacyCounter := &queryCounter{}
			legacyTxM := countedTxManager(t, h, legacyCounter)
			legacyCounter.count.Store(0)
			if err := legacyTxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
				ancestors, err := authz.NewStore().OrgAncestors(ctx, db, leaf)
				if err != nil {
					return err
				}
				_, err = legacyResolveRule(ctx, db, key, ancestors, at, json.RawMessage(`30`))
				return err
			}); err != nil {
				t.Fatalf("legacy resolve at depth %d: %v", depth, err)
			}
			legacyCounts[depth] = legacyCounter.count.Load()
		})
	}

	if !ran {
		t.Skip("query-count comparison requires a database")
	}

	t.Logf("observed SQL statements: legacy depth 3=%d, 10=%d, 50=%d; set-based depth 3=%d, 10=%d, 50=%d",
		legacyCounts[3], legacyCounts[10], legacyCounts[50], counts[3], counts[10], counts[50])
	if counts[3] != counts[10] || counts[10] != counts[50] {
		t.Fatalf("SQL count grew with ancestry depth: depth 3=%d, 10=%d, 50=%d", counts[3], counts[10], counts[50])
	}
	if legacyCounts[3] >= legacyCounts[10] || legacyCounts[10] >= legacyCounts[50] {
		t.Fatalf("legacy SQL count did not grow with depth: depth 3=%d, 10=%d, 50=%d",
			legacyCounts[3], legacyCounts[10], legacyCounts[50])
	}
}

func TestIntegrationResolverSetBasedParity(t *testing.T) {
	const key = "core.retention.audit_days"
	for _, scenario := range []string{
		"nearest_org",
		"ancestor_org",
		"tenant",
		"platform",
		"code_default",
		"historical_superseded",
	} {
		t.Run(scenario, func(t *testing.T) {
			h := testkit.NewDB(t)
			seedRuleDef(t, h, key)
			registry := reg(t, false)
			store := rules.NewStore(registry, model.UUIDv7())
			resolver := rules.NewResolver(registry, authz.NewStore().OrgAncestors)
			tenant := testkit.CreateTenant(t, h)
			ctx := testkit.TenantCtx(tenant.ID)
			ancestors := createRuleOrgChain(t, h, tenant.ID, 3)
			at := time.Now().UTC()

			switch scenario {
			case "nearest_org":
				proposeActivate(t, h, ctx, store, rules.Proposal{
					Key: key, Scope: rules.ScopePlatform, Value: json.RawMessage(`90`),
				})
				proposeActivate(t, h, ctx, store, rules.Proposal{
					Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`20`),
				})
				proposeActivate(t, h, ctx, store, rules.Proposal{
					Key: key, Scope: rules.ScopeOrg, ScopeID: ancestors[1], Value: json.RawMessage(`50`),
				})
				proposeActivate(t, h, ctx, store, rules.Proposal{
					Key: key, Scope: rules.ScopeOrg, ScopeID: ancestors[0], Value: json.RawMessage(`11`),
				})
			case "ancestor_org":
				proposeActivate(t, h, ctx, store, rules.Proposal{
					Key: key, Scope: rules.ScopePlatform, Value: json.RawMessage(`90`),
				})
				proposeActivate(t, h, ctx, store, rules.Proposal{
					Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`20`),
				})
				proposeActivate(t, h, ctx, store, rules.Proposal{
					Key: key, Scope: rules.ScopeOrg, ScopeID: ancestors[2], Value: json.RawMessage(`70`),
				})
			case "tenant":
				proposeActivate(t, h, ctx, store, rules.Proposal{
					Key: key, Scope: rules.ScopePlatform, Value: json.RawMessage(`90`),
				})
				proposeActivate(t, h, ctx, store, rules.Proposal{
					Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`20`),
				})
			case "platform":
				proposeActivate(t, h, ctx, store, rules.Proposal{
					Key: key, Scope: rules.ScopePlatform, Value: json.RawMessage(`90`),
				})
			case "historical_superseded":
				proposeActivate(t, h, ctx, store, rules.Proposal{
					Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`5`),
					EffectiveFrom: at.Add(-10 * 24 * time.Hour),
				})
				proposeActivate(t, h, ctx, store, rules.Proposal{
					Key: key, Scope: rules.ScopeTenant, Value: json.RawMessage(`9`),
					EffectiveFrom: at.Add(-2 * 24 * time.Hour),
				})
				at = at.Add(-5 * 24 * time.Hour)
			}

			var setBased, legacy rules.Resolved
			err := h.TxM.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
				var err error
				setBased, err = resolver.Resolve(ctx, db, key, ancestors[0], at)
				if err != nil {
					return err
				}
				legacy, err = legacyResolveRule(ctx, db, key, ancestors, at, json.RawMessage(`30`))
				return err
			})
			if err != nil {
				t.Fatalf("resolve %s: %v", scenario, err)
			}
			if setBased.Scope != legacy.Scope ||
				setBased.VersionID != legacy.VersionID ||
				setBased.IsDefault != legacy.IsDefault ||
				!bytes.Equal(setBased.Value, legacy.Value) {
				t.Fatalf("set-based result %+v does not match legacy result %+v", setBased, legacy)
			}
		})
	}
}

func createRuleOrgChain(t *testing.T, h *testkit.DBHandle, tenantID uuid.UUID, depth int) []uuid.UUID {
	t.Helper()
	rootFirst := make([]uuid.UUID, 0, depth)
	var parent *uuid.UUID
	for i := range depth {
		org := testkit.CreateOrg(t, h, tenantID, parent, fmt.Sprintf("parity-%d", i))
		rootFirst = append(rootFirst, org)
		parent = &org
	}
	selfFirst := make([]uuid.UUID, len(rootFirst))
	for i, id := range rootFirst {
		selfFirst[len(rootFirst)-1-i] = id
	}
	return selfFirst
}

func legacyResolveRule(
	ctx context.Context,
	db database.TenantDB,
	key string,
	ancestors []uuid.UUID,
	at time.Time,
	defaultValue json.RawMessage,
) (rules.Resolved, error) {
	for _, scopeID := range ancestors {
		resolved, found, err := legacyLookupRule(ctx, db, key, rules.ScopeOrg, scopeID, at)
		if err != nil || found {
			return resolved, err
		}
	}
	for _, scope := range []rules.ScopeKind{rules.ScopeTenant, rules.ScopePlatform} {
		resolved, found, err := legacyLookupRule(ctx, db, key, scope, uuid.Nil, at)
		if err != nil || found {
			return resolved, err
		}
	}
	return rules.Resolved{Key: key, Value: defaultValue, IsDefault: true}, nil
}

func legacyLookupRule(
	ctx context.Context,
	db database.TenantDB,
	key string,
	scope rules.ScopeKind,
	scopeID uuid.UUID,
	at time.Time,
) (rules.Resolved, bool, error) {
	var scopeArg any
	if scopeID != uuid.Nil {
		scopeArg = scopeID
	}
	var (
		id    uuid.UUID
		value []byte
	)
	err := db.QueryRow(ctx,
		`SELECT id, value FROM rule_versions
          WHERE rule_key = $1 AND scope_kind = $2
            AND (scope_id = $3 OR ($3 IS NULL AND scope_id IS NULL))
            AND status IN ('active','superseded')
            AND effective_from <= $4 AND (effective_to IS NULL OR effective_to > $4)
          ORDER BY effective_from DESC
          LIMIT 1`,
		key, string(scope), scopeArg, at).Scan(&id, &value)
	if errors.Is(err, pgx.ErrNoRows) {
		return rules.Resolved{}, false, nil
	}
	if err != nil {
		return rules.Resolved{}, false, err
	}
	return rules.Resolved{
		Key: key, Value: json.RawMessage(value), Scope: scope, VersionID: id,
	}, true, nil
}

func TestIntegrationResolverExplainFixtures(t *testing.T) {
	const backgroundKey = "core.perf.background"
	h := testkit.NewDB(t)
	seedDefFull(t, h, backgroundKey, `{"type":"integer"}`, `0`)

	registry := rules.NewRegistry()
	specs := []struct {
		name         string
		key          string
		depth        int
		historyCount int
	}{
		{name: "shallow-low", key: "core.perf.shallowlow", depth: 3, historyCount: 4},
		{name: "shallow-high", key: "core.perf.shallowhigh", depth: 3, historyCount: 1000},
		{name: "deep-low", key: "core.perf.deeplow", depth: 50, historyCount: 4},
		{name: "deep-high", key: "core.perf.deephigh", depth: 50, historyCount: 1000},
	}
	for _, spec := range specs {
		seedDefFull(t, h, spec.key, `{"type":"integer"}`, `0`)
		registry.Register("core", rules.Point{
			Key: spec.key, ValueSchema: json.RawMessage(`{"type":"integer"}`),
			Default: json.RawMessage(`0`), Description: "PERF-03 EXPLAIN fixture",
		})
	}
	if err := registry.Err(); err != nil {
		t.Fatalf("register fixture rule points: %v", err)
	}

	backgroundTenant := testkit.CreateTenant(t, h)
	seedRuleHistory(t, h, backgroundKey, backgroundTenant.ID, rules.ScopeTenant, uuid.Nil, 20_000,
		time.Now().UTC().Add(-30*24*time.Hour))

	for _, spec := range specs {
		t.Run(spec.name, func(t *testing.T) {
			tenant := testkit.CreateTenant(t, h)
			ctx := testkit.TenantCtx(tenant.ID)
			ancestors := createNamedRuleOrgChain(t, h, tenant.ID, spec.depth, spec.name)
			base := time.Now().UTC().Add(-time.Duration(spec.historyCount+2) * time.Minute)

			for _, scopeID := range ancestors {
				seedRuleHistory(t, h, spec.key, tenant.ID, rules.ScopeOrg, scopeID, spec.historyCount, base)
			}
			seedRuleHistory(t, h, spec.key, tenant.ID, rules.ScopeTenant, uuid.Nil, spec.historyCount, base)
			seedRuleHistory(t, h, spec.key, tenant.ID, rules.ScopePlatform, uuid.Nil, spec.historyCount, base)
			if _, err := h.Admin.Exec(context.Background(), `ANALYZE rule_versions`); err != nil {
				t.Fatalf("analyze rule_versions: %v", err)
			}

			historicalAt := base.Add(time.Duration(spec.historyCount/2)*time.Minute + 30*time.Second)
			currentAt := base.Add(time.Duration(spec.historyCount+1) * time.Minute)
			tracer := &captureQueryTracer{}
			txm := tracedTxManager(t, h, tracer)
			var historicalPlan, currentPlan []byte
			err := txm.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
				if _, err := rules.NewResolver(registry, authz.NewStore().OrgAncestors).
					Resolve(ctx, db, spec.key, ancestors[0], historicalAt); err != nil {
					return err
				}
				query, args := tracer.snapshot()
				if query == "" {
					return fmt.Errorf("resolver query was not captured")
				}
				if err := db.QueryRow(ctx,
					"EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON) "+query, args...).
					Scan(&historicalPlan); err != nil {
					return fmt.Errorf("historical EXPLAIN: %w", err)
				}
				return db.QueryRow(ctx,
					`EXPLAIN (ANALYZE, BUFFERS, FORMAT JSON)
                     SELECT id, value
                       FROM rule_versions
                      WHERE rule_key = $1
                        AND scope_kind = 'tenant'
                        AND scope_id IS NULL
                        AND tenant_id = $2
                        AND status = 'active'
                        AND effective_from <= $3
                        AND (effective_to IS NULL OR effective_to > $3)
                      ORDER BY effective_from DESC
                      LIMIT 1`,
					spec.key, tenant.ID, currentAt).Scan(&currentPlan)
			})
			if err != nil {
				t.Fatalf("build EXPLAIN fixture: %v", err)
			}

			assertRuleVersionsIndexPlan(t, historicalPlan, "rule_versions_history_resolution_idx")
			assertRuleVersionsIndexPlan(t, currentPlan, "")
			if os.Getenv("WOWAPI_UPDATE_EXPLAIN_FIXTURES") != "" {
				writeExplainFixture(t, spec.name, spec.depth, spec.historyCount, historicalPlan, currentPlan)
			}
		})
	}
}

func seedRuleHistory(
	t *testing.T,
	h *testkit.DBHandle,
	key string,
	tenantID uuid.UUID,
	scope rules.ScopeKind,
	scopeID uuid.UUID,
	count int,
	base time.Time,
) {
	t.Helper()
	rows := make([][]any, 0, count)
	for i := range count {
		var tenantArg, scopeArg, effectiveTo any
		if scope != rules.ScopePlatform {
			tenantArg = tenantID
		}
		if scopeID != uuid.Nil {
			scopeArg = scopeID
		}
		status := "superseded"
		if i == count-1 {
			status = "active"
		} else {
			effectiveTo = base.Add(time.Duration(i+1) * time.Minute)
		}
		rows = append(rows, []any{
			uuid.New(), key, tenantArg, string(scope), scopeArg,
			json.RawMessage(fmt.Sprintf("%d", i+1)),
			base.Add(time.Duration(i) * time.Minute), effectiveTo, status, uuid.Nil,
		})
	}
	_, err := h.Admin.CopyFrom(
		context.Background(),
		pgx.Identifier{"rule_versions"},
		[]string{
			"id", "rule_key", "tenant_id", "scope_kind", "scope_id",
			"value", "effective_from", "effective_to", "status", "created_by",
		},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		t.Fatalf("seed %s %s history: %v", key, scope, err)
	}
}

func assertRuleVersionsIndexPlan(t *testing.T, plan []byte, indexName string) {
	t.Helper()
	if indexName != "" && !strings.Contains(string(plan), `"`+indexName+`"`) {
		t.Fatalf("EXPLAIN did not use %s:\n%s", indexName, plan)
	}
	var decoded any
	if err := json.Unmarshal(plan, &decoded); err != nil {
		t.Fatalf("decode EXPLAIN JSON: %v", err)
	}
	var sawRuleVersions, sawIndexAccess bool
	var walk func(any)
	walk = func(value any) {
		switch value := value.(type) {
		case []any:
			for _, child := range value {
				walk(child)
			}
		case map[string]any:
			if value["Relation Name"] == "rule_versions" {
				sawRuleVersions = true
				if value["Node Type"] == "Seq Scan" {
					t.Errorf("rule_versions used a sequential scan: %v", value)
				}
				if strings.Contains(fmt.Sprint(value["Node Type"]), "Index") ||
					value["Node Type"] == "Bitmap Heap Scan" {
					sawIndexAccess = true
				}
			}
			for _, child := range value {
				walk(child)
			}
		}
	}
	walk(decoded)
	if !sawRuleVersions {
		t.Fatal("EXPLAIN did not contain a rule_versions plan node")
	}
	if !sawIndexAccess {
		t.Fatal("EXPLAIN did not use index access for rule_versions")
	}
}

func writeExplainFixture(
	t *testing.T,
	name string,
	depth int,
	historyCount int,
	historicalPlan []byte,
	currentPlan []byte,
) {
	t.Helper()
	fixture := struct {
		ArtifactID     string          `json:"artifact_id"`
		GeneratedAt    string          `json:"generated_at"`
		Environment    string          `json:"environment"`
		DECQ9          string          `json:"dec_q9"`
		AncestryDepth  int             `json:"ancestry_depth"`
		VersionsScope  int             `json:"versions_per_scope"`
		HistoricalPlan json.RawMessage `json:"historical_resolution_explain"`
		CurrentPlan    json.RawMessage `json:"current_predicate_explain"`
	}{
		ArtifactID: "ART-W07-E01-S002-004", GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Environment: "local PostgreSQL 16 container; relative evidence only",
		DECQ9:       "open; no absolute SLO claim", AncestryDepth: depth, VersionsScope: historyCount,
		HistoricalPlan: historicalPlan, CurrentPlan: currentPlan,
	}
	content, err := json.MarshalIndent(fixture, "", "  ")
	if err != nil {
		t.Fatalf("marshal EXPLAIN fixture: %v", err)
	}
	path := filepath.Join("..", "..", "perf", "results", "perf-03-explain-"+name+".json")
	if err := os.WriteFile(path, append(content, '\n'), 0o644); err != nil {
		t.Fatalf("write EXPLAIN fixture %s: %v", path, err)
	}
}

func createNamedRuleOrgChain(
	t *testing.T,
	h *testkit.DBHandle,
	tenantID uuid.UUID,
	depth int,
	namePrefix string,
) []uuid.UUID {
	t.Helper()
	rootFirst := make([]uuid.UUID, 0, depth)
	var parent *uuid.UUID
	for i := range depth {
		org := testkit.CreateOrg(t, h, tenantID, parent, fmt.Sprintf("%s-%d", namePrefix, i))
		rootFirst = append(rootFirst, org)
		parent = &org
	}
	selfFirst := make([]uuid.UUID, len(rootFirst))
	for i, id := range rootFirst {
		selfFirst[len(rootFirst)-1-i] = id
	}
	return selfFirst
}
