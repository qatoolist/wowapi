package filtering_test

import (
	"strings"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/filtering"
)

// The filter DSL takes untrusted client input (query params) and turns it into a
// SQL WHERE clause — the highest-value injection target in the framework (roadmap
// S8). These fuzz tests assert the security invariant under adversarial input:
// client text NEVER reaches the SQL as a literal; every value is a bound
// placeholder and every column is allowlisted. The seed corpus runs in normal
// CI; `go test -fuzz` drives deep exploration (see `make test-fuzz`).

// noRawSQL is the invariant: because all values are parameterized and all columns
// are allowlisted identifiers, the generated SQL can never contain a quote,
// semicolon, or comment marker sourced from input.
func noRawSQL(t *testing.T, sql string) {
	t.Helper()
	if strings.ContainsAny(sql, "';") || strings.Contains(sql, "--") || strings.Contains(sql, "/*") {
		t.Fatalf("client text leaked into SQL text: %q", sql)
	}
}

func FuzzFilterParse(f *testing.F) {
	for _, s := range []string{
		"eq:1", "in:1,2,3", "like:%x%", "gt:5", "neq:", "bogus:1", "",
		"eq:'; DROP TABLE users;--", "in:", "eq:eq:eq", ":", "like:/* */",
	} {
		f.Add("status", s)
	}
	allow := filtering.Allowlist{
		"status": {Col: "status", Ops: []filtering.Op{filtering.OpEq, filtering.OpNeq, filtering.OpIn, filtering.OpLike}},
		"score":  {Col: "score", Ops: []filtering.Op{filtering.OpGt, filtering.OpGte, filtering.OpLt, filtering.OpLte}},
	}
	f.Fuzz(func(t *testing.T, field, entry string) {
		set, err := filtering.Parse(map[string][]string{field: {entry}}, allow)
		if err != nil {
			return // rejected input is fine; it must simply never panic
		}
		sql, args, next := set.SQL(1)
		noRawSQL(t, sql)
		// Placeholder count must equal the bound-arg count — no value un-bound.
		if strings.Count(sql, "$") != len(args) {
			t.Fatalf("placeholder/arg mismatch: sql=%q args=%d", sql, len(args))
		}
		if next != len(args)+1 {
			t.Fatalf("nextArg=%d, want %d", next, len(args)+1)
		}
	})
}

func FuzzParseSort(f *testing.F) {
	for _, s := range []string{
		"created_at:desc,id:asc", "id", "id:bogus", "unknown", "", ",",
		"id:asc;DROP", ":::", "id:asc,id:asc", "id:'--",
	} {
		f.Add(s)
	}
	allow := filtering.SortAllowlist{"created_at": {Col: "created_at"}, "id": {Col: "id"}}
	f.Fuzz(func(t *testing.T, raw string) {
		sort, err := filtering.ParseSort(raw, allow)
		if err != nil {
			return
		}
		noRawSQL(t, sort.SQL())
		// Signature must be derivable without panic for any accepted sort.
		_ = sort.Signature()
	})
}
