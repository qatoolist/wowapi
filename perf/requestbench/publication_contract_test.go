package requestbench

import (
	"fmt"
	"testing"
	"time"
)

func TestPublicationRequiresCompleteMatrixAndAttribution(t *testing.T) {
	results := make([]publishedResult, 0, 36)
	for _, profile := range workloadProfiles {
		for _, cache := range cacheStates {
			for _, tenants := range concurrentTenantCounts {
				cost := map[string]int64{}
				for _, component := range costComponents {
					cost[component] = 1
				}
				results = append(results, publishedResult{
					Name:    fmt.Sprintf("%s/%s/tenants-%d", profile, cache, tenants),
					Profile: profile, Cache: cache, ConcurrentTenants: tenants,
					P50NS: 1, P95NS: 2, P99NS: 3, AllocationsPerRequest: 1,
					SQLStatementsPerRequest: 1, BytesPerRequest: 1, PoolWaitNSPerRequest: 1,
					TransactionNSPerRequest: 1, LockWaitNSPerRequest: 0,
					PlanHash: "sha256:fixture", CostNSPerRequest: cost,
					RelativeToReference: map[string]float64{
						"p50": 1, "p95": 1, "p99": 1, "allocations": 1, "sql_statements": 1,
					},
				})
			}
		}
	}
	pub := publication{
		SchemaVersion: 1, Reference: "perf/reference-v1.json",
		ComparisonKind: "initial-reference-capture", AbsoluteSLOStatus: "conditional-on-DEC-Q9",
		MeasuredAt: time.Now().UTC(),
		Environment: map[string]string{
			"goos": "linux", "goarch": "amd64", "go_version": "go1.26.5",
			"container_image": "sha256:go", "postgres_image": "sha256:postgres",
			"postgres_version": "16.9", "postgres_config": "pinned", "network": "container bridge",
		},
		Results: results,
	}
	if err := validatePublication(pub); err != nil {
		t.Fatalf("valid publication rejected: %v", err)
	}
	missingEnvironment := pub
	missingEnvironment.Environment = nil
	if err := validatePublication(missingEnvironment); err == nil {
		t.Fatal("publication without observed container environment accepted")
	}
	pub.Results = pub.Results[:35]
	if err := validatePublication(pub); err == nil {
		t.Fatal("incomplete 35-row matrix accepted")
	}
}
