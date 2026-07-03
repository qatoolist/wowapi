package policy_test

import (
	"encoding/json"
	"testing"

	"github.com/qatoolist/wowapi/kernel/authz"
	"github.com/qatoolist/wowapi/kernel/policy"
)

func raw(v any) json.RawMessage { b, _ := json.Marshal(v); return b }

func cond(attr, op string, v any) authz.Condition {
	return authz.Condition{Attribute: attr, Op: op, Value: raw(v)}
}

func TestMatchesEmptyIsTrue(t *testing.T) {
	ok, err := policy.New().Matches(nil, map[string]any{})
	if err != nil || !ok {
		t.Fatalf("empty conditions must match: %v %v", ok, err)
	}
}

func TestOperators(t *testing.T) {
	attrs := map[string]any{
		"status": "locked",
		"count":  float64(5),
		"tags":   []any{"a", "b"},
		"time":   "2026-07-03T12:00:00Z",
	}
	cases := []struct {
		name string
		c    authz.Condition
		want bool
	}{
		{"eq-hit", cond("status", "eq", "locked"), true},
		{"eq-miss", cond("status", "eq", "open"), false},
		{"neq-hit", cond("status", "neq", "open"), true},
		{"in-hit", cond("status", "in", []any{"open", "locked"}), true},
		{"in-miss", cond("status", "in", []any{"open"}), false},
		{"not_in-hit", cond("status", "not_in", []any{"open"}), true},
		{"contains-hit", cond("tags", "contains", "a"), true},
		{"contains-miss", cond("tags", "contains", "z"), false},
		{"gte-hit", cond("count", "gte", 5), true},
		{"gte-miss", cond("count", "gte", 6), false},
		{"lte-hit", cond("count", "lte", 5), true},
		{"within-hit", cond("time", "within", []any{"2026-07-03T00:00:00Z", "2026-07-03T23:59:59Z"}), true},
		{"within-miss", cond("time", "within", []any{"2026-07-01T00:00:00Z", "2026-07-02T00:00:00Z"}), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := policy.New().Matches([]authz.Condition{tc.c}, attrs)
			if err != nil {
				t.Fatal(err)
			}
			if got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAllConditionsMustHold(t *testing.T) {
	attrs := map[string]any{"a": "1", "b": "2"}
	ok, err := policy.New().Matches([]authz.Condition{
		cond("a", "eq", "1"),
		cond("b", "eq", "wrong"),
	}, attrs)
	if err != nil || ok {
		t.Fatalf("AND semantics: one failing condition must fail the set: %v %v", ok, err)
	}
}

func TestMissingAttributeDoesNotMatchPositiveOps(t *testing.T) {
	// eq/in/gte on an absent attribute → false (not an error, not a match).
	for _, c := range []authz.Condition{
		cond("absent", "eq", "x"),
		cond("absent", "in", []any{"x"}),
		cond("absent", "gte", 1),
	} {
		ok, err := policy.New().Matches([]authz.Condition{c}, map[string]any{})
		if err != nil || ok {
			t.Errorf("%s on absent attr: ok=%v err=%v", c.Op, ok, err)
		}
	}
}

func TestUnknownOperatorErrors(t *testing.T) {
	// A broken policy must fail loud, never silently stop denying.
	_, err := policy.New().Matches([]authz.Condition{cond("x", "regex", ".*")}, map[string]any{"x": "y"})
	if err == nil {
		t.Fatal("unknown operator must error")
	}
}
