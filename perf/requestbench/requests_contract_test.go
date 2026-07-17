package requestbench

import (
	"context"
	"testing"
	"time"

	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/database"
)

func TestRealPostgresRequestProfileMatrix(t *testing.T) {
	s := newRequestSuite(t)
	if err := s.assertRLS(context.Background()); err != nil {
		t.Fatalf("RLS precondition: %v", err)
	}
	for _, profile := range workloadProfiles {
		for _, cache := range cacheStates {
			for _, tenants := range concurrentTenantCounts {
				t.Run(profile+"/"+cache+"/tenants-"+itoa(tenants), func(t *testing.T) {
					result, err := s.runBatch(context.Background(), profile, cache, tenants)
					if err != nil {
						t.Fatalf("run batch: %v", err)
					}
					if result.Requests != tenants {
						t.Fatalf("requests = %d, want %d", result.Requests, tenants)
					}
					if result.Errors != 0 {
						t.Fatalf("errors = %d, want 0", result.Errors)
					}
					if result.SQLStatements == 0 {
						t.Fatal("SQLStatements = 0; workload did not exercise real PostgreSQL")
					}
					for _, component := range costComponents {
						if _, ok := result.Cost[component]; !ok {
							t.Errorf("missing cost component %q", component)
						}
					}
					if result.PlanHash == "" {
						t.Fatal("empty query plan hash")
					}
				})
			}
		}
	}
}

func TestSeedMatchesReferenceDatasetCardinality(t *testing.T) {
	s := newRequestSuite(t)
	tf := s.tenants[0]
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tf.tenant), tf.capacity)
	var resources, historical int
	err := s.txm.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		if err := db.QueryRow(ctx, `SELECT count(*) FROM resources`).Scan(&resources); err != nil {
			return err
		}
		return db.QueryRow(ctx, `SELECT count(*) FROM idempotency_keys WHERE actor_scope='seed'`).Scan(&historical)
	})
	if err != nil {
		t.Fatalf("count seed rows: %v", err)
	}
	if resources != 10 || historical != 25 {
		t.Fatalf("seed cardinality resources=%d historical=%d, want 10 and 25", resources, historical)
	}
}

func TestSeedGrantsRealPostgresAuthorization(t *testing.T) {
	s := newRequestSuite(t)
	tf := s.tenants[0]
	ctx := database.WithActorID(database.WithTenantID(context.Background(), tf.tenant), tf.capacity)
	var assignments []authz.Assignment
	var assignmentCount, roleCount, permissionCount int
	var boundTenant, boundActor, currentRole string
	err := s.txm.WithTenantRO(ctx, func(ctx context.Context, db database.TenantDB) error {
		if err := db.QueryRow(ctx, `SELECT current_setting('app.tenant_id'), current_setting('app.actor_id'), current_user`).Scan(&boundTenant, &boundActor, &currentRole); err != nil {
			return err
		}
		if err := db.QueryRow(ctx, `SELECT count(*) FROM actor_assignments WHERE capacity_id=$1`, tf.capacity).Scan(&assignmentCount); err != nil {
			return err
		}
		if err := db.QueryRow(ctx, `SELECT count(*) FROM roles WHERE tenant_id=$1`, tf.tenant).Scan(&roleCount); err != nil {
			return err
		}
		if err := db.QueryRow(ctx, `SELECT count(*) FROM role_permissions`).Scan(&permissionCount); err != nil {
			return err
		}
		var err error
		assignments, err = authz.NewStore().ActiveAssignments(ctx, db, authz.Actor{
			Kind: authz.ActorUser, UserID: tf.user, CapacityID: tf.capacity, TenantID: tf.tenant,
		}, time.Now())
		return err
	})
	if err != nil {
		t.Fatalf("load assignments: %v", err)
	}
	if len(assignments) != 1 {
		t.Fatalf("assignments = %#v, want one seeded assignment (bound tenant=%s actor=%s role=%s; visible actor_assignments=%d roles=%d role_permissions=%d)", assignments, boundTenant, boundActor, currentRole, assignmentCount, roleCount, permissionCount)
	}
}

func TestCostAttributionIsNonOverlapping(t *testing.T) {
	s := newRequestSuite(t)
	result, err := s.runBatch(context.Background(), "authenticated-read", "cold", 1)
	if err != nil {
		t.Fatalf("run batch: %v", err)
	}
	var attributed time.Duration
	for _, component := range costComponents {
		attributed += result.Cost[component]
	}
	if attributed > result.Elapsed {
		t.Fatalf("attributed cost %s exceeds elapsed request time %s", attributed, result.Elapsed)
	}
}

func TestSQLCountExcludesWarmupAndPriorBatches(t *testing.T) {
	s := newRequestSuite(t)
	// The first concurrent batch may create pool connections; their mandatory
	// SET ROLE and RLS guard statements correctly count as cold-path SQL.
	first, err := s.runBatch(context.Background(), "public", "warm", 10)
	if err != nil {
		t.Fatalf("prime concurrent pool: %v", err)
	}
	result, err := s.runBatch(context.Background(), "public", "warm", 10)
	if err != nil {
		t.Fatalf("run measured batch: %v", err)
	}
	// Each request issues SELECT 1. A newly opened runtime connection may also
	// issue SET ROLE and the RLS guard query; no other batch is permitted.
	if result.SQLStatements < result.Requests || result.SQLStatements > 3*result.Requests {
		t.Fatalf("SQL statements = %d, requests = %d; expected request SQL plus at most two connection-setup statements per request (first batch=%d)", result.SQLStatements, result.Requests, first.SQLStatements)
	}
}

func TestMatrixContractHasSixVariantsPerProfile(t *testing.T) {
	if got := len(cacheStates) * len(concurrentTenantCounts); got != 6 {
		t.Fatalf("variants per profile = %d, want 6", got)
	}
	if got := len(workloadProfiles); got != 6 {
		t.Fatalf("workload profiles = %d, want 6", got)
	}
	if concurrentTenantCounts[len(concurrentTenantCounts)-1] != 100 {
		t.Fatalf("maximum concurrent tenants = %d, want 100", concurrentTenantCounts[len(concurrentTenantCounts)-1])
	}
}
