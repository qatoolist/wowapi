package requestbench

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestReferenceV1FullFieldContract(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", "reference-v1.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}

	required := []string{
		"schema_version", "policy_status", "decision_dependency",
		"environment.os", "environment.arch", "environment.runner.name",
		"environment.runner.image", "environment.runner.image_digest",
		"environment.cpu.model_source", "environment.cpu.logical_cores",
		"environment.go.version", "environment.go.container_image", "environment.go.clean_cache_policy",
		"environment.postgres.image", "environment.postgres.image_digest",
		"environment.postgres.version", "environment.postgres.config.shared_buffers",
		"environment.postgres.config.max_connections", "environment.postgres.config.fsync",
		"environment.postgres.config.synchronous_commit", "environment.postgres.config.full_page_writes",
		"environment.postgres.config.jit", "environment.postgres.config.track_io_timing",
		"environment.pool.max_connections", "environment.pool.min_connections",
		"environment.pool.query_timeout", "environment.network.mode", "environment.network.shaping",
		"dataset.cardinality.tenants", "dataset.cardinality.users_per_tenant",
		"dataset.cardinality.capacities_per_tenant", "dataset.cardinality.organizations_per_tenant",
		"dataset.cardinality.resources_per_tenant", "dataset.cardinality.historical_rows_per_tenant",
		"dataset.tenant_distribution", "dataset.object_sizes_bytes", "dataset.fixture_checksum_algorithm",
		"provider_latency_model.mode", "provider_latency_model.authentication",
		"provider_latency_model.postgres", "provider_latency_model.object_storage",
		"provider_latency_model.queue_provider", "workload.seed", "workload.fixture",
		"workload.profiles", "workload.cache_states", "workload.concurrent_tenants",
		"procedure.warm_up_duration", "procedure.measurement_duration",
		"procedure.repetitions", "procedure.selected_run", "procedure.microbenchmark_count",
		"procedure.benchstat_alpha", "procedure.flaky_rerun_policy",
		"relative_ceilings.microbenchmark_ns_per_op_regression_percent",
		"relative_ceilings.allocation_increase_per_op", "relative_ceilings.request_p95_regression_percent",
		"relative_ceilings.request_p99_regression_percent", "relative_ceilings.sql_statement_count_increase",
		"relative_ceilings.error_rate_percent", "absolute_ceilings.status", "absolute_ceilings.dependency",
		"absolute_ceilings.request_p50", "absolute_ceilings.request_p95", "absolute_ceilings.request_p99",
		"absolute_ceilings.allocations_per_request", "absolute_ceilings.sql_statements_per_request",
		"absolute_ceilings.pool_wait_per_request", "absolute_ceilings.transaction_duration_per_request",
		"absolute_ceilings.lock_wait_per_request", "absolute_ceilings.memory_bytes",
		"publication.raw_output", "publication.plan_hashes", "publication.profiles",
		"publication.source_sha_addressed", "publication.comparison_mode",
	}
	for _, field := range required {
		if v, ok := lookup(doc, field); !ok || empty(v) {
			t.Errorf("required reference field %q is missing or empty", field)
		}
	}
	if got, _ := lookup(doc, "environment.os"); got != "linux" {
		t.Errorf("environment.os = %v, want linux", got)
	}
	if got, _ := lookup(doc, "environment.arch"); got != "amd64" {
		t.Errorf("environment.arch = %v, want amd64", got)
	}
	if got, _ := lookup(doc, "absolute_ceilings.status"); got != "conditional-on-DEC-Q9" {
		t.Errorf("absolute_ceilings.status = %v, want conditional-on-DEC-Q9", got)
	}
}

func TestWorkloadFixtureDefinesCompleteMatrix(t *testing.T) {
	t.Parallel()
	path := filepath.Join("..", "fixtures", "request-workloads-v1.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var fixture struct {
		Profiles          []string `json:"profiles"`
		Caches            []string `json:"cache_states"`
		ConcurrentTenants []int    `json:"concurrent_tenants"`
		Seed              struct {
			Tenants int `json:"tenants"`
		} `json:"seed"`
	}
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	wantProfiles := []string{"public", "authenticated-read", "authenticated-write", "resource-authz", "idempotent-write", "async-enqueue"}
	if !sameStrings(fixture.Profiles, wantProfiles) {
		t.Fatalf("profiles = %v, want %v", fixture.Profiles, wantProfiles)
	}
	if !sameStrings(fixture.Caches, []string{"cold", "warm"}) {
		t.Fatalf("cache_states = %v, want [cold warm]", fixture.Caches)
	}
	if len(fixture.ConcurrentTenants) != 3 || fixture.ConcurrentTenants[0] != 1 || fixture.ConcurrentTenants[1] != 10 || fixture.ConcurrentTenants[2] != 100 {
		t.Fatalf("concurrent_tenants = %v, want [1 10 100]", fixture.ConcurrentTenants)
	}
	if fixture.Seed.Tenants < 100 {
		t.Fatalf("seed.tenants = %d, want at least 100", fixture.Seed.Tenants)
	}
}

func TestReferenceContractIsHostIndependent(t *testing.T) {
	if runtime.GOOS == "linux" && runtime.GOARCH == "amd64" {
		return
	}
	// The committed policy describes the provisional reference runner, not the
	// developer host. This test intentionally still validates it off-runner.
	TestReferenceV1FullFieldContract(t)
}

func lookup(root map[string]any, dotted string) (any, bool) {
	current := any(root)
	start := 0
	for i := 0; i <= len(dotted); i++ {
		if i != len(dotted) && dotted[i] != '.' {
			continue
		}
		part := dotted[start:i]
		m, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = m[part]
		if !ok {
			return nil, false
		}
		start = i + 1
	}
	return current, true
}

func empty(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return x == ""
	case []any:
		return len(x) == 0
	case map[string]any:
		return len(x) == 0
	default:
		return false
	}
}

func sameStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
