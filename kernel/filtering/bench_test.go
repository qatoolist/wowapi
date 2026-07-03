package filtering_test

// Hot-path benchmarks for the filtering and keyset clause builders (criterion #17).
//
// Filter parsing + SQL building happens on every list request. The allowlist
// construction is once per endpoint registration; Parse + SQL happen per request.

import (
	"testing"

	"github.com/qatoolist/wowapi/kernel/filtering"
	"github.com/qatoolist/wowapi/kernel/pagination"
)

var benchAllow = filtering.Allowlist{
	"status":     {Col: "status", Ops: []filtering.Op{filtering.OpEq, filtering.OpIn}},
	"created_at": {Col: "created_at", Ops: []filtering.Op{filtering.OpGte, filtering.OpLte}},
	"assignee":   {Col: "assignee_id", Ops: []filtering.Op{filtering.OpEq, filtering.OpIn}},
}

var benchSortAllow = filtering.SortAllowlist{
	"created_at": {Col: "created_at"},
	"id":         {Col: "id"},
}

// BenchmarkFilteringParseAndSQL measures the full per-request path: parse
// raw input against the allowlist, then render to SQL. One condition.
func BenchmarkFilteringParseAndSQL(b *testing.B) {
	raw := map[string][]string{"status": {"eq:active"}}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set, err := filtering.Parse(raw, benchAllow)
		if err != nil {
			b.Fatal(err)
		}
		_, _, _ = set.SQL(1)
	}
}

// BenchmarkFilteringParseMulti measures parsing + SQL with three conditions
// (representative of a real list endpoint with multi-field filters).
func BenchmarkFilteringParseMulti(b *testing.B) {
	raw := map[string][]string{
		"status":     {"in:active,pending"},
		"created_at": {"gte:2026-01-01"},
		"assignee":   {"eq:alice"},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		set, err := filtering.Parse(raw, benchAllow)
		if err != nil {
			b.Fatal(err)
		}
		_, _, _ = set.SQL(1)
	}
}

// BenchmarkFilteringSQLOnly measures SQL rendering on an already-parsed Set.
// This is the inner loop when the Set is cached across paginated requests.
func BenchmarkFilteringSQLOnly(b *testing.B) {
	raw := map[string][]string{"status": {"in:active,pending"}}
	set, err := filtering.Parse(raw, benchAllow)
	if err != nil {
		b.Fatalf("setup: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = set.SQL(1)
	}
}

// BenchmarkSortParseAndSQL measures sort parsing + ORDER BY rendering.
func BenchmarkSortParseAndSQL(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, err := filtering.ParseSort("created_at:desc,id:asc", benchSortAllow)
		if err != nil {
			b.Fatal(err)
		}
		_ = s.SQL()
	}
}

// BenchmarkKeysetClause measures the keyset predicate builder: called on
// every non-first page of a paginated list. Exercises itoa + string joins.
func BenchmarkKeysetClause(b *testing.B) {
	sort, err := filtering.ParseSort("created_at:desc,id:asc", benchSortAllow)
	if err != nil {
		b.Fatalf("setup sort: %v", err)
	}
	cur, err := pagination.EncodeCursor(map[string]any{
		"created_at": "2026-06-01T00:00:00Z",
		"id":         int64(42),
	})
	if err != nil {
		b.Fatalf("setup cursor: %v", err)
	}
	decoded, err := pagination.DecodeCursor(cur)
	if err != nil {
		b.Fatalf("decode cursor: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _, _ = filtering.KeysetClause(sort, decoded, 1)
	}
}
