package policy_test

// Hot-path benchmarks for the ABAC policy engine (criterion #17).
//
// policy.Engine is called per-decision when ABAC policies are active. Each
// benchmark measures a different operator shape to expose the cost of JSON
// decode + comparison under common condition types.

import (
	"encoding/json"
	"testing"

	"github.com/qatoolist/wowapi/v2/kernel/authz"
	"github.com/qatoolist/wowapi/v2/kernel/policy"
)

func marshalJSON(t interface{}) json.RawMessage {
	b, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return b
}

// BenchmarkPolicyMatchesEq measures a single "eq" condition match.
// This is the most common condition type.
func BenchmarkPolicyMatchesEq(b *testing.B) {
	eng := policy.New()
	conds := []authz.Condition{
		{Attribute: "actor.impersonating", Op: "eq", Value: marshalJSON(false)},
	}
	attrs := map[string]any{
		"actor.impersonating": false,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eng.Matches(conds, attrs)
	}
}

// BenchmarkPolicyMatchesIn measures a single "in" condition with a small list.
func BenchmarkPolicyMatchesIn(b *testing.B) {
	eng := policy.New()
	conds := []authz.Condition{
		{Attribute: "actor.kind", Op: "in", Value: marshalJSON([]string{"user", "system"})},
	}
	attrs := map[string]any{"actor.kind": "user"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eng.Matches(conds, attrs)
	}
}

// BenchmarkPolicyMatchesGte measures a "gte" (numeric) condition match.
func BenchmarkPolicyMatchesGte(b *testing.B) {
	eng := policy.New()
	conds := []authz.Condition{
		{Attribute: "env.hour", Op: "gte", Value: marshalJSON(9)},
	}
	attrs := map[string]any{"env.hour": 12}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eng.Matches(conds, attrs)
	}
}

// BenchmarkPolicyMatchesMultiCondition measures a 3-condition AND: typical
// for a production deny policy (role + impersonation + hour window).
func BenchmarkPolicyMatchesMultiCondition(b *testing.B) {
	eng := policy.New()
	conds := []authz.Condition{
		{Attribute: "actor.kind", Op: "eq", Value: marshalJSON("user")},
		{Attribute: "actor.impersonating", Op: "eq", Value: marshalJSON(false)},
		{Attribute: "env.hour", Op: "gte", Value: marshalJSON(9)},
	}
	attrs := map[string]any{
		"actor.kind":          "user",
		"actor.impersonating": false,
		"env.hour":            12,
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = eng.Matches(conds, attrs)
	}
}
